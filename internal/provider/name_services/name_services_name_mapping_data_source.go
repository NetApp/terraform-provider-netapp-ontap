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
var _ datasource.DataSource = &NameServicesNameMappingDataSource{}

// NewNameServicesNameMappingDataSource is a helper function to simplify the provider implementation.
func NewNameServicesNameMappingDataSource() datasource.DataSource {
	return &NameServicesNameMappingDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "name_services_name_mapping",
		},
	}
}

// NameServicesNameMappingDataSource defines the data source implementation.
type NameServicesNameMappingDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesNameMappingDataSourceModel describes the data source data model.
type NameServicesNameMappingDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Direction     types.String `tfsdk:"direction"`
	Index         types.Int64  `tfsdk:"index"`
	Pattern       types.String `tfsdk:"pattern"`
	ClientMatch   types.String `tfsdk:"client_match"`
	Replacement   types.String `tfsdk:"replacement"`
	ID            types.String `tfsdk:"id"`
}

// Metadata returns the data source type name.
func (d *NameServicesNameMappingDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *NameServicesNameMappingDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "NameServicesNameMapping data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the SVM that owns the name mapping",
				Required:            true,
			},
			"direction": schema.StringAttribute{
				MarkdownDescription: "Direction of the name mapping. Possible values: krb_unix, win_unix, unix_win, s3_unix, s3_win.",
				Required:            true,
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "Position in the list of name mappings (1-2147483647).",
				Required:            true,
			},
			"pattern": schema.StringAttribute{
				MarkdownDescription: "UNIX-style regular expression used to match the name to be mapped (1-256 characters).",
				Computed:            true,
			},
			"client_match": schema.StringAttribute{
				MarkdownDescription: "Client workstation IP address or hostname matched when searching for the pattern.",
				Computed:            true,
			},
			"replacement": schema.StringAttribute{
				MarkdownDescription: "Replacement name used when the pattern matches (1-256 characters).",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Composite identifier: {svm_uuid}/{direction}/{index}",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NameServicesNameMappingDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *NameServicesNameMappingDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NameServicesNameMappingDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetNameServicesNameMappingByIndex(
		errorHandler,
		*client,
		data.SVMName.ValueString(),
		data.Direction.ValueString(),
		int(data.Index.ValueInt64()),
	)
	if err != nil {
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError(
			"No name mapping found",
			fmt.Sprintf("name mapping not found for svm_name=%s direction=%s index=%d", data.SVMName.ValueString(), data.Direction.ValueString(), data.Index.ValueInt64()),
		)
		return
	}

	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Direction = types.StringValue(restInfo.Direction)
	data.Index = types.Int64Value(int64(restInfo.Index))
	data.Pattern = types.StringValue(restInfo.Pattern)
	data.ClientMatch = types.StringValue(restInfo.ClientMatch)
	data.Replacement = types.StringValue(restInfo.Replacement)
	data.ID = types.StringValue(restInfo.SVM.UUID + "/" + restInfo.Direction + "/" + strconv.Itoa(restInfo.Index))

	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
