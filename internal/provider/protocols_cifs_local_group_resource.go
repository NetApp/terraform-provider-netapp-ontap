package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
var _ resource.Resource = &CifsLocalGroupResource{}
var _ resource.ResourceWithImportState = &CifsLocalGroupResource{}

// NewCifsLocalGroupResource is a helper function to simplify the provider implementation.
func NewCifsLocalGroupResource() resource.Resource {
	return &CifsLocalGroupResource{
		config: resourceOrDataSourceConfig{
			name: "protocols_cifs_local_group_resource",
		},
	}
}

// CifsLocalGroupResource defines the resource implementation.
type CifsLocalGroupResource struct {
	config resourceOrDataSourceConfig
}

// GroupMember describes the data source data model.
type GroupMember struct {
	Name types.String `tfsdk:"name"`
}

// CifsLocalGroupResourceModel describes the resource data model.
type CifsLocalGroupResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	SVMName       types.String `tfsdk:"svm_name"`
	ID            types.String `tfsdk:"id"`
	Description   types.String `tfsdk:"description"`
	Members       types.Set    `tfsdk:"members"`
	// Members       []GroupMember `tfsdk:"members"`
	// Members []types.Object `tfsdk:"members"`
}

// Metadata returns the resource type name.
func (r *CifsLocalGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *CifsLocalGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsLocalGroup resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "CifsLocalGroup name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "CifsLocalGroup svm name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "CifsLocalGroup description",
				Optional:            true,
			},
			"members": schema.SetNestedAttribute{
				MarkdownDescription: "Cifs Local Group members",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Cifs Local Group member",
							Computed:            true,
							// Optional:            true,
						},
					},
				},
				// PlanModifiers: []planmodifier.Set{
				// 	setplanmodifier.UseStateForUnknown(),
				// },
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "CifsLocalGroup ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *CifsLocalGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *CifsLocalGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CifsLocalGroupResourceModel
	tflog.Debug(ctx, fmt.Sprintf("RRR read in get cifs local group resource: %#v", ctx))
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	tflog.Debug(ctx, fmt.Sprintf("$$$ read after get cifs local group resource: %#v", data.Name))
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

	var restInfo *interfaces.CifsLocalGroupGetDataModelONTAP
	if data.ID.IsNull() {
		restInfo, err = interfaces.GetCifsLocalGroupByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
		data.ID = types.StringValue(restInfo.SID)
	} else {
		restInfo, err = interfaces.GetCifsLocalGroup(errorHandler, *client, svm.UUID, data.ID.ValueString())
	}

	if err != nil {
		// error reporting done inside GetCifsLocalGroup
		return
	}

	// if !strings.Contains(data.Name.ValueString(), "\\") {
	// 	// rawName in the format as XXXX/name in read response
	// 	rawName := strings.Split(restInfo.Name, "\\")
	// 	data.Name = types.StringValue(rawName[len(rawName)-1])
	// }
	if restInfo.Description != "" {
		data.Description = types.StringValue(restInfo.Description)
	}
	// data.SVMName = types.StringValue(restInfo.SVM.Name)
	// if len(restInfo.Members) > 0 {

	// comment out the code for memebers type []GroupMember
	// data.Members = make([]GroupMember, len(restInfo.Members))
	// for i, member := range restInfo.Members {
	// 	data.Members[i].Name = types.StringValue(member.Name)
	// }

	// }
	data.Members = memberSliceToLSet(ctx, restInfo.Members, &resp.Diagnostics)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (o GroupMember) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name": types.StringType,
	}
}

func memberSliceToLSet(ctx context.Context, membersSliceIn []interfaces.Member, diags *diag.Diagnostics) types.Set {
	keys, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: GroupMember{}.attrTypes()}, membersSliceIn)
	diags.Append(d...)
	return keys
}

// Create a resource and retrieve UUID
func (r *CifsLocalGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *CifsLocalGroupResourceModel
	tflog.Debug(ctx, fmt.Sprintf("### create in get cifs local group resource: %#v", ctx))
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	tflog.Debug(ctx, fmt.Sprintf("### create after get cifs local group resource: %#v", data.Name))
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("### create start cifs local group resource: %#v", data))
	var body interfaces.CifsLocalGroupResourceBodyDataModelONTAP
	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()
	if !data.Description.IsNull() {
		body.Description = data.Description.ValueString()
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateCifsLocalGroup(errorHandler, *client, body)
	if err != nil {
		return
	}

	// read the resource back to get the ID cause create doesn't return it
	restInfo, err := interfaces.GetCifsLocalGroupByName(errorHandler, *client, resource.Name, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetCifsLocalGroup
		return
	}

	data.ID = types.StringValue(restInfo.SID)

	data.Members = memberSliceToLSet(ctx, restInfo.Members, &resp.Diagnostics)

	tflog.Debug(ctx, fmt.Sprintf("### create set ID cifs local group resource: %#v", data))
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	tflog.Debug(ctx, fmt.Sprintf("### create end cifs local group resource: %#v", data))
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *CifsLocalGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *CifsLocalGroupResourceModel
	var dataOld *CifsLocalGroupResourceModel

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
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}
	var request interfaces.CifsLocalGroupResourceBodyDataModelONTAP

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}

	// only the changed values are sent in the request
	// svm.name cannot be set in update request
	if !data.Name.Equal(dataOld.Name) {
		request.Name = data.Name.ValueString()
	}

	if !data.Description.Equal(dataOld.Description) {
		request.Description = data.Description.ValueString()
	}
	_, err = interfaces.UpdateCifsLocalGroup(errorHandler, *client, request, svm.UUID, data.ID.ValueString())
	if err != nil {
		return
	}

	// read the resource back to get the ID cause create doesn't return it
	restInfo, err := interfaces.GetCifsLocalGroupByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetCifsLocalGroup
		return
	}

	data.Members = memberSliceToLSet(ctx, restInfo.Members, &resp.Diagnostics)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *CifsLocalGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *CifsLocalGroupResourceModel

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
		errorHandler.MakeAndReportError("ID is null", "protocols_cifs_local_group ID is null")
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}

	err = interfaces.DeleteCifsLocalGroup(errorHandler, *client, svm.UUID, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *CifsLocalGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a protocols cifs local group resource: %#v", req))
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
