package protocols

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsS3UserResource{}
var _ resource.ResourceWithImportState = &ProtocolsS3UserResource{}
var _ resource.ResourceWithModifyPlan = &ProtocolsS3UserResource{}

// NewProtocolsS3UserResource is a helper function to simplify the provider implementation.
func NewProtocolsS3UserResource() resource.Resource {
	return &ProtocolsS3UserResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "s3_user",
		},
	}
}

// ProtocolsS3UserResource defines the resource implementation.
type ProtocolsS3UserResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsS3UserResourceModel describes the resource data model.
type ProtocolsS3UserResourceModel struct {
	CxProfileName   types.String  `tfsdk:"cx_profile_name"`
	SVMName         types.String  `tfsdk:"svm_name"`
	Name            types.String  `tfsdk:"name"`
	Comment         types.String  `tfsdk:"comment"`
	AccessKey       types.String  `tfsdk:"access_key"`
	SecretKey	    types.String  `tfsdk:"secret_key"`
	KeyExpiryTime   types.String  `tfsdk:"key_expiry_time"`
	KeyTimeToLive   types.String  `tfsdk:"key_time_to_live"`
	RegenerateKeys  types.Bool    `tfsdk:"regenerate_keys"`
	DeleteKeys      types.Bool    `tfsdk:"delete_keys"`
	ID              types.Int64   `tfsdk:"id"`
}

