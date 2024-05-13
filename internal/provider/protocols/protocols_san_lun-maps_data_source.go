package protocols

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/svm"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsSanLunMapsDataSource{}

// NewProtocolsSanLunMapsDataSource is a helper function to simplify the provider implementation.
func NewProtocolsSanLunMapsDataSource() datasource.DataSource {
	return &ProtocolsSanLunMapsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_san_lun-maps_data_source",
		},
	}
}

// ProtocolsSanLunMapsDataSource defines the data source implementation.
type ProtocolsSanLunMapsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsSanLunMapsDataSourceModel describes the data source data model.
type ProtocolsSanLunMapsDataSourceModel struct {
	CxProfileName       types.String                              `tfsdk:"cx_profile_name"`
	ProtocolsSanLunMaps []ProtocolsSanLunMapDataSourceModel       `tfsdk:"protocols_san_lun_maps"`
	Filter              *ProtocolsSanLunMapsDataSourceFilterModel `tfsdk:"filter"`
}

// ProtocolsSanLunMapsDataSourceFilterModel describes the data source data model for queries.
type ProtocolsSanLunMapsDataSourceFilterModel struct {
	SVM    svm.SVM `tfsdk:"svm"`
	Lun    Lun     `tfsdk:"lun"`
	IGroup IGroup  `tfsdk:"igroup"`
}

// Metadata returns the data source type name.
func (d *ProtocolsSanLunMapsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsSanLunMapsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsSanLunMaps data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm": schema.SingleNestedAttribute{
						MarkdownDescription: "SVM details for ProtocolsSanLunMaps",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "name of the SVM",
								Optional:            true,
							},
						},
					},
					"igroup": schema.SingleNestedAttribute{
						MarkdownDescription: "SVM details for ProtocolsSanLunMaps",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "name of the igroup",
								Optional:            true,
							},
						},
					},
					"lun": schema.SingleNestedAttribute{
						MarkdownDescription: "SVM details for ProtocolsSanLunMaps",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "name of the lun",
								Optional:            true,
							},
						},
					},
				},
				Optional: true,
			},
			"protocols_san_lun_maps": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"svm": schema.SingleNestedAttribute{
							MarkdownDescription: "SVM details for ProtocolsSanLunMaps",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "name of the SVM",
									Required:            true,
								},
							},
						},
						"igroup": schema.SingleNestedAttribute{
							MarkdownDescription: "SVM details for ProtocolsSanLunMaps",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "name of the igroup",
									Required:            true,
								},
							},
						},
						"lun": schema.SingleNestedAttribute{
							MarkdownDescription: "SVM details for ProtocolsSanLunMaps",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "name of the lun",
									Required:            true,
								},
							},
						},
						"logical_unit_number": schema.Int64Attribute{
							MarkdownDescription: "If no value is provided, ONTAP assigns the lowest available value",
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
func (d *ProtocolsSanLunMapsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsSanLunMapsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsSanLunMapsDataSourceModel

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

	var filter *interfaces.ProtocolsSanLunMapsDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.ProtocolsSanLunMapsDataSourceFilterModel{
			SVM: interfaces.SVM{
				Name: data.Filter.SVM.Name.ValueString(),
			},
			Lun: interfaces.Lun{
				Name: data.Filter.Lun.Name.ValueString(),
			},
			IGroup: interfaces.IGroup{
				Name: data.Filter.IGroup.Name.ValueString(),
			},
		}
	}
	restInfo, err := interfaces.GetProtocolsSanLunMaps(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetProtocolsSanLunMaps
		return
	}

	data.ProtocolsSanLunMaps = make([]ProtocolsSanLunMapDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.ProtocolsSanLunMaps[index] = ProtocolsSanLunMapDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVM: svm.SVM{
				Name: types.StringValue(record.SVM.Name),
			},
			IGroup: IGroup{
				Name: types.StringValue(record.IGroup.Name),
			},
			Lun: Lun{
				Name: types.StringValue(record.Lun.Name),
			},
			LogicalUnitNumber: types.Int64Value(int64(record.LogicalUnitNumber)),
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
