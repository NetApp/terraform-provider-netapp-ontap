package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageLunDataSource{}

// NewStorageLunDataSource is a helper function to simplify the provider implementation.
func NewStorageLunDataSource() datasource.DataSource {
	return &StorageLunDataSource{
		config: resourceOrDataSourceConfig{
			name: "storage_lun_data_source",
		},
	}
}

// StorageLunDataSource defines the data source implementation.
type StorageLunDataSource struct {
	config resourceOrDataSourceConfig
}

// StorageLunDataSourceModel describes the data source data model.
type StorageLunDataSourceModel struct {
	CxProfileName types.String                        `tfsdk:"cx_profile_name"`
	Name          types.String                        `tfsdk:"name"`
	SVMName       types.String                        `tfsdk:"svm_name"`
	CreationTime  types.String                        `tfsdk:"create_time"`
	Location      *StorageLunDataSourceLocationModel  `tfsdk:"location"`
	OSType        types.String                        `tfsdk:"os_type"`
	QoSPolicy     *StorageLunDataSourceQoSPolicyModel `tfsdk:"qos_policy"`
	Space         *StorageLunDataSourceSpaceModel     `tfsdk:"space"`
	ID            types.String                        `tfsdk:"id"`
}

// StorageLunDataSourceLocationModel describes the data source data model for queries.
type StorageLunDataSourceLocationModel struct {
	LogicalUnit types.String                     `tfsdk:"logical_unit"`
	Volume      *StorageLunDataSourceVolumeModel `tfsdk:"volume"`
}

// StorageLunDataSourceVolumeModel describes the data source data model for queries.
type StorageLunDataSourceVolumeModel struct {
	Name types.String `tfsdk:"name"`
	UUID types.String `tfsdk:"uuid"`
}

// StorageLunDataSourceQoSPolicyModel describes the data source data model for queries.
type StorageLunDataSourceQoSPolicyModel struct {
	Name types.String `tfsdk:"name"`
	UUID types.String `tfsdk:"uuid"`
}

// StorageLunDataSourceSpaceModel describes the data source data model for queries.
type StorageLunDataSourceSpaceModel struct {
	Size types.Int64 `tfsdk:"size"`
	Used types.Int64 `tfsdk:"used"`
}

// Metadata returns the data source type name.
func (d *StorageLunDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *StorageLunDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageLun data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name for lun",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "svm name for lun",
				Required:            true,
			},
			"privileges": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of privileges",
				Required:            true,
				PlanModifiers:       []planmodifier.String{},
			},
			"create_time": schema.StringAttribute{
				MarkdownDescription: "Time when the lun was created",
				Computed:            true,
			},

			"location": schema.SingleNestedAttribute{
				Computed: true,
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"logical_unit": schema.StringAttribute{
						MarkdownDescription: "Logical unit name",
						Computed:            true,
					},
					"volume": schema.SingleNestedAttribute{
						Computed: true,
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "Volume name",
								Required:            true,
							},
							"uuid": schema.StringAttribute{
								MarkdownDescription: "Volume uuid",
								Computed:            true,
							},
						},
					},
				},
			},
			"os_type": schema.StringAttribute{
				MarkdownDescription: "OS type for lun",
				Computed:            true,
			},
			"qos_policy": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "QoS policy name",
						Computed:            true,
					},
					"uuid": schema.StringAttribute{
						MarkdownDescription: "QoS policy uuid",
						Computed:            true,
					},
				},
			},
			"space": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"size": schema.Int64Attribute{
						MarkdownDescription: "Size of the lun",
						Computed:            true,
					},
					"used": schema.Int64Attribute{
						MarkdownDescription: "Used space of the lun",
						Computed:            true,
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Lun uuid",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *StorageLunDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *StorageLunDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	tflog.Debug(ctx, fmt.Sprintf("carchi7py read a data source"))
	var data StorageLunDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("carchi7py read a data source: %#v", data))

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	restInfo, err := interfaces.GetStorageLunByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString(), data.Location.Volume.Name.ValueString())
	if err != nil {
		// error reporting done inside GetStorageLun
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("carchi7py read a rest info source: %#v", restInfo))
	data.Name = types.StringValue(restInfo.Name)
	data.CreationTime = types.StringValue(restInfo.CreateTime)
	data.Location.LogicalUnit = types.StringValue(restInfo.Location.LogicalUnit)
	data.Location.Volume.Name = types.StringValue(restInfo.Location.Volume.Name)
	data.Location.Volume.UUID = types.StringValue(restInfo.Location.Volume.UUID)
	data.OSType = types.StringValue(restInfo.OSType)
	if restInfo.QoSPolicy.Name != "" {
		data.QoSPolicy.Name = types.StringValue(restInfo.QoSPolicy.Name)
	}
	if restInfo.QoSPolicy.UUID != "" {
		data.QoSPolicy.UUID = types.StringValue(restInfo.QoSPolicy.UUID)
	}
	data.Space = &StorageLunDataSourceSpaceModel{
		Size: types.Int64Value(restInfo.Space.Size),
		Used: types.Int64Value(restInfo.Space.Used),
	}
	data.ID = types.StringValue(restInfo.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