// Metadata returns the resource type name
func (r *ProtocolsS3UserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsS3UserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsS3User resource",
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
				MarkdownDescription: "Specifies the name of the user.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Additional information about the user.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"access_key": schema.StringAttribute{
				MarkdownDescription: "Specifies the access key for the user.",
				Optional: 		     true,
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "Specifies the secret key for the user.",
				Optional: 		     true,
				Computed:            true,
				Sensitive:           true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_expiry_time": schema.StringAttribute{
				MarkdownDescription: "The date and time after which keys expire and are no longer valid.",
				Computed:            true,
			},
			"key_time_to_live": schema.StringAttribute{
				MarkdownDescription: `
				Indicates the time period from when this parameter is specified:
				when creating or modifying a user or when the user keys were last regenerated, after which the user keys expire and are no longer valid.
				Valid format is: 'PnDTnHnMnS|PnW'. For example, P2DT6H3M10S specifies a time period of 2 days, 6 hours, 3 minutes, and 10 seconds.
				If the value specified is '0' seconds, then the keys won't expire.
				It can be used during user creation (POST) and during keys regeneration (PATCH with regenerate_keys).
				`,
				Optional: 		     true,
			},
			"regenerate_keys": schema.BoolAttribute{
				MarkdownDescription: `
				Specifies if secret_key and access_key need to be regenerated.
				**Note:** This resource is not idempotent when this option is set.
				`,
				Optional:             true,
				WriteOnly:            true,
			},
			"delete_keys": schema.BoolAttribute{
				MarkdownDescription: "Specifies if secret_key and access_key need to be deleted.",
				Optional:             true,
				WriteOnly:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "The UUID of the S3 user.",
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ProtocolsS3UserResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// If this is a destroy plan, there is no planned state to modify.
	if req.Plan.Raw.IsNull() {
		return
	}

	var plan *ProtocolsS3UserResourceModel
	var config *ProtocolsS3UserResourceModel
	var state *ProtocolsS3UserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	// State is used to detect existing resources and ignore POST-only attributes.
	if !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// only force a diff when the user is explicitly requesting an action
	force := false
	changed := false
	if !config.RegenerateKeys.IsNull() && !config.RegenerateKeys.IsUnknown() && config.RegenerateKeys.ValueBool() {
		force = true
		// keep write-only attribute out of the planned state
		plan.RegenerateKeys = types.BoolNull()
		// keys, that are regenerated by the API, must not be
		// derived from prior state (UseStateForUnknown)
		plan.AccessKey = types.StringUnknown()
		plan.SecretKey = types.StringUnknown()
		// key regeneration also updates computed key expiry time
		plan.KeyExpiryTime = types.StringUnknown()
		plan.ID = types.Int64Unknown()
		changed = true
	}
	if !config.DeleteKeys.IsNull() && !config.DeleteKeys.IsUnknown() && config.DeleteKeys.ValueBool() {
		// keep write-only attribute out of the planned state
		plan.DeleteKeys = types.BoolNull()

		// only force an update if there are keys to delete
		keysPresent := false
		if state != nil {
			if !state.AccessKey.IsNull() && !state.AccessKey.IsUnknown() && strings.TrimSpace(state.AccessKey.ValueString()) != "" {
				keysPresent = true
			}
			if !state.SecretKey.IsNull() && !state.SecretKey.IsUnknown() && strings.TrimSpace(state.SecretKey.ValueString()) != "" {
				keysPresent = true
			}
		}
		if keysPresent {
			force = true
			// Keys will be deleted by the API, so avoid carrying forward prior state.
			plan.AccessKey = types.StringUnknown()
			plan.SecretKey = types.StringUnknown()
			plan.KeyExpiryTime = types.StringUnknown()
			plan.ID = types.Int64Unknown()
			changed = true
		}
	}

	// access_key/secret_key are POST-only
	// If the resource already exists, ignore any configured values to prevent perpetual diffs.
	if !force && state != nil {
		if (!config.AccessKey.IsNull() && !config.AccessKey.IsUnknown()) || (!config.SecretKey.IsNull() && !config.SecretKey.IsUnknown()) {
			plan.AccessKey = state.AccessKey
			plan.SecretKey = state.SecretKey
			changed = true
		}
	}

	if changed {
		resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsS3UserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Read refreshes the Terraform state with the latest data.
func (r *ProtocolsS3UserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsS3UserResourceModel

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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	var restInfo *interfaces.ProtocolsS3UserGetDataModelONTAP
	restInfo, err = interfaces.GetProtocolsS3User(errorHandler, *client, svm.UUID, data.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsS3UserbyID, GetProtocolsS3User
		return
	}

	data.ID = types.Int64Value(restInfo.ID)
	data.Name = types.StringValue(restInfo.Name)
	data.Comment = types.StringValue(restInfo.Comment)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	// The ONTAP GET response may omit access_key/secret_key when keys have been
	// deleted. In that case, reflect the deletion in state by setting them to
	// empty strings (rather than keeping prior state values).
	if restInfo.AccessKey == "" {
		data.AccessKey = types.StringValue("")
		data.SecretKey = types.StringValue("")
	} else {
		data.AccessKey = types.StringValue(restInfo.AccessKey)
	}
	// key_time_to_live is an input-style argument. Do not overwrite it from GET.
	// Only persist key_expiry_time when key_time_to_live is explicitly set in config/state.
	if !data.KeyTimeToLive.IsNull() && !data.KeyTimeToLive.IsUnknown() && strings.TrimSpace(data.KeyTimeToLive.ValueString()) != "" {
		if restInfo.KeyExpiryTime != "" {
			data.KeyExpiryTime = types.StringValue(restInfo.KeyExpiryTime)
		} else {
			data.KeyExpiryTime = types.StringNull()
		}
	} else {
		data.KeyExpiryTime = types.StringNull()
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// keep action-style arguments out of state so they can be re-applied
	// on subsequent runs when set in config
	data.RegenerateKeys = types.BoolNull()
	data.DeleteKeys = types.BoolNull()

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve ID
func (r *ProtocolsS3UserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProtocolsS3UserResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ProtocolsS3UserResourceBodyDataModel
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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	var errors []string

	nameValue := data.Name.ValueString()
	body.Name = &nameValue
	if !data.Comment.IsNull() {
		commentValue := data.Comment.ValueString()
		body.Comment = &commentValue
	}
	if !data.AccessKey.IsNull() && !data.AccessKey.IsUnknown() && data.AccessKey.ValueString() != "" {
		accessKeyValue := data.AccessKey.ValueString()
		body.AccessKey = &accessKeyValue
	}
	if !data.SecretKey.IsNull() && !data.SecretKey.IsUnknown() && data.SecretKey.ValueString() != "" {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 16 {
			secretKeyValue := data.SecretKey.ValueString()
			body.SecretKey = &secretKeyValue
		} else {
			errors = append(errors, "secret_key requires ONTAP 9.16 or later")
		}
	}
	if !data.KeyTimeToLive.IsNull() && !data.KeyTimeToLive.IsUnknown() && strings.TrimSpace(data.KeyTimeToLive.ValueString()) != "" {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 14 {
			ttlValue := data.KeyTimeToLive.ValueString()
			body.KeyTimeToLive = &ttlValue
		} else {
			errors = append(errors, "key_time_to_live requires ONTAP 9.14 or later")
		}
	}

	if len(errors) > 0 {
		errorsString := strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("The following options have ONTAP version constraints: %#v", errorsString))
		return
	}

	resource, err := interfaces.CreateProtocolsS3User(errorHandler, *client, body, svm.UUID)
	if err != nil {
		// error reporting done inside CreateProtocolsS3User
		return
	}
	data.ID = types.Int64Value(resource.ID)
	data.AccessKey = types.StringValue(resource.AccessKey)
	data.SecretKey = types.StringValue(resource.SecretKey)

	restInfo, err := interfaces.GetProtocolsS3User(errorHandler, *client, svm.UUID, data.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsS3UserbyID
		return
	}
	
	// Update the computed parameters
	data.Name = types.StringValue(restInfo.Name)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Comment = types.StringValue(restInfo.Comment)
	// key_time_to_live is an input-style argument. Keep the planned value in state.
	// Only persist key_expiry_time when key_time_to_live is explicitly set.
	if !data.KeyTimeToLive.IsNull() && !data.KeyTimeToLive.IsUnknown() && strings.TrimSpace(data.KeyTimeToLive.ValueString()) != "" {
		if restInfo.KeyExpiryTime != "" {
			data.KeyExpiryTime = types.StringValue(restInfo.KeyExpiryTime)
		} else {
			data.KeyExpiryTime = types.StringNull()
		}
	} else {
		data.KeyTimeToLive = types.StringNull()
		data.KeyExpiryTime = types.StringNull()
	}

	// keep action-style arguments out of state so they can be re-applied
	// on subsequent runs when set in config
	data.RegenerateKeys = types.BoolNull()
	data.DeleteKeys = types.BoolNull()

	tflog.Trace(ctx, fmt.Sprintf("created S3 user resource, ID=%d", data.ID.ValueInt64()))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsS3UserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state, config ProtocolsS3UserResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Read Terraform config data into the model (includes write-only attributes)
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	// Read Terraform state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
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
	var body interfaces.ProtocolsS3UserResourceBodyDataModel

	var errors []string

	body.Name = nil
	
	if !plan.Comment.Equal(state.Comment) {
		commentValue := plan.Comment.ValueString()
		body.Comment = &commentValue
	}
	
	if !config.RegenerateKeys.IsNull() && !config.RegenerateKeys.IsUnknown() && config.RegenerateKeys.ValueBool() {
		regenerateActionValue := true
		body.RegenerateKeys = &regenerateActionValue
		if !plan.KeyTimeToLive.IsNull() && !plan.KeyTimeToLive.IsUnknown() &&
		!plan.KeyTimeToLive.Equal(state.KeyTimeToLive) && plan.KeyTimeToLive.ValueString() != "" {
			if cluster.Version.Generation == 9 && cluster.Version.Major >= 14 {
				ttlValue := plan.KeyTimeToLive.ValueString()
				body.KeyTimeToLive = &ttlValue
			} else {
				errors = append(errors, "key_time_to_live requires ONTAP 9.14 or later")
			}
		}
	}
	if !config.DeleteKeys.IsNull() && !config.DeleteKeys.IsUnknown() && config.DeleteKeys.ValueBool() {
		// only call delete_keys when we actually have keys in state
		keysPresent := false
		if !state.AccessKey.IsNull() && !state.AccessKey.IsUnknown() && strings.TrimSpace(state.AccessKey.ValueString()) != "" {
			keysPresent = true
		}
		if !state.SecretKey.IsNull() && !state.SecretKey.IsUnknown() && strings.TrimSpace(state.SecretKey.ValueString()) != "" {
			keysPresent = true
		}
		if keysPresent {
			deleteActionValue := true
			body.DeleteKeys = &deleteActionValue
		}
	}

	if len(errors) > 0 {
		errorsString := strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("The following options have ONTAP version constraints: %#v", errorsString))
		return
	}

	resource, err := interfaces.UpdateProtocolsS3User(errorHandler, *client, body, svm.UUID, state.Name.ValueString())
	if err != nil {
		// error reporting done inside UpdateProtocolsS3User
		return
	}

	// PATCH may return no records for action-style operations.
	// Also note secret_key is not returned by GET, so prefer the PATCH response when it includes it.
	if resource != nil {
		if resource.AccessKey != "" {
			plan.AccessKey = types.StringValue(resource.AccessKey)
		}
		if resource.SecretKey != "" {
			plan.SecretKey = types.StringValue(resource.SecretKey)
		}
		// only persist key_expiry_time when key_time_to_live is explicitly set
		if !plan.KeyTimeToLive.IsNull() && !plan.KeyTimeToLive.IsUnknown() && strings.TrimSpace(plan.KeyTimeToLive.ValueString()) != "" {
			if resource.KeyExpiryTime != "" {
				plan.KeyExpiryTime = types.StringValue(resource.KeyExpiryTime)
			} else {
				plan.KeyExpiryTime = types.StringNull()
			}
		} else {
			plan.KeyExpiryTime = types.StringNull()
		}
	}

	restInfo, err := interfaces.GetProtocolsS3User(errorHandler, *client, svm.UUID, state.Name.ValueString(), cluster.Version)
	if err != nil {
		return
	}
	
	// Update the computed parameters
	plan.Name = types.StringValue(restInfo.Name)
	plan.SVMName = types.StringValue(restInfo.SVM.Name)
	plan.Comment = types.StringValue(restInfo.Comment)
	plan.ID = types.Int64Value(restInfo.ID)
	// ensure state reflects backend key deletion.
	if restInfo.AccessKey == "" {
		plan.AccessKey = types.StringValue("")
		plan.SecretKey = types.StringValue("")
	} else {
		plan.AccessKey = types.StringValue(restInfo.AccessKey)
	}
	// secret_key is not returned by GET. If it is still unknown here (e.g. regenerate
	// succeeded but ONTAP didn't return secret_key), keep prior state to avoid an invalid unknown value in state.
	if plan.SecretKey.IsUnknown() {
		plan.SecretKey = state.SecretKey
	}
	// key_time_to_live is an input-style argument. Do not overwrite it from GET.
	// Only persist key_expiry_time when key_time_to_live is explicitly set.
	if !plan.KeyTimeToLive.IsNull() && !plan.KeyTimeToLive.IsUnknown() && strings.TrimSpace(plan.KeyTimeToLive.ValueString()) != "" {
		if restInfo.KeyExpiryTime != "" {
			plan.KeyExpiryTime = types.StringValue(restInfo.KeyExpiryTime)
		} else {
			plan.KeyExpiryTime = types.StringNull()
		}
	} else {
		plan.KeyExpiryTime = types.StringNull()
	}

	// keep action-style arguments out of state so they can be re-applied
	// on subsequent runs when set in config.
	plan.RegenerateKeys = types.BoolNull()
	plan.DeleteKeys = types.BoolNull()

	tflog.Debug(ctx, fmt.Sprintf("updated S3 user resource: ID=%d", plan.ID.ValueInt64()))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsS3UserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data ProtocolsS3UserResourceModel

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

	if data.Name.IsNull() || data.Name.IsUnknown() || strings.TrimSpace(data.Name.ValueString()) == "" {
		errorHandler.MakeAndReportError("Name is null", "S3 user name is null")
		return
	}

	err = interfaces.DeleteProtocolsS3User(errorHandler, *client, svm.UUID, data.Name.ValueString())
	if err != nil {
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsS3UserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
