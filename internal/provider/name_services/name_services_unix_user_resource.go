package name_services

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
var _ resource.Resource = &NameServicesUnixUserResource{}
var _ resource.ResourceWithImportState = &NameServicesUnixUserResource{}

// NewNameServicesUnixUserResource is a helper function to simplify the provider implementation.
func NewNameServicesUnixUserResource() resource.Resource {
	return &NameServicesUnixUserResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "unix_user",
		},
	}
}

// NameServicesUnixUserResource defines the resource implementation.
type NameServicesUnixUserResource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesUnixUserResourceModel describes the resource data model.
type NameServicesUnixUserResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Name          types.String `tfsdk:"name"`
	FullName      types.String `tfsdk:"full_name"`
	PrimaryGID    types.Int64  `tfsdk:"primary_gid"`
	ID            types.Int64  `tfsdk:"id"`
}

// Metadata returns the resource type name
func (r *NameServicesUnixUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *NameServicesUnixUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesUnixUser resource",
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
				MarkdownDescription: "Specifies the name of the UNIX user.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "Full name of the UNIX user.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"primary_gid": schema.Int64Attribute{
				MarkdownDescription: "Primary group ID to which the UNIX user belongs.",
				Required: true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "UNIX user ID of the specified user.",
				Required: true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *NameServicesUnixUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *NameServicesUnixUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NameServicesUnixUserResourceModel

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

	var restInfo *interfaces.NameServicesUnixUserGetDataModelONTAP
	restInfo, err = interfaces.GetNameServicesUnixUser(errorHandler, *client, data.SVMName.ValueString(), data.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetNameServicesUnixUser
		return
	}

	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.FullName = types.StringValue(restInfo.FullName)
	data.PrimaryGID = types.Int64Value(restInfo.PrimaryGID)
	data.ID = types.Int64Value(restInfo.ID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve ID
func (r *NameServicesUnixUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NameServicesUnixUserResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.NameServicesUnixUserResourceBodyDataModel
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

	nameValue := data.Name.ValueString()
	body.Name = &nameValue
	body.SVM.Name = data.SVMName.ValueString()
	if !data.FullName.IsNull() && !data.FullName.IsUnknown() && data.FullName.ValueString() != "" {
		fullNameValue := data.FullName.ValueString()
		body.FullName = &fullNameValue
	}
	primaryGIDValue := data.PrimaryGID.ValueInt64()
	body.PrimaryGID = &primaryGIDValue
	idValue := data.ID.ValueInt64()
	body.ID = &idValue

	resource, err := interfaces.CreateNameServicesUnixUser(errorHandler, *client, body)
	if err != nil {
		// error reporting done inside CreateNameServicesUnixUser
		return
	}
	
	// Update the computed parameters
	data.SVMName = types.StringValue(resource.SVM.Name)
	data.Name = types.StringValue(resource.Name)
	data.FullName = types.StringValue(resource.FullName)
	data.PrimaryGID = types.Int64Value(resource.PrimaryGID)
	data.ID = types.Int64Value(resource.ID)

	tflog.Trace(ctx, fmt.Sprintf("created UNIX user resource, Name=%s", data.Name.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *NameServicesUnixUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config NameServicesUnixUserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Read Terraform config data into the model (includes write-only attributes)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
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
	var body interfaces.NameServicesUnixUserResourceBodyDataModel
	body.Name = nil
	if !plan.FullName.Equal(state.FullName) {
		fullNameValue := plan.FullName.ValueString()
		body.FullName = &fullNameValue
	}
	if !plan.PrimaryGID.Equal(state.PrimaryGID) {
		primaryGIDValue := plan.PrimaryGID.ValueInt64()
		body.PrimaryGID = &primaryGIDValue
	}
	if !plan.ID.Equal(state.ID) {
		idValue := plan.ID.ValueInt64()
		body.ID = &idValue
	}
	
	err = interfaces.UpdateNameServicesUnixUser(errorHandler, *client, body, svm.UUID, state.Name.ValueString())
	if err != nil {
		// error reporting done inside UpdateNameServicesUnixUser
		return
	}

	restInfo, err := interfaces.GetNameServicesUnixUser(errorHandler, *client, plan.SVMName.ValueString(), plan.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetNameServicesUnixUser
		return
	}
	
	// Update the computed parameters
	plan.SVMName = types.StringValue(restInfo.SVM.Name)
	plan.Name = types.StringValue(restInfo.Name)
	plan.FullName = types.StringValue(restInfo.FullName)
	plan.PrimaryGID = types.Int64Value(restInfo.PrimaryGID)
	plan.ID = types.Int64Value(restInfo.ID)
	
	tflog.Debug(ctx, fmt.Sprintf("updated UNIX user resource: Name=%s", plan.Name.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *NameServicesUnixUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data NameServicesUnixUserResourceModel

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
		errorHandler.MakeAndReportError("Name is null", "UNIX user name is null")
		return
	}

	err = interfaces.DeleteNameServicesUnixUser(errorHandler, *client, svm.UUID, data.Name.ValueString())
	if err != nil {
		// error reporting done inside DeleteNameServicesUnixUser
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *NameServicesUnixUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
