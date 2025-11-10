package support

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &AutoSupportResource{}
var _ resource.ResourceWithImportState = &AutoSupportResource{}

// NewAutoSupportResource is a helper function to simplify the provider implementation.
func NewAutoSupportResource() resource.Resource {
	return &AutoSupportResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "autosupport",
		},
	}
}

// AutoSupportResource defines the resource implementation.
type AutoSupportResource struct {
	config connection.ResourceOrDataSourceConfig
}

// AutoSupportResourceModel describes the resource data model.
type AutoSupportResourceModel struct {
	CxProfileName                 types.String `tfsdk:"cx_profile_name"`
	ID                            types.String `tfsdk:"id"`
	Enabled                       types.Bool   `tfsdk:"enabled"`
	Transport                     types.String `tfsdk:"transport"`
	To                            types.Set    `tfsdk:"to_addresses"`
	From                          types.String `tfsdk:"from"`
	ContactSupport                types.Bool   `tfsdk:"contact_support"`
	PartnerAddresses              types.Set    `tfsdk:"partner_addresses"`
	ProxyURL                      types.String `tfsdk:"proxy_url"`
	MailHosts                     types.Set    `tfsdk:"mail_hosts"`
	IsMinimal                     types.Bool   `tfsdk:"is_minimal"`
	OndemandEnabled               types.Bool   `tfsdk:"ondemand_enabled"`
	SmtpEncryption                types.String `tfsdk:"smtp_encryption"`
	Force                         types.Bool   `tfsdk:"force"`
}

