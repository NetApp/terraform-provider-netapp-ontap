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
var _ datasource.DataSource = &SnapshotPoliciesDataSource{}

// NewSnapshotPoliciesDataSource is a helper function to simplify the provider implementation.
func NewSnapshotPoliciesDataSource() datasource.DataSource {
	return &SnapshotPoliciesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "snapshot_policies",
		},
	}
}

// NewSnapshotPoliciesDataSourceAlias is a helper function to simplify the provider implementation.
func NewSnapshotPoliciesDataSourceAlias() datasource.DataSource {
	return &SnapshotPoliciesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "storage_snapshot_policies_data_source",
		},
	}
}

// SnapshotPoliciesDataSource defines the data source implementation.
type SnapshotPoliciesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// SnapshotPoliciesDataSourceModel describes the data source data model.
type SnapshotPoliciesDataSourceModel struct {
	CxProfileName    types.String                         `tfsdk:"cx_profile_name"`
	SnapshotPolicies []SnapshotPolicyDataSourceModel      `tfsdk:"storage_snapshot_policies"`
	Filter           *SnapshotPolicyDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *SnapshotPoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *SnapshotPoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SnapshotPolicies data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "SnapshotPolicy name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "SnapshotPolicy svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"storage_snapshot_policies": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "SnapshotPolicy name",
							Required:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "IPInterface svm name",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "SnapshotPolicy UUID",
							Computed:            true,
						},
						"copies": schema.ListNestedAttribute{
							MarkdownDescription: "Snapshot copy",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"count": schema.Int64Attribute{
										MarkdownDescription: "The number of Snapshot copies to maintain for this schedule",
										Computed:            true,
									},
									"schedule": schema.SingleNestedAttribute{
										MarkdownDescription: "Schedule at which Snapshot copies are captured on the volume",
										Required:            true,
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												MarkdownDescription: "Some common schedules already defined in the system are hourly, daily, weekly, at 15 minute intervals, and at 5 minute intervals. Snapshot copy policies with custom schedules can be referenced",
												Computed:            true,
											},
										},
									},
									"retention_period": schema.StringAttribute{
										MarkdownDescription: "The retention period of Snapshot copies for this schedule",
										Computed:            true,
									},
									"snapmirror_label": schema.StringAttribute{
										MarkdownDescription: "Label for SnapMirror operations",
										Computed:            true,
									},
									"prefix": schema.StringAttribute{
										MarkdownDescription: "The prefix to use while creating Snapshot copies at regular intervals",
										Computed:            true,
									},
								},
							},
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "A comment associated with the Snapshot copy policy",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Is the Snapshot copy policy enabled?",
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
func (d *SnapshotPoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SnapshotPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SnapshotPoliciesDataSourceModel

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

	var filter *interfaces.SnapshotPolicyGetDataFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.SnapshotPolicyGetDataFilterModel{
			Name:    data.Filter.Name.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetSnapshotPolicies(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetSnapshotPolicies
		return
	}

	data.SnapshotPolicies = make([]SnapshotPolicyDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		var copies = make([]CopyResourceModel, len(record.Copies))

		for i, copiesRecord := range record.Copies {
			copies[i] = CopyResourceModel{
				Count: types.Int64Value(copiesRecord.Count),
				Schedule: ScheduleResourceModel{
					Name: types.StringValue(copiesRecord.Schedule.Name),
				},
				SnapmirrorLabel: types.StringValue(copiesRecord.SnapmirrorLabel),
				Prefix:          types.StringValue(copiesRecord.Prefix),
			}

			if copiesRecord.RetentionPeriod != "" {
				copies[i].RetentionPeriod = types.StringValue(copiesRecord.RetentionPeriod)
			}
		}

		data.SnapshotPolicies[index] = SnapshotPolicyDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
			SVMName:       types.StringValue(record.SVM.Name),
			ID:            types.StringValue(record.UUID),
			Copies:        copies,
			Comment:       types.StringValue(record.Comment),
			Enabled:       types.BoolValue(record.Enabled),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
