package protocols

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsCIFSShareACLResource{}
var _ resource.ResourceWithImportState = &ProtocolsCIFSShareACLResource{}

// NewProtocolsCIFSShareACLResource is a helper function to simplify the provider implementation.
func NewProtocolsCIFSShareACLResource() resource.Resource {
	return &ProtocolsCIFSShareACLResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cifs_share_acl",
		},
	}
}

// NewProtocolsCIFSShareACLResourceAlias is a helper function to simplify the provider implementation.
func NewProtocolsCIFSShareACLResourceAlias() resource.Resource {
	return &ProtocolsCIFSShareACLResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_cifs_share_acl_resource",
		},
	}
}

// ProtocolsCIFSShareACLResource defines the resource implementation.
type ProtocolsCIFSShareACLResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsCIFSShareACLResourceModel describes the resource data model.
type ProtocolsCIFSShareACLResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	ShareName     types.String `tfsdk:"share_name"`
	UserOrGroup   types.String `tfsdk:"user_or_group"`
	Type          types.String `tfsdk:"type"`
	Permission    types.String `tfsdk:"permission"`
	ID            types.String `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *ProtocolsCIFSShareACLResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsCIFSShareACLResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsShareACL resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the SVM that owns the CIFS share",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"share_name": schema.StringAttribute{
				MarkdownDescription: "Name of the CIFS share that owns this ACL",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_or_group": schema.StringAttribute{
				MarkdownDescription: "Specifies the user or group name to add to the access control list of a CIFS share.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Specifies the type of the user or group to add to the access control list of a CIFS share.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("windows", "unix_user", "unix_group"),
				},
			},
			"permission": schema.StringAttribute{
				MarkdownDescription: "Specifies the access rights that a user or group has on the defined CIFS Share.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("full_control", "read", "change", "no_access"),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "CifsShareACL ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsCIFSShareACLResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ProtocolsCIFSShareACLResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsCIFSShareACLResourceModel

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

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		return
	}
	if svm == nil {
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	restInfo, err := interfaces.GetProtocolsCIFSShareACL(errorHandler, *client, svm.UUID, data.ShareName.ValueString(), data.UserOrGroup.ValueString(), data.Type.ValueString())
	if err != nil {
		// error reporting done inside GetProtocolsCIFSShareACL
		return
	}

	data.Permission = types.StringValue(restInfo.Permission)
	data.UserOrGroup = types.StringValue(restInfo.UserOrGroup)
	data.Type = types.StringValue(restInfo.Type)
	if restInfo.Share != "" {
		data.ShareName = types.StringValue(restInfo.Share)
	}
	data.ID = types.StringValue(restInfo.Name)
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve ID
func (r *ProtocolsCIFSShareACLResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProtocolsCIFSShareACLResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ProtocolsCIFSShareACLResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		return
	}

	if !data.Permission.IsUnknown() {
		body.Permission = data.Permission.ValueString()
	}
	if !data.Type.IsUnknown() {
		body.Type = data.Type.ValueString()
	}
	if !data.UserOrGroup.IsUnknown() {
		body.UserOrGroup = data.UserOrGroup.ValueString()
	}

	_, err = interfaces.CreateProtocolsCIFSShareACL(errorHandler, *client, body, svm.UUID, data.ShareName.ValueString())
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetProtocolsCIFSShareACL(errorHandler, *client, svm.UUID, data.ShareName.ValueString(), data.UserOrGroup.ValueString(), data.Type.ValueString())
	if err != nil {
		// error reporting done inside GetProtocolsCIFSShareACL
		return
	}

	data.ID = types.StringValue(restInfo.Name)

	data.Permission = types.StringValue(restInfo.Permission)
	data.UserOrGroup = types.StringValue(restInfo.UserOrGroup)
	data.Type = types.StringValue(restInfo.Type)

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsCIFSShareACLResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state ProtocolsCIFSShareACLResourceModel

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

	svm, err := interfaces.GetSvmByName(errorHandler, *client, state.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		return
	}
	if svm == nil {
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	var body interfaces.ProtocolsCIFSShareACLResourceBodyDataModelONTAP
	if !plan.Permission.Equal(state.Permission) {
		body.Permission = plan.Permission.ValueString()
	}

	err = interfaces.UpdateProtocolsCIFSShareACL(errorHandler, *client, body, svm.UUID, plan.ShareName.ValueString(), plan.UserOrGroup.ValueString(), plan.Type.ValueString())
	if err != nil {
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsCIFSShareACLResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsCIFSShareACLResourceModel

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
		// error reporting done inside GetSvmByName
		return
	}
	if svm == nil {
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	err = interfaces.DeleteProtocolsCIFSShareACL(errorHandler, *client, svm.UUID, data.ShareName.ValueString(), data.UserOrGroup.ValueString(), data.Type.ValueString())
	if err != nil {
		// error reporting done inside DeleteProtocolsCIFSShareACL
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsCIFSShareACLResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a protocols cifs share acl resource: %#v", req))
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 5 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" || idParts[4] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: user_or_group,type,share_name,svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_or_group"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("share_name"), idParts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[3])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[4])...)
}
