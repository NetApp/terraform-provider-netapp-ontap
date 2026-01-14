package protocols

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsFpolicyPolicyResource{}
var _ resource.ResourceWithImportState = &ProtocolsFpolicyPolicyResource{}

// NewProtocolsFpolicyPolicyResource is a helper function to simplify the provider implementation.
func NewProtocolsFpolicyPolicyResource() resource.Resource {
	return &ProtocolsFpolicyPolicyResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "fpolicy_policy",
		},
	}
}

// ProtocolsFpolicyPolicyResource defines the resource implementation.
type ProtocolsFpolicyPolicyResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsFpolicyPolicyScopeResourceModel describes the scope data model
type ProtocolsFpolicyPolicyScopeResourceModel struct {
	IncludeVolumes                  []types.String `tfsdk:"include_volumes"`
	ExcludeVolumes                  []types.String `tfsdk:"exclude_volumes"`
	IncludeShares                   []types.String `tfsdk:"include_shares"`
	ExcludeShares                   []types.String `tfsdk:"exclude_shares"`
	IncludeExtension                []types.String `tfsdk:"include_extension"`
	ExcludeExtension                []types.String `tfsdk:"exclude_extension"`
	IncludeExportPolicies           []types.String `tfsdk:"include_export_policies"`
	ExcludeExportPolicies           []types.String `tfsdk:"exclude_export_policies"`
	CheckExtensionsOnDirectories    types.Bool     `tfsdk:"check_extensions_on_directories"`
	ObjectMonitoringWithNoExtension types.Bool     `tfsdk:"object_monitoring_with_no_extension"`
}

