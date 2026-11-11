package networking

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &IPspacesDataSource{}

// NewIPspacesDataSource is a helper function to simplify the provider implementation.
func NewIPspacesDataSource() datasource.DataSource {
	return &IPspacesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "network_ipspaces",
		},
	}
}

// IPspacesDataSource defines the data source implementation.
type IPspacesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// IPspacesDataSourceModel describes the data source data model.
type IPspacesDataSourceModel struct {
	CxProfileName  types.String                    `tfsdk:"cx_profile_name"`
	IPspaces       []IPspaceDataSourceModel        `tfsdk:"ipspaces"`
	Filter         *IPspacesDataSourceFilterModel  `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *IPspacesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *IPspacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "IPspaces data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "The name of the IPspace.",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"ipspaces": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the IPspace.",
							Required:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "The UUID of the IPspace.",
							Computed:            true,
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "List of IPspaces",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *IPspacesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *IPspacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IPspacesDataSourceModel

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

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	filterName := ""
	if data.Filter != nil && !data.Filter.Name.IsNull() && !data.Filter.Name.IsUnknown() {
		filterName = data.Filter.Name.ValueString()
	}

	restInfo, err := interfaces.GetIPspaces(errorHandler, *client, filterName)
	if err != nil {
		// error reporting done inside GetIPspaces
		return
	}

	data.IPspaces = make([]IPspaceDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.IPspaces[index] = IPspaceDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
			UUID:          types.StringValue(record.UUID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
