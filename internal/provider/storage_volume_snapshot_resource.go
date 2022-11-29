package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &StorageVolumeSnapshotResource{}
var _ resource.ResourceWithImportState = &StorageVolumeResource{}

// NewStorageVolumeSnapshotResource is a helper function to simplify the provider implementation.
func NewStorageVolumeSnapshotResource() resource.Resource {
	return &StorageVolumeSnapshotResource{
		config: resourceOrDataSourceConfig{
			name: "storage_volume_snapshot_resource",
		},
	}
}

// StorageVolumeSnapshotResource defines the resource implementation.
type StorageVolumeSnapshotResource struct {
	config resourceOrDataSourceConfig
}

// StorageVolumeSnapshotResourceModel describes the resource data model.
type StorageVolumeSnapshotResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	VolumeUUID    types.String `tfsdk:"volume_uuid"`
	UUID          types.String `tfsdk:"uuid"`
}

// Metadata returns the resource type name.
func (r *StorageVolumeSnapshotResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// GetSchema defines the schema for the resource.
func (r *StorageVolumeSnapshotResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Storage Volume Snapshot resource",

		Attributes: map[string]tfsdk.Attribute{
			"cx_profile_name": {
				MarkdownDescription: "Connection profile name",
				Type:                types.StringType,
				Required:            true,
			},
			"name": {
				MarkdownDescription: "Snapshot name",
				Type:                types.StringType,
				Required:            true,
			},
			// TODO: replace UUID with Volume Name, and vserver name
			"volume_uuid": {
				MarkdownDescription: "Volume UUID",
				Type:                types.StringType,
				Required:            true,
			},
			"uuid": {
				MarkdownDescription: "Snapshot UUID",
				Type:                types.StringType,
				Computed:            true,
			},
		},
	}, nil
}

// Configure adds the provider configured client to the data source.
func (r *StorageVolumeSnapshotResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
}

// Create creates the resource and sets the initial Terraform state.
func (r *StorageVolumeSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageVolumeSnapshotResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var request interfaces.StorageVolumeSnapshotResourceModel

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	request.Name = data.Name.ValueString()

	snapshot, err := interfaces.CreateStorageVolumeSnapshot(errorHandler, *client, request, data.VolumeUUID.ValueString())
	if err != nil {
		return
	}
	// TODO: add async calls or add wait condition for create
	data.UUID = types.StringValue(snapshot.UUID)

	tflog.Trace(ctx, "created a resource")
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *StorageVolumeSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *StorageVolumeSnapshotResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}
	snapshot, err := interfaces.GetStorageVolumeSnapshots(errorHandler, *client, data.Name.ValueString(), data.VolumeUUID.ValueString())
	if err != nil {
		return
	}
	data.Name = types.StringValue(snapshot.Name)
	data.VolumeUUID = types.StringValue(snapshot.Volume.UUID)
	data.UUID = types.StringValue(snapshot.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *StorageVolumeSnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *StorageVolumeSnapshotResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *StorageVolumeSnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *StorageVolumeSnapshotResourceModel
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
	_, err = interfaces.DeleteStorageVolumeSnapshot(errorHandler, *client, data.VolumeUUID.ValueString(), data.UUID.ValueString())
	if err != nil {
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *StorageVolumeSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
