package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageVolumeDataSource{}

// NewStorageVolumeDataSource is a helper function to simplify the provider implementation.
func NewStorageVolumeDataSource() datasource.DataSource {
	return &StorageVolumeDataSource{
		config: resourceOrDataSourceConfig{
			name: "storage_volume_data_source",
		},
	}
}

// StorageVolumeDataSource defines the data source implementation.
type StorageVolumeDataSource struct {
	config resourceOrDataSourceConfig
}

// StorageVolumeDataSourceModel describes the data source data model.
type StorageVolumeDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	// Volume Variables
	Name                      types.String   `tfsdk:"name"`
	SVMName                   types.String   `tfsdk:"svm_name"`
	Size                      types.Int64    `tfsdk:"size"`
	SizeUnit                  types.String   `tfsdk:"size_unit"`
	IsOnline                  types.Bool     `tfsdk:"is_online"`
	Type                      types.String   `tfsdk:"type"`
	ExportPolicy              types.String   `tfsdk:"export_policy"`
	JunctionPath              types.String   `tfsdk:"junction_path"`
	SpaceGuarantee            types.String   `tfsdk:"space_guarantee"`
	PercentSnapshotSpace      types.Int64    `tfsdk:"percent_snapshot_space"`
	SecurityStyle             types.String   `tfsdk:"security_style"`
	Encrypt                   types.Bool     `tfsdk:"encrypt"`
	EfficiencyPolicy          types.String   `tfsdk:"efficiency_policy"`
	UnixPermissions           types.String   `tfsdk:"unix_permissions"`
	GroupID                   types.Int64    `tfsdk:"group_id"`
	UserID                    types.Int64    `tfsdk:"user_id"`
	SnapshotPolicy            types.String   `tfsdk:"snapshot_policy"`
	Language                  types.String   `tfsdk:"language"`
	QosPolicyGroup            types.String   `tfsdk:"qos_policy_group"`
	QosAdaptivePolicyGroup    types.String   `tfsdk:"qos_adaptive_policy_group"`
	TieringPolicy             types.String   `tfsdk:"tiering_policy"`
	Comment                   types.String   `tfsdk:"comment"`
	Compression               types.String   `tfsdk:"compression"`
	TieringMinimumCoolingDays types.Int64    `tfsdk:"tiering_minimum_cooling_days"`
	LogicalSpaceEnforcement   types.Bool     `tfsdk:"logical_space_enforcement"`
	LogicalSpaceReporting     types.Bool     `tfsdk:"logical_space_reporting"`
	Aggregates                []types.String `tfsdk:"aggregates"`
	SnaplockType              types.String   `tfsdk:"snaplock_type"`
	Analytics                 types.String   `tfsdk:"analytics"`
	UUID                      types.String   `tfsdk:"uuid"`
}

