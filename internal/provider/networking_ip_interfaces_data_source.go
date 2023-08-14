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
var _ datasource.DataSource = &IPInterfacesDataSource{}

// NewIPInterfacesDataSource is a helper function to simplify the provider implementation.
func NewIPInterfacesDataSource() datasource.DataSource {
	return &IPInterfacesDataSource{
		config: resourceOrDataSourceConfig{
			name: "networking_ip_interfaces_data_source",
		},
	}
}

// IPInterfacesDataSource defines the data source implementation.
type IPInterfacesDataSource struct {
	config resourceOrDataSourceConfig
}

// IPInterfacesDataSourceModel describes the data source data model.
type IPInterfacesDataSourceModel struct {
	CxProfileName types.String                      `tfsdk:"cx_profile_name"`
	IPInterfaces  []IPInterfaceDataSourceModel      `tfsdk:"ip_interfaces"`
	Filter        *IPInterfaceDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *IPInterfacesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *IPInterfacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "IPInterfaces data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "IPInterface name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "IPInterface svm name",
						Optional:            true,
					},
					"scope": schema.StringAttribute{
						MarkdownDescription: "IPInterface scope",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"ip_interfaces": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "IPInterface name",
							Required:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "IPInterface svm name",
							Optional:            true,
						},
						"scope": schema.StringAttribute{
							MarkdownDescription: "IPInterface scope",
							Computed:            true,
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
func (d *IPInterfacesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *IPInterfacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IPInterfacesDataSourceModel

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

	var filter *interfaces.IPInterfaceGetDataModelONTAP = nil
	if data.Filter != nil {
		filter = &interfaces.IPInterfaceGetDataModelONTAP{
			Name:    data.Filter.Name.ValueString(),
			Scope:   data.Filter.Scope.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetIPInterfaces(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetIPInterfaces
		return
	}

	data.IPInterfaces = make([]IPInterfaceDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.IPInterfaces[index] = IPInterfaceDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
			Scope:         types.StringValue(record.Scope),
			SVMName:       types.StringValue(record.SVMName),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
