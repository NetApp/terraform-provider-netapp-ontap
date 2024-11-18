package networking

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &EthernetPortResource{}
	_ resource.ResourceWithConfigure   = &EthernetPortResource{}
	_ resource.ResourceWithImportState = &EthernetPortResource{}
)

// NewEthernetPortResource is a helper function to simplify the provider implementation.
func NewEthernetPortResource() resource.Resource {
	return &EthernetPortResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "port",
		},
	}
}

// EthernetPortResource defines the resource implementation.
type EthernetPortResource struct {
	config connection.ResourceOrDataSourceConfig
}

// EthernetPortResourceModel describes the resource data model.
type EthernetPortResourceModel struct {
	BroadcastDomainID types.String       `tfsdk:"broadcast_domain_id"`
	CxProfileName     types.String       `tfsdk:"cx_profile_name"`
	LAG               *LAGResourceModel  `tfsdk:"lag"`
	Name              types.String       `tfsdk:"name"`
	NodeID            types.String       `tfsdk:"node_id"`
	Type              types.String       `tfsdk:"type"`
	VLAN              *VLANResourceModel `tfsdk:"vlan"`
	ID                types.String       `tfsdk:"id"`

	// Enabled        types.Bool   `tfsdk:"enabled"`
	// InterfaceCount types.Int64  `tfsdk:"interface_count"`
	// MACAddress     types.String `tfsdk:"mac_address"`
	// MTU            types.Int64  `tfsdk:"mtu"`
	// RDMAProtocols  types.Set    `tfsdk:"rdma_protocols"`
	// Reachability   types.String `tfsdk:"reachability"`
	// Speed          types.Int64  `tfsdk:"speed"`
	// State          types.String `tfsdk:"state"`
}

// LAGResourceModel describes the data source model for LAG ports, policy and mode.
type LAGResourceModel struct {
	ActivePortsID      types.Set    `tfsdk:"active_ports_id"`
	DistributionPolicy types.String `tfsdk:"distribution_policy"`
	MemberPortsID      types.Set    `tfsdk:"member_ports_id"`
	Mode               types.String `tfsdk:"mode"`
}

// VLANResourceModel describes the data source model for VLAN base port and tag.
type VLANResourceModel struct {
	BasePortID types.String `tfsdk:"base_port_id"`
	Tag        types.Int64  `tfsdk:"tag"`
}

