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
var _ datasource.DataSource = &EthernetPortsDataSource{}

// NewEthernetPortsDataSource is a helper function to simplify the provider implementation.
func NewEthernetPortsDataSource() datasource.DataSource {
	return &EthernetPortsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "ports",
		},
	}
}

// EthernetPortsDataSource defines the data source implementation.
type EthernetPortsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// EthernetPortDataSourceModel describes the data source data model.
type EthernetPortsDataSourceModel struct {
	CxProfileName types.String                        `tfsdk:"cx_profile_name"`
	Filter        *EthernetPortsDataSourceFilterModel `tfsdk:"filter"`
	Ports         []EthernetPortDataSourceModel       `tfsdk:"ports"`
}

// IPInterfaceDataSourceFilterModel describes the data source data model for queries.
type EthernetPortsDataSourceFilterModel struct {
	Name  types.String `tfsdk:"name"`
	State types.String `tfsdk:"state"`
	Type  types.String `tfsdk:"type"`
}

// Metadata returns the data source type name.
func (d *EthernetPortsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *EthernetPortsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Ethernet Ports data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Portname, such as e0a, e1b-100 (VLAN on Ethernet), a0c (LAG/ifgrp), a0d-200 (VLAN on LAG/ifgrp), e0a.pv1 (p-VLAN, in select environments only)",
					},
					"state": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Operational state of the port",
					},
					"type": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Type of physical or virtual port",
					},
				},
				MarkdownDescription: "Filter ethernet ports by their properties",
				Optional:            true,
			},
			"ports": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
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
				},
				MarkdownDescription: "Ethernet ports matching the filter",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *EthernetPortsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *EthernetPortsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EthernetPortsDataSourceModel

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

	// Prepare filter
	var filter *interfaces.EthernetPortsDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.EthernetPortsDataSourceFilterModel{
			Name:  data.Filter.Name.ValueString(),
			State: data.Filter.State.ValueString(),
			Type:  data.Filter.Type.ValueString(),
		}
	}

	// Call ONTAP REST API for reading ethernet_port info
	restInfo, err := interfaces.GetListEthernetPorts(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetListEthernetPorts
		return
	}

	// Copy ethernet_port info to data source model
	data.Ports = make([]EthernetPortDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.Ports[index] = EthernetPortDataSourceModel{
			BroadcastDomain: &EthernetPortBroadcastDomainDataSourceModel{
				ID:      types.StringValue(record.BroadcastDomain.UUID),
				IPSpace: types.StringValue(record.BroadcastDomain.IPSpace.Name),
				Name:    types.StringValue(record.BroadcastDomain.Name),
			},
			Enabled:        types.BoolValue(record.Enabled),
			InterfaceCount: types.Int64Value(record.InterfaceCount),
			MACAddress:     types.StringValue(record.MACAddress),
			MTU:            types.Int64Value(record.MTU),
			Name:           types.StringValue(record.Name),
			Node: &EthernetPortNodeDataSourceModel{
				ID:   types.StringValue(record.Node.UUID),
				Name: types.StringValue(record.Node.Name),
			},
			Reachability: types.StringValue(record.Reachability),
			Speed:        types.Int64Value(record.Speed),
			State:        types.StringValue(record.State),
			Type:         types.StringValue(record.Type),
			ID:           types.StringValue(record.UUID),
		}

		// rdma_protocols set
		var protocols []attr.Value
		for _, v := range record.RDMAProtocols {
			protocols = append(protocols, types.StringValue(v))
		}
		protocolsSet, diags := types.SetValue(types.StringType, protocols)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Ports[index].RDMAProtocols = protocolsSet

		switch data.Ports[index].Type.ValueString() {
		case "lag":
			// active_ports set
			var activePorts, memberPorts []attr.Value
			for _, v := range record.LAG.ActivePorts {
				activePorts = append(activePorts, types.StringValue(v.Name))
			}
			activePortsSet, diags := types.SetValue(types.StringType, activePorts)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			// member_ports set
			for _, v := range record.LAG.MemberPorts {
				memberPorts = append(memberPorts, types.StringValue(v.Name))
			}
			memberPortsSet, diags := types.SetValue(types.StringType, memberPorts)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			data.Ports[index].LAG = &LAGDataSourceModel{
				ActivePorts:        activePortsSet,
				DistributionPolicy: types.StringValue(record.LAG.DistributionPolicy),
				MemberPorts:        memberPortsSet,
				Mode:               types.StringValue(record.LAG.Mode),
			}
		case "vlan":
			data.Ports[index].VLAN = &VLANDataSourceModel{
				BasePort: types.StringValue(record.VLAN.BasePort.Name),
				Tag:      types.Int64Value(record.VLAN.Tag),
			}
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
