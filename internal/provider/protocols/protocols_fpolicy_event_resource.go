package protocols

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsFpolicyEventResource{}
var _ resource.ResourceWithImportState = &ProtocolsFpolicyEventResource{}

// NewProtocolsFpolicyEventResource is a helper function to simplify the provider implementation.
func NewProtocolsFpolicyEventResource() resource.Resource {
	return &ProtocolsFpolicyEventResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "fpolicy_event",
		},
	}
}

// ProtocolsFpolicyEventResource defines the resource implementation.
type ProtocolsFpolicyEventResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsFpolicyEventResourceModel describes the resource data model.
type ProtocolsFpolicyEventResourceModel struct {
	CxProfileName    types.String   `tfsdk:"cx_profile_name"`
	Name             types.String   `tfsdk:"name"`
	SVMName          types.String   `tfsdk:"svm_name"`
	Protocol         types.String   `tfsdk:"protocol"`
	FileOperations   []types.String `tfsdk:"file_operations"`
	Filters          []types.String `tfsdk:"filters"`
	VolumeMonitoring types.Bool     `tfsdk:"volume_monitoring"`
	ID               types.String   `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *ProtocolsFpolicyEventResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsFpolicyEventResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsFpolicyEvent resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Specifies the name of the FPolicy event",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the SVM to which this FPolicy event belongs",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol for which the FPolicy event is applicable. Valid values are cifs, nfsv3, or nfsv4",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("cifs", "nfsv3", "nfsv4"),
				},
			},
			"file_operations": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "File operations to be monitored",
				Optional:            true,
			},
			"filters": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Filters to be applied for the specified file operations",
				Optional:            true,
			},
			"volume_monitoring": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether volume operation monitoring is required",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "FPolicy event identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsFpolicyEventResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create a resource
func (r *ProtocolsFpolicyEventResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *ProtocolsFpolicyEventResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	client, err := connection.GetRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		return
	}

	// Get SVM info to retrieve UUID
	svm, err := interfaces.GetSvmByName(errorHandler, *client, plan.SVMName.ValueString())
	if err != nil {
		return
	}

	var request interfaces.ProtocolsFpolicyEventResourceBodyDataModelONTAP

	request.Name = plan.Name.ValueString()
	if !plan.Protocol.IsUnknown() {
		request.Protocol = plan.Protocol.ValueString()
	}

	// Convert file operations and filters using helper functions
	if len(plan.FileOperations) > 0 {
		request.FileOperations = convertFileOperationsToStruct(plan.FileOperations)
	}
	if len(plan.Filters) > 0 {
		request.Filters = convertFiltersToStruct(plan.Filters)
	}

	request.VolumeMonitoring = plan.VolumeMonitoring.ValueBool()

	resource, err := interfaces.CreateProtocolsFpolicyEvent(errorHandler, *client, request, svm.UUID)
	if err != nil {
		return
	}

	// Set ID
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.SVMName.ValueString(), plan.Name.ValueString()))

	// Read back the resource to get computed values
	readResource, err := interfaces.GetProtocolsFpolicyEventByName(errorHandler, *client, plan.Name.ValueString(), svm.UUID)
	if err != nil {
		return
	}

	// Always read back computed values from API
	plan.VolumeMonitoring = types.BoolValue(readResource.VolumeMonitoring)

	// Suppress the unused variable warning
	_ = resource

	tflog.Trace(ctx, "created a protocols_fpolicy_event resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ProtocolsFpolicyEventResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *ProtocolsFpolicyEventResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	client, err := connection.GetRestClient(errorHandler, r.config, state.CxProfileName)
	if err != nil {
		return
	}

	// Get SVM info to retrieve UUID
	svm, err := interfaces.GetSvmByName(errorHandler, *client, state.SVMName.ValueString())
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetProtocolsFpolicyEventByName(errorHandler, *client, state.Name.ValueString(), svm.UUID)
	if err != nil {
		return
	}

	// Update state with read values
	state.Name = types.StringValue(restInfo.Name)
	state.Protocol = types.StringValue(restInfo.Protocol)
	state.VolumeMonitoring = types.BoolValue(restInfo.VolumeMonitoring)

	// Convert file operations and filters using helper functions
	state.FileOperations = convertFileOperationsFromStruct(restInfo.FileOperations)
	state.Filters = convertFiltersFromStruct(restInfo.Filters)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsFpolicyEventResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *ProtocolsFpolicyEventResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	client, err := connection.GetRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		return
	}

	if plan.ID.IsNull() {
		errorHandler.MakeAndReportError("ID is null", "protocols_fpolicy_event ID is null")
		return
	}

	// Get SVM info to retrieve UUID
	svm, err := interfaces.GetSvmByName(errorHandler, *client, plan.SVMName.ValueString())
	if err != nil {
		return
	}

	// Create the request body (without name field for PATCH)
	request := interfaces.ProtocolsFpolicyEventResourceUpdateBodyDataModelONTAP{
		Protocol: plan.Protocol.ValueString(),
	}

	// Always convert file operations and filters (even if empty) to ensure proper updates
	request.FileOperations = convertFileOperationsToStruct(plan.FileOperations)
	request.Filters = convertFiltersToStruct(plan.Filters)

	val := plan.VolumeMonitoring.ValueBool()
	request.VolumeMonitoring = &val

	err = interfaces.UpdateProtocolsFpolicyEvent(errorHandler, *client, request, svm.UUID, plan.Name.ValueString())
	if err != nil {
		return
	}

	tflog.Trace(ctx, "updated a protocols_fpolicy_event resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsFpolicyEventResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *ProtocolsFpolicyEventResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	client, err := connection.GetRestClient(errorHandler, r.config, state.CxProfileName)
	if err != nil {
		return
	}

	if state.ID.IsNull() {
		errorHandler.MakeAndReportError("ID is null", "protocols_fpolicy_event ID is null")
		return
	}

	// Get SVM info to retrieve UUID
	svm, err := interfaces.GetSvmByName(errorHandler, *client, state.SVMName.ValueString())
	if err != nil {
		return
	}

	err = interfaces.DeleteProtocolsFpolicyEvent(errorHandler, *client, svm.UUID, state.Name.ValueString())
	if err != nil {
		return
	}

	tflog.Trace(ctx, "deleted a protocols_fpolicy_event resource")
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsFpolicyEventResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a protocols fpolicy event resource: %#v", req))
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
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), fmt.Sprintf("%s/%s", idParts[1], idParts[0]))...)
}

// convertFileOperationsToStruct converts a slice of types.String to FpolicyEventFileOperations struct
func convertFileOperationsToStruct(operations []types.String) interfaces.FpolicyEventFileOperations {
	ops := interfaces.FpolicyEventFileOperations{}
	for _, op := range operations {
		opStr := op.ValueString()
		switch opStr {
		case "access":
			ops.Access = true
		case "close":
			ops.Close = true
		case "create":
			ops.Create = true
		case "create_dir":
			ops.CreateDir = true
		case "delete":
			ops.Delete = true
		case "delete_dir":
			ops.DeleteDir = true
		case "getattr":
			ops.GetAttr = true
		case "link":
			ops.Link = true
		case "lookup":
			ops.Lookup = true
		case "open":
			ops.Open = true
		case "read":
			ops.Read = true
		case "rename":
			ops.Rename = true
		case "rename_dir":
			ops.RenameDir = true
		case "setattr":
			ops.SetAttr = true
		case "symlink":
			ops.Symlink = true
		case "write":
			ops.Write = true
		}
	}
	return ops
}

// convertFiltersToStruct converts a slice of types.String to FpolicyEventFilters struct
func convertFiltersToStruct(filters []types.String) interfaces.FpolicyEventFilters {
	f := interfaces.FpolicyEventFilters{}
	for _, filter := range filters {
		filterStr := filter.ValueString()
		switch filterStr {
		case "close_with_modification":
			f.CloseWithModification = true
		case "close_with_read":
			f.CloseWithRead = true
		case "close_without_modification":
			f.CloseWithoutModification = true
		case "exclude_directory":
			f.ExcludeDirectory = true
		case "first_read":
			f.FirstRead = true
		case "first_write":
			f.FirstWrite = true
		case "monitor_ads":
			f.MonitorAds = true
		case "offline_bit":
			f.OfflineBit = true
		case "open_with_delete_intent":
			f.OpenWithDeleteIntent = true
		case "open_with_write_intent":
			f.OpenWithWriteIntent = true
		case "setattr_with_access_time_change":
			f.SetAttrWithAccessTimeChange = true
		case "setattr_with_allocation_size_change":
			f.SetAttrWithAllocationSizeChange = true
		case "setattr_with_creation_time_change":
			f.SetAttrWithCreationTimeChange = true
		case "setattr_with_dacl_change":
			f.SetAttrWithDaclChange = true
		case "setattr_with_group_change":
			f.SetAttrWithGroupChange = true
		case "setattr_with_mode_change":
			f.SetAttrWithModeChange = true
		case "setattr_with_modify_time_change":
			f.SetAttrWithModifyTimeChange = true
		case "setattr_with_owner_change":
			f.SetAttrWithOwnerChange = true
		case "setattr_with_sacl_change":
			f.SetAttrWithSaclChange = true
		case "setattr_with_size_change":
			f.SetAttrWithSizeChange = true
		case "write_with_size_change":
			f.WriteWithSizeChange = true
		}
	}
	return f
}

// convertFileOperationsFromStruct converts FpolicyEventFileOperations struct to a slice of types.String
func convertFileOperationsFromStruct(ops interfaces.FpolicyEventFileOperations) []types.String {
	var operations []types.String
	if ops.Access {
		operations = append(operations, types.StringValue("access"))
	}
	if ops.Close {
		operations = append(operations, types.StringValue("close"))
	}
	if ops.Create {
		operations = append(operations, types.StringValue("create"))
	}
	if ops.CreateDir {
		operations = append(operations, types.StringValue("create_dir"))
	}
	if ops.Delete {
		operations = append(operations, types.StringValue("delete"))
	}
	if ops.DeleteDir {
		operations = append(operations, types.StringValue("delete_dir"))
	}
	if ops.GetAttr {
		operations = append(operations, types.StringValue("getattr"))
	}
	if ops.Link {
		operations = append(operations, types.StringValue("link"))
	}
	if ops.Lookup {
		operations = append(operations, types.StringValue("lookup"))
	}
	if ops.Open {
		operations = append(operations, types.StringValue("open"))
	}
	if ops.Read {
		operations = append(operations, types.StringValue("read"))
	}
	if ops.Rename {
		operations = append(operations, types.StringValue("rename"))
	}
	if ops.RenameDir {
		operations = append(operations, types.StringValue("rename_dir"))
	}
	if ops.SetAttr {
		operations = append(operations, types.StringValue("setattr"))
	}
	if ops.Symlink {
		operations = append(operations, types.StringValue("symlink"))
	}
	if ops.Write {
		operations = append(operations, types.StringValue("write"))
	}
	return operations
}

// convertFiltersFromStruct converts FpolicyEventFilters struct to a slice of types.String
func convertFiltersFromStruct(f interfaces.FpolicyEventFilters) []types.String {
	var filters []types.String
	if f.CloseWithModification {
		filters = append(filters, types.StringValue("close_with_modification"))
	}
	if f.CloseWithRead {
		filters = append(filters, types.StringValue("close_with_read"))
	}
	if f.CloseWithoutModification {
		filters = append(filters, types.StringValue("close_without_modification"))
	}
	if f.ExcludeDirectory {
		filters = append(filters, types.StringValue("exclude_directory"))
	}
	if f.FirstRead {
		filters = append(filters, types.StringValue("first_read"))
	}
	if f.FirstWrite {
		filters = append(filters, types.StringValue("first_write"))
	}
	if f.MonitorAds {
		filters = append(filters, types.StringValue("monitor_ads"))
	}
	if f.OfflineBit {
		filters = append(filters, types.StringValue("offline_bit"))
	}
	if f.OpenWithDeleteIntent {
		filters = append(filters, types.StringValue("open_with_delete_intent"))
	}
	if f.OpenWithWriteIntent {
		filters = append(filters, types.StringValue("open_with_write_intent"))
	}
	if f.SetAttrWithAccessTimeChange {
		filters = append(filters, types.StringValue("setattr_with_access_time_change"))
	}
	if f.SetAttrWithAllocationSizeChange {
		filters = append(filters, types.StringValue("setattr_with_allocation_size_change"))
	}
	if f.SetAttrWithCreationTimeChange {
		filters = append(filters, types.StringValue("setattr_with_creation_time_change"))
	}
	if f.SetAttrWithDaclChange {
		filters = append(filters, types.StringValue("setattr_with_dacl_change"))
	}
	if f.SetAttrWithGroupChange {
		filters = append(filters, types.StringValue("setattr_with_group_change"))
	}
	if f.SetAttrWithModeChange {
		filters = append(filters, types.StringValue("setattr_with_mode_change"))
	}
	if f.SetAttrWithModifyTimeChange {
		filters = append(filters, types.StringValue("setattr_with_modify_time_change"))
	}
	if f.SetAttrWithOwnerChange {
		filters = append(filters, types.StringValue("setattr_with_owner_change"))
	}
	if f.SetAttrWithSaclChange {
		filters = append(filters, types.StringValue("setattr_with_sacl_change"))
	}
	if f.SetAttrWithSizeChange {
		filters = append(filters, types.StringValue("setattr_with_size_change"))
	}
	if f.WriteWithSizeChange {
		filters = append(filters, types.StringValue("write_with_size_change"))
	}
	return filters
}
