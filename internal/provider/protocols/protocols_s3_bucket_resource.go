package protocols

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsS3BucketResource{}
var _ resource.ResourceWithImportState = &ProtocolsS3BucketResource{}
var _ resource.ResourceWithModifyPlan = &ProtocolsS3BucketResource{}

// NewProtocolsS3BucketResource is a helper function to simplify the provider implementation.
func NewProtocolsS3BucketResource() resource.Resource {
	return &ProtocolsS3BucketResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "s3_bucket",
		},
	}
}

// ProtocolsS3BucketResource defines the resource implementation.
type ProtocolsS3BucketResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsS3BucketResourceModel describes the resource data model.
type ProtocolsS3BucketResourceModel struct {
	CxProfileName             types.String  `tfsdk:"cx_profile_name"`
	SVMName                   types.String  `tfsdk:"svm_name"`
	Name                      types.String  `tfsdk:"name"`
	Size                      types.Int64   `tfsdk:"size"`
	Comment                   types.String  `tfsdk:"comment"`
	Type                      types.String  `tfsdk:"type"`
	NASPath                   types.String  `tfsdk:"nas_path"`
	VersioningState           types.String  `tfsdk:"versioning_state"`
	Policy                    types.Object  `tfsdk:"policy"`
	QoSPolicy                 types.Object  `tfsdk:"qos_policy"`
	SnapshotPolicy            types.String  `tfsdk:"snapshot_policy"`
	AuditEventSelector        types.Object  `tfsdk:"audit_event_selector"`
	Aggregates                types.List    `tfsdk:"aggregates"`
	ConstituentsPerAggregate  types.Int64   `tfsdk:"constituents_per_aggregate"`
	ID                        types.String  `tfsdk:"id"`
}

// PolicyResourceModel describes the policy data model using go types for mapping.
type PolicyResourceModel struct {
	Statements  []StatementsResourceModel  `tfsdk:"statements"`
}

func (m PolicyResourceModel) attributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"statements": types.ListType{ElemType: types.ObjectType{AttrTypes: StatementsResourceModel{}.attributeTypes()}},
	}
}

// StatementsResourceModel describes the statements data model using go types for mapping.
type StatementsResourceModel struct {
	SID           types.String             `tfsdk:"sid"`
	Resources     types.Set                `tfsdk:"resources"`
	Actions       types.Set                `tfsdk:"actions"`
	Effect        types.String             `tfsdk:"effect"`
	Principals    types.Set                `tfsdk:"principals"`
	Conditions  []ConditionsResourceModel  `tfsdk:"conditions"`
}

func (m StatementsResourceModel) attributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"sid":        types.StringType,
		"resources":  types.SetType{ElemType: types.StringType},
		"actions":    types.SetType{ElemType: types.StringType},
		"effect":     types.StringType,
		"principals": types.SetType{ElemType: types.StringType},
		"conditions": types.ListType{ElemType: types.ObjectType{AttrTypes: ConditionsResourceModel{}.attributeTypes()}},
	}
}

// ConditionsResourceModel describes the conditions data model using go types for mapping.
type ConditionsResourceModel struct {
	Operator    types.String  `tfsdk:"operator"`
	MaxKeys     types.Set     `tfsdk:"max_keys"`
	Delimiters  types.Set     `tfsdk:"delimiters"`
	SourceIPs   types.Set     `tfsdk:"source_ips"`
	Prefixes    types.Set     `tfsdk:"prefixes"`
	Usernames   types.Set     `tfsdk:"usernames"`
}

func (m ConditionsResourceModel) attributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"operator":   types.StringType,
		"max_keys":   types.SetType{ElemType: types.StringType},
		"delimiters": types.SetType{ElemType: types.StringType},
		"source_ips": types.SetType{ElemType: types.StringType},
		"prefixes":   types.SetType{ElemType: types.StringType},
		"usernames":  types.SetType{ElemType: types.StringType},
	}
}

// QoSPolicyResourceModel describes the qos_policy data model using go types for mapping.
type QoSPolicyResourceModel struct {
	Name               types.String  `tfsdk:"name"`
	MaxThroughputIops  types.Int64   `tfsdk:"max_throughput_iops"`
	MaxThroughputMbps  types.Int64   `tfsdk:"max_throughput_mbps"`
	MinThroughputIops  types.Int64   `tfsdk:"min_throughput_iops"`
	MinThroughputMbps  types.Int64   `tfsdk:"min_throughput_mbps"`
}

func (m QoSPolicyResourceModel) attributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":                types.StringType,
		"max_throughput_iops": types.Int64Type,
		"max_throughput_mbps": types.Int64Type,
		"min_throughput_iops": types.Int64Type,
		"min_throughput_mbps": types.Int64Type,
	}
}

// AuditEventSelectorResourceModel describes the audit_event_selector data model using go types for mapping.
type AuditEventSelectorResourceModel struct {
	Access      types.String  `tfsdk:"access"`
	Permission  types.String  `tfsdk:"permission"`
}

func (m AuditEventSelectorResourceModel) attributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"access":     types.StringType,
		"permission": types.StringType,
	}
}

