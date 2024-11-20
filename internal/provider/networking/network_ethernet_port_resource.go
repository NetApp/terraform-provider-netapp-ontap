package networking

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
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
	BroadcastDomain *EthernetPortBroadcastDomainResourceModel `tfsdk:"broadcast_domain"`
	CxProfileName   types.String                              `tfsdk:"cx_profile_name"`
	Enabled         types.Bool                                `tfsdk:"enabled"`
	LAG             *LAGResourceModel                         `tfsdk:"lag"`
	Name            types.String                              `tfsdk:"name"`
	Node            *EthernetPortNodeResourceModel            `tfsdk:"node"`
	Reachability    types.String                              `tfsdk:"reachability"`
	State           types.String                              `tfsdk:"state"`
	Type            types.String                              `tfsdk:"type"`
	VLAN            *VLANResourceModel                        `tfsdk:"vlan"`
	ID              types.String                              `tfsdk:"id"`

	// InterfaceCount types.Int64  `tfsdk:"interface_count"`
	// MACAddress     types.String `tfsdk:"mac_address"`
	// MTU            types.Int64  `tfsdk:"mtu"`
	// RDMAProtocols  types.Set    `tfsdk:"rdma_protocols"`
	// Speed          types.Int64  `tfsdk:"speed"`
}

// EthernetPortBroadcastDomainResourceModel describes the resource model for broadcast domains.
type EthernetPortBroadcastDomainResourceModel struct {
	ID      types.String `tfsdk:"id"`
	IPSpace types.String `tfsdk:"ipspace"`
	Name    types.String `tfsdk:"name"`
}

// LAGResourceModel describes the resource model for LAG ports, policy and mode.
type LAGResourceModel struct {
	ActivePorts        types.Set    `tfsdk:"active_ports"`
	DistributionPolicy types.String `tfsdk:"distribution_policy"`
	MemberPorts        types.Set    `tfsdk:"member_ports"`
	Mode               types.String `tfsdk:"mode"`
}

// EthernetPortNodeResourceModel describes the resource model for nodes.
type EthernetPortNodeResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

