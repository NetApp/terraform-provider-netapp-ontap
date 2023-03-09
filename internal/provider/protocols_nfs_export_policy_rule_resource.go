package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ExportPolicyRuleResource{}

var _ resource.ResourceWithImportState = &ExportPolicyRuleResource{}

// NewExportPolicyRuleResource is a helper function to simplify the provider implementation.
func NewExportPolicyRuleResource() resource.Resource {
	return &ExportPolicyRuleResource{
		config: resourceOrDataSourceConfig{
			name: "protocols_nfs_export_policy_rule_resource",
		},
	}
}

// ExportPolicyRuleResource defines the resource implementation.
type ExportPolicyRuleResource struct {
	config resourceOrDataSourceConfig
}

// ExportPolicyRuleResourceModel describes the resource data model.
type ExportPolicyRuleResourceModel struct {
	CxProfileName       types.String   `tfsdk:"cx_profile_name"`
	ExportPolicyID      types.String   `tfsdk:"export_policy_id"`
	Vserver             types.String   `tfsdk:"vserver"`
	RoRule              []types.String `tfsdk:"ro_rule"`
	RwRule              []types.String `tfsdk:"rw_rule"`
	Protocols           []types.String `tfsdk:"protocols"`
	AnonymousUser       types.String   `tfsdk:"anonymous_user"`
	Superuser           []types.String `tfsdk:"superuser"`
	AllowDeviceCreation types.Bool     `tfsdk:"allow_device_creation"`
	NtfsUnixSecurity    types.String   `tfsdk:"ntfs_unix_security"`
	ChownMode           types.String   `tfsdk:"chown_mode"`
	AllowSuid           types.Bool     `tfsdk:"allow_suid"`
	ClientsMatch        []types.String `tfsdk:"clients_match"`
	Index               types.Int64    `tfsdk:"index"`
}

// Metadata returns the resource type name.
func (r *ExportPolicyRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *ExportPolicyRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Export policy rule resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"vserver": schema.StringAttribute{
				MarkdownDescription: "Name of the vserver to use",
				Required:            true,
			},
			"export_policy_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Export policy identifier",
			},
			"ro_rule": schema.ListAttribute{
				Optional:            true,
				MarkdownDescription: "RO Access Rule",
				ElementType:         types.StringType,
			},
			"rw_rule": schema.ListAttribute{
				Optional:            true,
				MarkdownDescription: "RW Access Rule",
				ElementType:         types.StringType,
			},
			"clients_match": schema.ListAttribute{
				Required:            true,
				MarkdownDescription: "List of Client Match Hostnames, IP Addresses, Netgroups, or Domains",
				ElementType:         types.StringType,
			},
			"protocols": schema.ListAttribute{
				Optional:            true,
				MarkdownDescription: "Access Protocol",
				ElementType:         types.StringType,
			},
			"anonymous_user": schema.StringAttribute{
				MarkdownDescription: "User ID To Which Anonymous Users Are Mapped",
				Optional:            true,
			},
			"superuser": schema.ListAttribute{
				MarkdownDescription: "Superuser Security Types",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"allow_device_creation": schema.BoolAttribute{
				MarkdownDescription: "Allow Creation of Devices",
				Optional:            true,
			},
			"ntfs_unix_security": schema.StringAttribute{
				MarkdownDescription: "NTFS export UNIX security options",
				Optional:            true,
			},
			"chown_mode": schema.StringAttribute{
				MarkdownDescription: "Specifies who is authorized to change the ownership mode of a file",
				Optional:            true,
			},
			"allow_suid": schema.BoolAttribute{
				MarkdownDescription: "Honor SetUID Bits in SETATTR",
				Optional:            true,
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "rule index",
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					IntUseStateForUnknown(),
				},
			},
		},
	}
}

// IntPlanModify implements planmodifier.Int64
type IntPlanModify struct {
}

// IntUseStateForUnknown is the wrapper function returns the plan modifier
func IntUseStateForUnknown() planmodifier.Int64 {
	return IntPlanModify{}
}

