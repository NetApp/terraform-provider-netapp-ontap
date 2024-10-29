package snapmirror

import (
	"context"
	"fmt"
	"strconv"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SnapmirrorPolicyDataSource{}

// NewSnapmirrorPolicyDataSource is a helper function to simplify the provider implementation.
func NewSnapmirrorPolicyDataSource() datasource.DataSource {
	return &SnapmirrorPolicyDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "snapmirror_policy",
		},
	}
}

// NewSnapmirrorPolicyDataSourceAlias is a helper function to simplify the provider implementation.
func NewSnapmirrorPolicyDataSourceAlias() datasource.DataSource {
	return &SnapmirrorPolicyDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "snapmirror_policy_data_source",
		},
	}
}

// SnapmirrorPolicyDataSource defines the data source implementation.
type SnapmirrorPolicyDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// SnapmirrorPolicyDataSourceModel describes the data source data model.
type SnapmirrorPolicyDataSourceModel struct {
	CxProfileName             types.String            `tfsdk:"cx_profile_name"`
	Name                      types.String            `tfsdk:"name"`
	SVMName                   types.String            `tfsdk:"svm_name"`
	Type                      types.String            `tfsdk:"type"`
	SyncType                  types.String            `tfsdk:"sync_type"`
	Comment                   types.String            `tfsdk:"comment"`
	TransferScheduleName      types.String            `tfsdk:"transfer_schedule_name"`
	NetworkCompressionEnabled types.Bool              `tfsdk:"network_compression_enabled"`
	Retention                 []RetentionGetDataModel `tfsdk:"retention"`
	IdentityPreservation      types.String            `tfsdk:"identity_preservation"`
	CopyAllSourceSnapshots    types.Bool              `tfsdk:"copy_all_source_snapshots"`
	CopyLatestSourceSnapshot  types.Bool              `tfsdk:"copy_latest_source_snapshot"`
	ID                        types.String            `tfsdk:"id"`
	CreateSnapshotOnSource    types.Bool              `tfsdk:"create_snapshot_on_source"`
}

// RetentionGetDataModel defines the resource get retention model
type RetentionGetDataModel struct {
	CreationScheduleName types.String `tfsdk:"creation_schedule_name"`
	Count                types.Int64  `tfsdk:"count"`
	Label                types.String `tfsdk:"label"`
	Prefix               types.String `tfsdk:"prefix"`
}

// SnapmirrorPolicyDataSourceFilterModel describes the data source data model for queries.
type SnapmirrorPolicyDataSourceFilterModel struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *SnapmirrorPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *SnapmirrorPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SnapmirrorPolicy data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "SnapmirrorPolicy name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "SnapmirrorPolicy svm name",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "SnapmirrorPolicy type. [async, sync, continuous]",
				Computed:            true,
			},
			"sync_type": schema.StringAttribute{
				MarkdownDescription: "SnapmirrorPolicy sync type. [sync, strict_sync, automated_failover]",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment associated with the policy.",
				Computed:            true,
			},
			"transfer_schedule_name": schema.StringAttribute{
				MarkdownDescription: "The schedule used to update asynchronous relationships",
				Computed:            true,
			},
			"network_compression_enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether network compression is enabled for transfers",
				Computed:            true,
			},
			"retention": schema.ListNestedAttribute{
				MarkdownDescription: "Rules for Snapshot copy retention.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"creation_schedule_name": schema.StringAttribute{
							MarkdownDescription: "Schedule used to create Snapshot copies on the destination for long term retention.",
							Computed:            true,
						},
						"count": schema.Int64Attribute{
							MarkdownDescription: "Number of Snapshot copies to be kept for retention.",
							Computed:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: "Snapshot copy label",
							Computed:            true,
						},
						"prefix": schema.StringAttribute{
							MarkdownDescription: "Specifies the prefix for the Snapshot copy name to be created as per the schedule",
							Computed:            true,
						},
					},
				},
			},
			"identity_preservation": schema.StringAttribute{
				MarkdownDescription: "Specifies which configuration of the source SVM is replicated to the destination SVM.",
				Computed:            true,
			},
			"copy_all_source_snapshots": schema.BoolAttribute{
				MarkdownDescription: "Specifies that all the source Snapshot copies (including the one created by SnapMirror before the transfer begins) should be copied to the destination on a transfer.",
				Computed:            true,
			},
			"copy_latest_source_snapshot": schema.BoolAttribute{
				MarkdownDescription: "Specifies that the latest source Snapshot copy (created by SnapMirror before the transfer begins) should be copied to the destination on a transfer. 'Retention' properties cannot be specified along with this property. This is applicable only to async policies. Property can only be set to 'true'.",
				Computed:            true,
			},
			"create_snapshot_on_source": schema.BoolAttribute{
				MarkdownDescription: "Specifies that all the source Snapshot copies (including the one created by SnapMirror before the transfer begins) should be copied to the destination on a transfer.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "SnapmirrorPolicy uuid",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *SnapmirrorPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SnapmirrorPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SnapmirrorPolicyDataSourceModel

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

	restInfo, err := interfaces.GetSnapmirrorPolicyDataSourceByName(errorHandler, *client, data.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetSnapmirrorPolicy
		return
	}

	var retentions = make([]RetentionGetDataModel, len(restInfo.Retention))
	for i, retention := range restInfo.Retention {
		retentionCount, err := strconv.Atoi(retention.Count)
		if err != nil {
			errorHandler.MakeAndReportError("Decode count error", "snapmirror_policy retention count is not valid")
			return
		}
		retentions[i] = RetentionGetDataModel{
			CreationScheduleName: types.StringValue(retention.CreationSchedule.Name),
			Count:                types.Int64Value(int64(retentionCount)),
			Label:                types.StringValue(retention.Label),
			Prefix:               types.StringValue(retention.Prefix),
		}
	}

	data = SnapmirrorPolicyDataSourceModel{
		CxProfileName:             types.String(data.CxProfileName),
		Name:                      types.StringValue(restInfo.Name),
		SVMName:                   types.StringValue(restInfo.SVM.Name),
		Type:                      types.StringValue(restInfo.Type),
		SyncType:                  types.StringValue(restInfo.SyncType),
		Comment:                   types.StringValue(restInfo.Comment),
		TransferScheduleName:      types.StringValue(restInfo.TransferSchedule.Name),
		NetworkCompressionEnabled: types.BoolValue(restInfo.NetworkCompressionEnabled),
		IdentityPreservation:      types.StringValue(restInfo.IdentityPreservation),
		ID:                        types.StringValue(restInfo.UUID),
	}

	if len(restInfo.Retention) == 0 {
		data.Retention = nil
	} else {
		data.Retention = retentions
	}

	if cluster.Version.Generation == 9 && cluster.Version.Major > 9 {
		data.CopyAllSourceSnapshots = types.BoolValue(restInfo.CopyAllSourceSnapshots)
	}
	if cluster.Version.Generation == 9 && cluster.Version.Major > 10 {
		data.CreateSnapshotOnSource = types.BoolValue(restInfo.CreateSnapshotOnSource)
		data.CopyLatestSourceSnapshot = types.BoolValue(restInfo.CopyLatestSourceSnapshot)
	}
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
