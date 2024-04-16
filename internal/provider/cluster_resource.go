package provider

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ClusterResource{}
var _ resource.ResourceWithImportState = &ClusterResource{}

// NewClusterResource is a helper function to simplify the provider implementation.
func NewClusterResource() resource.Resource {
	return &ClusterResource{
		config: resourceOrDataSourceConfig{
			name: "cluster_resource",
		},
	}
}

// ClusterResource defines the resource implementation.
type ClusterResource struct {
	config resourceOrDataSourceConfig
}

// ClusterResourceModel describes the resource data model.
type ClusterResourceModel struct {
	CxProfileName        types.String `tfsdk:"cx_profile_name"`
	Name                 types.String `tfsdk:"name"`
	Version              types.Object `tfsdk:"version"`
	Contact              types.String `tfsdk:"contact"`
	Location             types.String `tfsdk:"location"`
	License              types.Object `tfsdk:"license"`
	Password             types.String `tfsdk:"password"`
	DNSDomains           types.Set    `tfsdk:"dns_domains"`
	NameServers          types.Set    `tfsdk:"name_servers"`
	TimeZone             types.Object `tfsdk:"timezone"`
	Certificate          types.Object `tfsdk:"certificate"`
	NtpServers           types.Set    `tfsdk:"ntp_servers"`
	ManagementInterface  types.Object `tfsdk:"management_interface"`
	ManagementInterfaces types.Set    `tfsdk:"management_interfaces"`
	ID                   types.String `tfsdk:"id"`
}

// ClusterResourceVersion describes the Version data model.
type ClusterResourceVersion struct {
	Full types.String `tfsdk:"full"`
}

// ClusterResourceManagementInterface describes the ManagementInterface data model.
type ClusterResourceManagementInterface struct {
	Address types.String `tfsdk:"address"`
	Gateway types.String `tfsdk:"gateway"`
	Netmask types.String `tfsdk:"netmask"`
}

// ClusterResourceCertificate describes the Certificate data model.
type ClusterResourceCertificate struct {
	ID types.String `tfsdk:"id"`
}

// ClusterResourceTimezone describes the Timezone data model.
type ClusterResourceTimezone struct {
	Name types.String `tfsdk:"name"`
}

// ClusterResourceLicense describes the License data model.
type ClusterResourceLicense struct {
	Keys types.Set `tfsdk:"keys"`
}

