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
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &StorageVolumeResource{}
var _ resource.ResourceWithImportState = &StorageVolumeResource{}

// NewStorageVolumeResource is a helper function to simplify the provider implementation.
func NewStorageVolumeResource() resource.Resource {
	return &StorageVolumeResource{
		config: resourceOrDataSourceConfig{
			name: "storage_volume_resource",
		},
	}
}

// StorageVolumeResource defines the resource implementation.
type StorageVolumeResource struct {
	config resourceOrDataSourceConfig
}

// StorageVolumeResourceModel describes the resource data model.
type StorageVolumeResourceModel struct {
	CxProfileName           types.String   `tfsdk:"cx_profile_name"`
	Name                    types.String   `tfsdk:"name"`
	SVMName                 types.String   `tfsdk:"svm_name"`
	Size                    types.Int64    `tfsdk:"size"`
	SizeUnit                types.String   `tfsdk:"size_unit"`
	IsOnline                types.Bool     `tfsdk:"is_online"`
	Type                    types.String   `tfsdk:"type"`
	ExportPolicy            types.String   `tfsdk:"export_policy"`
	JunctionPath            types.String   `tfsdk:"junction_path"`
	SpaceGuarantee          types.String   `tfsdk:"space_guarantee"`
	PercentSnapshotSpace    types.Int64    `tfsdk:"percent_snapshot_space"`
	SecurityStyle           types.String   `tfsdk:"security_style"`
	Encrypt                 types.Bool     `tfsdk:"encrypt"`
	EfficiencyPolicy        types.String   `tfsdk:"efficiency_policy"`
	UnixPermissions         types.String   `tfsdk:"unix_permissions"`
	GroupID                 types.Int64    `tfsdk:"group_id"`
	UserID                  types.Int64    `tfsdk:"user_id"`
	SnapshotPolicy          types.String   `tfsdk:"snapshot_policy"`
	Language                types.String   `tfsdk:"language"`
	QOSPolicyGroup          types.String   `tfsdk:"qos_policy_group"`
	QOSAdaptivePolicyGroup  types.String   `tfsdk:"qos_adaptive_policy_group"`
	TieringPolicy           types.String   `tfsdk:"tiering_policy"`
	Comment                 types.String   `tfsdk:"comment"`
	Compression             types.Bool     `tfsdk:"compression"`
	InlineCompression       types.Bool     `tfsdk:"inline_compression"`
	TieringMinCoolingDays   types.Int64    `tfsdk:"tiering_minimum_cooling_days"`
	LogicalSpaceEnforcement types.Bool     `tfsdk:"logical_space_enforcement"`
	LogicalSpaceReporting   types.Bool     `tfsdk:"logical_space_reporting"`
	SnaplockType            types.String   `tfsdk:"snaplock_type"`
	Analytics               types.String   `tfsdk:"analytics"`
	Aggregates              []types.String `tfsdk:"aggregates"`
	UUID                    types.String   `tfsdk:"uuid"`
	ID                      types.String   `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *StorageVolumeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
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
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the volume",
				Required:            true,
			},
			"size_unit": schema.StringAttribute{
				MarkdownDescription: "The unit used to interpret the size parameter",
				Required:            true,
			},
			"is_online": schema.BoolAttribute{
				MarkdownDescription: "Whether the specified volume is online, or not",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The volume type, either read-write (RW) or data-protection (DP)",
				Optional:            true,
			},
			"export_policy": schema.StringAttribute{
				MarkdownDescription: "The name of the export policy",
				Optional:            true,
			},
			"junction_path": schema.StringAttribute{
				MarkdownDescription: "Junction path of the volume",
				Optional:            true,
			},
			"space_guarantee": schema.StringAttribute{
				MarkdownDescription: "Space guarantee style for the volume",
				Optional:            true,
			},
			"percent_snapshot_space": schema.Int64Attribute{
				MarkdownDescription: "Amount of space reserved for snapshot copies of the volume",
				Optional:            true,
			},
			"security_style": schema.StringAttribute{
				MarkdownDescription: "The security style associated to the volume",
				Optional:            true,
			},
			"encrypt": schema.BoolAttribute{
				MarkdownDescription: "Whether or not to enable Volume Encryption",
				Optional:            true,
			},
			"efficiency_policy": schema.StringAttribute{
				MarkdownDescription: "Allows a storage efficiency policy to be set on volume creation",
				Optional:            true,
			},
			"unix_permissions": schema.StringAttribute{
				MarkdownDescription: "Unix permission bits in octal or symbolic format. For example, 0 is equivalent to ------------, 777 is equivalent to ---rwxrwxrwx,both formats are accepted",
				Optional:            true,
			},
			"group_id": schema.Int64Attribute{
				MarkdownDescription: "The UNIX group ID for the volume",
				Optional:            true,
			},
			"user_id": schema.Int64Attribute{
				MarkdownDescription: "The UNIX user ID for the volume",
				Optional:            true,
			},
			"snapshot_policy": schema.StringAttribute{
				MarkdownDescription: "The name of the snapshot policy",
				Optional:            true,
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "Language to use for volume",
				Optional:            true,
			},
			"qos_policy_group": schema.StringAttribute{
				MarkdownDescription: "Specifies a QoS policy group to be set on volume",
				Optional:            true,
			},
			"qos_adaptive_policy_group": schema.StringAttribute{
				MarkdownDescription: "Specifies a QoS adaptive policy group to be set on volume",
				Optional:            true,
			},
			"tiering_policy": schema.StringAttribute{
				MarkdownDescription: "The tiering policy that is to be associated with the volume",
				Optional:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Sets a comment associated with the volume",
				Optional:            true,
			},
			"compression": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable compression for the volume (HDD and Flash Pool aggregates)",
				Optional:            true,
			},
			"inline_compression": schema.BoolAttribute{
				MarkdownDescription: "Whether to enable inline compression for the volume (HDD and Flash Pool aggregates, AFF platforms)",
				Optional:            true,
			},
			"tiering_minimum_cooling_days": schema.Int64Attribute{
				MarkdownDescription: "Determines how many days must pass before inactive data in a volume using the Auto or Snapshot-Only policy is considered cold and eligible for tiering",
				Optional:            true,
			},
			"logical_space_enforcement": schema.BoolAttribute{
				MarkdownDescription: "Whether to perform logical space accounting on the volume",
				Optional:            true,
			},
			"logical_space_reporting": schema.BoolAttribute{
				MarkdownDescription: "Whether to report space logically",
				Optional:            true,
			},
			"aggregates": schema.ListAttribute{
				ElementType:         types.StringType,
				Required:            true,
				MarkdownDescription: "List of aggregates in which to create the volume",
			},
			"snaplock_type": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "The SnapLock type of the volume",
			},
			"analytics": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Set file system analytics state of the volume",
			},
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Volume identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *StorageVolumeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
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
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.UUID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "Volume UUID is null")
		return
	}
	data.ID = types.StringValue("example-id")

	_, err = interfaces.GetStorageVolume(errorHandler, *client, data.UUID.ValueString())
	if err != nil {
		return
	}

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

	aggregates := []interfaces.Aggregate{}
	for _, v := range data.Aggregates {
		aggr := interfaces.Aggregate{}
		aggr.Name = v.ValueString()
		aggregates = append(aggregates, aggr)
	}
	err := mapstructure.Decode(aggregates, &request.Aggregates)
	if err != nil {
		errorHandler.MakeAndReportError("error creating volume", fmt.Sprintf("error on encoding aggregates info: %s, aggregates %#v", err, aggregates))
		return
	}
	request.Name = data.Name.ValueString()
	request.SVM.Name = data.SVMName.ValueString()

	if _, ok := interfaces.POW2BYTEMAP[data.SizeUnit.ValueString()]; !ok {
		errorHandler.MakeAndReportError("error creating volume", fmt.Sprintf("invalid input for size_unit: %s, required one of: bytes, b, kb, mb, gb, tb, pb, eb, zb, yb", data.SizeUnit.ValueString()))
		return
	}
	request.Space.Size = int(data.Size.ValueInt64()) * interfaces.POW2BYTEMAP[data.SizeUnit.ValueString()]

	if !data.IsOnline.IsNull() {
		request.State = interfaces.BoolToOnline(data.IsOnline.ValueBool())
	}

	if !data.Type.IsNull() {
		request.Type = data.Type.ValueString()
	}

	if !data.ExportPolicy.IsNull() {
		request.NAS.ExportPolicy.Name = data.ExportPolicy.ValueString()
	}

	if !data.JunctionPath.IsNull() {
		request.NAS.JunctionPath = data.JunctionPath.ValueString()
	}

	if !data.SpaceGuarantee.IsNull() {
		request.SpaceGuarantee.Type = data.SpaceGuarantee.ValueString()
	}

	if !data.PercentSnapshotSpace.IsNull() {
		request.Space.Snapshot.ReservePercent = int(data.PercentSnapshotSpace.ValueInt64())
	}

	if !data.SecurityStyle.IsNull() {
		request.NAS.SecurityStyle = data.SecurityStyle.ValueString()
	}

	if !data.Encrypt.IsNull() {
		request.Encryption.Enabled = data.Encrypt.ValueBool()
	}

	if !data.EfficiencyPolicy.IsNull() {
		request.Efficiency.Policy.Name = data.EfficiencyPolicy.ValueString()
	}

	if !data.UnixPermissions.IsNull() {
		request.NAS.UnixPermissions = data.UnixPermissions.ValueString()
	}

	if !data.GroupID.IsNull() {
		request.NAS.GroupID = int(data.GroupID.ValueInt64())
	}

	if !data.UserID.IsNull() {
		request.NAS.UserID = int(data.UserID.ValueInt64())
	}

	if !data.SnapshotPolicy.IsNull() {
		request.SnapshotPolicy.Name = data.SnapshotPolicy.ValueString()
	}

	if !data.Language.IsNull() {
		request.Language = data.Language.ValueString()
	}

	if !data.QOSPolicyGroup.IsNull() && !data.QOSAdaptivePolicyGroup.IsNull() {
		errorHandler.MakeAndReportError("error creating volume",
			fmt.Sprintf("with Rest API qos_policy_group and qos_adaptive_policy_group are now the same thing and cannot be set at the same time"))
		return
	}

	if !data.QOSPolicyGroup.IsNull() {
		request.QOS.Policy.Name = data.QOSPolicyGroup.ValueString()
	}

	if !data.QOSAdaptivePolicyGroup.IsNull() {
		request.QOS.Policy.Name = data.QOSAdaptivePolicyGroup.ValueString()
	}

	if !data.TieringPolicy.IsNull() {
		request.TieringPolicy.Policy = data.TieringPolicy.ValueString()
	}

	if !data.Comment.IsNull() {
		request.Comment = data.Comment.ValueString()
	}

	if !data.Compression.IsNull() || !data.InlineCompression.IsNull() {
		request.Efficiency.Compression = interfaces.GetCompression(data.Compression.ValueBool(), data.InlineCompression.ValueBool())
	}

	if !data.TieringMinCoolingDays.IsNull() {
		request.TieringPolicy.MinCoolingDays = int(data.TieringMinCoolingDays.ValueInt64())
	}

	if !data.LogicalSpaceEnforcement.IsNull() {
		request.Space.LogicalSpace.Enforcement = data.LogicalSpaceEnforcement.ValueBool()
	}

	if !data.LogicalSpaceReporting.IsNull() {
		request.Space.LogicalSpace.Reporting = data.LogicalSpaceReporting.ValueBool()
	}

	if !data.SnaplockType.IsNull() {
		request.Snaplock.Type = data.SnaplockType.ValueString()
	}

	if !data.Analytics.IsNull() {
		request.Analytics.State = data.Analytics.ValueString()
	}

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	volume, err := interfaces.CreateStorageVolume(errorHandler, *client, request)
	if err != nil {
		return
	}

	data.UUID = types.StringValue(volume.UUID)
	data.ID = types.StringValue("example-id")

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *StorageVolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *StorageVolumeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
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
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.UUID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "Volume UUID is null")
		return
	}

	err = interfaces.DeleteStorageVolume(errorHandler, *client, data.UUID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *StorageVolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
