package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ExportPoliciesDataSource{}

// NewExportPoliciesDataSource is a helper function to simplify the provider implementation.
func NewExportPoliciesDataSource() datasource.DataSource {
	return &ExportPoliciesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_nfs_export_policies_data_source",
		},
	}
}

// ExportPoliciesDataSource defines the data source implementation.
type ExportPoliciesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ExportPolicyGetDataSourceModelONTAP describes the source data model.
type ExportPolicyGetDataSourceModelONTAP struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	ID            types.Int64  `tfsdk:"id"`
	SVMName       types.String `tfsdk:"svm_name"`
	SVMUUID       types.String `tfsdk:"svm_uuid"`
}

// ExportPoliciesDataSourceModel describes the data source data model.
type ExportPoliciesDataSourceModel struct {
	CxProfileName  types.String                          `tfsdk:"cx_profile_name"`
	ExportPolicies []ExportPolicyGetDataSourceModelONTAP `tfsdk:"protocols_nfs_export_policies"`
	Filter         *ExportPolicyDataSourceFilterModel    `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *ExportPoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ExportPoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ExportPolicies data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "ExportPolicy name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "ExportPolicy svm name name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_nfs_export_policies": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "ExportPolicy name",
							Required:            true,
						},
						"id": schema.Int64Attribute{
							MarkdownDescription: "Export policy identifier",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "Name of the svm name",
							Computed:            true,
						},
						"svm_uuid": schema.StringAttribute{
							MarkdownDescription: "UUID of the svm uuid",
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
func (d *ExportPoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ExportPoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExportPoliciesDataSourceModel

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

	var filter *interfaces.ExportPolicyGetDataFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.ExportPolicyGetDataFilterModel{
			Name:    data.Filter.Name.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetExportPoliciesList(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetExportPolicys
		return
	}

	data.ExportPolicies = make([]ExportPolicyGetDataSourceModelONTAP, len(restInfo))
	for index, record := range restInfo {
		data.ExportPolicies[index] = ExportPolicyGetDataSourceModelONTAP{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
			ID:            types.Int64Value(int64(record.ID)),
			SVMName:       types.StringValue(record.Svm.Name),
			SVMUUID:       types.StringValue(record.Svm.UUID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
