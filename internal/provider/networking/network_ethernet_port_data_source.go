package networking

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	BroadcastDomain *EthernetPortBroadcastDomainDataSourceModel `tfsdk:"broadcast_domain"`
	CxProfileName   types.String                                `tfsdk:"cx_profile_name"`
	Enabled         types.Bool                                  `tfsdk:"enabled"`
	InterfaceCount  types.Int64                                 `tfsdk:"interface_count"`
	LAG             *LAGDataSourceModel                         `tfsdk:"lag"`
	MACAddress      types.String                                `tfsdk:"mac_address"`
	MTU             types.Int64                                 `tfsdk:"mtu"`
	Name            types.String                                `tfsdk:"name"`
	Node            *EthernetPortNodeDataSourceModel            `tfsdk:"node"`
	RDMAProtocols   types.Set                                   `tfsdk:"rdma_protocols"`
	Reachability    types.String                                `tfsdk:"reachability"`
	Speed           types.Int64                                 `tfsdk:"speed"`
	State           types.String                                `tfsdk:"state"`
	Type            types.String                                `tfsdk:"type"`
	VLAN            *VLANDataSourceModel                        `tfsdk:"vlan"`
	ID              types.String                                `tfsdk:"id"`
}

// EthernetPortBroadcastDomainDataSourceModel describes the data source model for broadcast domains.
type EthernetPortBroadcastDomainDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	IPSpace types.String `tfsdk:"ipspace"`
	Name    types.String `tfsdk:"name"`
}

// LAGDataSourceModel describes the data source model for LAG ports, policy and mode.
type LAGDataSourceModel struct {
	ActivePorts        types.Set    `tfsdk:"active_ports"`
	DistributionPolicy types.String `tfsdk:"distribution_policy"`
	MemberPorts        types.Set    `tfsdk:"member_ports"`
	Mode               types.String `tfsdk:"mode"`
}

// EthernetPortBroadcastDomainDataSourceModel describes the data source model for nodes.
type EthernetPortNodeDataSourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// VLANDataSourceModel describes the data source model for VLAN base port and tag.
type VLANDataSourceModel struct {
	BasePort types.String `tfsdk:"base_port"`
	Tag      types.Int64  `tfsdk:"tag"`
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
			"broadcast_domain": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"ipspace": schema.StringAttribute{
						MarkdownDescription: "Name of the broadcast domain's IPspace",
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Name of the broadcast domain, scoped to its IPspace",
						Computed:            true,
					},
					"id": schema.StringAttribute{
						MarkdownDescription: "Broadcast domain UUID",
						Computed:            true,
					},
				},
				MarkdownDescription: "Broadcast domain properties",
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
					"active_ports": schema.SetAttribute{
						MarkdownDescription: "Active ports of a LAG (ifgrp)",
						ElementType:         types.StringType,
						Computed:            true,
					},
					"distribution_policy": schema.StringAttribute{
						MarkdownDescription: "Policy for mapping flows to ports for outbound packets through a LAG (ifgrp)",
						Computed:            true,
					},
					"member_ports": schema.SetAttribute{
						MarkdownDescription: "Array of ports belonging to the LAG, regardless of their state",
						ElementType:         types.StringType,
						Computed:            true,
					},
					"mode": schema.StringAttribute{
						MarkdownDescription: "Determines how the ports interact with the switch",
						Computed:            true,
					},
				},
				MarkdownDescription: "LAG (ifgrp) properties",
				Computed:            true,
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
			"node": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Name of the node on which the port is located",
						Computed:            true,
					},
					"id": schema.StringAttribute{
						MarkdownDescription: "Node UUID",
						Computed:            true,
					},
				},
				MarkdownDescription: "Node properties",
				Computed:            true,
			},
			"rdma_protocols": schema.SetAttribute{
				MarkdownDescription: "Supported RDMA offload protocols",
				ElementType:         types.StringType,
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
					"base_port": schema.StringAttribute{
						MarkdownDescription: "VLAN base port",
						Computed:            true,
					},
					"tag": schema.Int64Attribute{
						MarkdownDescription: "VLAN ID",
						Computed:            true,
					},
				},
				MarkdownDescription: "VLAN properties",
				Computed:            true,
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
	data.BroadcastDomain = &EthernetPortBroadcastDomainDataSourceModel{
		ID:      types.StringValue(restInfo.BroadcastDomain.UUID),
		IPSpace: types.StringValue(restInfo.BroadcastDomain.IPSpace.Name),
		Name:    types.StringValue(restInfo.BroadcastDomain.Name),
	}
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.InterfaceCount = types.Int64Value(restInfo.InterfaceCount)
	data.MACAddress = types.StringValue(restInfo.MACAddress)
	data.MTU = types.Int64Value(restInfo.MTU)
	data.Name = types.StringValue(restInfo.Name)
	data.Node = &EthernetPortNodeDataSourceModel{
		ID:   types.StringValue(restInfo.Node.UUID),
		Name: types.StringValue(restInfo.Node.Name),
	}
	data.Reachability = types.StringValue(restInfo.Reachability)
	data.Speed = types.Int64Value(restInfo.Speed)
	data.State = types.StringValue(restInfo.State)
	data.Type = types.StringValue(restInfo.Type)
	data.ID = types.StringValue(restInfo.UUID)

	// rdma_protocols set
	if len(restInfo.RDMAProtocols) > 0 {
		var protocols []attr.Value
		for _, v := range restInfo.RDMAProtocols {
			protocols = append(protocols, types.StringValue(v))
		}
		protocolsSet, diags := types.SetValue(types.StringType, protocols)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.RDMAProtocols = protocolsSet
	}

	switch data.Type.ValueString() {
	case "lag":
		// active_ports set
		var activePorts, memberPorts []attr.Value
		for _, v := range restInfo.LAG.ActivePorts {
			activePorts = append(activePorts, types.StringValue(v.Name))
		}
		activePortsSet, diags := types.SetValue(types.StringType, activePorts)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// member_ports set
		for _, v := range restInfo.LAG.MemberPorts {
			memberPorts = append(memberPorts, types.StringValue(v.Name))
		}
		memberPortsSet, diags := types.SetValue(types.StringType, memberPorts)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.LAG = &LAGDataSourceModel{
			ActivePorts:        activePortsSet,
			DistributionPolicy: types.StringValue(restInfo.LAG.DistributionPolicy),
			MemberPorts:        memberPortsSet,
			Mode:               types.StringValue(restInfo.LAG.Mode),
		}
	case "vlan":
		data.VLAN = &VLANDataSourceModel{
			BasePort: types.StringValue(restInfo.VLAN.BasePort.Name),
			Tag:      types.Int64Value(restInfo.VLAN.Tag),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
