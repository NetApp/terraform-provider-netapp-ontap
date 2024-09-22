package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &VolumesFilesResource{}
var _ resource.ResourceWithImportState = &VolumesFilesResource{}

// NewVolumesFilesResource is a helper function to simplify the provider implementation.
func NewVolumesFileResource() resource.Resource {
	return &VolumesFilesResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "volumes_file",
		},
	}
}

// VolumesFilesResource defines the resource implementation.
type VolumesFilesResource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageVolumesFileDataSourceModel describes the data source data model.
type StorageVolumesFileResourceModel struct {
	CxProfileName   types.String `tfsdk:"cx_profile_name"`
	VolumeName      types.String `tfsdk:"volume_name"`
	SVMName         types.String `tfsdk:"svm_name"`
	Path            types.String `tfsdk:"path"`
	ByteOffset      types.Int64  `tfsdk:"byte_offset"`
	Overwrite       types.Bool   `tfsdk:"overwrite"`
	UnixPermissions types.Int64  `tfsdk:"unix_permissions"`
	Name            types.String `tfsdk:"name"`
	Type            types.String `tfsdk:"type"`
	Size            types.Int64  `tfsdk:"size"`
	ID              types.String `tfsdk:"id"`
}

// QOSPolicy describes the efficiency model.
type QOSPolicy struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the resource type name.
func (r *VolumesFilesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *VolumesFilesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "VolumesFiles resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"volume_name": schema.StringAttribute{
				MarkdownDescription: "Volume name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the svm to use",
				Required:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Relative path of a file or directory in the volume",
				Required:            true,
			},
			"byte_offset": schema.Int64Attribute{
				MarkdownDescription: "The number of bytes used",
				Optional:            true,
			},
			"overwrite": schema.BoolAttribute{
				MarkdownDescription: "Whether the file can be overwritten",
				Optional:            true,
			},
			"unix_permissions": schema.Int64Attribute{
				MarkdownDescription: "UNIX permissions to be viewed as an octal number",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the file or directory",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the file or directory",
				Optional:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the file or directory",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "VolumesFiles path is used as ID here",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *VolumesFilesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *VolumesFilesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StorageVolumesFileResourceModel

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

	restVolInfo, err := interfaces.GetStorageVolumeByName(errorHandler, *client, data.VolumeName.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetStorageVolumeByName
		return
	}
	if restVolInfo == nil {
		errorHandler.MakeAndReportError("No volume found", fmt.Sprintf("volume %s not found.", data.VolumeName))
		return
	}

	restInfo, err := interfaces.GetStorageVolumesFiles(errorHandler, *client, restVolInfo.UUID, data.Path.ValueString())
	if err != nil {
		// error reporting done inside GetVolumesFiles
		return
	}

	if restInfo == nil {
		errorHandler.MakeAndReportError("error reading info", "No volume files found")
		return
	}

	pathFound := false
	for _, record := range restInfo {
		if record.Path == data.Path.ValueString() {
			data.Type = types.StringValue(record.Type)
			data.ID = types.StringValue(record.Path)
			data.Size = types.Int64Value(int64(record.Size))
			pathFound = true
			break
		}
	}

	if !pathFound {
		errorHandler.MakeAndReportError("error reading info", "No data found")
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *VolumesFilesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageVolumesFileResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var request interfaces.VolumesFilesResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	volume, err := interfaces.GetStorageVolumeByName(errorHandler, *client, data.VolumeName.ValueString(), data.SVMName.ValueString())
	if err != nil {
		return
	}
	if volume == nil {
		errorHandler.MakeAndReportError("No volume found", fmt.Sprintf("volume %s not found.", data.VolumeName))
		return
	}

	request.Path = data.Path.ValueString()
	if !data.ByteOffset.IsNull() {
		request.ByteOffset = data.ByteOffset.ValueInt64()
	}
	if !data.Overwrite.IsNull() {
		request.Overwrite = data.Overwrite.ValueBool()
	}
	if !data.Name.IsNull() {
		request.Name = data.Name.ValueString()
	}
	if !data.Type.IsNull() {
		request.Type = data.Type.ValueString()
	}
	if !data.UnixPermissions.IsNull() {
		request.UnixPermissions = data.UnixPermissions.ValueInt64()
	}

	_, err = interfaces.CreateVolumesFiles(errorHandler, *client, request, volume.UUID)
	if err != nil {
		return
	}

	restInfoGet, err := interfaces.GetStorageVolumesFiles(errorHandler, *client, volume.UUID, data.Path.ValueString())
	if err != nil {
		// error reporting done inside GetVolumesFiles
		return
	}
	if restInfoGet == nil {
		errorHandler.MakeAndReportError("error reading info", "No volume files found")
		return
	}

	pathFound := false
	for _, record := range restInfoGet {
		if record.Path == data.Path.ValueString() {
			data.ID = types.StringValue(record.Path)
			data.Type = types.StringValue(record.Type)
			data.ID = types.StringValue(record.Path)
			data.Size = types.Int64Value(int64(record.Size))
			pathFound = true
			break
		}
	}

	if !pathFound {
		errorHandler.MakeAndReportError("error reading info", "No data found")
		return
	}

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *VolumesFilesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *StorageVolumesFileResourceModel
	var state *StorageVolumesFileResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	volume, err := interfaces.GetStorageVolumeByName(errorHandler, *client, plan.VolumeName.ValueString(), plan.SVMName.ValueString())
	if err != nil {
		return
	}
	if volume == nil {
		errorHandler.MakeAndReportError("No volume found", fmt.Sprintf("volume %s not found.", plan.VolumeName))
		return
	}
	var request interfaces.VolumesFilesResourceBodyDataModelONTAP
	if !plan.Path.Equal(state.Path) {
		request.Path = plan.Path.ValueString()
	}
	tflog.Debug(ctx, fmt.Sprintf("update a resource %s: %#v", state.ID.ValueString(), request))
	err = interfaces.UpdateVolumesFiles(errorHandler, *client, request, volume.UUID, state.ID.ValueString())
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *VolumesFilesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *StorageVolumesFileResourceModel

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
		errorHandler.MakeAndReportError("UUID is null", "volumes_files UUID is null")
		return
	}

	volume, err := interfaces.GetStorageVolumeByName(errorHandler, *client, data.VolumeName.ValueString(), data.SVMName.ValueString())
	if err != nil {
		return
	}
	if volume == nil {
		errorHandler.MakeAndReportError("No volume found", fmt.Sprintf("volume %s not found.", data.VolumeName))
		return
	}

	err = interfaces.DeleteVolumesFiles(errorHandler, *client, data.ID.ValueString(), volume.UUID)
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *VolumesFilesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: volume_name,svm_name,path,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("volume_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("path"), idParts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[3])...)
}
