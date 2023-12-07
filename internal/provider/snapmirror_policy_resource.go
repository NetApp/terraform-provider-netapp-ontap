package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/mitchellh/mapstructure"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &SnapmirrorPolicyResource{}
var _ resource.ResourceWithImportState = &SnapmirrorPolicyResource{}

// NewSnapmirrorPolicyResource is a helper function to simplify the provider implementation.
func NewSnapmirrorPolicyResource() resource.Resource {
	return &SnapmirrorPolicyResource{
		config: resourceOrDataSourceConfig{
			name: "snapmirror_policy_resource",
		},
	}
}

// SnapmirrorPolicyResource defines the resource implementation.
type SnapmirrorPolicyResource struct {
	config resourceOrDataSourceConfig
}

// SnapmirrorPolicyResourceModel describes the resource data model.
type SnapmirrorPolicyResourceModel struct {
	CxProfileName             types.String     `tfsdk:"cx_profile_name"`
	Name                      types.String     `tfsdk:"name"`
	SVMName                   types.String     `tfsdk:"svm_name"`
	Type                      types.String     `tfsdk:"type"`
	SyncType                  types.String     `tfsdk:"sync_type"`
	Comment                   types.String     `tfsdk:"comment"`
	TransferScheduleName      types.String     `tfsdk:"transfer_schedule_name"`
	NetworkCompressionEnabled types.Bool       `tfsdk:"network_compression_enabled"`
	Retention                 []RetentionModel `tfsdk:"retention"`
	IdentityPreservation      types.String     `tfsdk:"identity_preservation"`
	CopyAllSourceSnapshots    types.Bool       `tfsdk:"copy_all_source_snapshots"`
	CopyLatestSourceSnapshot  types.Bool       `tfsdk:"copy_latest_source_snapshot"`
	CreateSnapshotOnSource    types.Bool       `tfsdk:"create_snapshot_on_source"`
	ID                        types.String     `tfsdk:"id"`
}

// RetentionModel describes retention data model.
type RetentionModel struct {
	CreationScheduleName types.String `tfsdk:"creation_schedule_name"`
	Count                types.Int64  `tfsdk:"count"`
	Label                types.String `tfsdk:"label"`
	Prefix               types.String `tfsdk:"prefix"`
}

// Metadata returns the resource type name
func (r *SnapmirrorPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *SnapmirrorPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SnapmirrorPolicy resource",
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
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "SnapmirrorPolicy type. [async, sync, continuous]",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"async", "sync", "continuous"}...),
				},
			},
			"sync_type": schema.StringAttribute{
				MarkdownDescription: "SnapmirrorPolicy sync type. [sync, strict_sync, automated_failover]",
				Optional:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"sync", "strict_sync", "automated_failover"}...),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment associated with the policy.",
				Optional:            true,
			},
			"transfer_schedule_name": schema.StringAttribute{
				MarkdownDescription: "The schedule used to update asynchronous relationships",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("sync_type"),
					}...),
				},
			},
			"network_compression_enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether network compression is enabled for transfers",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"retention": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Rules for Snapshot copy retention.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"creation_schedule_name": schema.StringAttribute{
							MarkdownDescription: "Schedule used to create Snapshot copies on the destination for long term retention.",
							Optional:            true,
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.Expressions{
									path.MatchRoot("sync_type"),
								}...),
							},
						},
						"count": schema.Int64Attribute{
							MarkdownDescription: "Number of Snapshot copies to be kept for retention.",
							Optional:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: "Snapshot copy label",
							Required:            true,
						},
						"prefix": schema.StringAttribute{
							MarkdownDescription: "Specifies the prefix for the Snapshot copy name to be created as per the schedule",
							Optional:            true,
							Computed:            true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
							Validators: []validator.String{
								stringvalidator.ConflictsWith(path.Expressions{
									path.MatchRoot("sync_type"),
								}...),
							},
						},
					},
				},
			},
			"identity_preservation": schema.StringAttribute{
				MarkdownDescription: "Specifies which configuration of the source SVM is replicated to the destination SVM.",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"full", "exclude_network_config", "exclude_network_and_protocol_config"}...),
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("sync_type"),
					}...),
				},
			},
			"copy_all_source_snapshots": schema.BoolAttribute{
				MarkdownDescription: "Specifies that all the source Snapshot copies (including the one created by SnapMirror before the transfer begins) should be copied to the destination on a transfer.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("create_snapshot_on_source"),
						path.MatchRoot("copy_latest_source_snapshot"),
					}...),
				},
			},
			"copy_latest_source_snapshot": schema.BoolAttribute{
				MarkdownDescription: "Specifies that the latest source Snapshot copy (created by SnapMirror before the transfer begins) should be copied to the destination on a transfer. 'Retention' properties cannot be specified along with this property. This is applicable only to async policies. Property can only be set to 'true'.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"create_snapshot_on_source": schema.BoolAttribute{
				MarkdownDescription: "Specifies that all the source Snapshot copies (including the one created by SnapMirror before the transfer begins) should be copied to the destination on a transfer.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("copy_all_source_snapshots"),
						path.MatchRoot("copy_latest_source_snapshot"),
					}...),
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

