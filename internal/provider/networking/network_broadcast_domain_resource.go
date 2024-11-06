package networking

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource                = &BroadcastDomainResource{}
	_ resource.ResourceWithConfigure   = &BroadcastDomainResource{}
	_ resource.ResourceWithImportState = &BroadcastDomainResource{}
)

// NewBroadcastDomainResource is a helper function to simplify the provider implementation.
func NewBroadcastDomainResource() resource.Resource {
	return &BroadcastDomainResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "network_broadcast_domain",
		},
	}
}

// BroadcastDomainResource defines the resource implementation.
type BroadcastDomainResource struct {
	config connection.ResourceOrDataSourceConfig
}

// BroadcastDomainResourceModel describes the resource data model.
type BroadcastDomainResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	IPSpace       types.String `tfsdk:"ipspace"`
	Name          types.String `tfsdk:"name"`
	MTU           types.Int64  `tfsdk:"mtu"`
	Ports         types.Set    `tfsdk:"ports"`
	ID            types.String `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *BroadcastDomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *BroadcastDomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Broadcast Domain resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"ipspace": schema.StringAttribute{
				MarkdownDescription: "Name of the IPspace the broadcast domain belongs to",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("Default"),
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the broadcast domain, scoped to its IPspace",
				Required:            true,
			},
			"mtu": schema.Int64Attribute{
				MarkdownDescription: "Maximum transmission unit, largest packet size on this network",
				Required:            true,
			},
			"ports": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Ports that belong to the broadcast domain",
				Computed:            true,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Broadcast domain UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *BroadcastDomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *BroadcastDomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data BroadcastDomainResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use existing-, or create new REST API client
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for reading broadcast_domain info
	var restInfo *interfaces.BroadcastDomainGetDataModelONTAP
	if data.ID.IsNull() {
		restInfo, err = interfaces.GetBroadcastDomainByName(
			errorHandler,
			*client,
			data.IPSpace.ValueString(),
			data.Name.ValueString(),
		)
		if err != nil {
			// error reporting done inside GetBroadcastDomainByName
			return
		}
	} else {
		restInfo, err = interfaces.GetBroadcastDomain(
			errorHandler,
			*client,
			data.ID.ValueString(),
		)
		if err != nil {
			// error reporting done inside GetBroadcastDomain
			return
		}
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError(
			"No Broadcast Domain Found",
			fmt.Sprintf("No broadcast domain '%s' found.", data.Name.ValueString()),
		)

		return
	}

	// Copy broadcast_domain info to data source model
	data.IPSpace = types.StringValue(restInfo.IPspace.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.MTU = types.Int64Value(restInfo.MTU)
	data.ID = types.StringValue(restInfo.UUID)

	var ports []attr.Value
	for _, v := range restInfo.Ports {
		ports = append(ports, types.StringValue(v.Name))
	}
	data.Ports, _ = types.SetValue(types.StringType, ports)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *BroadcastDomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *BroadcastDomainResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Copy broadcast_domain info to request body
	var body interfaces.BroadcastDomainResourceBodyDataModelONTAP
	body.IPspace.Name = data.IPSpace.ValueString()
	body.MTU = data.MTU.ValueInt64()
	body.Name = data.Name.ValueString()

	// Use existing-, or create new REST API client
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for creating broadcast_domain
	resource, err := interfaces.CreateBroadcastDomain(errorHandler, *client, body)
	if err != nil {
		return
	}

	// Copy broadcast_domain info to data source model
	data.ID = types.StringValue(resource.UUID)

	var ports []attr.Value
	for _, v := range resource.Ports {
		ports = append(ports, types.StringValue(v.Name))
	}
	data.Ports, _ = types.SetValue(types.StringType, ports)

	tflog.Trace(ctx, fmt.Sprintf("created a resource, UUID=%s", data.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *BroadcastDomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *BroadcastDomainResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	// Read state file data
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Copy broadcast_domain info to request body
	var body interfaces.BroadcastDomainResourceBodyDataModelONTAP
	if !data.IPSpace.Equal(state.IPSpace) {
		body.IPspace.Name = data.IPSpace.ValueString()
	}
	if !data.Name.Equal(state.Name) {
		body.Name = data.Name.ValueString()
	}
	if !data.MTU.Equal(state.MTU) {
		body.MTU = data.MTU.ValueInt64()
	}

	// Use existing-, or create new REST API client
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for updating broadcast_domain
	err = interfaces.UpdateBroadcastDomain(errorHandler, *client, body, data.ID.ValueString())
	if err != nil {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("updated a resource, UUID=%s", data.ID))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *BroadcastDomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *BroadcastDomainResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
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
		errorHandler.MakeAndReportError("ID Is Null", "Broadcast domain ID is null.")

		return
	}

	// Call ONTAP REST API for deleting broadcast_domain
	err = interfaces.DeleteBroadcastDomain(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

	tflog.Trace(ctx, fmt.Sprintf("deleted a resource, UUID=%s", data.ID))
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *BroadcastDomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a network broadcast domain resource: %#v", req))

	// Extract broadcast_domain info from import identifier
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: cx_profile_name,ipspace,name, got: %q.", req.ID),
		)

		return
	}

	// Save broadcast_domain info to attributes
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ipspace"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[2])...)
}
