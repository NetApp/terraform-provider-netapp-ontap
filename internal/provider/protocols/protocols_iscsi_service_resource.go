package protocols

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsIscsiServiceResource{}
var _ resource.ResourceWithImportState = &ProtocolsIscsiServiceResource{}

// NewProtocolsIscsiServiceResource is a helper function to simplify the provider implementation.
func NewProtocolsIscsiServiceResource() resource.Resource {
	return &ProtocolsIscsiServiceResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "iscsi_service",
		},
	}
}

// ProtocolsIscsiServiceResource defines the resource implementation.
type ProtocolsIscsiServiceResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsIscsiServiceResourceModel describes the resource data model.
type ProtocolsIscsiServiceResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	// Protocols iSCSI Services specific
	Enabled types.Bool   `tfsdk:"enabled"`
	Target  types.Object `tfsdk:"target"`
	ID      types.String `tfsdk:"id"`
}

// ProtocolsIscsiServiceTargetResourceModel describes the resource model for iSCSI target.
type ProtocolsIscsiServiceTargetResourceModel struct {
	Alias types.String `tfsdk:"alias"`
	Name  types.String `tfsdk:"name"`
}

func (m ProtocolsIscsiServiceTargetResourceModel) attributeTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"alias": types.StringType,
		"name":  types.StringType,
	}
}

// Metadata returns the resource type name.
func (r *ProtocolsIscsiServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsIscsiServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Protocols iSCSI Service resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "iSCSI SVM name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "iSCSI should be enabled or disabled",
				Optional:            true,
				Computed:            true,
			},
			"target": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"alias": schema.StringAttribute{
						MarkdownDescription: "iSCSI target alias of the iSCSI service",
						Optional:            true,
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "iSCSI target name of the iSCSI service",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
				MarkdownDescription: "iSCSI target properties",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "iSCSI Service UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsIscsiServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

		return
	}

	r.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *ProtocolsIscsiServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsIscsiServiceResourceModel

	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Use existing-, or create new REST API client
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for reading protcols iscsi service info
	restInfo, err := interfaces.GetProtocolsIscsiService(
		errorHandler,
		*client,
		data.SVMName.ValueString(),
	)
	if err != nil {
		// error reporting done inside GetProtocolsIscsiService
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No iSCSI service found", "iSCSI service not found.")
		return
	}

	// Copy iSCSI service info to data source model
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.ID = types.StringValue(restInfo.SVM.UUID)

	targetElement := ProtocolsIscsiServiceTargetResourceModel{
		Alias: types.StringValue(restInfo.Target.Alias),
		Name:  types.StringValue(restInfo.Target.Name),
	}
	targetValue, diags := types.ObjectValueFrom(
		ctx,
		targetElement.attributeTypes(),
		targetElement,
	)
	if !diags.HasError() {
		data.Target = targetValue
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Create a resource and retrieve UUID
func (r *ProtocolsIscsiServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProtocolsIscsiServiceResourceModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Copy iscsi service info to request body
	var body interfaces.ProtocolsIscsiServiceResourceDataModelONTAP
	body.SVM.Name = data.SVMName.ValueString()
	// optional fields
	enabledEmpty := false
	if data.Enabled.IsNull() || data.Enabled.IsUnknown() {
		enabledEmpty = true
	} else {
		body.Enabled = data.Enabled.ValueBool()
	}
	var target ProtocolsIscsiServiceTargetResourceModel
	diags = data.Target.As(ctx, &target, basetypes.ObjectAsOptions{})
	if !diags.HasError() {
		body.Target.Alias = target.Alias.ValueString()
	}

	// Use existing-, or create new REST API client
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for creating iscsi service
	resource, err := interfaces.CreateProtocolsIscsiService(
		errorHandler,
		*client,
		body,
		enabledEmpty,
	)
	if err != nil {
		return
	}

	// Copy iscsi service info to resource model
	data.ID = types.StringValue(resource.SVM.UUID)
	data.Enabled = types.BoolValue(resource.Enabled)

	targetElement := ProtocolsIscsiServiceTargetResourceModel{
		Alias: types.StringValue(resource.Target.Alias),
		Name:  types.StringValue(resource.Target.Name),
	}
	targetValue, diags := types.ObjectValueFrom(
		ctx,
		targetElement.attributeTypes(),
		targetElement,
	)
	if !diags.HasError() {
		data.Target = targetValue
	}

	tflog.Trace(ctx, fmt.Sprintf("created a resource, UUID=%s", data.ID))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsIscsiServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ProtocolsIscsiServiceResourceModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Copy iscsi service info to request body
	var body interfaces.ProtocolsIscsiServiceResourceDataModelONTAP
	// optional fields
	enabledEmpty := false
	if data.Enabled.IsNull() || data.Enabled.IsUnknown() {
		enabledEmpty = true
	} else {
		body.Enabled = data.Enabled.ValueBool()
	}
	var target ProtocolsIscsiServiceTargetResourceModel
	diags = data.Target.As(ctx, &target, basetypes.ObjectAsOptions{})
	if !diags.HasError() {
		body.Target.Alias = target.Alias.ValueString()
	}

	// Use existing-, or create new REST API client
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for updating protcols iscsi service info
	err = interfaces.UpdateProtocolsIscsiService(
		errorHandler,
		*client,
		body,
		data.ID.ValueString(),
		enabledEmpty,
	)
	if err != nil {
		return
	}

	// Call ONTAP REST API for reading protcols iscsi service info
	restInfo, err := interfaces.GetProtocolsIscsiService(
		errorHandler,
		*client,
		data.SVMName.ValueString(),
	)
	if err != nil {
		// error reporting done inside GetProtocolsIscsiService
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No iSCSI service found", "iSCSI service not found.")
		return
	}

	// Copy iSCSI service info to data source model
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.ID = types.StringValue(restInfo.SVM.UUID)

	targetElement := ProtocolsIscsiServiceTargetResourceModel{
		Alias: types.StringValue(restInfo.Target.Alias),
		Name:  types.StringValue(restInfo.Target.Name),
	}
	targetValue, diags := types.ObjectValueFrom(
		ctx,
		targetElement.attributeTypes(),
		targetElement,
	)
	if !diags.HasError() {
		data.Target = targetValue
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a resource, UUID=%s", data.ID))

	// Save updated data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsIscsiServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsIscsiServiceResourceModel

	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Copy iscsi service info to request body (stop iscsi service)
	var body interfaces.ProtocolsIscsiServiceResourceDataModelONTAP
	body.Enabled = false

	// Use existing-, or create new REST API client
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for updating protcols iscsi service info
	err = interfaces.UpdateProtocolsIscsiService(
		errorHandler,
		*client,
		body,
		data.ID.ValueString(),
		false,
	)
	if err != nil {
		return
	}

	// Call ONTAP REST API for deleting iscsi service
	err = interfaces.DeleteProtocolsIscsiService(
		errorHandler,
		*client,
		data.ID.ValueString(),
	)
	if err != nil {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a resource, UUID=%s", data.ID))
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsIscsiServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	/*idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[1])...)*/
}
