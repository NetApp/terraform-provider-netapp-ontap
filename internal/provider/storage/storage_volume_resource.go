package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &StorageVolumeResource{}
var _ resource.ResourceWithImportState = &StorageVolumeResource{}

// var _ resource.ResourceWithModifyPlan = &StorageVolumeResource{}

// NewStorageVolumeResource is a helper function to simplify the provider implementation.
func NewStorageVolumeResource() resource.Resource {
	return &StorageVolumeResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "volume",
		},
	}
}

// NewStorageVolumeResourceAlias is a helper function to simplify the provider implementation.
func NewStorageVolumeResourceAlias() resource.Resource {
	return &StorageVolumeResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "storage_volume_resource",
		},
	}
}

// StorageVolumeResource defines the resource implementation.
type StorageVolumeResource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageVolumeResourceModel describes the resource data model.
type StorageVolumeResourceModel struct {
	CxProfileName  types.String                      `tfsdk:"cx_profile_name"`
	Name           types.String                      `tfsdk:"name"`
	SVMName        types.String                      `tfsdk:"svm_name"`
	State          types.String                      `tfsdk:"state"`
	Type           types.String                      `tfsdk:"type"`
	SpaceGuarantee types.String                      `tfsdk:"space_guarantee"`
	Encrypt        types.Bool                        `tfsdk:"encryption"`
	SnapshotPolicy types.String                      `tfsdk:"snapshot_policy"`
	Language       types.String                      `tfsdk:"language"`
	QOSPolicyGroup types.String                      `tfsdk:"qos_policy_group"`
	Comment        types.String                      `tfsdk:"comment"`
	Aggregates     []StorageVolumeResourceAggregates `tfsdk:"aggregates"`
	ID             types.String                      `tfsdk:"id"`
	Space          types.Object                      `tfsdk:"space"`
	Nas            types.Object                      `tfsdk:"nas"`
	Tiering        types.Object                      `tfsdk:"tiering"`
	Efficiency     types.Object                      `tfsdk:"efficiency"`
	SnapLock       types.Object                      `tfsdk:"snaplock"`
	Analytics      types.Object                      `tfsdk:"analytics"`
	Autosize       types.Object                      `tfsdk:"autosize"`
}

// StorageVolumeResourceAggregates describes the analytics model.
type StorageVolumeResourceAggregates struct {
	Name types.String `tfsdk:"name"`
}

// StorageVolumeResourceAnalytics describes the analytics model.
type StorageVolumeResourceAnalytics struct {
	State types.String `tfsdk:"state"`
}

// StorageVolumeResourceSnapLock describes the snaplock model.
type StorageVolumeResourceSnapLock struct {
	SnaplockType types.String `tfsdk:"type"`
}

// StorageVolumeResourceEfficiency describes the efficiency model.
type StorageVolumeResourceEfficiency struct {
	Policy      types.String `tfsdk:"policy_name"`
	Compression types.String `tfsdk:"compression"`
}

// StorageVolumeResourceTiering describes the tiering model.
type StorageVolumeResourceTiering struct {
	Policy             types.String `tfsdk:"policy_name"`
	MinimumCoolingDays types.Int64  `tfsdk:"minimum_cooling_days"`
}

// StorageVolumeResourceNas describes the Nas model.
type StorageVolumeResourceNas struct {
	ExportPolicy    types.String `tfsdk:"export_policy_name"`
	JunctionPath    types.String `tfsdk:"junction_path"`
	GroupID         types.Int64  `tfsdk:"group_id"`
	UserID          types.Int64  `tfsdk:"user_id"`
	SecurityStyle   types.String `tfsdk:"security_style"`
	UnixPermissions types.Int64  `tfsdk:"unix_permissions"`
}

// StorageVolumeResourceSpace describes the space model.
type StorageVolumeResourceSpace struct {
	Size                 types.Int64  `tfsdk:"size"`
	SizeUnit             types.String `tfsdk:"size_unit"`
	PercentSnapshotSpace types.Int64  `tfsdk:"percent_snapshot_space"`
	LogicalSpace         types.Object `tfsdk:"logical_space"`
}

// StorageVolumeResourceSpaceLogicalSpace describes the logical space model within sapce model.
type StorageVolumeResourceSpaceLogicalSpace struct {
	Enforcement types.Bool `tfsdk:"enforcement"`
	Reporting   types.Bool `tfsdk:"reporting"`
}

// StorageVolumeResourceAutosize describes the autosize model.
type StorageVolumeResourceAutosize struct {
	Minimum         types.Int64  `tfsdk:"minimum"`
	Maximum         types.Int64  `tfsdk:"maximum"`
	ShrinkThreshold types.Int64  `tfsdk:"shrink_threshold"`
	GrowThreshold   types.Int64  `tfsdk:"grow_threshold"`
	Mode            types.String `tfsdk:"mode"`
	SizeUnit        types.String `tfsdk:"size_unit"`
}

