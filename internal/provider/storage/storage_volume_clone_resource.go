package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &StorageVolumeCloneResource{}
var _ resource.ResourceWithImportState = &StorageVolumeCloneResource{}
var _ resource.ResourceWithModifyPlan = &StorageVolumeCloneResource{}

// NewStorageVolumeCloneResource is a helper function to simplify the provider implementation.
func NewStorageVolumeCloneResource() resource.Resource {
	return &StorageVolumeCloneResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "volume_clone",
		},
	}
}

// StorageVolumeCloneResource defines the resource implementation.
type StorageVolumeCloneResource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageVolumeCloneResourceModel describes the resource data model.
type StorageVolumeCloneResourceModel struct {
	CxProfileName      types.String `tfsdk:"cx_profile_name"`
	SVMName            types.String `tfsdk:"svm_name"`
	Name               types.String `tfsdk:"name"`
	Type               types.String `tfsdk:"type"`
	Clone              types.Object `tfsdk:"clone"`
	NAS                types.Object `tfsdk:"nas"`
	ID                 types.String `tfsdk:"id"`
}

// StorageVolumeCloneResourceClone describes the clone model.
type StorageVolumeCloneResourceClone struct {
	IsFlexclone    types.Bool   `tfsdk:"is_flexclone"`
	Split 		   types.Bool   `tfsdk:"split"`
	ParentVolume   types.String `tfsdk:"parent_volume"`
	ParentSnapshot types.String `tfsdk:"parent_snapshot"`
	ParentSVM      types.String `tfsdk:"parent_svm"`
}

// StorageVolumeCloneResourceNAS describes the NAS model.
type StorageVolumeCloneResourceNAS struct {
	JunctionPath types.String `tfsdk:"junction_path"`
	GroupID      types.Int64  `tfsdk:"group_id"`
	UserID       types.Int64  `tfsdk:"user_id"`
}

// StorageVolumeCloneResourceName describes the resource name model.
type StorageVolumeCloneResourceName struct {
	Name string `mapstructure:"name"`
}

