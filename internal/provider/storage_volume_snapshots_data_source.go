package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageVolumeSnapshotsDataSource{}

// NewStorageVolumeSnapshotsDataSource is a helper function to simplify the provider implementation.
func NewStorageVolumeSnapshotsDataSource() datasource.DataSource {
	return &StorageVolumeSnapshotsDataSource{
		config: resourceOrDataSourceConfig{
			name: "storage_volume_snapshots_data_source",
		},
	}
}

// StorageVolumeSnapshotsDataSource defines the data source implementation.
type StorageVolumeSnapshotsDataSource struct {
	config resourceOrDataSourceConfig
}

// StorageVolumeSnapshotsDataSourceModel describes the data source data model.
type StorageVolumeSnapshotsDataSourceModel struct {
	CxProfileName          types.String                                `tfsdk:"cx_profile_name"`
	StorageVolumeSnapshots []StorageVolumeSnapshotDataSourceModel      `tfsdk:"storage_volume_snapshots"`
	Filter                 *StorageVolumeSnapshotDataSourceFilterModel `tfsdk:"filter"`
}

// StorageVolumeSnapshotDataSourceFilterModel describes the data source data model for queries.
type StorageVolumeSnapshotDataSourceFilterModel struct {
	Name       types.String `tfsdk:"name"`
	SVMName    types.String `tfsdk:"svm_name"`
	VolumeName types.String `tfsdk:"volume_name"`
}

// Metadata returns the data source type name.
func (d *StorageVolumeSnapshotsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *StorageVolumeSnapshotsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageVolumeSnapshots data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "StorageVolumeSnapshot name",
						Required:            true,
					},
					"volume_name": schema.StringAttribute{
						MarkdownDescription: "StorageVolumeSnapshot volume name",
						Required:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "StorageVolumeSnapshot svm name",
						Required:            true,
					},
				},
				Required: true,
			},
			"storage_volume_snapshots": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "StorageVolumeSnapshot name",
							Required:            true,
						},
						"volume_name": schema.StringAttribute{
							MarkdownDescription: "Volume Name",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "SVM Name",
							Required:            true,
						},
						"create_time": schema.StringAttribute{
							MarkdownDescription: "Create time",
							Computed:            true,
						},
						"expiry_time": schema.StringAttribute{
							MarkdownDescription: "Expiry time",
							Computed:            true,
						},
						"state": schema.StringAttribute{
							MarkdownDescription: "State",
							Computed:            true,
						},
						"size": schema.Float64Attribute{
							MarkdownDescription: "Size",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Comment",
							Computed:            true,
						},
						"snapmirror_label": schema.StringAttribute{
							MarkdownDescription: "Snapmirror Label",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "volume snapshot UUID",
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
func (d *StorageVolumeSnapshotsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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
	d.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *StorageVolumeSnapshotsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageVolumeSnapshotsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.Filter.Name.IsNull() {
		errorHandler.MakeAndReportError("error reading snapshot", "filter.name is required")
		return
	}
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.Filter.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetStorageVolumeSnapshots
		return
	}
	volume, err := interfaces.GetStorageVolumeByName(errorHandler, *client, data.Filter.VolumeName.ValueString(), svm.Name)
	if err != nil {
		// error reporting done inside GetStorageVolumeSnapshots
		return
	}

	var filter *interfaces.StorageVolumeSnapshotDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.StorageVolumeSnapshotDataSourceFilterModel{
			Name: data.Filter.Name.ValueString(),
		}
	}

	restInfo, err := interfaces.GetListStorageVolumeSnapshots(errorHandler, *client, volume.UUID, filter)
	if err != nil {
		// error reporting done inside GetStorageVolumeSnapshots
		return
	}

	data.StorageVolumeSnapshots = make([]StorageVolumeSnapshotDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.StorageVolumeSnapshots[index] = StorageVolumeSnapshotDataSourceModel{
			CxProfileName:   types.String(data.CxProfileName),
			Name:            types.StringValue(record.Name),
			SVMName:         types.StringValue(record.SVM.Name),
			CreateTime:      types.StringValue(record.CreateTime),
			ExpiryTime:      types.StringValue(record.ExpiryTime),
			Size:            types.Float64Value(record.Size),
			SnapmirrorLabel: types.StringValue(record.SnapmirrorLabel),
			State:           types.StringValue(record.State),
			VolumeName:      types.StringValue(record.Volume.Name),
			ID:              types.StringValue(record.UUID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
