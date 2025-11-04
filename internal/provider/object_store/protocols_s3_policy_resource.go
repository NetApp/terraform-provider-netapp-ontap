package object_store

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsS3PolicyResource{}
var _ resource.ResourceWithImportState = &ProtocolsS3PolicyResource{}



// NewProtocolsS3PolicyResource is a helper function to simplify the provider implementation.
func NewProtocolsS3PolicyResource() resource.Resource {
	return &ProtocolsS3PolicyResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "s3_policy",
		},
	}
}

// ProtocolsS3PolicyResource defines the resource implementation.
type ProtocolsS3PolicyResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsS3PolicyStatement describes the statement data model.
type ProtocolsS3PolicyStatement struct {
	SID        types.String `tfsdk:"sid"`
	Effect     types.String `tfsdk:"effect"`
	Actions    types.Set    `tfsdk:"actions"`
	Resources  types.Set    `tfsdk:"resources"`
	Index      types.Int64  `tfsdk:"index"`
}

// ProtocolsS3PolicyResourceModel describes the resource data model.
type ProtocolsS3PolicyResourceModel struct {
	CxProfileName types.String                      `tfsdk:"cx_profile_name"`
	Name          types.String                      `tfsdk:"name"`
	Comment       types.String                      `tfsdk:"comment"`
	SVMName       types.String                      `tfsdk:"svm_name"`
	ReadOnly      types.Bool                        `tfsdk:"read_only"`
	Statements    []ProtocolsS3PolicyStatement      `tfsdk:"statements"`
	ID            types.String                      `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *ProtocolsS3PolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsS3PolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Protocols S3 Policy resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "S3 policy name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the SVM",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Optional comment for the S3 policy",
				Optional:            true,
			},
			"read_only": schema.BoolAttribute{
				MarkdownDescription: "Indicates if the policy is read-only",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
			"statements": schema.ListNestedAttribute{
				MarkdownDescription: "S3 policy statements",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"sid": schema.StringAttribute{
							MarkdownDescription: "Specifies the statement identifier which contains additional information about the statement",
							Required:            true,
						},
						"effect": schema.StringAttribute{
							MarkdownDescription: "Specifies whether access is allowed or denied",
							Required:            true,
						},
						"actions": schema.SetAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "The resource operations allowed or denied are identified by an action list",
							Optional:            true,
							Computed:            true,
						},
						"resources": schema.SetAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "List of S3 resources",
							Required:            true,
						},
						"index": schema.Int64Attribute{
							MarkdownDescription: "Specifies a unique statement index used to identify a particular statement",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// attrTypes returns a map of the attribute types for the statement.
func (o ProtocolsS3PolicyStatement) attrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"sid":        types.StringType,
		"effect":     types.StringType,
		"actions":    types.SetType{ElemType: types.StringType},
		"resources":  types.SetType{ElemType: types.StringType},

		"index":      types.Int64Type,
	}
}

// statementsSliceToList converts a slice of statements to a Go slice
func statementsSliceToList(ctx context.Context, statementsSliceIn []interfaces.ProtocolsS3PolicyStatementONTAP, diags *diag.Diagnostics) []ProtocolsS3PolicyStatement {
	var statements []ProtocolsS3PolicyStatement

	for i, stmt := range statementsSliceIn {
		// Convert actions
		actionsSet, actionsDiags := types.SetValueFrom(ctx, types.StringType, stmt.Actions)
		diags.Append(actionsDiags...)

		// Convert resources
		resourcesSet, resourcesDiags := types.SetValueFrom(ctx, types.StringType, stmt.Resources)
		diags.Append(resourcesDiags...)

		statement := ProtocolsS3PolicyStatement{
			SID:        types.StringValue(stmt.SID),
			Effect:     types.StringValue(stmt.Effect),
			Actions:    actionsSet,
			Resources:  resourcesSet,
			Index:      types.Int64Value(int64(i)),
		}

		statements = append(statements, statement)
	}

	return statements
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsS3PolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *ProtocolsS3PolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProtocolsS3PolicyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	var request interfaces.ProtocolsS3PolicyResourceBodyDataModelONTAP
	request.Name = data.Name.ValueString()
	
	if !data.Comment.IsNull() {
		request.Comment = data.Comment.ValueString()
	}

	// Convert statements (if provided)
	if len(data.Statements) > 0 {
		for _, stmt := range data.Statements {
			var stmtRequest interfaces.ProtocolsS3PolicyStatementONTAP
			stmtRequest.SID = stmt.SID.ValueString()
			stmtRequest.Effect = stmt.Effect.ValueString()

			// Convert actions - handle null/empty case
			if !stmt.Actions.IsNull() && !stmt.Actions.IsUnknown() {
				resp.Diagnostics.Append(stmt.Actions.ElementsAs(ctx, &stmtRequest.Actions, false)...)
			}
			
			// Convert resources  
			resp.Diagnostics.Append(stmt.Resources.ElementsAs(ctx, &stmtRequest.Resources, false)...)

			request.Statements = append(request.Statements, stmtRequest)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	_, err = interfaces.CreateProtocolsS3Policy(errorHandler, *client, request, data.SVMName.ValueString())
	if err != nil {
		return
	}

	// Set ID and read_only from response
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.SVMName.ValueString(), data.Name.ValueString()))
	
	// Read the resource back to get complete data
	restInfo, err := interfaces.GetProtocolsS3Policy(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		return
	}

	// Handle empty comment as null for idempotency (only if not explicitly set)
	if data.Comment.IsNull() && restInfo.Comment == "" {
		data.Comment = types.StringNull()
	} else if !data.Comment.IsNull() {
		data.Comment = types.StringValue(restInfo.Comment)
	}

	// Update statements with indices from API response - use nil for empty to preserve user intention
	if len(restInfo.Statements) == 0 {
		data.Statements = nil
	} else {
		data.Statements = statementsSliceToList(ctx, restInfo.Statements, &resp.Diagnostics)
	}

	tflog.Trace(ctx, "created S3 policy resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ProtocolsS3PolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ProtocolsS3PolicyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetProtocolsS3Policy(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		tflog.Debug(ctx, fmt.Sprintf("Error reading S3 policy %s from SVM %s: %v", data.Name.ValueString(), data.SVMName.ValueString(), err))
		return
	}
	if restInfo == nil {
		tflog.Debug(ctx, fmt.Sprintf("S3 policy %s not found in SVM %s, removing from state", data.Name.ValueString(), data.SVMName.ValueString()))
		resp.State.RemoveResource(ctx)
		return
	}

	// Map API response to model
	data.Name = types.StringValue(restInfo.Name)
	
	// Handle empty comment as null for idempotency
	if restInfo.Comment == "" {
		data.Comment = types.StringNull()
	} else {
		data.Comment = types.StringValue(restInfo.Comment)
	}
	
	data.SVMName = types.StringValue(restInfo.SVM.Name)

	// Convert statements - handle empty array as nil for consistency
	if len(restInfo.Statements) == 0 {
		data.Statements = nil
	} else {
		data.Statements = statementsSliceToList(ctx, restInfo.Statements, &resp.Diagnostics)
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.SVMName.ValueString(), data.Name.ValueString()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsS3PolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ProtocolsS3PolicyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	// Create update request body as map for proper API field naming
	updateRequest := make(map[string]interface{})
	
	if !data.Comment.IsNull() {
		updateRequest["comment"] = data.Comment.ValueString()
	}

	// Convert statements (if provided)
	if len(data.Statements) > 0 {
		statements := make([]map[string]interface{}, 0, len(data.Statements))
		for _, stmt := range data.Statements {
			stmtMap := make(map[string]interface{})
			stmtMap["sid"] = stmt.SID.ValueString()
			stmtMap["effect"] = stmt.Effect.ValueString()

			// Convert actions
			if !stmt.Actions.IsNull() && !stmt.Actions.IsUnknown() {
				var actions []string
				resp.Diagnostics.Append(stmt.Actions.ElementsAs(ctx, &actions, false)...)
				stmtMap["actions"] = actions
			}
			
			// Convert resources
			var resources []string
			resp.Diagnostics.Append(stmt.Resources.ElementsAs(ctx, &resources, false)...)
			stmtMap["resources"] = resources
			
			statements = append(statements, stmtMap)
		}
		updateRequest["statements"] = statements
	}

	if resp.Diagnostics.HasError() {
		return
	}

	err = interfaces.UpdateProtocolsS3Policy(errorHandler, *client, data.SVMName.ValueString(), data.Name.ValueString(), updateRequest)
	if err != nil {
		return
	}

	// Read back the updated configuration to ensure consistency
	restInfo, err := interfaces.GetProtocolsS3Policy(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		return
	}

	// Update data with current state from API
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.SVMName.ValueString(), data.Name.ValueString()))
	
	// Handle empty comment as null for idempotency
	if restInfo.Comment == "" {
		data.Comment = types.StringNull()
	} else {
		data.Comment = types.StringValue(restInfo.Comment)
	}
	
	// Set read_only to false as default if not provided by API
	data.ReadOnly = types.BoolValue(false)

	// Update statements - use API response for consistency
	if len(restInfo.Statements) == 0 {
		data.Statements = nil
	} else {
		data.Statements = statementsSliceToList(ctx, restInfo.Statements, &resp.Diagnostics)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsS3PolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsS3PolicyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	err = interfaces.DeleteProtocolsS3Policy(errorHandler, *client, data.SVMName.ValueString(), data.Name.ValueString())
	if err != nil {
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsS3PolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req S3 policy resource: %#v", req))
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
