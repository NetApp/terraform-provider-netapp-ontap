package networking

import (
	"context"
	"fmt"
	"strconv"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &IPInterfaceDataSource{}

// NewIPInterfaceDataSource is a helper function to simplify the provider implementation.
func NewIPInterfaceDataSource() datasource.DataSource {
	return &IPInterfaceDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "network_ip_interface",
		},
	}
}

// IPInterfaceDataSource defines the data source implementation.
type IPInterfaceDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// IPInterfaceDataSourceModel describes the data source data model.
type IPInterfaceDataSourceModel struct {
	CxProfileName types.String             `tfsdk:"cx_profile_name"`
	Name          types.String             `tfsdk:"name"`
	SVMName       types.String             `tfsdk:"svm_name"`
	Scope         types.String             `tfsdk:"scope"`
	IP            *IPDataSourceModel       `tfsdk:"ip"`
	Location      *LocationDataSourceModel `tfsdk:"location"`
}

// IPDataSourceModel describes the data source model for IP address and mask.
type IPDataSourceModel struct {
	Address types.String `tfsdk:"address"`
	Netmask types.Int64  `tfsdk:"netmask"`
}

// LocationDataSourceModel describes the data source model for home node/port.
type LocationDataSourceModel struct {
	HomeNode types.String `tfsdk:"home_node"`
	HomePort types.String `tfsdk:"home_port"`
}

// Metadata returns the data source type name.
func (d *IPInterfaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *IPInterfaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "IPInterface data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Connection profile name",
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "IPInterface name",
			},
			"svm_name": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "IPInterface svm name",
			},
			"scope": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "IPInterface scope",
			},
			"ip": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"address": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "IPInterface IP address",
					},
					"netmask": schema.Int64Attribute{
						Computed:            true,
						MarkdownDescription: "IPInterface IP netmask",
					},
				},
				Computed: true,
			},
			"location": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"home_node": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "IPInterface home node",
					},
					"home_port": schema.StringAttribute{
						Computed:            true,
						MarkdownDescription: "IPInterface home port",
					},
				},
				Computed: true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *IPInterfaceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
	}
	d.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *IPInterfaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IPInterfaceDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	restInfo, err := interfaces.GetIPInterfaceByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetIPInterface
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No Interface found", fmt.Sprintf("NO interface, %s found.", data.Name.ValueString()))
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.Scope = types.StringValue(restInfo.Scope)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	intNetmask, err := strconv.Atoi(restInfo.IP.Netmask)
	if err != nil {
		errorHandler.MakeAndReportError("Failed to read ip interface", fmt.Sprintf("Error: failed to convert string value '%s' to int for net mask.", restInfo.IP.Netmask))
		return
	}
	data.IP = &IPDataSourceModel{
		Address: types.StringValue(restInfo.IP.Address),
		Netmask: types.Int64Value(int64(intNetmask)),
	}
	data.Location = &LocationDataSourceModel{
		HomeNode: types.StringValue(restInfo.Location.HomeNode.Name),
		HomePort: types.StringValue(restInfo.Location.HomePort.Name),
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
