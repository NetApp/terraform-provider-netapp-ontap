package name_services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &NameServicesNameMappingResource{}
var _ resource.ResourceWithImportState = &NameServicesNameMappingResource{}

// NewNameServicesNameMappingResource is a helper function to simplify the provider implementation.
func NewNameServicesNameMappingResource() resource.Resource {
	return &NameServicesNameMappingResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "name_services_name_mapping",
		},
	}
}

// NameServicesNameMappingResource defines the resource implementation.
type NameServicesNameMappingResource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesNameMappingResourceModel describes the resource data model.
type NameServicesNameMappingResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Direction     types.String `tfsdk:"direction"`
	Index         types.Int64  `tfsdk:"index"`
	Pattern       types.String `tfsdk:"pattern"`
	ClientMatch   types.String `tfsdk:"client_match"`
	Replacement   types.String `tfsdk:"replacement"`
	ID            types.String `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *NameServicesNameMappingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *NameServicesNameMappingResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a name mapping. Name mappings are used to map CIFS identities to UNIX identities, Kerberos identities to UNIX identities, UNIX identities to CIFS identities, S3 to UNIX identities, and S3 to CIFS identities.",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the SVM that owns the name mapping",
				Required:            true,
			},
			"direction": schema.StringAttribute{
				MarkdownDescription: "Direction of the name mapping. Possible values: krb_unix, win_unix, unix_win, s3_unix, s3_win. Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "Position in the list of name mappings (1–2147483647). Cannot be changed after creation.",
				Required:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"pattern": schema.StringAttribute{
				MarkdownDescription: "UNIX-style regular expression used to match the name to be mapped (1–256 characters).",
				Optional:            true,
			},
			"client_match": schema.StringAttribute{
				MarkdownDescription: "Client workstation IP address or hostname matched when searching for the pattern.",
				Optional:            true,
			},
			"replacement": schema.StringAttribute{
				MarkdownDescription: "Replacement name used when the pattern matches (1–256 characters).",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Composite identifier: {svm_uuid}/{direction}/{index}",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *NameServicesNameMappingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Read refreshes the Terraform state with the latest data.
func (r *NameServicesNameMappingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NameServicesNameMappingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetNameServicesNameMappingByIndex(errorHandler, *client,
		data.SVMName.ValueString(), data.Direction.ValueString(), int(data.Index.ValueInt64()))
	if err != nil {
		return
	}

	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Direction = types.StringValue(restInfo.Direction)
	data.Index = types.Int64Value(int64(restInfo.Index))
	data.Pattern = types.StringValue(restInfo.Pattern)
	data.ClientMatch = types.StringValue(restInfo.ClientMatch)
	data.Replacement = types.StringValue(restInfo.Replacement)
	data.ID = types.StringValue(restInfo.SVM.UUID + "/" + restInfo.Direction + "/" + strconv.Itoa(restInfo.Index))

	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *NameServicesNameMappingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *NameServicesNameMappingResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.NameServicesNameMappingResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.SVM.Name = data.SVMName.ValueString()
	body.Direction = data.Direction.ValueString()
	body.Index = int(data.Index.ValueInt64())
	if !data.Pattern.IsNull() {
		body.Pattern = data.Pattern.ValueString()
	}
	if !data.ClientMatch.IsNull() {
		body.ClientMatch = data.ClientMatch.ValueString()
	}
	if !data.Replacement.IsNull() {
		body.Replacement = data.Replacement.ValueString()
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	restInfo, err := interfaces.CreateNameServicesNameMapping(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Direction = types.StringValue(restInfo.Direction)
	data.Index = types.Int64Value(int64(restInfo.Index))
	data.Pattern = types.StringValue(restInfo.Pattern)
	data.ClientMatch = types.StringValue(restInfo.ClientMatch)
	data.Replacement = types.StringValue(restInfo.Replacement)
	data.ID = types.StringValue(restInfo.SVM.UUID + "/" + restInfo.Direction + "/" + strconv.Itoa(restInfo.Index))

	tflog.Trace(ctx, "created a resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *NameServicesNameMappingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *NameServicesNameMappingResourceModel

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

	var body interfaces.NameServicesNameMappingResourceBodyDataModelONTAP
	if !plan.Pattern.Equal(state.Pattern) {
		body.Pattern = plan.Pattern.ValueString()
	}
	if !plan.ClientMatch.Equal(state.ClientMatch) {
		body.ClientMatch = plan.ClientMatch.ValueString()
	}
	if !plan.Replacement.Equal(state.Replacement) {
		body.Replacement = plan.Replacement.ValueString()
	}

	err = interfaces.UpdateNameServicesNameMapping(errorHandler, *client, state.ID.ValueString(), body)
	if err != nil {
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *NameServicesNameMappingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *NameServicesNameMappingResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("ID is null", "name_services_name_mapping ID is null")
		return
	}

	err = interfaces.DeleteNameServicesNameMapping(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
// Import ID format: "svm_name,direction,index,cx_profile_name"
func (r *NameServicesNameMappingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a name_services_name_mapping resource: %#v", req))
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprint("Expected ID in the format 'svm_name,direction,index,cx_profile_name', got: ", req.ID),
		)
		return
	}

	index, err := strconv.ParseInt(idParts[2], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError(
			"Invalid index in Import Identifier",
			fmt.Sprintf("Expected integer for index, got: %s", idParts[2]),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("direction"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("index"), index)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[3])...)
}
