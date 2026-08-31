package snapmirror

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &SnapmirrorResource{}
var _ resource.ResourceWithImportState = &SnapmirrorResource{}
var _ resource.ResourceWithModifyPlan = &SnapmirrorResource{}

// NewSnapmirrorResource is a helper function to simplify the provider implementation.
func NewSnapmirrorResource() resource.Resource {
	return &SnapmirrorResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "snapmirror",
		},
	}
}

// NewSnapmirrorResourceAlias is a helper function to simplify the provider implementation.
func NewSnapmirrorResourceAlias() resource.Resource {
	return &SnapmirrorResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "snapmirror_resource",
		},
	}
}

// SnapmirrorResource defines the resource implementation.
type SnapmirrorResource struct {
	config connection.ResourceOrDataSourceConfig
}

// SnapmirrorResourceModel describes the resource data model.
type SnapmirrorResourceModel struct {
	CxProfileName       types.String       `tfsdk:"cx_profile_name"`
	SourceEndPoint      *EndPoint          `tfsdk:"source_endpoint"`
	DestinationEndPoint *EndPoint          `tfsdk:"destination_endpoint"`
	CreateDestination   *CreateDestination `tfsdk:"create_destination"`
	Policy              *Policy            `tfsdk:"policy"`
	Initialize          types.Bool         `tfsdk:"initialize"`
	Force               types.Bool         `tfsdk:"force"`
	QuickResync         types.Bool         `tfsdk:"quick_resync"`
	TransferringTimeOut types.Int64        `tfsdk:"transferring_time_out"`
	Healthy             types.Bool         `tfsdk:"healthy"`
	State               types.String       `tfsdk:"state"`
	ID                  types.String       `tfsdk:"id"`
}

// EndPoint describes source/destination endpoint data model.
type EndPoint struct {
	Cluster *Cluster     `tfsdk:"cluster"`
	Path    types.String `tfsdk:"path"`
}

// CreateDestination describes CreateDestination data model.
type CreateDestination struct {
	Enabled types.Bool `tfsdk:"enabled"`
}

// Cluster describes Cluster data model.
type Cluster struct {
	Name types.String `tfsdk:"name"`
}

// Policy describes Policy data model.
type Policy struct {
	Name             types.String      `tfsdk:"name"`
	TransferSchedule *TransferSchedule `tfsdk:"transfer_schedule"`
}

// TransferSchedule describes the transfer schedule data model.
type TransferSchedule struct {
	Name types.String `tfsdk:"name"`
}

const (
	snapmirrorTransitionTimeout = 2 * time.Minute
	snapmirrorTransitionPoll    = 5 * time.Second
)

// Metadata returns the resource type name
func (r *SnapmirrorResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *SnapmirrorResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Snapmirror resource",
		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"source_endpoint": schema.SingleNestedAttribute{
				MarkdownDescription: "Snapmirror source endpoint",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"cluster": schema.SingleNestedAttribute{
						MarkdownDescription: "Cluster details",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "cluster name",
								Required:            true,
							},
						},
					},
					"path": schema.StringAttribute{
						MarkdownDescription: "Path to the source endpoint of the SnapMirror relationship",
						Required:            true,
					},
				},
			},
			"destination_endpoint": schema.SingleNestedAttribute{
				MarkdownDescription: "Snapmirror destination endpoint",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"cluster": schema.SingleNestedAttribute{
						MarkdownDescription: "Cluster details",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "cluster name",
								Required:            true,
							},
						},
					},
					"path": schema.StringAttribute{
						MarkdownDescription: "Path to the destination endpoint of the SnapMirror relationship",
						Required:            true,
					},
				},
			},
			"create_destination": schema.SingleNestedAttribute{
				MarkdownDescription: "Snapmirror provision destination",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable this property to provision the destination endpoint",
						Required:            true,
					},
				},
			},
			"initialize": schema.BoolAttribute{
				MarkdownDescription: "initialize the relationship",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"force": schema.BoolAttribute{
				MarkdownDescription: "If set to true while specifying state as broken_off, performs a forced failover overriding validation errors.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"quick_resync": schema.BoolAttribute{
				MarkdownDescription: "Optional modify-only flag. Set true to speed resync by not preserving storage efficiency; applicable for FlexVol and SVMDR when PATCH state changes to snapmirrored.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"transferring_time_out": schema.Int64Attribute{
				MarkdownDescription: "Maximum time in seconds to wait for SnapMirror state transitions before failing the operation.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(300),
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"healthy": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Set state to snapmirrored for async policies or in_sync for sync policies when initializing or resyncing a SnapMirror relationship.",
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("broken_off", "paused", "snapmirrored", "uninitialized", "in_sync", "out_of_sync", "synchronizing", "expanding", "shrinking"),
				},
			},
			"policy": schema.SingleNestedAttribute{
				MarkdownDescription: "policy details",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "policy name",
						Required:            true,
					},
					"transfer_schedule": schema.SingleNestedAttribute{
						MarkdownDescription: "transfer schedule details",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "schedule name",
								Required:            true,
							},
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// ModifyPlan suppresses one-shot flag diffs unless they are relevant for the current state transition.
func (r *SnapmirrorResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var plan, state *SnapmirrorResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	changed := false
	stateChanging := !plan.State.Equal(state.State)

	forceRelevant := stateChanging && plan.State.ValueString() == "broken_off"
	if !forceRelevant && !plan.Force.Equal(state.Force) {
		plan.Force = state.Force
		changed = true
	}

	quickResyncRelevant := stateChanging && state.State.ValueString() == "broken_off"
	if !quickResyncRelevant && !plan.QuickResync.Equal(state.QuickResync) {
		plan.QuickResync = state.QuickResync
		changed = true
	}

	if changed {
		resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
	}
}