// Metadata returns the resource type name
func (r *StorageVolumeCloneResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *StorageVolumeCloneResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageVolumeClone resource",
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
				MarkdownDescription: "Specifies the name of the volume clone.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"clone": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"is_flexclone": schema.BoolAttribute{
						MarkdownDescription: "Specifies if this volume is a normal FlexVol or FlexClone.",
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"split": schema.BoolAttribute{
						MarkdownDescription: `Initiates a split of the volume clone from its parent volume.
						Once set to true, the clone will be split asynchronously.
						After the split completes, is_flexclone will become false.`,
						Optional:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"parent_svm": schema.StringAttribute{
						MarkdownDescription: "SVM of parent volume in which clone is created off.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"parent_volume": schema.StringAttribute{
						MarkdownDescription: "Parent volume of the clone.",
						Required:            true,
					},
					"parent_snapshot": schema.StringAttribute{
						MarkdownDescription: "Parent snapshot of the clone.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The volume-type setting which should be used for the volume clone",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf(
						"rw",
						"dp",
					),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"nas": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"junction_path": schema.StringAttribute{
						MarkdownDescription: "Junction path of the clone.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"group_id": schema.Int64Attribute{
						MarkdownDescription: "The UNIX group ID for the clone volume.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"user_id": schema.Int64Attribute{
						MarkdownDescription: "The UNIX user ID for the clone volume.",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "UUID of the volume clone.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *StorageVolumeCloneResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// ModifyPlan suppresses drift-only changes once a clone has already been split.
// If state shows is_flexclone=false, the clone is detached from source and
// parent_* fields are no longer modifiable. Keep the existing state in plan.
func (r *StorageVolumeCloneResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	if req.State.Raw.IsNull() || req.Plan.Raw.IsNull() {
		return
	}

	var plan, state StorageVolumeCloneResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var stateClone StorageVolumeCloneResourceClone
	if state.Clone.IsNull() || state.Clone.IsUnknown() {
		return
	}
	resp.Diagnostics.Append(state.Clone.As(ctx, &stateClone, basetypes.ObjectAsOptions{})...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !stateClone.IsFlexclone.IsNull() && !stateClone.IsFlexclone.IsUnknown() && !stateClone.IsFlexclone.ValueBool() {
		plan.Clone = state.Clone
		resp.Diagnostics.Append(resp.Plan.Set(ctx, &plan)...)
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *StorageVolumeCloneResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StorageVolumeCloneResourceModel

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

	var restInfo *interfaces.StorageVolumeCloneGetDataModelONTAP
	restInfo, err = interfaces.GetVolumeClone(errorHandler, *client, data.SVMName.ValueString(), data.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetVolumeClone
		return
	}

	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.Type = types.StringValue(restInfo.Type)

	// Clone
	elementTypes := map[string]attr.Type{
		"is_flexclone":    types.BoolType,
		"split":           types.BoolType,
		"parent_volume":   types.StringType,
		"parent_snapshot": types.StringType,
		"parent_svm":      types.StringType,
	}
	elements := map[string]attr.Value{
		"is_flexclone":    types.BoolValue(restInfo.Clone.IsFlexclone),
		"split":           types.BoolNull(), // split is write-only; preserve state value via UseStateForUnknown() plan modifier
		"parent_volume":   types.StringNull(),
		"parent_snapshot": types.StringNull(),
		"parent_svm":      types.StringNull(),
	}
	if restInfo.Clone.ParentVolume.Name != "" {
		elements["parent_volume"] = types.StringValue(restInfo.Clone.ParentVolume.Name)
	}
	if restInfo.Clone.ParentSnapshot != nil && restInfo.Clone.ParentSnapshot.Name != "" {
		elements["parent_snapshot"] = types.StringValue(restInfo.Clone.ParentSnapshot.Name)
	}
	if restInfo.Clone.ParentSVM.Name != "" {
		elements["parent_svm"] = types.StringValue(restInfo.Clone.ParentSVM.Name)
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Clone = objectValue
	
	// NAS
	elementTypes = map[string]attr.Type{
		"junction_path": types.StringType,
		"group_id":      types.Int64Type,
		"user_id":       types.Int64Type,
	}
	elements = map[string]attr.Value{
		"junction_path": types.StringNull(),
		"group_id":      types.Int64Null(),
		"user_id":       types.Int64Null(),
	}
	if restInfo.NAS.JunctionPath != nil {
		elements["junction_path"] = types.StringValue(*restInfo.NAS.JunctionPath)
	}
	if restInfo.NAS.GroupID != nil {
		elements["group_id"] = types.Int64Value(int64(*restInfo.NAS.GroupID))
	}
	if restInfo.NAS.UserID != nil {
		elements["user_id"] = types.Int64Value(int64(*restInfo.NAS.UserID))
	}
	data.NAS, diags = types.ObjectValue(elementTypes, elements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue(restInfo.ID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve ID
func (r *StorageVolumeCloneResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data StorageVolumeCloneResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.StorageVolumeCloneResourceBodyDataModel
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

	nameValue := data.Name.ValueString()
	body.Name = &nameValue
	body.SVM.Name = data.SVMName.ValueString()
	if !data.Type.IsNull() && !data.Type.IsUnknown() {
		typeValue := data.Type.ValueString()
		body.Type = &typeValue
	}
	
	if !data.Clone.IsNull() && !data.Clone.IsUnknown() {
		var clone StorageVolumeCloneResourceClone
		diags := data.Clone.As(ctx, &clone, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.Clone = &interfaces.CloneFields{IsFlexclone: true}
		body.Clone.ParentVolume.Name = clone.ParentVolume.ValueString()
		if !clone.ParentSnapshot.IsNull() && !clone.ParentSnapshot.IsUnknown() {
			parentSnapshot := clone.ParentSnapshot.ValueString()
			if parentSnapshot != "" {
				body.Clone.ParentSnapshot = &interfaces.Name{Name: parentSnapshot}
			}
		}
		if !clone.ParentSVM.IsNull() && !clone.ParentSVM.IsUnknown() {
			parentSVM := clone.ParentSVM.ValueString()
			if parentSVM != "" {
				body.Clone.ParentSVM.Name = parentSVM
			}
		}
	}

	if !data.NAS.IsNull() && !data.NAS.IsUnknown() {
		var nas StorageVolumeCloneResourceNAS
		diags := data.NAS.As(ctx, &nas, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.NAS = &interfaces.NASFields{}
		if !nas.JunctionPath.IsNull() && !nas.JunctionPath.IsUnknown() {
			junctionPathValue := nas.JunctionPath.ValueString()
			body.NAS.JunctionPath = &junctionPathValue
		}
		if !nas.GroupID.IsNull() && !nas.GroupID.IsUnknown() {
			groupIDValue := int64(nas.GroupID.ValueInt64())
			body.NAS.GroupID = &groupIDValue
		}
		if !nas.UserID.IsNull() && !nas.UserID.IsUnknown() {
			userIDValue := int64(nas.UserID.ValueInt64())
			body.NAS.UserID = &userIDValue
		}
	}

	resource, err := interfaces.CreateVolumeClone(errorHandler, *client, body)
	if err != nil {
		// error reporting done inside CreateVolumeClone
		return
	}
	
	// Update the computed parameters
	data.ID = types.StringValue(resource.ID)
	data.SVMName = types.StringValue(resource.SVM.Name)
	data.Name = types.StringValue(resource.Name)
	data.Type = types.StringValue(resource.Type)

	// Clone
	elementTypes := map[string]attr.Type{
		"is_flexclone":   types.BoolType,
		"split":          types.BoolType,
		"parent_volume":  types.StringType,
		"parent_snapshot": types.StringType,
		"parent_svm":     types.StringType,
	}
	
	// Preserve the user's configured split value (write-only field)
	var cloneConfigValue types.Bool
	if !data.Clone.IsNull() && !data.Clone.IsUnknown() {
		var configClone StorageVolumeCloneResourceClone
		diags := data.Clone.As(ctx, &configClone, basetypes.ObjectAsOptions{})
		if !diags.HasError() && !configClone.Split.IsNull() && !configClone.Split.IsUnknown() {
			cloneConfigValue = configClone.Split
		}
	}
	
	elements := map[string]attr.Value{
		"is_flexclone":   types.BoolValue(resource.Clone.IsFlexclone),
		"split":          cloneConfigValue,
		"parent_volume":  types.StringNull(),
		"parent_snapshot": types.StringNull(),
		"parent_svm":     types.StringNull(),
	}
	if resource.Clone.ParentVolume.Name != "" {
		elements["parent_volume"] = types.StringValue(resource.Clone.ParentVolume.Name)
	}
	if resource.Clone.ParentSnapshot != nil && resource.Clone.ParentSnapshot.Name != "" {
		elements["parent_snapshot"] = types.StringValue(resource.Clone.ParentSnapshot.Name)
	}
	if resource.Clone.ParentSVM.Name != "" {
		elements["parent_svm"] = types.StringValue(resource.Clone.ParentSVM.Name)
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Clone = objectValue
	
	// NAS
	elementTypes = map[string]attr.Type{
		"junction_path": types.StringType,
		"group_id":      types.Int64Type,
		"user_id":       types.Int64Type,
	}
	elements = map[string]attr.Value{
		"junction_path": types.StringNull(),
		"group_id":      types.Int64Null(),
		"user_id":       types.Int64Null(),
	}
	if resource.NAS.JunctionPath != nil {
		elements["junction_path"] = types.StringValue(*resource.NAS.JunctionPath)
	}
	if resource.NAS.GroupID != nil {
		elements["group_id"] = types.Int64Value(int64(*resource.NAS.GroupID))
	}
	if resource.NAS.UserID != nil {
		elements["user_id"] = types.Int64Value(int64(*resource.NAS.UserID))
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.NAS = objectValue

	tflog.Trace(ctx, fmt.Sprintf("created volume clone resource, Name=%s", data.Name.ValueString()))

	// if split action is requested, perform the split operation
	if !data.Clone.IsNull() && !data.Clone.IsUnknown() {
		var clone StorageVolumeCloneResourceClone
		diags := data.Clone.As(ctx, &clone, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !clone.Split.IsNull() && !clone.Split.IsUnknown() &&
		  clone.IsFlexclone.ValueBool() && clone.Split.ValueBool() {
			err = interfaces.SplitVolumeClone(errorHandler, *client, data.ID.ValueString())
			if err != nil {
				// error reporting done inside SplitVolumeClone
				return
			}
			tflog.Trace(ctx, fmt.Sprintf("split volume clone resource, Name=%s", data.Name.ValueString()))
		}
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *StorageVolumeCloneResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config StorageVolumeCloneResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Read Terraform state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// Read Terraform config to know whether split was explicitly configured.
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

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	var planClone, stateClone, configClone StorageVolumeCloneResourceClone
	if !plan.Clone.IsNull() && !plan.Clone.IsUnknown() {
		resp.Diagnostics.Append(plan.Clone.As(ctx, &planClone, basetypes.ObjectAsOptions{})...)
	}
	if !state.Clone.IsNull() && !state.Clone.IsUnknown() {
		resp.Diagnostics.Append(state.Clone.As(ctx, &stateClone, basetypes.ObjectAsOptions{})...)
	}
	if !config.Clone.IsNull() && !config.Clone.IsUnknown() {
		resp.Diagnostics.Append(config.Clone.As(ctx, &configClone, basetypes.ObjectAsOptions{})...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Update the resource
	if stateClone.IsFlexclone.ValueBool() &&
		!configClone.Split.IsNull() && !configClone.Split.IsUnknown() && configClone.Split.ValueBool() {
		err = interfaces.SplitVolumeClone(errorHandler, *client, state.ID.ValueString())
		if err != nil {
			// error reporting done inside SplitVolumeClone
			return
		}
	}
	// Update the computed parameters
	if (planClone.IsFlexclone.IsNull() || planClone.IsFlexclone.IsUnknown()) &&
		!stateClone.IsFlexclone.IsNull() && !stateClone.IsFlexclone.IsUnknown() {
		planClone.IsFlexclone = stateClone.IsFlexclone
	}
	if (planClone.Split.IsNull() || planClone.Split.IsUnknown()) &&
		!stateClone.Split.IsNull() && !stateClone.Split.IsUnknown() {
		planClone.Split = stateClone.Split
	}
	
	// Clone
	elementTypes := map[string]attr.Type{
		"is_flexclone":    types.BoolType,
		"split":           types.BoolType,
		"parent_volume":   types.StringType,
		"parent_snapshot": types.StringType,
		"parent_svm":      types.StringType,
	}
	elements := map[string]attr.Value{
		"is_flexclone":    planClone.IsFlexclone,
		"split":           planClone.Split,
		"parent_volume":   planClone.ParentVolume,
		"parent_snapshot": planClone.ParentSnapshot,
		"parent_svm":      planClone.ParentSVM,
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Clone = objectValue
	
	tflog.Debug(ctx, fmt.Sprintf("updated volume clone resource: Name=%s", plan.Name.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *StorageVolumeCloneResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *StorageVolumeCloneResourceModel

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
func (r *StorageVolumeCloneResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
