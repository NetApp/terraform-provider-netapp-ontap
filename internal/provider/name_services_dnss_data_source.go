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
var _ datasource.DataSource = &NameServicesDNSsDataSource{}

// NewNameServicesDNSsDataSource is a helper function to simplify the provider implementation.
func NewNameServicesDNSsDataSource() datasource.DataSource {
	return &NameServicesDNSsDataSource{
		config: resourceOrDataSourceConfig{
			name: "name_services_dnss_data_source",
		},
	}
}

// NameServicesDNSsDataSource defines the data source implementation.
type NameServicesDNSsDataSource struct {
	config resourceOrDataSourceConfig
}

// NameServicesDNSsDataSourceModel describes the data source data model.
type NameServicesDNSsDataSourceModel struct {
	CxProfileName    types.String                          `tfsdk:"cx_profile_name"`
	NameServicesDNSs []NameServicesDNSDataSourceModel      `tfsdk:"name_services_dnss"`
	Filter           *NameServicesDNSDataSourceFilterModel `tfsdk:"filter"`
}

// NameServicesDNSDataSourceFilterModel describes the data source data model for queries.
type NameServicesDNSDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
	Domains types.String `tfsdk:"dns_domains"`
	Servers types.String `tfsdk:"name_servers"`
}

// Metadata returns the data source type name.
func (d *NameServicesDNSsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *NameServicesDNSsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesDNSs data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Connection profile name",
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "IPInterface svm name.",
					},
					"dns_domains": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "DNS domain such as 'sales.bar.com'. The first domain is the one that the svm belongs to.",
					},
					"name_servers": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "IPv4 address of name servers such as '123.123.123.123'.",
					},
				},
				Optional: true,
			},
			"name_services_dnss": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Connection profile name",
						},
						"svm_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "IPInterface svm name",
						},
						"svm_uuid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "UUID of svm",
						},
						"dns_domains": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "List of DNS domains such as 'sales.bar.com'. The first domain is the one that the svm belongs to.",
						},
						"name_servers": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "List of IPv4 addresses of name servers such as '123.123.123.123'.",
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "List of IPv4 addresses of name servers such as '123.123.123.123'.",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NameServicesDNSsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *NameServicesDNSsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NameServicesDNSsDataSourceModel

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

	var filter *interfaces.NameServicesDNSDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.NameServicesDNSDataSourceFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
			Domains: data.Filter.Domains.ValueString(),
			Servers: data.Filter.Servers.ValueString(),
		}
	}
	restInfo, err := interfaces.GetListNameServicesDNSs(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetNameServicesDNSs
		return
	}

	data.NameServicesDNSs = make([]NameServicesDNSDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.NameServicesDNSs[index] = NameServicesDNSDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVMName:       types.StringValue(record.SVM.Name),
			SVMUUID:       types.StringValue(record.SVM.UUID),
			Domains:       flattenTypesStringList(record.Domains),
			NameServers:   flattenTypesStringList(record.Servers),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
