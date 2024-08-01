package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &StorageVolumeEfficiencyPoliciesResource{}
var _ resource.ResourceWithImportState = &StorageVolumeEfficiencyPoliciesResource{}

// NewStorageVolumeEfficiencyPoliciesResource is a helper function to simplify the provider implementation.
func NewStorageVolumeEfficiencyPoliciesResource() resource.Resource {
	return &StorageVolumeEfficiencyPoliciesResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "volume_efficiency_policies",
		},
	}
}

// StorageVolumeEfficiencyPoliciesResource defines the resource implementation.
type StorageVolumeEfficiencyPoliciesResource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageVolumeEfficiencyPoliciesResourceModel describes the resource data model.
type StorageVolumeEfficiencyPoliciesResourceModel struct {
	CxProfileName         types.String `tfsdk:"cx_profile_name"`
	Name                  types.String `tfsdk:"name"`
	SVM                   SVM          `tfsdk:"svm"`
	Type                  types.String `tfsdk:"type"`
	Schedule              types.Object `tfsdk:"schedule"`
	Duration              types.Int64  `tfsdk:"duration"`
	StartThresholdPercent types.Int64  `tfsdk:"start_threshold_percent"`
	QOSPolicy             types.String `tfsdk:"qos_policy"`
	Comment               types.String `tfsdk:"comment"`
	Enabled               types.Bool   `tfsdk:"enabled"`
	ID                    types.String `tfsdk:"id"`
}

// SVM describes SVM data model.
type SVM struct {
	Name types.String `tfsdk:"name"`
}

