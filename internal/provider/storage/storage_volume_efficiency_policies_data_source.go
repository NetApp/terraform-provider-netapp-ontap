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
var _ datasource.DataSource = &VolumeEfficiencyPoliciesDataSource{}

// NewVolumeEfficiencyPoliciesDataSource is a helper function to simplify the provider implementation.
func NewVolumeEfficiencyPoliciesDataSource() datasource.DataSource {
	return &VolumeEfficiencyPoliciesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "volume_efficiency_policies",
		},
	}
}

// VolumeEfficiencyPoliciesDataSource defines the data source implementation.
type VolumeEfficiencyPoliciesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// VolumeEfficiencyPoliciesDataSourceModel describes the data source data model.
type VolumeEfficiencyPoliciesDataSourceModel struct {
	CxProfileName            types.String                                   `tfsdk:"cx_profile_name"`
	VolumeEfficiencyPolicies []VolumeEfficiencyPolicyDataSourceModel        `tfsdk:"volume_efficiency_policies"`
	Filter                   *VolumeEfficiencyPoliciesDataSourceFilterModel `tfsdk:"filter"`
}

// VolumeEfficiencyPoliciesDataSourceFilterModel describes the data source data model for queries.
type VolumeEfficiencyPoliciesDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *VolumeEfficiencyPoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *VolumeEfficiencyPoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "VolumeEfficiencyPolicies data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "VolumeEfficiencyPolicy name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "VolumeEfficiencyPolicy svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"volume_efficiency_policies": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "VolumeEfficiencyPolicy name",
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
							Computed:            true,
						},
						"schedule": schema.SingleNestedAttribute{
							MarkdownDescription: "schedule details for StorageVolumeEfficiencyPolicies",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "name of the schedule",
									Computed:            true,
								},
							},
						},
						"duration": schema.Int64Attribute{
							MarkdownDescription: "StorageVolumeEfficiencyPolicies duration",
							Computed:            true,
						},
						"start_threshold_percent": schema.Int64Attribute{
							MarkdownDescription: "StorageVolumeEfficiencyPolicies duration",
							Computed:            true,
						},
						"qos_policy": schema.StringAttribute{
							MarkdownDescription: "StorageVolumeEfficiencyPolicies duration",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "StorageVolumeEfficiencyPolicies duration",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "StorageVolumeEfficiencyPolicies duration",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "StorageVolumeEfficiencyPolicies UUID",
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
func (d *VolumeEfficiencyPoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *VolumeEfficiencyPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VolumeEfficiencyPoliciesDataSourceModel

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

	var filter *interfaces.StorageVolumeEfficiencyPoliciesDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.StorageVolumeEfficiencyPoliciesDataSourceFilterModel{
			Name:    data.Filter.Name.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetStorageVolumeEfficiencyPolicies(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetVolumeEfficiencyPolicies
		return
	}

	data.VolumeEfficiencyPolicies = make([]VolumeEfficiencyPolicyDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.VolumeEfficiencyPolicies[index] = VolumeEfficiencyPolicyDataSourceModel{}
		data.VolumeEfficiencyPolicies[index].CxProfileName = types.String(data.CxProfileName)
		data.VolumeEfficiencyPolicies[index].Name = types.StringValue(record.Name)
		data.VolumeEfficiencyPolicies[index].SVM.Name = types.StringValue(record.SVM.Name)
		data.VolumeEfficiencyPolicies[index].Type = types.StringValue(record.Type)
		data.VolumeEfficiencyPolicies[index].QOSPolicy = types.StringValue(record.QOSPolicy)
		data.VolumeEfficiencyPolicies[index].Enabled = types.BoolValue(record.Enabled)
		elementTypes := map[string]attr.Type{
			"name": types.StringType,
		}
		elements := map[string]attr.Value{
			"name": types.StringValue(record.Schedule.Name),
		}
		objectValue, diags := types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
		data.VolumeEfficiencyPolicies[index].Schedule = objectValue
		if record.Duration != types.Int64Null().ValueInt64() {
			data.VolumeEfficiencyPolicies[index].Duration = types.Int64Value(record.Duration)
		}
		if record.StartThresholdPercent != types.Int64Null().ValueInt64() {
			data.VolumeEfficiencyPolicies[index].StartThresholdPercent = types.Int64Value(record.StartThresholdPercent)
		}
		if record.Comment != "" {
			data.VolumeEfficiencyPolicies[index].Comment = types.StringValue(record.Comment)
		}

	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
