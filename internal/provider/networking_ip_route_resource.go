package provider

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"

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

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &IPRouteResource{}
var _ resource.ResourceWithImportState = &IPRouteResource{}

// NewIPRouteResource is a helper function to simplify the provider implementation.
func NewIPRouteResource() resource.Resource {
	return &IPRouteResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "networking_ip_route_resource",
		},
	}
}

// IPRouteResource defines the resource implementation.
type IPRouteResource struct {
	config connection.ResourceOrDataSourceConfig
}

// IPRouteResourceModel describes the resource data model.
type IPRouteResourceModel struct {
	CxProfileName types.String                `tfsdk:"cx_profile_name"`
	SVMName       types.String                `tfsdk:"svm_name"`
	Destination   *DestinationDataSourceModel `tfsdk:"destination"`
	Gateway       types.String                `tfsdk:"gateway"`
	Metric        types.Int64                 `tfsdk:"metric"`
	ID            types.String                `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *IPRouteResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *IPRouteResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NetRoute resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"destination": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "destination IP address information",
				Computed:            true,
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"address": types.StringType,
						"netmask": types.StringType,
					},
					map[string]attr.Value{
						"address": types.StringValue("0.0.0.0"),
						"netmask": types.StringValue("0"),
					})),
				PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplace()},
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						MarkdownDescription: "IPv4 or IPv6 address",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("0.0.0.0"),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
					"netmask": schema.StringAttribute{
						MarkdownDescription: "netmask length (16) or IPv4 mask (255.255.0.0). For IPv6, valid range is 1 to 127.",
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("0"),
						PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
					},
				},
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface svm name",
				Optional:            true,
			},
			"gateway": schema.StringAttribute{
				MarkdownDescription: "The IP address of the gateway router leading to the destination.",
				Required:            true,
			},
			"metric": schema.Int64Attribute{
				MarkdownDescription: "Indicates a preference order between several routes to the same destination.",
				Optional:            true,
				Computed:            true,
				Default:             int64default.StaticInt64(20),
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "IP Route UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *IPRouteResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	}
	r.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *IPRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IPRouteResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
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
		errorHandler.MakeAndReportError("No cluster found", "No Cluster found")
		return
	}

	var restInfo *interfaces.IPRouteGetDataModelONTAP
	if data.Destination != nil {
		restInfo, err = interfaces.GetIPRoute(errorHandler, *client, data.Destination.Address.ValueString(), data.SVMName.ValueString(), data.Gateway.ValueString(), cluster.Version)
		if err != nil {
			// error reporting done inside GetIPInterface
			return
		}
	} else {
		restInfo, err = interfaces.GetIPRouteByGatewayAndSVM(errorHandler, *client, data.SVMName.ValueString(), data.Gateway.ValueString(), cluster.Version)
		if err != nil {
			// error reporting done inside GetIPInterface
			return
		}
	}

	if restInfo == nil {
		errorHandler.MakeAndReportError("No IP Route found", fmt.Sprintf("No IP Route %s found", data.Destination.Address.ValueString()))
		return
	}

	if data.Destination == nil {
		data.Destination = &DestinationDataSourceModel{}
	}
	data.Destination.Address = types.StringValue(restInfo.Destination.Address)
	data.Destination.Netmask = types.StringValue(restInfo.Destination.Netmask)
	data.Gateway = types.StringValue(restInfo.Gateway)
	data.Metric = types.Int64Value(restInfo.Metric)
	data.SVMName = types.StringValue(restInfo.SVMName.Name)
	data.ID = types.StringValue(restInfo.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *IPRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *IPRouteResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.IPRouteResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Destination != nil {
		if !data.Destination.Address.IsNull() {
			body.Destination.Address = data.Destination.Address.ValueString()
		}
		if !data.Destination.Netmask.IsNull() {
			body.Destination.Netmask = data.Destination.Netmask.ValueString()
		}
	}
	if !data.SVMName.IsNull() {
		body.SVM.Name = data.SVMName.ValueString()
	}
	if !data.Gateway.IsNull() {
		body.Gateway = data.Gateway.ValueString()
	}
	if !data.Metric.IsNull() {
		body.Metric = data.Metric.ValueInt64()
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateIPRoute(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.ID = types.StringValue(resource.UUID)

	tflog.Trace(ctx, fmt.Sprintf("created a resource, UUID=%s", data.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *IPRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *IPRouteResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Error(ctx, "Update not supported by REST API for IP Routes")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *IPRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *IPRouteResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "ip_interface UUID is null")
		return
	}

	err = interfaces.DeleteIPRoute(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *IPRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: svm_name,gateway,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("gateway"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
}
