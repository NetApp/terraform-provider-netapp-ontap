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
// copy this file to match you data source (should match internal/provider/protocols_san_igroup_data_source.go)
// replace ProtocolsSanIgroups with the name of the resource, following go conventions, eg IPInterfaces
// replace protocols_san_igroups with the name of the resource, for logging purposes, eg ip_interfaces
// make sure to create internal/interfaces/protocols_san_igroup.go too)
// delete these 5 lines

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsSanIgroupsDataSource{}

// NewProtocolsSanIgroupsDataSource is a helper function to simplify the provider implementation.
func NewProtocolsSanIgroupsDataSource() datasource.DataSource {
	return &ProtocolsSanIgroupsDataSource{
		config: resourceOrDataSourceConfig{
			name: "protocols_san_igroups_data_source",
		},
	}
}

// ProtocolsSanIgroupsDataSource defines the data source implementation.
type ProtocolsSanIgroupsDataSource struct {
	config resourceOrDataSourceConfig
}

// ProtocolsSanIgroupsDataSourceModel describes the data source data model.
type ProtocolsSanIgroupsDataSourceModel struct {
	CxProfileName types.String                      `tfsdk:"cx_profile_name"`
	ProtocolsSanIgroups   []ProtocolsSanIgroupDataSourceModel         `tfsdk:"protocols_san_igroups"`
	Filter        *ProtocolsSanIgroupsDataSourceFilterModel `tfsdk:"filter"`
}

// ProtocolsSanIgroupsDataSourceFilterModel describes the data source data model for queries.
type ProtocolsSanIgroupsDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsSanIgroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *ProtocolsSanIgroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsSanIgroups data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "ProtocolsSanIgroup name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "ProtocolsSanIgroup svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_san_igroups": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "ProtocolsSanIgroup name",
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
func (d *ProtocolsSanIgroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsSanIgroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsSanIgroupsDataSourceModel

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

	var filter *interfaces.ProtocolsSanIgroupDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.ProtocolsSanIgroupDataSourceFilterModel{
			Name: data.Filter.Name.ValueString(),
		}
	}
	restInfo, err := interfaces.GetProtocolsSanIgroups(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetProtocolsSanIgroups
		return
	}

	data.ProtocolsSanIgroups = make([]ProtocolsSanIgroupDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.ProtocolsSanIgroups[index] = ProtocolsSanIgroupDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}