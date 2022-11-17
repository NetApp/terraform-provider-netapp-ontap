package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ClusterDataSource{}

// NewClusterDataSource is a helper function to simplify the provider implementation.
func NewClusterDataSource() datasource.DataSource {
	return &ClusterDataSource{}
}

// ClusterDataSource defines the data source implementation.
type ClusterDataSource struct {
	client *restclient.RestClient
	config Config
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
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

// GetSchema defines the schema for the data source.
func (d *ClusterDataSource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cluster data source",

		Attributes: map[string]tfsdk.Attribute{
			"cx_profile_name": {
				MarkdownDescription: "Connection profile name",
				Type:                types.StringType,
				Required:            true,
			},
			"name": {
				MarkdownDescription: "Cluster name",
				Type:                types.StringType,
				Computed:            true,
			},
			"version": {
				MarkdownDescription: "ONTAP software version",
				Computed:            true,
				Optional:            true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"full": {
						Type:     types.StringType,
						Computed: true,
					},
				}),
			},
			"nodes": {
				MarkdownDescription: "Cluster Nodes",
				Computed:            true,
				Optional:            true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						Type:     types.StringType,
						Computed: true,
					},
					"management_ip_addresses": {
						Type: types.ListType{
							ElemType: types.StringType,
						},
						Computed: true,
					},
				}),
			},
		},
	}, nil
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
	d.config = config
	// we need to defer setting the client until we can read the connection profile name
	d.client = nil
}

// Read refreshes the Terraform state with the latest data.
func (d *ClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := d.config.NewClient(ctx, resp.Diagnostics, data.CxProfileName.ValueString())
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	cluster, err := interfaces.GetCluster(ctx, resp.Diagnostics, *client)
	if err != nil {
		msg := fmt.Sprintf("error reading cluster: %s", err)
		tflog.Error(ctx, msg)
		resp.Diagnostics.AddError("error reading cluster", msg)
		return
	}

	data.Name = types.StringValue(cluster.Name)
	data.Version = &versionModel{
		Full: types.StringValue(cluster.Version.Full),
	}

	nodes, err := interfaces.GetClusterNodes(ctx, resp.Diagnostics, *client)
	if err != nil {
		msg := fmt.Sprintf("error reading cluster nodes: %s", err)
		tflog.Error(ctx, msg)
		resp.Diagnostics.AddError("error reading cluster nodes", msg)
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