// Metadata returns the resource type name
func (r *ProtocolsS3BucketResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsS3BucketResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsS3Bucket resource",
		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "The name of the SVM.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the S3 or NAS bucket.",
				Required:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the S3 bucket in bytes. This option is not supported when type is set to NAS.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Additional information about the bucket.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: `
				Specifies the bucket type.
				Requires ONTAP 9.12.1 or later.
				`,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("s3", "nas"),
				},
			},
			"nas_path": schema.StringAttribute{
				MarkdownDescription: `
				The NAS path to which the NAS bucket corresponds.
				Requires ONTAP 9.12.1 or later.
				`,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"versioning_state": schema.StringAttribute{
				MarkdownDescription: `
				The versioning state of the bucket. The versioning state cannot be modified to 'disabled' from any other state.
				Requires ONTAP 9.11.1 or later.
				`,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("disabled", "enabled", "suspended"),
				},
			},
			"policy": schema.SingleNestedAttribute{
				MarkdownDescription: "Access policy uses the AWS policy language syntax to allow S3 tenants to create access policies to their data.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"statements": schema.ListNestedAttribute{
						MarkdownDescription: `
						Policy statements are built using this structure to specify permissions.
						Grant <Effect> to allow/deny <Principal> to perform <Action> on <Resource> when <Condition> applies.
						`,
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.List{
							listplanmodifier.UseStateForUnknown(),
						},
						NestedObject: schema.NestedAttributeObject{
							Attributes: map[string]schema.Attribute{
								"sid": schema.StringAttribute{
									MarkdownDescription: "Statement ID",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
								"resources": schema.SetAttribute{
									MarkdownDescription: `
									The bucket and any object it contains.
									The wildcard characters * and ? can be used to form a regular expression for specifying a resource.
									`,
									Optional:            true,
									Computed:            true,
									ElementType:         types.StringType,
									PlanModifiers: []planmodifier.Set{
										setplanmodifier.UseStateForUnknown(),
									},
								},
								"actions": schema.SetAttribute{
									MarkdownDescription: "You can specify * to mean all actions, or a list of one or more actions",
									Optional:            true,
									Computed:            true,
									ElementType:         types.StringType,
									PlanModifiers: []planmodifier.Set{
										setplanmodifier.UseStateForUnknown(),
									},
									Validators: []validator.Set{
										setvalidator.ValueStringsAre(stringvalidator.OneOf(
											"GetObject",
											"PutObject",
											"DeleteObject",
											"ListBucket",
											"GetBucketAcl",
											"GetObjectAcl",
											"ListBucketMultipartUploads",
											"ListMultipartUploadParts",
										)),
									},
								},
								"effect": schema.StringAttribute{
									MarkdownDescription: "The statement may allow or deny access.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.String{
										stringplanmodifier.UseStateForUnknown(),
									},
								},
								"principals": schema.SetAttribute{
									MarkdownDescription: "A list of one or more S3 users or groups.",
									Optional:            true,
									Computed:            true,
									ElementType:         types.StringType,
									PlanModifiers: []planmodifier.Set{
										setplanmodifier.UseStateForUnknown(),
									},
								},
								"conditions": schema.ListNestedAttribute{
									MarkdownDescription: "Conditions for when a policy is in effect.",
									Optional:            true,
									Computed:            true,
									PlanModifiers: []planmodifier.List{
										listplanmodifier.UseStateForUnknown(),
									},
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"operator": schema.StringAttribute{
												MarkdownDescription: "The operator to use for the condition.",
												Required:            true,
												Validators: []validator.String{
													stringvalidator.OneOf(
														"ip_address",
														"not_ip_address",
														"string_equals",
														"string_not_equals",
														"string_equals_ignore_case",
														"string_not_equals_ignore_case",
														"string_like",
														"string_not_like",
														"numeric_equals",
														"numeric_not_equals",
														"numeric_greater_than",
														"numeric_greater_than_equals",
														"numeric_less_than",
														"numeric_less_than_equals",
													),
												},
											},
											"max_keys": schema.SetAttribute{
												MarkdownDescription: "The maximum number of keys that can be returned in a request.",
												Optional:            true,
												Computed:            true,
												ElementType:         types.StringType,
												PlanModifiers: []planmodifier.Set{
													setplanmodifier.UseStateForUnknown(),
												},
											},
											"delimiters": schema.SetAttribute{
												MarkdownDescription: "The delimiter used to identify a prefix in a list of objects.",
												Optional:            true,
												Computed:            true,
												ElementType:         types.StringType,
												PlanModifiers: []planmodifier.Set{
													setplanmodifier.UseStateForUnknown(),
												},
											},
											"source_ips": schema.SetAttribute{
												MarkdownDescription: "The source IP addresses of the request.",
												Optional:            true,
												Computed:            true,
												ElementType:         types.StringType,
												PlanModifiers: []planmodifier.Set{
													setplanmodifier.UseStateForUnknown(),
												},
											},
											"prefixes": schema.SetAttribute{
												MarkdownDescription: "The prefixes of the objects that you want to list.",
												Optional:            true,
												Computed:            true,
												ElementType:         types.StringType,
												PlanModifiers: []planmodifier.Set{
													setplanmodifier.UseStateForUnknown(),
												},
											},
											"usernames": schema.SetAttribute{
												MarkdownDescription: "The user names that you want to allow to access the bucket.",
												Optional:            true,
												Computed:            true,
												ElementType:         types.StringType,
												PlanModifiers: []planmodifier.Set{
													setplanmodifier.UseStateForUnknown(),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"qos_policy": schema.SingleNestedAttribute{
				MarkdownDescription: `
				A policy group defines measurable service level objectives (SLOs) that apply to the storage objects with which the policy group is associated.
				If you do not assign a policy group to a bucket, the system wil not monitor and control the traffic to it.
				This option is not supported when type is set to NAS.
				`,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "The QoS policy group name. This is mutually exclusive with other QoS attributes.",
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							stringvalidator.ConflictsWith(
								path.MatchRelative().AtParent().AtName("max_throughput_iops"),
								path.MatchRelative().AtParent().AtName("max_throughput_mbps"),
								path.MatchRelative().AtParent().AtName("min_throughput_iops"),
								path.MatchRelative().AtParent().AtName("min_throughput_mbps"),
							),
						},
					},
					"max_throughput_iops": schema.Int64Attribute{
						MarkdownDescription: "The maximum throughput in IOPS.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("name")),
						},
					},
					"max_throughput_mbps": schema.Int64Attribute{
						MarkdownDescription: "The maximum throughput in MBPS.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("name")),
						},
					},
					"min_throughput_iops": schema.Int64Attribute{
						MarkdownDescription: "The minimum throughput in IOPS.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("name")),
						},
					},
					"min_throughput_mbps": schema.Int64Attribute{
						MarkdownDescription: "The minimum throughput in MBPS.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Int64{
							int64validator.ConflictsWith(path.MatchRelative().AtParent().AtName("name")),
						},
					},
				},
			},
			"snapshot_policy": schema.StringAttribute{
				MarkdownDescription: `
				The snapshot policy for the bucket.
				Requires ONTAP 9.16.1 or later.
				`,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"audit_event_selector": schema.SingleNestedAttribute{
				MarkdownDescription: `
				Audit event selector allows you to specify access and permission types to audit.
				This option is not supported when type is set to NAS.
				Requires ONTAP 9.10.1 or later.
				`,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"access": schema.StringAttribute{
						MarkdownDescription: "The type of event access to be audited.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("read", "write", "all"),
						},
					},
					"permission": schema.StringAttribute{
						MarkdownDescription: "The type of event permission to be audited.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("allow", "deny", "all"),
						},
					},
				},
			},
			"aggregates": schema.ListAttribute{
				MarkdownDescription: "List of aggregates to use for the S3 bucket. This option is not supported when type is set to NAS.",
				Optional:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			"constituents_per_aggregate": schema.Int64Attribute{
				MarkdownDescription: "Number of constituents per aggregate. This option is not supported when type is set to NAS.",
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.MatchRoot("aggregates")),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the S3 bucket.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsS3BucketResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ModifyPlan preserves the auto-generated qos_policy.name from state when it is not set in config,
// preventing a plan diff of (known value) -> null or (known after apply).
func (r *ProtocolsS3BucketResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var plan, state *ProtocolsS3BucketResourceModel
	var config *ProtocolsS3BucketResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planChanged := false

	// qos_policy.name can be set by the user, but if it is omitted the API auto-generates it when QoS is sent.
	// In that case the planned value is often a *known null*, and then Apply fails with:
	// "produced an unexpected new value ... was null, but now ...".
	//
	// Fix: if name is not explicitly set in config and QoS throughput settings are being created/changed,
	// force the planned name to unknown (otherwise would be known null) so the API-generated value is accepted.
	// Otherwise, preserve the existing state name to avoid diffs.
	if plan != nil && !plan.QoSPolicy.IsNull() && !plan.QoSPolicy.IsUnknown() {
		var planQoS QoSPolicyResourceModel
		diags := plan.QoSPolicy.As(ctx, &planQoS, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateQoS *QoSPolicyResourceModel
		if state != nil && !state.QoSPolicy.IsNull() && !state.QoSPolicy.IsUnknown() {
			var decoded QoSPolicyResourceModel
			diags := state.QoSPolicy.As(ctx, &decoded, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			stateQoS = &decoded
		}

		nameExplicitInConfig := false
		if config != nil && !config.QoSPolicy.IsNull() && !config.QoSPolicy.IsUnknown() {
			var configQoS QoSPolicyResourceModel
			diags := config.QoSPolicy.As(ctx, &configQoS, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			if !configQoS.Name.IsNull() && !configQoS.Name.IsUnknown() && configQoS.Name.ValueString() != "" {
				nameExplicitInConfig = true
			}
		}

		if !nameExplicitInConfig && (planQoS.Name.IsNull() || planQoS.Name.IsUnknown()) {
			qosThroughputChangingOrCreating := false
			if stateQoS == nil {
				// Create (or no prior QoS state): if any throughput is set/known, QoS will be sent and name will be generated.
				if (!planQoS.MaxThroughputIops.IsNull() && !planQoS.MaxThroughputIops.IsUnknown()) ||
					(!planQoS.MaxThroughputMbps.IsNull() && !planQoS.MaxThroughputMbps.IsUnknown()) ||
					(!planQoS.MinThroughputIops.IsNull() && !planQoS.MinThroughputIops.IsUnknown()) ||
					(!planQoS.MinThroughputMbps.IsNull() && !planQoS.MinThroughputMbps.IsUnknown()) {
					qosThroughputChangingOrCreating = true
				}
			} else {
				// Update: mimic Update() logic - QoS is sent only when a known throughput value differs from state.
				if !planQoS.MaxThroughputIops.IsNull() && !planQoS.MaxThroughputIops.IsUnknown() &&
					!planQoS.MaxThroughputIops.Equal(stateQoS.MaxThroughputIops) {
					qosThroughputChangingOrCreating = true
				}
				if !planQoS.MaxThroughputMbps.IsNull() && !planQoS.MaxThroughputMbps.IsUnknown() &&
					!planQoS.MaxThroughputMbps.Equal(stateQoS.MaxThroughputMbps) {
					qosThroughputChangingOrCreating = true
				}
				if !planQoS.MinThroughputIops.IsNull() && !planQoS.MinThroughputIops.IsUnknown() &&
					!planQoS.MinThroughputIops.Equal(stateQoS.MinThroughputIops) {
					qosThroughputChangingOrCreating = true
				}
				if !planQoS.MinThroughputMbps.IsNull() && !planQoS.MinThroughputMbps.IsUnknown() &&
					!planQoS.MinThroughputMbps.Equal(stateQoS.MinThroughputMbps) {
					qosThroughputChangingOrCreating = true
				}
			}

			if qosThroughputChangingOrCreating {
				if planQoS.Name.IsNull() {
					planQoS.Name = types.StringUnknown()
					newQoS, diags := types.ObjectValueFrom(ctx, planQoS.attributeTypes(), planQoS)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
					plan.QoSPolicy = newQoS
					planChanged = true
				}
			} else if stateQoS != nil && !stateQoS.Name.IsNull() {
				// No QoS throughput update, keep the existing API-generated name.
				planQoS.Name = stateQoS.Name
				newQoS, diags := types.ObjectValueFrom(ctx, planQoS.attributeTypes(), planQoS)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				plan.QoSPolicy = newQoS
				planChanged = true
			}
		}
	}

	if planChanged {
		resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *ProtocolsS3BucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsS3BucketResourceModel

	// Read Terraform prior state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	trackPolicy := !data.Policy.IsNull() && !data.Policy.IsUnknown()

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

	var restInfo *interfaces.ProtocolsS3BucketGetDataModelONTAP
	restInfo, err = interfaces.GetProtocolsS3Bucket(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsS3Bucket
		return
	}

	data.ID = types.StringValue(restInfo.UUID)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.Size = types.Int64Value(restInfo.Size)
	data.Comment = types.StringValue(restInfo.Comment)
	data.Type = types.StringValue(restInfo.Type)
	data.NASPath = types.StringValue(restInfo.NASPath)
	data.VersioningState = types.StringValue(restInfo.VersioningState)

	// policy
	if restInfo.Policy == nil {
		// ONTAP may omit the policy object when it is empty, even if previously sent statements=[]
		// keep it as an empty policy (statements=[]) to avoid flipping between null and []
		if trackPolicy {
			emptyPolicy := PolicyResourceModel{Statements: []StatementsResourceModel{}}
			policyValue, diags := types.ObjectValueFrom(ctx, emptyPolicy.attributeTypes(), emptyPolicy)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.Policy = policyValue
		} else {
			data.Policy = types.ObjectNull(PolicyResourceModel{}.attributeTypes())
		}
	} else {
		var policy PolicyResourceModel
		policy.Statements = []StatementsResourceModel{}
		for _, item := range restInfo.Policy.Statements {
			var statement StatementsResourceModel
			statement.SID = types.StringValue(item.SID)
			statement.Effect = types.StringValue(item.Effect)
			// resources
			var resources = make([]string, len(item.Resources))
			copy(resources, item.Resources)
			resourcesSet, diags := types.SetValueFrom(ctx, types.StringType, resources)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			statement.Resources = resourcesSet
			//actions
			var actions = make([]string, len(item.Actions))
			copy(actions, item.Actions)
			actionsSet, diags := types.SetValueFrom(ctx, types.StringType, actions)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			statement.Actions = actionsSet
			// principals
			var principals = make([]string, len(item.Principals))
			copy(principals, item.Principals)
			principalsSet, diags := types.SetValueFrom(ctx, types.StringType, principals)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			statement.Principals = principalsSet
			//conditions
			conditions := []ConditionsResourceModel{}
			for _, cond := range item.Conditions {
				var condition ConditionsResourceModel
				condition.Operator = types.StringValue(cond.Operator)
				// max_keys
				var maxKeys = make([]string, len(cond.MaxKeys))
				copy(maxKeys, cond.MaxKeys)
				maxKeysSet, diags := types.SetValueFrom(ctx, types.StringType, maxKeys)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if len(cond.MaxKeys) == 0 {
					condition.MaxKeys = types.SetNull(types.StringType)
				} else {
					condition.MaxKeys = maxKeysSet
				}
				// delimiters
				var delimiters = make([]string, len(cond.Delimiters))
				copy(delimiters, cond.Delimiters)
				delimitersSet, diags := types.SetValueFrom(ctx, types.StringType, delimiters)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if len(cond.Delimiters) == 0 {
					condition.Delimiters = types.SetNull(types.StringType)
				} else {
					condition.Delimiters = delimitersSet
				}
				// source_ips
				var sourceIPs = make([]string, len(cond.SourceIPs))
				copy(sourceIPs, cond.SourceIPs)
				sourceIPsSet, diags := types.SetValueFrom(ctx, types.StringType, sourceIPs)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if len(cond.SourceIPs) == 0 {
					condition.SourceIPs = types.SetNull(types.StringType)
				} else {
					condition.SourceIPs = sourceIPsSet
				}
				// prefixes
				var prefixes = make([]string, len(cond.Prefixes))
				copy(prefixes, cond.Prefixes)
				prefixesSet, diags := types.SetValueFrom(ctx, types.StringType, prefixes)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if len(cond.Prefixes) == 0 {
					condition.Prefixes = types.SetNull(types.StringType)
				} else {
					condition.Prefixes = prefixesSet
				}
				// usernames
				var usernames = make([]string, len(cond.Usernames))
				copy(usernames, cond.Usernames)
				usernamesSet, diags := types.SetValueFrom(ctx, types.StringType, usernames)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				condition.Usernames = usernamesSet
				conditions = append(conditions, condition)
			}
			statement.Conditions = conditions

			policy.Statements = append(policy.Statements, statement)
		}
		policyValue, diags := types.ObjectValueFrom(ctx, policy.attributeTypes(), policy)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Policy = policyValue
	}

	// qos_policy
	if restInfo.QoSPolicy == nil || (restInfo.QoSPolicy.Name == "" &&
		restInfo.QoSPolicy.MaxThroughputIops == 0 && restInfo.QoSPolicy.MaxThroughputMbps == 0 &&
		restInfo.QoSPolicy.MinThroughputIops == 0 && restInfo.QoSPolicy.MinThroughputMbps == 0) {
		data.QoSPolicy = types.ObjectNull(QoSPolicyResourceModel{}.attributeTypes())
	} else {
		qos := restInfo.QoSPolicy
		qosName := types.StringNull()
		if qos.Name != "" {
			qosName = types.StringValue(qos.Name)
		}
		maxIops := types.Int64Null()
		if qos.MaxThroughputIops != 0 {
			maxIops = types.Int64Value(int64(qos.MaxThroughputIops))
		}
		maxMbps := types.Int64Null()
		if qos.MaxThroughputMbps != 0 {
			maxMbps = types.Int64Value(int64(qos.MaxThroughputMbps))
		}
		minIops := types.Int64Null()
		if qos.MinThroughputIops != 0 {
			minIops = types.Int64Value(int64(qos.MinThroughputIops))
		}
		minMbps := types.Int64Null()
		if qos.MinThroughputMbps != 0 {
			minMbps = types.Int64Value(int64(qos.MinThroughputMbps))
		}
		qosModel := QoSPolicyResourceModel{
			Name:              qosName,
			MaxThroughputIops: maxIops,
			MaxThroughputMbps: maxMbps,
			MinThroughputIops: minIops,
			MinThroughputMbps: minMbps,
		}
		qosValue, diags := types.ObjectValueFrom(ctx, qosModel.attributeTypes(), qosModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.QoSPolicy = qosValue
	}

	// snapshot_policy
	if restInfo.SnapshotPolicy.Name != "" {
		data.SnapshotPolicy = types.StringValue(restInfo.SnapshotPolicy.Name)
	} else {
		data.SnapshotPolicy = types.StringNull()
	}

	// audit_event_selector
	if restInfo.AuditEventSelector == nil || (restInfo.AuditEventSelector.Access == "" && restInfo.AuditEventSelector.Permission == "") {
		data.AuditEventSelector = types.ObjectNull(AuditEventSelectorResourceModel{}.attributeTypes())
	} else {
		selector := AuditEventSelectorResourceModel{
			Access:     types.StringNull(),
			Permission: types.StringNull(),
		}
		if restInfo.AuditEventSelector.Access != "" {
			selector.Access = types.StringValue(restInfo.AuditEventSelector.Access)
		}
		if restInfo.AuditEventSelector.Permission != "" {
			selector.Permission = types.StringValue(restInfo.AuditEventSelector.Permission)
		}
		selectorValue, diags := types.ObjectValueFrom(ctx, selector.attributeTypes(), selector)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.AuditEventSelector = selectorValue
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *ProtocolsS3BucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProtocolsS3BucketResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	trackPolicy := data != nil && !data.Policy.IsNull() && !data.Policy.IsUnknown()

	var body interfaces.ProtocolsS3BucketResourceBodyDataModel
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

	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()
	if !data.Size.IsNull() {
		body.Size = data.Size.ValueInt64()
	}
	if !data.Comment.IsNull() {
		commentValue := data.Comment.ValueString()
		body.Comment = &commentValue
	}
	if !data.Type.IsNull() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 12 {
			body.Type = data.Type.ValueString()
		} else {
			errors = append(errors, "type requires ONTAP 9.12 or later")
		}
	}
	if !data.NASPath.IsNull() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 12 {
			body.NASPath = data.NASPath.ValueString()
		} else {
			errors = append(errors, "nas_path requires ONTAP 9.12 or later")
		}
	}
	if !data.VersioningState.IsNull() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 11 {
			body.VersioningState = data.VersioningState.ValueString()
		} else {
			errors = append(errors, "versioning_state requires ONTAP 9.11 or later")
		}
	}
	if !data.SnapshotPolicy.IsNull() && !data.SnapshotPolicy.IsUnknown() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 16 {
			if snapshotPolicy := data.SnapshotPolicy.ValueString(); snapshotPolicy != "" {
				body.SnapshotPolicy = snapshotPolicy
			}
		} else {
			errors = append(errors, "snapshot_policy requires ONTAP 9.16 or later")
		}
	}

	if !data.Policy.IsNull() && !data.Policy.IsUnknown() {
		var policy PolicyResourceModel
		diags := data.Policy.As(ctx, &policy, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		if len(policy.Statements) > 0 {
			body.Policy = &interfaces.PolicyDataModel{}
			for _, stmt := range policy.Statements {
			// decode directly into API types to correctly handle empty sets (actions=[], principals=[])
			var apiStmt interfaces.StatementsDataModel
			if !stmt.SID.IsNull() && !stmt.SID.IsUnknown() {
				apiStmt.SID = stmt.SID.ValueString()
			}
			if !stmt.Effect.IsNull() && !stmt.Effect.IsUnknown() {
				apiStmt.Effect = stmt.Effect.ValueString()
			}
			if !stmt.Resources.IsNull() && !stmt.Resources.IsUnknown() {
				resp.Diagnostics.Append(stmt.Resources.ElementsAs(ctx, &apiStmt.Resources, false)...)
			}
			if !stmt.Actions.IsNull() && !stmt.Actions.IsUnknown() {
				resp.Diagnostics.Append(stmt.Actions.ElementsAs(ctx, &apiStmt.Actions, false)...)
			}
			if !stmt.Principals.IsNull() && !stmt.Principals.IsUnknown() {
				resp.Diagnostics.Append(stmt.Principals.ElementsAs(ctx, &apiStmt.Principals, false)...)
			}
			// initialize to empty slice so conditions=[] sends [] rather than null
			apiStmt.Conditions = []interfaces.ConditionsDataModel{}
			for _, cond := range stmt.Conditions {
				var apiCond interfaces.ConditionsDataModel
				if !cond.Operator.IsNull() && !cond.Operator.IsUnknown() {
					apiCond.Operator = cond.Operator.ValueString()
				}
				if !cond.MaxKeys.IsNull() && !cond.MaxKeys.IsUnknown() {
					resp.Diagnostics.Append(cond.MaxKeys.ElementsAs(ctx, &apiCond.MaxKeys, false)...)
				}
				if !cond.Delimiters.IsNull() && !cond.Delimiters.IsUnknown() {
					resp.Diagnostics.Append(cond.Delimiters.ElementsAs(ctx, &apiCond.Delimiters, false)...)
				}
				if !cond.SourceIPs.IsNull() && !cond.SourceIPs.IsUnknown() {
					resp.Diagnostics.Append(cond.SourceIPs.ElementsAs(ctx, &apiCond.SourceIPs, false)...)
				}
				if !cond.Prefixes.IsNull() && !cond.Prefixes.IsUnknown() {
					resp.Diagnostics.Append(cond.Prefixes.ElementsAs(ctx, &apiCond.Prefixes, false)...)
				}
				if !cond.Usernames.IsNull() && !cond.Usernames.IsUnknown() {
					resp.Diagnostics.Append(cond.Usernames.ElementsAs(ctx, &apiCond.Usernames, false)...)
				}
				apiStmt.Conditions = append(apiStmt.Conditions, apiCond)
			}
			body.Policy.Statements = append(body.Policy.Statements, apiStmt)
		}
		}
	}

	if !data.QoSPolicy.IsNull() && !data.QoSPolicy.IsUnknown() {
		var qosModel QoSPolicyResourceModel
		diags := data.QoSPolicy.As(ctx, &qosModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		qos := &interfaces.QoSPolicyDataModel{}
		hasValue := false
		if !qosModel.Name.IsNull() && !qosModel.Name.IsUnknown() && qosModel.Name.ValueString() != "" {
			qos.Name = qosModel.Name.ValueString()
			hasValue = true
		}
		if !qosModel.MaxThroughputIops.IsNull() && !qosModel.MaxThroughputIops.IsUnknown() {
			qos.MaxThroughputIops = int(qosModel.MaxThroughputIops.ValueInt64())
			hasValue = true
		}
		if !qosModel.MaxThroughputMbps.IsNull() && !qosModel.MaxThroughputMbps.IsUnknown() {
			qos.MaxThroughputMbps = int(qosModel.MaxThroughputMbps.ValueInt64())
			hasValue = true
		}
		if !qosModel.MinThroughputIops.IsNull() && !qosModel.MinThroughputIops.IsUnknown() {
			qos.MinThroughputIops = int(qosModel.MinThroughputIops.ValueInt64())
			hasValue = true
		}
		if !qosModel.MinThroughputMbps.IsNull() && !qosModel.MinThroughputMbps.IsUnknown() {
			qos.MinThroughputMbps = int(qosModel.MinThroughputMbps.ValueInt64())
			hasValue = true
		}
		if hasValue {
			body.QoSPolicy = qos
		}
	}

	if !data.AuditEventSelector.IsNull() && !data.AuditEventSelector.IsUnknown() {
		var selectorModel AuditEventSelectorResourceModel
		diags := data.AuditEventSelector.As(ctx, &selectorModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		selector := &interfaces.AuditEventSelectorDataModel{}
		hasValue := false
		// only set audit_event_selector fields if they have non-empty values
		if !selectorModel.Access.IsNull() && !selectorModel.Access.IsUnknown() && selectorModel.Access.ValueString() != "" {
			selector.Access = selectorModel.Access.ValueString()
			hasValue = true
		}
		if !selectorModel.Permission.IsNull() && !selectorModel.Permission.IsUnknown() && selectorModel.Permission.ValueString() != "" {
			selector.Permission = selectorModel.Permission.ValueString()
			hasValue = true
		}
		if hasValue {
			if cluster.Version.Generation == 9 && cluster.Version.Major >= 10 {
				body.AuditEventSelector = selector
			} else {
				errors = append(errors, "audit_event_selector requires ONTAP 9.10 or later")
			}
		}
	}

	if !data.Aggregates.IsNull() {
		resp.Diagnostics.Append(data.Aggregates.ElementsAs(ctx, &body.Aggregates, false)...)
	}
	if !data.ConstituentsPerAggregate.IsNull() {
		body.ConstituentsPerAggregate = data.ConstituentsPerAggregate.ValueInt64()
	}

	if len(errors) > 0 {
		errorsString := strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("The following options have ONTAP version constraints: %#v", errorsString))
		return
	}

	resource, err := interfaces.CreateProtocolsS3Bucket(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.ID = types.StringValue(resource.UUID)
	restInfo, err := interfaces.GetProtocolsS3Bucket(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsS3Bucket
		return
	}

	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(resource.Name)
	data.Size = types.Int64Value(resource.Size)
	data.Comment = types.StringValue(resource.Comment)
	data.Type = types.StringValue(resource.Type)
	data.NASPath = types.StringValue(resource.NASPath)
	data.VersioningState = types.StringValue(resource.VersioningState)

	// policy
	if restInfo.Policy == nil {
		// ONTAP may omit the policy object when it is empty, even if previously sent statements=[]
		// keep it as an empty policy (statements=[]) to avoid flipping between null and []
		if trackPolicy {
			emptyPolicy := PolicyResourceModel{Statements: []StatementsResourceModel{}}
			policyValue, diags := types.ObjectValueFrom(ctx, emptyPolicy.attributeTypes(), emptyPolicy)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.Policy = policyValue
		} else {
			data.Policy = types.ObjectNull(PolicyResourceModel{}.attributeTypes())
		}
	} else {
		var policy PolicyResourceModel
		policy.Statements = []StatementsResourceModel{}
		for _, item := range restInfo.Policy.Statements {
			var statement StatementsResourceModel
			statement.SID = types.StringValue(item.SID)
			statement.Effect = types.StringValue(item.Effect)
			// resources
			var resources = make([]string, len(item.Resources))
			copy(resources, item.Resources)
			resourcesSet, diags := types.SetValueFrom(ctx, types.StringType, resources)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			statement.Resources = resourcesSet
			//actions
			var actions = make([]string, len(item.Actions))
			copy(actions, item.Actions)
			actionsSet, diags := types.SetValueFrom(ctx, types.StringType, actions)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			statement.Actions = actionsSet
			// principals
			var principals = make([]string, len(item.Principals))
			copy(principals, item.Principals)
			principalsSet, diags := types.SetValueFrom(ctx, types.StringType, principals)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			statement.Principals = principalsSet
			//conditions
			conditions := []ConditionsResourceModel{}
			for _, cond := range item.Conditions {
				var condition ConditionsResourceModel
				condition.Operator = types.StringValue(cond.Operator)
				// max_keys
				var maxKeys = make([]string, len(cond.MaxKeys))
				copy(maxKeys, cond.MaxKeys)
				maxKeysSet, diags := types.SetValueFrom(ctx, types.StringType, maxKeys)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if len(cond.MaxKeys) == 0 {
					condition.MaxKeys = types.SetNull(types.StringType)
				} else {
					condition.MaxKeys = maxKeysSet
				}
				// delimiters
				var delimiters = make([]string, len(cond.Delimiters))
				copy(delimiters, cond.Delimiters)
				delimitersSet, diags := types.SetValueFrom(ctx, types.StringType, delimiters)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if len(cond.Delimiters) == 0 {
					condition.Delimiters = types.SetNull(types.StringType)
				} else {
					condition.Delimiters = delimitersSet
				}
				// source_ips
				var sourceIPs = make([]string, len(cond.SourceIPs))
				copy(sourceIPs, cond.SourceIPs)
				sourceIPsSet, diags := types.SetValueFrom(ctx, types.StringType, sourceIPs)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if len(cond.SourceIPs) == 0 {
					condition.SourceIPs = types.SetNull(types.StringType)
				} else {
					condition.SourceIPs = sourceIPsSet
				}
				// prefixes
				var prefixes = make([]string, len(cond.Prefixes))
				copy(prefixes, cond.Prefixes)
				prefixesSet, diags := types.SetValueFrom(ctx, types.StringType, prefixes)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if len(cond.Prefixes) == 0 {
					condition.Prefixes = types.SetNull(types.StringType)
				} else {
					condition.Prefixes = prefixesSet
				}
				// usernames
				var usernames = make([]string, len(cond.Usernames))
				copy(usernames, cond.Usernames)
				usernamesSet, diags := types.SetValueFrom(ctx, types.StringType, usernames)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				condition.Usernames = usernamesSet
				conditions = append(conditions, condition)
			}
			statement.Conditions = conditions

			policy.Statements = append(policy.Statements, statement)
		}
		policyValue, diags := types.ObjectValueFrom(ctx, policy.attributeTypes(), policy)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Policy = policyValue
	}

	// qos_policy
	if restInfo.QoSPolicy == nil || (restInfo.QoSPolicy.Name == "" &&
		restInfo.QoSPolicy.MaxThroughputIops == 0 && restInfo.QoSPolicy.MaxThroughputMbps == 0 &&
		restInfo.QoSPolicy.MinThroughputIops == 0 && restInfo.QoSPolicy.MinThroughputMbps == 0) {
		data.QoSPolicy = types.ObjectNull(QoSPolicyResourceModel{}.attributeTypes())
	} else {
		qos := restInfo.QoSPolicy
		qosName := types.StringNull()
		if qos.Name != "" {
			qosName = types.StringValue(qos.Name)
		}
		maxIops := types.Int64Null()
		if qos.MaxThroughputIops != 0 {
			maxIops = types.Int64Value(int64(qos.MaxThroughputIops))
		}
		maxMbps := types.Int64Null()
		if qos.MaxThroughputMbps != 0 {
			maxMbps = types.Int64Value(int64(qos.MaxThroughputMbps))
		}
		minIops := types.Int64Null()
		if qos.MinThroughputIops != 0 {
			minIops = types.Int64Value(int64(qos.MinThroughputIops))
		}
		minMbps := types.Int64Null()
		if qos.MinThroughputMbps != 0 {
			minMbps = types.Int64Value(int64(qos.MinThroughputMbps))
		}
		qosModel := QoSPolicyResourceModel{
			Name:              qosName,
			MaxThroughputIops: maxIops,
			MaxThroughputMbps: maxMbps,
			MinThroughputIops: minIops,
			MinThroughputMbps: minMbps,
		}
		qosValue, diags := types.ObjectValueFrom(ctx, qosModel.attributeTypes(), qosModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.QoSPolicy = qosValue
	}

	// snapshot_policy
	if restInfo.SnapshotPolicy.Name != "" {
		data.SnapshotPolicy = types.StringValue(restInfo.SnapshotPolicy.Name)
	} else {
		data.SnapshotPolicy = types.StringNull()
	}

	// audit_event_selector
	if restInfo.AuditEventSelector == nil || (restInfo.AuditEventSelector.Access == "" && restInfo.AuditEventSelector.Permission == "") {
		data.AuditEventSelector = types.ObjectNull(AuditEventSelectorResourceModel{}.attributeTypes())
	} else {
		selector := AuditEventSelectorResourceModel{
			Access:     types.StringNull(),
			Permission: types.StringNull(),
		}
		if restInfo.AuditEventSelector.Access != "" {
			selector.Access = types.StringValue(restInfo.AuditEventSelector.Access)
		}
		if restInfo.AuditEventSelector.Permission != "" {
			selector.Permission = types.StringValue(restInfo.AuditEventSelector.Permission)
		}
		selectorValue, diags := types.ObjectValueFrom(ctx, selector.attributeTypes(), selector)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.AuditEventSelector = selectorValue
	}

	tflog.Trace(ctx, fmt.Sprintf("created S3 bucket resource, ID=%s", data.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsS3BucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *ProtocolsS3BucketResourceModel

	// read Terraform plan & state data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	trackPolicy := plan != nil && !plan.Policy.IsNull() && !plan.Policy.IsUnknown()
	if !trackPolicy {
		trackPolicy = state != nil && !state.Policy.IsNull() && !state.Policy.IsUnknown()
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

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
	var body interfaces.ProtocolsS3BucketResourceBodyDataModel

	var errors []string

	body.Name = plan.Name.ValueString()
	if !plan.Size.Equal(state.Size) {
		body.Size = plan.Size.ValueInt64()
	}
	if !plan.Comment.Equal(state.Comment) {
		commentValue := plan.Comment.ValueString()
		body.Comment = &commentValue
	}
	if !plan.VersioningState.Equal(state.VersioningState) {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 11 {
			body.VersioningState = plan.VersioningState.ValueString()
		} else {
			errors = append(errors, "versioning_state requires ONTAP 9.11 or later")
		}
	}
	if !plan.SnapshotPolicy.Equal(state.SnapshotPolicy) &&
		!plan.SnapshotPolicy.IsNull() && !plan.SnapshotPolicy.IsUnknown() {
		if plan.SnapshotPolicy.ValueString() == "" {
			// snapshot policy cannot be modified as empty name
			errorHandler.MakeAndReportError("update S3 bucket", "snapshot_policy cannot be updated with empty string")
			return
		}
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 16 {
			body.SnapshotPolicy = plan.SnapshotPolicy.ValueString()
		} else {
			errors = append(errors, "snapshot_policy requires ONTAP 9.16 or later")
		}
	}
	// policy
	// only attempt to update policy when it exists in the planned value
	planHasPolicy := !plan.Policy.IsNull() && !plan.Policy.IsUnknown()
	planPolicy := PolicyResourceModel{}
	if planHasPolicy {
		diags := plan.Policy.As(ctx, &planPolicy, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
	planHasStatements := planHasPolicy && len(planPolicy.Statements) > 0

	if planHasPolicy {
		if state == nil || plan.Policy.IsUnknown() || state.Policy.IsUnknown() || !plan.Policy.Equal(state.Policy) {
			if planHasStatements {
				// plan has statements - update the policy with plan data, decoded directly into API types
				body.Policy = &interfaces.PolicyDataModel{}
				for _, stmt := range planPolicy.Statements {
					var apiStmt interfaces.StatementsDataModel
					if !stmt.SID.IsNull() && !stmt.SID.IsUnknown() {
						apiStmt.SID = stmt.SID.ValueString()
					}
					if !stmt.Effect.IsNull() && !stmt.Effect.IsUnknown() {
						apiStmt.Effect = stmt.Effect.ValueString()
					}
					if !stmt.Resources.IsNull() && !stmt.Resources.IsUnknown() {
						resp.Diagnostics.Append(stmt.Resources.ElementsAs(ctx, &apiStmt.Resources, false)...)
					}
					if !stmt.Actions.IsNull() && !stmt.Actions.IsUnknown() {
						resp.Diagnostics.Append(stmt.Actions.ElementsAs(ctx, &apiStmt.Actions, false)...)
					}
					if !stmt.Principals.IsNull() && !stmt.Principals.IsUnknown() {
						resp.Diagnostics.Append(stmt.Principals.ElementsAs(ctx, &apiStmt.Principals, false)...)
					}
					// initialize to empty slice so conditions=[] sends [] rather than null
					apiStmt.Conditions = []interfaces.ConditionsDataModel{}
					for _, cond := range stmt.Conditions {
						var apiCond interfaces.ConditionsDataModel
						if !cond.Operator.IsNull() && !cond.Operator.IsUnknown() {
							apiCond.Operator = cond.Operator.ValueString()
						}
						if !cond.MaxKeys.IsNull() && !cond.MaxKeys.IsUnknown() {
							resp.Diagnostics.Append(cond.MaxKeys.ElementsAs(ctx, &apiCond.MaxKeys, false)...)
						}
						if !cond.Delimiters.IsNull() && !cond.Delimiters.IsUnknown() {
							resp.Diagnostics.Append(cond.Delimiters.ElementsAs(ctx, &apiCond.Delimiters, false)...)
						}
						if !cond.SourceIPs.IsNull() && !cond.SourceIPs.IsUnknown() {
							resp.Diagnostics.Append(cond.SourceIPs.ElementsAs(ctx, &apiCond.SourceIPs, false)...)
						}
						if !cond.Prefixes.IsNull() && !cond.Prefixes.IsUnknown() {
							resp.Diagnostics.Append(cond.Prefixes.ElementsAs(ctx, &apiCond.Prefixes, false)...)
						}
						if !cond.Usernames.IsNull() && !cond.Usernames.IsUnknown() {
							resp.Diagnostics.Append(cond.Usernames.ElementsAs(ctx, &apiCond.Usernames, false)...)
						}
						apiStmt.Conditions = append(apiStmt.Conditions, apiCond)
					}
					body.Policy.Statements = append(body.Policy.Statements, apiStmt)
				}
			} else {
				// plan has policy but no statements
				// initialize the policy structure explicitly to ensure statements field is sent
				body.Policy = &interfaces.PolicyDataModel{
					Statements: []interfaces.StatementsDataModel{},
				}
			}
		}
	}

	// qos_policy
	if !plan.QoSPolicy.IsNull() && !plan.QoSPolicy.IsUnknown() {
		var planQoS QoSPolicyResourceModel
		diags := plan.QoSPolicy.As(ctx, &planQoS, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		var stateQoS *QoSPolicyResourceModel
		if state != nil && !state.QoSPolicy.IsNull() && !state.QoSPolicy.IsUnknown() {
			var decoded QoSPolicyResourceModel
			diags := state.QoSPolicy.As(ctx, &decoded, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			stateQoS = &decoded
		}

		qos := &interfaces.QoSPolicyDataModel{}
		hasValue := false

		// only send fields that are explicitly known and differ from state
		if !planQoS.Name.IsNull() && !planQoS.Name.IsUnknown() && planQoS.Name.ValueString() != "" &&
			(stateQoS == nil || !planQoS.Name.Equal(stateQoS.Name)) {
			qos.Name = planQoS.Name.ValueString()
			hasValue = true
		}
		if !planQoS.MaxThroughputIops.IsNull() && !planQoS.MaxThroughputIops.IsUnknown() &&
			(stateQoS == nil || !planQoS.MaxThroughputIops.Equal(stateQoS.MaxThroughputIops)) {
			qos.MaxThroughputIops = int(planQoS.MaxThroughputIops.ValueInt64())
			hasValue = true
		}
		if !planQoS.MaxThroughputMbps.IsNull() && !planQoS.MaxThroughputMbps.IsUnknown() &&
			(stateQoS == nil || !planQoS.MaxThroughputMbps.Equal(stateQoS.MaxThroughputMbps)) {
			qos.MaxThroughputMbps = int(planQoS.MaxThroughputMbps.ValueInt64())
			hasValue = true
		}
		if !planQoS.MinThroughputIops.IsNull() && !planQoS.MinThroughputIops.IsUnknown() &&
			(stateQoS == nil || !planQoS.MinThroughputIops.Equal(stateQoS.MinThroughputIops)) {
			qos.MinThroughputIops = int(planQoS.MinThroughputIops.ValueInt64())
			hasValue = true
		}
		if !planQoS.MinThroughputMbps.IsNull() && !planQoS.MinThroughputMbps.IsUnknown() &&
			(stateQoS == nil || !planQoS.MinThroughputMbps.Equal(stateQoS.MinThroughputMbps)) {
			qos.MinThroughputMbps = int(planQoS.MinThroughputMbps.ValueInt64())
			hasValue = true
		}

		// name is treated as API-computed/auto-generated
		if hasValue {
			body.QoSPolicy = qos
		}
	}

	// audit_event_selector
	if !plan.AuditEventSelector.IsNull() && !plan.AuditEventSelector.IsUnknown() &&
		(state == nil || state.AuditEventSelector.IsUnknown() || !plan.AuditEventSelector.Equal(state.AuditEventSelector)) {
		var selectorModel AuditEventSelectorResourceModel
		diags := plan.AuditEventSelector.As(ctx, &selectorModel, basetypes.ObjectAsOptions{UnhandledNullAsEmpty: true, UnhandledUnknownAsEmpty: true})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		selector := &interfaces.AuditEventSelectorDataModel{}
		hasValue := false
		if !selectorModel.Access.IsNull() && !selectorModel.Access.IsUnknown() && selectorModel.Access.ValueString() != "" {
			selector.Access = selectorModel.Access.ValueString()
			hasValue = true
		}
		if !selectorModel.Permission.IsNull() && !selectorModel.Permission.IsUnknown() && selectorModel.Permission.ValueString() != "" {
			selector.Permission = selectorModel.Permission.ValueString()
			hasValue = true
		}
		if hasValue {
			if cluster.Version.Generation == 9 && cluster.Version.Major >= 10 {
				body.AuditEventSelector = selector
			} else {
				errors = append(errors, "audit_event_selector requires ONTAP 9.10 or later")
			}
		}
	}

	if len(errors) > 0 {
		errorsString := strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("The following options have ONTAP version constraints: %#v", errorsString))
		return
	}

	err = interfaces.UpdateProtocolsS3Bucket(errorHandler, *client, body, plan.ID.ValueString(), svm.UUID)
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetProtocolsS3Bucket(errorHandler, *client, plan.Name.ValueString(), plan.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsS3Bucket
		return
	}

	// Update the computed parameters
	plan.ID = types.StringValue(restInfo.UUID)
	plan.SVMName = types.StringValue(restInfo.SVM.Name)
	plan.Name = types.StringValue(restInfo.Name)
	plan.Size = types.Int64Value(restInfo.Size)
	plan.Comment = types.StringValue(restInfo.Comment)
	plan.Type = types.StringValue(restInfo.Type)
	plan.NASPath = types.StringValue(restInfo.NASPath)
	plan.VersioningState = types.StringValue(restInfo.VersioningState)

	// policy
	if restInfo.Policy == nil {
		// ONTAP may omit the policy object when it is empty, even if previously sent statements=[]
		// keep it as an empty policy (statements=[]) to avoid flipping between null and []
		if trackPolicy {
			emptyPolicy := PolicyResourceModel{Statements: []StatementsResourceModel{}}
			policyValue, diags := types.ObjectValueFrom(ctx, emptyPolicy.attributeTypes(), emptyPolicy)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			plan.Policy = policyValue
		} else {
			plan.Policy = types.ObjectNull(PolicyResourceModel{}.attributeTypes())
		}
	} else {
		var policy PolicyResourceModel
		policy.Statements = []StatementsResourceModel{}
		for _, item := range restInfo.Policy.Statements {
			var statement StatementsResourceModel
			statement.SID = types.StringValue(item.SID)
			statement.Effect = types.StringValue(item.Effect)
			// resources
			var resources = make([]string, len(item.Resources))
			copy(resources, item.Resources)
			resourcesSet, diags := types.SetValueFrom(ctx, types.StringType, resources)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			statement.Resources = resourcesSet
			//actions
			var actions = make([]string, len(item.Actions))
			copy(actions, item.Actions)
			actionsSet, diags := types.SetValueFrom(ctx, types.StringType, actions)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			statement.Actions = actionsSet
			// principals
			var principals = make([]string, len(item.Principals))
			copy(principals, item.Principals)
			principalsSet, diags := types.SetValueFrom(ctx, types.StringType, principals)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			statement.Principals = principalsSet
			//conditions
			conditions := []ConditionsResourceModel{}
			for _, cond := range item.Conditions {
				var condition ConditionsResourceModel
				condition.Operator = types.StringValue(cond.Operator)
				// max_keys
				var maxKeys = make([]string, len(cond.MaxKeys))
				copy(maxKeys, cond.MaxKeys)
				maxKeysSet, diags := types.SetValueFrom(ctx, types.StringType, maxKeys)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if len(cond.MaxKeys) == 0 {
					condition.MaxKeys = types.SetNull(types.StringType)
				} else {
					condition.MaxKeys = maxKeysSet
				}
				// delimiters
				var delimiters = make([]string, len(cond.Delimiters))
				copy(delimiters, cond.Delimiters)
				delimitersSet, diags := types.SetValueFrom(ctx, types.StringType, delimiters)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if len(cond.Delimiters) == 0 {
					condition.Delimiters = types.SetNull(types.StringType)
				} else {
					condition.Delimiters = delimitersSet
				}
				// source_ips
				var sourceIPs = make([]string, len(cond.SourceIPs))
				copy(sourceIPs, cond.SourceIPs)
				sourceIPsSet, diags := types.SetValueFrom(ctx, types.StringType, sourceIPs)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if len(cond.SourceIPs) == 0 {
					condition.SourceIPs = types.SetNull(types.StringType)
				} else {
					condition.SourceIPs = sourceIPsSet
				}
				// prefixes
				var prefixes = make([]string, len(cond.Prefixes))
				copy(prefixes, cond.Prefixes)
				prefixesSet, diags := types.SetValueFrom(ctx, types.StringType, prefixes)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				if len(cond.Prefixes) == 0 {
					condition.Prefixes = types.SetNull(types.StringType)
				} else {
					condition.Prefixes = prefixesSet
				}
				// usernames
				var usernames = make([]string, len(cond.Usernames))
				copy(usernames, cond.Usernames)
				usernamesSet, diags := types.SetValueFrom(ctx, types.StringType, usernames)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				condition.Usernames = usernamesSet
				conditions = append(conditions, condition)
			}
			statement.Conditions = conditions

			policy.Statements = append(policy.Statements, statement)
		}
		policyValue, diags := types.ObjectValueFrom(ctx, policy.attributeTypes(), policy)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Policy = policyValue
	}

	// qos_policy
	if restInfo.QoSPolicy == nil || (restInfo.QoSPolicy.Name == "" &&
		restInfo.QoSPolicy.MaxThroughputIops == 0 && restInfo.QoSPolicy.MaxThroughputMbps == 0 &&
		restInfo.QoSPolicy.MinThroughputIops == 0 && restInfo.QoSPolicy.MinThroughputMbps == 0) {
		plan.QoSPolicy = types.ObjectNull(QoSPolicyResourceModel{}.attributeTypes())
	} else {
		qos := restInfo.QoSPolicy
		qosName := types.StringNull()
		if qos.Name != "" {
			qosName = types.StringValue(qos.Name)
		}
		maxIops := types.Int64Null()
		if qos.MaxThroughputIops != 0 {
			maxIops = types.Int64Value(int64(qos.MaxThroughputIops))
		}
		maxMbps := types.Int64Null()
		if qos.MaxThroughputMbps != 0 {
			maxMbps = types.Int64Value(int64(qos.MaxThroughputMbps))
		}
		minIops := types.Int64Null()
		if qos.MinThroughputIops != 0 {
			minIops = types.Int64Value(int64(qos.MinThroughputIops))
		}
		minMbps := types.Int64Null()
		if qos.MinThroughputMbps != 0 {
			minMbps = types.Int64Value(int64(qos.MinThroughputMbps))
		}
		qosModel := QoSPolicyResourceModel{
			Name:              qosName,
			MaxThroughputIops: maxIops,
			MaxThroughputMbps: maxMbps,
			MinThroughputIops: minIops,
			MinThroughputMbps: minMbps,
		}
		qosValue, diags := types.ObjectValueFrom(ctx, qosModel.attributeTypes(), qosModel)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.QoSPolicy = qosValue
	}

	// snapshot_policy
	if restInfo.SnapshotPolicy.Name != "" {
		plan.SnapshotPolicy = types.StringValue(restInfo.SnapshotPolicy.Name)
	} else {
		plan.SnapshotPolicy = types.StringNull()
	}

	// audit_event_selector
	if restInfo.AuditEventSelector == nil || (restInfo.AuditEventSelector.Access == "" && restInfo.AuditEventSelector.Permission == "") {
		plan.AuditEventSelector = types.ObjectNull(AuditEventSelectorResourceModel{}.attributeTypes())
	} else {
		selector := AuditEventSelectorResourceModel{
			Access:     types.StringNull(),
			Permission: types.StringNull(),
		}
		if restInfo.AuditEventSelector.Access != "" {
			selector.Access = types.StringValue(restInfo.AuditEventSelector.Access)
		}
		if restInfo.AuditEventSelector.Permission != "" {
			selector.Permission = types.StringValue(restInfo.AuditEventSelector.Permission)
		}
		selectorValue, diags := types.ObjectValueFrom(ctx, selector.attributeTypes(), selector)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.AuditEventSelector = selectorValue
	}

	tflog.Debug(ctx, fmt.Sprintf("updated S3 bucket resource: ID=%s", plan.ID))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsS3BucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsS3BucketResourceModel

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

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "S3 bucket UUID is null")
		return
	}

	err = interfaces.DeleteProtocolsS3Bucket(errorHandler, *client, data.ID.ValueString(), svm.UUID)
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsS3BucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
