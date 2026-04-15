package name_services

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &NameServicesDNSResource{}
var _ resource.ResourceWithImportState = &NameServicesDNSResource{}

// NewNameServicesDNSResource is a helper function to simplify the provider implementation.
func NewNameServicesDNSResource() resource.Resource {
	return &NameServicesDNSResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "dns",
		},
	}
}

// NewNameServicesDNSResourceAlias is a helper function to simplify the provider implementation.
func NewNameServicesDNSResourceAlias() resource.Resource {
	return &NameServicesDNSResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "name_services_dns_resource",
		},
	}
}

// NameServicesDNSResource defines the resource implementation.
type NameServicesDNSResource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesDNSResourceModel describes the resource data model.
type NameServicesDNSResourceModel struct {
	CxProfileName        types.String   `tfsdk:"cx_profile_name"`
	SVMName              types.String   `tfsdk:"svm_name"`
	ID                   types.String   `tfsdk:"id"`
	SkipConfigValidation types.Bool     `tfsdk:"skip_config_validation"`
	Domains              []types.String `tfsdk:"dns_domains"`
	NameServers          []types.String `tfsdk:"name_servers"`
	DynamicDNS           types.Object   `tfsdk:"dynamic_dns"`
}

type NameServicesDNSDynamicDNSModel struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	FQDN               types.String `tfsdk:"fqdn"`
	SkipFQDNValidation types.Bool   `tfsdk:"skip_fqdn_validation"`
	TimeToLive         types.String `tfsdk:"time_to_live"`
	UseSecure          types.Bool   `tfsdk:"use_secure"`
}

