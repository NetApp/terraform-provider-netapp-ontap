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
var _ datasource.DataSource = &VolumeEfficiencyPolicyDataSource{}

// NewVolumeEfficiencyPolicyDataSource is a helper function to simplify the provider implementation.
func NewVolumeEfficiencyPolicyDataSource() datasource.DataSource {
	return &VolumeEfficiencyPolicyDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "volume_efficiency_policy",
		},
	}
}

// VolumeEfficiencyPolicyDataSource defines the data source implementation.
type VolumeEfficiencyPolicyDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// VolumeEfficiencyPolicyDataSourceModel describes the data source data model.
type VolumeEfficiencyPolicyDataSourceModel struct {
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

// Metadata returns the data source type name.
func (d *VolumeEfficiencyPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *VolumeEfficiencyPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "VolumeEfficiencyPolicy data source",

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
	}
}

// Configure adds the provider configured client to the data source.
func (d *VolumeEfficiencyPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *VolumeEfficiencyPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VolumeEfficiencyPolicyDataSourceModel

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

	restInfo, err := interfaces.GetStorageVolumeEfficiencyPoliciesByName(errorHandler, *client, data.Name.ValueString(), data.SVM.Name.ValueString())
	if err != nil {
		// error reporting done inside GetVolumeEfficiencyPolicy
		return
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
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
