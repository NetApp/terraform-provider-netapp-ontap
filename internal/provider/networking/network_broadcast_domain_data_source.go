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
var _ datasource.DataSource = &BroadcastDomainDataSource{}

// NewBroadcastDomainDataSource is a helper function to simplify the provider implementation.
func NewBroadcastDomainDataSource() datasource.DataSource {
	return &BroadcastDomainDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "network_broadcast_domain",
		},
	}
}

// BroadcastDomainDataSource defines the data source implementation.
type BroadcastDomainDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// BroadcastDomainDataSourceModel describes the data source data model.
type BroadcastDomainDataSourceModel struct {
	CxProfileName types.String   `tfsdk:"cx_profile_name"`
	IPSpace       types.String   `tfsdk:"ipspace"`
	Name          types.String   `tfsdk:"name"`
	MTU           types.Int64    `tfsdk:"mtu"`
	Ports         []types.String `tfsdk:"ports"`
	ID            types.String   `tfsdk:"id"`
}

// Metadata returns the data source type name.
func (d *BroadcastDomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *BroadcastDomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Broadcast Domain data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"ipspace": schema.StringAttribute{
				MarkdownDescription: "Name of the IPspace",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the broadcast domain, scoped to its IPspace",
				Required:            true,
			},
			"mtu": schema.Int64Attribute{
				MarkdownDescription: "Maximum transmission unit, largest packet size on this network",
				Computed:            true,
			},
			"ports": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Ports that belong to the broadcast domain",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Broadcast domain UUID",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *BroadcastDomainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *BroadcastDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BroadcastDomainDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Use existing-, or create new REST API client
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for reading broadcast_domain info
	restInfo, err := interfaces.GetBroadcastDomainByName(
		errorHandler,
		*client,
		data.IPSpace.ValueString(),
		data.Name.ValueString(),
	)
	if err != nil {
		// error reporting done inside GetBroadcastDomainByName
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError(
			"No Broadcast Domain found",
			fmt.Sprintf("No broadcast-domain '%s' found.", data.Name.ValueString()))
		return
	}

	// Copy broadcast_domain info to data source model
	var ports []types.String
	for _, v := range restInfo.Ports {
		ports = append(ports, types.StringValue(v.Name))
	}
	data.IPSpace = types.StringValue(restInfo.IPspace.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.MTU = types.Int64Value(restInfo.MTU)
	data.Ports = ports
	data.ID = types.StringValue(restInfo.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
