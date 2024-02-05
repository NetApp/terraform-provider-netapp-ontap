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
var _ datasource.DataSource = &StorageLunsDataSource{}

// NewStorageLunsDataSource is a helper function to simplify the provider implementation.
func NewStorageLunsDataSource() datasource.DataSource {
	return &StorageLunsDataSource{
		config: resourceOrDataSourceConfig{
			name: "storage_luns_data_source",
		},
	}
}

// StorageLunsDataSource defines the data source implementation.
type StorageLunsDataSource struct {
	config resourceOrDataSourceConfig
}

// StorageLunsDataSourceModel describes the data source data model.
type StorageLunsDataSourceModel struct {
	CxProfileName types.String                      `tfsdk:"cx_profile_name"`
	StorageLuns   []StorageLunDataSourceModel       `tfsdk:"storage_luns"`
	Filter        *StorageLunsDataSourceFilterModel `tfsdk:"filter"`
}

// StorageLunsDataSourceFilterModel describes the data source data model for queries.
type StorageLunsDataSourceFilterModel struct {
	Name       types.String `tfsdk:"name"`
	SVMName    types.String `tfsdk:"svm_name"`
	VolumeName types.String `tfsdk:"volume_name"`
}

// Metadata returns the data source type name.
func (d *StorageLunsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *StorageLunsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageLuns data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "StorageLun name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "StorageLun svm name",
						Optional:            true,
					},
					"volume_name": schema.StringAttribute{
						MarkdownDescription: "StorageLun volume name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"storage_luns": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "StorageLun name",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "StorageLun svm name",
							Computed:            true,
						},
						"create_time": schema.StringAttribute{
							MarkdownDescription: "Time when the lun was created",
							Computed:            true,
						},
						"location": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"logical_unit": schema.StringAttribute{
									MarkdownDescription: "Logical unit name",
									Computed:            true,
								},
								"volume": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											MarkdownDescription: "Volume name",
											Computed:            true,
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
									MarkdownDescription: "Size of lun in bytes",
									Computed:            true,
								},
								"used": schema.Int64Attribute{
									MarkdownDescription: "Used space of lun in bytes",
									Computed:            true,
								},
							},
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "StorageLun UUID",
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
func (d *StorageLunsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *StorageLunsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageLunsDataSourceModel

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

	var filter *interfaces.StorageLunDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.StorageLunDataSourceFilterModel{
			Name:       data.Filter.Name.ValueString(),
			SVMName:    data.Filter.SVMName.ValueString(),
			VolumeName: data.Filter.VolumeName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetStorageLuns(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetStorageLuns
		return
	}

	data.StorageLuns = make([]StorageLunDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.StorageLuns[index] = StorageLunDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
			SVMName:       types.StringValue(record.SVMName),
			CreationTime:  types.StringValue(record.CreateTime),
			Location: &StorageLunDataSourceLocationModel{
				LogicalUnit: types.StringValue(record.Location.LogicalUnit),
				Volume: &StorageLunDataSourceVolumeModel{
					Name: types.StringValue(record.Location.Volume.Name),
					UUID: types.StringValue(record.Location.Volume.UUID),
				},
			},
			OSType: types.StringValue(record.OSType),
			QoSPolicy: &StorageLunDataSourceQoSPolicyModel{
				Name: types.StringValue(record.QoSPolicy.Name),
				UUID: types.StringValue(record.QoSPolicy.UUID),
			},
			Space: &StorageLunDataSourceSpaceModel{
				Size: types.Int64Value(record.Space.Size),
				Used: types.Int64Value(record.Space.Used),
			},
			ID: types.StringValue(record.UUID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