// Metadata returns the resource type name.
func (r *NameServicesDNSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *NameServicesDNSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesDNS resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface svm name",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "UUID of svm",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dns_domains": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of DNS domains such as 'sales.bar.com'. The first domain is the one that the svm belongs to",
				Optional:            true,
			},
			"name_servers": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of IPv4 addresses of name servers such as '123.123.123.123'.",
				Optional:            true,
			},
			"skip_config_validation": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether or not the validation for the specified DNS configuration is disabled. (9.9)",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dynamic_dns": schema.SingleNestedAttribute{
				MarkdownDescription: "Dynamic DNS update configuration for the SVM.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"fqdn": schema.StringAttribute{
						MarkdownDescription: "Fully Qualified Domain Name (FQDN) to be used for dynamic DNS updates",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"time_to_live": schema.StringAttribute{
						MarkdownDescription: "Time to live value for the dynamic DNS updates, in an ISO-8601 duration formatted string",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"skip_fqdn_validation": schema.BoolAttribute{
						MarkdownDescription: "Enable or disable FQDN validation",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"use_secure": schema.BoolAttribute{
						MarkdownDescription: "Enable or disable secure dynamic DNS updates for the specified SVM",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable or disable Dynamic DNS (DDNS) updates for the specified SVM",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
		},
	}
}
// Configure adds the provider configured client to the resource.
func (r *NameServicesDNSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *NameServicesDNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NameServicesDNSResourceModel

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

	restInfo, err := interfaces.GetNameServicesDNS(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetNameServicesDNS
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No DNS found", fmt.Sprintf("NO DNS on svm %s found.", data.SVMName.ValueString()))
		return
	}

	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.ID = types.StringValue(restInfo.SVM.UUID)

	if restInfo.Servers != nil {
		data.NameServers = make([]types.String, len(restInfo.Servers))
		for index, server := range restInfo.Servers {
			data.NameServers[index] = types.StringValue(server)
		}
	}

	if restInfo.Domains != nil {
		data.Domains = make([]types.String, len(restInfo.Domains))
		for index, domain := range restInfo.Domains {
			data.Domains[index] = types.StringValue(domain)
		}
	}

	dynamicDNSAttrTypes := map[string]attr.Type{
		"enabled":              types.BoolType,
		"fqdn":                 types.StringType,
		"skip_fqdn_validation": types.BoolType,
		"time_to_live":         types.StringType,
		"use_secure":           types.BoolType,
	}

	if restInfo.DynamicDNS != nil {
		// If state already has skip_fqdn_validation, prefer state to avoid drift when ONTAP omits/normalizes this field in dynamic DNS read responses.
		skipFQDNValidation := types.BoolValue(restInfo.DynamicDNS.SkipFQDNValidation)
		if !data.DynamicDNS.IsNull() && !data.DynamicDNS.IsUnknown() {
			var stateDynamicDNSModel NameServicesDNSDynamicDNSModel
			diags := data.DynamicDNS.As(ctx, &stateDynamicDNSModel, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !stateDynamicDNSModel.SkipFQDNValidation.IsNull() && !stateDynamicDNSModel.SkipFQDNValidation.IsUnknown() {
				skipFQDNValidation = stateDynamicDNSModel.SkipFQDNValidation
			}
		}

		dynamicDNSValues := map[string]attr.Value{
			"enabled":              types.BoolValue(restInfo.DynamicDNS.Enabled),
			"fqdn":                 types.StringValue(restInfo.DynamicDNS.FQDN),
			"skip_fqdn_validation": skipFQDNValidation,
			"time_to_live":         types.StringValue(restInfo.DynamicDNS.TimeToLive),
			"use_secure":           types.BoolValue(restInfo.DynamicDNS.UseSecure),
		}
		dynamicDNSObject, diags := types.ObjectValue(dynamicDNSAttrTypes, dynamicDNSValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.DynamicDNS = dynamicDNSObject
	} else {
		data.DynamicDNS = types.ObjectNull(dynamicDNSAttrTypes)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *NameServicesDNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *NameServicesDNSResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body := interfaces.NameServicesDNSGetDataModelONTAP{
		Servers: []string{},
		Domains: []string{},
	}
	for _, value := range data.NameServers {
		body.Servers = append(body.Servers, value.ValueString())
	}
	for _, value := range data.Domains {
		body.Domains = append(body.Domains, value.ValueString())
	}
	if !data.SkipConfigValidation.IsNull() && !data.SkipConfigValidation.IsUnknown() {
		body.SkipConfigValidation = data.SkipConfigValidation.ValueBool()
	}
	if !data.DynamicDNS.IsNull() && !data.DynamicDNS.IsUnknown() {
		var dynamicDNSModel NameServicesDNSDynamicDNSModel
		diags := data.DynamicDNS.As(ctx, &dynamicDNSModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		dynamicDNS := interfaces.DynamicDNS{}
		hasDynamicDNS := false

		if !dynamicDNSModel.Enabled.IsNull() && !dynamicDNSModel.Enabled.IsUnknown() {
			dynamicDNS.Enabled = dynamicDNSModel.Enabled.ValueBool()
			hasDynamicDNS = true
		}
		if !dynamicDNSModel.FQDN.IsNull() && !dynamicDNSModel.FQDN.IsUnknown() {
			dynamicDNS.FQDN = dynamicDNSModel.FQDN.ValueString()
			hasDynamicDNS = true
		}
		if !dynamicDNSModel.SkipFQDNValidation.IsNull() && !dynamicDNSModel.SkipFQDNValidation.IsUnknown() {
			dynamicDNS.SkipFQDNValidation = dynamicDNSModel.SkipFQDNValidation.ValueBool()
			hasDynamicDNS = true
		}
		if !dynamicDNSModel.TimeToLive.IsNull() && !dynamicDNSModel.TimeToLive.IsUnknown() {
			dynamicDNS.TimeToLive = dynamicDNSModel.TimeToLive.ValueString()
			hasDynamicDNS = true
		}
		if !dynamicDNSModel.UseSecure.IsNull() && !dynamicDNSModel.UseSecure.IsUnknown() {
			dynamicDNS.UseSecure = dynamicDNSModel.UseSecure.ValueBool()
			hasDynamicDNS = true
		}

		if hasDynamicDNS {
			body.DynamicDNS = &dynamicDNS
		}
	}
	body.SVM.Name = data.SVMName.ValueString()
	body.SVM.UUID = data.ID.ValueString()

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	dns, createErr := interfaces.CreateNameServicesDNS(errorHandler, *client, body)
	if createErr != nil {
		return
	}
	data.ID = types.StringValue(dns.SVM.UUID)

	restInfo, err := interfaces.GetNameServicesDNS(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No DNS found", fmt.Sprintf("NO DNS on svm %s found.", data.SVMName.ValueString()))
		return
	}

	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.ID = types.StringValue(restInfo.SVM.UUID)

	if restInfo.Servers != nil {
		data.NameServers = make([]types.String, len(restInfo.Servers))
		for index, server := range restInfo.Servers {
			data.NameServers[index] = types.StringValue(server)
		}
	}

	if restInfo.Domains != nil {
		data.Domains = make([]types.String, len(restInfo.Domains))
		for index, domain := range restInfo.Domains {
			data.Domains[index] = types.StringValue(domain)
		}
	}

	dynamicDNSAttrTypes := map[string]attr.Type{
		"enabled":              types.BoolType,
		"fqdn":                 types.StringType,
		"skip_fqdn_validation": types.BoolType,
		"time_to_live":         types.StringType,
		"use_secure":           types.BoolType,
	}

	if restInfo.DynamicDNS != nil {
		// If state already has skip_fqdn_validation, prefer state to avoid drift when ONTAP omits/normalizes this field in dynamic DNS read responses.
		skipFQDNValidation := types.BoolValue(restInfo.DynamicDNS.SkipFQDNValidation)
		if !data.DynamicDNS.IsNull() && !data.DynamicDNS.IsUnknown() {
			var stateDynamicDNSModel NameServicesDNSDynamicDNSModel
			diags := data.DynamicDNS.As(ctx, &stateDynamicDNSModel, basetypes.ObjectAsOptions{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !stateDynamicDNSModel.SkipFQDNValidation.IsNull() && !stateDynamicDNSModel.SkipFQDNValidation.IsUnknown() {
				skipFQDNValidation = stateDynamicDNSModel.SkipFQDNValidation
			}
		}

		dynamicDNSValues := map[string]attr.Value{
			"enabled":              types.BoolValue(restInfo.DynamicDNS.Enabled),
			"fqdn":                 types.StringValue(restInfo.DynamicDNS.FQDN),
			"skip_fqdn_validation": skipFQDNValidation,
			"time_to_live":         types.StringValue(restInfo.DynamicDNS.TimeToLive),
			"use_secure":           types.BoolValue(restInfo.DynamicDNS.UseSecure),
		}
		dynamicDNSObject, diags := types.ObjectValue(dynamicDNSAttrTypes, dynamicDNSValues)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.DynamicDNS = dynamicDNSObject
	} else {
		data.DynamicDNS = types.ObjectNull(dynamicDNSAttrTypes)
	}

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *NameServicesDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *NameServicesDNSResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}


	request := interfaces.NameServicesDNSGetDataModelONTAP{
		Servers: []string{},
		Domains: []string{},
	}

	for _, value := range data.NameServers {
		request.Servers = append(request.Servers, value.ValueString())
	}
	for _, value := range data.Domains {
		request.Domains = append(request.Domains, value.ValueString())
	}

	if !data.SkipConfigValidation.IsNull() && !data.SkipConfigValidation.IsUnknown() {
		request.SkipConfigValidation = data.SkipConfigValidation.ValueBool()
	}

	if !data.DynamicDNS.IsNull() && !data.DynamicDNS.IsUnknown() {
		var dynamicDNSModel NameServicesDNSDynamicDNSModel
		diags := data.DynamicDNS.As(ctx, &dynamicDNSModel, basetypes.ObjectAsOptions{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		dynamicDNS := interfaces.DynamicDNS{}
		hasDynamicDNS := false

		if !dynamicDNSModel.Enabled.IsNull() && !dynamicDNSModel.Enabled.IsUnknown() {
			dynamicDNS.Enabled = dynamicDNSModel.Enabled.ValueBool()
			hasDynamicDNS = true
		}
		if !dynamicDNSModel.FQDN.IsNull() && !dynamicDNSModel.FQDN.IsUnknown() {
			dynamicDNS.FQDN = dynamicDNSModel.FQDN.ValueString()
			hasDynamicDNS = true
		}
		if !dynamicDNSModel.SkipFQDNValidation.IsNull() && !dynamicDNSModel.SkipFQDNValidation.IsUnknown() {
			dynamicDNS.SkipFQDNValidation = dynamicDNSModel.SkipFQDNValidation.ValueBool()
			hasDynamicDNS = true
		}
		if !dynamicDNSModel.TimeToLive.IsNull() && !dynamicDNSModel.TimeToLive.IsUnknown() {
			dynamicDNS.TimeToLive = dynamicDNSModel.TimeToLive.ValueString()
			hasDynamicDNS = true
		}
		if !dynamicDNSModel.UseSecure.IsNull() && !dynamicDNSModel.UseSecure.IsUnknown() {
			dynamicDNS.UseSecure = dynamicDNSModel.UseSecure.ValueBool()
			hasDynamicDNS = true
		}

		if hasDynamicDNS {
			request.DynamicDNS = &dynamicDNS
		}
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("ID is null", "name_services_dns ID is null")
		return
	}

	err = interfaces.UpdateNameServicesDNS(errorHandler, *client, request, data.ID.ValueString())
	if err != nil {
		return
	}
	tflog.Trace(ctx, "updated a name_services_dns resource")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *NameServicesDNSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *NameServicesDNSResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	err = interfaces.DeleteNameServicesDNS(errorHandler, *client, svm.UUID)
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *NameServicesDNSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[1])...)
}
