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

// SecurityAccountDataSourceFilterModel describes the data source data model for queries.
type SecurityAccountDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm_name"`
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
						MarkdownDescription: "SecurityAccount svm name (Owner name)",
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
						"owner": schema.SingleNestedAttribute{
							MarkdownDescription: "SecurityAccount owner",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "SecurityAccount owner name",
									Computed:            true,
								},
								"uuid": schema.StringAttribute{
									MarkdownDescription: "SecurityAccount owner uuid",
									Computed:            true,
								},
							},
						},
						"locked": schema.BoolAttribute{
							MarkdownDescription: "SecurityAccount locked",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "SecurityAccount comment",
							Computed:            true,
						},
						"role": schema.SingleNestedAttribute{
							MarkdownDescription: "SecurityAccount role",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "SecurityAccount role name",
									Computed:            true,
								},
							},
						},
						"scope": schema.StringAttribute{
							MarkdownDescription: "SecurityAccount scope",
							Computed:            true,
						},
						"applications": schema.ListNestedAttribute{
							MarkdownDescription: "SecurityAccount applications",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"application": schema.StringAttribute{
										MarkdownDescription: "SecurityAccount application",
										Computed:            true,
									},
									"second_authentication_method": schema.StringAttribute{
										MarkdownDescription: "SecurityAccount second authentication method",
										Computed:            true,
									},
									"authentication_methods": schema.ListAttribute{
										MarkdownDescription: "SecurityAccount authentication methods",
										Computed:            true,
										ElementType:         types.StringType,
									},
								},
							},
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

	var filter *interfaces.SecurityAccountDataSourceFilterModel = nil
	if data.Filter != nil {
		if data.Filter.SVMName.IsNull() {
			filter = &interfaces.SecurityAccountDataSourceFilterModel{
				Name: data.Filter.Name.ValueString(),
			}
		} else {
			filter = &interfaces.SecurityAccountDataSourceFilterModel{
				Name: data.Filter.Name.ValueString(),
				Owner: interfaces.SecurityAccountOwner{
					Name: data.Filter.SVMName.ValueString(),
				},
			}
		}

	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("security account filter: %+v", filter))
	restInfo, err := interfaces.GetSecurityAccounts(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetSecurityAccounts
		return
	}
	data.SecurityAccounts = make([]SecurityAccountDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.SecurityAccounts[index] = SecurityAccountDataSourceModel{
			CxProfileName: data.CxProfileName,
			Name:          types.StringValue(record.Name),
			Owner: &OwnerDataSourceModel{
				Name:    types.StringValue(record.Owner.Name),
				OwnerID: types.StringValue(record.Owner.UUID),
			},
			Locked:  types.BoolValue(record.Locked),
			Comment: types.StringValue(record.Comment),
			Role: &RoleDataSourceModel{
				Name: types.StringValue(record.Role.Name),
			},
			Scope:        types.StringValue(record.Scope),
			Applications: make([]ApplicationsDataSourceModel, len(record.Applications)),
		}
		for i, application := range record.Applications {
			data.SecurityAccounts[index].Applications[i] = ApplicationsDataSourceModel{
				Application:                types.StringValue(application.Application),
				SecondAuthentiactionMethod: types.StringValue(application.SecondAuthenticationMethod),
			}
			var authenticationMethods []types.String
			for _, authenticationMethod := range application.AuthenticationMethods {
				authenticationMethods = append(authenticationMethods, types.StringValue(authenticationMethod))
			}
			data.SecurityAccounts[index].Applications[i].AuthenticationMethods = &authenticationMethods
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
