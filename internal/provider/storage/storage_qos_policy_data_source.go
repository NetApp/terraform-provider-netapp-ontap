package storage

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &QOSPolicyDataSource{}

// NewQOSPolicyDataSource is a helper function to simplify the provider implementation.
func NewStorageQOSPolicyDataSource() datasource.DataSource {
	return &QOSPolicyDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "qos_policy",
		},
	}
}

// QOSPolicyDataSource defines the data source implementation.
type QOSPolicyDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// QOSPolicyDataSourceModel describes the data source data model.
type QOSPolicyDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Fixed         types.Object `tfsdk:"fixed"`
	Adaptive      types.Object `tfsdk:"adaptive"`
	Scope         types.String `tfsdk:"scope"`
	ID            types.String `tfsdk:"id"`
}

// Metadata returns the data source type name.
func (d *QOSPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *QOSPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "QOSPolicy data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "QOSPolicy name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "QOSPolicy svm name",
				Required:            true,
			},
			"fixed": schema.SingleNestedAttribute{
				MarkdownDescription: "Fixed QoS policy",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"max_throughput_iops": schema.Int64Attribute{
						MarkdownDescription: "Maximum throughput in IOPS",
						Computed:            true,
					},
					"capacity_shared": schema.BoolAttribute{
						MarkdownDescription: "Capacity shared",
						Computed:            true,
					},
					"max_throughput_mbps": schema.Int64Attribute{
						MarkdownDescription: "Maximum throughput in MBPS",
						Computed:            true,
					},
					"min_throughput_iops": schema.Int64Attribute{
						MarkdownDescription: "Minimum throughput in IOPS",
						Computed:            true,
					},
					"min_throughput_mbps": schema.Int64Attribute{
						MarkdownDescription: "Minimum throughput in MBPS",
						Computed:            true,
					},
				},
			},
			"adaptive": schema.SingleNestedAttribute{
				MarkdownDescription: "Adaptive QoS policy",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"expected_iops_allocation": schema.StringAttribute{
						MarkdownDescription: "Expected IOPS allocation",
						Computed:            true,
					},
					"expected_iops": schema.Int64Attribute{
						MarkdownDescription: "Expected IOPS",
						Computed:            true,
					},
					"peak_iops_allocation": schema.StringAttribute{
						MarkdownDescription: "Peak IOPS allocation",
						Computed:            true,
					},
					"block_size": schema.StringAttribute{
						MarkdownDescription: "Block size",
						Computed:            true,
					},
					"peak_iops": schema.Int64Attribute{
						MarkdownDescription: "Peak IOPS",
						Required:            true,
					},
					"absolute_min_iops": schema.Int64Attribute{
						MarkdownDescription: "Absolute minimum IOPS",
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "QoS policy scope",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "QOSPolicies UUID",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *QOSPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	d.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *QOSPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data QOSPolicyDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// var restInfo *interfaces.QOSPoliciesGetDataModelONTAP
	restInfo, err := interfaces.GetQOSPoliciesByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
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

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
