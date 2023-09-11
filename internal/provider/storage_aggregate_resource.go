package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &AggregateResource{}

var _ resource.ResourceWithImportState = &AggregateResource{}

// NewAggregateResource is a helper function to simplify the provider implementation.
func NewAggregateResource() resource.Resource {
	return &AggregateResource{
		config: resourceOrDataSourceConfig{
			name: "storage_aggregate_resource",
		},
	}
}

// AggregateResource defines the resource implementation.
type AggregateResource struct {
	config resourceOrDataSourceConfig
}

// AggregateResourceModel describes the resource data model.
type AggregateResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	ID            types.String `tfsdk:"id"`
	State         types.String `tfsdk:"state"`
	Node          types.String `tfsdk:"node"`
	DiskClass     types.String `tfsdk:"disk_class"`
	DiskCount     types.Int64  `tfsdk:"disk_count"`
	DiskSize      types.Int64  `tfsdk:"disk_size"`
	DiskSizeUnit  types.String `tfsdk:"disk_size_unit"`
	RaidSize      types.Int64  `tfsdk:"raid_size"`
	RaidType      types.String `tfsdk:"raid_type"`
	IsMirrored    types.Bool   `tfsdk:"is_mirrored"`
	SnaplockType  types.String `tfsdk:"snaplock_type"`
	Encryption    types.Bool   `tfsdk:"encryption"`
}

