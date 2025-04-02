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
var _ datasource.DataSource = &BroadcastDomainsDataSource{}

// NewBroadcastDomainDataSource is a helper function to simplify the provider implementation.
func NewBroadcastDomainsDataSource() datasource.DataSource {
	return &BroadcastDomainsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "network_broadcast_domains",
		},
	}
}

// BroadcastDomainsDataSource defines the data source implementation.
type BroadcastDomainsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// BroadcastDomainsDataSourceModel describes the data source data model.
type BroadcastDomainsDataSourceModel struct {
	CxProfileName    types.String                          `tfsdk:"cx_profile_name"`
	BroadcastDomains []BroadcastDomainDataSourceModel      `tfsdk:"broadcast_domains"`
	Filter           *BroadcastDomainDataSourceFilterModel `tfsdk:"filter"`
}

// BroadcastDomainDataSourceFilterModel describes the data source data model for queries.
type BroadcastDomainDataSourceFilterModel struct {
	IPspace types.String `tfsdk:"ipspace"`
	Name    types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *BroadcastDomainsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *BroadcastDomainsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Broadcast Domains data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"ipspace": schema.StringAttribute{
						MarkdownDescription: "Name of the IPspace the broadcast domain belongs to",
						Optional:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Name of the broadcast domain, scoped to its IPspace",
						Optional:            true,
					},
				},
				MarkdownDescription: "Filter broadcast domains by their properties",
				Optional:            true,
			},
			"broadcast_domains": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"ipspace": schema.StringAttribute{
							MarkdownDescription: "Name of the IPspace the broadcast domain belongs to",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the broadcast domain, scoped to its IPspace",
							Computed:            true,
						},
						"mtu": schema.Int64Attribute{
							MarkdownDescription: "Maximum transmission unit, largest packet size on this network",
							Computed:            true,
						},
						"ports": schema.SetAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "Ports that belong to the broadcast domain",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "Broadcast domain UUID",
							Computed:            true,
						},
					},
				},
				MarkdownDescription: "Broadcast domains matching the filter",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *BroadcastDomainsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *BroadcastDomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BroadcastDomainsDataSourceModel

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
	var filter *interfaces.BroadcastDomainDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.BroadcastDomainDataSourceFilterModel{
			IPspace: data.Filter.IPspace.ValueString(),
			Name:    data.Filter.Name.ValueString(),
		}
	}

	// Call ONTAP REST API for reading broadcast_domain info
	restInfo, err := interfaces.GetListBroadcastDomains(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetListBroadcastDomains
		return
	}

	// Copy broadcast_domain info to data source model
	data.BroadcastDomains = make([]BroadcastDomainDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		var ports []attr.Value
		for _, v := range record.Ports {
			ports = append(ports, types.StringValue(
				fmt.Sprintf("%s:%s", v.Node.Name, v.Name),
			))
		}
		portsSet, diags := types.SetValue(types.StringType, ports)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.BroadcastDomains[index] = BroadcastDomainDataSourceModel{
			IPSpace: types.StringValue(record.IPspace.Name),
			Name:    types.StringValue(record.Name),
			MTU:     types.Int64Value(record.MTU),
			Ports:   portsSet,
			ID:      types.StringValue(record.UUID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