// Metadata returns the data source type name.
func (d *StorageVolumeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
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
				MarkdownDescription: "Volume name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the svm to use",
				Required:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the volume",
				Computed:            true,
			},
			"size_unit": schema.StringAttribute{
				MarkdownDescription: "The unit used to interpret the size parameter",
				Computed:            true,
			},
			"is_online": schema.BoolAttribute{
				MarkdownDescription: "Whether the specified volume is online, or not",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The volume type, either read-write (RW) or data-protection (DP)",
				Computed:            true,
			},
			"export_policy": schema.StringAttribute{
				MarkdownDescription: "The name of the export policy",
				Computed:            true,
			},
			"junction_path": schema.StringAttribute{
				MarkdownDescription: "Junction path of the volume",
				Computed:            true,
			},
			"space_guarantee": schema.StringAttribute{
				MarkdownDescription: "Space guarantee style for the volume",
				Computed:            true,
			},
			"percent_snapshot_space": schema.Int64Attribute{
				MarkdownDescription: "Amount of space reserved for snapshot copies of the volume",
				Computed:            true,
			},
			"security_style": schema.StringAttribute{
				MarkdownDescription: "The security style associated to the volume",
				Computed:            true,
			},
			"encrypt": schema.BoolAttribute{
				MarkdownDescription: "Whether or not to enable Volume Encryption",
				Computed:            true,
			},
			"efficiency_policy": schema.StringAttribute{
				MarkdownDescription: "Allows a storage efficiency policy to be set on volume creation",
				Computed:            true,
			},
			"unix_permissions": schema.StringAttribute{
				MarkdownDescription: "Unix permission bits in octal or symbolic format. For example, 0 is equivalent to ------------, 777 is equivalent to ---rwxrwxrwx,both formats are accepted",
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
			"snapshot_policy": schema.StringAttribute{
				MarkdownDescription: "The name of the snapshot policy",
				Computed:            true,
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "Language to use for volume",
				Computed:            true,
			},
			"qos_policy_group": schema.StringAttribute{
				MarkdownDescription: "Specifies a QoS policy group to be set on volume",
				Computed:            true,
			},
			"qos_adaptive_policy_group": schema.StringAttribute{
				MarkdownDescription: "Specifies a QoS adaptive policy group to be set on volume",
				Computed:            true,
			},
			"tiering_policy": schema.StringAttribute{
				MarkdownDescription: "The tiering policy that is to be associated with the volume",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Sets a comment associated with the volume",
				Computed:            true,
			},
			"compression": schema.StringAttribute{
				MarkdownDescription: "Whether to enable compression for the volume (HDD and Flash Pool aggregates, AFF platforms)",
				Computed:            true,
			},
			"tiering_minimum_cooling_days": schema.Int64Attribute{
				MarkdownDescription: "Determines how many days must pass before inactive data in a volume using the Auto or Snapshot-Only policy is considered cold and eligible for tiering",
				Computed:            true,
			},
			"logical_space_enforcement": schema.BoolAttribute{
				MarkdownDescription: "Whether to perform logical space accounting on the volume",
				Computed:            true,
			},
			"logical_space_reporting": schema.BoolAttribute{
				MarkdownDescription: "Whether to report space logically",
				Computed:            true,
			},
			"aggregates": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of aggregates in which to create the volume",
			},
			"snaplock_type": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The SnapLock type of the volume",
			},
			"analytics": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Set file system analytics state of the volume",
			},
			"uuid": schema.StringAttribute{
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
	client, err := getRestClient(errorHandler, d.config, data.CxProfileName)
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
	data.UUID = types.StringValue(volume.UUID)
	data.SVMName = types.StringValue(volume.SVM.Name)
	var aggregates []types.String
	for _, e := range volume.Aggregates {
		aggregates = append(aggregates, types.StringValue(e.Name))
	}
	data.Aggregates = aggregates
	size, sizeUnit := interfaces.ByteFormat(int64(volume.Space.Size))
	data.Size = types.Int64Value(size)
	data.SizeUnit = types.StringValue(sizeUnit)
	data.IsOnline = types.BoolValue(interfaces.OnlineToBool(volume.State))
	data.Type = types.StringValue(volume.Type)
	data.ExportPolicy = types.StringValue(volume.NAS.ExportPolicy.Name)
	data.JunctionPath = types.StringValue(volume.NAS.JunctionPath)
	data.SpaceGuarantee = types.StringValue(volume.SpaceGuarantee.Type)
	data.PercentSnapshotSpace = types.Int64Value(int64(volume.Space.Snapshot.ReservePercent))
	data.SecurityStyle = types.StringValue(volume.NAS.SecurityStyle)
	data.Encrypt = types.BoolValue(volume.Encryption.Enabled)
	data.EfficiencyPolicy = types.StringValue(volume.Efficiency.Policy.Name)
	data.UnixPermissions = types.StringValue(strconv.Itoa(volume.NAS.UnixPermissions))
	data.GroupID = types.Int64Value(int64(volume.NAS.GroupID))
	data.UserID = types.Int64Value(int64(volume.NAS.UserID))
	data.SnapshotPolicy = types.StringValue(volume.SnapshotPolicy.Name)
	data.Language = types.StringValue(volume.Language)
	data.QosPolicyGroup = types.StringValue(volume.QOS.Policy.Name)
	data.QosAdaptivePolicyGroup = types.StringValue(volume.QOS.Policy.Name)
	data.TieringPolicy = types.StringValue(volume.TieringPolicy.Policy)
	data.Comment = types.StringValue(volume.Comment)
	data.Compression = types.StringValue(volume.Efficiency.Compression)
	data.TieringMinimumCoolingDays = types.Int64Value(int64(volume.TieringPolicy.MinCoolingDays))
	data.LogicalSpaceEnforcement = types.BoolValue(volume.Space.LogicalSpace.Enforcement)
	data.LogicalSpaceReporting = types.BoolValue(volume.Space.LogicalSpace.Reporting)
	data.SnaplockType = types.StringValue(volume.Snaplock.Type)
	data.Analytics = types.StringValue(volume.Analytics.State)

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
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	d.config.providerConfig = config
}