// Metadata returns the resource type name.
func (r *StorageVolumeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *StorageVolumeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Volume resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the volume to manage",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the svm to use",
				Required:            true,
			},
			"aggregates": schema.SetNestedAttribute{
				Required:            true,
				MarkdownDescription: "List of aggregates to place volume on",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the aggregate",
							Required:            true,
						},
					},
				},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Whether the specified volume is online, or not",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The volume type, either read-write (RW) or data-protection (DP)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space_guarantee": schema.StringAttribute{
				MarkdownDescription: "Space guarantee style for the volume",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"encryption": schema.BoolAttribute{
				MarkdownDescription: "Whether or not to enable Volume Encryption",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"snapshot_policy": schema.StringAttribute{
				MarkdownDescription: "The name of the snapshot policy",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "Language to use for volume",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			// with Rest API qos_policy_group and qos_adaptive_policy_group are now the same thing and cannot be set at the same time
			"qos_policy_group": schema.StringAttribute{
				MarkdownDescription: "Specifies a QoS policy group to be set on volume",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Sets a comment associated with the volume",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"space": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"size": schema.Int64Attribute{
						MarkdownDescription: "The size of the volume",
						Required:            true,
					},
					"size_unit": schema.StringAttribute{
						MarkdownDescription: "The unit used to interpret the size parameter",
						Required:            true,
					},
					"percent_snapshot_space": schema.Int64Attribute{
						MarkdownDescription: "Amount of space reserved for snapshot copies of the volume",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"logical_space": schema.SingleNestedAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
						Attributes: map[string]schema.Attribute{
							"enforcement": schema.BoolAttribute{
								MarkdownDescription: "Whether to perform logical space accounting on the volume",
								Optional:            true,
								Computed:            true,
							},
							"reporting": schema.BoolAttribute{
								MarkdownDescription: "Whether to report space logically",
								Optional:            true,
								Computed:            true,
							},
						},
					},
				},
			},
			"nas": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"export_policy_name": schema.StringAttribute{
						MarkdownDescription: "The name of the export policy",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"junction_path": schema.StringAttribute{
						MarkdownDescription: "Junction path of the volume",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"group_id": schema.Int64Attribute{
						MarkdownDescription: "The UNIX group ID for the volume",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"user_id": schema.Int64Attribute{
						MarkdownDescription: "The UNIX user ID for the volume",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"security_style": schema.StringAttribute{
						MarkdownDescription: "The security style associated to the volume",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"unix_permissions": schema.Int64Attribute{
						MarkdownDescription: "Unix permission bits in octal or symbolic format. For example, 0 is equivalent to ------------, 777 is equivalent to ---rwxrwxrwx,both formats are accepted",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"tiering": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"policy_name": schema.StringAttribute{
						MarkdownDescription: "The tiering policy that is to be associated with the volume",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"minimum_cooling_days": schema.Int64Attribute{
						MarkdownDescription: "Determines how many days must pass before inactive data in a volume using the Auto or Snapshot-Only policy is considered cold and eligible for tiering",
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"efficiency": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"policy_name": schema.StringAttribute{
						MarkdownDescription: "Allows a storage efficiency policy to be set on volume creation",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"compression": schema.StringAttribute{
						MarkdownDescription: "Whether to enable compression for the volume (HDD and Flash Pool aggregates)",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},

			"snaplock": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "The SnapLock type of the volume",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"analytics": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"state": schema.StringAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Set file system analytics state of the volume",
					},
				},
			},
			// 'autosize' can not use UseStateForUnknown because it will changed by PATCH API on changing size.
			"autosize": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"minimum": schema.Int64Attribute{
						MarkdownDescription: "Minimum size in bytes up to which the volume shrinks automatically. This size cannot be greater than or equal to the maximum size of volume.",
						Optional:            true,
						Computed:            true,
					},
					"maximum": schema.Int64Attribute{
						MarkdownDescription: "Maximum size in bytes up to which a volume grows automatically. This size cannot be less than the current volume size, or less than or equal to the minimum size of volume.",
						Optional:            true,
						Computed:            true,
					},
					"shrink_threshold": schema.Int64Attribute{
						MarkdownDescription: "Used space threshold size, in percentage, for the automatic shrinkage of the volume.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"grow_threshold": schema.Int64Attribute{
						MarkdownDescription: "Used space threshold size, in percentage, for the automatic growth of the volume.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"mode": schema.StringAttribute{
						MarkdownDescription: `
											 Autosize mode for the volume.
											 grow - Volume automatically grows when the amount of used space is above the 'grow_threshold' value.
							   				 grow_shrink - Volume grows or shrinks in response to the amount of space used.
											 off - Autosizing of the volume is disabled.
											 `,
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("off", "grow", "grow_shrink"),
						},
					},
					"size_unit": schema.StringAttribute{
						MarkdownDescription: "The unit used to interpret the minimum or maximum size parameters",
						Optional:            true,
						Computed:            true,
						Validators: []validator.String{
							stringvalidator.OneOf("bytes", "b", "kb", "mb", "gb", "tb", "pb", "eb", "zb", "yb"),
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Volume identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// ModifyPlan makes terraform errors if config or state sets state of the volume offline.
// TO DO: when offline, values change from API response.
func (r *StorageVolumeResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Fill in logic.
	var plan, state, config *StorageVolumeResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if state != nil && !state.State.IsUnknown() && state.State.ValueString() == "offline" {
		resp.Diagnostics.AddError("Volume is offline", "Provider is not supported to manage offline volume. Please manually switch the volume online")
		return
	}
	if plan != nil && !plan.State.IsUnknown() && plan.State.ValueString() == "offline" {

		resp.Diagnostics.AddError("Volume is offline", "Provider is not supported to manage offline volume. Please manually switch the volume online")
		return
	}
}

// Configure adds the provider configured client to the resource.
func (r *StorageVolumeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ConfigValidators validates entire resource configurations
func (d *StorageVolumeResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.RequiredTogether(
			path.MatchRoot("autosize").AtName("minimum"),
			path.MatchRoot("autosize").AtName("size_unit"),
		),
		resourcevalidator.RequiredTogether(
			path.MatchRoot("autosize").AtName("maximum"),
			path.MatchRoot("autosize").AtName("size_unit"),
		),
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *StorageVolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *StorageVolumeResourceModel
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
	// Import don't have id's so we need to get the id from the name
	var response *interfaces.StorageVolumeGetDataModelONTAP
	if data.ID.ValueString() == "" {
		response, err = interfaces.GetStorageVolumeByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
		if err != nil {
			return
		}
		data.ID = types.StringValue(response.UUID)
	} else {
		response, err = interfaces.GetStorageVolume(errorHandler, *client, data.ID.ValueString())
		if err != nil {
			return
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("read a volume resource: %#v", data))

	data.Comment = types.StringValue(response.Comment)
	data.Encrypt = types.BoolValue(response.Encryption.Enabled)
	data.State = types.StringValue(response.State)
	data.Language = types.StringValue(response.Language)
	data.QOSPolicyGroup = types.StringValue(response.QOS.Policy.Name)
	data.SpaceGuarantee = types.StringValue(response.SpaceGuarantee.Type)
	data.SnapshotPolicy = types.StringValue(response.SnapshotPolicy.Name)
	data.Type = types.StringValue(response.Type)

	//Space
	nestedElementTypes := map[string]attr.Type{
		"reporting":   types.BoolType,
		"enforcement": types.BoolType,
	}
	nestedEslements := map[string]attr.Value{
		"reporting":   types.BoolValue(response.Space.LogicalSpace.Reporting),
		"enforcement": types.BoolValue(response.Space.LogicalSpace.Enforcement),
	}
	logicalObjectValue, _ := types.ObjectValue(nestedElementTypes, nestedEslements)
	elementTypes := map[string]attr.Type{
		"size":                   types.Int64Type,
		"size_unit":              types.StringType,
		"percent_snapshot_space": types.Int64Type,
		"logical_space":          types.ObjectType{AttrTypes: nestedElementTypes},
	}

	var sizeUnit string
	var size int64
	// If state does not have size unit, which is the case for importing volume, we pick the proper size unit.
	if data.Space.IsNull() {
		size, sizeUnit = interfaces.ByteFormat(int64(response.Space.Size))
	} else {
		var space StorageVolumeResourceSpace
		diags := data.Space.As(ctx, &space, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		sizeUnit = space.SizeUnit.ValueString()
		size = interfaces.ConvertBytesToUnitInt(int64(response.Space.Size), sizeUnit)
	}

	elements := map[string]attr.Value{
		"size":                   types.Int64Value(size),
		"size_unit":              types.StringValue(sizeUnit),
		"percent_snapshot_space": types.Int64PointerValue(response.Space.Snapshot.ReservePercent),
		"logical_space":          logicalObjectValue,
	}

	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Space = objectValue

	//Snaplock
	elementTypes = map[string]attr.Type{
		"type": types.StringType,
	}
	elements = map[string]attr.Value{
		"type": types.StringValue(response.Snaplock.Type),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.SnapLock = objectValue

	//Efficiency
	elementTypes = map[string]attr.Type{
		"compression": types.StringType,
		"policy_name": types.StringType,
	}
	policyName := response.Efficiency.Policy.Name
	if policyName == "" || policyName == "-" {
		policyName = "default"
	}
	elements = map[string]attr.Value{
		"compression": types.StringValue(response.Efficiency.Compression),
		"policy_name": types.StringValue(policyName),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Efficiency = objectValue

	//Tiering
	elementTypes = map[string]attr.Type{
		"minimum_cooling_days": types.Int64Type,
		"policy_name":          types.StringType,
	}
	elements = map[string]attr.Value{
		"minimum_cooling_days": types.Int64Value(int64(response.TieringPolicy.MinCoolingDays)),
		"policy_name":          types.StringValue(response.TieringPolicy.Policy),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Tiering = objectValue

	//Nas
	elementTypes = map[string]attr.Type{
		"unix_permissions":   types.Int64Type,
		"junction_path":      types.StringType,
		"group_id":           types.Int64Type,
		"user_id":            types.Int64Type,
		"security_style":     types.StringType,
		"export_policy_name": types.StringType,
	}
	elements = map[string]attr.Value{
		"unix_permissions":   types.Int64Value(int64(response.NAS.UnixPermissions)),
		"junction_path":      types.StringValue(response.NAS.JunctionPath),
		"group_id":           types.Int64Value(int64(response.NAS.GroupID)),
		"user_id":            types.Int64Value(int64(response.NAS.UserID)),
		"security_style":     types.StringValue(response.NAS.SecurityStyle),
		"export_policy_name": types.StringValue(response.NAS.ExportPolicy.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Nas = objectValue

	//Analytics
	elementTypes = map[string]attr.Type{
		"state": types.StringType,
	}
	elements = map[string]attr.Value{
		"state": types.StringValue(response.Analytics.State),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Analytics = objectValue

	//Aggregates
	var aggregates []StorageVolumeResourceAggregates
	for _, v := range response.Aggregates {
		var aggregate StorageVolumeResourceAggregates
		aggregate.Name = types.StringValue(v.Name)
		aggregates = append(aggregates, aggregate)
	}
	data.Aggregates = aggregates

	//Autosize
	elementTypes = map[string]attr.Type{
		"minimum":          types.Int64Type,
		"maximum":          types.Int64Type,
		"shrink_threshold": types.Int64Type,
		"grow_threshold":   types.Int64Type,
		"mode":             types.StringType,
		"size_unit":        types.StringType,
	}
	// var sizeUnit is already defined as part of Space model
	var minimum int64
	var maximum int64
	var autoSizeUnit string
	// If state does not have size unit, which is the case for importing volume, we pick the proper size unit.
	if !data.Autosize.IsNull() {
		var autosize StorageVolumeResourceAutosize
		diags := data.Autosize.As(ctx, &autosize, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		autoSizeUnit = autosize.SizeUnit.ValueString()
		minimum = interfaces.ConvertBytesToUnitInt(int64(response.Autosize.Minimum), autoSizeUnit)
		maximum = interfaces.ConvertBytesToUnitInt(int64(response.Autosize.Maximum), autoSizeUnit)
	} else {
		minimum, autoSizeUnit = interfaces.ByteFormat(int64(response.Autosize.Minimum))
		maximum, _ = interfaces.ByteFormat(int64(response.Autosize.Maximum))
	}

	elements = map[string]attr.Value{
		"minimum":          types.Int64Value(minimum),
		"maximum":          types.Int64Value(maximum),
		"shrink_threshold": types.Int64Value(int64(response.Autosize.ShrinkThreshold)),
		"grow_threshold":   types.Int64Value(int64(response.Autosize.GrowThreshold)),
		"mode":             types.StringValue(response.Autosize.Mode),
		"size_unit":        types.StringValue(autoSizeUnit),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Autosize = objectValue

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *StorageVolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageVolumeResourceModel
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var request interfaces.StorageVolumeResourceModel

	//var aggregates = make([]interfaces.Aggregate, len(data.Aggregates))
	//for i, v := range data.Aggregates {
	//	aggregates[i].Name = v.Name.ValueString()
	//}

	aggrgatges := []interfaces.Aggregate{}
	for _, v := range data.Aggregates {
		var aggr interfaces.Aggregate
		aggr.Name = v.Name.ValueString()
		aggrgatges = append(aggrgatges, aggr)
	}
	err := mapstructure.Decode(aggrgatges, &request.Aggregates)
	if err != nil {
		errorHandler.MakeAndReportError("error creating Volume", fmt.Sprintf("error on encoding copies info: %s, copies %#v", err, aggrgatges))
		return
	}

	request.Name = data.Name.ValueString()
	request.SVM.Name = data.SVMName.ValueString()

	if !data.State.IsUnknown() {
		request.State = data.State.ValueString()
	}

	if !data.Type.IsUnknown() {
		request.Type = data.Type.ValueString()
	}
	if !data.SpaceGuarantee.IsUnknown() {
		request.SpaceGuarantee.Type = data.SpaceGuarantee.ValueString()
	}
	if !data.Encrypt.IsUnknown() {
		request.Encryption.Enabled = data.Encrypt.ValueBool()
	}
	if !data.SnapshotPolicy.IsUnknown() {
		request.SnapshotPolicy.Name = data.SnapshotPolicy.ValueString()
	}
	if !data.Language.IsUnknown() {
		request.Language = data.Language.ValueString()
	}
	if !data.QOSPolicyGroup.IsUnknown() {
		request.QOS.Policy.Name = data.QOSPolicyGroup.ValueString()
	}
	if !data.Comment.IsUnknown() {
		request.Comment = data.Comment.ValueString()
	}

	if !data.Nas.IsUnknown() {
		var nas StorageVolumeResourceNas
		diags := data.Nas.As(ctx, &nas, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !nas.ExportPolicy.IsUnknown() {
			request.NAS.ExportPolicy.Name = nas.ExportPolicy.ValueString()
		}
		if !nas.JunctionPath.IsUnknown() {
			request.NAS.JunctionPath = nas.JunctionPath.ValueString()
		}
		if !nas.SecurityStyle.IsUnknown() {
			request.NAS.SecurityStyle = nas.SecurityStyle.ValueString()
		}
		if !nas.UnixPermissions.IsUnknown() {
			request.NAS.UnixPermissions = int(nas.UnixPermissions.ValueInt64())
		}
		if !nas.GroupID.IsUnknown() {
			request.NAS.GroupID = int(nas.GroupID.ValueInt64())
		}
		if !nas.UserID.IsUnknown() {
			request.NAS.UserID = int(nas.UserID.ValueInt64())
		}
	}

	var sizeUnit string
	autoSizeUnit := "mb"
	var space StorageVolumeResourceSpace
	diags := data.Space.As(ctx, &space, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	if _, ok := interfaces.POW2BYTEMAP[space.SizeUnit.ValueString()]; !ok {
		errorHandler.MakeAndReportError("error creating volume", fmt.Sprintf("invalid input for size_unit: %s, required one of: bytes, b, kb, mb, gb, tb, pb, eb, zb, yb", space.SizeUnit.ValueString()))
		return
	}
	sizeUnit = space.SizeUnit.ValueString()
	request.Space.Size = int(space.Size.ValueInt64()) * interfaces.POW2BYTEMAP[space.SizeUnit.ValueString()]

	if !space.PercentSnapshotSpace.IsUnknown() {
		request.Space.Snapshot.ReservePercent = space.PercentSnapshotSpace.ValueInt64Pointer()
	}
	if !space.LogicalSpace.IsUnknown() {
		var logicalSpace StorageVolumeResourceSpaceLogicalSpace
		diags = space.LogicalSpace.As(ctx, &logicalSpace, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !logicalSpace.Enforcement.IsUnknown() {
			request.Space.LogicalSpace.Enforcement = logicalSpace.Enforcement.ValueBool()
		}
		if !logicalSpace.Reporting.IsUnknown() {
			request.Space.LogicalSpace.Reporting = logicalSpace.Reporting.ValueBool()
		}
	}

	if !data.Efficiency.IsUnknown() {
		var efficiency StorageVolumeResourceEfficiency
		diags := data.Efficiency.As(ctx, &efficiency, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if efficiency.Policy.ValueString() == "" || efficiency.Policy.ValueString() == "-" {
			request.Efficiency.Policy.Name = "default"
		} else {
			request.Efficiency.Policy.Name = efficiency.Policy.ValueString()
		}
		if !efficiency.Compression.IsUnknown() {
			request.Efficiency.Compression = efficiency.Compression.ValueString()
		}
	}

	if !data.Tiering.IsUnknown() {
		var tiering StorageVolumeResourceTiering
		diags := data.Tiering.As(ctx, &tiering, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !tiering.Policy.IsUnknown() {
			request.TieringPolicy.Policy = tiering.Policy.ValueString()
		}
		if !tiering.MinimumCoolingDays.IsUnknown() {
			request.TieringPolicy.MinCoolingDays = int(tiering.MinimumCoolingDays.ValueInt64())
		}
	}

	if !data.SnapLock.IsUnknown() {
		var snapLock StorageVolumeResourceSnapLock
		diags := data.SnapLock.As(ctx, &snapLock, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		request.Snaplock.Type = snapLock.SnaplockType.ValueString()
	}

	if !data.Analytics.IsUnknown() {
		var analytics StorageVolumeResourceAnalytics
		diags := data.Analytics.As(ctx, &analytics, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		request.Analytics.State = analytics.State.ValueString()
	}

	if !data.Autosize.IsUnknown() {
		var autosize StorageVolumeResourceAutosize
		diags := data.Autosize.As(ctx, &autosize, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		autoSizeUnit = autosize.SizeUnit.ValueString()

		if !autosize.Minimum.IsUnknown() {
			request.Autosize.Minimum = int(autosize.Minimum.ValueInt64()) * interfaces.POW2BYTEMAP[autosize.SizeUnit.ValueString()]
		}
		if !autosize.Maximum.IsUnknown() {
			request.Autosize.Maximum = int(autosize.Maximum.ValueInt64()) * interfaces.POW2BYTEMAP[autosize.SizeUnit.ValueString()]
		}
		if !autosize.ShrinkThreshold.IsUnknown() {
			request.Autosize.ShrinkThreshold = int(autosize.ShrinkThreshold.ValueInt64())
		}
		if !autosize.GrowThreshold.IsUnknown() {
			request.Autosize.GrowThreshold = int(autosize.GrowThreshold.ValueInt64())
		}
		if !autosize.Mode.IsUnknown() {
			request.Autosize.Mode = autosize.Mode.ValueString()
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	response, err := interfaces.CreateStorageVolume(errorHandler, *client, request)
	if err != nil {
		return
	}

	data.ID = types.StringValue(response.UUID)
	data.Comment = types.StringValue(response.Comment)
	data.Encrypt = types.BoolValue(response.Encryption.Enabled)
	data.State = types.StringValue(response.State)
	data.Language = types.StringValue(response.Language)
	data.QOSPolicyGroup = types.StringValue(response.QOS.Policy.Name)
	data.SpaceGuarantee = types.StringValue(response.SpaceGuarantee.Type)
	data.SnapshotPolicy = types.StringValue(response.SnapshotPolicy.Name)
	data.Type = types.StringValue(response.Type)

	//Space
	nestedElementTypes := map[string]attr.Type{
		"reporting":   types.BoolType,
		"enforcement": types.BoolType,
	}
	nestedEslements := map[string]attr.Value{
		"reporting":   types.BoolValue(response.Space.LogicalSpace.Reporting),
		"enforcement": types.BoolValue(response.Space.LogicalSpace.Enforcement),
	}
	logicalObjectValue, _ := types.ObjectValue(nestedElementTypes, nestedEslements)

	elementTypes := map[string]attr.Type{
		"size":                   types.Int64Type,
		"size_unit":              types.StringType,
		"percent_snapshot_space": types.Int64Type,
		"logical_space":          types.ObjectType{AttrTypes: nestedElementTypes},
	}
	elements := map[string]attr.Value{
		"size":                   types.Int64Value(int64(response.Space.Size / interfaces.POW2BYTEMAP[sizeUnit])),
		"size_unit":              types.StringValue(sizeUnit),
		"percent_snapshot_space": types.Int64PointerValue(response.Space.Snapshot.ReservePercent),
		"logical_space":          logicalObjectValue,
	}

	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Space = objectValue

	//Snaplock
	elementTypes = map[string]attr.Type{
		"type": types.StringType,
	}
	elements = map[string]attr.Value{
		"type": types.StringValue(response.Snaplock.Type),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.SnapLock = objectValue

	//Efficiency
	elementTypes = map[string]attr.Type{
		"compression": types.StringType,
		"policy_name": types.StringType,
	}
	policyName := response.Efficiency.Policy.Name
	if policyName == "" || policyName == "-" {
		policyName = "default"
	}
	elements = map[string]attr.Value{
		"compression": types.StringValue(response.Efficiency.Compression),
		"policy_name": types.StringValue(policyName),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Efficiency = objectValue

	//Tiering
	elementTypes = map[string]attr.Type{
		"minimum_cooling_days": types.Int64Type,
		"policy_name":          types.StringType,
	}
	elements = map[string]attr.Value{
		"minimum_cooling_days": types.Int64Value(int64(response.TieringPolicy.MinCoolingDays)),
		"policy_name":          types.StringValue(response.TieringPolicy.Policy),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Tiering = objectValue

	//Nas
	elementTypes = map[string]attr.Type{
		"unix_permissions":   types.Int64Type,
		"junction_path":      types.StringType,
		"group_id":           types.Int64Type,
		"user_id":            types.Int64Type,
		"security_style":     types.StringType,
		"export_policy_name": types.StringType,
	}
	elements = map[string]attr.Value{
		"unix_permissions":   types.Int64Value(int64(response.NAS.UnixPermissions)),
		"junction_path":      types.StringValue(response.NAS.JunctionPath),
		"group_id":           types.Int64Value(int64(response.NAS.GroupID)),
		"user_id":            types.Int64Value(int64(response.NAS.UserID)),
		"security_style":     types.StringValue(response.NAS.SecurityStyle),
		"export_policy_name": types.StringValue(response.NAS.ExportPolicy.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Nas = objectValue

	//Analytics
	elementTypes = map[string]attr.Type{
		"state": types.StringType,
	}
	elements = map[string]attr.Value{
		"state": types.StringValue(response.Analytics.State),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Analytics = objectValue

	//Autosize
	elementTypes = map[string]attr.Type{
		"minimum":          types.Int64Type,
		"maximum":          types.Int64Type,
		"shrink_threshold": types.Int64Type,
		"grow_threshold":   types.Int64Type,
		"mode":             types.StringType,
		"size_unit":        types.StringType,
	}
	elements = map[string]attr.Value{
		"minimum":          types.Int64Value(int64(response.Autosize.Minimum / interfaces.POW2BYTEMAP[autoSizeUnit])),
		"maximum":          types.Int64Value(int64(response.Autosize.Maximum / interfaces.POW2BYTEMAP[autoSizeUnit])),
		"shrink_threshold": types.Int64Value(int64(response.Autosize.ShrinkThreshold)),
		"grow_threshold":   types.Int64Value(int64(response.Autosize.GrowThreshold)),
		"mode":             types.StringValue(response.Autosize.Mode),
		"size_unit":        types.StringValue(autoSizeUnit),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Autosize = objectValue

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

// Update updates the resource and sets the updated Terraform state on success.
func (r *StorageVolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *StorageVolumeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	var request interfaces.StorageVolumeResourceModel

	if !plan.Type.Equal(state.Type) {
		request.State = plan.State.ValueString()
	}

	if !plan.Type.Equal(state.Type) {
		request.Type = plan.Type.ValueString()
	}

	if !plan.SnapshotPolicy.Equal(state.SnapshotPolicy) {
		request.SnapshotPolicy.Name = plan.SnapshotPolicy.ValueString()
	}

	if !plan.Language.Equal(state.Language) {
		request.Language = plan.Language.ValueString()
	}

	if !plan.QOSPolicyGroup.Equal(state.QOSPolicyGroup) {
		request.QOS.Policy.Name = plan.QOSPolicyGroup.ValueString()
	}

	if !plan.Comment.Equal(state.Comment) {
		request.Comment = plan.Comment.ValueString()
	}

	if !plan.SpaceGuarantee.Equal(state.SpaceGuarantee) {
		request.SpaceGuarantee.Type = plan.SpaceGuarantee.ValueString()
	}

	// encryption is tricky in PATCH. The meaning of this option more like the action rather than the end state.
	// if it is set to false, it means that we want to disable encryption, if it is set to true, it means that we want to enable encryption.
	// Regardless of the current state. And the API returns errors depend on the cuurent state and the end action.
	// so if there is no change in the encryption state, we don't send it to the API.
	// But omitempty is annoying because it ignores the value if it is false.
	// Use a flag 'ignoreEncrption' to determine if we should send the encryption state to the API.
	ignoreEncrption := true
	if !plan.Encrypt.IsUnknown() {
		if !plan.Encrypt.Equal(state.Encrypt) {
			ignoreEncrption = false
			request.Encryption.Enabled = plan.Encrypt.ValueBool()
		}
	}

	if !plan.Nas.Equal(state.Nas) {
		var nas StorageVolumeResourceNas
		diags := plan.Nas.As(ctx, &nas, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		request.NAS.ExportPolicy.Name = nas.ExportPolicy.ValueString()
		request.NAS.JunctionPath = nas.JunctionPath.ValueString()
		request.NAS.SecurityStyle = nas.SecurityStyle.ValueString()
		request.NAS.UnixPermissions = int(nas.UnixPermissions.ValueInt64())
		request.NAS.GroupID = int(nas.GroupID.ValueInt64())
		request.NAS.UserID = int(nas.UserID.ValueInt64())

	}

	ignoreLogicalSpace := true
	var space StorageVolumeResourceSpace
	diags := plan.Space.As(ctx, &space, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	if _, ok := interfaces.POW2BYTEMAP[space.SizeUnit.ValueString()]; !ok {
		errorHandler.MakeAndReportError("error updating volume", fmt.Sprintf("invalid input for size_unit: %s, required one of: bytes, b, kb, mb, gb, tb, pb, eb, zb, yb", space.SizeUnit.ValueString()))
		return
	}
	if !plan.Space.Equal(state.Space) {
		ignoreLogicalSpace = false
		request.Space.Size = int(space.Size.ValueInt64()) * interfaces.POW2BYTEMAP[space.SizeUnit.ValueString()]
		request.Space.Snapshot.ReservePercent = space.PercentSnapshotSpace.ValueInt64Pointer()

		var logicalSpace StorageVolumeResourceSpaceLogicalSpace
		space.LogicalSpace.As(ctx, &logicalSpace, basetypes.ObjectAsOptions{})
		request.Space.LogicalSpace.Enforcement = logicalSpace.Enforcement.ValueBool()
		request.Space.LogicalSpace.Reporting = logicalSpace.Reporting.ValueBool()

	}

	if !plan.Efficiency.Equal(state.Efficiency) {
		var efficiency StorageVolumeResourceEfficiency
		diags := plan.Efficiency.As(ctx, &efficiency, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if efficiency.Policy.ValueString() == "" || efficiency.Policy.ValueString() == "-" {
			request.Efficiency.Policy.Name = "default"
		} else {
			request.Efficiency.Policy.Name = efficiency.Policy.ValueString()
		}
		if !efficiency.Compression.IsUnknown() {
			request.Efficiency.Compression = efficiency.Compression.ValueString()
		}
	}

	if !plan.Tiering.Equal(state.Tiering) {
		var tiering StorageVolumeResourceTiering
		diags := plan.Tiering.As(ctx, &tiering, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		request.TieringPolicy.Policy = tiering.Policy.ValueString()
		request.TieringPolicy.MinCoolingDays = int(tiering.MinimumCoolingDays.ValueInt64())
	}

	if !plan.SnapLock.Equal(state.SnapLock) {
		var snapLock StorageVolumeResourceSnapLock
		diags := plan.SnapLock.As(ctx, &snapLock, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		request.Snaplock.Type = snapLock.SnaplockType.ValueString()
	}

	if !plan.Analytics.Equal(state.Analytics) {
		var analytics StorageVolumeResourceAnalytics
		diags := plan.Analytics.As(ctx, &analytics, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		request.Analytics.State = analytics.State.ValueString()
	}

	if !plan.Autosize.IsUnknown() {
		// if size changes, autosize might change as well from the API side. When autosize is not unknown, we should set it in the PATCH request.
		var autosize StorageVolumeResourceAutosize
		diags := plan.Autosize.As(ctx, &autosize, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		request.Autosize.Minimum = int(autosize.Minimum.ValueInt64()) * interfaces.POW2BYTEMAP[autosize.SizeUnit.ValueString()]
		request.Autosize.Maximum = int(autosize.Maximum.ValueInt64()) * interfaces.POW2BYTEMAP[autosize.SizeUnit.ValueString()]
		request.Autosize.ShrinkThreshold = int(autosize.ShrinkThreshold.ValueInt64())
		request.Autosize.GrowThreshold = int(autosize.GrowThreshold.ValueInt64())
		request.Autosize.Mode = autosize.Mode.ValueString()
	}

	err = interfaces.UpddateStorageVolume(errorHandler, *client, request, plan.ID.ValueString(), ignoreEncrption, ignoreLogicalSpace)
	if err != nil {
		return
	}
	// Save updated data into Terraform state
	planEncryption := plan.Encrypt.ValueBool()
	readDiags := readVolume(ctx, client, plan)
	resp.Diagnostics.Append(readDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	timeout := 10
	// Wait for the encryption state to be updated if it was changed.
	for !ignoreEncrption && plan.Encrypt.ValueBool() != planEncryption {
		time.Sleep(time.Second * 30)
		readDiags := readVolume(ctx, client, plan)
		resp.Diagnostics.Append(readDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		timeout--
		if timeout <= 0 {
			break
		}
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *StorageVolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *StorageVolumeResourceModel

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

	if data.ID.IsUnknown() {
		errorHandler.MakeAndReportError("UUID is null", "Volume UUID is null")
		return
	}

	err = interfaces.DeleteStorageVolume(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *StorageVolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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

func readVolume(ctx context.Context, client *restclient.RestClient, data *StorageVolumeResourceModel) diag.Diagnostics {
	var allDiags diag.Diagnostics

	errorHandler := utils.NewErrorHandler(ctx, &allDiags)

	response, returnedError := interfaces.GetStorageVolume(errorHandler, *client, data.ID.ValueString())
	if returnedError != nil {
		allDiags.AddError("Error reading volume", returnedError.Error())
		return allDiags
	}
	data.Comment = types.StringValue(response.Comment)
	data.Encrypt = types.BoolValue(response.Encryption.Enabled)
	data.State = types.StringValue(response.State)
	data.Language = types.StringValue(response.Language)
	data.QOSPolicyGroup = types.StringValue(response.QOS.Policy.Name)
	data.SpaceGuarantee = types.StringValue(response.SpaceGuarantee.Type)
	data.SnapshotPolicy = types.StringValue(response.SnapshotPolicy.Name)
	data.Type = types.StringValue(response.Type)

	//Space
	nestedElementTypes := map[string]attr.Type{
		"reporting":   types.BoolType,
		"enforcement": types.BoolType,
	}
	nestedEslements := map[string]attr.Value{
		"reporting":   types.BoolValue(response.Space.LogicalSpace.Reporting),
		"enforcement": types.BoolValue(response.Space.LogicalSpace.Enforcement),
	}
	logicalObjectValue, _ := types.ObjectValue(nestedElementTypes, nestedEslements)
	elementTypes := map[string]attr.Type{
		"size":                   types.Int64Type,
		"size_unit":              types.StringType,
		"percent_snapshot_space": types.Int64Type,
		"logical_space":          types.ObjectType{AttrTypes: nestedElementTypes},
	}
	var sizeUnit string
	var space StorageVolumeResourceSpace
	diags := data.Space.As(ctx, &space, basetypes.ObjectAsOptions{})
	if diags.HasError() {
		allDiags.Append(diags...)
		return allDiags
	}
	if _, ok := interfaces.POW2BYTEMAP[space.SizeUnit.ValueString()]; !ok {
		errorHandler.MakeAndReportError("error creating volume", fmt.Sprintf("invalid input for size_unit: %s, required one of: bytes, b, kb, mb, gb, tb, pb, eb, zb, yb", space.SizeUnit.ValueString()))
		return allDiags
	}
	sizeUnit = space.SizeUnit.ValueString()

	elements := map[string]attr.Value{
		"size":                   types.Int64Value(int64(response.Space.Size / interfaces.POW2BYTEMAP[sizeUnit])),
		"size_unit":              types.StringValue(sizeUnit),
		"percent_snapshot_space": types.Int64PointerValue(response.Space.Snapshot.ReservePercent),
		"logical_space":          logicalObjectValue,
	}

	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		allDiags.Append(diags...)
	}
	data.Space = objectValue

	//Snaplock
	elementTypes = map[string]attr.Type{
		"type": types.StringType,
	}
	elements = map[string]attr.Value{
		"type": types.StringValue(response.Snaplock.Type),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		allDiags.Append(diags...)
	}
	data.SnapLock = objectValue

	//Efficiency
	elementTypes = map[string]attr.Type{
		"compression": types.StringType,
		"policy_name": types.StringType,
	}
	policyName := response.Efficiency.Policy.Name
	if policyName == "" || policyName == "-" {
		policyName = "default"
	}
	elements = map[string]attr.Value{
		"compression": types.StringValue(response.Efficiency.Compression),
		"policy_name": types.StringValue(policyName),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		allDiags.Append(diags...)
	}
	data.Efficiency = objectValue

	//Tiering
	elementTypes = map[string]attr.Type{
		"minimum_cooling_days": types.Int64Type,
		"policy_name":          types.StringType,
	}
	elements = map[string]attr.Value{
		"minimum_cooling_days": types.Int64Value(int64(response.TieringPolicy.MinCoolingDays)),
		"policy_name":          types.StringValue(response.TieringPolicy.Policy),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		allDiags.Append(diags...)
	}
	data.Tiering = objectValue

	//Nas
	elementTypes = map[string]attr.Type{
		"unix_permissions":   types.Int64Type,
		"junction_path":      types.StringType,
		"group_id":           types.Int64Type,
		"user_id":            types.Int64Type,
		"security_style":     types.StringType,
		"export_policy_name": types.StringType,
	}
	elements = map[string]attr.Value{
		"unix_permissions":   types.Int64Value(int64(response.NAS.UnixPermissions)),
		"junction_path":      types.StringValue(response.NAS.JunctionPath),
		"group_id":           types.Int64Value(int64(response.NAS.GroupID)),
		"user_id":            types.Int64Value(int64(response.NAS.UserID)),
		"security_style":     types.StringValue(response.NAS.SecurityStyle),
		"export_policy_name": types.StringValue(response.NAS.ExportPolicy.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		allDiags.Append(diags...)
	}
	data.Nas = objectValue

	//Analytics
	elementTypes = map[string]attr.Type{
		"state": types.StringType,
	}
	elements = map[string]attr.Value{
		"state": types.StringValue(response.Analytics.State),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		allDiags.Append(diags...)
	}
	data.Analytics = objectValue

	//Autosize
	elementTypes = map[string]attr.Type{
		"minimum":          types.Int64Type,
		"maximum":          types.Int64Type,
		"shrink_threshold": types.Int64Type,
		"grow_threshold":   types.Int64Type,
		"mode":             types.StringType,
		"size_unit":        types.StringType,
	}
	// var sizeUnit is already defined as part of Space model
	if !data.Autosize.IsUnknown() {
		var autosize StorageVolumeResourceAutosize
		diags = data.Autosize.As(ctx, &autosize, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			allDiags.Append(diags...)
			return allDiags
		}
		sizeUnit = autosize.SizeUnit.ValueString()
	}

	elements = map[string]attr.Value{
		"minimum":          types.Int64Value(int64(response.Autosize.Minimum / interfaces.POW2BYTEMAP[sizeUnit])),
		"maximum":          types.Int64Value(int64(response.Autosize.Maximum / interfaces.POW2BYTEMAP[sizeUnit])),
		"shrink_threshold": types.Int64Value(int64(response.Autosize.ShrinkThreshold)),
		"grow_threshold":   types.Int64Value(int64(response.Autosize.GrowThreshold)),
		"mode":             types.StringValue(response.Autosize.Mode),
		"size_unit":        types.StringValue(sizeUnit),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		allDiags.Append(diags...)
	}
	data.Autosize = objectValue

	return allDiags
}
