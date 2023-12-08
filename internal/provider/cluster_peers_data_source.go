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
var _ datasource.DataSource = &ClusterPeersDataSource{}

// NewClusterPeersDataSource is a helper function to simplify the provider implementation.
func NewClusterPeersDataSource() datasource.DataSource {
	return &ClusterPeersDataSource{
		config: resourceOrDataSourceConfig{
			name: "cluster_peers_data_source",
		},
	}
}

// ClusterPeersDataSource defines the data source implementation.
type ClusterPeersDataSource struct {
	config resourceOrDataSourceConfig
}

// ClusterPeersDataSourceModel describes the data source data model.
type ClusterPeersDataSourceModel struct {
	CxProfileName types.String                      `tfsdk:"cx_profile_name"`
	ClusterPeers  []ClusterPeerDataSourceModel      `tfsdk:"cluster_peers"`
	Filter        *ClusterPeerDataSourceFilterModel `tfsdk:"filter"`
}

// ClusterPeerDataSourceFilterModel describes the data source data model for queries.
type ClusterPeerDataSourceFilterModel struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *ClusterPeersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *ClusterPeersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ClusterPeers data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "ClusterPeer name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"cluster_peers": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "ClusterPeer name",
							Required:            true,
						},
						"remote": schema.SingleNestedAttribute{
							MarkdownDescription: "Remote cluster",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"ip_addresses": schema.ListAttribute{
									ElementType:         types.StringType,
									MarkdownDescription: "List of IP addresses of remote cluster",
									Computed:            true,
								},
								"name": schema.StringAttribute{
									MarkdownDescription: "Name of remote cluster",
									Computed:            true,
								},
							},
						},
						"status": schema.SingleNestedAttribute{
							MarkdownDescription: "Status of cluster peer",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"state": schema.StringAttribute{
									MarkdownDescription: "State of cluster peer",
									Computed:            true,
								},
							},
						},
						"peer_applications": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "List of peer applications",
							Computed:            true,
						},
						"encryption": schema.SingleNestedAttribute{
							MarkdownDescription: "Encryption of cluster peer",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"proposed": schema.StringAttribute{
									MarkdownDescription: "Proposed encryption of cluster peer",
									Computed:            true,
								},
								"state": schema.StringAttribute{
									MarkdownDescription: "State of encryption of cluster peer",
									Computed:            true,
								},
							},
						},
						"ip_address": schema.StringAttribute{
							MarkdownDescription: "IP address",
							Computed:            true,
						},
						"ipspace": schema.SingleNestedAttribute{
							MarkdownDescription: "Ipspace of cluster peer",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Name of ipspace of cluster peer",
									Computed:            true,
								},
							},
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "ID of cluster peer",
							Computed:            true,
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "Cluster Peers",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ClusterPeersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ClusterPeersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterPeersDataSourceModel

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

	var filter *interfaces.ClusterPeerDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.ClusterPeerDataSourceFilterModel{
			Name: data.Filter.Name.ValueString(),
		}
	}
	restInfo, err := interfaces.GetClusterPeers(errorHandler, *client, filter)
	if err != nil {
		return
	}

	data.ClusterPeers = make([]ClusterPeerDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.ClusterPeers[index] = ClusterPeerDataSourceModel{
			CxProfileName: data.CxProfileName,
			Name:          types.StringValue(record.Name),
			ID:            types.StringValue(record.UUID),
			Remote: &ClusterPeerDataSourceRemote{
				IPAddresses: make([]types.String, len(record.Remote.IPAddress)),
				Name:        types.StringValue(record.Remote.Name),
			},
			Status: &ClusterPeerDataSourceStatus{
				State: types.StringValue(record.Status.State),
			},
			PeerApplications: make([]types.String, len(record.PeerApplications)),
			Encryption: &ClusterPeerDataSourceEncryption{
				Proposed: types.StringValue(record.Encryption.Propsed),
				State:    types.StringValue(record.Encryption.State),
			},
			IPAddress: types.StringValue(record.IPAddress),
			Ipspace: &ClusterPeerDataSourceIpspace{
				Name: types.StringValue(record.Ipspace.Name),
			},
		}
		for index, IPAddress := range record.Remote.IPAddress {
			data.ClusterPeers[index].Remote.IPAddresses[index] = types.StringValue(IPAddress)
		}
		for index, peerApplication := range record.PeerApplications {
			data.ClusterPeers[index].PeerApplications[index] = types.StringValue(peerApplication)
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
