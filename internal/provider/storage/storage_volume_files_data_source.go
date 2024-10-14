package storage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageVolumesFilesDataSource{}

// NewStorageVolumesFilesDataSource is a helper function to simplify the provider implementation.
func NewStorageVolumesFilesDataSource() datasource.DataSource {
	return &StorageVolumesFilesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "volumes_files",
		},
	}
}

// StorageVolumesFilesDataSource defines the data source implementation.
type StorageVolumesFilesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageVolumesFileDataSourceModel describes the data source data model.
type StorageVolumesFileDataSourceModel struct {
	CxProfileName    types.String `tfsdk:"cx_profile_name"`
	VolumeName       types.String `tfsdk:"volume_name"`
	Path             types.String `tfsdk:"path"`
	BytesUsed        types.Int64  `tfsdk:"bytes_used"`
	Name             types.String `tfsdk:"name"`
	OverwriteEnabled types.Bool   `tfsdk:"overwrite_enabled"`
	Type             types.String `tfsdk:"type"`
	GroupID          types.Int64  `tfsdk:"group_id"`
	HardLinksCount   types.Int64  `tfsdk:"hard_links_count"`
	Size             types.Int64  `tfsdk:"size"`
	OwnerID          types.Int64  `tfsdk:"owner_id"`
	InodeNumber      types.Int64  `tfsdk:"inode_number"`
	IsEmpty          types.Bool   `tfsdk:"is_empty"`
	Target           types.String `tfsdk:"target"`
}

// StorageVolumesFilesDataSourceModel describes the data source data model.
type StorageVolumesFilesDataSourceModel struct {
	StorageVolumesFiles []StorageVolumesFileDataSourceModel `tfsdk:"storage_volumes_files"`
	CxProfileName       types.String                        `tfsdk:"cx_profile_name"`
	VolumeName          types.String                        `tfsdk:"volume_name"`
	SVMName             types.String                        `tfsdk:"svm_name"`
	Path                types.String                        `tfsdk:"path"`
	ByteOffset          types.Int64                         `tfsdk:"byte_offset"`
	Name                types.String                        `tfsdk:"name"`
	OverwriteEnabled    types.Bool                          `tfsdk:"overwrite_enabled"`
}

// Metadata returns the data source type name.
func (d *StorageVolumesFilesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *StorageVolumesFilesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "StorageVolumesFiles data source",
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
				MarkdownDescription: "svm name",
				Required:            true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: "Relative path of a file or directory in the volume",
				Required:            true,
			},
			"byte_offset": schema.Int64Attribute{
				MarkdownDescription: "The file offset to start reading from",
				Optional:            true,
			},
			"overwrite_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the file can be overwritten",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the file or directory",
				Optional:            true,
			},
			"storage_volumes_files": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"volume_name": schema.StringAttribute{
							MarkdownDescription: "Volume name",
							Computed:            true,
						},
						"path": schema.StringAttribute{
							MarkdownDescription: "Relative path of a file or directory in the volume",
							Computed:            true,
						},
						"bytes_used": schema.Int64Attribute{
							MarkdownDescription: "The number of bytes used",
							Computed:            true,
						},
						"overwrite_enabled": schema.BoolAttribute{
							MarkdownDescription: "Whether the file can be overwritten",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the file or directory",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the file or directory",
							Computed:            true,
						},
						"group_id": schema.Int64Attribute{
							MarkdownDescription: "The group ID of the file or directory",
							Computed:            true,
						},
						"hard_links_count": schema.Int64Attribute{
							MarkdownDescription: "The number of hard links to the file or directory",
							Computed:            true,
						},
						"size": schema.Int64Attribute{
							MarkdownDescription: "The size of the file or directory",
							Computed:            true,
						},
						"owner_id": schema.Int64Attribute{
							MarkdownDescription: "The owner ID of the file or directory",
							Computed:            true,
						},
						"inode_number": schema.Int64Attribute{
							MarkdownDescription: "The inode number of the file or directory",
							Computed:            true,
						},
						"is_empty": schema.BoolAttribute{
							MarkdownDescription: "Whether the file or directory is empty",
							Computed:            true,
						},
						"target": schema.StringAttribute{
							MarkdownDescription: "Whether the file or directory is empty",
							Computed:            true,
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *StorageVolumesFilesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	d.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *StorageVolumesFilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageVolumesFilesDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	restInfo, err := interfaces.GetStorageVolumeByName(errorHandler, *client, data.VolumeName.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetStorageVolumeByName
		return
	}

	storageVolumesFiles, err := interfaces.GetStorageVolumesFiles(errorHandler, *client, restInfo.UUID, data.Path.ValueString())
	if err != nil {
		// error reporting done inside GetStorageVolumesFiles
		return
	}

	data.StorageVolumesFiles = make([]StorageVolumesFileDataSourceModel, len(storageVolumesFiles))
	for index, record := range storageVolumesFiles {
		data.StorageVolumesFiles[index] = StorageVolumesFileDataSourceModel{
			CxProfileName:    types.String(data.CxProfileName),
			Name:             types.StringValue(record.Name),
			VolumeName:       types.StringValue(record.Volume.Name),
			Path:             types.StringValue(record.Path),
			OverwriteEnabled: types.BoolValue(record.OverwriteEnabled),
			Type:             types.StringValue(record.Type),
			GroupID:          types.Int64Value(int64(record.GroupID)),
			HardLinksCount:   types.Int64Value(int64(record.HardLinksCount)),
			BytesUsed:        types.Int64Value(int64(record.BytesUsed)),
			Size:             types.Int64Value(int64(record.Size)),
			OwnerID:          types.Int64Value(int64(record.OwnerID)),
			InodeNumber:      types.Int64Value(int64(record.InodeNumber)),
			IsEmpty:          types.BoolValue(record.IsEmpty),
			Target:           types.StringValue(record.Target),
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
