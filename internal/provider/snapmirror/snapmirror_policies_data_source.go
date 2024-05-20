package snapmirror

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SnapmirrorPoliciesDataSource{}

// NewSnapmirrorPoliciesDataSource is a helper function to simplify the provider implementation.
func NewSnapmirrorPoliciesDataSource() datasource.DataSource {
	return &SnapmirrorPoliciesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "snapmirror_policies",
		},
	}
}

// SnapmirrorPoliciesDataSource defines the data source implementation.
type SnapmirrorPoliciesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// SnapmirrorPoliciesDataSourceModel describes the data source data model.
type SnapmirrorPoliciesDataSourceModel struct {
	CxProfileName      types.String                           `tfsdk:"cx_profile_name"`
	SnapmirrorPolicies []SnapmirrorPolicyDataSourceModel      `tfsdk:"snapmirror_policies"`
	Filter             *SnapmirrorPolicyDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *SnapmirrorPoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *SnapmirrorPoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SnapmirrorPolicies data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "SnapmirrorPolicy name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"snapmirror_policies": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
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
				},
				Computed:            true,
				MarkdownDescription: "",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *SnapmirrorPoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SnapmirrorPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SnapmirrorPoliciesDataSourceModel

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

	var filter *interfaces.SnapmirrorPolicyFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.SnapmirrorPolicyFilterModel{
			Name: data.Filter.Name.ValueString(),
		}
	}
	restInfo, err := interfaces.GetSnapmirrorPolicies(errorHandler, *client, filter, cluster.Version)
	if err != nil {
		// error reporting done inside GetSnapmirrorPolicies
		return
	}

	data.SnapmirrorPolicies = make([]SnapmirrorPolicyDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		var retentions = make([]RetentionGetDataModel, len(record.Retention))
		for i, retention := range record.Retention {
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

		data.SnapmirrorPolicies[index] = SnapmirrorPolicyDataSourceModel{
			CxProfileName:             types.String(data.CxProfileName),
			Name:                      types.StringValue(record.Name),
			SVMName:                   types.StringValue(record.SVM.Name),
			Type:                      types.StringValue(record.Type),
			SyncType:                  types.StringValue(record.SyncType),
			Comment:                   types.StringValue(record.Comment),
			TransferScheduleName:      types.StringValue(record.TransferSchedule.Name),
			NetworkCompressionEnabled: types.BoolValue(record.NetworkCompressionEnabled),
			IdentityPreservation:      types.StringValue(record.IdentityPreservation),
			ID:                        types.StringValue(record.UUID),
		}

		if cluster.Version.Generation == 9 && cluster.Version.Major > 9 {
			data.SnapmirrorPolicies[index].CopyAllSourceSnapshots = types.BoolValue(record.CopyAllSourceSnapshots)
		}
		if cluster.Version.Generation == 9 && cluster.Version.Major > 10 {
			data.SnapmirrorPolicies[index].CreateSnapshotOnSource = types.BoolValue(record.CreateSnapshotOnSource)
			data.SnapmirrorPolicies[index].CopyLatestSourceSnapshot = types.BoolValue(record.CopyLatestSourceSnapshot)
		}

		if len(record.Retention) == 0 {
			data.SnapmirrorPolicies[index].Retention = nil
		} else {
			data.SnapmirrorPolicies[index].Retention = retentions
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