// Configure adds the provider configured client to the resource.
func (r *SnapmirrorPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please resport this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *SnapmirrorPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SnapmirrorPolicyResourceModel

	// Read Terraform prior state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside New Client
		return
	}

	var restInfo *interfaces.SnapmirrorPolicyGetRawDataModelONTAP
	if data.ID.ValueString() != "" {
		restInfo, err = interfaces.GetSnapmirrorPolicy(errorHandler, *client, data.ID.ValueString())
	} else {
		restInfo, err = interfaces.GetSnapmirrorPolicyByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	}
	if err != nil {
		// error reporting done inside GETSnapmirrorPolicy
		return
	}

	if restInfo.TransferSchedule.Name != "" {
		data.TransferScheduleName = types.StringValue(restInfo.TransferSchedule.Name)
	}

	data.Type = types.StringValue(restInfo.Type)
	if restInfo.SyncType != "" {
		data.SyncType = types.StringValue(restInfo.SyncType)
	}
	if restInfo.Comment != "" {
		data.Comment = types.StringValue(restInfo.Comment)
	}
	if restInfo.IdentityPreservation != "" {
		data.IdentityPreservation = types.StringValue(restInfo.IdentityPreservation)
	}
	data.CopyAllSourceSnapshots = types.BoolValue(restInfo.CopyAllSourceSnapshots)
	data.NetworkCompressionEnabled = types.BoolValue(restInfo.NetworkCompressionEnabled)
	data.CopyLatestSourceSnapshot = types.BoolValue(restInfo.CopyLatestSourceSnapshot)
	data.CreateSnapshotOnSource = types.BoolValue(restInfo.CreateSnapshotOnSource)
	data.ID = types.StringValue(restInfo.UUID)

	// if len(restInfo.Retention) == 0 {
	if restInfo.Retention == nil {
		data.Retention = nil
	} else {
		data.Retention = []RetentionModel{}
		for _, item := range restInfo.Retention {
			var retention RetentionModel
			// conver count from string to int
			count, err := strconv.Atoi(item.Count)
			if err != nil {
				errorHandler.MakeAndReportError("Decode count error", "snapmirror_policy retention count is not valid")
				return
			}
			retention.Count = types.Int64Value(int64(count))
			if item.CreationSchedule.Name != "" {
				retention.CreationScheduleName = types.StringValue(item.CreationSchedule.Name)
			}
			if item.Label != "" {
				retention.Label = types.StringValue(item.Label)
			}
			if item.Prefix != "" {
				retention.Prefix = types.StringValue(item.Prefix)
			}
			data.Retention = append(data.Retention, retention)
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a snapmirror policy resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *SnapmirrorPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SnapmirrorPolicyResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.SnapmirrorPolicyResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()
	if !data.IdentityPreservation.IsNull() {
		body.IdentityPreservation = data.IdentityPreservation.ValueString()
	}
	if !data.Comment.IsNull() {
		body.Comment = data.Comment.ValueString()
	}
	if !data.CopyAllSourceSnapshots.IsNull() {
		body.CopyAllSourceSnapshots = data.CopyAllSourceSnapshots.ValueBool()
	}
	if !data.NetworkCompressionEnabled.IsNull() {
		body.NetworkCompressionEnabled = data.NetworkCompressionEnabled.ValueBool()
	}
	if !data.TransferScheduleName.IsNull() {
		body.TransferSchedule.Name = data.TransferScheduleName.ValueString()
	}
	if !data.Type.IsNull() {
		body.Type = data.Type.ValueString()
	}
	if !data.SyncType.IsNull() {
		body.SyncType = data.SyncType.ValueString()
	}
	if !data.CopyLatestSourceSnapshot.IsNull() {
		body.CopyLatestSourceSnapshot = data.CopyLatestSourceSnapshot.ValueBool()
	}
	if !data.CreateSnapshotOnSource.IsNull() {
		body.CreateSnapshotOnSource = data.CreateSnapshotOnSource.ValueBool()
	}

	if data.Retention == nil {
		body.Retention = nil
	} else {
		retention := []interfaces.RetentionGetDataModel{}
		for _, item := range data.Retention {
			var aRetention interfaces.RetentionGetDataModel
			aRetention.Count = item.Count.ValueInt64()
			if !item.CreationScheduleName.IsNull() {
				aRetention.CreationSchedule.Name = item.CreationScheduleName.ValueString()
			}
			if !item.Label.IsNull() {
				aRetention.Label = item.Label.ValueString()
			}
			if !item.Prefix.IsNull() {
				aRetention.Prefix = item.Prefix.ValueString()
			}
			retention = append(retention, aRetention)
		}
		err := mapstructure.Decode(retention, &body.Retention)
		if err != nil {
			errorHandler.MakeAndReportError("error creating snapshot policies", fmt.Sprintf("error on encoding copies info: %s, copies %#v", err, retention))
			return
		}
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateSnapmirrorPolicy(errorHandler, *client, body)
	if err != nil {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("create snapmirror policy get resource: %#v", resource))
	// Update the computed parameters
	data.ID = types.StringValue(resource.UUID)
	if resource.Retention == nil {
		data.Retention = nil
		tflog.Debug(ctx, fmt.Sprintf("create snapmirror policy retention is nil: %#v", data.Retention))
	} else {
		data.Retention = []RetentionModel{}
		for _, item := range resource.Retention {
			var retention RetentionModel
			// conver count from string to int
			count, err := strconv.Atoi(item.Count)
			if err != nil {
				errorHandler.MakeAndReportError("decode count error", "snapmirror_policy retention count is not valid")
				return
			}
			retention.Count = types.Int64Value(int64(count))
			if item.CreationSchedule.Name != "" {
				retention.CreationScheduleName = types.StringValue(item.CreationSchedule.Name)
			}
			if item.Label != "" {
				retention.Label = types.StringValue(item.Label)
			}
			if item.Prefix != "" {
				retention.Prefix = types.StringValue(item.Prefix)
			}
			data.Retention = append(data.Retention, retention)
		}
	}
	data.Type = types.StringValue(resource.Type)

	tflog.Trace(ctx, fmt.Sprintf("created a snapmirror policy resource, UUID=%s", data.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SnapmirrorPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SnapmirrorPolicyResourceModel
	var state SnapmirrorPolicyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Read Terraform state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	client, err := getRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	// sync type -
	// not support: transfer_schedule_name, retention.prefix, retention.creation_schedule_name, identity_preservation
	// max count of retention is 1
	// modify retention is not allowed
	if !plan.SyncType.IsNull() {
		var body interfaces.UpdateSyncSnapmirrorPolicyResourceBodyDataModelONTAP
		body.Comment = plan.Comment.ValueString()
		body.NetworkCompressionEnabled = plan.NetworkCompressionEnabled.ValueBool()
		// The policy properties "copy_all_source_snapshots", "copy_latest_source_snapshot", and "create_snapshot_on_source" cannot be modified.
		if !plan.CopyAllSourceSnapshots.Equal(state.CopyAllSourceSnapshots) ||
			!plan.CopyLatestSourceSnapshot.Equal(state.CopyLatestSourceSnapshot) ||
			!plan.CreateSnapshotOnSource.Equal(state.CreateSnapshotOnSource) {
			errorHandler.MakeAndReportError("error updating snapshot policies",
				"error copy_all_source_snapshots, copy_latest_source_snapshot, and create_snapshot_on_sourc cannot be modified")
			return
		}
		if len(plan.Retention) == 0 && len(state.Retention) != 0 && !state.CreateSnapshotOnSource.ValueBool() {
			errorHandler.MakeAndReportError("error updating snapshot policies",
				"error deleting all retention rules of a policy that has the create_snapshot_on_source is set to false is not supported.")
			return
		}
		if plan.Retention == nil {
			body.Retention = nil
		} else {
			// add a retention
			if len(state.Retention) == 0 && len(plan.Retention) == 1 {
				retention := []interfaces.RetentionGetDataModel{}
				for _, item := range plan.Retention {
					var aRetention interfaces.RetentionGetDataModel
					aRetention.Count = item.Count.ValueInt64()
					aRetention.Label = item.Label.ValueString()

					retention = append(retention, aRetention)
				}
				err := mapstructure.Decode(retention, &body.Retention)
				if err != nil {
					errorHandler.MakeAndReportError("error updating snapshot policie in sync", fmt.Sprintf("error on encoding copies info: %s, copies %#v", err, retention))
					return
				}
			} else if len(state.Retention) == 1 && len(plan.Retention) == 1 {
				tflog.Debug(ctx, fmt.Sprintf("update snapmirror policy retention is not allowed, so keep the original one %#v. plan:%#v", state.Retention, plan.Retention))
			} else {
				errorHandler.MakeAndReportError("error updating sync snapshot policies",
					"error modifying retention rule of a policy is not supported.")
				return
			}
		}
		err = interfaces.UpdateSnapmirrorPolicy(errorHandler, *client, body, plan.ID.ValueString())
		if err != nil {
			return
		}
	} else { // async or continuous
		var body interfaces.UpdateSnapmirrorPolicyResourceBodyDataModelONTAP
		body.Comment = plan.Comment.ValueString()
		body.NetworkCompressionEnabled = plan.NetworkCompressionEnabled.ValueBool()
		body.IdentityPreservation = plan.IdentityPreservation.ValueString()

		if !plan.TransferScheduleName.IsNull() && plan.TransferScheduleName.ValueString() != "" {
			transferschedule := interfaces.UpdateTransferScheduleType{}
			transferschedule.Name = plan.TransferScheduleName.ValueString()
			err := mapstructure.Decode(transferschedule, &body.TransferSchedule)
			if err != nil {
				errorHandler.MakeAndReportError("error updating snapshot policies", fmt.Sprintf("error on encoding transfer_schedule info: %s, transferschedule %#v", err, transferschedule))
				return
			}
		} else {
			body.TransferSchedule = nil
		}

		// tflog.Debug(ctx, fmt.Sprintf("Call update plan trasfer name:%#v, state:%#v", plan.TransferScheduleName, state.TransferScheduleName))
		// The policy properties "copy_all_source_snapshots", "copy_latest_source_snapshot", and "create_snapshot_on_source" cannot be modified.
		if !plan.CopyAllSourceSnapshots.Equal(state.CopyAllSourceSnapshots) ||
			!plan.CopyLatestSourceSnapshot.Equal(state.CopyLatestSourceSnapshot) ||
			!plan.CreateSnapshotOnSource.Equal(state.CreateSnapshotOnSource) {
			errorHandler.MakeAndReportError("error updating snapshot policies",
				"error copy_all_source_snapshots, copy_latest_source_snapshot, and create_snapshot_on_sourc cannot be modified")
			return
		}

		if len(plan.Retention) == 0 && len(state.Retention) != 0 && !state.CreateSnapshotOnSource.ValueBool() {
			errorHandler.MakeAndReportError("error updating snapshot policies",
				"error deleting all retention rules of a policy that has the create_snapshot_on_source is set to false is not supported.")
			return
		}
		if plan.Retention == nil {
			body.Retention = nil
		} else {
			retention := []interfaces.RetentionGetDataModel{}
			for _, item := range plan.Retention {
				var aRetention interfaces.RetentionGetDataModel
				aRetention.Count = item.Count.ValueInt64()
				aRetention.Label = item.Label.ValueString()
				aRetention.Prefix = item.Prefix.ValueString()
				aRetention.CreationSchedule.Name = item.CreationScheduleName.ValueString()

				retention = append(retention, aRetention)
			}
			err := mapstructure.Decode(retention, &body.Retention)
			if err != nil {
				errorHandler.MakeAndReportError("error updating snapshot policies", fmt.Sprintf("error on encoding copies info: %s, copies %#v", err, retention))
				return
			}
		}
		err = interfaces.UpdateSnapmirrorPolicy(errorHandler, *client, body, plan.ID.ValueString())
		if err != nil {
			return
		}
	}

	restInfo, err := interfaces.GetSnapmirrorPolicy(errorHandler, *client, plan.ID.ValueString())
	if err != nil {
		// error reporting done inside GETSnapmirrorPolicy
		return
	}

	if restInfo.Retention == nil {
		plan.Retention = nil
	} else {
		plan.Retention = []RetentionModel{}
		for _, item := range restInfo.Retention {
			var retention RetentionModel
			// conver count from string to int
			count, err := strconv.Atoi(item.Count)
			if err != nil {
				errorHandler.MakeAndReportError("decode count error", "snapmirror_policy retention count is not valid")
				return
			}
			retention.Count = types.Int64Value(int64(count))
			if item.CreationSchedule.Name != "" {
				retention.CreationScheduleName = types.StringValue(item.CreationSchedule.Name)
			}
			if item.Label != "" {
				retention.Label = types.StringValue(item.Label)
			}
			if item.Prefix != "" {
				retention.Prefix = types.StringValue(item.Prefix)
			}
			plan.Retention = append(plan.Retention, retention)
		}
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a snapmirror policy resource, UUID=%s", plan.ID))
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SnapmirrorPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SnapmirrorPolicyResourceModel

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

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "snapmirror_policy UUID is null")
		return
	}

	err = interfaces.DeleteSnapmirrorPolicy(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SnapmirrorPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
}
