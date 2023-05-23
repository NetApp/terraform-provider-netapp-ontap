package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &NameServicesDNSDataSource{}

// NewNameServicesDNSDataSource is a helper function to simplify the provider implementation.
func NewNameServicesDNSDataSource() datasource.DataSource {
	return &NameServicesDNSDataSource{
		config: resourceOrDataSourceConfig{
			name: "name_services_dns_data_source",
		},
	}
}

// NameServicesDNSDataSource defines the data source implementation.
type NameServicesDNSDataSource struct {
	config resourceOrDataSourceConfig
}

// NameServicesDNSDataSourceModel describes the data source data model.
type NameServicesDNSDataSourceModel struct {
	CxProfileName types.String   `tfsdk:"cx_profile_name"`
	SVMName       types.String   `tfsdk:"svm_name"`
	SVMUUID       types.String   `tfsdk:"svm_uuid"`
	Domains       []types.String `tfsdk:"dns_domains"`
	NameServers   []types.String `tfsdk:"name_servers"`
}

// NameServicesDNSDataSourceFilterModel describes the data source data model for queries.
type NameServicesDNSDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm.name"`
}

// Metadata returns the data source type name.
func (d *NameServicesDNSDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *NameServicesDNSDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesDNS data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface vserver name",
				Required:            true,
			},
			"svm_uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of Vserver",
				Computed:            true,
			},
			"dns_domains": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of DNS domains such as 'sales.bar.com'. The first domain is the one that the Vserver belongs to",
			},
			"name_servers": schema.ListAttribute{
				ElementType:         types.StringType,
				Computed:            true,
				MarkdownDescription: "List of IPv4 addresses of name servers such as '123.123.123.123'.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NameServicesDNSDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	d.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *NameServicesDNSDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NameServicesDNSDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	restInfo, err := interfaces.GetNameServicesDNS(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetNameServicesDNS
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("read a data source rest: %#v", restInfo))
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.SVMUUID = types.StringValue(restInfo.SVM.UUID)
	var servers []types.String
	for _, v := range restInfo.Servers {
		servers = append(data.NameServers, types.StringValue(v))
	}
	data.NameServers = servers
	var domains []types.String
	for _, v := range restInfo.Domains {
		domains = append(data.Domains, types.StringValue(v))
	}
	data.Domains = domains

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
