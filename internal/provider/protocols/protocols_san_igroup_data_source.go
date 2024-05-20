package protocols

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsSanIgroupDataSource{}

// NewProtocolsSanIgroupDataSource is a helper function to simplify the provider implementation.
func NewProtocolsSanIgroupDataSource() datasource.DataSource {
	return &ProtocolsSanIgroupDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "san_igroup",
		},
	}
}

// ProtocolsSanIgroupDataSource defines the data source implementation.
type ProtocolsSanIgroupDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsSanIgroupDataSourceModel describes the data source data model.
type ProtocolsSanIgroupDataSourceModel struct {
	CxProfileName types.String                                 `tfsdk:"cx_profile_name"`
	Name          types.String                                 `tfsdk:"name"`
	SVMName       types.String                                 `tfsdk:"svm_name"`
	Comment       types.String                                 `tfsdk:"comment"`
	Igroups       []ProtocolsSanIgroupDataSourceIgroupModel    `tfsdk:"igroups"`
	Initiators    []ProtocolsSanIgroupDataSourceInitiatorModel `tfsdk:"initiators"`
	LunMaps       []ProtocolsSanIgroupDataSourceLunMapModel    `tfsdk:"lun_maps"`
	OsType        types.String                                 `tfsdk:"os_type"`
	Portset       *ProtocolsSanIgroupDataSourcePortsetModel    `tfsdk:"portset"`
	Protocol      types.String                                 `tfsdk:"protocol"`
	ID            types.String                                 `tfsdk:"id"`
}

// ProtocolsSanIgroupDataSourceIgroupModel describes the data source data model.
type ProtocolsSanIgroupDataSourceIgroupModel struct {
	Comment types.String `tfsdk:"comment"`
	Name    types.String `tfsdk:"name"`
	UUID    types.String `tfsdk:"uuid"`
}

// ProtocolsSanIgroupDataSourceInitiatorModel describes the data source data model.
type ProtocolsSanIgroupDataSourceInitiatorModel struct {
	Comment types.String `tfsdk:"comment"`
	Name    types.String `tfsdk:"name"`
}

// ProtocolsSanIgroupDataSourceLunMapModel describes the data source data model.
type ProtocolsSanIgroupDataSourceLunMapModel struct {
	LogicalUnitNumber types.Int64                          `tfsdk:"logical_unit_number"`
	Lun               ProtocolsSanIgroupDataSourceLunModel `tfsdk:"lun"`
}

// ProtocolsSanIgroupDataSourceLunModel describes the data source data model.
type ProtocolsSanIgroupDataSourceLunModel struct {
	Name types.String `tfsdk:"name"`
	UUID types.String `tfsdk:"uuid"`
}

// ProtocolsSanIgroupDataSourcePortsetModel describes the data source data model.
type ProtocolsSanIgroupDataSourcePortsetModel struct {
	Name types.String `tfsdk:"name"`
	UUID types.String `tfsdk:"uuid"`
}

// Metadata returns the data source type name.
func (d *ProtocolsSanIgroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsSanIgroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsSanIgroup data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the initiator group.",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "The name of the SVM",
				Required:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Required ONTAP 9.9 or greater. Comment",
				Computed:            true,
			},
			"igroups": schema.SetNestedAttribute{
				MarkdownDescription: "Required ONTAP 9.9 or greater. The initiator groups that are members of the group.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"comment": schema.StringAttribute{
							MarkdownDescription: "Comment",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name",
							Computed:            true,
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "UUID",
							Computed:            true,
						},
					},
				},
			},
			"initiators": schema.SetNestedAttribute{
				MarkdownDescription: "Required ONTAP 9.9 or greater. The initiators that are members of the group or any group nested below this group.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"comment": schema.StringAttribute{
							MarkdownDescription: "Comment",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name",
							Computed:            true,
						},
					},
				},
			},
			"lun_maps": schema.SetNestedAttribute{
				MarkdownDescription: "All LUN maps with which the initiator is associated.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"logical_unit_number": schema.Int64Attribute{
							MarkdownDescription: "The logical unit number assigned to the LUN for initiators in the initiator group.",
							Computed:            true,
						},
						"lun": schema.SingleNestedAttribute{
							MarkdownDescription: "The LUN to which the initiator group is mapped",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "The name of the LUN.",
									Computed:            true,
								},
								"uuid": schema.StringAttribute{
									MarkdownDescription: "The UUID of the LUN.",
									Computed:            true,
								},
							},
						},
					},
				},
			},
			"os_type": schema.StringAttribute{
				MarkdownDescription: "The host operating system of the initiator group. All initiators in the group should be hosts of the same operating system.",
				Computed:            true,
			},
			"portset": schema.SingleNestedAttribute{
				MarkdownDescription: "Required ONTAP 9.9 or greater. The portset to which the initiator group is bound. Binding the initiator group to a portset restricts the initiators of the group to accessing mapped LUNs only through network interfaces in the portset.",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "The name of the LUN.",
						Computed:            true,
					},
					"uuid": schema.StringAttribute{
						MarkdownDescription: "The UUID of the LUN.",
						Computed:            true,
					},
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "The protocols supported by the initiator group. This restricts the type of initiators that can be added to the initiator group.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the initiator group.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsSanIgroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsSanIgroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsSanIgroupDataSourceModel

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
	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	restInfo, err := interfaces.GetProtocolsSanIgroupByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsSanIgroup
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Comment = types.StringValue(restInfo.Comment)
	data.OsType = types.StringValue(restInfo.OsType)
	data.Protocol = types.StringValue(restInfo.Protocol)
	data.ID = types.StringValue(restInfo.UUID)
	data.Portset = &ProtocolsSanIgroupDataSourcePortsetModel{
		Name: types.StringValue(restInfo.Portset.Name),
		UUID: types.StringValue(restInfo.Portset.UUID),
	}
	data.Igroups = make([]ProtocolsSanIgroupDataSourceIgroupModel, len(restInfo.Igroups))
	for index, record := range restInfo.Igroups {
		data.Igroups[index] = ProtocolsSanIgroupDataSourceIgroupModel{
			Comment: types.StringValue(record.Comment),
			Name:    types.StringValue(record.Name),
			UUID:    types.StringValue(record.UUID),
		}
	}
	data.Initiators = make([]ProtocolsSanIgroupDataSourceInitiatorModel, len(restInfo.Initiators))
	for index, record := range restInfo.Initiators {
		data.Initiators[index] = ProtocolsSanIgroupDataSourceInitiatorModel{
			Comment: types.StringValue(record.Comment),
			Name:    types.StringValue(record.Name),
		}
	}
	data.LunMaps = make([]ProtocolsSanIgroupDataSourceLunMapModel, len(restInfo.LunMaps))
	for index, record := range restInfo.LunMaps {
		data.LunMaps[index] = ProtocolsSanIgroupDataSourceLunMapModel{
			LogicalUnitNumber: types.Int64Value(int64(record.LogicalUnitNumber)),
			Lun: ProtocolsSanIgroupDataSourceLunModel{
				Name: types.StringValue(record.Lun.Name),
				UUID: types.StringValue(record.Lun.UUID),
			},
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