// Metadata returns the resource type name.
func (r *EthernetPortResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *EthernetPortResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Ethernet Port resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"broadcast_domain_id": schema.StringAttribute{
				MarkdownDescription: "Broadcast domain UUID",
				Optional:            true,
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
				MarkdownDescription: "",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Portname, such as e0a, e1b-100 (VLAN on Ethernet), a0c (LAG/ifgrp), a0d-200 (VLAN on LAG/ifgrp), e0a.pv1 (p-VLAN, in select environments only)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"node_id": schema.StringAttribute{
				MarkdownDescription: "UUID of node on which the port is located",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of physical or virtual port",
				Required:            true,
			},
			"vlan": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"base_port_id": schema.StringAttribute{
						MarkdownDescription: "VLAN base port UUID",
						Required:            true,
					},
					"tag": schema.Int64Attribute{
						MarkdownDescription: "VLAN ID",
						Required:            true,
					},
				},
				MarkdownDescription: "",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Port UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *EthernetPortResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *EthernetPortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data EthernetPortResourceModel

	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Use existing-, or create new REST API client
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for reading ethernet_port info
	var restInfo *interfaces.EthernetPortGetDataModelONTAP
	if data.ID.IsNull() {
		restInfo, err = interfaces.GetEthernetPortByName(
			errorHandler,
			*client,
			data.Name.ValueString(),
		)
		if err != nil {
			// error reporting done inside GetEthernetPortByName
			return
		}
	} else {
		restInfo, err = interfaces.GetEthernetPort(
			errorHandler,
			*client,
			data.ID.ValueString(),
		)
		if err != nil {
			// error reporting done inside GetEthernetPort
			return
		}
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError(
			"No Ethernet Port Found",
			fmt.Sprintf("No ethernet port '%s' found.", data.Name.ValueString()),
		)

		return
	}

	// Copy ethernet_port info to data source model
	data.BroadcastDomainID = types.StringValue(restInfo.BroadcastDomain.UUID)
	data.Name = types.StringValue(restInfo.Name)
	data.NodeID = types.StringValue(restInfo.Node.UUID)
	data.Type = types.StringValue(restInfo.Type)
	data.ID = types.StringValue(restInfo.UUID)

	switch data.Type.ValueString() {
	case "lag":
		// active_ports_id set
		var activePorts, memberPorts []attr.Value
		for _, v := range restInfo.LAG.ActivePorts {
			activePorts = append(activePorts, types.StringValue(v.UUID))
		}
		activePortsSet, diags := types.SetValue(types.StringType, activePorts)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// member_ports_id set
		for _, v := range restInfo.LAG.MemberPorts {
			memberPorts = append(memberPorts, types.StringValue(v.UUID))
		}
		memberPortsSet, diags := types.SetValue(types.StringType, memberPorts)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.LAG = &LAGResourceModel{
			ActivePortsID:      activePortsSet,
			DistributionPolicy: types.StringValue(restInfo.LAG.DistributionPolicy),
			MemberPortsID:      memberPortsSet,
			Mode:               types.StringValue(restInfo.LAG.Mode),
		}
	case "vlan":
		data.VLAN = &VLANResourceModel{
			BasePortID: types.StringValue(restInfo.VLAN.BasePort.UUID),
			Tag:        types.Int64Value(restInfo.VLAN.Tag),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Create a resource and retrieve UUID
func (r *EthernetPortResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *EthernetPortResourceModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Copy ethernet_port info to request body
	var body interfaces.EthernetPortResourceBodyDataModelONTAP
	body.BroadcastDomain = interfaces.EthernetPortBroadcastDomain{
		UUID: data.BroadcastDomainID.ValueString(),
	}
	body.Node = interfaces.EthernetPortNode{
		UUID: data.NodeID.ValueString(),
	}
	body.Type = data.Type.ValueString()

	switch body.Type {
	case "lag":
		// active_ports_id set
		activePorts := make([]types.String, 0, len(data.LAG.ActivePortsID.Elements()))
		diags := data.LAG.ActivePortsID.ElementsAs(ctx, &activePorts, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		activePortsList := make([]interfaces.EthernetPortLAGPort, 0, len(activePorts))
		for _, port := range activePorts {
			activePortsList = append(activePortsList, interfaces.EthernetPortLAGPort{
				UUID: port.ValueString(),
			})
		}

		// member_ports_id set
		memberPorts := make([]types.String, 0, len(data.LAG.MemberPortsID.Elements()))
		diags = data.LAG.MemberPortsID.ElementsAs(ctx, &memberPorts, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		memberPortsList := make([]interfaces.EthernetPortLAGPort, 0, len(memberPorts))
		for _, port := range memberPorts {
			memberPortsList = append(memberPortsList, interfaces.EthernetPortLAGPort{
				UUID: port.ValueString(),
			})
		}

		body.LAG = interfaces.EthernetPortLAG{
			ActivePorts:        activePortsList,
			DistributionPolicy: data.LAG.DistributionPolicy.ValueString(),
			MemberPorts:        memberPortsList,
			Mode:               data.LAG.Mode.ValueString(),
		}
	case "vlan":
		body.VLAN = interfaces.EthernetPortVLAN{
			BasePort: interfaces.EthernetPortVLANBasePort{
				UUID: data.VLAN.BasePortID.ValueString(),
			},
			Tag: data.VLAN.Tag.ValueInt64(),
		}
	}

	// Use existing-, or create new REST API client
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for creating broadcast_domain
	resource, err := interfaces.CreateEthernetPort(errorHandler, *client, body)
	if err != nil {
		return
	}

	// Copy ethernet_port info to data source model
	data.ID = types.StringValue(resource.UUID)
	data.Name = types.StringValue(resource.Name)

	tflog.Trace(ctx, fmt.Sprintf("created a resource, UUID=%s", data.ID))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *EthernetPortResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data /*, state*/ *EthernetPortResourceModel

	// 	// Read Terraform plan data into the model
	// 	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	// 	// Read state file data
	// 	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	// 	if resp.Diagnostics.HasError() {
	// 		return
	// 	}
	// 	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// 	// Copy broadcast_domain info to request body
	// 	var body interfaces.BroadcastDomainResourceBodyDataModelONTAP
	// 	if !data.IPSpace.Equal(state.IPSpace) {
	// 		body.IPspace.Name = data.IPSpace.ValueString()
	// 	}
	// 	if !data.Name.Equal(state.Name) {
	// 		body.Name = data.Name.ValueString()
	// 	}
	// 	if !data.MTU.Equal(state.MTU) {
	// 		body.MTU = data.MTU.ValueInt64()
	// 	}

	// 	// Use existing-, or create new REST API client
	// 	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	// 	if err != nil {
	// 		// error reporting done inside NewClient
	// 		return
	// 	}

	// 	// Call ONTAP REST API for updating broadcast_domain
	// 	err = interfaces.UpdateBroadcastDomain(errorHandler, *client, body, data.ID.ValueString())
	// 	if err != nil {
	// 		return
	// 	}

	tflog.Trace(ctx, fmt.Sprintf("updated a resource, UUID=%s", data.ID))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *EthernetPortResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *EthernetPortResourceModel

	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Use existing-, or create new REST API client
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Ensure that ID in known
	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("ID Is Null", "Ethernet port ID is null.")

		return
	}

	// Call ONTAP REST API for deleting broadcast_domain
	err = interfaces.DeleteEthernetPort(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a resource, UUID=%s", data.ID))
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *EthernetPortResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a network ethernet port resource: %#v", req))

	// 	// Extract broadcast_domain info from import identifier
	// 	idParts := strings.Split(req.ID, ",")
	// 	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
	// 		resp.Diagnostics.AddError(
	// 			"Unexpected Import Identifier",
	// 			fmt.Sprintf("Expected import identifier with format: cx_profile_name,ipspace,name, got: %q.", req.ID),
	// 		)

	// 		return
	// 	}

	// // Save broadcast_domain info to attributes
	// resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[0])...)
	// resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ipspace"), idParts[1])...)
	// resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[2])...)
}