// VLANResourceModel describes the resource model for VLAN base port and tag.
type VLANResourceModel struct {
	BasePort types.String `tfsdk:"base_port"`
	Tag      types.Int64  `tfsdk:"tag"`
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
			"broadcast_domain": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"ipspace": schema.StringAttribute{
						MarkdownDescription: "Name of the broadcast domain's IPspace",
						Optional:            true,
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Name of the broadcast domain, scoped to its IPspace",
						Optional:            true,
						Computed:            true,
					},
					"id": schema.StringAttribute{
						MarkdownDescription: "Broadcast domain UUID",
						Optional:            true,
						Computed:            true,
					},
				},
				MarkdownDescription: "Broadcast domain properties",
				Optional:            true,
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Port enabled",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
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
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"member_ports": schema.SetAttribute{
						MarkdownDescription: "Array of ports belonging to the LAG, regardless of their state",
						ElementType:         types.StringType,
						Required:            true,
					},
					"mode": schema.StringAttribute{
						MarkdownDescription: "Determines how the ports interact with the switch",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
				},
				MarkdownDescription: "LAG (ifgrp) properties",
				Optional:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Portname, such as e0a, e1b-100 (VLAN on Ethernet), a0c (LAG/ifgrp), a0d-200 (VLAN on LAG/ifgrp), e0a.pv1 (p-VLAN, in select environments only)",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"node": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Name of the node on which the port is located",
						Optional:            true,
						Computed:            true,
					},
					"id": schema.StringAttribute{
						MarkdownDescription: "Node UUID",
						Optional:            true,
						Computed:            true,
					},
				},
				MarkdownDescription: "Node properties",
				Required:            true,
			},
			"reachability": schema.StringAttribute{
				MarkdownDescription: "Reachability status of the port",
				Computed:            true,
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "Operational state of the port",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of physical or virtual port",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vlan": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"base_port": schema.StringAttribute{
						MarkdownDescription: "VLAN base port",
						Required:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.RequiresReplace(),
						},
					},
					"tag": schema.Int64Attribute{
						MarkdownDescription: "VLAN ID",
						Required:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.RequiresReplace(),
						},
					},
				},
				MarkdownDescription: "VLAN properties",
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

	// Copy ethernet_port info to resource model
	data.BroadcastDomain.ID = types.StringValue(restInfo.BroadcastDomain.UUID)
	data.BroadcastDomain.IPSpace = types.StringValue(restInfo.BroadcastDomain.IPSpace.Name)
	data.BroadcastDomain.Name = types.StringValue(restInfo.BroadcastDomain.Name)
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.Name = types.StringValue(restInfo.Name)
	data.Node.ID = types.StringValue(restInfo.Node.UUID)
	data.Node.Name = types.StringValue(restInfo.Node.Name)
	data.Reachability = types.StringValue(restInfo.Reachability)
	data.State = types.StringValue(restInfo.State)
	data.Type = types.StringValue(restInfo.Type)
	data.ID = types.StringValue(restInfo.UUID)

	switch data.Type.ValueString() {
	case "lag":
		data.LAG.DistributionPolicy = types.StringValue(restInfo.LAG.DistributionPolicy)
		data.LAG.Mode = types.StringValue(restInfo.LAG.Mode)

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
		data.LAG.ActivePorts = activePortsSet

		// member_ports set
		for _, v := range restInfo.LAG.MemberPorts {
			memberPorts = append(memberPorts, types.StringValue(v.Name))
		}
		memberPortsSet, diags := types.SetValue(types.StringType, memberPorts)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.LAG.MemberPorts = memberPortsSet
	case "vlan":
		data.VLAN.BasePort = types.StringValue(restInfo.VLAN.BasePort.Name)
		data.VLAN.Tag = types.Int64Value(restInfo.VLAN.Tag)
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
		IPSpace: interfaces.EthernetPortIPSpace{
			Name: data.BroadcastDomain.IPSpace.ValueString(),
		},
		Name: data.BroadcastDomain.Name.ValueString(),
		UUID: data.BroadcastDomain.ID.ValueString(),
	}
	body.Enabled = data.Enabled.ValueBool()
	body.Node = interfaces.EthernetPortNode{
		Name: data.Node.Name.ValueString(),
		UUID: data.Node.ID.ValueString(),
	}
	body.Type = data.Type.ValueString()

	switch body.Type {
	case "lag":
		// member_ports set
		memberPorts := make([]types.String, 0, len(data.LAG.MemberPorts.Elements()))
		diags = data.LAG.MemberPorts.ElementsAs(ctx, &memberPorts, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		memberPortsList := make([]interfaces.EthernetPortLAGPort, 0, len(memberPorts))
		for _, port := range memberPorts {
			memberPortsList = append(memberPortsList, interfaces.EthernetPortLAGPort{
				Name: port.ValueString(),
			})
		}

		body.LAG = interfaces.EthernetPortLAG{
			DistributionPolicy: data.LAG.DistributionPolicy.ValueString(),
			MemberPorts:        memberPortsList,
			Mode:               data.LAG.Mode.ValueString(),
		}
	case "vlan":
		body.VLAN = interfaces.EthernetPortVLAN{
			BasePort: interfaces.EthernetPortVLANBasePort{
				Name: data.VLAN.BasePort.ValueString(),
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

	// Copy ethernet_port info to resource model
	data.BroadcastDomain = &EthernetPortBroadcastDomainResourceModel{
		ID:      types.StringValue(resource.BroadcastDomain.UUID),
		IPSpace: types.StringValue(resource.BroadcastDomain.IPSpace.Name),
		Name:    types.StringValue(resource.BroadcastDomain.Name),
	}
	data.Enabled = types.BoolValue(resource.Enabled)
	data.Node = &EthernetPortNodeResourceModel{
		ID:   types.StringValue(resource.Node.UUID),
		Name: types.StringValue(resource.Node.Name),
	}
	data.ID = types.StringValue(resource.UUID)

	// TODO: interfaces.GetEthernetPort to populate remaining attributes

	// ONTAP API POST response does not contain the following attributes.
	// We still include them in the data model, and read them here, so that
	// subsequent GET request (terraform refresh) can populate them.
	data.Name = types.StringValue(resource.Name)
	data.Reachability = types.StringValue(resource.Reachability)
	data.State = types.StringValue(resource.State)

	if body.Type == "lag" {
		// active_ports set
		var activePorts []attr.Value
		for _, v := range resource.LAG.ActivePorts {
			activePorts = append(activePorts, types.StringValue(v.Name))
		}
		activePortsSet, diags := types.SetValue(types.StringType, activePorts)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.LAG.ActivePorts = activePortsSet
	}

	tflog.Trace(ctx, fmt.Sprintf("created a resource, UUID=%s", data.ID))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *EthernetPortResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *EthernetPortResourceModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read state file data
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Copy ethernet_port info to request body
	var body interfaces.EthernetPortResourceBodyDataModelONTAP

	// Ports can be enabled / disabled in-place
	body.Enabled = data.Enabled.ValueBool()

	// broadcast_domain can be updated in-place
	if !data.BroadcastDomain.IPSpace.Equal(state.BroadcastDomain.IPSpace) {
		body.BroadcastDomain.IPSpace.Name = data.BroadcastDomain.IPSpace.ValueString()
	}
	if !data.BroadcastDomain.Name.Equal(state.BroadcastDomain.Name) {
		body.BroadcastDomain.Name = data.BroadcastDomain.Name.ValueString()
	}
	if !data.BroadcastDomain.ID.Equal(state.BroadcastDomain.ID) {
		body.BroadcastDomain.UUID = data.BroadcastDomain.ID.ValueString()
	}

	// node can be updated in-place
	if !data.Node.Name.Equal(state.Node.Name) {
		body.Node.Name = data.Node.Name.ValueString()
	}
	if !data.Node.ID.Equal(state.Node.ID) {
		body.Node.UUID = data.Node.ID.ValueString()
	}

	switch data.Type.ValueString() {
	case "lag":
		// For LAG, member_ports can be updated in-place
		memberPorts := make([]types.String, 0, len(data.LAG.MemberPorts.Elements()))
		diags = data.LAG.MemberPorts.ElementsAs(ctx, &memberPorts, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		memberPortsList := make([]interfaces.EthernetPortLAGPort, 0, len(memberPorts))
		for _, port := range memberPorts {
			memberPortsList = append(memberPortsList, interfaces.EthernetPortLAGPort{
				Name: port.ValueString(),
			})
		}
		body.LAG.MemberPorts = memberPortsList
	case "vlan":
		// For VLAN, no attributes can be updated in-place
	}

	// Use existing-, or create new REST API client
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for updating ethernet_port
	err = interfaces.UpdateEthernetPort(
		errorHandler,
		*client,
		body,
		data.ID.ValueString(),
		data.Type.ValueString(),
	)
	if err != nil {
		return
	}

	// Call ONTAP REST API again for reading ethernet_port info
	restInfo, err := interfaces.GetEthernetPort(
		errorHandler,
		*client,
		data.ID.ValueString(),
	)
	if err != nil {
		return
	}

	// Copy ethernet_port info to resource model
	data.BroadcastDomain.ID = types.StringValue(restInfo.BroadcastDomain.UUID)
	data.BroadcastDomain.IPSpace = types.StringValue(restInfo.BroadcastDomain.IPSpace.Name)
	data.BroadcastDomain.Name = types.StringValue(restInfo.BroadcastDomain.Name)
	data.Node.ID = types.StringValue(restInfo.Node.UUID)
	data.Node.Name = types.StringValue(restInfo.Node.Name)
	data.Reachability = types.StringValue(restInfo.Reachability)
	data.State = types.StringValue(restInfo.State)

	if data.Type.ValueString() == "lag" {
		// active_ports set
		var activePorts []attr.Value
		for _, v := range restInfo.LAG.ActivePorts {
			activePorts = append(activePorts, types.StringValue(v.Name))
		}
		activePortsSet, diags := types.SetValue(types.StringType, activePorts)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.LAG.ActivePorts = activePortsSet
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a resource, UUID=%s", data.ID))

	// Save updated data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
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

	// Call ONTAP REST API for deleting ethernet_port
	err = interfaces.DeleteEthernetPort(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a resource, UUID=%s", data.ID))
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *EthernetPortResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a network ethernet port resource: %#v", req))

	// 	// Extract ethernet_port info from import identifier
	// 	idParts := strings.Split(req.ID, ",")
	// 	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
	// 		resp.Diagnostics.AddError(
	// 			"Unexpected Import Identifier",
	// 			fmt.Sprintf("Expected import identifier with format: cx_profile_name,ipspace,name, got: %q.", req.ID),
	// 		)

	// 		return
	// 	}

	// // Save ethernet_port info to attributes
	// resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[0])...)
	// resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ipspace"), idParts[1])...)
	// resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[2])...)
}
