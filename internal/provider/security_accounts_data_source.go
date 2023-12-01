package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/provider/security_account_data_source.go)
// replace SecurityAccounts with the name of the resource, following go conventions, eg IPInterfaces
// replace security_accounts with the name of the resource, for logging purposes, eg ip_interfaces
// make sure to create internal/interfaces/security_account.go too)
// delete these 5 lines

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SecurityAccountsDataSource{}

// NewSecurityAccountsDataSource is a helper function to simplify the provider implementation.
func NewSecurityAccountsDataSource() datasource.DataSource {
	return &SecurityAccountsDataSource{
		config: resourceOrDataSourceConfig{
			name: "security_accounts_data_source",
		},
	}
}

// SecurityAccountsDataSource defines the data source implementation.
type SecurityAccountsDataSource struct {
	config resourceOrDataSourceConfig
}

// SecurityAccountsDataSourceModel describes the data source data model.
type SecurityAccountsDataSourceModel struct {
	CxProfileName    types.String                          `tfsdk:"cx_profile_name"`
	SecurityAccounts []SecurityAccountDataSourceModel      `tfsdk:"security_accounts"`
	Filter           *SecurityAccountDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *SecurityAccountsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *SecurityAccountsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityAccounts data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "SecurityAccount name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "SecurityAccount svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"security_accounts": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "SecurityAccount name",
							Required:            true,
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *SecurityAccountsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SecurityAccountsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityAccountsDataSourceModel

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
	if client == nil {
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}