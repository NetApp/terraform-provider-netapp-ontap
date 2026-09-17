package storage

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	
	"github.com/hashicorp/terraform-plugin-log/tflog"
	
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageVolumeClonesDataSource{}

// NewStorageVolumeClonesDataSource is a helper function to simplify the provider implementation.
func NewStorageVolumeClonesDataSource() datasource.DataSource {
	return &StorageVolumeClonesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "volume_clones",
		},
	}
}

// StorageVolumeClonesDataSource defines the data source implementation.
type StorageVolumeClonesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageVolumeClonesDataSourceModel describes the data source data model.
type StorageVolumeClonesDataSourceModel struct {
	CxProfileName       types.String                             `tfsdk:"cx_profile_name"`
	StorageVolumeClones []StorageVolumeCloneDataSourceModel      `tfsdk:"storage_volume_clones"`
	Filter              *StorageVolumeCloneDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *StorageVolumeClonesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *StorageVolumeClonesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageVolumeClones data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "The name of the SVM",
						Optional:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "The name of the volume clone",
						Optional:            true,
					},
				},
				Required: true,
			},
			"storage_volume_clones": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
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
							MarkdownDescription: "Specifies the name of the volume clone.",
							Required:            true,
						},
						"clone": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"is_flexclone": schema.BoolAttribute{
									MarkdownDescription: "Specifies if this volume is a normal FlexVol or FlexClone.",
									Computed:            true,
								},
								"split": schema.BoolAttribute{
									MarkdownDescription: "This field is set when split is executed on any FlexClone.",
									Computed:            true,
								},
								"parent_svm": schema.StringAttribute{
									MarkdownDescription: "SVM of parent volume in which clone is created off.",
									Computed:            true,
								},
								"parent_volume": schema.StringAttribute{
									MarkdownDescription: "Parent volume of the clone.",
									Computed:            true,
								},
								"parent_snapshot": schema.StringAttribute{
									MarkdownDescription: "Parent snapshot of the clone.",
									Computed:            true,
								},
							},
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the volume clone",
							Computed:            true,
						},
						"nas": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"junction_path": schema.StringAttribute{
									MarkdownDescription: "Junction path of the clone.",
									Computed:            true,
								},
								"group_id": schema.Int64Attribute{
									MarkdownDescription: "The UNIX group ID for the clone volume.",
									Computed:            true,
								},
								"user_id": schema.Int64Attribute{
									MarkdownDescription: "The UNIX user ID for the clone volume.",
									Computed:            true,
								},
							},
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "UUID of the volume clone.",
							Computed:            true,
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "List of Volume Clones",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *StorageVolumeClonesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *StorageVolumeClonesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageVolumeClonesDataSourceModel

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

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	if data.Filter == nil || data.Filter.SVMName.IsNull() {
		errorHandler.MakeAndReportError("No SVM specified", "svm_name must be specified in filter")
		return
	}

	var filter *interfaces.StorageVolumeClonesFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.StorageVolumeClonesFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
			Name:    data.Filter.Name.ValueString(),
		}
	}
	
	restInfo, err := interfaces.GetVolumeClones(errorHandler, *client, filter, cluster.Version)
	if err != nil {
		// error reporting done inside GetVolumeClones
		return
	}

	data.StorageVolumeClones = make([]StorageVolumeCloneDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		// Clone
		clone := &StorageVolumeCloneDataSourceClone{
			IsFlexclone:    types.BoolValue(record.Clone.IsFlexclone),
			Split:          types.BoolValue(!record.Clone.IsFlexclone),
			ParentVolume:   types.StringNull(),
			ParentSnapshot: types.StringNull(),
			ParentSVM:      types.StringNull(),
		}
		if record.Clone.ParentVolume.Name != "" {
			clone.ParentVolume = types.StringValue(record.Clone.ParentVolume.Name)
		}
		if record.Clone.ParentSnapshot != nil && record.Clone.ParentSnapshot.Name != "" {
			clone.ParentSnapshot = types.StringValue(record.Clone.ParentSnapshot.Name)
		}
		if record.Clone.ParentSVM.Name != "" {
			clone.ParentSVM = types.StringValue(record.Clone.ParentSVM.Name)
		}

		// NAS
		nas := &StorageVolumeCloneDataSourceNAS{
			JunctionPath: types.StringNull(),
			GroupID:      types.Int64Null(),
			UserID:       types.Int64Null(),
		}
		if record.NAS.JunctionPath != nil {
			nas.JunctionPath = types.StringValue(*record.NAS.JunctionPath)
		}
		if record.NAS.GroupID != nil {
			nas.GroupID = types.Int64Value(int64(*record.NAS.GroupID))
		}
		if record.NAS.UserID != nil {
			nas.UserID = types.Int64Value(int64(*record.NAS.UserID))
		}

		data.StorageVolumeClones[index] = StorageVolumeCloneDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVMName:       types.StringValue(record.SVM.Name),
			Name:          types.StringValue(record.Name),
			Type:          types.StringValue(record.Type),
			Clone:		   clone,
			NAS:           nas,
			ID:            types.StringValue(record.ID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
