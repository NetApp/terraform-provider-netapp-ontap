package provider

import (
	"context"
	"fmt"
	"strings"

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

// TODO:
// copy this file to match you resource (should match internal/provider/protocols_cifs_user_group_privilege_resource.go)
// replace CifsUserGroupPrivilege with the name of the resource, following go conventions, eg IPInterface
// replace protocols_cifs_user_group_privilege with the name of the resource, for logging purposes, eg ip_interface
// make sure to create internal/interfaces/protocols_cifs_user_group_privilege.go too)
// delete these 5 lines

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
			"privileges": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of privileges",
				Required:            true,
				PlanModifiers: []planmodifier.List{
					ListConvertToLowercasePlanModifier(),
				},
				// Validators: []validator.List{
				// 	// Validate this List must contain string values which are not sensitive.
				// 	listvalidator.ValueStringsAre([]string{"SeSecurityPrivilege", "SeTakeOwnershipPrivilege", "SeChangeNotifyPrivilege"}),
				// },
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

func ListConvertToLowercasePlanModifier() planmodifier.List {
	return &listConvertToLowercasePlanModifier{}
}

type listConvertToLowercasePlanModifier struct {
}

func (d *listConvertToLowercasePlanModifier) Description(ctx context.Context) string {
	return "convert list of strings to lowercase"
}

func (d *listConvertToLowercasePlanModifier) MarkdownDescription(ctx context.Context) string {
	return d.Description(ctx)
}

func (d *listConvertToLowercasePlanModifier) PlanModifyList(ctx context.Context, req planmodifier.ListRequest, resp *planmodifier.ListResponse) {
	var attr, attrstate types.List

	diags := req.Plan.GetAttribute(ctx, path.Root("privileges"), &attr)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.GetAttribute(ctx, path.Root("privileges"), &attrstate)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("*********read a resource plan: %#v", attr))
	tflog.Debug(ctx, fmt.Sprintf("*********read a resource state: %#v", attrstate))
	// basetypes.ListValue{elements:[]attr.Value{basetypes.StringValue{state:0x2, value:"SeSecurityPrivilege"}, basetypes.StringValue{state:0x2, value:"SeTakeOwnershipPrivilege"}, basetypes.StringValue{state:0x2, value:"SeChangeNotifyPrivilege"}}, elementType:basetypes.StringType{}, state:0x2}
	planValues := attr.Elements()
	stateValues := attrstate.Elements()
	for i, planValue := range planValues {
		// Convert the value to lowercase for case-insensitive comparison
		lv := strings.ToLower(planValue.(types.String).String())
		lowerValue := types.StringValue(lv)
		tflog.Debug(ctx, fmt.Sprintf("\t*** plan: %#v lv: %#v state: %#v", lowerValue, lv, stateValues[i]))
		if stateValues[i].Equal(lowerValue) {
			// If the plan value is equal to the state value, then we don't need to
			// modify the plan.
			tflog.Debug(ctx, fmt.Sprintf("\t*** inside plan: %#v state: %#v", lowerValue, stateValues[i]))
			continue
		}
		// theValue := types.StringValue(planValue.String())
		// lowerValue := types.String(strings.ToLower(theValue.ToTerraformValue()))
		planValues[i] = lowerValue
	}
	tflog.Debug(ctx, fmt.Sprintf("*********out side plan change: %#v", planValues))
	req.Plan.SetAttribute(ctx, path.Root("privileges"), planValues)

	// resp.Diagnostics.Append(planValues)
	// if req.PlanValue.Equal(req.StateValue) {
	// 	// If the plan value is equal to the state value, then we don't need to
	// 	// modify the plan.
	// 	return
	// }
	//resp.SetAttribute(ctx, path.Root("privileges"), planValues)
	// resp.Diagnostics.Append(resp.PlanValue.Elements(), planValues...)
	// resp.Diagnostics.Append(resp.PlanValue.SetAttribute(ctx, path.Root("privileges"), planValues))

	tflog.Debug(ctx, fmt.Sprintf("*********plan: %#v", planValues))
	// Convert the value to lowercase for case-insensitive comparison
	// for _, value := range attr.Value() {
	// 	// Convert the value to lowercase for case-insensitive comparison
	// 	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", value))
	// 	// lowerValue := strings.ToLower(value.(string))
	// 	// updatedList.Append(lowerValue)
	// }
	//resp.Diagnostics.Append(resp.PlanValue.Elements()(ctx, path.Root("privileges"), req.PlanValue.Elements())...)
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
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
