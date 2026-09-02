package name_services

import (
	"context"
	"fmt"
	"strconv"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &NameServicesNameMappingsDataSource{}

// NewNameServicesNameMappingsDataSource is a helper function to simplify the provider implementation.
func NewNameServicesNameMappingsDataSource() datasource.DataSource {
	return &NameServicesNameMappingsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "name_services_name_mappings",
		},
	}
}

// NameServicesNameMappingsDataSource defines the data source implementation.
type NameServicesNameMappingsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesNameMappingsDataSourceModel describes the data source data model.
type NameServicesNameMappingsDataSourceModel struct {
	CxProfileName            types.String                                   `tfsdk:"cx_profile_name"`
	NameServicesNameMappings []NameServicesNameMappingDataSourceModel       `tfsdk:"name_services_name_mappings"`
	Filter                   *NameServicesNameMappingsDataSourceFilterModel `tfsdk:"filter"`
}

// NameServicesNameMappingsDataSourceFilterModel describes the data source filter model.
type NameServicesNameMappingsDataSourceFilterModel struct {
	SVMName   types.String `tfsdk:"svm_name"`
	Direction types.String `tfsdk:"direction"`
}

// Metadata returns the data source type name.
func (d *NameServicesNameMappingsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *NameServicesNameMappingsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "NameServicesNameMappings data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "Name of the SVM that owns the name mappings",
						Optional:            true,
					},
					"direction": schema.StringAttribute{
						MarkdownDescription: "Direction of the name mappings. Possible values: krb_unix, win_unix, unix_win, s3_unix, s3_win.",
						Optional:            true,
					},
				},
				Required: true,
			},
			"name_services_name_mappings": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "Name of the SVM that owns the name mapping",
							Computed:            true,
						},
						"direction": schema.StringAttribute{
							MarkdownDescription: "Direction of the name mapping",
							Computed:            true,
						},
						"index": schema.Int64Attribute{
							MarkdownDescription: "Position in the list of name mappings",
							Computed:            true,
						},
						"pattern": schema.StringAttribute{
							MarkdownDescription: "UNIX-style regular expression used to match the name to be mapped",
							Computed:            true,
						},
						"client_match": schema.StringAttribute{
							MarkdownDescription: "Client workstation IP address or hostname matched when searching for the pattern",
							Computed:            true,
						},
						"replacement": schema.StringAttribute{
							MarkdownDescription: "Replacement name used when the pattern matches",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "Composite identifier: {svm_uuid}/{direction}/{index}",
							Computed:            true,
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "List of name mappings",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NameServicesNameMappingsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *NameServicesNameMappingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NameServicesNameMappingsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		return
	}

	if data.Filter == nil || data.Filter.SVMName.IsNull() {
		errorHandler.MakeAndReportError("No SVM specified", "svm_name must be specified in filter")
		return
	}

	filter := &interfaces.NameServicesNameMappingDataSourceFilterModel{
		SVMName: data.Filter.SVMName.ValueString(),
	}
	if data.Filter != nil && !data.Filter.Direction.IsNull() {
		filter.Direction = data.Filter.Direction.ValueString()
	}

	restInfo, err := interfaces.GetNameServicesNameMappings(errorHandler, *client, filter)
	if err != nil {
		return
	}

	data.NameServicesNameMappings = make([]NameServicesNameMappingDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.NameServicesNameMappings[index] = NameServicesNameMappingDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVMName:       types.StringValue(record.SVM.Name),
			Direction:     types.StringValue(record.Direction),
			Index:         types.Int64Value(int64(record.Index)),
			Pattern:       types.StringValue(record.Pattern),
			ClientMatch:   types.StringValue(record.ClientMatch),
			Replacement:   types.StringValue(record.Replacement),
			ID:            types.StringValue(record.SVM.UUID + "/" + record.Direction + "/" + strconv.Itoa(record.Index)),
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
