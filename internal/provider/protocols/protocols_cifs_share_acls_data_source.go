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
var _ datasource.DataSource = &ProtocolsCIFSShareACLsDataSource{}

// NewProtocolsCIFSShareACLsDataSource is a helper function to simplify the provider implementation.
func NewProtocolsCIFSShareACLsDataSource() datasource.DataSource {
	return &ProtocolsCIFSShareACLsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cifs_share_acls",
		},
	}
}

// NewProtocolsCIFSShareACLsDataSourceAlias is a helper function to simplify the provider implementation.
func NewProtocolsCIFSShareACLsDataSourceAlias() datasource.DataSource {
	return &ProtocolsCIFSShareACLsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_cifs_share_acls_data_source",
		},
	}
}

// ProtocolsCIFSShareACLsDataSource defines the data source implementation.
type ProtocolsCIFSShareACLsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsCIFSShareACLsDataSourceModel describes the data source data model.
type ProtocolsCIFSShareACLsDataSourceModel struct {
	CxProfileName          types.String                                 `tfsdk:"cx_profile_name"`
	SVMName                types.String                                 `tfsdk:"svm_name"`
	ProtocolsCIFSShareACLs []ProtocolsCIFSShareACLDataSourceModel       `tfsdk:"protocols_cifs_share_acls"`
	Filter                 *ProtocolsCIFSShareACLsDataSourceFilterModel `tfsdk:"filter"`
}

// ProtocolsCIFSShareACLsDataSourceFilterModel describes the data source data model for queries.
type ProtocolsCIFSShareACLsDataSourceFilterModel struct {
	ShareName types.String `tfsdk:"share_name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsCIFSShareACLsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsCIFSShareACLsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsCIFSShareACLs data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the SVM that owns the CIFS shares",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"share_name": schema.StringAttribute{
						MarkdownDescription: "Name of the CIFS share",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_cifs_share_acls": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "Name of the SVM that owns the CIFS share",
							Computed:            true,
						},
						"share_name": schema.StringAttribute{
							MarkdownDescription: "Name of the CIFS share that owns this ACL",
							Computed:            true,
						},
						"user_or_group": schema.StringAttribute{
							MarkdownDescription: "Specifies the user or group name in the access control list of a CIFS share.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Specifies the type of the user or group in the access control list of a CIFS share.",
							Computed:            true,
						},
						"permission": schema.StringAttribute{
							MarkdownDescription: "Specifies the access rights that a user or group has on the defined CIFS Share.",
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
func (d *ProtocolsCIFSShareACLsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsCIFSShareACLsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsCIFSShareACLsDataSourceModel

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

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		return
	}

	var filter *interfaces.ProtocolsCIFSShareACLDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.ProtocolsCIFSShareACLDataSourceFilterModel{
			Share: data.Filter.ShareName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetProtocolsCIFSShareAcls(errorHandler, *client, filter, svm.UUID)
	if err != nil {
		// error reporting done inside GetProtocolsCIFSShareAcls
		return
	}

	data.ProtocolsCIFSShareACLs = make([]ProtocolsCIFSShareACLDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.ProtocolsCIFSShareACLs[index] = ProtocolsCIFSShareACLDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVMName:       types.StringValue(record.SVM.Name),
			ShareName:     types.StringValue(record.Share),
			UserOrGroup:   types.StringValue(record.UserOrGroup),
			Type:          types.StringValue(record.Type),
			Permission:    types.StringValue(record.Permission),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
