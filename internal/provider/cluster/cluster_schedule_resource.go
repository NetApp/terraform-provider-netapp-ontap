package cluster

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
		config: connection.ResourceOrDataSourceConfig{
			Name: "cluster_schedule",
		},
	}
}

// ClusterScheduleResource defines the resource implementation.
type ClusterScheduleResource struct {
	config connection.ResourceOrDataSourceConfig
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
	ID            types.String               `tfsdk:"id"`
	Interval      types.String               `tfsdk:"interval"`
	Cron          *CronScheduleResourceModel `tfsdk:"cron"`
}

// Metadata returns the resource type name.
func (r *ClusterScheduleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
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
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster/Job schedule identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cron": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"minutes": schema.SetAttribute{
						ElementType:         types.Int64Type,
						MarkdownDescription: "List of cluster schedule minutes",
						Optional:            true,
					},
					"hours": schema.SetAttribute{
						ElementType:         types.Int64Type,
						MarkdownDescription: "List of cluster schedule hours",
						Optional:            true,
					},
					"days": schema.SetAttribute{
						ElementType:         types.Int64Type,
						MarkdownDescription: "List of cluster schedule days",
						Optional:            true,
					},
					"weekdays": schema.SetAttribute{
						ElementType:         types.Int64Type,
						MarkdownDescription: "List of cluster schedule weekdays",
						Optional:            true,
					},
					"months": schema.SetAttribute{
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
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("cron"),
					}...),
				},
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
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected  Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.ProviderConfig = config
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
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	var restInfo *interfaces.ClusterScheduleGetDataModelONTAP
	if data.ID.ValueString() == "" {
		restInfo, err = interfaces.GetClusterScheduleByName(errorHandler, *client, data.Name.ValueString())
		if err != nil {
			// error reporting done inside GetClusterScheduleByName
			return
		}
		data.ID = types.StringValue(restInfo.UUID)
	} else {
		restInfo, err = interfaces.GetClusterSchedule(errorHandler, *client, data.ID.ValueString())
		if err != nil {
			// error reporting done inside GetClusterSchedule
			return
		}
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No Cluster Schedule found", fmt.Sprintf("Cluster Schedule %s not found.", data.Name.ValueString()))
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("Restinfo a cluster schedule job data source: %#v", restInfo))
	if restInfo.Interval == "" {
		tflog.Debug(ctx, fmt.Sprintf("Cron: %#v", restInfo.Cron))
		var cron CronScheduleResourceModel

		if restInfo.Cron.Days != nil {
			for _, v := range restInfo.Cron.Days {
				cron.Days = append(cron.Days, types.Int64Value(v))
			}
		}
		if restInfo.Cron.Hours != nil {
			for _, v := range restInfo.Cron.Hours {
				cron.Hours = append(cron.Hours, types.Int64Value(v))
			}
		}
		if restInfo.Cron.Minutes != nil {
			var minutes []types.Int64
			for _, v := range restInfo.Cron.Minutes {
				minutes = append(minutes, types.Int64Value(v))
			}
			cron.Minutes = minutes
		}
		if restInfo.Cron.Months != nil {
			for _, v := range restInfo.Cron.Months {
				cron.Months = append(cron.Months, types.Int64Value(v))
			}
		}
		if restInfo.Cron.Weekdays != nil {
			for _, v := range restInfo.Cron.Weekdays {
				cron.Weekdays = append(cron.Weekdays, types.Int64Value(v))
			}
		}
		data.Cron = &cron
	} else {
		data.Interval = types.StringValue(restInfo.Interval)
	}
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

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateClusterSchedule(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.ID = types.StringValue(resource.UUID)

	tflog.Trace(ctx, fmt.Sprintf("created a resource, UUID=%s", data.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ClusterScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ClusterScheduleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	var request interfaces.ClusterScheduleResourceBodyDataModelONTAP
	if !data.Interval.IsNull() {
		request.Interval = data.Interval.ValueString()
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
		request.Cron.Minutes = minutes
		request.Cron.Hours = hours
		request.Cron.Weekdays = weekdays
		request.Cron.Days = days
		request.Cron.Months = months
	}

	err = interfaces.UpdateClusterSchedule(errorHandler, *client, request, data.ID.ValueString())
	if err != nil {
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
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "cluster_schedule UUID is null")
		return
	}

	err = interfaces.DeleteClusterSchedule(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ClusterScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a scluster schedule resource: %#v", req))
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[1])...)
}
