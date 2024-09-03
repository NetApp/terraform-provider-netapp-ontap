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
var _ datasource.DataSource = &QOSPoliciesDataSource{}

// NewStorageQOSPoliciesDataSource is a helper function to simplify the provider implementation.
func NewStorageQOSPoliciesDataSource() datasource.DataSource {
	return &QOSPoliciesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "qos_policies",
		},
	}
}

// QOSPoliciesDataSource defines the data source implementation.
type QOSPoliciesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// QOSPoliciesDataSourceModel describes the data source data model.
type QOSPoliciesDataSourceModel struct {
	CxProfileName types.String                      `tfsdk:"cx_profile_name"`
	QOSPolicies   []QOSPolicyDataSourceModel        `tfsdk:"qos_policies"`
	Filter        *QOSPoliciesDataSourceFilterModel `tfsdk:"filter"`
}

// QOSPoliciesDataSourceFilterModel describes the data source data model for queries.
type QOSPoliciesDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *QOSPoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *QOSPoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "QOSPolicies data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "QOSPolicy name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "QOSPolicy svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"qos_policies": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
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
				},
				Computed:            true,
				MarkdownDescription: "",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *QOSPoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *QOSPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data QOSPoliciesDataSourceModel

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

	var filter *interfaces.QOSPoliciesDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.QOSPoliciesDataSourceFilterModel{
			Name:    data.Filter.Name.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetQOSPolicies(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetQOSPolicies
		return
	}

	data.QOSPolicies = make([]QOSPolicyDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.QOSPolicies[index] = QOSPolicyDataSourceModel{}
		data.QOSPolicies[index].CxProfileName = types.String(data.CxProfileName)
		data.QOSPolicies[index].Name = types.StringValue(record.Name)
		data.QOSPolicies[index].Scope = types.StringValue(record.Scope)
		data.QOSPolicies[index].ID = types.StringValue(record.UUID)

		// Fixed QoS policy
		elementTypes := map[string]attr.Type{
			"max_throughput_iops": types.Int64Type,
			"capacity_shared":     types.BoolType,
			"max_throughput_mbps": types.Int64Type,
			"min_throughput_iops": types.Int64Type,
			"min_throughput_mbps": types.Int64Type,
		}
		elements := map[string]attr.Value{
			"max_throughput_iops": types.Int64Value(int64(record.Fixed.MaxThroughputIOPS)),
			"capacity_shared":     types.BoolValue(record.Fixed.CapacityShared),
			"max_throughput_mbps": types.Int64Value(int64(record.Fixed.MaxThroughputMBPS)),
			"min_throughput_iops": types.Int64Value(int64(record.Fixed.MinThroughputIOPS)),
			"min_throughput_mbps": types.Int64Value(int64(record.Fixed.MinThroughputMBPS)),
		}
		objectValue, diags := types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
		data.QOSPolicies[index].Fixed = objectValue

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
			"expected_iops_allocation": types.StringValue(record.Adaptive.ExpectedIOPSAllocation),
			"expected_iops":            types.Int64Value(int64(record.Adaptive.ExpectedIOPS)),
			"peak_iops_allocation":     types.StringValue(record.Adaptive.PeakIOPSAllocation),
			"block_size":               types.StringValue(record.Adaptive.BlockSize),
			"peak_iops":                types.Int64Value(int64(record.Adaptive.PeakIOPS)),
			"absolute_min_iops":        types.Int64Value(int64(record.Adaptive.AbsoluteMinIOPS)),
		}
		objectValue, diags = types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
		data.QOSPolicies[index].Adaptive = objectValue
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