// Schedule describes Schedule data model.
type Schedule struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the resource type name.
func (r *StorageVolumeEfficiencyPoliciesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *StorageVolumeEfficiencyPoliciesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageVolumeEfficiencyPolicies resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "StorageVolumeEfficiencyPolicies name",
				Required:            true,
			},
			"svm": schema.SingleNestedAttribute{
				MarkdownDescription: "SVM details for StorageVolumeEfficiencyPolicies",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "name of the SVM",
						Required:            true,
					},
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "StorageVolumeEfficiencyPolicies type",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("scheduled"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule": schema.SingleNestedAttribute{
				MarkdownDescription: "schedule details for StorageVolumeEfficiencyPolicies",
				Optional:            true,
				Computed:            true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"name": types.StringType,
					},
					map[string]attr.Value{
						"name": types.StringValue("daily"),
					})),
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "name of the schedule",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("daily"),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"duration": schema.Int64Attribute{
				MarkdownDescription: "StorageVolumeEfficiencyPolicies duration",
				Optional:            true,
			},
			"start_threshold_percent": schema.Int64Attribute{
				MarkdownDescription: "StorageVolumeEfficiencyPolicies duration",
				Optional:            true,
			},
			"qos_policy": schema.StringAttribute{
				MarkdownDescription: "StorageVolumeEfficiencyPolicies duration",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("best_effort"),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "StorageVolumeEfficiencyPolicies duration",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "StorageVolumeEfficiencyPolicies duration",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "StorageVolumeEfficiencyPolicies UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *StorageVolumeEfficiencyPoliciesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *StorageVolumeEfficiencyPoliciesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StorageVolumeEfficiencyPoliciesResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	var restInfo *interfaces.StorageVolumeEfficiencyPoliciesGetDataModelONTAP
	if data.ID.ValueString() != "" {
		restInfo, err = interfaces.GetStorageVolumeEfficiencyPoliciesByUUID(errorHandler, *client, data.ID.ValueString())
		if err != nil {
			// error reporting done inside GetStorageVolumeEfficiencyPoliciesByUUID
			return
		}
	} else {
		restInfo, err = interfaces.GetStorageVolumeEfficiencyPoliciesByName(errorHandler, *client, data.Name.ValueString(), data.SVM.Name.ValueString())
		if err != nil {
			// error reporting done inside GetStorageVolumeEfficiencyPoliciesByName
			return
		}
	}

	if restInfo == nil {
		errorHandler.MakeAndReportError("error reading info", "No Storage Volume Efficiency Policy found")
		return
	}

	data.ID = types.StringValue(restInfo.UUID)
	data.Name = types.StringValue(restInfo.Name)
	data.SVM.Name = types.StringValue(restInfo.SVM.Name)
	data.Type = types.StringValue(restInfo.Type)
	data.QOSPolicy = types.StringValue(restInfo.QOSPolicy)
	data.Enabled = types.BoolValue(restInfo.Enabled)
	elementTypes := map[string]attr.Type{
		"name": types.StringType,
	}
	elements := map[string]attr.Value{
		"name": types.StringValue(restInfo.Schedule.Name),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Schedule = objectValue

	if restInfo.Duration != types.Int64Null().ValueInt64() {
		data.Duration = types.Int64Value(restInfo.Duration)
	}
	if restInfo.StartThresholdPercent != types.Int64Null().ValueInt64() {
		data.StartThresholdPercent = types.Int64Value(restInfo.StartThresholdPercent)
	}
	if restInfo.Comment != "" {
		data.Comment = types.StringValue(restInfo.Comment)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *StorageVolumeEfficiencyPoliciesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageVolumeEfficiencyPoliciesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.StorageVolumeEfficiencyPoliciesResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVM.Name.ValueString()
	body.Type = data.Type.ValueString()
	if !data.Schedule.IsNull() {
		var schedule Schedule
		diags := data.Schedule.As(ctx, &schedule, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.Schedule.Name = schedule.Name.ValueString()
	}
	if !data.Duration.IsNull() {
		body.Duration = data.Duration.ValueInt64()
	}
	if !data.StartThresholdPercent.IsNull() {
		body.StartThresholdPercent = data.StartThresholdPercent.ValueInt64()
	}
	body.QOSPolicy = data.QOSPolicy.ValueString()
	if !data.Comment.IsNull() {
		body.Comment = data.Comment.ValueString()
	}
	body.Enabled = data.Enabled.ValueBool()

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateStorageVolumeEfficiencyPolicies(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.ID = types.StringValue(resource.UUID)
	data.Comment = types.StringValue(resource.Comment)
	// elementTypes := map[string]attr.Type{
	// 	"name": types.StringType,
	// }
	// elements := map[string]attr.Value{
	// 	"name": types.StringValue(resource.Schedule.Name),
	// }
	// objectValue, diags := types.ObjectValue(elementTypes, elements)
	// if diags.HasError() {
	// 	resp.Diagnostics.Append(diags...)
	// }
	// data.Schedule = objectValue

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *StorageVolumeEfficiencyPoliciesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *StorageVolumeEfficiencyPoliciesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	client, err := connection.GetRestClient(utils.NewErrorHandler(ctx, &resp.Diagnostics), r.config, plan.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	var request interfaces.UpdateStorageVolumeEfficiencyPoliciesResourceBodyDataModelONTAP
	if plan.StartThresholdPercent != state.StartThresholdPercent {
		request.StartThresholdPercent = plan.StartThresholdPercent.ValueInt64()
	}
	if plan.Duration != state.Duration {
		request.Duration = plan.Duration.ValueInt64()
	}
	if !plan.Comment.Equal(state.Comment) {
		request.Comment = plan.Comment.ValueString()
	}
	if !plan.Schedule.Equal(state.Schedule) {
		var schedule Schedule
		diags := plan.Schedule.As(ctx, &schedule, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		request.Schedule.Name = schedule.Name.ValueString()
	}
	if !plan.QOSPolicy.Equal(state.QOSPolicy) {
		request.QOSPolicy = plan.QOSPolicy.ValueString()
	}
	if plan.Enabled.Equal(state.Enabled) {
		request.Enabled = plan.Enabled.ValueBool()
	}
	if !plan.Type.Equal(state.Type) {
		request.Type = plan.Type.ValueString()
	}
	err = interfaces.UpdateStorageVolumeEfficiencyPolicies(errorHandler, *client, state.ID.ValueString(), request)
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *StorageVolumeEfficiencyPoliciesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *StorageVolumeEfficiencyPoliciesResourceModel

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

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "storage_volume_efficiency_policies UUID is null")
		return
	}

	err = interfaces.DeleteStorageVolumeEfficiencyPolicies(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *StorageVolumeEfficiencyPoliciesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req an Storage Volume Efficiency Policies resource: %#v", req))
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprint("Expected ID in the format 'name,svm_name,cx_profile_name', got: ", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm").AtName("name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
}