// Metadata returns the resource type name.
func (r *AutoSupportResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *AutoSupportResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AutoSupport resource. Manages the AutoSupport configuration that always exists on ONTAP clusters. The 'create' operation configures the existing AutoSupport service.",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "AutoSupport identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the AutoSupport daemon is enabled. When enabled, AutoSupport messages are generated",
				Optional:            true,
				Computed:            true,
			},
			"transport": schema.StringAttribute{
				MarkdownDescription: "The name of the transport protocol used to deliver AutoSupport messages",
				Optional:            true,
				Computed:            true,
			},
			"to_addresses": schema.SetAttribute{
				MarkdownDescription: "Specifies up to five recipients of full AutoSupport e-mail messages",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"from": schema.StringAttribute{
				MarkdownDescription: "The e-mail address from which the node sends AutoSupport messages",
				Optional:            true,
				Computed:            true,
			},
			"contact_support": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether AutoSupport notification to technical support is enabled",
				Optional:            true,
				Computed:            true,
			},
			"partner_addresses": schema.SetAttribute{
				MarkdownDescription: "Specifies up to five partner vendor recipients of full AutoSupport e-mail messages",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"proxy_url": schema.StringAttribute{
				MarkdownDescription: "HTTP or HTTPS proxy if the transport parameter is set to HTTP or HTTPS",
				Optional:            true,
				Computed:            true,
			},
			"mail_hosts": schema.SetAttribute{
				MarkdownDescription: "List of mail server(s) used to deliver AutoSupport messages via SMTP",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
			},
			"is_minimal": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the system information is collected in compliant form to remove private data or in complete form to enhance diagnostics",
				Optional:            true,
				Computed:            true,
			},
			"ondemand_enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the AutoSupport OnDemand Download feature is enabled",
				Optional:            true,
				Computed:            true,
			},
			"smtp_encryption": schema.StringAttribute{
				MarkdownDescription: "SMTP encryption type for AutoSupport messages",
				Optional:            true,
				Computed:            true,
			},
			"force": schema.BoolAttribute{
				MarkdownDescription: "Force the configuration update even if it might disrupt AutoSupport operations. Requires ONTAP 9.16.0 or higher.",
				Optional:            true,
				Computed:            false,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *AutoSupportResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *AutoSupportResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data AutoSupportResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	restInfo, err := interfaces.GetAutoSupport(errorHandler, *client, cluster.Version)
	if err != nil {
		// error reporting done inside GetAutoSupport
		return
	}

	data.ID = data.CxProfileName // For singleton resources, use cx_profile_name as ID
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.Transport = types.StringValue(restInfo.Transport)
	data.From = types.StringValue(restInfo.From)
	data.ContactSupport = types.BoolValue(restInfo.ContactSupport)
	data.ProxyURL = types.StringValue(restInfo.ProxyURL)
	data.IsMinimal = types.BoolValue(restInfo.IsMinimal)
	data.OndemandEnabled = types.BoolValue(restInfo.OndemandEnabled)
	data.SmtpEncryption = types.StringValue(restInfo.SmtpEncryption)

	// Handle string slices for sets
	if restInfo.To != nil {
		elements := []attr.Value{}
		for _, address := range restInfo.To {
			elements = append(elements, types.StringValue(address))
		}
		setVal, diags := types.SetValue(types.StringType, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.To = setVal
	} else {
		data.To = types.SetNull(types.StringType)
	}

	if restInfo.PartnerAddresses != nil {
		elements := []attr.Value{}
		for _, address := range restInfo.PartnerAddresses {
			elements = append(elements, types.StringValue(address))
		}
		setVal, diags := types.SetValue(types.StringType, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.PartnerAddresses = setVal
	} else {
		data.PartnerAddresses = types.SetNull(types.StringType)
	}

	if restInfo.MailHosts != nil {
		elements := []attr.Value{}
		for _, host := range restInfo.MailHosts {
			elements = append(elements, types.StringValue(host))
		}
		setVal, diags := types.SetValue(types.StringType, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.MailHosts = setVal
	} else {
		data.MailHosts = types.SetNull(types.StringType)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *AutoSupportResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Note: AutoSupport configuration always exists on ONTAP clusters.
	// "Create" operation actually configures the existing AutoSupport service.
	
	var data *AutoSupportResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var bodyMap = make(map[string]interface{})
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}
	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}
	var errors []string

	// Convert Terraform plan data to API body map - only include configured fields
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		bodyMap["enabled"] = data.Enabled.ValueBool()
	}
	if !data.Transport.IsNull() && !data.Transport.IsUnknown() {
		bodyMap["transport"] = data.Transport.ValueString()
	}
	if !data.From.IsNull() && !data.From.IsUnknown() {
		bodyMap["from"] = data.From.ValueString()
	}
	if !data.ContactSupport.IsNull() && !data.ContactSupport.IsUnknown() {
		bodyMap["contact_support"] = data.ContactSupport.ValueBool()
	}
	if !data.ProxyURL.IsNull() && !data.ProxyURL.IsUnknown() {
		bodyMap["proxy_url"] = data.ProxyURL.ValueString()
	}
	if !data.IsMinimal.IsNull() && !data.IsMinimal.IsUnknown() {
		bodyMap["is_minimal"] = data.IsMinimal.ValueBool()
	}
	if !data.OndemandEnabled.IsNull() && !data.OndemandEnabled.IsUnknown() {
		ondemandEnabled := data.OndemandEnabled.ValueBool()
		// ondemand_enabled is only supported in ONTAP 9.16.1 or higher
		if cluster.Version.Generation == 9 && (cluster.Version.Major > 16 || (cluster.Version.Major == 16 && cluster.Version.Minor >= 1)) {
			bodyMap["ondemand_enabled"] = ondemandEnabled
		} else {
			errors = append(errors, "ondemand_enabled")
		}
	}
	if !data.SmtpEncryption.IsNull() && !data.SmtpEncryption.IsUnknown() {
		smtpEncryption := data.SmtpEncryption.ValueString()
		// smtp_encryption is only supported in ONTAP 9.15.0 or higher
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 15 {
			bodyMap["smtp_encryption"] = smtpEncryption
		} else {
			errors = append(errors, "smtp_encryption")
		}
	}

	// Handle sets
	if !data.To.IsNull() && !data.To.IsUnknown() {
		var toAddresses []string
		diags := data.To.ElementsAs(ctx, &toAddresses, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		bodyMap["to"] = toAddresses
	}
	if !data.PartnerAddresses.IsNull() && !data.PartnerAddresses.IsUnknown() {
		var partnerAddresses []string
		diags := data.PartnerAddresses.ElementsAs(ctx, &partnerAddresses, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		bodyMap["partner_addresses"] = partnerAddresses
	}
	if !data.MailHosts.IsNull() && !data.MailHosts.IsUnknown() {
		var mailHosts []string
		diags := data.MailHosts.ElementsAs(ctx, &mailHosts, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		bodyMap["mail_hosts"] = mailHosts
	}

	force := false
	if !data.Force.IsNull() && !data.Force.IsUnknown() {
		// force parameter is only supported in ONTAP 9.16.0 or higher
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 16 {
			force = data.Force.ValueBool()
		} else {
			errors = append(errors, "force")
		}
	}

	// Check for version restriction errors before making the API call
	if len(errors) > 0 {
		errorMessage := "The following fields are not supported in this ONTAP version: " + strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("AutoSupport configuration validation failed: %v", errorMessage))
		resp.Diagnostics.AddError(
			"AutoSupport configuration validation failed",
			errorMessage,
		)
		return
	}

	err = interfaces.UpdateAutoSupport(errorHandler, *client, force, bodyMap)
	if err != nil {
		return
	}

	// Read back the updated configuration
	// Get cluster version for reading updated configuration
	readCluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if readCluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	restInfo, err := interfaces.GetAutoSupport(errorHandler, *client, readCluster.Version)
	if err != nil {
		// error reporting done inside GetAutoSupport
		return
	}

	// Update data with current state
	data.ID = data.CxProfileName // For singleton resources, use cx_profile_name as ID
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.Transport = types.StringValue(restInfo.Transport)
	data.From = types.StringValue(restInfo.From)
	data.ContactSupport = types.BoolValue(restInfo.ContactSupport)
	data.ProxyURL = types.StringValue(restInfo.ProxyURL)
	data.IsMinimal = types.BoolValue(restInfo.IsMinimal)
	data.OndemandEnabled = types.BoolValue(restInfo.OndemandEnabled)
	data.SmtpEncryption = types.StringValue(restInfo.SmtpEncryption)

	// Handle string slices for sets
	if restInfo.To != nil {
		elements := []attr.Value{}
		for _, address := range restInfo.To {
			elements = append(elements, types.StringValue(address))
		}
		setVal, diags := types.SetValue(types.StringType, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.To = setVal
	} else {
		data.To = types.SetNull(types.StringType)
	}

	if restInfo.PartnerAddresses != nil {
		elements := []attr.Value{}
		for _, address := range restInfo.PartnerAddresses {
			elements = append(elements, types.StringValue(address))
		}
		setVal, diags := types.SetValue(types.StringType, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.PartnerAddresses = setVal
	} else {
		data.PartnerAddresses = types.SetNull(types.StringType)
	}

	if restInfo.MailHosts != nil {
		elements := []attr.Value{}
		for _, host := range restInfo.MailHosts {
			elements = append(elements, types.StringValue(host))
		}
		setVal, diags := types.SetValue(types.StringType, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.MailHosts = setVal
	} else {
		data.MailHosts = types.SetNull(types.StringType)
	}

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AutoSupportResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *AutoSupportResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var bodyMap = make(map[string]interface{})
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Get cluster version for version restrictions
	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	var errors []string

	// Convert Terraform plan data to API body map - only include configured fields
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		bodyMap["enabled"] = data.Enabled.ValueBool()
	}
	if !data.Transport.IsNull() && !data.Transport.IsUnknown() {
		bodyMap["transport"] = data.Transport.ValueString()
	}
	if !data.From.IsNull() && !data.From.IsUnknown() {
		bodyMap["from"] = data.From.ValueString()
	}
	if !data.ContactSupport.IsNull() && !data.ContactSupport.IsUnknown() {
		bodyMap["contact_support"] = data.ContactSupport.ValueBool()
	}
	if !data.ProxyURL.IsNull() && !data.ProxyURL.IsUnknown() {
		bodyMap["proxy_url"] = data.ProxyURL.ValueString()
	}
	if !data.IsMinimal.IsNull() && !data.IsMinimal.IsUnknown() {
		bodyMap["is_minimal"] = data.IsMinimal.ValueBool()
	}
	if !data.OndemandEnabled.IsNull() && !data.OndemandEnabled.IsUnknown() {
		ondemandEnabled := data.OndemandEnabled.ValueBool()
		// ondemand_enabled is only supported in ONTAP 9.16.1 or higher
		if cluster.Version.Generation == 9 && (cluster.Version.Major > 16 || (cluster.Version.Major == 16 && cluster.Version.Minor >= 1)) {
			bodyMap["ondemand_enabled"] = ondemandEnabled
		} else {
			errors = append(errors, "ondemand_enabled")
		}
	}
	if !data.SmtpEncryption.IsNull() && !data.SmtpEncryption.IsUnknown() {
		smtpEncryption := data.SmtpEncryption.ValueString()
		// smtp_encryption is only supported in ONTAP 9.15.0 or higher
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 15 {
			bodyMap["smtp_encryption"] = smtpEncryption
		} else {
			errors = append(errors, "smtp_encryption")
		}
	}

	// Handle sets
	if !data.To.IsNull() && !data.To.IsUnknown() {
		var toAddresses []string
		diags := data.To.ElementsAs(ctx, &toAddresses, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		bodyMap["to"] = toAddresses
	}
	if !data.PartnerAddresses.IsNull() && !data.PartnerAddresses.IsUnknown() {
		var partnerAddresses []string
		diags := data.PartnerAddresses.ElementsAs(ctx, &partnerAddresses, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		bodyMap["partner_addresses"] = partnerAddresses
	}
	if !data.MailHosts.IsNull() && !data.MailHosts.IsUnknown() {
		var mailHosts []string
		diags := data.MailHosts.ElementsAs(ctx, &mailHosts, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		bodyMap["mail_hosts"] = mailHosts
	}

	// Check for version restriction errors before making the API call
	if len(errors) > 0 {
		errorMessage := "The following fields are not supported in this ONTAP version: " + strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("AutoSupport configuration validation failed: %v", errorMessage))
		resp.Diagnostics.AddError(
			"AutoSupport configuration validation failed",
			errorMessage,
		)
		return
	}

	// Extract force parameter separately and check version compatibility
	force := false
	if !data.Force.IsNull() && !data.Force.IsUnknown() {
		// force parameter is only supported in ONTAP 9.16.0 or higher
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 16 {
			force = data.Force.ValueBool()
		} else {
			errors = append(errors, "force")
		}
	}

	// Check for version restriction errors before making the API call
	if len(errors) > 0 {
		errorMessage := "The following fields are not supported in this ONTAP version: " + strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("AutoSupport configuration validation failed: %v", errorMessage))
		resp.Diagnostics.AddError(
			"AutoSupport configuration validation failed",
			errorMessage,
		)
		return
	}

	err = interfaces.UpdateAutoSupport(errorHandler, *client, force, bodyMap)
	if err != nil {
		return
	}

	// Get cluster version for reading updated configuration
	readCluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if readCluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	restInfo, err := interfaces.GetAutoSupport(errorHandler, *client, readCluster.Version)
	if err != nil {
		// error reporting done inside GetAutoSupport
		return
	}

	// Update data with current state
	if err != nil {
		return
	}

	// Update data with current state
	data.ID = data.CxProfileName // For singleton resources, use cx_profile_name as ID
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.Transport = types.StringValue(restInfo.Transport)
	data.From = types.StringValue(restInfo.From)
	data.ContactSupport = types.BoolValue(restInfo.ContactSupport)
	data.ProxyURL = types.StringValue(restInfo.ProxyURL)
	data.IsMinimal = types.BoolValue(restInfo.IsMinimal)
	data.OndemandEnabled = types.BoolValue(restInfo.OndemandEnabled)
	data.SmtpEncryption = types.StringValue(restInfo.SmtpEncryption)

	// Handle string slices for sets
	if restInfo.To != nil {
		elements := []attr.Value{}
		for _, address := range restInfo.To {
			elements = append(elements, types.StringValue(address))
		}
		setVal, diags := types.SetValue(types.StringType, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.To = setVal
	} else {
		data.To = types.SetNull(types.StringType)
	}

	if restInfo.PartnerAddresses != nil {
		elements := []attr.Value{}
		for _, address := range restInfo.PartnerAddresses {
			elements = append(elements, types.StringValue(address))
		}
		setVal, diags := types.SetValue(types.StringType, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.PartnerAddresses = setVal
	} else {
		data.PartnerAddresses = types.SetNull(types.StringType)
	}

	if restInfo.MailHosts != nil {
		elements := []attr.Value{}
		for _, host := range restInfo.MailHosts {
			elements = append(elements, types.StringValue(host))
		}
		setVal, diags := types.SetValue(types.StringType, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.MailHosts = setVal
	} else {
		data.MailHosts = types.SetNull(types.StringType)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AutoSupportResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// AutoSupport configuration cannot be deleted from the cluster.
	// This function exists to satisfy the Resource interface.
	// When the resource is removed from Terraform, the state is simply removed
	// but the AutoSupport configuration remains on the cluster.
	tflog.Debug(ctx, "AutoSupport resource removed from Terraform state - configuration remains on cluster")
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *AutoSupportResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// For singleton resources, we just need to set the cx_profile_name from the import ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), req.ID)...)
}
