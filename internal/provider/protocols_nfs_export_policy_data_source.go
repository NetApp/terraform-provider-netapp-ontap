package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ExportPolicyDataSource{}

// var _ resource.ResourceWithImportState = &ExportPolicyResource{}

// NewExportPolicyDataSource is a helper function to simplify the provider implementation.
func NewExportPolicyDataSource() datasource.DataSource {
	return &ExportPolicyDataSource{
		config: resourceOrDataSourceConfig{
			name: "protocols_nfs_export_policy_data_source",
		},
	}
}

// ExportPolicyDataSource defines the source implementation.
type ExportPolicyDataSource struct {
	config resourceOrDataSourceConfig
}

// ExportPolicyDataSourceModel describes the source data model.
type ExportPolicyDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Vserver       types.String `tfsdk:"vserver"`
	Name          types.String `tfsdk:"name"`
	ID            types.String `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (d *ExportPolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the resource.
func (d *ExportPolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Export policy rule resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"vserver": schema.StringAttribute{
				MarkdownDescription: "Name of the vserver to use",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Export policy name",
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Export policy identifier",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ExportPolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ExportPolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *ExportPolicyDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	filter := map[string]string{"name": data.Name.ValueString()}
	exportPolicy, err := interfaces.GetExportPolicies(errorHandler, *client, &filter)
	if err != nil {
		return
	}
	if exportPolicy == nil {
		errorHandler.MakeAndReportError("No export policy found", fmt.Sprintf("Export Policy %s not found.", data.Name))
		return
	}
	data.ID = types.StringValue(strconv.Itoa(exportPolicy.ID))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
