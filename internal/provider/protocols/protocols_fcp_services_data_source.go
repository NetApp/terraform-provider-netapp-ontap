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
var _ datasource.DataSource = &ProtocolsFcpServicesDataSource{}

// NewProtocolsFcpServicesDataSource is a helper function to simplify the provider implementation.
func NewProtocolsFcpServicesDataSource() datasource.DataSource {
	return &ProtocolsFcpServicesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "fcp_services",
		},
	}
}

// ProtocolsFcpServicesDataSource defines the data source implementation.
type ProtocolsFcpServicesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsFcpServicesDataSourceModel describes the data source data model.
type ProtocolsFcpServicesDataSourceModel struct {
	CxProfileName        types.String                              `tfsdk:"cx_profile_name"`
	ProtocolsFcpServices []ProtocolsFcpServiceDataSourceModel      `tfsdk:"fcp_services"`
	Filter               *ProtocolsFcpServiceDataSourceFilterModel `tfsdk:"filter"`
}

// ProtocolsFcpServiceDataSourceFilterModel describes the data source data model for queries.
type ProtocolsFcpServiceDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsFcpServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsFcpServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Protocols FC Protocol services data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "FC Protocol SVM name",
						Optional:            true,
					},
				},
				MarkdownDescription: "Filter FC Protocol services by their properties",
				Optional:            true,
			},
			"fcp_services": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "FC Protocol SVM name",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "The administrative state of the FC Protocol service",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "FC Protocol service SVM UUID",
							Computed:            true,
						},
					},
				},
				MarkdownDescription: "FC Protocol services matching the filter",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsFcpServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsFcpServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsFcpServicesDataSourceModel

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

	var filter *interfaces.FcpServicesFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.FcpServicesFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}

	restInfo, err := interfaces.GetProtocolsFcpServices(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetProtocolsFcpServices
		return
	}

	data.ProtocolsFcpServices = make([]ProtocolsFcpServiceDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.ProtocolsFcpServices[index] = ProtocolsFcpServiceDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVMName:       types.StringValue(record.SVM.Name),
			Enabled:       types.BoolValue(record.Enabled),
			ID:            types.StringValue(record.SVM.UUID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