// Description is the method required to implement planmodifier.Int64
func (s IntPlanModify) Description(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// MarkdownDescription is the method required to implement planmodifier.Int64
func (s IntPlanModify) MarkdownDescription(_ context.Context) string {
	return "Once set, the value of this attribute in state will not change."
}

// PlanModifyInt64 is the method required to implement planmodifier.Int64
func (s IntPlanModify) PlanModifyInt64(_ context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	// Do nothing if there is no state value.
	if req.StateValue.IsNull() {
		return
	}

	// Do nothing if there is a known planned value.
	if !req.PlanValue.IsUnknown() {
		return
	}

	// Do nothing if there is an unknown configuration value, otherwise interpolation gets messed up.
	if req.ConfigValue.IsUnknown() {
		return
	}
	resp.PlanValue = req.StateValue
}

// Configure adds the provider configured client to the resource.
func (r *ExportPolicyRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ExportPolicyRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ExportPolicyRuleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var request interfaces.ExportpolicyRuleResourceModel
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	var roRule, rwRule, protocols, superuser []string
	for _, e := range data.RoRule {
		roRule = append(roRule, e.ValueString())
	}
	for _, e := range data.RwRule {
		rwRule = append(rwRule, e.ValueString())
	}
	for _, e := range data.Protocols {
		protocols = append(protocols, e.ValueString())
	}
	for _, e := range data.Superuser {
		superuser = append(superuser, e.ValueString())
	}
	for _, e := range data.ClientsMatch {
		request.ClientsMatch = append(request.ClientsMatch, map[string]string{"match": e.ValueString()})
	}
	request.RoRule = roRule
	request.RwRule = rwRule
	request.Protocols = protocols
	request.Superuser = superuser

	//optional params
	if !data.AnonymousUser.IsNull() {
		request.AnonymousUser = data.AnonymousUser.ValueString()
	}
	if !data.AllowDeviceCreation.IsNull() {
		request.AllowDeviceCreation = data.AllowDeviceCreation.ValueBool()
	}
	if !data.AllowSuid.IsNull() {
		request.AllowSuid = data.AllowSuid.ValueBool()
	}
	if !data.ChownMode.IsNull() {
		request.ChownMode = data.ChownMode.ValueString()
	}
	if !data.NtfsUnixSecurity.IsNull() {
		request.NtfsUnixSecurity = data.NtfsUnixSecurity.ValueString()
	}

	exportPolicyRule, err := interfaces.CreateExportPolicyRule(errorHandler, *client, request, data.ExportPolicyID.ValueString())
	if err != nil {
		return
	}
	data.Index = types.Int64Value(exportPolicyRule.Index)

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ExportPolicyRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ExportPolicyRuleResourceModel

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

	restInfo, err := interfaces.GetExportPolicyRule(errorHandler, *client, data.ExportPolicyID.ValueString(), data.Index.ValueInt64())
	if err != nil {
		return
	}
	var roRule, rwRule, protocols, superuser, clientsMatch []types.String
	for _, e := range restInfo.RoRule {
		roRule = append(roRule, types.StringValue(e))
	}

	for _, e := range restInfo.RwRule {
		rwRule = append(rwRule, types.StringValue(e))
	}
	for _, e := range restInfo.Protocols {
		protocols = append(protocols, types.StringValue(e))
	}
	for _, e := range restInfo.Superuser {
		superuser = append(superuser, types.StringValue(e))
	}
	data.RoRule = roRule
	data.RwRule = rwRule
	data.Protocols = protocols
	data.Superuser = superuser

	for _, e := range restInfo.ClientsMatch {
		clientsMatch = append(clientsMatch, types.StringValue(e.Match))
	}
	data.ClientsMatch = clientsMatch

	if !data.AllowDeviceCreation.IsNull() {
		data.AllowDeviceCreation = types.BoolValue(restInfo.AllowDeviceCreation)
	}

	if !data.AllowSuid.IsNull() {
		data.AllowSuid = types.BoolValue(restInfo.AllowSuid)
	}

	if !data.ChownMode.IsNull() {
		data.ChownMode = types.StringValue(restInfo.ChownMode)
	}
	if !data.NtfsUnixSecurity.IsNull() {
		data.NtfsUnixSecurity = types.StringValue(restInfo.NtfsUnixSecurity)
	}
	if !data.AnonymousUser.IsNull() {
		data.AnonymousUser = types.StringValue(restInfo.AnonymousUser)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ExportPolicyRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ExportPolicyRuleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	var request interfaces.ExportpolicyRuleResourceModel

	var roRule, rwRule, protocols, superuser []string
	for _, e := range data.RoRule {
		roRule = append(roRule, e.ValueString())
	}
	for _, e := range data.RwRule {
		rwRule = append(rwRule, e.ValueString())
	}
	for _, e := range data.Protocols {
		protocols = append(protocols, e.ValueString())
	}
	for _, e := range data.Superuser {
		superuser = append(superuser, e.ValueString())
	}
	for _, e := range data.ClientsMatch {
		request.ClientsMatch = append(request.ClientsMatch, map[string]string{"match": e.ValueString()})
	}
	request.RoRule = roRule
	request.RwRule = rwRule
	request.Protocols = protocols
	request.Superuser = superuser

	//optional params
	if !data.AnonymousUser.IsNull() {
		request.AnonymousUser = data.AnonymousUser.ValueString()
	}
	if !data.AllowDeviceCreation.IsNull() {
		request.AllowDeviceCreation = data.AllowDeviceCreation.ValueBool()
	}
	if !data.AllowSuid.IsNull() {
		request.AllowSuid = data.AllowSuid.ValueBool()
	}
	if !data.ChownMode.IsNull() {
		request.ChownMode = data.ChownMode.ValueString()
	}
	if !data.NtfsUnixSecurity.IsNull() {
		request.NtfsUnixSecurity = data.NtfsUnixSecurity.ValueString()
	}

	_, err = interfaces.UpdateExportPolicyRule(errorHandler, *client, request, data.ExportPolicyID.ValueString(), data.Index.ValueInt64())
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ExportPolicyRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ExportPolicyRuleResourceModel

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

	err = interfaces.DeleteExportPolicyRule(errorHandler, *client, data.ExportPolicyID.ValueString(), data.Index.ValueInt64())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ExportPolicyRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
