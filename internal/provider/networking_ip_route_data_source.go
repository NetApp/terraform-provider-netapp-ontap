package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &IPRouteDataSource{}

// NewIPRouteDataSource is a helper function to simplify the provider implementation.
func NewIPRouteDataSource() datasource.DataSource {
	return &IPRouteDataSource{
		config: resourceOrDataSourceConfig{
			name: "networking_ip_route_data_source",
		},
	}
}

// IPRouteDataSource defines the data source implementation.
type IPRouteDataSource struct {
	config resourceOrDataSourceConfig
}

// IPRouteDataSourceModel describes the data source data model.
type IPRouteDataSourceModel struct {
	CxProfileName types.String                `tfsdk:"cx_profile_name"`
	SVMName       types.String                `tfsdk:"svm_name"`
	Destination   *DestinationDataSourceModel `tfsdk:"destination"`
	Gateway       types.String                `tfsdk:"gateway"`
	Metric        types.Int64                 `tfsdk:"metric"`
}

// DestinationDataSourceModel describes the data source of Protocols
type DestinationDataSourceModel struct {
	Address types.String `tfsdk:"address"`
	Netmask types.String `tfsdk:"netmask"`
}

// IPRouteDataSourceFilterModel describes the data source data model for queries.
type IPRouteDataSourceFilterModel struct {
	SVMName     types.String               `tfsdk:"svm.name"`
	Destination DestinationDataSourceModel `tfsdk:"destination"`
}

// Metadata returns the data source type name.
func (d *IPRouteDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *IPRouteDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NetRoute data source",

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
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface vserver name",
				Optional:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "The IP address of the gateway router leading to the destination.",
				Computed:            true,
			},
			"metric": schema.Int64Attribute{
				MarkdownDescription: "Indicates a preference order between several routes to the same destination.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *IPRouteDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	d.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *IPRouteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IPRouteDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", fmt.Sprintf("No Cluster found"))
		return
	}

	restInfo, err := interfaces.GetIPRoute(errorHandler, *client, data.Destination.Address.ValueString(), data.SVMName.ValueString(), data.Gateway.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetNetRoute
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No IP Route found", fmt.Sprintf("No IP Route %s found", data.Destination.Address.ValueString()))
		return
	}

	data.Destination.Address = types.StringValue(restInfo.Destination.Address)
	data.Destination.Netmask = types.StringValue(restInfo.Destination.Netmask)
	data.Gateway = types.StringValue(restInfo.Gateway)
	data.Metric = types.Int64Value(restInfo.Metric)
	data.SVMName = types.StringValue(restInfo.SVMName.Name)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
