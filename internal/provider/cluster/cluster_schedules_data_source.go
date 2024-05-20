package cluster

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
var _ datasource.DataSource = &ClusterSchedulesDataSource{}

// NewClusterSchedulesDataSource is a helper function to simplify the provider implementation.
func NewClusterSchedulesDataSource() datasource.DataSource {
	return &ClusterSchedulesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cluster_schedules",
		},
	}
}

// ClusterSchedulesDataSource defines the data source implementation.
type ClusterSchedulesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ClusterSchedulesDataSourceModel describes the data source data model.
type ClusterSchedulesDataSourceModel struct {
	CxProfileName    types.String                          `tfsdk:"cx_profile_name"`
	ClusterSchedules []ClusterScheduleDataSourceModel      `tfsdk:"cluster_schedules"`
	Filter           *ClusterScheduleDataSourceFilterModel `tfsdk:"filter"`
}

// ClusterScheduleDataSourceFilterModel describes the data source data model for queries.
type ClusterScheduleDataSourceFilterModel struct {
	Type types.String `tfsdk:"type"`
}

// Metadata returns the data source type name.
func (d *ClusterSchedulesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ClusterSchedulesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ClusterSchedules data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Cluster schdeule type",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"cluster_schedules": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "ClusterSchedule name",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Cluster schdeule type",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "Cluster schedule UUID",
							Computed:            true,
						},
						"cron": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"minutes": schema.ListAttribute{
									ElementType:         types.Int64Type,
									MarkdownDescription: "List of cluster schedule minutes",
									Computed:            true,
								},
								"hours": schema.ListAttribute{
									ElementType:         types.Int64Type,
									MarkdownDescription: "List of cluster schedule hours",
									Computed:            true,
								},
								"days": schema.ListAttribute{
									ElementType:         types.Int64Type,
									MarkdownDescription: "List of cluster schedule days",
									Computed:            true,
								},
								"weekdays": schema.ListAttribute{
									ElementType:         types.Int64Type,
									MarkdownDescription: "List of cluster schedule weekdays",
									Computed:            true,
								},
								"months": schema.ListAttribute{
									ElementType:         types.Int64Type,
									MarkdownDescription: "List of cluster schedule months",
									Computed:            true,
								},
							},
							Computed: true,
						},
						"interval": schema.StringAttribute{
							MarkdownDescription: "Cluster schedule interval",
							Computed:            true,
						},
						"scope": schema.StringAttribute{
							MarkdownDescription: "Cluster schedule scope",
							Computed:            true,
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "Cluster Schedules data source",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ClusterSchedulesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ClusterSchedulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterSchedulesDataSourceModel

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

	var filter *interfaces.ClusterScheduleFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.ClusterScheduleFilterModel{
			Type: data.Filter.Type.ValueString(),
		}
	}
	restInfo, err := interfaces.GetListClusterSchedules(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetClusterSchedules
		return
	}

	data.ClusterSchedules = make([]ClusterScheduleDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.ClusterSchedules[index] = ClusterScheduleDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
			ID:            types.StringValue(record.UUID),
			Type:          types.StringValue(record.Type),
			Scope:         types.StringValue(record.Scope),
		}

		if record.Type == "cron" {
			data.ClusterSchedules[index].Cron = &CronScheduleModel{
				Minutes:  connection.FlattenTypesInt64List(record.Cron.Minutes),
				Hours:    connection.FlattenTypesInt64List(record.Cron.Hours),
				Days:     connection.FlattenTypesInt64List(record.Cron.Days),
				Weekdays: connection.FlattenTypesInt64List(record.Cron.Weekdays),
				Months:   connection.FlattenTypesInt64List(record.Cron.Months),
			}
		} else {
			data.ClusterSchedules[index].Interval = types.StringValue(record.Interval)
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
