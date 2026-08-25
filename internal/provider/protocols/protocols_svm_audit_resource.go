package protocols

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsSVMAuditResource{}
var _ resource.ResourceWithImportState = &ProtocolsSVMAuditResource{}
var _ resource.ResourceWithModifyPlan = &ProtocolsSVMAuditResource{}

// NewProtocolsSVMAuditResource is a helper function to simplify the provider implementation.
func NewProtocolsSVMAuditResource() resource.Resource {
	return &ProtocolsSVMAuditResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "svm_audit",
		},
	}
}

// ProtocolsSVMAuditResource defines the resource implementation.
type ProtocolsSVMAuditResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsSVMAuditResourceModel describes the resource data model.
type ProtocolsSVMAuditResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	Events        types.Object `tfsdk:"events"`
	Guarantee     types.Bool   `tfsdk:"guarantee"`
	Log           types.Object `tfsdk:"log"`
	LogPath		  types.String `tfsdk:"log_path"`
	ID            types.String `tfsdk:"id"`
}

// EventsResourceModel describes the events model.
type EventsResourceModel struct {
	AuthorizationPolicy types.Bool `tfsdk:"authorization_policy"`
	CapStaging 		    types.Bool `tfsdk:"cap_staging"`
	CIFSLogonLogoff     types.Bool `tfsdk:"cifs_logon_logoff"`
	FileOperations      types.Bool `tfsdk:"file_operations"`
	FileShare           types.Bool `tfsdk:"file_share"`
	SecurityGroup		types.Bool `tfsdk:"security_group"`
	UserAccount         types.Bool `tfsdk:"user_account"`
}

// LogResourceModel describes the log model.
type LogResourceModel struct {
	Format    types.String           `tfsdk:"format"`
	Retention RetentionResourceModel `tfsdk:"retention"`
	Rotation  RotationResourceModel  `tfsdk:"rotation"`
}

// RetentionResourceModel describes the retention model.
type RetentionResourceModel struct {
	Count    types.Int64  `tfsdk:"count"`
	Duration types.String `tfsdk:"duration"`
}

// RotationResourceModel describes the rotation model.
type RotationResourceModel struct {
	Size     types.Int64            `tfsdk:"size"`
	Schedule *ScheduleResourceModel `tfsdk:"schedule"`
}

// ScheduleResourceModel describes the schedule model.
type ScheduleResourceModel struct {
	Days     []types.Int64 `tfsdk:"days"`
	Hours    []types.Int64 `tfsdk:"hours"`
	Minutes  []types.Int64 `tfsdk:"minutes"`
	Months   []types.Int64 `tfsdk:"months"`
	Weekdays []types.Int64 `tfsdk:"weekdays"`
}

// converts Terraform model values []types.Int64 to API payload values []int64
// used when sending create/update requests
func int64SliceFromTypesInt64Slice(values []types.Int64) []int64 {
	converted := make([]int64, 0, len(values))
	for _, value := range values {
		converted = append(converted, value.ValueInt64())
	}
	return converted
}

// converts API response values []int64 to Terraform state list values (types.List...) and preserves nil as null
func int64ListValueFromOptionalSlice(values []int64) (basetypes.ListValue, diag.Diagnostics) {
	if values == nil {
		return types.ListNull(types.Int64Type), nil
	}

	converted := make([]attr.Value, 0, len(values))
	for _, value := range values {
		converted = append(converted, types.Int64Value(value))
	}

	return types.ListValue(types.Int64Type, converted)
}

func typesInt64SlicesEqual(left []types.Int64, right []types.Int64) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if !left[i].Equal(right[i]) {
			return false
		}
	}
	return true
}

func RequiresRetry(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(
		err.Error(),
		"Reason: Final consolidation is in progress. Retry after sometime.",
	)
}

func DeleteSVMAuditConfigWithRetry(ctx context.Context, errorHandler *utils.ErrorHandler, client restclient.RestClient, svmUUID string) error {
	delay := 10 * time.Second
	retries := 3
	for attempt := 1; attempt <= retries; attempt++ {
		err := interfaces.DeleteSVMAuditConfig(errorHandler, client, svmUUID)
		if err == nil {
			// error reporting done inside DeleteSVMAuditConfig
			return nil
		}

		if !RequiresRetry(err) || attempt == retries {
			return err
		}

		tflog.Warn(ctx, "Could not delete audit configuration for SVM. Reason: Final consolidation is in progress.")
		tflog.Warn(ctx, fmt.Sprintf("retrying SVM audit config DELETE after failure (attempt %d/%d), waiting %s", attempt, retries, delay))
		time.Sleep(delay)
	}

	return fmt.Errorf("failed to delete SVM audit config after retries")
}

