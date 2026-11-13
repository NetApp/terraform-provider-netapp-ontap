package protocols

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsFcpServiceResource{}
var _ resource.ResourceWithImportState = &ProtocolsFcpServiceResource{}

// NewProtocolsFcpServiceResource is a helper function to simplify the provider implementation.
func NewProtocolsFcpServiceResource() resource.Resource {
	return &ProtocolsFcpServiceResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "fcp_service",
		},
	}
}

// ProtocolsFcpServiceResource defines the resource implementation.
type ProtocolsFcpServiceResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsFcpServiceResourceModel describes the resource data model.
type ProtocolsFcpServiceResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	ID            types.String `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *ProtocolsFcpServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsFcpServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Protocols FC Protocol service resource",
		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "FC Protocol SVM name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "The administrative state of the FC Protocol service. The FC Protocol service can be disabled to block all FC Protocol connectivity to the SVM. Defaults to true.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "FC Protocol service SVM UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsFcpServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *ProtocolsFcpServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsFcpServiceResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetProtocolsFcpService(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No FCP service found", "FCP service not found.")
		return
	}

	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.ID = types.StringValue(restInfo.SVM.UUID)
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Create a resource and retrieve UUID
func (r *ProtocolsFcpServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProtocolsFcpServiceResourceModel

	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	var body interfaces.ProtocolsFcpServiceResourceDataModelONTAP
	body.SVM.Name = data.SVMName.ValueString()
	body.Enabled = data.Enabled.ValueBool()

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	resource, err := interfaces.CreateProtocolsFcpService(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.ID = types.StringValue(resource.SVM.UUID)
	data.Enabled = types.BoolValue(resource.Enabled)
	tflog.Trace(ctx, fmt.Sprintf("created a resource, UUID=%s", data.ID))

	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsFcpServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ProtocolsFcpServiceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	client, err := connection.GetRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		return
	}

	var body interfaces.ProtocolsFcpServiceResourceDataModelONTAP
	enabledEmpty := true
	if !plan.Enabled.Equal(state.Enabled) {
		enabledEmpty = plan.Enabled.IsNull() || plan.Enabled.IsUnknown()
		if !enabledEmpty {
			body.Enabled = plan.Enabled.ValueBool()
		}
	}

	err = interfaces.UpdateProtocolsFcpService(errorHandler, *client, body, state.ID.ValueString(), enabledEmpty)
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetProtocolsFcpService(errorHandler, *client, plan.SVMName.ValueString())
	if err != nil {
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No FCP service found", "FCP service not found.")
		return
	}

	plan.Enabled = types.BoolValue(restInfo.Enabled)
	plan.ID = types.StringValue(restInfo.SVM.UUID)
	tflog.Trace(ctx, fmt.Sprintf("updated a resource, UUID=%s", plan.ID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsFcpServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsFcpServiceResourceModel

	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	var body interfaces.ProtocolsFcpServiceResourceDataModelONTAP
	body.Enabled = false

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	err = interfaces.UpdateProtocolsFcpService(errorHandler, *client, body, data.ID.ValueString(), false)
	if err != nil {
		return
	}

	err = interfaces.DeleteProtocolsFcpService(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}
	tflog.Trace(ctx, fmt.Sprintf("deleted a resource, UUID=%s", data.ID))
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsFcpServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
