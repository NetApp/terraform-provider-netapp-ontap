package provider

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &CifsLocalUserResource{}
var _ resource.ResourceWithImportState = &CifsLocalUserResource{}

// NewCifsLocalUserResource is a helper function to simplify the provider implementation.
func NewCifsLocalUserResource() resource.Resource {
	return &CifsLocalUserResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_cifs_local_user_resource",
		},
	}
}

// CifsLocalUserResource defines the resource implementation.
type CifsLocalUserResource struct {
	config connection.ResourceOrDataSourceConfig
}

// UserMember describes the data source data model.
type UserMember struct {
	Name types.String `tfsdk:"name"`
}

// CifsLocalUserResourceModel describes the resource data model.
type CifsLocalUserResourceModel struct {
	CxProfileName   types.String `tfsdk:"cx_profile_name"`
	Name            types.String `tfsdk:"name"`
	SVMName         types.String `tfsdk:"svm_name"`
	Password        types.String `tfsdk:"password"`
	ID              types.String `tfsdk:"id"`
	Description     types.String `tfsdk:"description"`
	FullName        types.String `tfsdk:"full_name"`
	Membership      types.Set    `tfsdk:"membership"`
	AccountDisabled types.Bool   `tfsdk:"account_disabled"`
}

// Metadata returns the resource type name.
func (r *CifsLocalUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *CifsLocalUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsLocalUser resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "CifsLocalUser name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "CifsLocalUser svm name",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "CifsLocalUser password",
				Required:            true,
				Sensitive:           true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "CifsLocalUser description",
				Optional:            true,
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "CifsLocalUser full name",
				Optional:            true,
			},
			"membership": schema.SetNestedAttribute{
				Computed:            true,
				MarkdownDescription: "CifsLocalUser membership",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "CifsLocalUser membership name",
						},
					},
				},
			},
			"account_disabled": schema.BoolAttribute{
				Computed:            true,
				Optional:            true,
				Default:             booldefault.StaticBool(false),
				MarkdownDescription: "CifsLocalUser account disabled",
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "CifsLocalUser SID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *CifsLocalUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// attrTypes returns a map of the attribute types for the resource.
func (o UserMember) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
	}
}

// membershipSliceToSet converts a slice of UserMember to a types.Set
func membershipSliceToSet(ctx context.Context, membersSliceIn []interfaces.Membership, diags *diag.Diagnostics) types.Set {
	members := make([]UserMember, len(membersSliceIn))
	for i, member := range membersSliceIn {
		members[i].Name = types.StringValue(member.Name)
	}

	keys, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: UserMember{}.attrTypes()}, members)
	diags.Append(d...)

	return keys
}

// Read refreshes the Terraform state with the latest data.
func (r *CifsLocalUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CifsLocalUserResourceModel

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

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}

	var restInfo *interfaces.CifsLocalUserGetDataModelONTAP
	if data.ID.IsNull() {
		restInfo, err = interfaces.GetCifsLocalUserByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
		if restInfo == nil || err != nil {
			// error reporting done inside GetCifsLocalUser
			return
		}
		data.ID = types.StringValue(restInfo.SID)
	} else {
		// we have the UUID, so we can use it to get the record
		restInfo, err = interfaces.GetCifsLocalUser(errorHandler, *client, svm.UUID, data.ID.ValueString())
	}

	if err != nil {
		// error reporting done inside GetCifsLocalUser
		return
	}

	data.SVMName = types.StringValue(svm.Name)
	data.Membership = membershipSliceToSet(ctx, restInfo.Membership, &resp.Diagnostics)

	// name in formate xxx\xxx from GET record
	if !strings.Contains(restInfo.Name, "\\") {
		errorHandler.MakeAndReportError("invalid name", fmt.Sprintf("protocols_cifs_local_user name %s is invalid", restInfo.Name))
		return
	}
	// check if name is in the form of xxx\xxx
	if !strings.Contains(data.Name.ValueString(), "\\") {
		name := strings.Split(restInfo.Name, "\\")[1]
		data.Name = types.StringValue(name)
	} else {
		data.Name = types.StringValue(restInfo.Name)
	}

	if restInfo.FullName != "" {
		data.FullName = types.StringValue(restInfo.FullName)
	}
	if restInfo.Description != "" {
		data.Description = types.StringValue(restInfo.Description)
	}
	data.AccountDisabled = types.BoolValue(restInfo.AccountDisabled)
	// password is not returned by GET

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *CifsLocalUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *CifsLocalUserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.CifsLocalUserResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()
	body.Password = data.Password.ValueString()
	if !data.Description.IsNull() {
		body.Description = data.Description.ValueString()
	}
	if !data.FullName.IsNull() {
		body.FullName = data.FullName.ValueString()
	}
	if !data.AccountDisabled.IsNull() {
		body.AccountDisabled = data.AccountDisabled.ValueBool()
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateCifsLocalUser(errorHandler, *client, body)
	if err != nil {
		return
	}

	// read the resource back to get the ID cause create API does not return the record
	restInfo, err := interfaces.GetCifsLocalUserByName(errorHandler, *client, resource.Name, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetCifsLocalGroup
		return
	}
	data.ID = types.StringValue(restInfo.SID)
	data.Membership = membershipSliceToSet(ctx, restInfo.Membership, &resp.Diagnostics)

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *CifsLocalUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *CifsLocalUserResourceModel
	var dataOld *CifsLocalUserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &dataOld)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	var request interfaces.CifsLocalUserResourceBodyDataModelONTAP
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}

	// only the changed values are sent in the request
	// svm.name cannot be set in update request
	// update memgbership is not supported
	if !data.Name.Equal(dataOld.Name) {
		request.Name = data.Name.ValueString()
	}
	if !data.Password.Equal(dataOld.Password) {
		request.Password = data.Password.ValueString()
	}
	if !data.Description.Equal(dataOld.Description) {
		request.Description = data.Description.ValueString()
	}
	if !data.FullName.Equal(dataOld.FullName) {
		request.FullName = data.FullName.ValueString()
	}
	if !data.AccountDisabled.Equal(dataOld.AccountDisabled) {
		request.AccountDisabled = data.AccountDisabled.ValueBool()
	}

	_, err = interfaces.UpdateCifsLocalUser(errorHandler, *client, request, svm.UUID, data.ID.ValueString())
	if err != nil {
		return
	}

	// read the resource back to get the ID cause create API does not return the record
	restInfo, err := interfaces.GetCifsLocalUserByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetCifsLocalGroup
		return
	}
	data.Membership = membershipSliceToSet(ctx, restInfo.Membership, &resp.Diagnostics)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *CifsLocalUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *CifsLocalUserResourceModel

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

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "protocols_cifs_local_user UUID is null")
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}

	err = interfaces.DeleteCifsLocalUser(errorHandler, *client, svm.UUID, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *CifsLocalUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a protocols cifs local user resource: %#v", req))
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
