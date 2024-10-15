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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
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
var _ resource.Resource = &QOSPoliciesResource{}
var _ resource.ResourceWithImportState = &QOSPoliciesResource{}

// NewQOSPoliciesResource is a helper function to simplify the provider implementation.
func NewQOSPolicyResource() resource.Resource {
	return &QOSPoliciesResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "qos_policy",
		},
	}
}

// QOSPoliciesResource defines the resource implementation.
type QOSPoliciesResource struct {
	config connection.ResourceOrDataSourceConfig
}

// QOSPoliciesResourceModel describes the resource data model.
type QOSPoliciesResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Fixed         types.Object `tfsdk:"fixed"`
	Adaptive      types.Object `tfsdk:"adaptive"`
	Scope         types.String `tfsdk:"scope"`
	ID            types.String `tfsdk:"id"`
}

// fixed describes the data model using go types for mapping.
type fixed struct {
	MaxThroughputIOPS types.Int64 `tfsdk:"max_throughput_iops"`
	CapacityShared    types.Bool  `tfsdk:"capacity_shared"`
	MaxThroughputMBPS types.Int64 `tfsdk:"max_throughput_mbps"`
	MinThroughputIOPS types.Int64 `tfsdk:"min_throughput_iops"`
	MinThroughputMBPS types.Int64 `tfsdk:"min_throughput_mbps"`
}

// adaptive describes the data model using go types for mapping.
type adaptive struct {
	ExpectedIOPSAllocation types.String `tfsdk:"expected_iops_allocation"`
	ExpectedIOPS           types.Int64  `tfsdk:"expected_iops"`
	PeakIOPSAllocation     types.String `tfsdk:"peak_iops_allocation"`
	BlockSize              types.String `tfsdk:"block_size"`
	PeakIOPS               types.Int64  `tfsdk:"peak_iops"`
	AbsoluteMinIOPS        types.Int64  `tfsdk:"absolute_min_iops"`
}

