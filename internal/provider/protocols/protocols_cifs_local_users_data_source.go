package protocols

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
var _ datasource.DataSource = &CifsLocalUsersDataSource{}

// NewCifsLocalUsersDataSource is a helper function to simplify the provider implementation.
func NewCifsLocalUsersDataSource() datasource.DataSource {
	return &CifsLocalUsersDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cifs_local_users",
		},
	}
}

// NewCifsLocalUsersDataSourceAlias is a helper function to simplify the provider implementation.
func NewCifsLocalUsersDataSourceAlias() datasource.DataSource {
	return &CifsLocalUsersDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_cifs_local_users_data_source",
		},
	}
}

// CifsLocalUsersDataSource defines the data source implementation.
type CifsLocalUsersDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// CifsLocalUserDataSourceFilterModel describes the data source data model.
type CifsLocalUserDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm_name"`
}

// CifsLocalUsersDataSourceModel describes the data source data model.
type CifsLocalUsersDataSourceModel struct {
	CxProfileName  types.String                        `tfsdk:"cx_profile_name"`
	CifsLocalUsers []CifsLocalUserDataSourceModel      `tfsdk:"protocols_cifs_local_users"`
	Filter         *CifsLocalUserDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *CifsLocalUsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *CifsLocalUsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsLocalUsers data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "CifsLocalUser name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "CifsLocalUser svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_cifs_local_users": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "CifsLocalUser name",
							Required:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "CifsLocalUser svm name",
							Computed:            true,
						},
						"full_name": schema.StringAttribute{
							MarkdownDescription: "CifsLocalUser full name",
							Computed:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "CifsLocalUser description",
							Computed:            true,
						},
						"membership": schema.ListNestedAttribute{
							Computed:            true,
							MarkdownDescription: "CifsLocalUser membership",
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Computed:            true,
										MarkdownDescription: "CifsLocalUser membership name",
									},
								},
							},
						},
						"account_disabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "CifsLocalUser account disabled",
						},
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "CifsLocalUser id",
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "Protocols CIFS local users",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CifsLocalUsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *CifsLocalUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CifsLocalUsersDataSourceModel

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

	var filter *interfaces.CifsLocalUserDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.CifsLocalUserDataSourceFilterModel{
			Name:    data.Filter.Name.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetCifsLocalUsers(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetCifsLocalUsers
		return
	}

	data.CifsLocalUsers = make([]CifsLocalUserDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		var membership = make([]Membership, len(record.Membership))
		for i, v := range record.Membership {
			membership[i].Name = types.StringValue(v.Name)
		}
		data.CifsLocalUsers[index] = CifsLocalUserDataSourceModel{
			CxProfileName:   types.String(data.CxProfileName),
			Name:            types.StringValue(record.Name),
			SVMName:         types.StringValue(record.SVM.Name),
			FullName:        types.StringValue(record.FullName),
			Description:     types.StringValue(record.Description),
			AccountDisabled: types.BoolValue(record.AccountDisabled),
			ID:              types.StringValue(record.SID),
			Membership:      membership,
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
