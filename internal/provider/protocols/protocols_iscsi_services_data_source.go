package protocols

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsIscsiServicesDataSource{}

// NewProtocolsIscsiServicesDataSource is a helper function to simplify the provider implementation.
func NewProtocolsIscsiServicesDataSource() datasource.DataSource {
	return &ProtocolsIscsiServicesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "iscsi_services",
		},
	}
}

// ProtocolsIscsiServicesDataSource defines the data source implementation.
type ProtocolsIscsiServicesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsIscsiServicesDataSourceModel describes the data source data model.
type ProtocolsIscsiServicesDataSourceModel struct {
	CxProfileName          types.String                                `tfsdk:"cx_profile_name"`
	ProtocolsIscsiServices []ProtocolsIscsiServiceDataSourceModel      `tfsdk:"iscsi_services"`
	Filter                 *ProtocolsIscsiServiceDataSourceFilterModel `tfsdk:"filter"`
}

// ProtocolsIscsiServiceDataSourceFilterModel describes the data source data model for queries.
type ProtocolsIscsiServiceDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsIscsiServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsIscsiServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Protocols iSCSI Services data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "iSCSI SVM name",
						Optional:            true,
					},
				},
				MarkdownDescription: "Filter iSCSI services by their properties",
				Optional:            true,
			},
			"iscsi_services": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "iSCSI SVM name",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "iSCSI should be enabled or disabled",
							Computed:            true,
						},
						"target": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"alias": schema.StringAttribute{
									MarkdownDescription: "iSCSI target alias of the iSCSI service",
									Computed:            true,
								},
								"name": schema.StringAttribute{
									MarkdownDescription: "iSCSI target name of the iSCSI service",
									Computed:            true,
								},
							},
							MarkdownDescription: "iSCSI target properties",
							Computed:            true,
						},
					},
				},
				MarkdownDescription: "iSCSI services matching the filter",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsIscsiServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

		return
	}

	d.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *ProtocolsIscsiServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsIscsiServicesDataSourceModel

	// Read Terraform configuration data into the model
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Use existing-, or create new REST API client
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Prepare filter
	var filter *interfaces.IscsiServicesFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.IscsiServicesFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}

	// Call ONTAP REST API for reading protcols iscsi services info
	restInfo, err := interfaces.GetProtocolsIscsiServices(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetProtocolsIscsiServices
		return
	}

	// Copy iscsi services info to data source model
	data.ProtocolsIscsiServices = make([]ProtocolsIscsiServiceDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.ProtocolsIscsiServices[index] = ProtocolsIscsiServiceDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVMName:       types.StringValue(record.SVM.Name),
			Enabled:       types.BoolValue(record.Enabled),
			Target: &ProtocolsIscsiServiceTargetDataSourceModel{
				Alias: types.StringValue(record.Target.Alias),
				Name:  types.StringValue(record.Target.Name),
			},
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
