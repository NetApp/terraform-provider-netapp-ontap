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
var _ datasource.DataSource = &SvmDataSource{}

// NewSvmDataSource is a helper function to simplify the provider implementation.
func NewSvmDataSource() datasource.DataSource {
	return &SvmDataSource{
		config: resourceOrDataSourceConfig{
			name: "svm_data_source",
		},
	}
}

// SvmDataSource defines the data source implementation.
type SvmDataSource struct {
	config resourceOrDataSourceConfig
}

// SvmDataSourceModel describes the data source data model.
type SvmDataSourceModel struct {
	CxProfileName  types.String   `tfsdk:"cx_profile_name"`
	Name           types.String   `tfsdk:"name"`
	Ipspace        types.String   `tfsdk:"ipspace"`
	SnapshotPolicy types.String   `tfsdk:"snapshot_policy"`
	SubType        types.String   `tfsdk:"subtype"`
	Comment        types.String   `tfsdk:"comment"`
	Language       types.String   `tfsdk:"language"`
	Aggregates     []types.String `tfsdk:"aggregates"`
	MaxVolumes     types.String   `tfsdk:"max_volumes"`
	ID             types.String   `tfsdk:"id"`
}

// SvmDataSourceFilterModel describes the data source data model for queries.
type SvmDataSourceFilterModel struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *SvmDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *SvmDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Svm data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Svm name",
				Required:            true,
			},
			"ipspace": schema.StringAttribute{
				MarkdownDescription: "The name of the ipspace to manage",
				Computed:            true,
			},
			"snapshot_policy": schema.StringAttribute{
				MarkdownDescription: "The name of the snapshot policy to manage",
				Computed:            true,
			},
			"subtype": schema.StringAttribute{
				MarkdownDescription: "The subtype for vserver to be created",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment for vserver to be created",
				Computed:            true,
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "Language to use for vserver",
				Computed:            true,
			},
			"aggregates": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Aggregates to be assigned use for vserver",
				Computed:            true,
			},
			"max_volumes": schema.StringAttribute{
				MarkdownDescription: "Maximum number of volumes that can be created on the vserver. Expects an integer or unlimited",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *SvmDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SvmDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SvmDataSourceModel

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

	restInfo, err := interfaces.GetSvmByNameDataSource(errorHandler, *client, data.Name.ValueString())
	if err != nil {
		// error reporting done inside GetSvm
		return
	}

	var aggregates []types.String
	for _, v := range restInfo.Aggregates {
		aggregates = append(aggregates, types.StringValue(v.Name))
	}

	data.Name = types.StringValue(restInfo.Name)
	data.ID = types.StringValue(restInfo.UUID)
	data.Ipspace = types.StringValue(restInfo.Ipspace.Name)
	data.SnapshotPolicy = types.StringValue(restInfo.SnapshotPolicy.Name)
	data.SubType = types.StringValue(restInfo.SubType)
	data.Comment = types.StringValue(restInfo.Comment)
	data.Language = types.StringValue(restInfo.Language)
	data.Aggregates = aggregates
	data.MaxVolumes = types.StringValue(restInfo.MaxVolumes)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