// Metadata returns the resource type name.
func (r *AggregateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *AggregateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Export policy resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the aggregate to manage",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Aggregate identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				MarkdownDescription: "Whether the specified aggregate should be enabled or disabled.",
				Validators: []validator.String{
					stringvalidator.OneOf("online", "offline"),
				},
			},
			"node": schema.StringAttribute{
				// required for REST
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Node for the aggregate to be created on. If no node specified, mgmt lif home will be used. If disk_count is present, node name is required.",
			},
			"disk_class": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Class of disk to use to build aggregate. capacity_flash is listed in swagger, but rejected as invalid by ONTAP.",
				Validators: []validator.String{
					stringvalidator.OneOfCaseInsensitive("capacity", "performance", "archive", "solid_state", "array", "virtual", "data_center", "capacity_flash"),
				},
			},
			"disk_count": schema.Int64Attribute{
				// required for REST
				Required: true,
				MarkdownDescription: `Number of disks to place into the aggregate, including parity disks.
				The disks in this newly-created aggregate come from the spare disk pool.
				The smallest disks in this pool join the aggregate first, unless the disk_size argument is provided.
				Modifiable only if specified disk_count is larger than current disk_count.
				If the disk_count % raid_size == 1, only disk_count/raid_size * raid_size will be added.
				If disk_count is 6, raid_type is raid4, raid_size 4, all 6 disks will be added.
				If disk_count is 5, raid_type is raid4, raid_size 4, 5/4 * 4 = 4 will be added. 1 will not be added.
				`,
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.Expressions{
						path.MatchRoot("node")}...),
				},
			},
			"disk_size": schema.Int64Attribute{
				Optional:            true,
				MarkdownDescription: "Disk size to use in 4K block size.  Disks within 10 precent of specified size will be used.",
				Validators: []validator.Int64{
					int64validator.AlsoRequires(path.Expressions{
						path.MatchRoot("disk_size_unit")}...),
				},
			},
			"disk_size_unit": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: `Disk size to use in the specified unit. This is converted to bytes, assuming K=1024.`,
				Validators: []validator.String{
					stringvalidator.AlsoRequires(path.Expressions{
						path.MatchRoot("disk_size")}...),
				},
			},
			"raid_size": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
				MarkdownDescription: "Sets the maximum number of drives per raid group.",
			},
			"raid_type": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
				Validators: []validator.String{
					stringvalidator.OneOf("raid4", "raid_dp", "raid_tec", "raid_0"),
				},
			},
			"is_mirrored": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: `Specifies that the new aggregate be mirrored (have two plexes).
				If set to true, then the indicated disks will be split across the two plexes. By default, the new aggregate will not be mirrored.`,
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown(), boolplanmodifier.RequiresReplace()},
			},
			// TODO: options in ansible, but it uses different REST API endpoint.
			// 'storage/aggregates/%s/cloud-stores' % self.uuid
			// "object_store_name": schema.StringAttribute{
			// 	Optional: true,
			// },
			// same above
			// "allow_flexgroups": schema.BoolAttribute{
			// 	Optional: true,
			// },
			"snaplock_type": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Type of snaplock for the aggregate being created.",
				Validators: []validator.String{
					stringvalidator.OneOf("compliance", "enterprise", "non_snaplock"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"encryption": schema.BoolAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Whether to enable software encryption. This is equivalent to -encrypt-with-aggr-key when using the CLI.Requires a VE license.",
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *AggregateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *AggregateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *AggregateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var request interfaces.StorageAggregateResourceModel
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	request.Name = data.Name.ValueString()

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	var diskSize int
	if !data.DiskSizeUnit.IsNull() && !data.DiskSize.IsNull() {
		diskSize = int(data.DiskSize.ValueInt64()) * interfaces.POW2BYTEMAP[data.DiskSizeUnit.ValueString()]
	}
	if !data.SnaplockType.IsNull() {
		request.SnaplockType = data.SnaplockType.ValueString()
	}
	if !data.Node.IsNull() {
		request.Node = map[string]string{
			"name": data.Node.ValueString(),
		}
	}
	if !data.Encryption.IsNull() {
		request.DataEncryption = map[string]bool{}
		request.DataEncryption["software_encryption_enabled"] = data.Encryption.ValueBool()
	}
	var BlockStoragePrimary interfaces.AggregateBlockStoragePrimary
	if !data.DiskClass.IsNull() {
		BlockStoragePrimary.DiskClass = data.DiskClass.ValueString()
	}
	if !data.DiskCount.IsNull() {
		BlockStoragePrimary.DiskCount = data.DiskCount.ValueInt64()
	}
	if !data.RaidSize.IsNull() {
		BlockStoragePrimary.RaidSize = data.RaidSize.ValueInt64()
	}
	if !data.RaidType.IsNull() {
		BlockStoragePrimary.RaidType = data.RaidType.ValueString()
	}
	if (BlockStoragePrimary != interfaces.AggregateBlockStoragePrimary{}) {
		request.BlockStorage = map[string]interface{}{}
		var body map[string]interface{}
		if err := mapstructure.Decode(BlockStoragePrimary, &body); err != nil {
			return
		}
		request.BlockStorage["primary"] = body
	}
	if !data.IsMirrored.IsNull() {
		if request.BlockStorage == nil {
			request.BlockStorage = map[string]interface{}{}
		}
		request.BlockStorage["mirror"] = map[string]bool{
			"enabled": data.IsMirrored.ValueBool(),
		}
	}
	// state parameter only can be updated
	// if set the state at the creation stage is illegal
	if data.State.ValueString() != "" {
		errorHandler.MakeAndReportError("set state is not allowed on creation", "error on setting state during resource creation")
		return
	}
	aggregate, err := interfaces.CreateStorageAggregate(errorHandler, *client, request, diskSize)
	if err != nil {
		return
	}

	// ONTAP will return the aggregate state as "onlining" when it is being created, Encryption is not enabled until the aggregate is online.
	// So we need to wait until the aggregate is online.
	waitTime := 1
	for aggregate.State == "onlining" {
		aggregate, err = interfaces.GetStorageAggregate(errorHandler, *client, aggregate.UUID)
		if err != nil {
			return
		}
		waitTime = ExpontentialBackoff(waitTime, 360)
	}

	data.ID = types.StringValue(aggregate.UUID)
	data.DiskCount = types.Int64Value(aggregate.BlockStorage.Primary.DiskCount)
	data.DiskClass = types.StringValue(aggregate.BlockStorage.Primary.DiskClass)
	data.RaidType = types.StringValue(aggregate.BlockStorage.Primary.RaidType)
	data.RaidSize = types.Int64Value(aggregate.BlockStorage.Primary.RaidSize)
	data.Encryption = types.BoolValue(aggregate.DataEncryption.SoftwareEncryptionEnabled)
	data.IsMirrored = types.BoolValue(aggregate.BlockStorage.Mirror.Enabled)
	data.SnaplockType = types.StringValue(aggregate.SnaplockType)
	data.State = types.StringValue(aggregate.State)
	data.Name = types.StringValue(aggregate.Name)
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *AggregateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *AggregateResourceModel

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

	aggregate, err := interfaces.GetStorageAggregate(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}
	if aggregate == nil {
		errorHandler.MakeAndReportError("No aggregate found", fmt.Sprintf("aggregate %s not found.", data.Name.ValueString()))
		return
	}
	data.DiskCount = types.Int64Value(aggregate.BlockStorage.Primary.DiskCount)
	data.DiskClass = types.StringValue(aggregate.BlockStorage.Primary.DiskClass)
	data.RaidType = types.StringValue(aggregate.BlockStorage.Primary.RaidType)
	data.RaidSize = types.Int64Value(aggregate.BlockStorage.Primary.RaidSize)
	data.Encryption = types.BoolValue(aggregate.DataEncryption.SoftwareEncryptionEnabled)
	data.SnaplockType = types.StringValue(aggregate.SnaplockType)
	data.IsMirrored = types.BoolValue(aggregate.BlockStorage.Mirror.Enabled)
	data.SnaplockType = types.StringValue(aggregate.SnaplockType)
	data.State = types.StringValue(aggregate.State)
	data.Name = types.StringValue(aggregate.Name)
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *AggregateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *AggregateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := getRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		return
	}

	var request interfaces.StorageAggregateResourceModel

	var diskSize int
	if !plan.Name.Equal(state.Name) {
		request.Name = plan.Name.ValueString()
	}
	if !plan.State.Equal(state.State) {
		request.State = plan.State.ValueString()
	}
	if !plan.Encryption.Equal(state.Encryption) {
		request.DataEncryption = map[string]bool{}
		request.DataEncryption["software_encryption_enabled"] = plan.Encryption.ValueBool()
	}
	var BlockStoragePrimary interfaces.AggregateBlockStoragePrimary
	if !plan.RaidType.Equal(state.RaidType) {
		BlockStoragePrimary.RaidType = plan.RaidType.ValueString()
	}
	if !plan.DiskCount.Equal(state.DiskCount) {
		BlockStoragePrimary.DiskCount = plan.DiskCount.ValueInt64()
	}
	if !plan.RaidSize.Equal(state.RaidSize) {
		BlockStoragePrimary.RaidSize = plan.RaidSize.ValueInt64()
	}
	if (BlockStoragePrimary != interfaces.AggregateBlockStoragePrimary{}) {
		request.BlockStorage = map[string]interface{}{}
		var body map[string]interface{}
		if err := mapstructure.Decode(BlockStoragePrimary, &body); err != nil {
			return
		}
		request.BlockStorage["primary"] = body
	}
	// \"disk_size\" can only be specified in a PATCH operation when \"block_storage.primary.disk_count\" is being modified."
	if request.BlockStorage != nil {
		if _, ok := request.BlockStorage["disk_count"]; !ok {
			diskSize = 0
		}
	}
	if (!plan.DiskSize.Equal(state.DiskSize) || !plan.DiskSizeUnit.Equal(state.DiskSizeUnit)) && plan.DiskCount.Equal(state.DiskCount) {
		errorHandler.MakeAndReportError("error updating aggregate", "disk_size and disk_unit can only be specified in a PATCH operation when disk_count is being modified.")
		return
	}

	err = interfaces.UpdateStorageAggregate(errorHandler, *client, request, diskSize, plan.ID.ValueString())
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *AggregateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *AggregateResourceModel

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

	err = interfaces.DeleteStorageAggregate(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *AggregateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
