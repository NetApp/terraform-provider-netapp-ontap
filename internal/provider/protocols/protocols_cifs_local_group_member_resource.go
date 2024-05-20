package protocols

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"regexp"
	"strings"

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
var _ resource.Resource = &CifsLocalGroupMemberResource{}
var _ resource.ResourceWithImportState = &CifsLocalGroupMemberResource{}

// NewCifsLocalGroupMemberResource is a helper function to simplify the provider implementation.
func NewCifsLocalGroupMemberResource() resource.Resource {
	return &CifsLocalGroupMemberResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_cifs_local_group_member_resource",
		},
	}
}

// CifsLocalGroupMemberResource defines the resource implementation.
type CifsLocalGroupMemberResource struct {
	config connection.ResourceOrDataSourceConfig
}

// CifsLocalGroupMemberResourceModel describes the resource data model.
type CifsLocalGroupMemberResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	GroupName     types.String `tfsdk:"group_name"`
	Member        types.String `tfsdk:"member"`
	SVMName       types.String `tfsdk:"svm_name"`
	ID            types.String `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *CifsLocalGroupMemberResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *CifsLocalGroupMemberResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsLocalGroupMember resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"group_name": schema.StringAttribute{
				MarkdownDescription: "CifsLocalGroupMember name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`\\`),
						"must contain \\\\",
					),
				},
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "CifsLocalGroupMember svm name",
				Required:            true,
			},
			"member": schema.StringAttribute{
				MarkdownDescription: "Member name",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(
						regexp.MustCompile(`\\`),
						"must contain \\\\",
					),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "CifsLocalGroup UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *CifsLocalGroupMemberResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *CifsLocalGroupMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CifsLocalGroupMemberResourceModel

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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("invalid svm name", fmt.Sprintf("protocols_cifs_local_group_members svm_name %s is invalid", data.SVMName.ValueString()))
		return
	}

	// Get group info
	restInfo, err := interfaces.GetCifsLocalGroupByName(errorHandler, *client, data.GroupName.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetCifsLocalGroup
		errorHandler.MakeAndReportError("invalid group name", fmt.Sprintf("protocols_cifs_local_group_members group_name %s is invalid", data.GroupName.ValueString()))
		return
	}
	// Get member
	restInfoMember, err := interfaces.GetCifsLocalGroupMemberByName(errorHandler, *client, svm.UUID, restInfo.SID, data.Member.ValueString())
	if err != nil {
		// error reporting done inside GetCifsLocalGroupMember
		return
	}

	data.Member = types.StringValue(restInfoMember.Name)
	// Set ID
	data.ID = types.StringValue(fmt.Sprintf("%s_%s_%s_%s", data.CxProfileName.ValueString(), data.SVMName.ValueString(), data.GroupName.ValueString(), data.Member.ValueString()))
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *CifsLocalGroupMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *CifsLocalGroupMemberResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.CifsLocalGroupMemberResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Member.ValueString()

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("invalid svm name", fmt.Sprintf("protocols_cifs_local_group_members svm_name %s is invalid", data.SVMName.ValueString()))
		return
	}

	// Get group info
	restInfo, err := interfaces.GetCifsLocalGroupByName(errorHandler, *client, data.GroupName.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetCifsLocalGroup
		errorHandler.MakeAndReportError("invalid group name", fmt.Sprintf("protocols_cifs_local_group_members group_name %s is invalid", data.GroupName.ValueString()))
		return
	}

	resource, err := interfaces.CreateCifsLocalGroupMember(errorHandler, *client, body, svm.UUID, restInfo.SID)
	if err != nil {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("created a resource: %#v", resource))
	data.ID = types.StringValue(fmt.Sprintf("%s_%s_%s_%s", data.CxProfileName.ValueString(), data.SVMName.ValueString(), data.GroupName.ValueString(), data.Member.ValueString()))

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *CifsLocalGroupMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *CifsLocalGroupMemberResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Error(ctx, "Update not supported for protocols_cifs_local_group_member_resource")
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *CifsLocalGroupMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *CifsLocalGroupMemberResourceModel
	var body interfaces.CifsLocalGroupMemberResourceBodyDataModelONTAP
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
		errorHandler.MakeAndReportError("invalid svm name", fmt.Sprintf("protocols_cifs_local_group_members svm_name %s is invalid", data.SVMName.ValueString()))
		return
	}

	// Get group info
	restInfo, err := interfaces.GetCifsLocalGroupByName(errorHandler, *client, data.GroupName.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetCifsLocalGroup
		errorHandler.MakeAndReportError("invalid group name", fmt.Sprintf("protocols_cifs_local_group_members group_name %s is invalid", data.GroupName.ValueString()))
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "protocols_cifs_local_group_member UUID is null")
		return
	}

	body.Name = data.Member.ValueString()
	err = interfaces.DeleteCifsLocalGroupMember(errorHandler, *client, body, svm.UUID, restInfo.SID)
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *CifsLocalGroupMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a protocols cifs local group member resource: %#v", req))
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: member,group_name,svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("member"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[3])...)
}
