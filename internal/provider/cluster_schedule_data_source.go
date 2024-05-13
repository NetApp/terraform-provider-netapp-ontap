package provider

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
var _ datasource.DataSource = &ClusterScheduleDataSource{}

// NewClusterScheduleDataSource is a helper function to simplify the provider implementation.
func NewClusterScheduleDataSource() datasource.DataSource {
	return &ClusterScheduleDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cluster_schedule_data_source",
		},
	}
}

// ClusterScheduleDataSource defines the data source implementation.
type ClusterScheduleDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ClusterScheduleDataSourceModel describes the data source data model.
type ClusterScheduleDataSourceModel struct {
	CxProfileName types.String       `tfsdk:"cx_profile_name"`
	Name          types.String       `tfsdk:"name"`
	ID            types.String       `tfsdk:"id"`
	Type          types.String       `tfsdk:"type"`
	Interval      types.String       `tfsdk:"interval"`
	Scope         types.String       `tfsdk:"scope"`
	Cron          *CronScheduleModel `tfsdk:"cron"`
}

// CronScheduleModel describe the cron data model
type CronScheduleModel struct {
	Minutes  []types.Int64 `tfsdk:"minutes"`
	Hours    []types.Int64 `tfsdk:"hours"`
	Days     []types.Int64 `tfsdk:"days"`
	Weekdays []types.Int64 `tfsdk:"weekdays"`
	Months   []types.Int64 `tfsdk:"months"`
}

// Metadata returns the data source type name.
func (d *ClusterScheduleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ClusterScheduleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cluster Schedule data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Schedule name",
				Required:            true,
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *ClusterScheduleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ClusterScheduleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterScheduleDataSourceModel

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
	restInfo, err := interfaces.GetClusterSchedule(errorHandler, *client, data.Name.ValueString())
	if err != nil {
		// error reporting done inside GetClusterSchedule
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No Cluster Schedule found", fmt.Sprintf("Cluster Schedule %s not found.", data.Name.ValueString()))
		return
	}
	data.Name = types.StringValue(restInfo.Name)
	data.ID = types.StringValue(restInfo.UUID)
	data.Type = types.StringValue(restInfo.Type)
	data.Scope = types.StringValue(restInfo.Scope)

	if restInfo.Type == "cron" {
		data.Cron = &CronScheduleModel{
			Minutes:  connection.FlattenTypesInt64List(restInfo.Cron.Minutes),
			Hours:    connection.FlattenTypesInt64List(restInfo.Cron.Hours),
			Days:     connection.FlattenTypesInt64List(restInfo.Cron.Days),
			Weekdays: connection.FlattenTypesInt64List(restInfo.Cron.Weekdays),
			Months:   connection.FlattenTypesInt64List(restInfo.Cron.Months),
		}
	} else {
		data.Interval = types.StringValue(restInfo.Interval)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
