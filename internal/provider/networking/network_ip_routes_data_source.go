package networking

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &IPRoutesDataSource{}

// NewIPRoutesDataSource is a helper function to simplify the provider implementation.
func NewIPRoutesDataSource() datasource.DataSource {
	return &IPRoutesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "network_ip_routes",
		},
	}
}

// IPRoutesDataSource defines the data source implementation.
type IPRoutesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// IPRoutesDataSourceModel describes the data source data model.
type IPRoutesDataSourceModel struct {
	CxProfileName types.String                  `tfsdk:"cx_profile_name"`
	Gateway       types.String                  `tfsdk:"gateway"`
	IPRoutes      []IPRouteDataSourceModel      `tfsdk:"ip_routes"`
	Filter        *IPRouteDataSourceFilterModel `tfsdk:"filter"`
}

// IPRouteDataSourceFilterModel describes the data source data model for queries.
type IPRouteDataSourceFilterModel struct {
	SVMName     types.String               `tfsdk:"svm_name"`
	Destination DestinationDataSourceModel `tfsdk:"destination"`
	Gateway     types.String               `tfsdk:"gateway"`
}

// Metadata returns the data source type name.
func (d *IPRoutesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *IPRoutesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "IP Routes data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "The IP address of the gateway router leading to the destination.",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "IP Route svm name",
						Optional:            true,
					},
					"destination": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "destination IP address information",
						Attributes: map[string]schema.Attribute{
							"address": schema.StringAttribute{
								MarkdownDescription: "IPv4 or IPv6 address",
								Optional:            true,
							},
							"netmask": schema.StringAttribute{
								MarkdownDescription: "netmask length (16) or IPv4 mask (255.255.0.0). For IPv6, valid range is 1 to 127.",
								Optional:            true,
							},
						},
					},
					"gateway": schema.StringAttribute{
						MarkdownDescription: "The IP address of the gateway router leading to the destination.",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"ip_routes": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"destination": schema.SingleNestedAttribute{
							Required:            true,
							MarkdownDescription: "destination IP address information",
							Attributes: map[string]schema.Attribute{
								"address": schema.StringAttribute{
									MarkdownDescription: "IPv4 or IPv6 address",
									Required:            true,
								},
								"netmask": schema.StringAttribute{
									MarkdownDescription: "netmask length (16) or IPv4 mask (255.255.0.0). For IPv6, valid range is 1 to 127.",
									Computed:            true,
								},
							},
						},
						"gateway": schema.StringAttribute{
							MarkdownDescription: "The IP address of the gateway router leading to the destination.",
							Required:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "IPInterface svm name",
							Optional:            true,
						},
						"metric": schema.Int64Attribute{
							MarkdownDescription: "Indicates a preference order between several routes to the same destination.",
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
func (d *IPRoutesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *IPRoutesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IPRoutesDataSourceModel

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

	var filter *interfaces.IPRouteDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.IPRouteDataSourceFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
			Destination: interfaces.DestinationDataSourceModel{
				Address: data.Filter.Destination.Address.ValueString(),
				Netmask: data.Filter.Destination.Netmask.ValueString(),
			},
			Gateway: data.Filter.Gateway.ValueString(),
		}
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

	restInfo, err := interfaces.GetListIPRoutes(errorHandler, *client, data.Gateway.ValueString(), filter, cluster.Version)
	if err != nil {
		// error reporting done inside GetIPRoutes
		return
	}

	data.IPRoutes = make([]IPRouteDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.IPRoutes[index] = IPRouteDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVMName:       types.StringValue(record.SVMName.Name),
			Destination: &DestinationDataSourceModel{
				Address: types.StringValue(record.Destination.Address),
				Netmask: types.StringValue(record.Destination.Netmask),
			},
			Gateway: types.StringValue(record.Gateway),
			Metric:  types.Int64Value(record.Metric),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