// Metadata returns the resource type name.
func (r *QOSPoliciesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *QOSPoliciesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "QOSPolicies resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "QOSPolicies name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "QOSPolicies svm name",
				Required:            true,
			},
			"fixed": schema.SingleNestedAttribute{
				MarkdownDescription: "Fixed QoS policy",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"max_throughput_iops": schema.Int64Attribute{
						MarkdownDescription: "Maximum throughput in IOPS",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"capacity_shared": schema.BoolAttribute{
						MarkdownDescription: "Capacity shared",
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"max_throughput_mbps": schema.Int64Attribute{
						MarkdownDescription: "Maximum throughput in MBPS",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"min_throughput_iops": schema.Int64Attribute{
						MarkdownDescription: "Minimum throughput in IOPS",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"min_throughput_mbps": schema.Int64Attribute{
						MarkdownDescription: "Minimum throughput in MBPS",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"adaptive": schema.SingleNestedAttribute{
				MarkdownDescription: "Adaptive QoS policy",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				Attributes: map[string]schema.Attribute{
					"expected_iops_allocation": schema.StringAttribute{
						MarkdownDescription: "Expected IOPS allocation",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("allocated_space"),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"expected_iops": schema.Int64Attribute{
						MarkdownDescription: "Expected IOPS",
						Required:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"peak_iops_allocation": schema.StringAttribute{
						MarkdownDescription: "Peak IOPS allocation",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("used_space"),
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"block_size": schema.StringAttribute{
						MarkdownDescription: "Block size",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"peak_iops": schema.Int64Attribute{
						MarkdownDescription: "Peak IOPS",
						Required:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"absolute_min_iops": schema.Int64Attribute{
						MarkdownDescription: "Absolute minimum IOPS",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "QoS policy scope",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "QOSPolicies UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *QOSPoliciesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *QOSPoliciesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data QOSPoliciesResourceModel

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

	var restInfo *interfaces.QOSPoliciesGetDataModelONTAP
	if data.ID.ValueString() != "" {
		restInfo, err = interfaces.GetQOSPoliciesByUUID(errorHandler, *client, data.ID.ValueString())
		if err != nil {
			// error reporting done inside GetQOSPolicies
			return
		}
	} else {
		restInfo, err = interfaces.GetQOSPoliciesByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
		if err != nil {
			// error reporting done inside GetQOSPolicies
			return
		}
	}

	if restInfo == nil {
		errorHandler.MakeAndReportError("error reading info", "No QOS policy found")
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Scope = types.StringValue(restInfo.Scope)
	data.ID = types.StringValue(restInfo.UUID)

	// Fixed QoS policy
	elementTypes := map[string]attr.Type{
		"max_throughput_iops": types.Int64Type,
		"capacity_shared":     types.BoolType,
		"max_throughput_mbps": types.Int64Type,
		"min_throughput_iops": types.Int64Type,
		"min_throughput_mbps": types.Int64Type,
	}
	elements := map[string]attr.Value{
		"max_throughput_iops": types.Int64Value(int64(restInfo.Fixed.MaxThroughputIOPS)),
		"capacity_shared":     types.BoolValue(restInfo.Fixed.CapacityShared),
		"max_throughput_mbps": types.Int64Value(int64(restInfo.Fixed.MaxThroughputMBPS)),
		"min_throughput_iops": types.Int64Value(int64(restInfo.Fixed.MinThroughputIOPS)),
		"min_throughput_mbps": types.Int64Value(int64(restInfo.Fixed.MinThroughputMBPS)),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Fixed = objectValue

	// Adaptive QoS policy
	elementTypes = map[string]attr.Type{
		"expected_iops_allocation": types.StringType,
		"expected_iops":            types.Int64Type,
		"peak_iops_allocation":     types.StringType,
		"block_size":               types.StringType,
		"peak_iops":                types.Int64Type,
		"absolute_min_iops":        types.Int64Type,
	}
	elements = map[string]attr.Value{
		"expected_iops_allocation": types.StringValue(restInfo.Adaptive.ExpectedIOPSAllocation),
		"expected_iops":            types.Int64Value(int64(restInfo.Adaptive.ExpectedIOPS)),
		"peak_iops_allocation":     types.StringValue(restInfo.Adaptive.PeakIOPSAllocation),
		"block_size":               types.StringValue(restInfo.Adaptive.BlockSize),
		"peak_iops":                types.Int64Value(int64(restInfo.Adaptive.PeakIOPS)),
		"absolute_min_iops":        types.Int64Value(int64(restInfo.Adaptive.AbsoluteMinIOPS)),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Adaptive = objectValue

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *QOSPoliciesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *QOSPoliciesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.QOSPoliciesResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()
	if !data.Scope.IsUnknown() {
		body.Scope = data.Scope.ValueString()
	}

	if !data.Fixed.IsUnknown() {
		var fixed fixed
		diags := data.Fixed.As(ctx, &fixed, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.Fixed.CapacityShared = fixed.CapacityShared.ValueBool()
		if !fixed.MaxThroughputIOPS.IsNull() {
			body.Fixed.MaxThroughputIOPS = int(fixed.MaxThroughputIOPS.ValueInt64())
		}
		if !fixed.MaxThroughputMBPS.IsNull() {
			body.Fixed.MaxThroughputMBPS = int(fixed.MaxThroughputMBPS.ValueInt64())
		}
		if !fixed.MinThroughputIOPS.IsNull() {
			body.Fixed.MinThroughputIOPS = int(fixed.MinThroughputIOPS.ValueInt64())
		}
		if !fixed.MinThroughputMBPS.IsNull() {
			body.Fixed.MinThroughputMBPS = int(fixed.MinThroughputMBPS.ValueInt64())
		}
	}

	if !data.Adaptive.IsUnknown() {
		var adaptive adaptive
		diags := data.Adaptive.As(ctx, &adaptive, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.Adaptive.ExpectedIOPSAllocation = adaptive.ExpectedIOPSAllocation.ValueString()
		body.Adaptive.PeakIOPSAllocation = adaptive.PeakIOPSAllocation.ValueString()
		if !adaptive.ExpectedIOPS.IsNull() {
			body.Adaptive.ExpectedIOPS = int(adaptive.ExpectedIOPS.ValueInt64())
		}
		if !adaptive.BlockSize.IsNull() {
			body.Adaptive.BlockSize = adaptive.BlockSize.ValueString()
		}
		if !adaptive.PeakIOPS.IsNull() {
			body.Adaptive.PeakIOPS = int(adaptive.PeakIOPS.ValueInt64())
		}
		if !adaptive.AbsoluteMinIOPS.IsNull() {
			body.Adaptive.AbsoluteMinIOPS = int(adaptive.AbsoluteMinIOPS.ValueInt64())
		}
	}

	if data.Fixed.IsUnknown() && data.Adaptive.IsUnknown() {
		errorHandler.MakeAndReportError("Fixed and Adaptive QoS policies are both empty, one of Fixed or Adaptive QoS policies are required", "One of Fixed or Adaptive QoS policies are required")
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateQOSPolicies(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.ID = types.StringValue(resource.UUID)
	var restInfo *interfaces.QOSPoliciesGetDataModelONTAP
	restInfo, err = interfaces.GetQOSPoliciesByUUID(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		// error reporting done inside GetQOSPolicies
		return
	}

	if restInfo == nil {
		errorHandler.MakeAndReportError("error reading info", "No QOS policy found")
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Scope = types.StringValue(restInfo.Scope)
	data.ID = types.StringValue(restInfo.UUID)

	// Fixed QoS policy
	elementTypes := map[string]attr.Type{
		"max_throughput_iops": types.Int64Type,
		"capacity_shared":     types.BoolType,
		"max_throughput_mbps": types.Int64Type,
		"min_throughput_iops": types.Int64Type,
		"min_throughput_mbps": types.Int64Type,
	}
	elements := map[string]attr.Value{
		"max_throughput_iops": types.Int64Value(int64(restInfo.Fixed.MaxThroughputIOPS)),
		"capacity_shared":     types.BoolValue(restInfo.Fixed.CapacityShared),
		"max_throughput_mbps": types.Int64Value(int64(restInfo.Fixed.MaxThroughputMBPS)),
		"min_throughput_iops": types.Int64Value(int64(restInfo.Fixed.MinThroughputIOPS)),
		"min_throughput_mbps": types.Int64Value(int64(restInfo.Fixed.MinThroughputMBPS)),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Fixed = objectValue

	// Adaptive QoS policy
	elementTypes = map[string]attr.Type{
		"expected_iops_allocation": types.StringType,
		"expected_iops":            types.Int64Type,
		"peak_iops_allocation":     types.StringType,
		"block_size":               types.StringType,
		"peak_iops":                types.Int64Type,
		"absolute_min_iops":        types.Int64Type,
	}
	elements = map[string]attr.Value{
		"expected_iops_allocation": types.StringValue(restInfo.Adaptive.ExpectedIOPSAllocation),
		"expected_iops":            types.Int64Value(int64(restInfo.Adaptive.ExpectedIOPS)),
		"peak_iops_allocation":     types.StringValue(restInfo.Adaptive.PeakIOPSAllocation),
		"block_size":               types.StringValue(restInfo.Adaptive.BlockSize),
		"peak_iops":                types.Int64Value(int64(restInfo.Adaptive.PeakIOPS)),
		"absolute_min_iops":        types.Int64Value(int64(restInfo.Adaptive.AbsoluteMinIOPS)),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Adaptive = objectValue

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *QOSPoliciesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *QOSPoliciesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	var request interfaces.QOSPoliciesUpdateResourceBodyDataModelONTAP

	if !plan.Fixed.IsUnknown() {
		if !plan.Fixed.Equal(state.Fixed) {
			var fixed fixed
			diags := plan.Fixed.As(ctx, &fixed, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			request.Fixed.CapacityShared = fixed.CapacityShared.ValueBool()
			if !fixed.MaxThroughputIOPS.IsUnknown() {
				request.Fixed.MaxThroughputIOPS = int(fixed.MaxThroughputIOPS.ValueInt64())
			}
			if !fixed.MaxThroughputMBPS.IsUnknown() {
				request.Fixed.MaxThroughputMBPS = int(fixed.MaxThroughputMBPS.ValueInt64())
			}
			if !fixed.MinThroughputIOPS.IsUnknown() {
				request.Fixed.MinThroughputIOPS = int(fixed.MinThroughputIOPS.ValueInt64())
			}
			if !fixed.MinThroughputMBPS.IsUnknown() {
				request.Fixed.MinThroughputMBPS = int(fixed.MinThroughputMBPS.ValueInt64())
			}
		}
	}

	if !plan.Adaptive.IsUnknown() {
		if !plan.Adaptive.Equal(state.Adaptive) {
			var adaptive adaptive
			diags := plan.Adaptive.As(ctx, &adaptive, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			request.Adaptive.ExpectedIOPSAllocation = adaptive.ExpectedIOPSAllocation.ValueString()
			request.Adaptive.PeakIOPSAllocation = adaptive.PeakIOPSAllocation.ValueString()
			if !adaptive.ExpectedIOPS.IsUnknown() {
				request.Adaptive.ExpectedIOPS = int(adaptive.ExpectedIOPS.ValueInt64())
			}
			if !adaptive.BlockSize.IsUnknown() {
				request.Adaptive.BlockSize = adaptive.BlockSize.ValueString()
			}
			if !adaptive.PeakIOPS.IsUnknown() {
				request.Adaptive.PeakIOPS = int(adaptive.PeakIOPS.ValueInt64())
			}
			if !adaptive.AbsoluteMinIOPS.IsUnknown() {
				request.Adaptive.AbsoluteMinIOPS = int(adaptive.AbsoluteMinIOPS.ValueInt64())
			}
		}
	}

	if !plan.Name.Equal(state.Name) {
		request.Name = plan.Name.ValueString()
	}

	err = interfaces.UpdateQOSPolicies(errorHandler, *client, request, plan.ID.ValueString())
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *QOSPoliciesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *QOSPoliciesResourceModel

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
		errorHandler.MakeAndReportError("UUID is null", "qos_policies UUID is null")
		return
	}

	err = interfaces.DeleteQOSPolicies(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *QOSPoliciesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