// Configure adds the provider configured client to the resource.
func (r *SnapmirrorResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please resport this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *SnapmirrorResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SnapmirrorResourceModel

	// Read Terraform prior state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside New Client
		return
	}

	if data.ID.ValueString() != "" {
		restInfo, err := interfaces.GetSnapmirrorByID(errorHandler, *client, data.ID.ValueString())
		if err != nil {
			// error reporting done inside GetSnapmirrorByID
			return
		}
		data.ID = types.StringValue(restInfo.UUID)
		data.Healthy = types.BoolValue(restInfo.Healthy)
		data.State = types.StringValue(restInfo.State)
		if data.SourceEndPoint != nil && restInfo.Source.Path != "" {
			data.SourceEndPoint.Path = types.StringValue(restInfo.Source.Path)
		}
		if data.DestinationEndPoint != nil && restInfo.Destination.Path != "" {
			data.DestinationEndPoint.Path = types.StringValue(restInfo.Destination.Path)
		}
		if data.TransferringTimeOut.IsNull() || data.TransferringTimeOut.IsUnknown() {
			data.TransferringTimeOut = types.Int64Value(300)
		}
		if data.Force.IsNull() || data.Force.IsUnknown() {
			data.Force = types.BoolValue(false)
		}
		if data.QuickResync.IsNull() || data.QuickResync.IsUnknown() {
			data.QuickResync = types.BoolValue(false)
		}
		// only refresh policy fields if policy and policy.transfer_schedule were configured
		// and already exist in prior state
		if data.Policy != nil {
			if restInfo.Policy.Name != "" {
				data.Policy.Name = types.StringValue(restInfo.Policy.Name)
			}
			if data.Policy.TransferSchedule != nil {
				if restInfo.Policy.TransferSchedule != nil && restInfo.Policy.TransferSchedule.Name != "" {
					data.Policy.TransferSchedule = &TransferSchedule{
						Name: types.StringValue(restInfo.Policy.TransferSchedule.Name),
					}
				}
			}
		}
	} else {
		restInfoImport, err := interfaces.GetSnapmirrorByDestinationPath(errorHandler, *client, data.DestinationEndPoint.Path.ValueString(), nil)
		if err != nil {
			// error reporting done inside GetSnapmirrorByDestinationPath
			return
		}
		data.ID = types.StringValue(restInfoImport.UUID)
		data.Healthy = types.BoolValue(restInfoImport.Healthy)
		data.State = types.StringValue(restInfoImport.State)
		data.DestinationEndPoint.Path = types.StringValue(restInfoImport.Destination.Path)
		if data.TransferringTimeOut.IsNull() || data.TransferringTimeOut.IsUnknown() {
			data.TransferringTimeOut = types.Int64Value(300)
		}
		if data.Force.IsNull() || data.Force.IsUnknown() {
			data.Force = types.BoolValue(false)
		}
		if data.QuickResync.IsNull() || data.QuickResync.IsUnknown() {
			data.QuickResync = types.BoolValue(false)
		}
		// source_endpoint is a required attribute and is not part of the import ID,
		// so it has to be filled from the REST response for the imported state to be
		// usable. The source cluster is left unset on purpose: it is optional in the
		// configuration and filling it would create a permanent diff for the usual
		// path-only configurations.
		if data.SourceEndPoint == nil {
			data.SourceEndPoint = &EndPoint{}
		}
		data.SourceEndPoint.Path = types.StringValue(restInfoImport.Source.Path)
		// on import the prior state never carries a policy, fill it from the REST
		// response so the imported relationship keeps its policy in state.
		if restInfoImport.Policy.Name != "" {
			if data.Policy == nil {
				data.Policy = &Policy{}
			}
			data.Policy.Name = types.StringValue(restInfoImport.Policy.Name)
			if restInfoImport.Policy.TransferSchedule != nil && restInfoImport.Policy.TransferSchedule.Name != "" {
				data.Policy.TransferSchedule = &TransferSchedule{
					Name: types.StringValue(restInfoImport.Policy.TransferSchedule.Name),
				}
			}
		}
		// Set initialize to its default (true) so RequiresReplace does not trigger on first plan.
		if data.Initialize.IsNull() || data.Initialize.IsUnknown() {
			data.Initialize = types.BoolValue(true)
		}
		// initialize only matters at create time, but it is Computed with a default of
		// true and requires replacement on change. Leaving it null in the imported
		// state makes the next plan replace the relationship, so set the default here.
		if data.Initialize.IsNull() {
			data.Initialize = types.BoolValue(true)
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a snapmirror resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *SnapmirrorResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SnapmirrorResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.SnapmirrorResourceBodyDataModelONTAP

	if resp.Diagnostics.HasError() {
		return
	}
	if data.TransferringTimeOut.IsNull() || data.TransferringTimeOut.IsUnknown() {
		data.TransferringTimeOut = types.Int64Value(300)
	}

	body.SourceEndPoint.Path = data.SourceEndPoint.Path.ValueString()
	body.DestinationEndPoint.Path = data.DestinationEndPoint.Path.ValueString()
	if data.SourceEndPoint.Cluster != nil {
		if !data.SourceEndPoint.Cluster.Name.IsNull() {
			body.SourceEndPoint.Cluster.Name = data.SourceEndPoint.Cluster.Name.ValueString()
		}
	}
	if data.DestinationEndPoint.Cluster != nil {
		if !data.DestinationEndPoint.Cluster.Name.IsNull() {
			body.DestinationEndPoint.Cluster.Name = data.DestinationEndPoint.Cluster.Name.ValueString()
		}
	}
	if data.CreateDestination != nil {
		if !data.CreateDestination.Enabled.IsNull() {
			body.CreateDestination.Enabled = data.CreateDestination.Enabled.ValueBool()
		}
	}
	if data.Policy != nil {
		if !data.Policy.Name.IsNull() {
			body.Policy.Name = data.Policy.Name.ValueString()
		}
		if data.Policy.TransferSchedule != nil && !data.Policy.TransferSchedule.Name.IsNull() {
			body.Policy.TransferSchedule = &interfaces.TransferSchedule{
				Name: data.Policy.TransferSchedule.Name.ValueString(),
			}
		}
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateSnapmirror(errorHandler, *client, body)
	if err != nil {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("create snapmirror resource: %#v", resource))

	data.ID = types.StringValue(resource.UUID)
	restInfo, err := interfaces.GetSnapmirrorByID(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		// error reporting done inside GetSnapmirror
		return
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read snapmirror info: %#v", restInfo))
	data.Healthy = types.BoolValue(restInfo.Healthy)
	data.State = types.StringValue(restInfo.State)

	if data.Initialize.ValueBool() && data.State.ValueString() == "uninitialized" {
		time.Sleep(3 * time.Second)
		err := interfaces.InitializeSnapmirror(errorHandler, *client, data.ID.ValueString(), "snapmirrored")
		if err != nil {
			// error reporting done inside InitializeSnapmirror
			return
		}
		// Poll until ONTAP reaches snapmirrored or the transferring_time_out elapses.
		transitionTimeout := snapmirrorTransitionTimeout
		if !data.TransferringTimeOut.IsNull() && !data.TransferringTimeOut.IsUnknown() {
			transitionTimeout = time.Duration(data.TransferringTimeOut.ValueInt64()) * time.Second
		}
		deadline := time.Now().Add(transitionTimeout)
		for {
			restInfo, err = interfaces.GetSnapmirrorByID(errorHandler, *client, data.ID.ValueString())
			if err != nil {
				return
			}
			tflog.Debug(errorHandler.Ctx, fmt.Sprintf("waiting for snapmirror snapmirrored state, current: %s", restInfo.State))
			if restInfo.State == "snapmirrored" {
				break
			}
			if time.Now().After(deadline) {
				errorHandler.MakeAndReportError("timeout waiting for snapmirror initialization",
					fmt.Sprintf("expected state snapmirrored, got %s after %s", restInfo.State, transitionTimeout))
				return
			}
			time.Sleep(snapmirrorTransitionPoll)
		}
		data.Healthy = types.BoolValue(restInfo.Healthy)
		data.State = types.StringValue(restInfo.State)
	} else {
		restInfo, err = interfaces.GetSnapmirrorByID(errorHandler, *client, data.ID.ValueString())
		if err != nil {
			return
		}
		data.Healthy = types.BoolValue(restInfo.Healthy)
		data.State = types.StringValue(restInfo.State)
	}
	data.ID = types.StringValue(resource.UUID)

	tflog.Trace(ctx, fmt.Sprintf("created a snapmirror resource, UUID=%s", data.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SnapmirrorResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config *SnapmirrorResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Read Terraform state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// Read Terraform configuration (used for write-only inputs like quick_resync)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, state.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Update the resource
	var body interfaces.UpdateSnapmirrorResourceBodyDataModelONTAP

	if !plan.SourceEndPoint.Path.Equal(state.SourceEndPoint.Path) {
		body.SourceEndPoint = &interfaces.EndPoint{Path: plan.SourceEndPoint.Path.ValueString()}
		if plan.SourceEndPoint.Cluster != nil && !plan.SourceEndPoint.Cluster.Name.IsNull() {
			body.SourceEndPoint.Cluster.Name = plan.SourceEndPoint.Cluster.Name.ValueString()
		}
	}
	if !plan.DestinationEndPoint.Path.Equal(state.DestinationEndPoint.Path) {
		body.DestinationEndPoint = &interfaces.EndPoint{Path: plan.DestinationEndPoint.Path.ValueString()}
		if plan.DestinationEndPoint.Cluster != nil && !plan.DestinationEndPoint.Cluster.Name.IsNull() {
			body.DestinationEndPoint.Cluster.Name = plan.DestinationEndPoint.Cluster.Name.ValueString()
		}
	}
	if plan.Policy != nil {
		body.Policy = &interfaces.PolicySnapmirror{}
		if !plan.Policy.Name.IsNull() {
			body.Policy.Name = plan.Policy.Name.ValueString()
		}
		if plan.Policy.TransferSchedule != nil && !plan.Policy.TransferSchedule.Name.IsNull() {
			body.Policy.TransferSchedule = &interfaces.TransferSchedule{
				Name: plan.Policy.TransferSchedule.Name.ValueString(),
			}
		}
	}

	stateChanging := !plan.State.Equal(state.State)
	if stateChanging {
		body.State = plan.State.ValueString()
	}

	// force is passed as a query param only when breaking the relationship and user explicitly set force=true.
	useForce := stateChanging && plan.State.ValueString() == "broken_off" && plan.Force.ValueBool()
	// quick_resync is only relevant when moving a relationship from broken_off back to snapmirrored.
	quickResync := stateChanging &&
		state.State.ValueString() == "broken_off" &&
		!config.QuickResync.IsNull() &&
		!config.QuickResync.IsUnknown() &&
		config.QuickResync.ValueBool()
	err = interfaces.UpdateSnapmirror(errorHandler, *client, body, plan.ID.ValueString(), useForce, quickResync)
	if err != nil {
		return
	}

	if stateChanging {
		desiredState := plan.State.ValueString()
		transitionTimeout := snapmirrorTransitionTimeout
		if !plan.TransferringTimeOut.IsNull() && !plan.TransferringTimeOut.IsUnknown() {
			transitionTimeout = time.Duration(plan.TransferringTimeOut.ValueInt64()) * time.Second
		}
		deadline := time.Now().Add(transitionTimeout)
		for {
			restInfo, err := interfaces.GetSnapmirrorByID(errorHandler, *client, plan.ID.ValueString())
			if err != nil {
				return
			}
			if restInfo.State == desiredState {
				break
			}
			if time.Now().After(deadline) {
				errorHandler.MakeAndReportError("timeout waiting for snapmirror state transition", fmt.Sprintf("expected state %s, got %s", desiredState, restInfo.State))
				return
			}
			time.Sleep(snapmirrorTransitionPoll)
		}
	}

	restInfo, err := interfaces.GetSnapmirrorByID(errorHandler, *client, plan.ID.ValueString())
	if err != nil {
		return
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read snapmirror info: %#v", restInfo))
	// Update the computed parameters
	plan.Healthy = types.BoolValue(restInfo.Healthy)
	plan.State = types.StringValue(restInfo.State)
	if plan.SourceEndPoint != nil && restInfo.Source.Path != "" {
		plan.SourceEndPoint.Path = types.StringValue(restInfo.Source.Path)
	}
	if plan.DestinationEndPoint != nil && restInfo.Destination.Path != "" {
		plan.DestinationEndPoint.Path = types.StringValue(restInfo.Destination.Path)
	}

	tflog.Debug(ctx, fmt.Sprintf("updated a snapmirror resource: UUID=%s", plan.ID))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SnapmirrorResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SnapmirrorResourceModel

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
		errorHandler.MakeAndReportError("UUID is null", "snapmirror UUID is null")
		return
	}

	err = interfaces.DeleteSnapmirror(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SnapmirrorResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: destination_path,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("destination_endpoint").AtName("path"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[1])...)
}
