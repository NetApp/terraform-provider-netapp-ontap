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
var _ datasource.DataSource = &StorageVolumeDataSource{}

// NewStorageVolumeDataSource is a helper function to simplify the provider implementation.
func NewStorageVolumeDataSource() datasource.DataSource {
	return &StorageVolumeDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "volume",
		},
	}
}

// StorageVolumeDataSource defines the data source implementation.
type StorageVolumeDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageVolumeDataSourceModel describes the data source data model.
type StorageVolumeDataSourceModel struct {
	CxProfileName  types.String                        `tfsdk:"cx_profile_name"`
	Name           types.String                        `tfsdk:"name"`
	SVMName        types.String                        `tfsdk:"svm_name"`
	State          types.String                        `tfsdk:"state"`
	Type           types.String                        `tfsdk:"type"`
	SpaceGuarantee types.String                        `tfsdk:"space_guarantee"`
	Encrypt        types.Bool                          `tfsdk:"encryption"`
	SnapshotPolicy types.String                        `tfsdk:"snapshot_policy"`
	Language       types.String                        `tfsdk:"language"`
	QOSPolicyGroup types.String                        `tfsdk:"qos_policy_group"`
	Comment        types.String                        `tfsdk:"comment"`
	Aggregates     []StorageVolumeDataSourceAggregates `tfsdk:"aggregates"`
	ID             types.String                        `tfsdk:"id"`
	Space          *StorageVolumeDataSourceSpace       `tfsdk:"space"`
	Nas            *StorageVolumeDataSourceNas         `tfsdk:"nas"`
	Tiering        *StorageVolumeDataSourceTiering     `tfsdk:"tiering"`
	Efficiency     *StorageVolumeDataSourceEfficiency  `tfsdk:"efficiency"`
	SnapLock       *StorageVolumeDataSourceSnapLock    `tfsdk:"snaplock"`
	Analytics      *StorageVolumeDataSourceAnalytics   `tfsdk:"analytics"`
}

// StorageVolumeDataSourceAggregates describes the analytics model.
type StorageVolumeDataSourceAggregates struct {
	Name types.String `tfsdk:"name"`
}

// StorageVolumeDataSourceAnalytics describes the analytics model.
type StorageVolumeDataSourceAnalytics struct {
	State types.String `tfsdk:"state"`
}

// StorageVolumeDataSourceSnapLock describes the snaplock model.
type StorageVolumeDataSourceSnapLock struct {
	SnaplockType types.String `tfsdk:"type"`
}

// StorageVolumeDataSourceEfficiency describes the efficiency model.
type StorageVolumeDataSourceEfficiency struct {
	Policy      types.String `tfsdk:"policy_name"`
	Compression types.String `tfsdk:"compression"`
}

// StorageVolumeDataSourceTiering describes the tiering model.
type StorageVolumeDataSourceTiering struct {
	Policy             types.String `tfsdk:"policy_name"`
	MinimumCoolingDays types.Int64  `tfsdk:"minimum_cooling_days"`
}

// StorageVolumeDataSourceNas describes the Nas model.
type StorageVolumeDataSourceNas struct {
	ExportPolicy    types.String `tfsdk:"export_policy_name"`
	JunctionPath    types.String `tfsdk:"junction_path"`
	GroupID         types.Int64  `tfsdk:"group_id"`
	UserID          types.Int64  `tfsdk:"user_id"`
	SecurityStyle   types.String `tfsdk:"security_style"`
	UnixPermissions types.Int64  `tfsdk:"unix_permissions"`
}

// StorageVolumeDataSourceSpace describes the space model.
type StorageVolumeDataSourceSpace struct {
	Size                 types.Int64                               `tfsdk:"size"`
	SizeUnit             types.String                              `tfsdk:"size_unit"`
	PercentSnapshotSpace types.Int64                               `tfsdk:"percent_snapshot_space"`
	LogicalSpace         *StorageVolumeDataSourceSpaceLogicalSpace `tfsdk:"logical_space"`
}

// StorageVolumeDataSourceSpaceLogicalSpace describes the logical space model within sapce model.
type StorageVolumeDataSourceSpaceLogicalSpace struct {
	Enforcement types.Bool `tfsdk:"enforcement"`
	Reporting   types.Bool `tfsdk:"reporting"`
}

