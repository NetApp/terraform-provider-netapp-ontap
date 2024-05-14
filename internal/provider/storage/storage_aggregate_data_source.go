package storage

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageAggregateDataSource{}

// NewStorageAggregateDataSource is a helper function to simplify the provider implementation.
func NewStorageAggregateDataSource() datasource.DataSource {
	return &StorageAggregateDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "aggregate",
		},
	}
}

// StorageAggregateDataSource defines the data source implementation.
type StorageAggregateDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageAggregateDataSourceModel describes the data source data model.
type StorageAggregateDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	ID            types.String `tfsdk:"id"`
	State         types.String `tfsdk:"state"`
	Node          types.String `tfsdk:"node"`
	DiskClass     types.String `tfsdk:"disk_class"`
	DiskCount     types.Int64  `tfsdk:"disk_count"`
	RaidSize      types.Int64  `tfsdk:"raid_size"`
	RaidType      types.String `tfsdk:"raid_type"`
	IsMirrored    types.Bool   `tfsdk:"is_mirrored"`
	SnaplockType  types.String `tfsdk:"snaplock_type"`
	Encryption    types.Bool   `tfsdk:"encryption"`
}

// StorageAggregateDataSourceFilterModel describes the data source data model for queries.
type StorageAggregateDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *StorageAggregateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *StorageAggregateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageAggregate data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "StorageAggregate name",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Aggregate identifier",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Whether the specified aggregate should be enabled or disabled. Creates aggregate if doesnt exist.",
				Computed:            true,
			},
			"node": schema.StringAttribute{
				MarkdownDescription: "Node for the aggregate to be created on. If no node specified, mgmt lif home will be used. If disk_count is present, node name is required.",
				Computed:            true,
			},
			"disk_class": schema.StringAttribute{
				MarkdownDescription: "Class of disk to use to build aggregate. capacity_flash is listed in swagger, but rejected as invalid by ONTAP.",
				Computed:            true,
			},
			"disk_count": schema.Int64Attribute{
				MarkdownDescription: `Number of disks to place into the aggregate, including parity disks.
				The disks in this newly-created aggregate come from the spare disk pool.
				The smallest disks in this pool join the aggregate first, unless the disk_size argument is provided.
				Modifiable only if specified disk_count is larger than current disk_count.
				If the disk_count % raid_size == 1, only disk_count/raid_size * raid_size will be added.
				If disk_count is 6, raid_type is raid4, raid_size 4, all 6 disks will be added.
				If disk_count is 5, raid_type is raid4, raid_size 4, 5/4 * 4 = 4 will be added. 1 will not be added.
				`,
				Computed: true,
			},
			"raid_size": schema.Int64Attribute{
				MarkdownDescription: "Sets the maximum number of drives per raid group.",
				Computed:            true,
			},
			"raid_type": schema.StringAttribute{
				Computed: true,
			},
			"is_mirrored": schema.BoolAttribute{
				MarkdownDescription: `Specifies that the new aggregate be mirrored (have two plexes).
				If set to true, then the indicated disks will be split across the two plexes. By default, the new aggregate will not be mirrored.`,
				Computed: true,
			},
			"snaplock_type": schema.StringAttribute{
				MarkdownDescription: "Type of snaplock for the aggregate being created.",
				Computed:            true,
			},
			"encryption": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable software encryption. This is equivalent to -encrypt-with-aggr-key when using the CLI.Requires a VE license.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *StorageAggregateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *StorageAggregateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageAggregateDataSourceModel

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

	restInfo, err := interfaces.GetStorageAggregateByName(errorHandler, *client, data.Name.ValueString())
	if err != nil {
		// error reporting done inside GetStorageAggregate
		return
	}

	data.ID = types.StringValue(restInfo.UUID)
	data.DiskCount = types.Int64Value(restInfo.BlockStorage.Primary.DiskCount)
	data.DiskClass = types.StringValue(restInfo.BlockStorage.Primary.DiskClass)
	data.RaidType = types.StringValue(restInfo.BlockStorage.Primary.RaidType)
	data.RaidSize = types.Int64Value(restInfo.BlockStorage.Primary.RaidSize)
	data.Encryption = types.BoolValue(restInfo.DataEncryption.SoftwareEncryptionEnabled)
	data.IsMirrored = types.BoolValue(restInfo.BlockStorage.Mirror.Enabled)
	data.SnaplockType = types.StringValue(restInfo.SnaplockType)
	data.State = types.StringValue(restInfo.State)
	data.Name = types.StringValue(restInfo.Name)
	data.Node = types.StringValue(restInfo.Node.Name)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
