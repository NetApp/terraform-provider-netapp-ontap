package cluster

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
var _ datasource.DataSource = &ClusterDataSource{}

// NewClusterDataSource is a helper function to simplify the provider implementation.
func NewClusterDataSource() datasource.DataSource {
	return &ClusterDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cluster_data_source",
		},
	}
}

// ClusterDataSource defines the data source implementation.
//
//nolint:golint
type ClusterDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ClusterDataSourceModel describes the data source data model.
//
//nolint:golint
type ClusterDataSourceModel struct {
	// ConfigurableAttribute types.String `tfsdk:"configurable_attribute"`
	// ID                    types.String `tfsdk:"id"`
	CxProfileName        types.String          `tfsdk:"cx_profile_name"`
	Name                 types.String          `tfsdk:"name"`
	Version              *versionModel         `tfsdk:"version"`
	Nodes                []NodeDataSourceModel `tfsdk:"nodes"`
	Contact              types.String          `tfsdk:"contact"`
	Location             types.String          `tfsdk:"location"`
	DNSDomains           types.Set             `tfsdk:"dns_domains"`
	NameServers          types.Set             `tfsdk:"name_servers"`
	TimeZone             types.Object          `tfsdk:"timezone"`
	Certificate          types.Object          `tfsdk:"certificate"`
	NtpServers           types.Set             `tfsdk:"ntp_servers"`
	ManagementInterfaces types.Set             `tfsdk:"management_interfaces"`
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
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
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
			"contact": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Contact information. Example: support@company.com",
			},
			"location": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Location information",
			},
			"dns_domains": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "A list of DNS domains.",
			},
			"name_servers": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "The list of IP addresses of the DNS servers. Addresses can be either IPv4 or IPv6 addresses.",
			},
			"timezone": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Time zone",
					},
				},
				Computed:            true,
				MarkdownDescription: "Time zone",
			},
			"certificate": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed: true,
					},
				},
				Computed:            true,
				MarkdownDescription: "Certificate",
			},
			"ntp_servers": schema.SetAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "Host name, IPv4 address, or IPv6 address for the external NTP time servers.",
			},
			"management_interfaces": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"address": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "IP address",
								},
							},
							Computed:            true,
							MarkdownDescription: "IP address",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "ID",
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "A list of network interface",
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
func (d *ClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterDataSourceModel

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
	data.Contact = types.StringValue(cluster.Contact)
	data.Location = types.StringValue(cluster.Location)

	// dns domains
	elements := []attr.Value{}
	for _, dnsDomain := range cluster.DNSDomains {
		elements = append(elements, types.StringValue(dnsDomain))
	}
	setValue, diags := types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	data.DNSDomains = setValue

	//name servers
	elements = []attr.Value{}
	for _, nameServer := range cluster.NameServers {
		elements = append(elements, types.StringValue(nameServer))
	}
	setValue, diags = types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	data.NameServers = setValue
	// time zone
	elementTypes := map[string]attr.Type{
		"name": types.StringType,
	}
	objectElements := map[string]attr.Value{
		"name": types.StringValue(cluster.TimeZone.Name),
	}
	objectValue, diags := types.ObjectValue(elementTypes, objectElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.TimeZone = objectValue

	// certificate
	elementTypes = map[string]attr.Type{
		"id": types.StringType,
	}
	objectElements = map[string]attr.Value{
		"id": types.StringValue(cluster.ClusterCertificate.ID),
	}
	objectValue, diags = types.ObjectValue(elementTypes, objectElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Certificate = objectValue

	// ntp servers
	elements = []attr.Value{}
	for _, ntpServer := range cluster.NtpServers {
		elements = append(elements, types.StringValue(ntpServer))
	}
	setValue, diags = types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	data.NtpServers = setValue

	// management interfaces
	setElements := []attr.Value{}
	for _, mgmInterface := range cluster.ManagementInterfaces {
		nestedElementTypes := map[string]attr.Type{
			"address": types.StringType,
		}
		nestedVolumeElements := map[string]attr.Value{
			"address": types.StringValue(mgmInterface.IP.Address),
		}
		originVolumeObjectValue, diags := types.ObjectValue(nestedElementTypes, nestedVolumeElements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		elementTypes := map[string]attr.Type{
			"ip":   types.ObjectType{AttrTypes: nestedElementTypes},
			"name": types.StringType,
			"id":   types.StringType,
		}
		elements := map[string]attr.Value{
			"ip":   originVolumeObjectValue,
			"name": types.StringValue(mgmInterface.Name),
			"id":   types.StringValue(mgmInterface.ID),
		}
		objectValue, diags := types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		setElements = append(setElements, objectValue)
	}

	setValue, diags = types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"ip": types.ObjectType{AttrTypes: map[string]attr.Type{
				"address": types.StringType,
			}},
			"name": types.StringType,
			"id":   types.StringType,
		},
	}, setElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.ManagementInterfaces = setValue

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