// Metadata returns the data source type name.
func (d *StorageVolumeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *StorageVolumeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Storage Volume data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the volume to manage",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the svm to use",
				Required:            true,
			},
			"aggregates": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of aggregates that the volume is on",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the aggregate",
							Computed:            true,
						},
					},
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Whether the specified volume is online, or not",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The volume type, either read-write (RW) or data-protection (DP)",
				Computed:            true,
			},
			"space_guarantee": schema.StringAttribute{
				MarkdownDescription: "Space guarantee style for the volume",
				Computed:            true,
			},
			"encryption": schema.BoolAttribute{
				MarkdownDescription: "Whether or not to enable Volume Encryption",
				Computed:            true,
			},
			"snapshot_policy": schema.StringAttribute{
				MarkdownDescription: "The name of the snapshot policy",
				Computed:            true,
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "Language to use for volume",
				Computed:            true,
			},
			// with Rest API qos_policy_group and qos_adaptive_policy_group are now the same thing and cannot be set at the same time
			"qos_policy_group": schema.StringAttribute{
				MarkdownDescription: "Specifies a QoS policy group to be set on volume",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Sets a comment associated with the volume",
				Computed:            true,
			},
			"space": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"size": schema.Int64Attribute{
						MarkdownDescription: "The size of the volume",
						Computed:            true,
					},
					"size_unit": schema.StringAttribute{
						MarkdownDescription: "The unit used to interpret the size parameter",
						Computed:            true,
					},
					"percent_snapshot_space": schema.Int64Attribute{
						MarkdownDescription: "Amount of space reserved for snapshot copies of the volume",
						Computed:            true,
					},
					"logical_space": schema.SingleNestedAttribute{
						Computed: true,
						Attributes: map[string]schema.Attribute{
							"enforcement": schema.BoolAttribute{
								MarkdownDescription: "Whether to perform logical space accounting on the volume",
								Computed:            true,
							},
							"reporting": schema.BoolAttribute{
								MarkdownDescription: "Whether to report space logically",
								Computed:            true,
							},
						},
					},
				},
			},
			"nas": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"export_policy_name": schema.StringAttribute{
						MarkdownDescription: "The name of the export policy",
						Computed:            true,
					},
					"junction_path": schema.StringAttribute{
						MarkdownDescription: "Junction path of the volume",
						Computed:            true,
					},
					"group_id": schema.Int64Attribute{
						MarkdownDescription: "The UNIX group ID for the volume",
						Computed:            true,
					},
					"user_id": schema.Int64Attribute{
						MarkdownDescription: "The UNIX user ID for the volume",
						Computed:            true,
					},
					"security_style": schema.StringAttribute{
						MarkdownDescription: "The security style associated to the volume",
						Computed:            true,
					},
					"unix_permissions": schema.Int64Attribute{
						MarkdownDescription: "Unix permission bits in octal or symbolic format. For example, 0 is equivalent to ------------, 777 is equivalent to ---rwxrwxrwx,both formats are accepted",
						Computed:            true,
					},
				},
			},
			"tiering": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"policy_name": schema.StringAttribute{
						MarkdownDescription: "The tiering policy that is to be associated with the volume",
						Computed:            true,
					},
					"minimum_cooling_days": schema.Int64Attribute{
						MarkdownDescription: "Determines how many days must pass before inactive data in a volume using the Auto or Snapshot-Only policy is considered cold and eligible for tiering",
						Computed:            true,
					},
				},
			},
			"efficiency": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"policy_name": schema.StringAttribute{
						MarkdownDescription: "Allows a storage efficiency policy to be set on volume creation",
						Computed:            true,
					},
					"compression": schema.StringAttribute{
						MarkdownDescription: "Whether to enable compression for the volume (HDD and Flash Pool aggregates)",
						Computed:            true,
					},
				},
			},

			"snaplock": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "The SnapLock type of the volume",
						Computed:            true,
					},
				},
			},
			"analytics": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"state": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Set file system analytics state of the volume",
					},
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Volume identifier",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *StorageVolumeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageVolumeDataSourceModel

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

	volume, err := interfaces.GetStorageVolumeByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		return
	}
	if volume == nil {
		errorHandler.MakeAndReportError("No volume found", fmt.Sprintf("Volume %s not found.", data.Name))
		return
	}
	data.Name = types.StringValue(volume.Name)
	data.SVMName = types.StringValue(volume.SVM.Name)
	var aggregates = make([]StorageVolumeDataSourceAggregates, len(volume.Aggregates))
	for i, v := range volume.Aggregates {
		aggregates[i].Name = types.StringValue(v.Name)
	}
	data.Aggregates = aggregates
	data.State = types.StringValue(volume.State)
	data.Type = types.StringValue(volume.Type)
	data.SpaceGuarantee = types.StringValue(volume.SpaceGuarantee.Type)
	data.Encrypt = types.BoolValue(volume.Encryption.Enabled)
	data.SnapshotPolicy = types.StringValue(volume.SnapshotPolicy.Name)
	data.Language = types.StringValue(volume.Language)
	data.QOSPolicyGroup = types.StringValue(volume.QOS.Policy.Name)
	data.Comment = types.StringValue(volume.Comment)
	vsize, vunits := interfaces.ByteFormat(int64(volume.Space.Size))
	data.Space = &StorageVolumeDataSourceSpace{
		Size:                 types.Int64Value(vsize),
		SizeUnit:             types.StringValue(vunits),
		PercentSnapshotSpace: types.Int64Value(int64(volume.Space.Snapshot.ReservePercent)),
		LogicalSpace: &StorageVolumeDataSourceSpaceLogicalSpace{
			Enforcement: types.BoolValue(volume.Space.LogicalSpace.Enforcement),
			Reporting:   types.BoolValue(volume.Space.LogicalSpace.Reporting),
		},
	}
	data.Nas = &StorageVolumeDataSourceNas{
		ExportPolicy:    types.StringValue(volume.NAS.ExportPolicy.Name),
		JunctionPath:    types.StringValue(volume.NAS.JunctionPath),
		GroupID:         types.Int64Value(int64(volume.NAS.GroupID)),
		UserID:          types.Int64Value(int64(volume.NAS.UserID)),
		SecurityStyle:   types.StringValue(volume.NAS.SecurityStyle),
		UnixPermissions: types.Int64Value(int64(volume.NAS.UnixPermissions)),
	}
	data.Tiering = &StorageVolumeDataSourceTiering{
		Policy:             types.StringValue(volume.TieringPolicy.Policy),
		MinimumCoolingDays: types.Int64Value(int64(volume.TieringPolicy.MinCoolingDays)),
	}
	data.Efficiency = &StorageVolumeDataSourceEfficiency{
		Policy:      types.StringValue(volume.Efficiency.Policy.Name),
		Compression: types.StringValue(volume.Efficiency.Compression),
	}
	data.SnapLock = &StorageVolumeDataSourceSnapLock{
		SnaplockType: types.StringValue(volume.Snaplock.Type),
	}
	data.Analytics = &StorageVolumeDataSourceAnalytics{
		State: types.StringValue(volume.Analytics.State),
	}
	data.ID = types.StringValue(volume.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source.
func (d *StorageVolumeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