// Metadata returns the resource type name.
func (r *ClusterResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *ClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cluster resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Cluster name",
			},
			"contact": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Contact information. Example: support@company.com",
			},
			"location": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Location information",
			},
			"version": schema.SingleNestedAttribute{ // read only
				Attributes: map[string]schema.Attribute{
					"full": schema.StringAttribute{
						MarkdownDescription: "ONTAP software version",
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Computed: true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "This returns the cluster version information. When the cluster has more than one node, the cluster version is equivalent to the lowest of generation, major, and minor versions on all nodes.",
			},
			"license": schema.SingleNestedAttribute{ //create only
				Attributes: map[string]schema.Attribute{
					"keys": schema.SetAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						MarkdownDescription: "list of license keys",
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Optional:            true,
				MarkdownDescription: "License keys or NLF contents.",
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.UseStateForUnknown(),
				},
			},
			"password": schema.StringAttribute{ // create only
				Optional:            true,
				Sensitive:           true,
				MarkdownDescription: "Password",
			},
			"dns_domains": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "A list of DNS domains.",
			},
			"name_servers": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "The list of IP addresses of the DNS servers. Addresses can be either IPv4 or IPv6 addresses.",
			},
			"timezone": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Name of the time zone",
					},
				},
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Time zone information.",
			},
			"certificate": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Optional: true,
						Computed: true,
					},
				},
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Certificate",
			},
			"ntp_servers": schema.SetAttribute{
				ElementType:         types.StringType,
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Host name, IPv4 address, or IPv6 address for the external NTP time servers.",
			},
			"management_interface": schema.SingleNestedAttribute{ //create only
				Attributes: map[string]schema.Attribute{
					"ip": schema.SingleNestedAttribute{
						Attributes: map[string]schema.Attribute{
							"address": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "IPv4 or IPv6 address",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"gateway": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "The IPv4 or IPv6 address of the default router.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
							"netmask": schema.StringAttribute{
								Optional:            true,
								MarkdownDescription: "Input as netmask length (16) or IPv4 mask (255.255.0.0). For IPv6, the default value is 64 with a valid range of 1 to 127. Output is always netmask length.",
								PlanModifiers: []planmodifier.String{
									stringplanmodifier.UseStateForUnknown(),
								},
							},
						},
						Optional:            true,
						MarkdownDescription: "Object to setup an interface along with its default router.",
						PlanModifiers: []planmodifier.Object{
							objectplanmodifier.UseStateForUnknown(),
						},
					},
				},
				Optional:            true,
				MarkdownDescription: "The management interface of the cluster. The subnet mask and gateway for this interface are used for the node management interfaces provided in the node configuration.",
			},
			"management_interfaces": schema.SetNestedAttribute{ //read only
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ip": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"address": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "IP address",
								},
							},
							Computed:            true,
							MarkdownDescription: "IP formation",
						},
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The name of the interface. If only the name is provided, the SVM scope must be provided by the object this object is embedded in.",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "The UUID that uniquely identifies the interface.",
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "A list of network interface",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the cluster",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ClusterResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClusterResourceModel

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

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("Cluster Not found", fmt.Sprintf("cluster %s not found.", data.Name))
		return
	}

	// ID, Name, Contact, Location
	data.ID = types.StringValue(cluster.ID)
	data.Name = types.StringValue(cluster.Name)
	data.Contact = types.StringValue(cluster.Contact)
	data.Location = types.StringValue(cluster.Location)

	//version
	elementTypes := map[string]attr.Type{
		"full": types.StringType,
	}
	objectElements := map[string]attr.Value{
		"full": types.StringValue(cluster.TimeZone.Name),
	}
	objectValue, diags := types.ObjectValue(elementTypes, objectElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Version = objectValue

	// dns domains
	elements := []attr.Value{}
	for _, dnsDomain := range cluster.DNSDomains {
		elements = append(elements, types.StringValue(dnsDomain))
	}
	setValue, diags := types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	data.DNSDomains = setValue

	//name servers
	elements = []attr.Value{}
	for _, nameServer := range cluster.NameServers {
		elements = append(elements, types.StringValue(nameServer))
	}
	setValue, diags = types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	data.NameServers = setValue

	// timezone
	elementTypes = map[string]attr.Type{
		"name": types.StringType,
	}
	objectElements = map[string]attr.Value{
		"name": types.StringValue(cluster.TimeZone.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, objectElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.TimeZone = objectValue

	// certificate
	elementTypes = map[string]attr.Type{
		"id": types.StringType,
	}
	objectElements = map[string]attr.Value{
		"id": types.StringValue(cluster.ClusterCertificate.ID),
	}
	objectValue, diags = types.ObjectValue(elementTypes, objectElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Certificate = objectValue

	// ntp servers
	elements = []attr.Value{}
	for _, ntpServer := range cluster.NtpServers {
		elements = append(elements, types.StringValue(ntpServer))
	}
	setValue, diags = types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	data.NtpServers = setValue

	// management interfaces
	setElements := []attr.Value{}
	for _, mgmInterface := range cluster.ManagementInterfaces {
		nestedElementTypes := map[string]attr.Type{
			"address": types.StringType,
		}
		nestedVolumeElements := map[string]attr.Value{
			"address": types.StringValue(mgmInterface.IP.Address),
		}
		originVolumeObjectValue, diags := types.ObjectValue(nestedElementTypes, nestedVolumeElements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		elementTypes := map[string]attr.Type{
			"ip":   types.ObjectType{AttrTypes: nestedElementTypes},
			"name": types.StringType,
			"id":   types.StringType,
		}
		elements := map[string]attr.Value{
			"ip":   originVolumeObjectValue,
			"name": types.StringValue(mgmInterface.Name),
			"id":   types.StringValue(mgmInterface.ID),
		}
		objectValue, diags := types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		setElements = append(setElements, objectValue)
	}
	setValue, diags = types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"ip": types.ObjectType{AttrTypes: map[string]attr.Type{
				"address": types.StringType,
			}},
			"name": types.StringType,
			"id":   types.StringType,
		},
	}, setElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.ManagementInterfaces = setValue

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *ClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ClusterResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	//requried fields
	body.Name = data.Name.ValueString()

	// password
	if data.Password.IsUnknown() {
		errorHandler.MakeAndReportError("Password is required for cluster create", "Attribute 'password' is missing when creating a cluster.")
		return
	}
	body.Password = data.Password.ValueString()

	//contact
	if !data.Contact.IsUnknown() {
		body.Contact = data.Contact.ValueString()
	}

	//location
	if !data.Location.IsUnknown() {
		body.Location = data.Location.ValueString()
	}

	//license
	if !data.License.IsUnknown() {
		var license ClusterResourceLicense
		diags := data.License.As(ctx, &license, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		keys := make([]types.String, 0, len(license.Keys.Elements()))
		diags = license.Keys.ElementsAs(ctx, &keys, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		keysList := make([]string, 0, len(keys))
		for _, key := range keys {
			keysList = append(keysList, key.ValueString())
		}
		body.License.Keys = keysList
	}

	// dns domains
	if !data.DNSDomains.IsUnknown() {
		var dnsDomains []string
		diags := data.DNSDomains.ElementsAs(ctx, &dnsDomains, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.DNSDomains = dnsDomains
	}

	//name servers
	if !data.NameServers.IsUnknown() {
		var nameServers []string
		diags := data.NameServers.ElementsAs(ctx, &nameServers, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.NameServers = nameServers
	}

	// timezone
	if !data.TimeZone.IsUnknown() {
		var timeZone ClusterResourceTimezone
		diags := data.TimeZone.As(ctx, &timeZone, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.TimeZone.Name = timeZone.Name.ValueString()
	}

	// certificate
	if !data.Certificate.IsUnknown() {
		var certificate ClusterResourceCertificate
		diags := data.Certificate.As(ctx, &certificate, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.ClusterCertificate.ID = certificate.ID.ValueString()
	}

	// ntp servers
	if !data.NtpServers.IsUnknown() {
		var ntpServers []string
		diags := data.NtpServers.ElementsAs(ctx, &ntpServers, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.NtpServers = ntpServers
	}

	// management interface
	if !data.ManagementInterface.IsUnknown() {
		var mgmtInterface ClusterResourceManagementInterface
		diags := data.ManagementInterface.As(ctx, &mgmtInterface, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !mgmtInterface.Address.IsUnknown() {
			body.ManagementInterface.IP.Address = mgmtInterface.Address.ValueString()
		}
		if !mgmtInterface.Gateway.IsUnknown() {
			body.ManagementInterface.IP.Gateway = mgmtInterface.Gateway.ValueString()
		}
		if !mgmtInterface.Netmask.IsUnknown() {
			body.ManagementInterface.IP.Netmask = mgmtInterface.Netmask.ValueString()
		}
	}

	// Create API is async
	err = interfaces.CreateCluster(errorHandler, *client, body)
	if err != nil {
		// error reporting done inside CreateCluster
		return
	}

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("Cluster Not found", fmt.Sprintf("cluster %s not found.", data.Name))
		return
	}

	// ID, Name, Contact, Location
	data.ID = types.StringValue(cluster.ID)
	data.Name = types.StringValue(cluster.Name)
	data.Contact = types.StringValue(cluster.Contact)
	data.Location = types.StringValue(cluster.Location)

	//version
	elementTypes := map[string]attr.Type{
		"full": types.StringType,
	}
	objectElements := map[string]attr.Value{
		"full": types.StringValue(cluster.TimeZone.Name),
	}
	objectValue, diags := types.ObjectValue(elementTypes, objectElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Version = objectValue

	// dns domains
	elements := []attr.Value{}
	for _, dnsDomain := range cluster.DNSDomains {
		elements = append(elements, types.StringValue(dnsDomain))
	}
	setValue, diags := types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	data.DNSDomains = setValue

	//name servers
	elements = []attr.Value{}
	for _, nameServer := range cluster.NameServers {
		elements = append(elements, types.StringValue(nameServer))
	}
	setValue, diags = types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	data.NameServers = setValue

	// timezone
	elementTypes = map[string]attr.Type{
		"name": types.StringType,
	}
	objectElements = map[string]attr.Value{
		"name": types.StringValue(cluster.TimeZone.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, objectElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.TimeZone = objectValue

	// certificate
	elementTypes = map[string]attr.Type{
		"id": types.StringType,
	}
	objectElements = map[string]attr.Value{
		"id": types.StringValue(cluster.ClusterCertificate.ID),
	}
	objectValue, diags = types.ObjectValue(elementTypes, objectElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Certificate = objectValue

	// ntp servers
	elements = []attr.Value{}
	for _, ntpServer := range cluster.NtpServers {
		elements = append(elements, types.StringValue(ntpServer))
	}
	setValue, diags = types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	data.NtpServers = setValue

	// management interfaces
	setElements := []attr.Value{}
	for _, mgmInterface := range cluster.ManagementInterfaces {
		nestedElementTypes := map[string]attr.Type{
			"address": types.StringType,
		}
		nestedVolumeElements := map[string]attr.Value{
			"address": types.StringValue(mgmInterface.IP.Address),
		}
		originVolumeObjectValue, diags := types.ObjectValue(nestedElementTypes, nestedVolumeElements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		elementTypes := map[string]attr.Type{
			"ip":   types.ObjectType{AttrTypes: nestedElementTypes},
			"name": types.StringType,
			"id":   types.StringType,
		}
		elements := map[string]attr.Value{
			"ip":   originVolumeObjectValue,
			"name": types.StringValue(mgmInterface.Name),
			"id":   types.StringValue(mgmInterface.ID),
		}
		objectValue, diags := types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		setElements = append(setElements, objectValue)
	}

	setValue, diags = types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"ip": types.ObjectType{AttrTypes: map[string]attr.Type{
				"address": types.StringType,
			}},
			"name": types.StringType,
			"id":   types.StringType,
		},
	}, setElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.ManagementInterfaces = setValue

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *ClusterResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	client, err := getRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	var body interfaces.ClusterResourceBodyDataModelONTAP
	if !plan.Contact.IsUnknown() {
		body.Contact = plan.Contact.ValueString()
	}
	if !plan.Location.IsUnknown() {
		body.Location = plan.Location.ValueString()
	}
	if !plan.DNSDomains.IsUnknown() {
		var dnsDomains []string
		diags := plan.DNSDomains.ElementsAs(ctx, &dnsDomains, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.DNSDomains = dnsDomains
	}
	if !plan.NameServers.IsUnknown() {
		var nameServers []string
		diags := plan.NameServers.ElementsAs(ctx, &nameServers, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.NameServers = nameServers
	}
	if !plan.TimeZone.IsUnknown() {
		var timeZone ClusterResourceTimezone
		diags := plan.TimeZone.As(ctx, &timeZone, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.TimeZone.Name = timeZone.Name.ValueString()
	}
	if !plan.Certificate.IsUnknown() {
		var certificate ClusterResourceCertificate
		diags := plan.Certificate.As(ctx, &certificate, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.ClusterCertificate.ID = certificate.ID.ValueString()
	}
	if !plan.NtpServers.IsUnknown() {
		var ntpServers []string
		diags := plan.NtpServers.ElementsAs(ctx, &ntpServers, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.NtpServers = ntpServers
	}

	// Update API is async
	err = interfaces.UpdateCluster(errorHandler, *client, body)
	if err != nil {
		// error reporting done inside UpdateCluster
		return
	}

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("Cluster Not found", fmt.Sprintf("cluster %s not found.", plan.Name))
		return
	}

	plan.Name = types.StringValue(cluster.Name)
	plan.Contact = types.StringValue(cluster.Contact)
	plan.Location = types.StringValue(cluster.Location)

	//version
	elementTypes := map[string]attr.Type{
		"full": types.StringType,
	}
	objectElements := map[string]attr.Value{
		"full": types.StringValue(cluster.TimeZone.Name),
	}
	objectValue, diags := types.ObjectValue(elementTypes, objectElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	plan.Version = objectValue

	// dns domains
	elements := []attr.Value{}
	for _, dnsDomain := range cluster.DNSDomains {
		elements = append(elements, types.StringValue(dnsDomain))
	}
	setValue, diags := types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	plan.DNSDomains = setValue

	//name servers
	elements = []attr.Value{}
	for _, nameServer := range cluster.NameServers {
		elements = append(elements, types.StringValue(nameServer))
	}
	setValue, diags = types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	plan.NameServers = setValue
	// time zone
	elementTypes = map[string]attr.Type{
		"name": types.StringType,
	}
	objectElements = map[string]attr.Value{
		"name": types.StringValue(cluster.TimeZone.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, objectElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	plan.TimeZone = objectValue

	// certificate
	elementTypes = map[string]attr.Type{
		"id": types.StringType,
	}
	objectElements = map[string]attr.Value{
		"id": types.StringValue(cluster.ClusterCertificate.ID),
	}
	objectValue, diags = types.ObjectValue(elementTypes, objectElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	plan.Certificate = objectValue

	// ntp servers
	elements = []attr.Value{}
	for _, ntpServer := range cluster.NtpServers {
		elements = append(elements, types.StringValue(ntpServer))
	}
	setValue, diags = types.SetValue(types.StringType, elements)
	resp.Diagnostics.Append(diags...)
	plan.NtpServers = setValue

	// management interfaces
	setElements := []attr.Value{}
	for _, mgmInterface := range cluster.ManagementInterfaces {
		nestedElementTypes := map[string]attr.Type{
			"address": types.StringType,
		}
		nestedVolumeElements := map[string]attr.Value{
			"address": types.StringValue(mgmInterface.IP.Address),
		}
		originVolumeObjectValue, diags := types.ObjectValue(nestedElementTypes, nestedVolumeElements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		elementTypes := map[string]attr.Type{
			"ip":   types.ObjectType{AttrTypes: nestedElementTypes},
			"name": types.StringType,
			"id":   types.StringType,
		}
		elements := map[string]attr.Value{
			"ip":   originVolumeObjectValue,
			"name": types.StringValue(mgmInterface.Name),
			"id":   types.StringValue(mgmInterface.ID),
		}
		objectValue, diags := types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		setElements = append(setElements, objectValue)
	}

	setValue, diags = types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"ip": types.ObjectType{AttrTypes: map[string]attr.Type{
				"address": types.StringType,
			}},
			"name": types.StringType,
			"id":   types.StringType,
		},
	}, setElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	plan.ManagementInterfaces = setValue

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	errorHandler.MakeAndReportError("Update not available", "No update can be done on flexcache resource.")

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a cluster resource: %#v", req))
	idParts := strings.Split(req.ID, ",")
	log.Printf("idParts: %v", idParts)
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[1])...)
}