// ProtocolsFpolicyPolicyResourceModel describes the resource data model.
type ProtocolsFpolicyPolicyResourceModel struct {
	CxProfileName         types.String                              `tfsdk:"cx_profile_name"`
	Name                  types.String                              `tfsdk:"name"`
	SVMName               types.String                              `tfsdk:"svm_name"`
	Engine                types.String                              `tfsdk:"engine"`
	Events                []types.String                            `tfsdk:"events"`
	Priority              types.Int64                               `tfsdk:"priority"`
	Mandatory             types.Bool                                `tfsdk:"mandatory"`
	AllowPrivilegedAccess types.Bool                                `tfsdk:"allow_privileged_access"`
	PrivilegedUser        types.String                              `tfsdk:"privileged_user"`
	PassthroughRead       types.Bool                                `tfsdk:"passthrough_read"`
	Scope                 *ProtocolsFpolicyPolicyScopeResourceModel `tfsdk:"scope"`
	ID                    types.String                              `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *ProtocolsFpolicyPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsFpolicyPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsFpolicyPolicy resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Specifies the name of the FPolicy policy",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the SVM to which this FPolicy policy belongs",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"engine": schema.StringAttribute{
				MarkdownDescription: "Name of the engine to use with this policy",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"events": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of event names that will be used for event monitoring",
				Required:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "Priority for this FPolicy policy. The value must be between 1 and 10. The lower the number, the higher the priority",
				Required:            true,
			},
			"mandatory": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether file access is blocked if the policy engine is unavailable. If set to true, file access events will be blocked when the policy engine is unavailable",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"allow_privileged_access": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether privileged access is allowed. Note: This attribute can only be set during resource updates, not during initial creation",
				Optional:            true,
			},
			"privileged_user": schema.StringAttribute{
				MarkdownDescription: "Specifies the privileged user name",
				Optional:            true,
			},
			"passthrough_read": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether passthrough-read should be allowed for FPolicy servers registered for the policy",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"scope": schema.SingleNestedAttribute{
				MarkdownDescription: "Scope of the policy. Applies to the FPolicy policy to specific volumes, shares, export policies, or file extensions",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"include_volumes": schema.SetAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "List of volumes to include in the FPolicy policy scope",
						Optional:            true,
					},
					"exclude_volumes": schema.SetAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "List of volumes to exclude from the FPolicy policy scope",
						Optional:            true,
					},
					"include_shares": schema.SetAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "List of shares to include in the FPolicy policy scope",
						Optional:            true,
					},
					"exclude_shares": schema.SetAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "List of shares to exclude from the FPolicy policy scope",
						Optional:            true,
					},
					"include_extension": schema.SetAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "List of file extensions to include in the FPolicy policy scope",
						Optional:            true,
					},
					"exclude_extension": schema.SetAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "List of file extensions to exclude from the FPolicy policy scope",
						Optional:            true,
					},
					"include_export_policies": schema.SetAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "List of export policies to include in the FPolicy policy scope",
						Optional:            true,
					},
					"exclude_export_policies": schema.SetAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "List of export policies to exclude from the FPolicy policy scope",
						Optional:            true,
					},
					"check_extensions_on_directories": schema.BoolAttribute{
						MarkdownDescription: "Specifies whether the file extension check also applies to directory objects",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"object_monitoring_with_no_extension": schema.BoolAttribute{
						MarkdownDescription: "Specifies whether to monitor files that have no extension",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "FPolicy policy identifier",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsFpolicyPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ProtocolsFpolicyPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *ProtocolsFpolicyPolicyResourceModel

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

	// Validate that allow_privileged_access is not provided during create
	if !plan.AllowPrivilegedAccess.IsNull() && !plan.AllowPrivilegedAccess.IsUnknown() {
		resp.Diagnostics.AddError(
			"Invalid Attribute Configuration",
			"The allow_privileged_access attribute cannot be set during resource creation. It can only be modified during resource updates",
		)
		return
	}

	// Get SVM info to retrieve UUID
	svm, err := interfaces.GetSvmByName(errorHandler, *client, plan.SVMName.ValueString())
	if err != nil {
		return
	}

	var request interfaces.ProtocolsFpolicyPolicyResourceBodyDataModelONTAP

	request.Name = plan.Name.ValueString()

	if !plan.Engine.IsNull() && !plan.Engine.IsUnknown() {
		request.Engine = &interfaces.FpolicyPolicyEngine{Name: plan.Engine.ValueString()}
	}
	// Convert events to string array
	var events []string
	for _, event := range plan.Events {
		events = append(events, event.ValueString())
	}
	request.Events = events

	request.Priority = plan.Priority.ValueInt64()

	if !plan.Mandatory.IsNull() && !plan.Mandatory.IsUnknown() {
		request.Mandatory = plan.Mandatory.ValueBool()
	}

	if !plan.PrivilegedUser.IsNull() && !plan.PrivilegedUser.IsUnknown() {
		request.PrivilegedUser = plan.PrivilegedUser.ValueString()
	}

	if !plan.PassthroughRead.IsNull() && !plan.PassthroughRead.IsUnknown() {
		request.PassthroughRead = plan.PassthroughRead.ValueBool()
	}

	// Scope is required by the API, always provide it
	scope := interfaces.FpolicyPolicyScope{}
	if len(plan.Scope.IncludeVolumes) > 0 {
		for _, vol := range plan.Scope.IncludeVolumes {
			scope.IncludeVolumes = append(scope.IncludeVolumes, vol.ValueString())
		}
	}

	if len(plan.Scope.ExcludeVolumes) > 0 {
		for _, vol := range plan.Scope.ExcludeVolumes {
			scope.ExcludeVolumes = append(scope.ExcludeVolumes, vol.ValueString())
		}
	}

	if len(plan.Scope.IncludeShares) > 0 {
		for _, share := range plan.Scope.IncludeShares {
			scope.IncludeShares = append(scope.IncludeShares, share.ValueString())
		}
	}

	if len(plan.Scope.ExcludeShares) > 0 {
		for _, share := range plan.Scope.ExcludeShares {
			scope.ExcludeShares = append(scope.ExcludeShares, share.ValueString())
		}
	}

	if len(plan.Scope.IncludeExtension) > 0 {
		for _, ext := range plan.Scope.IncludeExtension {
			scope.IncludeExtension = append(scope.IncludeExtension, ext.ValueString())
		}
	}

	if len(plan.Scope.ExcludeExtension) > 0 {
		for _, ext := range plan.Scope.ExcludeExtension {
			scope.ExcludeExtension = append(scope.ExcludeExtension, ext.ValueString())
		}
	}

	if len(plan.Scope.IncludeExportPolicies) > 0 {
		for _, pol := range plan.Scope.IncludeExportPolicies {
			scope.IncludeExportPolicies = append(scope.IncludeExportPolicies, pol.ValueString())
		}
	}

	if len(plan.Scope.ExcludeExportPolicies) > 0 {
		for _, pol := range plan.Scope.ExcludeExportPolicies {
			scope.ExcludeExportPolicies = append(scope.ExcludeExportPolicies, pol.ValueString())
		}
	}

	if !plan.Scope.CheckExtensionsOnDirectories.IsNull() && !plan.Scope.CheckExtensionsOnDirectories.IsUnknown() {
		scope.CheckExtensionsOnDirectories = plan.Scope.CheckExtensionsOnDirectories.ValueBool()
	}

	if !plan.Scope.ObjectMonitoringWithNoExtension.IsNull() && !plan.Scope.ObjectMonitoringWithNoExtension.IsUnknown() {
		scope.ObjectMonitoringWithNoExtension = plan.Scope.ObjectMonitoringWithNoExtension.ValueBool()
	}

	request.Scope = &scope

	_, err = interfaces.CreateProtocolsFpolicyPolicy(errorHandler, *client, request, svm.UUID)
	if err != nil {
		return
	}

	// Set ID
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.SVMName.ValueString(), plan.Name.ValueString()))

	// Read back the resource to get computed values
	restInfo, err := interfaces.GetProtocolsFpolicyPolicyByName(errorHandler, *client, plan.Name.ValueString(), svm.UUID)
	if err != nil {
		return
	}

	plan.Engine = types.StringValue(restInfo.Engine.Name)
	plan.Mandatory = types.BoolValue(restInfo.Mandatory)
	plan.PassthroughRead = types.BoolValue(restInfo.PassthroughRead)
	plan.Scope.CheckExtensionsOnDirectories = types.BoolValue(restInfo.Scope.CheckExtensionsOnDirectories)
	plan.Scope.ObjectMonitoringWithNoExtension = types.BoolValue(restInfo.Scope.ObjectMonitoringWithNoExtension)

	tflog.Trace(ctx, "created a protocols_fpolicy_policy resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ProtocolsFpolicyPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *ProtocolsFpolicyPolicyResourceModel

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

	restInfo, err := interfaces.GetProtocolsFpolicyPolicyByName(errorHandler, *client, state.Name.ValueString(), svm.UUID)
	if err != nil {
		return
	}

	// Update state with read values
	state.Name = types.StringValue(restInfo.Name)

	state.Engine = types.StringValue(restInfo.Engine.Name)

	// Convert events from FpolicyPolicyEvent array to string array
	var events []types.String
	for _, event := range restInfo.Events {
		events = append(events, types.StringValue(event.Name))
	}
	state.Events = events

	state.Priority = types.Int64Value(restInfo.Priority)
	state.Mandatory = types.BoolValue(restInfo.Mandatory)
	state.PassthroughRead = types.BoolValue(restInfo.PassthroughRead)

	// Set allow_privileged_access as bool
	if !state.AllowPrivilegedAccess.IsNull() {
		state.AllowPrivilegedAccess = types.BoolValue(restInfo.AllowPrivilegedAccess)
	}

	if restInfo.PrivilegedUser != "" {
		state.PrivilegedUser = types.StringValue(restInfo.PrivilegedUser)
	}

	// Convert scope
	if len(restInfo.Scope.IncludeVolumes) > 0 || len(restInfo.Scope.ExcludeVolumes) > 0 ||
		len(restInfo.Scope.IncludeShares) > 0 || len(restInfo.Scope.ExcludeShares) > 0 ||
		len(restInfo.Scope.IncludeExtension) > 0 || len(restInfo.Scope.ExcludeExtension) > 0 ||
		len(restInfo.Scope.IncludeExportPolicies) > 0 || len(restInfo.Scope.ExcludeExportPolicies) > 0 {

		scope := &ProtocolsFpolicyPolicyScopeResourceModel{}

		for _, vol := range restInfo.Scope.IncludeVolumes {
			scope.IncludeVolumes = append(scope.IncludeVolumes, types.StringValue(vol))
		}
		for _, vol := range restInfo.Scope.ExcludeVolumes {
			scope.ExcludeVolumes = append(scope.ExcludeVolumes, types.StringValue(vol))
		}
		for _, share := range restInfo.Scope.IncludeShares {
			scope.IncludeShares = append(scope.IncludeShares, types.StringValue(share))
		}
		for _, share := range restInfo.Scope.ExcludeShares {
			scope.ExcludeShares = append(scope.ExcludeShares, types.StringValue(share))
		}
		for _, ext := range restInfo.Scope.IncludeExtension {
			scope.IncludeExtension = append(scope.IncludeExtension, types.StringValue(ext))
		}
		for _, ext := range restInfo.Scope.ExcludeExtension {
			scope.ExcludeExtension = append(scope.ExcludeExtension, types.StringValue(ext))
		}
		for _, pol := range restInfo.Scope.IncludeExportPolicies {
			scope.IncludeExportPolicies = append(scope.IncludeExportPolicies, types.StringValue(pol))
		}
		for _, pol := range restInfo.Scope.ExcludeExportPolicies {
			scope.ExcludeExportPolicies = append(scope.ExcludeExportPolicies, types.StringValue(pol))
		}

		scope.CheckExtensionsOnDirectories = types.BoolValue(restInfo.Scope.CheckExtensionsOnDirectories)
		scope.ObjectMonitoringWithNoExtension = types.BoolValue(restInfo.Scope.ObjectMonitoringWithNoExtension)

		state.Scope = scope
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsFpolicyPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *ProtocolsFpolicyPolicyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	client, err := connection.GetRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		return
	}

	if plan.ID.IsNull() {
		errorHandler.MakeAndReportError("ID is null", "protocols_fpolicy_policy ID is null")
		return
	}

	// Get SVM info to retrieve UUID
	svm, err := interfaces.GetSvmByName(errorHandler, *client, plan.SVMName.ValueString())
	if err != nil {
		return
	}

	// Create the request body (without name field for PATCH)
	request := interfaces.ProtocolsFpolicyPolicyResourceUpdateBodyDataModelONTAP{}

	// Engine: only set if not null/empty
	if !plan.Engine.IsNull() && !plan.Engine.IsUnknown() {
		engine := interfaces.FpolicyPolicyEngine{Name: plan.Engine.ValueString()}
		request.Engine = &engine
	}

	// Convert events to string array
	var events []string
	for _, event := range plan.Events {
		events = append(events, event.ValueString())
	}
	request.Events = events

	if plan.Priority.ValueInt64() != state.Priority.ValueInt64() {
		request.Priority = plan.Priority.ValueInt64Pointer()
	}

	if !plan.Mandatory.IsNull() && !plan.Mandatory.IsUnknown() {
		if plan.Mandatory.ValueBool() != state.Mandatory.ValueBool() {
			request.Mandatory = plan.Mandatory.ValueBoolPointer()
		}
	}

	if !plan.AllowPrivilegedAccess.IsNull() && !plan.AllowPrivilegedAccess.IsUnknown() {
		if plan.AllowPrivilegedAccess.ValueBool() != state.AllowPrivilegedAccess.ValueBool() {
			request.AllowPrivilegedAccess = plan.AllowPrivilegedAccess.ValueBoolPointer()
		}
	}

	if !plan.PrivilegedUser.IsNull() && !plan.PrivilegedUser.IsUnknown() {
		if plan.PrivilegedUser.ValueString() != state.PrivilegedUser.ValueString() {
			request.PrivilegedUser = plan.PrivilegedUser.ValueString()
		}
	}

	if !plan.PassthroughRead.IsNull() && !plan.PassthroughRead.IsUnknown() {
		if plan.PassthroughRead.ValueBool() != state.PassthroughRead.ValueBool() {
			request.PassthroughRead = plan.PassthroughRead.ValueBoolPointer()
		}
	}
	// Convert scope if provided
	if plan.Scope != nil {
		scope := interfaces.FpolicyPolicyScope{}

		for _, vol := range plan.Scope.IncludeVolumes {
			scope.IncludeVolumes = append(scope.IncludeVolumes, vol.ValueString())
		}
		for _, vol := range plan.Scope.ExcludeVolumes {
			scope.ExcludeVolumes = append(scope.ExcludeVolumes, vol.ValueString())
		}
		for _, share := range plan.Scope.IncludeShares {
			scope.IncludeShares = append(scope.IncludeShares, share.ValueString())
		}
		for _, share := range plan.Scope.ExcludeShares {
			scope.ExcludeShares = append(scope.ExcludeShares, share.ValueString())
		}
		for _, ext := range plan.Scope.IncludeExtension {
			scope.IncludeExtension = append(scope.IncludeExtension, ext.ValueString())
		}
		for _, ext := range plan.Scope.ExcludeExtension {
			scope.ExcludeExtension = append(scope.ExcludeExtension, ext.ValueString())
		}
		for _, pol := range plan.Scope.IncludeExportPolicies {
			scope.IncludeExportPolicies = append(scope.IncludeExportPolicies, pol.ValueString())
		}
		for _, pol := range plan.Scope.ExcludeExportPolicies {
			scope.ExcludeExportPolicies = append(scope.ExcludeExportPolicies, pol.ValueString())
		}

		if !plan.Scope.CheckExtensionsOnDirectories.IsNull() && !plan.Scope.CheckExtensionsOnDirectories.IsUnknown() {
			if plan.Scope.CheckExtensionsOnDirectories.ValueBool() != state.Scope.CheckExtensionsOnDirectories.ValueBool() {
				scope.CheckExtensionsOnDirectories = plan.Scope.CheckExtensionsOnDirectories.ValueBool()
			}
		}
		if !plan.Scope.ObjectMonitoringWithNoExtension.IsNull() && !plan.Scope.ObjectMonitoringWithNoExtension.IsUnknown() {
			if plan.Scope.ObjectMonitoringWithNoExtension.ValueBool() != state.Scope.ObjectMonitoringWithNoExtension.ValueBool() {
				scope.ObjectMonitoringWithNoExtension = plan.Scope.ObjectMonitoringWithNoExtension.ValueBool()
			}
		}

		request.Scope = &scope
	}

	err = interfaces.UpdateProtocolsFpolicyPolicy(errorHandler, *client, request, svm.UUID, plan.Name.ValueString())
	if err != nil {
		return
	}

	tflog.Trace(ctx, "updated a protocols_fpolicy_policy resource")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsFpolicyPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state *ProtocolsFpolicyPolicyResourceModel

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
		errorHandler.MakeAndReportError("ID is null", "protocols_fpolicy_policy ID is null")
		return
	}

	// Get SVM info to retrieve UUID
	svm, err := interfaces.GetSvmByName(errorHandler, *client, state.SVMName.ValueString())
	if err != nil {
		return
	}

	err = interfaces.DeleteProtocolsFpolicyPolicy(errorHandler, *client, svm.UUID, state.Name.ValueString())
	if err != nil {
		return
	}

	tflog.Trace(ctx, "deleted a protocols_fpolicy_policy resource")
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsFpolicyPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a protocols fpolicy policy resource: %#v", req))
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
