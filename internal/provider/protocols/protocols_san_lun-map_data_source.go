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
var _ datasource.DataSource = &ProtocolsSanLunMapDataSource{}

// NewProtocolsSanLunMapDataSource is a helper function to simplify the provider implementation.
func NewProtocolsSanLunMapDataSource() datasource.DataSource {
	return &ProtocolsSanLunMapDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "san_lun-map",
		},
	}
}

// ProtocolsSanLunMapDataSource defines the data source implementation.
type ProtocolsSanLunMapDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsSanLunMapDataSourceModel describes the data source data model.
type ProtocolsSanLunMapDataSourceModel struct {
	CxProfileName     types.String `tfsdk:"cx_profile_name"`
	SVM               svm.SVM      `tfsdk:"svm"`
	Lun               Lun          `tfsdk:"lun"`
	IGroup            IGroup       `tfsdk:"igroup"`
	LogicalUnitNumber types.Int64  `tfsdk:"logical_unit_number"`
}

// Metadata returns the data source type name.
func (d *ProtocolsSanLunMapDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsSanLunMapDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsSanLunMap data source",

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
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsSanLunMapDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsSanLunMapDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsSanLunMapDataSourceModel

	// Read Terraform prior state data into the model
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

	restInfo, err := interfaces.GetProtocolsSanLunMapsByName(errorHandler, *client, data.IGroup.Name.ValueString(), data.Lun.Name.ValueString(), data.SVM.Name.ValueString())
	if err != nil {
		// error reporting done inside GetProtocolsSanLunMaps
		return
	}

	data.LogicalUnitNumber = types.Int64Value(int64(restInfo.LogicalUnitNumber))

	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
