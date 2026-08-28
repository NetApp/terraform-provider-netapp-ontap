package protocols

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsS3ServiceResource{}
var _ resource.ResourceWithImportState = &ProtocolsS3ServiceResource{}
var _ resource.ResourceWithModifyPlan = &ProtocolsS3ServiceResource{}

// NewProtocolsS3ServiceResource is a helper function to simplify the provider implementation.
func NewProtocolsS3ServiceResource() resource.Resource {
	return &ProtocolsS3ServiceResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "s3_service",
		},
	}
}

// ProtocolsS3ServiceResource defines the resource implementation.
type ProtocolsS3ServiceResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsS3ServiceResourceModel describes the resource data model.
type ProtocolsS3ServiceResourceModel struct {
	CxProfileName   types.String `tfsdk:"cx_profile_name"`
	SVMName         types.String `tfsdk:"svm_name"`
	Name            types.String `tfsdk:"name"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Comment         types.String `tfsdk:"comment"`
	CertificateName types.String `tfsdk:"certificate_name"`
	IsHTTPEnabled   types.Bool   `tfsdk:"is_http_enabled"`
	IsHTTPSEnabled  types.Bool   `tfsdk:"is_https_enabled"`
	Port            types.Int64  `tfsdk:"port"`
	SecurePort      types.Int64  `tfsdk:"secure_port"`
	ID              types.String `tfsdk:"id"`
}

// Metadata returns the resource type name
func (r *ProtocolsS3ServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsS3ServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsS3Service resource",
		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "The name of the SVM.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Specifies the name of the S3 server.",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not the S3 service should be enabled.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Additional information about the server being created or modified.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate_name": schema.StringAttribute{
				MarkdownDescription: "Specifies the certificate that will be used for creating HTTPS connections to the S3 server.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_http_enabled": schema.BoolAttribute{
				MarkdownDescription: `
				Specifies whether HTTP is enabled on the S3 server being created or modified.
				By default, HTTP is disabled on the S3 server.
				`,
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"is_https_enabled": schema.BoolAttribute{
				MarkdownDescription: `
				Specifies whether HTTPS is enabled on the S3 server being created or modified.
				By default, HTTPS is enabled on the S3 server.
				`,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: `
				Specifies the HTTP listener port for the S3 server.
				By default, HTTP is enabled on port 80. Valid values range from 1 to 65535.
				`,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"secure_port": schema.Int64Attribute{
				MarkdownDescription: `
				Specifies the HTTPS listener port for the S3 server.
				By default, HTTPS is enabled on port 443. Valid values range from 1 to 65535.
				`,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The name of the SVM.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsS3ServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ModifyPlan warns when ONTAP's default HTTPS setting requires a certificate
// that was not supplied during resource creation.
func (r *ProtocolsS3ServiceResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Only validate create operation.
	// Update operation already has the backend HTTPS and certificate values in state.
	if req.Plan.Raw.IsNull() || !req.State.Raw.IsNull() {
		return
	}

	var config ProtocolsS3ServiceResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpsOmitted := config.IsHTTPSEnabled.IsNull()
	certificateMissing := config.CertificateName.IsNull() ||
		(!config.CertificateName.IsUnknown() && strings.TrimSpace(config.CertificateName.ValueString()) == "")

	if httpsOmitted && certificateMissing {
		resp.Diagnostics.AddAttributeWarning(
			path.Root("certificate_name"),
			"Certificate Name Required by Default HTTPS Setting",
			"When is_https_enabled is omitted during S3 service creation, ONTAP defaults it to true and requires certificate_name. Set certificate_name, or explicitly set is_https_enabled to false.",
		)
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ProtocolsS3ServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsS3ServiceResourceModel

	// Read Terraform prior state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside New Client
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

	var restInfo *interfaces.ProtocolsS3ServiceGetDataModelONTAP
	restInfo, err = interfaces.GetS3Server(errorHandler, *client, data.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetS3Server
		return
	}

	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.Comment = types.StringValue(restInfo.Comment)
	// certificate_name
	if restInfo.Certificate.Name != "" {
		data.CertificateName = types.StringValue(restInfo.Certificate.Name)
	} else {
		data.CertificateName = types.StringNull()
	}
	data.IsHTTPEnabled = types.BoolValue(restInfo.IsHTTPEnabled)
	data.IsHTTPSEnabled = types.BoolValue(restInfo.IsHTTPSEnabled)
	data.Port = types.Int64Value(restInfo.Port)
	data.SecurePort = types.Int64Value(restInfo.SecurePort)
	data.ID = types.StringValue(restInfo.SVM.Name)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve ID
func (r *ProtocolsS3ServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProtocolsS3ServiceResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ProtocolsS3ServiceResourceBodyDataModel
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

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

	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		enabledValue := data.Enabled.ValueBool()
		body.Enabled = &enabledValue
	}
	if !data.Comment.IsNull() && !data.Comment.IsUnknown() {
		commentValue := data.Comment.ValueString()
		body.Comment = &commentValue
	}
	if !data.CertificateName.IsNull() && !data.CertificateName.IsUnknown() {
		certificateValue := data.CertificateName.ValueString()
		body.Certificate = &interfaces.CertificateDataModel{
			Name: certificateValue,
		}
	}
	if !data.IsHTTPEnabled.IsNull() && !data.IsHTTPEnabled.IsUnknown() {
		isHTTPEnabledValue := data.IsHTTPEnabled.ValueBool()
		body.IsHTTPEnabled = &isHTTPEnabledValue
	}
	if !data.IsHTTPSEnabled.IsNull() && !data.IsHTTPSEnabled.IsUnknown() {
		isHTTPSEnabledValue := data.IsHTTPSEnabled.ValueBool()
		body.IsHTTPSEnabled = &isHTTPSEnabledValue
	}
	if !data.Port.IsNull() && !data.Port.IsUnknown() {
		portValue := data.Port.ValueInt64()
		body.Port = &portValue
	}
	if !data.SecurePort.IsNull() && !data.SecurePort.IsUnknown() {
		securePortValue := data.SecurePort.ValueInt64()
		body.SecurePort = &securePortValue
	}

	_, err = interfaces.CreateS3Server(errorHandler, *client, body)
	if err != nil {
		// error reporting done inside CreateS3Server
		return
	}

	restInfo, err := interfaces.GetS3Server(errorHandler, *client, data.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetS3Server
		return
	}

	// Update the computed parameters
	data.ID = types.StringValue(restInfo.SVM.Name)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.Comment = types.StringValue(restInfo.Comment)
	// certificate_name
	if restInfo.Certificate.Name != "" {
		data.CertificateName = types.StringValue(restInfo.Certificate.Name)
	} else {
		data.CertificateName = types.StringNull()
	}
	data.IsHTTPEnabled = types.BoolValue(restInfo.IsHTTPEnabled)
	data.IsHTTPSEnabled = types.BoolValue(restInfo.IsHTTPSEnabled)
	data.Port = types.Int64Value(restInfo.Port)
	data.SecurePort = types.Int64Value(restInfo.SecurePort)

	tflog.Trace(ctx, fmt.Sprintf("created S3 service resource for SVM=%s", data.SVMName.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsS3ServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ProtocolsS3ServiceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Read Terraform state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, state.CxProfileName)
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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, state.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	// Update the resource
	var body interfaces.ProtocolsS3ServiceResourceBodyDataModel

	body.Name = plan.Name.ValueString()
	if !plan.Enabled.Equal(state.Enabled) {
		enabledValue := plan.Enabled.ValueBool()
		body.Enabled = &enabledValue
	}
	if !plan.Comment.Equal(state.Comment) {
		commentValue := plan.Comment.ValueString()
		body.Comment = &commentValue
	}
	if !plan.CertificateName.Equal(state.CertificateName) &&
		!plan.CertificateName.IsNull() && !plan.CertificateName.IsUnknown() {
		if plan.CertificateName.ValueString() != "" {
			certificateValue := plan.CertificateName.ValueString()
			body.Certificate = &interfaces.CertificateDataModel{
				Name: certificateValue,
			}
		}
	}
	if !plan.IsHTTPEnabled.Equal(state.IsHTTPEnabled) {
		isHTTPEnabledValue := plan.IsHTTPEnabled.ValueBool()
		body.IsHTTPEnabled = &isHTTPEnabledValue
	}
	if !plan.IsHTTPSEnabled.Equal(state.IsHTTPSEnabled) {
		isHTTPSEnabledValue := plan.IsHTTPSEnabled.ValueBool()
		body.IsHTTPSEnabled = &isHTTPSEnabledValue
	}
	if !plan.Port.Equal(state.Port) {
		portValue := plan.Port.ValueInt64()
		body.Port = &portValue
	}
	if !plan.SecurePort.Equal(state.SecurePort) {
		securePortValue := plan.SecurePort.ValueInt64()
		body.SecurePort = &securePortValue
	}

	err = interfaces.UpdateS3Server(errorHandler, *client, body, svm.UUID)
	if err != nil {
		// error reporting done inside UpdateS3Server
		return
	}

	restInfo, err := interfaces.GetS3Server(errorHandler, *client, plan.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetS3Server
		return
	}

	// Update the computed parameters
	plan.ID = types.StringValue(restInfo.SVM.Name)
	plan.SVMName = types.StringValue(restInfo.SVM.Name)
	plan.Name = types.StringValue(restInfo.Name)
	plan.Enabled = types.BoolValue(restInfo.Enabled)
	plan.Comment = types.StringValue(restInfo.Comment)
	// certificate_name
	if restInfo.Certificate.Name != "" {
		plan.CertificateName = types.StringValue(restInfo.Certificate.Name)
	} else {
		plan.CertificateName = types.StringNull()
	}
	plan.IsHTTPEnabled = types.BoolValue(restInfo.IsHTTPEnabled)
	plan.IsHTTPSEnabled = types.BoolValue(restInfo.IsHTTPSEnabled)
	plan.Port = types.Int64Value(restInfo.Port)
	plan.SecurePort = types.Int64Value(restInfo.SecurePort)

	tflog.Debug(ctx, fmt.Sprintf("updated S3 service resource for SVM=%s", plan.SVMName.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsS3ServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsS3ServiceResourceModel

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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	err = interfaces.DeleteS3Server(errorHandler, *client, svm.UUID)
	if err != nil {
		// error reporting done inside DeleteS3Server
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsS3ServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
