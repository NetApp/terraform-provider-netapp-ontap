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
var _ datasource.DataSource = &SvmsDataSource{}

// NewSvmsDataSource is a helper function to simplify the provider implementation.
func NewSvmsDataSource() datasource.DataSource {
	return &SvmsDataSource{
		config: resourceOrDataSourceConfig{
			name: "svms_data_source",
		},
	}
}

// SvmsDataSource defines the data source implementation.
type SvmsDataSource struct {
	config resourceOrDataSourceConfig
}

// SvmsDataSourceModel describes the data source data model.
type SvmsDataSourceModel struct {
	CxProfileName types.String              `tfsdk:"cx_profile_name"`
	Svms          []SvmDataSourceModel      `tfsdk:"svms"`
	Filter        *SvmDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *SvmsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *SvmsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Svms data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"svms": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
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
							MarkdownDescription: "The subtype for svm to be created",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Comment for svm to be created",
							Computed:            true,
						},
						"language": schema.StringAttribute{
							MarkdownDescription: "Language to use for svm",
							Computed:            true,
						},
						"aggregates": schema.ListAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "Aggregates to be assigned use for svm",
							Computed:            true,
						},
						"max_volumes": schema.StringAttribute{
							MarkdownDescription: "Maximum number of volumes that can be created on the svm. Expects an integer or unlimited",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							Computed: true,
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
func (d *SvmsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SvmsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SvmsDataSourceModel

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

	var filter *interfaces.SvmDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.SvmDataSourceFilterModel{
			Name: data.Filter.Name.ValueString(),
		}
	}
	restInfo, err := interfaces.GetSvmsByName(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetSvms
		return
	}

	data.Svms = make([]SvmDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		var aggregates []types.String
		for _, v := range record.Aggregates {
			aggregates = append(aggregates, types.StringValue(v.Name))
		}

		data.Svms[index] = SvmDataSourceModel{
			CxProfileName:  data.CxProfileName,
			Name:           types.StringValue(record.Name),
			ID:             types.StringValue(record.UUID),
			Ipspace:        types.StringValue(record.Ipspace.Name),
			SnapshotPolicy: types.StringValue(record.SnapshotPolicy.Name),
			SubType:        types.StringValue(record.SubType),
			Comment:        types.StringValue(record.Comment),
			Language:       types.StringValue(record.Language),
			Aggregates:     aggregates,
			MaxVolumes:     types.StringValue(record.MaxVolumes),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
