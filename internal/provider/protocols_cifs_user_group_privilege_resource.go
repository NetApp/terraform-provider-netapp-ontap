package provider

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
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
var _ resource.Resource = &CifsUserGroupPrivilegeResource{}
var _ resource.ResourceWithImportState = &CifsUserGroupPrivilegeResource{}

// NewCifsUserGroupPrivilegeResource is a helper function to simplify the provider implementation.
func NewCifsUserGroupPrivilegeResource() resource.Resource {
	return &CifsUserGroupPrivilegeResource{
		config: resourceOrDataSourceConfig{
			name: "protocols_cifs_user_group_privilege_resource",
		},
	}
}

// CifsUserGroupPrivilegeResource defines the resource implementation.
type CifsUserGroupPrivilegeResource struct {
	config resourceOrDataSourceConfig
}

// CifsUserGroupPrivilegeResourceModel describes the resource data model.
type CifsUserGroupPrivilegeResourceModel struct {
	CxProfileName types.String   `tfsdk:"cx_profile_name"`
	Name          types.String   `tfsdk:"name"`
	SVMName       types.String   `tfsdk:"svm_name"`
	Privileges    []types.String `tfsdk:"privileges"`
	ID            types.String   `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *CifsUserGroupPrivilegeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *CifsUserGroupPrivilegeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsUserGroupPrivilege resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "CifsUserGroupPrivilege name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "CifsUserGroupPrivilege svm name",
				Required:            true,
			},
			"privileges": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of privileges",
				Required:            true,
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.RegexMatches(
						regexp.MustCompile(`^[a-z]*$`),
						"must only contain lower case characters",
					)),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "CifsUserGroupPrivilege ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *CifsUserGroupPrivilegeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *CifsUserGroupPrivilegeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CifsUserGroupPrivilegeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}
	restInfo, err := interfaces.GetCifsUserGroupPrivilegeByName(errorHandler, *client, data.Name.ValueString(), svm.Name)
	if err != nil {
		// error reporting done inside GetCifsUserGroupPrivilege
		return
	}

	data.Privileges = make([]types.String, len(restInfo.Privileges))
	for index, privilege := range restInfo.Privileges {
		data.Privileges[index] = types.StringValue(privilege)
	}
	// Set the ID
	data.ID = types.StringValue(fmt.Sprintf("%s_%s_%s", data.CxProfileName.ValueString(), data.SVMName.ValueString(), data.Name.ValueString()))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve ID
func (r *CifsUserGroupPrivilegeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *CifsUserGroupPrivilegeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.CifsUserGroupPrivilegeResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()
	body.Privileges = make([]string, len(data.Privileges))
	for i, privilege := range data.Privileges {
		body.Privileges[i] = privilege.ValueString()
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	_, err = interfaces.CreateCifsUserGroupPrivilege(errorHandler, *client, body)
	if err != nil {
		return
	}

	// Set the ID
	data.ID = types.StringValue(fmt.Sprintf("%s_%s_%s", data.CxProfileName.ValueString(), data.SVMName.ValueString(), data.Name.ValueString()))

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *CifsUserGroupPrivilegeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *CifsUserGroupPrivilegeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("ID is null", "protocols_cifs_user_group_privilege ID is null")
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}

	var body interfaces.CifsUserGroupPrivilegeResourceBodyDataModelONTAP

	body.Privileges = make([]string, len(data.Privileges))
	for i, privilege := range data.Privileges {
		body.Privileges[i] = privilege.ValueString()
	}
	_, err = interfaces.UpdateCifsUserGroupPrivilege(errorHandler, *client, body, svm.UUID, data.Name.ValueString())
	if err != nil {
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *CifsUserGroupPrivilegeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *CifsUserGroupPrivilegeResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("ID is null", "protocols_cifs_user_group_privilege ID is null")
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}

	var body interfaces.CifsUserGroupPrivilegeResourceBodyDataModelONTAP

	// reset privileges to empty list
	body.Privileges = []string{}
	_, err = interfaces.UpdateCifsUserGroupPrivilege(errorHandler, *client, body, svm.UUID, data.Name.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *CifsUserGroupPrivilegeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a protocols cifs user group resource: %#v", req))
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
}
