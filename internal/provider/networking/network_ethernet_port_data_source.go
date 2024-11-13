package networking

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
var _ datasource.DataSource = &EthernetPortDataSource{}

// NewEthernetPortDataSource is a helper function to simplify the provider implementation.
func NewEthernetPortDataSource() datasource.DataSource {
	return &EthernetPortDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "port",
		},
	}
}

// EthernetPortDataSource defines the data source implementation.
type EthernetPortDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// EthernetPortDataSourceModel describes the data source data model.
type EthernetPortDataSourceModel struct {
	BroadcastDomainID types.String         `tfsdk:"broadcast_domain_id"`
	CxProfileName     types.String         `tfsdk:"cx_profile_name"`
	Enabled           types.Bool           `tfsdk:"enabled"`
	InterfaceCount    types.Int64          `tfsdk:"interface_count"`
	LAG               *LAGDataSourceModel  `tfsdk:"lag"`
	MACAddress        types.String         `tfsdk:"mac_address"`
	MTU               types.Int64          `tfsdk:"mtu"`
	Name              types.String         `tfsdk:"name"`
	NodeID            types.String         `tfsdk:"node_id"`
	RDMAProtocols     []types.String       `tfsdk:"rdma_protocols"`
	Reachability      types.String         `tfsdk:"reachability"`
	Speed             types.Int64          `tfsdk:"speed"`
	State             types.String         `tfsdk:"state"`
	Type              types.String         `tfsdk:"type"`
	VLAN              *VLANDataSourceModel `tfsdk:"vlan"`
	ID                types.String         `tfsdk:"id"`
}

// LAGDataSourceModel describes the data source model for LAG ports, policy and mode.
type LAGDataSourceModel struct {
	ActivePortsID      []types.String `tfsdk:"active_ports_id"`
	DistributionPolicy types.String   `tfsdk:"distribution_policy"`
	MemberPortsID      []types.String `tfsdk:"member_ports_id"`
	Mode               types.String   `tfsdk:"mode"`
}

// VLANDataSourceModel describes the data source model for VLAN base port and tag.
type VLANDataSourceModel struct {
	BasePortID types.String `tfsdk:"base_port_id"`
	Tag        types.Int64  `tfsdk:"tag"`
}

// Metadata returns the data source type name.
func (d *EthernetPortDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *EthernetPortDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Ethernet Port data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"broadcast_domain_id": schema.StringAttribute{
				MarkdownDescription: "Broadcast domain UUID",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Port enabled",
				Computed:            true,
			},
			"interface_count": schema.Int64Attribute{
				MarkdownDescription: "Number of interfaces hosted",
				Computed:            true,
			},
			"lag": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"active_ports_id": schema.SetAttribute{
						ElementType:         types.StringType,
						Computed:            true,
						MarkdownDescription: "Active ports of a LAG (ifgrp)",
					},
					"distribution_policy": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Policy for mapping flows to ports for outbound packets through a LAG (ifgrp)",
					},
					"member_ports_id": schema.SetAttribute{
						ElementType:         types.StringType,
						Computed:            true,
						MarkdownDescription: "Array of ports belonging to the LAG, regardless of their state",
					},
					"mode": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "Determines how the ports interact with the switch",
					},
				},
				Computed: true,
			},
			"mac_address": schema.StringAttribute{
				MarkdownDescription: "Port MAC address",
				Computed:            true,
			},
			"mtu": schema.Int64Attribute{
				MarkdownDescription: "MTU of the port in bytes",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Portname, such as e0a, e1b-100 (VLAN on Ethernet), a0c (LAG/ifgrp), a0d-200 (VLAN on LAG/ifgrp), e0a.pv1 (p-VLAN, in select environments only)",
				Required:            true,
			},
			"node_id": schema.StringAttribute{
				MarkdownDescription: "UUID of node on which the port is located",
				Computed:            true,
			},
			"rdma_protocols": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Supported RDMA offload protocols",
				Computed:            true,
			},
			"reachability": schema.StringAttribute{
				MarkdownDescription: "Reachability status of the port",
				Computed:            true,
			},
			"speed": schema.Int64Attribute{
				MarkdownDescription: "Link speed in Mbps",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Operational state of the port",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of physical or virtual port",
				Computed:            true,
			},
			"vlan": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"base_port_id": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "VLAN base port UUID",
					},
					"tag": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "VLAN ID",
					},
				},
				Computed: true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Port UUID",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *EthernetPortDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *EthernetPortDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EthernetPortDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
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

	// Call ONTAP REST API for reading ethernet_port info
	restInfo, err := interfaces.GetEthernetPortByName(
		errorHandler,
		*client,
		data.Name.ValueString(),
	)
	if err != nil {
		// error reporting done inside GetEthernetPortByName
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError(
			"No Ethernet Port Found",
			fmt.Sprintf("No ethernet port '%s' found.", data.Name.ValueString()),
		)

		return
	}

	// Copy ethernet_port info to data source model
	var protocols []types.String
	for _, v := range restInfo.RDMAProtocols {
		protocols = append(protocols, types.StringValue(v))
	}

	data.BroadcastDomainID = types.StringValue(restInfo.BroadcastDomain.UUID)
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.InterfaceCount = types.Int64Value(restInfo.InterfaceCount)
	data.MACAddress = types.StringValue(restInfo.MACAddress)
	data.MTU = types.Int64Value(restInfo.MTU)
	data.Name = types.StringValue(restInfo.Name)
	data.NodeID = types.StringValue(restInfo.Node.UUID)
	data.RDMAProtocols = protocols
	data.Reachability = types.StringValue(restInfo.Reachability)
	data.Speed = types.Int64Value(restInfo.Speed)
	data.State = types.StringValue(restInfo.State)
	data.Type = types.StringValue(restInfo.Type)
	data.ID = types.StringValue(restInfo.UUID)

	switch data.Type.ValueString() {
	case "lag":
		var activePorts, memberPorts []types.String
		for _, v := range restInfo.LAG.ActivePorts {
			activePorts = append(activePorts, types.StringValue(v.UUID))
		}
		for _, v := range restInfo.LAG.MemberPorts {
			memberPorts = append(memberPorts, types.StringValue(v.UUID))
		}
		data.LAG = &LAGDataSourceModel{
			ActivePortsID:      activePorts,
			DistributionPolicy: types.StringValue(restInfo.LAG.DistributionPolicy),
			MemberPortsID:      memberPorts,
			Mode:               types.StringValue(restInfo.LAG.Mode),
		}
	case "vlan":
		data.VLAN = &VLANDataSourceModel{
			BasePortID: types.StringValue(restInfo.VLAN.BasePort.UUID),
			Tag:        types.Int64Value(restInfo.VLAN.Tag),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
