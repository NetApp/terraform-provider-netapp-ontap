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
var _ datasource.DataSource = &ClusterDataSource{}

// NewClusterDataSource is a helper function to simplify the provider implementation.
func NewClusterDataSource() datasource.DataSource {
	return &ClusterDataSource{
		config: resourceOrDataSourceConfig{
			name: "cluster_data_source",
		},
	}
}

// TODO - Add the data source implementation here

// ClusterDataSource defines the data source implementation.
type ClusterDataSource struct {
	config resourceOrDataSourceConfig
}

// ClusterDataSourceModel describes the data source data model.
type ClusterDataSourceModel struct {
	// ConfigurableAttribute types.String `tfsdk:"configurable_attribute"`
	// ID                    types.String `tfsdk:"id"`
	CxProfileName types.String          `tfsdk:"cx_profile_name"`
	Name          types.String          `tfsdk:"name"`
	Version       *versionModel         `tfsdk:"version"`
	Nodes         []NodeDataSourceModel `tfsdk:"nodes"`
}

// NodeDataSourceModel describes the data source data model.
type NodeDataSourceModel struct {
	Name            types.String `tfsdk:"name"`
	MgmtIPAddresses types.List   `tfsdk:"management_ip_addresses"`
}

type versionModel struct {
	Full types.String `tfsdk:"full"`
}

// Metadata returns the data source type name.
func (d *ClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *ClusterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cluster data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster name",
			},
			"version": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"full": schema.StringAttribute{
						MarkdownDescription: "ONTAP software version",
						Computed:            true,
					},
				},
				Computed:            true,
				MarkdownDescription: "ONTAP software version",
			},
			"nodes": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Cluster Nodes",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"management_ip_addresses": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ClusterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// TODO: world domnication

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
		errorHandler.MakeAndReportError("Cluster Not found", fmt.Sprintf("cluster %s not found.", data.Name))
		return
	}

	data.Name = types.StringValue(cluster.Name)
	data.Version = &versionModel{
		Full: types.StringValue(cluster.Version.Full),
	}

	nodes, err := interfaces.GetClusterNodes(errorHandler, *client)
	if err != nil {
		return
	}
	if nodes == nil {
		errorHandler.MakeAndReportError("Cluster Nodes Not found", fmt.Sprintf("cluster nodes not found."))
		return
	}

	data.Nodes = make([]NodeDataSourceModel, 1)
	ipAddressesIn := make([]string, 1)
	ipAddressesIn[0] = nodes[0].ManagementInterfaces[0].IP.Address
	ipAddressesOut, _ := types.ListValueFrom(ctx, types.StringType, ipAddressesIn)
	data.Nodes[0] = NodeDataSourceModel{
		Name:            types.StringValue(nodes[0].Name),
		MgmtIPAddresses: ipAddressesOut,
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
