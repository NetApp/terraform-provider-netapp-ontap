package cluster

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
var _ datasource.DataSource = &ClusterLicensingLicensesDataSource{}

// NewClusterLicensingLicensesDataSource is a helper function to simplify the provider implementation.
func NewClusterLicensingLicensesDataSource() datasource.DataSource {
	return &ClusterLicensingLicensesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cluster_licensing_licenses",
		},
	}
}

// NewClusterLicensingLicensesDataSourceAlias is a helper function to simplify the provider implementation.
func NewClusterLicensingLicensesDataSourceAlias() datasource.DataSource {
	return &ClusterLicensingLicensesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cluster_licensing_licenses_data_source",
		},
	}
}

// ClusterLicensingLicensesDataSource defines the data source implementation.
type ClusterLicensingLicensesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ClusterLicensingLicensesDataSourceModel describes the data source data model.
type ClusterLicensingLicensesDataSourceModel struct {
	CxProfileName            types.String                                   `tfsdk:"cx_profile_name"`
	ClusterLicensingLicenses []ClusterLicensingLicenseDataSourceModel       `tfsdk:"cluster_licensing_licenses"`
	Filter                   *ClusterLicensingLicensesDataSourceFilterModel `tfsdk:"filter"`
}

// ClusterLicensingLicensesDataSourceFilterModel describes the data source data model for queries.
type ClusterLicensingLicensesDataSourceFilterModel struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *ClusterLicensingLicensesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ClusterLicensingLicensesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ClusterLicensingLicenses data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "ClusterLicensingLicense name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"cluster_licensing_licenses": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "ClusterLicensingLicense name",
							Required:            true,
						},
						"scope": schema.StringAttribute{
							MarkdownDescription: "Scope of the license",
							Computed:            true,
						},
						"state": schema.StringAttribute{
							MarkdownDescription: "State of the license",
							Computed:            true,
						},
						"licenses": schema.ListNestedAttribute{
							MarkdownDescription: "Licenses of the license",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"serial_number": schema.StringAttribute{
										MarkdownDescription: "Serial Number of the license",
										Computed:            true,
									},
									"owner": schema.StringAttribute{
										MarkdownDescription: "owner of the license",
										Computed:            true,
									},
									"compliance": schema.SingleNestedAttribute{
										MarkdownDescription: "compliance of the license",
										Computed:            true,
										Attributes: map[string]schema.Attribute{
											"state": schema.StringAttribute{
												MarkdownDescription: "state of the license",
												Computed:            true,
											},
										},
									},
									"active": schema.BoolAttribute{
										MarkdownDescription: "active of the license",
										Computed:            true,
									},
									"evaluation": schema.BoolAttribute{
										MarkdownDescription: "evaluation of the license",
										Computed:            true,
									},
									"installed_license": schema.StringAttribute{
										MarkdownDescription: "installed_license of the license",
										Computed:            true,
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
func (d *ClusterLicensingLicensesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ClusterLicensingLicensesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterLicensingLicensesDataSourceModel

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

	var filter *interfaces.ClusterLicensingLicenseFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.ClusterLicensingLicenseFilterModel{
			Name: data.Filter.Name.ValueString(),
		}
	}
	restInfo, err := interfaces.GetListClusterLicensingLicenses(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetClusterLicensingLicenses
		return
	}

	data.ClusterLicensingLicenses = make([]ClusterLicensingLicenseDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		var licenses = make([]LicensesModel, len(record.Licenses))
		for i, v := range record.Licenses {
			license := LicensesModel{
				SerialNumber: types.StringValue(v.SerialNumber),
				Owner:        types.StringValue(v.Owner),
				Compliance: &Compliance{
					State: types.StringValue(v.Compliance.State),
				},
				Active:           types.BoolValue(v.Active),
				Evaluation:       types.BoolValue(v.Evaluation),
				InstalledLicense: types.StringValue(v.InstalledLicense),
			}

			licenses[i] = license
		}

		data.ClusterLicensingLicenses[index] = ClusterLicensingLicenseDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
			Licenses:      licenses,
			State:         types.StringValue(record.State),
			Scope:         types.StringValue(record.Scope),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