// Metadata returns the resource type name
func (r *ProtocolsSVMAuditResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsSVMAuditResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsSVMAudit resource",
		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "The name of the SVM.",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether or not auditing is enabled on the SVM.",
				Optional:            true,
				Computed:	         true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"events": schema.SingleNestedAttribute{
				MarkdownDescription: "Indicates events for which auditing is enabled on the SVM.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"authorization_policy": schema.BoolAttribute{
						MarkdownDescription: "Authorization policy change events.",
						Optional:            true,
						Computed:	         true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"cap_staging": schema.BoolAttribute{
						MarkdownDescription: "Central access policy staging events.",
						Optional:            true,
						Computed:	         true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"cifs_logon_logoff": schema.BoolAttribute{
						MarkdownDescription: "CIFS logon and logoff events.",
						Optional:            true,
						Computed:	         true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"file_operations": schema.BoolAttribute{
						MarkdownDescription: "File operation events.",
						Optional:            true,
						Computed:	         true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"file_share": schema.BoolAttribute{
						MarkdownDescription: "File share category events.",
						Optional:            true,
						Computed:	         true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"security_group": schema.BoolAttribute{
						MarkdownDescription: "Local security group management events.",
						Optional:            true,
						Computed:	         true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"user_account": schema.BoolAttribute{
						MarkdownDescription: "Local user account management events.",
						Optional:            true,
						Computed:	         true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"guarantee": schema.BoolAttribute{
				MarkdownDescription: `
				Indicates whether there is a strict Guarantee of Auditing.
				Requires ONTAP 9.10.1 or later.
				`,
				Optional:            true,
				Computed:	         true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"log": schema.SingleNestedAttribute{
				MarkdownDescription: "Indicates the configuration options for audit log files.",
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"format": schema.StringAttribute{
						MarkdownDescription: `
						Describes the format in which the logs are generated by consolidation process.
						Possible values are,
          				xml - Data ONTAP-specific XML log format
						evtx - Microsoft Windows EVTX log format
						`,
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("xml", "evtx"),
						},
					},
					"retention": schema.SingleNestedAttribute{
						MarkdownDescription: "Describes the count and time to retain the audit log file.",
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"count": schema.Int64Attribute{
								MarkdownDescription: `
								Determines how many audit log files to retain before rotating the oldest log file out.
								This is mutually exclusive with duration.
								`,
								Optional:            true,
								Computed:            true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.UseStateForUnknown(),
								},
							},
							"duration": schema.StringAttribute{
								MarkdownDescription: `
								Specifies an ISO-8601 format date and time to retain the audit log file.
								The audit log files are deleted once they reach the specified date/time.
								`,
								Optional:            true,
								Computed:            true,
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
								Validators: []validator.String{
									stringvalidator.ConflictsWith(path.MatchRelative().AtParent().AtName("count")),
								},
							},
						},
					},
					"rotation": schema.SingleNestedAttribute{
						MarkdownDescription: `
						Audit event log files are rotated when they reach a configured threshold log size or are on a configured schedule.
						When an event log file is rotated, the scheduled consolidation task first renames the active converted file to a time-stamped archive file,
            			and then creates a new active converted event log file.
						`,
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"size": schema.Int64Attribute{
								MarkdownDescription: `
								Rotates logs based on log size in bytes.
								Default value is 104857600.
								`,
								Optional:            true,
								Computed:            true,
								PlanModifiers: []planmodifier.Int64{
									int64planmodifier.UseStateForUnknown(),
								},
							},
							"schedule": schema.SingleNestedAttribute{
								MarkdownDescription: `
								Rotates the audit logs based on a schedule by using the time-based rotation parameters in any combination.
								The rotation schedule is calculated by using all the time-related values.
								`,
								Optional:            true,
								Computed:            true,
								PlanModifiers: []planmodifier.Object{
									objectplanmodifier.UseStateForUnknown(),
								},
								Attributes: map[string]schema.Attribute{
									"days": schema.ListAttribute{
										MarkdownDescription: `
										Specifies the day of the month schedule to rotate audit log.
										Specify -1 to rotate the audit logs all days of a month.
										`,
										ElementType:         types.Int64Type,
										Optional:			 true,
										Computed:            true,
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
									},
									"hours": schema.ListAttribute{
										MarkdownDescription: `
										Specifies the hourly schedule to rotate audit log.
										Specify -1 to rotate the audit logs every hour.
										`,
										ElementType:         types.Int64Type,
										Optional:			 true,
										Computed:            true,
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
									},
									"minutes": schema.ListAttribute{
										MarkdownDescription: "Specifies the minutes schedule to rotate the audit log.",
										ElementType:         types.Int64Type,
										Optional:			 true,
										Computed:            true,
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
									},
									"months": schema.ListAttribute{
										MarkdownDescription: `
										Specifies the months schedule to rotate audit log.
										Specify -1 to rotate the audit logs every month.
										`,
										ElementType:         types.Int64Type,
										Optional:			 true,
										Computed:            true,
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
									},
									"weekdays": schema.ListAttribute{
										MarkdownDescription: `
										Specifies the weekdays schedule to rotate audit log.
										Specify -1 to rotate the audit logs every day.
										`,
										ElementType:         types.Int64Type,
										Optional:			 true,
										Computed:            true,
										PlanModifiers: []planmodifier.List{
											listplanmodifier.UseStateForUnknown(),
										},
									},
								},
							},
						},
					},
				},
			},
			"log_path": schema.StringAttribute{
				MarkdownDescription: "The audit log destination path where consolidated audit logs are stored.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The name of the SVM.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsSVMAuditResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.ProviderConfig = config
}

// ModifyPlan handles log.retention plan/state drift on update.
// ONTAP normalizes the non-configured retention field:
// - when duration is configured, count is returned as 0
// - when count is configured, duration is returned as PT0S
func (r *ProtocolsSVMAuditResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Destroy or create plans do not need this adjustment.
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var plan *ProtocolsSVMAuditResourceModel
	var config *ProtocolsSVMAuditResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() || plan == nil || config == nil {
		return
	}

	if config.Log.IsNull() || config.Log.IsUnknown() || plan.Log.IsNull() || plan.Log.IsUnknown() {
		return
	}

	var configLog LogResourceModel
	diags := config.Log.As(ctx, &configLog, basetypes.ObjectAsOptions{
		UnhandledNullAsEmpty:    true,
		UnhandledUnknownAsEmpty: true,
	})
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	hasDurationConfig := !configLog.Retention.Duration.IsNull() && !configLog.Retention.Duration.IsUnknown() &&
		strings.TrimSpace(configLog.Retention.Duration.ValueString()) != ""
	hasCountConfig := !configLog.Retention.Count.IsNull() && !configLog.Retention.Count.IsUnknown()

	if !hasDurationConfig && !hasCountConfig {
		return
	}

	logValues := plan.Log.Attributes()
	retentionValue, ok := logValues["retention"]
	if !ok {
		return
	}

	retentionObject, ok := retentionValue.(basetypes.ObjectValue)
	if !ok || retentionObject.IsNull() || retentionObject.IsUnknown() {
		return
	}

	retentionValues := retentionObject.Attributes()
	if hasDurationConfig {
		retentionValues["count"] = types.Int64Value(0)
	}
	if hasCountConfig {
		retentionValues["duration"] = types.StringValue("PT0S")
	}

	logScheduleAttrTypes := map[string]attr.Type{
		"days":     types.ListType{ElemType: types.Int64Type},
		"hours":    types.ListType{ElemType: types.Int64Type},
		"minutes":  types.ListType{ElemType: types.Int64Type},
		"months":   types.ListType{ElemType: types.Int64Type},
		"weekdays": types.ListType{ElemType: types.Int64Type},
	}
	logRotationAttrTypes := map[string]attr.Type{
		"size":     types.Int64Type,
		"schedule": types.ObjectType{AttrTypes: logScheduleAttrTypes},
	}
	logRetentionAttrTypes := map[string]attr.Type{
		"count":    types.Int64Type,
		"duration": types.StringType,
	}
	logAttrTypes := map[string]attr.Type{
		"format":    types.StringType,
		"retention": types.ObjectType{AttrTypes: logRetentionAttrTypes},
		"rotation":  types.ObjectType{AttrTypes: logRotationAttrTypes},
	}

	newRetention, d := types.ObjectValue(logRetentionAttrTypes, retentionValues)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	logValues["retention"] = newRetention
	newLog, d := types.ObjectValue(logAttrTypes, logValues)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.Log = newLog
	resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ProtocolsSVMAuditResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsSVMAuditResourceModel

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

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	var restInfo *interfaces.ProtocolsSVMAuditGetDataModelONTAP
	restInfo, err = interfaces.GetSVMAuditConfig(errorHandler, *client, data.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetSVMAuditConfig
		return
	}

	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Enabled = types.BoolValue(restInfo.Enabled)

	// events
	eventsAttrTypes := map[string]attr.Type{
		"authorization_policy": types.BoolType,
		"cap_staging": types.BoolType,
		"cifs_logon_logoff": types.BoolType,
		"file_operations": types.BoolType,
		"file_share": types.BoolType,
		"security_group": types.BoolType,
		"user_account": types.BoolType,
	}
	eventsValues := map[string]attr.Value{
		"authorization_policy": types.BoolNull(),
		"cap_staging": types.BoolNull(),
		"cifs_logon_logoff": types.BoolNull(),
		"file_operations": types.BoolNull(),
		"file_share": types.BoolNull(),
		"security_group": types.BoolNull(),
		"user_account": types.BoolNull(),
	}
	if restInfo.Events != nil {
		eventsValues["authorization_policy"] = types.BoolValue(restInfo.Events.AuthorizationPolicy)
		eventsValues["cap_staging"] = types.BoolValue(restInfo.Events.CapStaging)
		eventsValues["cifs_logon_logoff"] = types.BoolValue(restInfo.Events.CIFSLogonLogoff)
		eventsValues["file_operations"] = types.BoolValue(restInfo.Events.FileOperations)
		eventsValues["file_share"] = types.BoolValue(restInfo.Events.FileShare)
		eventsValues["security_group"] = types.BoolValue(restInfo.Events.SecurityGroup)
		eventsValues["user_account"] = types.BoolValue(restInfo.Events.UserAccount)
	}
	objectValue, diags := types.ObjectValue(eventsAttrTypes, eventsValues)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Events = objectValue

	data.Guarantee = types.BoolValue(restInfo.Guarantee)

	// log
	logScheduleAttrTypes := map[string]attr.Type{
		"days":     types.ListType{ElemType: types.Int64Type},
		"hours":    types.ListType{ElemType: types.Int64Type},
		"minutes":  types.ListType{ElemType: types.Int64Type},
		"months":   types.ListType{ElemType: types.Int64Type},
		"weekdays": types.ListType{ElemType: types.Int64Type},
	}
	logRotationAttrTypes := map[string]attr.Type{
		"size":     types.Int64Type,
		"schedule": types.ObjectType{AttrTypes: logScheduleAttrTypes},
	}
	logRetentionAttrTypes := map[string]attr.Type{
		"count":    types.Int64Type,
		"duration": types.StringType,
	}
	logAttrTypes := map[string]attr.Type{
		"format":    types.StringType,
		"retention": types.ObjectType{AttrTypes: logRetentionAttrTypes},
		"rotation":  types.ObjectType{AttrTypes: logRotationAttrTypes},
	}
	logValues := map[string]attr.Value{
		"format":    types.StringNull(),
		"retention": types.ObjectNull(logRetentionAttrTypes),
		"rotation":  types.ObjectNull(logRotationAttrTypes),
	}
	if restInfo.Log != nil {
		logValues["format"] = types.StringValue(restInfo.Log.Format)
		if restInfo.Log.Retention != nil {
			retentionValue, d := types.ObjectValue(logRetentionAttrTypes, map[string]attr.Value{
				"count":    types.Int64Value(restInfo.Log.Retention.Count),
				"duration": types.StringValue(restInfo.Log.Retention.Duration),
			})
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			logValues["retention"] = retentionValue
		}
		if restInfo.Log.Rotation != nil {
			scheduleValue := types.ObjectNull(logScheduleAttrTypes)
			if restInfo.Log.Rotation.Schedule != nil {
				daysValue, d := int64ListValueFromOptionalSlice(restInfo.Log.Rotation.Schedule.Days)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				hoursValue, d := int64ListValueFromOptionalSlice(restInfo.Log.Rotation.Schedule.Hours)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				minutesValue, d := int64ListValueFromOptionalSlice(restInfo.Log.Rotation.Schedule.Minutes)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				monthsValue, d := int64ListValueFromOptionalSlice(restInfo.Log.Rotation.Schedule.Months)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				weekdaysValue, d := int64ListValueFromOptionalSlice(restInfo.Log.Rotation.Schedule.Weekdays)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				scheduleValueTemp, d := types.ObjectValue(logScheduleAttrTypes, map[string]attr.Value{
					"days":     daysValue,
					"hours":    hoursValue,
					"minutes":  minutesValue,
					"months":   monthsValue,
					"weekdays": weekdaysValue,
				})
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				scheduleValue = scheduleValueTemp
			}
			rotationValue, d := types.ObjectValue(logRotationAttrTypes, map[string]attr.Value{
				"size":     types.Int64Value(restInfo.Log.Rotation.Size),
				"schedule": scheduleValue,
			})
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			logValues["rotation"] = rotationValue
		}
	}
	logObjectValue, d := types.ObjectValue(logAttrTypes, logValues)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Log = logObjectValue
	
	data.LogPath = types.StringValue(restInfo.LogPath)
	data.ID = types.StringValue(restInfo.SVM.Name)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve ID
func (r *ProtocolsSVMAuditResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProtocolsSVMAuditResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ProtocolsSVMAuditResourceBodyDataModel
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
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

	var errors []string

	body.SVM.Name = data.SVMName.ValueString()
	logPathValue := data.LogPath.ValueString()
	body.LogPath = &logPathValue
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() {
		enabledValue := data.Enabled.ValueBool()
		body.Enabled = &enabledValue
	}

	// events
	if !data.Events.IsNull() && !data.Events.IsUnknown() {
		var events EventsResourceModel
		diags := data.Events.As(ctx, &events, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.Events = &interfaces.EventsDataModel{
			AuthorizationPolicy: events.AuthorizationPolicy.ValueBool(),
			CapStaging:          events.CapStaging.ValueBool(),
			CIFSLogonLogoff:     events.CIFSLogonLogoff.ValueBool(),
			FileOperations:      events.FileOperations.ValueBool(),
			FileShare:           events.FileShare.ValueBool(),
			SecurityGroup:       events.SecurityGroup.ValueBool(),
			UserAccount:         events.UserAccount.ValueBool(),
		}
	}

	if !data.Guarantee.IsNull() && !data.Guarantee.IsUnknown() {
		if (cluster.Version.Generation == 9 && cluster.Version.Major >= 10) || cluster.Version.Generation > 9 {
			guaranteeValue := data.Guarantee.ValueBool()
			body.Guarantee = &guaranteeValue
		} else {
			errors = append(errors, "guarantee requires ONTAP 9.10 or later")
		}
	}

	// log
	if !data.Log.IsNull() && !data.Log.IsUnknown() {
		var log LogResourceModel
		diags := data.Log.As(ctx, &log, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		var scheduleBody *interfaces.ScheduleDataModel
		if log.Rotation.Schedule != nil {
			scheduleBody = &interfaces.ScheduleDataModel{
				Days:     int64SliceFromTypesInt64Slice(log.Rotation.Schedule.Days),
				Hours:    int64SliceFromTypesInt64Slice(log.Rotation.Schedule.Hours),
				Minutes:  int64SliceFromTypesInt64Slice(log.Rotation.Schedule.Minutes),
				Months:   int64SliceFromTypesInt64Slice(log.Rotation.Schedule.Months),
				Weekdays: int64SliceFromTypesInt64Slice(log.Rotation.Schedule.Weekdays),
			}
		}

		body.Log = &interfaces.LogDataModel{
			Format: log.Format.ValueString(),
			Retention: &interfaces.RetentionDataModel{
				Count:    int64(log.Retention.Count.ValueInt64()),
				Duration: log.Retention.Duration.ValueString(),
			},
			Rotation: &interfaces.RotationDataModel{
				Size:     int64(log.Rotation.Size.ValueInt64()),
				Schedule: scheduleBody,
			},
		}
	}

	if len(errors) > 0 {
		errorsString := strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("The following options have ONTAP version constraints: %#v", errorsString))
		return
	}

	resource, err := interfaces.CreateSVMAuditConfig(errorHandler, *client, body)
	if err != nil {
		// error reporting done inside CreateSVMAuditConfig
		return
	}
	
	// Update the computed parameters
	data.ID = types.StringValue(resource.SVM.Name)
	data.SVMName = types.StringValue(resource.SVM.Name)
	data.Enabled = types.BoolValue(resource.Enabled)

	// events
	eventsAttrTypes := map[string]attr.Type{
		"authorization_policy": types.BoolType,
		"cap_staging": types.BoolType,
		"cifs_logon_logoff": types.BoolType,
		"file_operations": types.BoolType,
		"file_share": types.BoolType,
		"security_group": types.BoolType,
		"user_account": types.BoolType,
	}
	eventsValues := map[string]attr.Value{
		"authorization_policy": types.BoolNull(),
		"cap_staging": types.BoolNull(),
		"cifs_logon_logoff": types.BoolNull(),
		"file_operations": types.BoolNull(),
		"file_share": types.BoolNull(),
		"security_group": types.BoolNull(),
		"user_account": types.BoolNull(),
	}
	if resource.Events != nil {
		eventsValues["authorization_policy"] = types.BoolValue(resource.Events.AuthorizationPolicy)
		eventsValues["cap_staging"] = types.BoolValue(resource.Events.CapStaging)
		eventsValues["cifs_logon_logoff"] = types.BoolValue(resource.Events.CIFSLogonLogoff)
		eventsValues["file_operations"] = types.BoolValue(resource.Events.FileOperations)
		eventsValues["file_share"] = types.BoolValue(resource.Events.FileShare)
		eventsValues["security_group"] = types.BoolValue(resource.Events.SecurityGroup)
		eventsValues["user_account"] = types.BoolValue(resource.Events.UserAccount)
	}
	objectValue, diags := types.ObjectValue(eventsAttrTypes, eventsValues)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Events = objectValue

	data.Guarantee = types.BoolValue(resource.Guarantee)
	
	// log
	logScheduleAttrTypes := map[string]attr.Type{
		"days":     types.ListType{ElemType: types.Int64Type},
		"hours":    types.ListType{ElemType: types.Int64Type},
		"minutes":  types.ListType{ElemType: types.Int64Type},
		"months":   types.ListType{ElemType: types.Int64Type},
		"weekdays": types.ListType{ElemType: types.Int64Type},
	}
	logRotationAttrTypes := map[string]attr.Type{
		"size":     types.Int64Type,
		"schedule": types.ObjectType{AttrTypes: logScheduleAttrTypes},
	}
	logRetentionAttrTypes := map[string]attr.Type{
		"count":    types.Int64Type,
		"duration": types.StringType,
	}
	logAttrTypes := map[string]attr.Type{
		"format":    types.StringType,
		"retention": types.ObjectType{AttrTypes: logRetentionAttrTypes},
		"rotation":  types.ObjectType{AttrTypes: logRotationAttrTypes},
	}
	logValues := map[string]attr.Value{
		"format":    types.StringNull(),
		"retention": types.ObjectNull(logRetentionAttrTypes),
		"rotation":  types.ObjectNull(logRotationAttrTypes),
	}
	if resource.Log != nil {
		logValues["format"] = types.StringValue(resource.Log.Format)
		if resource.Log.Retention != nil {
			retentionValue, d := types.ObjectValue(logRetentionAttrTypes, map[string]attr.Value{
				"count":    types.Int64Value(resource.Log.Retention.Count),
				"duration": types.StringValue(resource.Log.Retention.Duration),
			})
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			logValues["retention"] = retentionValue
		}
		if resource.Log.Rotation != nil {
			scheduleValue := types.ObjectNull(logScheduleAttrTypes)
			if resource.Log.Rotation.Schedule != nil {
				daysValue, d := int64ListValueFromOptionalSlice(resource.Log.Rotation.Schedule.Days)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				hoursValue, d := int64ListValueFromOptionalSlice(resource.Log.Rotation.Schedule.Hours)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				minutesValue, d := int64ListValueFromOptionalSlice(resource.Log.Rotation.Schedule.Minutes)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				monthsValue, d := int64ListValueFromOptionalSlice(resource.Log.Rotation.Schedule.Months)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				weekdaysValue, d := int64ListValueFromOptionalSlice(resource.Log.Rotation.Schedule.Weekdays)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				scheduleValueTemp, d := types.ObjectValue(logScheduleAttrTypes, map[string]attr.Value{
					"days":     daysValue,
					"hours":    hoursValue,
					"minutes":  minutesValue,
					"months":   monthsValue,
					"weekdays": weekdaysValue,
				})
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				scheduleValue = scheduleValueTemp
			}
			rotationValue, d := types.ObjectValue(logRotationAttrTypes, map[string]attr.Value{
				"size":     types.Int64Value(resource.Log.Rotation.Size),
				"schedule": scheduleValue,
			})
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			logValues["rotation"] = rotationValue
		}
	}
	logObjectValue, d := types.ObjectValue(logAttrTypes, logValues)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Log = logObjectValue

	data.LogPath = types.StringValue(resource.LogPath)

	tflog.Trace(ctx, fmt.Sprintf("created SVM audit resource for SVM=%s", data.SVMName.ValueString()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsSVMAuditResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config ProtocolsSVMAuditResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Read Terraform state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// Read Terraform config data in to the model
	// This is needed to determine which of retention.count/retention.duration the user actually configured,
	// since ONTAP normalizes the non-configured field to a sentinel value (count=0 or duration=PT0S)
	// that can otherwise be indistinguishable from "unchanged" when comparing plan against prior state.
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var configLog LogResourceModel
	if !config.Log.IsNull() && !config.Log.IsUnknown() {
		diags := config.Log.As(ctx, &configLog, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	hasDurationConfig := !configLog.Retention.Duration.IsNull() && !configLog.Retention.Duration.IsUnknown() &&
		strings.TrimSpace(configLog.Retention.Duration.ValueString()) != ""
	hasCountConfig := !configLog.Retention.Count.IsNull() && !configLog.Retention.Count.IsUnknown()

	client, err := connection.GetRestClient(errorHandler, r.config, state.CxProfileName)
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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, state.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	// Update the resource
	var body interfaces.ProtocolsSVMAuditResourceBodyDataModel

	var errors []string

	// No other fields can be specified when enabled is specified for modify
	if !plan.Enabled.Equal(state.Enabled) && !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		err = interfaces.EnableDisableSVMAudit(errorHandler, *client, svm.UUID, plan.Enabled.ValueBool())
		if err != nil {
			// error reporting done inside EnableDisableSVMAudit
			return
		}
	}
	// Update the rest of the fields
	// events
	if !plan.Events.IsNull() && !plan.Events.IsUnknown() {
		var planEvents EventsResourceModel
		diags := plan.Events.As(ctx, &planEvents, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		var stateEvents EventsResourceModel
		hasStateEvents := false
		if !state.Events.IsNull() && !state.Events.IsUnknown() {
			diags = state.Events.As(ctx, &stateEvents, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			hasStateEvents = true
		}
		
		events := &interfaces.EventsDataModel{}
		hasChange := false

		// only send fields that are explicitly known and differ from state
		if !planEvents.AuthorizationPolicy.IsNull() && !planEvents.AuthorizationPolicy.IsUnknown() &&
			(!hasStateEvents || !planEvents.AuthorizationPolicy.Equal(stateEvents.AuthorizationPolicy)) {
			events.AuthorizationPolicy = planEvents.AuthorizationPolicy.ValueBool()
			hasChange = true
		}
		if !planEvents.CapStaging.IsNull() && !planEvents.CapStaging.IsUnknown() &&
			(!hasStateEvents || !planEvents.CapStaging.Equal(stateEvents.CapStaging)) {
			events.CapStaging = planEvents.CapStaging.ValueBool()
			hasChange = true
		}
		if !planEvents.CIFSLogonLogoff.IsNull() && !planEvents.CIFSLogonLogoff.IsUnknown() &&
			(!hasStateEvents || !planEvents.CIFSLogonLogoff.Equal(stateEvents.CIFSLogonLogoff)) {
			events.CIFSLogonLogoff = planEvents.CIFSLogonLogoff.ValueBool()
			hasChange = true
		}
		if !planEvents.FileOperations.IsNull() && !planEvents.FileOperations.IsUnknown() &&
			(!hasStateEvents || !planEvents.FileOperations.Equal(stateEvents.FileOperations)) {
			events.FileOperations = planEvents.FileOperations.ValueBool()
			hasChange = true
		}
		if !planEvents.FileShare.IsNull() && !planEvents.FileShare.IsUnknown() &&
			(!hasStateEvents || !planEvents.FileShare.Equal(stateEvents.FileShare)) {
			events.FileShare = planEvents.FileShare.ValueBool()
			hasChange = true
		}
		if !planEvents.SecurityGroup.IsNull() && !planEvents.SecurityGroup.IsUnknown() &&
			(!hasStateEvents || !planEvents.SecurityGroup.Equal(stateEvents.SecurityGroup)) {
			events.SecurityGroup = planEvents.SecurityGroup.ValueBool()
			hasChange = true
		}
		if !planEvents.UserAccount.IsNull() && !planEvents.UserAccount.IsUnknown() &&
			(!hasStateEvents || !planEvents.UserAccount.Equal(stateEvents.UserAccount)) {
			events.UserAccount = planEvents.UserAccount.ValueBool()
			hasChange = true
		}

		if hasChange {
			body.Events = events
		}
	}

	if !plan.Guarantee.Equal(state.Guarantee) &&
		!plan.Guarantee.IsNull() && !plan.Guarantee.IsUnknown() {
		if (cluster.Version.Generation == 9 && cluster.Version.Major >= 10) || cluster.Version.Generation > 9 {
			guaranteeValue := plan.Guarantee.ValueBool()
			body.Guarantee = &guaranteeValue
		} else {
			errors = append(errors, "guarantee requires ONTAP 9.16 or later")
		}
	}

	// log
	if !plan.Log.IsNull() && !plan.Log.IsUnknown() {
		var planLog LogResourceModel
		diags := plan.Log.As(ctx, &planLog, basetypes.ObjectAsOptions{
			UnhandledNullAsEmpty:    true,
			UnhandledUnknownAsEmpty: true,
		})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		var stateLog LogResourceModel
		hasStateLog := false
		if !state.Log.IsNull() && !state.Log.IsUnknown() {
			diags = state.Log.As(ctx, &stateLog, basetypes.ObjectAsOptions{
				UnhandledNullAsEmpty:    true,
				UnhandledUnknownAsEmpty: true,
			})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			hasStateLog = true
		}

		log := &interfaces.LogDataModel{}
		hasLogChange := false

		if !planLog.Format.IsNull() && !planLog.Format.IsUnknown() &&
			(!hasStateLog || !planLog.Format.Equal(stateLog.Format)) {
			log.Format = planLog.Format.ValueString()
			hasLogChange = true
		}

		retention := &interfaces.RetentionDataModel{}
		hasRetentionChange := false
		countKnown := !planLog.Retention.Count.IsNull() && !planLog.Retention.Count.IsUnknown()
		durationKnown := !planLog.Retention.Duration.IsNull() && !planLog.Retention.Duration.IsUnknown()

		countChanged := countKnown && (!hasStateLog || !planLog.Retention.Count.Equal(stateLog.Retention.Count))
		durationChanged := durationKnown && (!hasStateLog || !planLog.Retention.Duration.Equal(stateLog.Retention.Duration))

		countValue := int64(0)
		if countKnown {
			countValue = planLog.Retention.Count.ValueInt64()
		}
		durationValue := ""
		if durationKnown {
			durationValue = strings.TrimSpace(planLog.Retention.Duration.ValueString())
		}

		// Decide which field to send based on what the user actually configured
		// (hasDurationConfig/hasCountConfig), not on a plan-vs-state text diff.
		// ONTAP normalizes the non-configured retention field to a sentinel value (count=0 or duration=PT0S)
		// as a side effect of the other field being present in the PATCH body.
		// That side effect can make the "active" field look unchanged
		// (e.g. switching away from count while duration was already PT0S in state),
		// so we must still send it whenever the retention mode is switching, even if its text matches the prior state.
		switch {
		// case hasDurationConfig && hasCountConfig:
		// Schema validator (ConflictsWith) should already prevent this.
		case hasDurationConfig:
			stateCountActive := hasStateLog && stateLog.Retention.Count.ValueInt64() != 0
			if durationChanged || stateCountActive {
				retention.Duration = durationValue
				hasRetentionChange = true
			}
		case hasCountConfig:
			stateDuration := ""
			if hasStateLog {
				stateDuration = strings.TrimSpace(stateLog.Retention.Duration.ValueString())
			}
			stateDurationActive := stateDuration != "" && stateDuration != "PT0S"
			if countChanged || stateDurationActive {
				retention.Count = countValue
				hasRetentionChange = true
			}
		}
		if hasRetentionChange {
			log.Retention = retention
			hasLogChange = true
		}

		rotation := &interfaces.RotationDataModel{}
		hasRotationChange := false
		if !planLog.Rotation.Size.IsNull() && !planLog.Rotation.Size.IsUnknown() &&
			(!hasStateLog || !planLog.Rotation.Size.Equal(stateLog.Rotation.Size)) {
			rotation.Size = planLog.Rotation.Size.ValueInt64()
			hasRotationChange = true
		}

		planSchedule := planLog.Rotation.Schedule
		var stateSchedule *ScheduleResourceModel
		if hasStateLog {
			stateSchedule = stateLog.Rotation.Schedule
		}

		planDays := []types.Int64{}
		if planSchedule != nil {
			planDays = planSchedule.Days
		}
		stateDays := []types.Int64{}
		if stateSchedule != nil {
			stateDays = stateSchedule.Days
		}

		planHours := []types.Int64{}
		if planSchedule != nil {
			planHours = planSchedule.Hours
		}
		stateHours := []types.Int64{}
		if stateSchedule != nil {
			stateHours = stateSchedule.Hours
		}

		planMinutes := []types.Int64{}
		if planSchedule != nil {
			planMinutes = planSchedule.Minutes
		}
		stateMinutes := []types.Int64{}
		if stateSchedule != nil {
			stateMinutes = stateSchedule.Minutes
		}

		planMonths := []types.Int64{}
		if planSchedule != nil {
			planMonths = planSchedule.Months
		}
		stateMonths := []types.Int64{}
		if stateSchedule != nil {
			stateMonths = stateSchedule.Months
		}

		planWeekdays := []types.Int64{}
		if planSchedule != nil {
			planWeekdays = planSchedule.Weekdays
		}
		stateWeekdays := []types.Int64{}
		if stateSchedule != nil {
			stateWeekdays = stateSchedule.Weekdays
		}

		scheduleBody := &interfaces.ScheduleDataModel{}
		hasScheduleChange := false
		if hasStateLog && len(stateDays) > 0 && len(planDays) == 0 {
			scheduleBody.Days = []int64{}
			hasScheduleChange = true
		} else if !hasStateLog || !typesInt64SlicesEqual(planDays, stateDays) {
			scheduleBody.Days = int64SliceFromTypesInt64Slice(planDays)
			hasScheduleChange = true
		}
		if hasStateLog && len(stateHours) > 0 && len(planHours) == 0 {
			scheduleBody.Hours = []int64{}
			hasScheduleChange = true
		} else if !hasStateLog || !typesInt64SlicesEqual(planHours, stateHours) {
			scheduleBody.Hours = int64SliceFromTypesInt64Slice(planHours)
			hasScheduleChange = true
		}
		if hasStateLog && len(stateMinutes) > 0 && len(planMinutes) == 0 {
			scheduleBody.Minutes = []int64{}
			hasScheduleChange = true
		} else if !hasStateLog || !typesInt64SlicesEqual(planMinutes, stateMinutes) {
			scheduleBody.Minutes = int64SliceFromTypesInt64Slice(planMinutes)
			hasScheduleChange = true
		}
		if hasStateLog && len(stateMonths) > 0 && len(planMonths) == 0 {
			scheduleBody.Months = []int64{}
			hasScheduleChange = true
		} else if !hasStateLog || !typesInt64SlicesEqual(planMonths, stateMonths) {
			scheduleBody.Months = int64SliceFromTypesInt64Slice(planMonths)
			hasScheduleChange = true
		}
		if hasStateLog && len(stateWeekdays) > 0 && len(planWeekdays) == 0 {
			scheduleBody.Weekdays = []int64{}
			hasScheduleChange = true
		} else if !hasStateLog || !typesInt64SlicesEqual(planWeekdays, stateWeekdays) {
			scheduleBody.Weekdays = int64SliceFromTypesInt64Slice(planWeekdays)
			hasScheduleChange = true
		}
		if hasScheduleChange {
			rotation.Schedule = scheduleBody
			hasRotationChange = true
		}
		if hasRotationChange {
			log.Rotation = rotation
			hasLogChange = true
		}

		if hasLogChange {
			body.Log = log
		}
	}

	if !plan.LogPath.Equal(state.LogPath) &&
		!plan.LogPath.IsNull() && !plan.LogPath.IsUnknown() {
		logPathValue := plan.LogPath.ValueString()
		body.LogPath = &logPathValue
	}

	if len(errors) > 0 {
		errorsString := strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("The following options have ONTAP version constraints: %#v", errorsString))
		return
	}

	err = interfaces.UpdateSVMAuditConfig(errorHandler, *client, body, svm.UUID)
	if err != nil {
		// error reporting done inside UpdateSVMAuditConfig
		return
	}

	restInfo, err := interfaces.GetSVMAuditConfig(errorHandler, *client, plan.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetSVMAuditConfig
		return
	}

	// Update the computed parameters
	plan.ID = types.StringValue(restInfo.SVM.Name)
	plan.SVMName = types.StringValue(restInfo.SVM.Name)
	plan.Enabled = types.BoolValue(restInfo.Enabled)

	// events
	eventsAttrTypes := map[string]attr.Type{
		"authorization_policy": types.BoolType,
		"cap_staging": types.BoolType,
		"cifs_logon_logoff": types.BoolType,
		"file_operations": types.BoolType,
		"file_share": types.BoolType,
		"security_group": types.BoolType,
		"user_account": types.BoolType,
	}
	eventsValues := map[string]attr.Value{
		"authorization_policy": types.BoolNull(),
		"cap_staging": types.BoolNull(),
		"cifs_logon_logoff": types.BoolNull(),
		"file_operations": types.BoolNull(),
		"file_share": types.BoolNull(),
		"security_group": types.BoolNull(),
		"user_account": types.BoolNull(),
	}
	eventsValues["authorization_policy"] = types.BoolValue(restInfo.Events.AuthorizationPolicy)
	eventsValues["cap_staging"] = types.BoolValue(restInfo.Events.CapStaging)
	eventsValues["cifs_logon_logoff"] = types.BoolValue(restInfo.Events.CIFSLogonLogoff)
	eventsValues["file_operations"] = types.BoolValue(restInfo.Events.FileOperations)
	eventsValues["file_share"] = types.BoolValue(restInfo.Events.FileShare)
	eventsValues["security_group"] = types.BoolValue(restInfo.Events.SecurityGroup)
	eventsValues["user_account"] = types.BoolValue(restInfo.Events.UserAccount)
	objectValue, diags := types.ObjectValue(eventsAttrTypes, eventsValues)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	plan.Events = objectValue
	plan.Guarantee = types.BoolValue(restInfo.Guarantee)
	
	// log
	logScheduleAttrTypes := map[string]attr.Type{
		"days":     types.ListType{ElemType: types.Int64Type},
		"hours":    types.ListType{ElemType: types.Int64Type},
		"minutes":  types.ListType{ElemType: types.Int64Type},
		"months":   types.ListType{ElemType: types.Int64Type},
		"weekdays": types.ListType{ElemType: types.Int64Type},
	}
	logRotationAttrTypes := map[string]attr.Type{
		"size":     types.Int64Type,
		"schedule": types.ObjectType{AttrTypes: logScheduleAttrTypes},
	}
	logRetentionAttrTypes := map[string]attr.Type{
		"count":    types.Int64Type,
		"duration": types.StringType,
	}
	logAttrTypes := map[string]attr.Type{
		"format":    types.StringType,
		"retention": types.ObjectType{AttrTypes: logRetentionAttrTypes},
		"rotation":  types.ObjectType{AttrTypes: logRotationAttrTypes},
	}
	logValues := map[string]attr.Value{
		"format":    types.StringNull(),
		"retention": types.ObjectNull(logRetentionAttrTypes),
		"rotation":  types.ObjectNull(logRotationAttrTypes),
	}
	if restInfo.Log != nil {
		logValues["format"] = types.StringValue(restInfo.Log.Format)
		if restInfo.Log.Retention != nil {
			retentionValue, d := types.ObjectValue(logRetentionAttrTypes, map[string]attr.Value{
				"count":    types.Int64Value(restInfo.Log.Retention.Count),
				"duration": types.StringValue(restInfo.Log.Retention.Duration),
			})
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			logValues["retention"] = retentionValue
		}
		if restInfo.Log.Rotation != nil {
			scheduleValue := types.ObjectNull(logScheduleAttrTypes)
			if restInfo.Log.Rotation.Schedule != nil {
				daysValue, d := int64ListValueFromOptionalSlice(restInfo.Log.Rotation.Schedule.Days)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				hoursValue, d := int64ListValueFromOptionalSlice(restInfo.Log.Rotation.Schedule.Hours)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				minutesValue, d := int64ListValueFromOptionalSlice(restInfo.Log.Rotation.Schedule.Minutes)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				monthsValue, d := int64ListValueFromOptionalSlice(restInfo.Log.Rotation.Schedule.Months)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				weekdaysValue, d := int64ListValueFromOptionalSlice(restInfo.Log.Rotation.Schedule.Weekdays)
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				scheduleValueTemp, d := types.ObjectValue(logScheduleAttrTypes, map[string]attr.Value{
					"days":     daysValue,
					"hours":    hoursValue,
					"minutes":  minutesValue,
					"months":   monthsValue,
					"weekdays": weekdaysValue,
				})
				resp.Diagnostics.Append(d...)
				if resp.Diagnostics.HasError() {
					return
				}
				scheduleValue = scheduleValueTemp
			}
			rotationValue, d := types.ObjectValue(logRotationAttrTypes, map[string]attr.Value{
				"size":     types.Int64Value(restInfo.Log.Rotation.Size),
				"schedule": scheduleValue,
			})
			resp.Diagnostics.Append(d...)
			if resp.Diagnostics.HasError() {
				return
			}
			logValues["rotation"] = rotationValue
		}
	}
	logObjectValue, d := types.ObjectValue(logAttrTypes, logValues)
	resp.Diagnostics.Append(d...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Log = logObjectValue

	plan.LogPath = types.StringValue(restInfo.LogPath)
	
	tflog.Debug(ctx, fmt.Sprintf("updated SVM audit resource for SVM=%s", plan.SVMName.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsSVMAuditResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsSVMAuditResourceModel

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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	// Auditing must be disabled before deleting the audit configuration
	if !data.Enabled.IsNull() && !data.Enabled.IsUnknown() && data.Enabled.ValueBool() {
		tflog.Warn(ctx, "SVM auditing is currently enabled; disabling it before destroying the audit configuration.")
		err = interfaces.EnableDisableSVMAudit(errorHandler, *client, svm.UUID, false)
		if err != nil {
			// error reporting done inside EnableDisableSVMAudit
			return
		}
	}

	err = DeleteSVMAuditConfigWithRetry(ctx, errorHandler, *client, svm.UUID)
	if err != nil {
		// error reporting done inside DeleteSVMAuditConfig
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsSVMAuditResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[1])...)
}
