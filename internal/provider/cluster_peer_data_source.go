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

// TODO:
// copy this file to match you data source (should match internal/provider/cluster_peer_data_source.go)
// replace ClusterPeer with the name of the resource, following go conventions, eg IPInterface
// replace cluster_peer with the name of the resource, for logging purposes, eg ip_interface
// make sure to create internal/interfaces/cluster_peer.go too)
// delete these 5 lines

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ClusterPeerDataSource{}

// NewClusterPeerDataSource is a helper function to simplify the provider implementation.
func NewClusterPeerDataSource() datasource.DataSource {
	return &ClusterPeerDataSource{
		config: resourceOrDataSourceConfig{
			name: "cluster_peer_data_source",
		},
	}
}

// ClusterPeerDataSource defines the data source implementation.
type ClusterPeerDataSource struct {
	config resourceOrDataSourceConfig
}

// ClusterPeerDataSourceModel describes the data source data model.
type ClusterPeerDataSourceModel struct {
	CxProfileName    types.String                     `tfsdk:"cx_profile_name"`
	Name             types.String                     `tfsdk:"name"`
	Remote           *ClusterPeerDataSourceRemote     `tfsdk:"remote"`
	Status           *ClusterPeerDataSourceStatus     `tfsdk:"status"`
	PeerApplications []types.String                   `tfsdk:"peer_applications"`
	Encryption       *ClusterPeerDataSourceEncryption `tfsdk:"encryption"`
	IpAddress        types.String                     `tfsdk:"ip_address"`
	Ipspace          *ClusterPeerDataSourceIpspace    `tfsdk:"ipspace"`
	ID               types.String                     `tfsdk:"id"`
}

// ClusterPeerDataSourceFilterModel describes the data source data model for queries.
type ClusterPeerDataSourceFilterModel struct {
	Name types.String `tfsdk:"name"`
}

// ClusterPeerDataSourceRemote describes the data source data model for remote cluster.
type ClusterPeerDataSourceRemote struct {
	IpAddresses []types.String `tfsdk:"ip_addresses"`
	Name        types.String   `tfsdk:"name"`
}

// ClusterPeerDataSourceStatus describes the data source data model for status.
type ClusterPeerDataSourceStatus struct {
	State types.String `tfsdk:"state"`
}

// ClusterPeerDataSourceEncryption describes the data source data model for encryption.
type ClusterPeerDataSourceEncryption struct {
	Proposed types.String `tfsdk:"proposed"`
	State    types.String `tfsdk:"state"`
}

// ClusterPeerDataSourceIpspace describes the data source data model for ipspace.
type ClusterPeerDataSourceIpspace struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *ClusterPeerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *ClusterPeerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ClusterPeer data source",

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
	}
}

// Configure adds the provider configured client to the data source.
func (d *ClusterPeerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ClusterPeerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterPeerDataSourceModel

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

	restInfo, err := interfaces.GetClusterPeerByName(errorHandler, *client, data.Name.ValueString())
	if err != nil {
		// error reporting done inside GetClusterPeer
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.ID = types.StringValue(restInfo.UUID)
	data.Remote = &ClusterPeerDataSourceRemote{
		IpAddresses: make([]types.String, len(restInfo.Remote.IpAddress)),
		Name:        types.StringValue(restInfo.Remote.Name),
	}
	for index, ipAddress := range restInfo.Remote.IpAddress {
		data.Remote.IpAddresses[index] = types.StringValue(ipAddress)
	}
	data.Status = &ClusterPeerDataSourceStatus{
		State: types.StringValue(restInfo.Status.State),
	}
	data.PeerApplications = make([]types.String, len(restInfo.PeerApplications))
	for index, peerApplication := range restInfo.PeerApplications {
		data.PeerApplications[index] = types.StringValue(peerApplication)
	}
	data.Encryption = &ClusterPeerDataSourceEncryption{
		Proposed: types.StringValue(restInfo.Encryption.Propsed),
		State:    types.StringValue(restInfo.Encryption.State),
	}
	data.IpAddress = types.StringValue(restInfo.IpAddress)
	data.Ipspace = &ClusterPeerDataSourceIpspace{
		Name: types.StringValue(restInfo.Ipspace.Name),
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
