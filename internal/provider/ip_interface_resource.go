package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you resource (should match internal/provider/ip_interface_resource.go)
// replace IPInterface with the name of the resource, following go conventions, eg IPInterface
// replace ip_interface with the name of the resource, for logging purposes, eg ip_interface
// make sure to create internal/interfaces/ip_interface.go too)
// delete these 5 lines

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &IPInterfaceResource{}
var _ resource.ResourceWithImportState = &IPInterfaceResource{}

// NewIPInterfaceResource is a helper function to simplify the provider implementation.
func NewIPInterfaceResource() resource.Resource {
	return &IPInterfaceResource{
		config: resourceOrDataSourceConfig{
			name: "ip_interface_resource",
		},
	}
}

// IPInterfaceResource defines the resource implementation.
type IPInterfaceResource struct {
	config resourceOrDataSourceConfig
}

// IPInterfaceResourceIP describes the resource data model for IP address and mask.
type IPInterfaceResourceIP struct {
	Address types.String `tfsdk:"address"`
	Netmask types.Int64  `tfsdk:"netmask"`
}

// IPInterfaceResourceLocation describes the resource data model for home node/port.
type IPInterfaceResourceLocation struct {
	HomeNode types.String `tfsdk:"home_node"`
	HomePort types.String `tfsdk:"home_port"`
}

// IPInterfaceResourceModel describes the resource data model.
type IPInterfaceResourceModel struct {
	CxProfileName types.String                 `tfsdk:"cx_profile_name"`
	Name          types.String                 `tfsdk:"name"`
	SVMName       types.String                 `tfsdk:"svm_name"`
	IP            *IPInterfaceResourceIP       `tfsdk:"ip"`
	Location      *IPInterfaceResourceLocation `tfsdk:"location"`
	UUID          types.String                 `tfsdk:"uuid"`
}

// Metadata returns the resource type name.
func (r *IPInterfaceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *IPInterfaceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "IPInterface resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "IPInterface name",
				Required:            true,
			},
			// TODO: Make svm_name optional for cluster scoped interface
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface vserver name",
				Required:            true,
			},
			// TODO: Make IP optional once subnet is supported
			"ip": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "IPInterface IP address",
						Required:            true,
					},
					"netmask": schema.Int64Attribute{
						MarkdownDescription: "IPInterface IP netmask",
						Required:            true,
					},
				},
				Required: true,
			},
			// TODO: Make location fields optionals once other options are supported
			"location": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"home_node": schema.StringAttribute{
						MarkdownDescription: "IPInterface home node",
						Required:            true,
					},
					"home_port": schema.StringAttribute{
						MarkdownDescription: "IPInterface home port",
						Required:            true,
					},
				},
				Required: true,
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "IPInterface UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *IPInterfaceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *IPInterfaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IPInterfaceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	restInfo, err := interfaces.GetIPInterface(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetIPInterface
		return
	}

	data.Name = types.StringValue(restInfo.Name)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *IPInterfaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *IPInterfaceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.IPInterfaceResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	// TODO: check for empty values for optional fields
	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()
	body.IP.Address = data.IP.Address.ValueString()
	body.IP.Netmask = data.IP.Netmask.ValueInt64()
	body.Location.HomePort = &interfaces.IPInterfaceResourceHomePort{
		Name: data.Location.HomePort.ValueString(),
		Node: interfaces.IPInterfaceResourceHomeNode{
			Name: data.Location.HomeNode.ValueString(),
		},
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateIPInterface(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.UUID = types.StringValue(resource.UUID)

	tflog.Trace(ctx, fmt.Sprintf("created a resource, UUID=%s", data.UUID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *IPInterfaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *IPInterfaceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *IPInterfaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *IPInterfaceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.UUID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "ip_interface UUID is null")
		return
	}

	err = interfaces.DeleteIPInterface(errorHandler, *client, data.UUID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *IPInterfaceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
