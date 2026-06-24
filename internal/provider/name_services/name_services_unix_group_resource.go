package name_services

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &NameServicesUnixGroupResource{}
var _ resource.ResourceWithImportState = &NameServicesUnixGroupResource{}

// NewNameServicesUnixGroupResource is a helper function to simplify the provider implementation.
func NewNameServicesUnixGroupResource() resource.Resource {
	return &NameServicesUnixGroupResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "unix_group",
		},
	}
}

// NameServicesUnixGroupResource defines the resource implementation.
type NameServicesUnixGroupResource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesUnixGroupResourceModel describes the resource data model.
type NameServicesUnixGroupResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Name          types.String `tfsdk:"name"`
	Users         types.Set    `tfsdk:"users"`
	ID            types.Int64  `tfsdk:"id"`
}

// Metadata returns the resource type name
func (r *NameServicesUnixGroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *NameServicesUnixGroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesUnixGroup resource",
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
				MarkdownDescription: "Specifies the name of the UNIX group.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"users": schema.SetAttribute{
				MarkdownDescription: "Users belonging to the UNIX group.",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "UNIX group ID of the specified group.",
				Required:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *NameServicesUnixGroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *NameServicesUnixGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NameServicesUnixGroupResourceModel

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

	var restInfo *interfaces.NameServicesUnixGroupGetDataModelONTAP
	restInfo, err = interfaces.GetNameServicesUnixGroup(errorHandler, *client, data.SVMName.ValueString(), data.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetNameServicesUnixGroup
		return
	}

	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.ID = types.Int64Value(restInfo.ID)

	// Users - convert to types.Set
	users := make([]string, 0, len(restInfo.Users))
	for _, user := range restInfo.Users {
		users = append(users, user.Name)
	}
	usersSet, diags := types.SetValueFrom(ctx, types.StringType, users)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Users = usersSet

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve ID
func (r *NameServicesUnixGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NameServicesUnixGroupResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.NameServicesUnixGroupResourceBodyDataModel
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

	// create the UNIX group first,
	// then add the UNIX users separately since the API does not allow setting all parameters at creation
	nameValue := data.Name.ValueString()
	body.Name = &nameValue
	body.SVM.Name = data.SVMName.ValueString()
	idValue := data.ID.ValueInt64()
	body.ID = &idValue

	resource, err := interfaces.CreateNameServicesUnixGroup(errorHandler, *client, body)
	if err != nil {
		// error reporting done inside CreateNameServicesUnixGroup
		return
	}
	
	// Update the computed parameters
	data.SVMName = types.StringValue(resource.SVM.Name)
	data.Name = types.StringValue(resource.Name)
	data.ID = types.Int64Value(resource.ID)

	tflog.Trace(ctx, fmt.Sprintf("created UNIX group resource, Name=%s", data.Name.ValueString()))

	// now add UNIX users to the group using the other API
	var usersToAdd []interfaces.UnixGroupUserRecord
	if !data.Users.IsNull() && !data.Users.IsUnknown() {
		userElements := make([]types.String, 0, len(data.Users.Elements()))
		diags := data.Users.ElementsAs(ctx, &userElements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, userName := range userElements {
			usersToAdd = append(usersToAdd, interfaces.UnixGroupUserRecord{Name: userName.ValueString()})
		}
	}

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	if len(usersToAdd) > 0 {
		addUsersRequest := interfaces.UnixGroupUsersResourceBodyDataModel{Records: usersToAdd}
		err = interfaces.AddUnixGroupUsers(errorHandler, *client, addUsersRequest, svm.UUID, data.Name.ValueString())
		if err != nil {
			// error reporting done inside AddUnixGroupUsers
			return
		}
	}

	restInfo, err := interfaces.GetNameServicesUnixGroup(errorHandler, *client, data.SVMName.ValueString(), data.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetNameServicesUnixGroup
		return
	}
	// Users - convert to types.Set
	users := make([]string, 0, len(restInfo.Users))
	for _, user := range restInfo.Users {
		users = append(users, user.Name)
	}
	usersSet, diags := types.SetValueFrom(ctx, types.StringType, users)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Users = usersSet

	tflog.Trace(ctx, "added UNIX users to UNIX group")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *NameServicesUnixGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state NameServicesUnixGroupResourceModel

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

	// Update UNIX group fields supported by the PATCH endpoint
	var body interfaces.NameServicesUnixGroupResourceBodyDataModel
	if !plan.ID.Equal(state.ID) {
		idValue := plan.ID.ValueInt64()
		body.ID = &idValue
		err = interfaces.UpdateNameServicesUnixGroup(errorHandler, *client, body, svm.UUID, state.Name.ValueString())
		if err != nil {
			// error reporting done inside UpdateNameServicesUnixGroup
			return
		}
	}

	// Reconcile users: add missing users; delete removed users
	if !plan.Users.IsNull() && !plan.Users.IsUnknown() && !state.Users.IsNull() && !state.Users.IsUnknown() {
		var desiredElems []types.String
		resp.Diagnostics.Append(plan.Users.ElementsAs(ctx, &desiredElems, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		desired := make(map[string]struct{}, len(desiredElems))
		for _, u := range desiredElems {
			name := strings.TrimSpace(u.ValueString())
			if name != "" {
				desired[name] = struct{}{}
			}
		}

		var currentElems []types.String
		resp.Diagnostics.Append(state.Users.ElementsAs(ctx, &currentElems, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		current := make(map[string]struct{}, len(currentElems))
		for _, u := range currentElems {
			name := strings.TrimSpace(u.ValueString())
			if name != "" {
				current[name] = struct{}{}
			}
		}

		// Delete users removed from config
		for userName := range current {
			if _, ok := desired[userName]; !ok {
				err = interfaces.DeleteUnixGroupUser(errorHandler, *client, svm.UUID, state.Name.ValueString(), userName)
				if err != nil {
					return
				}
			}
		}

		// Add newly desired users
		var usersToAdd []interfaces.UnixGroupUserRecord
		for userName := range desired {
			if _, ok := current[userName]; !ok {
				usersToAdd = append(usersToAdd, interfaces.UnixGroupUserRecord{Name: userName})
			}
		}
		if len(usersToAdd) > 0 {
			addUsersRequest := interfaces.UnixGroupUsersResourceBodyDataModel{Records: usersToAdd}
			err = interfaces.AddUnixGroupUsers(errorHandler, *client, addUsersRequest, svm.UUID, state.Name.ValueString())
			if err != nil {
				return
			}
		}
	}

	restInfo, err := interfaces.GetNameServicesUnixGroup(errorHandler, *client, plan.SVMName.ValueString(), plan.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetNameServicesUnixGroup
		return
	}
	
	// Update the computed parameters
	plan.SVMName = types.StringValue(restInfo.SVM.Name)
	plan.Name = types.StringValue(restInfo.Name)
	plan.ID = types.Int64Value(restInfo.ID)
	
	// users - map to set
	users := make([]string, 0, len(restInfo.Users))
	for _, user := range restInfo.Users {
		users = append(users, user.Name)
	}
	usersSet, diags := types.SetValueFrom(ctx, types.StringType, users)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Users = usersSet
	
	tflog.Debug(ctx, fmt.Sprintf("updated UNIX group resource: Name=%s", plan.Name.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *NameServicesUnixGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NameServicesUnixGroupResourceModel

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

	if data.Name.IsNull() || data.Name.IsUnknown() || strings.TrimSpace(data.Name.ValueString()) == "" {
		errorHandler.MakeAndReportError("Name is null", "UNIX group name is null")
		return
	}

	err = interfaces.DeleteNameServicesUnixGroup(errorHandler, *client, svm.UUID, data.Name.ValueString())
	if err != nil {
		// error reporting done inside DeleteNameServicesUnixGroup
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *NameServicesUnixGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
