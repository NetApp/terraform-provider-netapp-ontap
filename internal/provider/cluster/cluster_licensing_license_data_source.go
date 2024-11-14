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
var _ datasource.DataSource = &ClusterLicensingLicenseDataSource{}

// NewClusterLicensingLicenseDataSource is a helper function to simplify the provider implementation.
func NewClusterLicensingLicenseDataSource() datasource.DataSource {
	return &ClusterLicensingLicenseDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cluster_licensing_license",
		},
	}
}

// NewClusterLicensingLicenseDataSourceAlias is a helper function to simplify the provider implementation.
func NewClusterLicensingLicenseDataSourceAlias() datasource.DataSource {
	return &ClusterLicensingLicenseDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cluster_licensing_license_data_source",
		},
	}
}

// ClusterLicensingLicenseDataSource defines the data source implementation.
type ClusterLicensingLicenseDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ClusterLicensingLicenseDataSourceModel describes the data source data model.
type ClusterLicensingLicenseDataSourceModel struct {
	CxProfileName types.String    `tfsdk:"cx_profile_name"`
	Name          types.String    `tfsdk:"name"`
	Licenses      []LicensesModel `tfsdk:"licenses"`
	State         types.String    `tfsdk:"state"`
	Scope         types.String    `tfsdk:"scope"`
}

// LicensesModel describes data source model.
type LicensesModel struct {
	SerialNumber     types.String `tfsdk:"serial_number"`
	Owner            types.String `tfsdk:"owner"`
	Compliance       *Compliance  `tfsdk:"compliance"`
	Active           types.Bool   `tfsdk:"active"`
	Evaluation       types.Bool   `tfsdk:"evaluation"`
	InstalledLicense types.String `tfsdk:"installed_license"`
}

// Entitlement describes data source model.
type Entitlement struct {
	Action types.String `tfsdk:"action"`
	Risk   types.String `tfsdk:"risk"`
}

// Compliance describes data source model.
type Compliance struct {
	State types.String `tfsdk:"state"`
}

// Capacity describes data source model.
type Capacity struct {
	MaximumSize types.Int64 `tfsdk:"maximum_size"`
	UsedSize    types.Int64 `tfsdk:"used_size"`
}

// ClusterLicensingLicenseDataSourceFilterModel describes the data source data model for queries.
type ClusterLicensingLicenseDataSourceFilterModel struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *ClusterLicensingLicenseDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ClusterLicensingLicenseDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ClusterLicensingLicense data source",

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
							MarkdownDescription: "installed license of the license",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ClusterLicensingLicenseDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ClusterLicensingLicenseDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ClusterLicensingLicenseDataSourceModel

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

	restInfo, err := interfaces.GetClusterLicensingLicenseByName(errorHandler, *client, data.Name.ValueString())
	if err != nil {
		// error reporting done inside GetClusterLicensingLicense
		return
	}

	var licenses = make([]LicensesModel, len(restInfo.Licenses))
	for i, v := range restInfo.Licenses {
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

	data = ClusterLicensingLicenseDataSourceModel{
		CxProfileName: data.CxProfileName,
		Name:          types.StringValue(restInfo.Name),
		Licenses:      licenses,
		State:         types.StringValue(restInfo.State),
		Scope:         types.StringValue(restInfo.Scope),
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
