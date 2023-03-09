package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ClusterScheduleResource{}

var _ resource.ResourceWithImportState = &ClusterScheduleResource{}

// NewClusterScheduleResource is a helper function to simplify the provider implementation.
func NewClusterScheduleResource() resource.Resource {
	return &ClusterScheduleResource{
		config: resourceOrDataSourceConfig{
			name: "cluster_schedule_resource",
		},
	}
}

// ClusterScheduleResource defines the resource implementation.
type ClusterScheduleResource struct {
	config resourceOrDataSourceConfig
}

// CronScheduleResourceModel describe the cron data model
type CronScheduleResourceModel struct {
	Minutes  []types.Int64 `tfsdk:"minutes"`
	Hours    []types.Int64 `tfsdk:"hours"`
	Days     []types.Int64 `tfsdk:"days"`
	Weekdays []types.Int64 `tfsdk:"weekdays"`
	Months   []types.Int64 `tfsdk:"months"`
}

// ClusterScheduleResourceModel describes the resource data model.
type ClusterScheduleResourceModel struct {
	CxProfileName types.String               `tfsdk:"cx_profile_name"`
	Name          types.String               `tfsdk:"name"`
	UUID          types.String               `tfsdk:"uuid"`
	Interval      types.String               `tfsdk:"interval"`
	Cron          *CronScheduleResourceModel `tfsdk:"cron"`
}

// Metadata returns the resource type name.
func (r *ClusterScheduleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *ClusterScheduleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cluster schedule resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the cluster schedule",
				Required:            true,
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster/Job schedule identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cron": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"minutes": schema.ListAttribute{
						ElementType:         types.Int64Type,
						MarkdownDescription: "List of cluster schedule minutes",
						Optional:            true,
					},
					"hours": schema.ListAttribute{
						ElementType:         types.Int64Type,
						MarkdownDescription: "List of cluster schedule hours",
						Optional:            true,
					},
					"days": schema.ListAttribute{
						ElementType:         types.Int64Type,
						MarkdownDescription: "List of cluster schedule days",
						Optional:            true,
					},
					"weekdays": schema.ListAttribute{
						ElementType:         types.Int64Type,
						MarkdownDescription: "List of cluster schedule weekdays",
						Optional:            true,
					},
					"months": schema.ListAttribute{
						ElementType:         types.Int64Type,
						MarkdownDescription: "List of cluster schedule months",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"interval": schema.StringAttribute{
				MarkdownDescription: "Cluster schedule interval",
				Optional:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ClusterScheduleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected  Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *ClusterScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClusterScheduleResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	restInfo, err := interfaces.GetClusterSchedule(errorHandler, *client, data.Name.ValueString())
	if err != nil {
		// error reporting done inside GetClusterSchedule
		return
	}
	// data.Name = types.StringValue(restInfo.Name)
	data.UUID = types.StringValue(restInfo.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *ClusterScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ClusterScheduleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	var body interfaces.ClusterScheduleResourceBodyDataModelONTAP
	body.Name = data.Name.ValueString()

	if !data.Interval.IsNull() {
		body.Interval = data.Interval.ValueString()
	} else {
		var minutes, hours, days, months, weekdays []int64
		for _, v := range data.Cron.Minutes {
			minutes = append(minutes, v.ValueInt64())
		}
		for _, v := range data.Cron.Hours {
			hours = append(hours, v.ValueInt64())
		}
		for _, v := range data.Cron.Weekdays {
			weekdays = append(weekdays, v.ValueInt64())
		}
		for _, v := range data.Cron.Days {
			days = append(days, v.ValueInt64())
		}
		for _, v := range data.Cron.Months {
			months = append(months, v.ValueInt64())
		}
		body.Cron.Minutes = minutes
		body.Cron.Hours = hours
		body.Cron.Weekdays = weekdays
		body.Cron.Days = days
		body.Cron.Months = months
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateClusterSchedule(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.UUID = types.StringValue(resource.UUID)

	tflog.Trace(ctx, fmt.Sprintf("created a resource, UUID=%s", data.UUID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ClusterScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ClusterScheduleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ClusterScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ClusterScheduleResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.UUID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "cluster_schedule UUID is null")
		return
	}

	err = interfaces.DeleteClusterSchedule(errorHandler, *client, data.UUID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ClusterScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
