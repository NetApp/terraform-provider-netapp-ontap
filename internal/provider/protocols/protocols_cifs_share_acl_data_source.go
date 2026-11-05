package protocols

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsCIFSShareACLDataSource{}

// NewProtocolsCIFSShareACLDataSource is a helper function to simplify the provider implementation.
func NewProtocolsCIFSShareACLDataSource() datasource.DataSource {
	return &ProtocolsCIFSShareACLDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cifs_share_acl",
		},
	}
}

// NewProtocolsCIFSShareACLDataSourceAlias is a helper function to simplify the provider implementation.
func NewProtocolsCIFSShareACLDataSourceAlias() datasource.DataSource {
	return &ProtocolsCIFSShareACLDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_cifs_share_acl_data_source",
		},
	}
}

// ProtocolsCIFSShareACLDataSource defines the data source implementation.
type ProtocolsCIFSShareACLDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsCIFSShareACLDataSourceModel describes the data source data model.
type ProtocolsCIFSShareACLDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	ShareName     types.String `tfsdk:"share_name"`
	UserOrGroup   types.String `tfsdk:"user_or_group"`
	Type          types.String `tfsdk:"type"`
	Permission    types.String `tfsdk:"permission"`
}

// Metadata returns the data source type name.
func (d *ProtocolsCIFSShareACLDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsCIFSShareACLDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsCIFSShareACL data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the SVM that owns the CIFS share",
				Required:            true,
			},
			"share_name": schema.StringAttribute{
				MarkdownDescription: "Name of the CIFS share that owns this ACL",
				Required:            true,
			},
			"user_or_group": schema.StringAttribute{
				MarkdownDescription: "Specifies the user or group name in the access control list of a CIFS share.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Specifies the type of the user or group in the access control list of a CIFS share.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("windows", "unix_user", "unix_group"),
				},
			},
			"permission": schema.StringAttribute{
				MarkdownDescription: "Specifies the access rights that a user or group has on the defined CIFS Share.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsCIFSShareACLDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsCIFSShareACLDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsCIFSShareACLDataSourceModel

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

	restInfo, err := interfaces.GetProtocolsCIFSShareACL(errorHandler, *client, svm.UUID, data.ShareName.ValueString(), data.UserOrGroup.ValueString(), data.Type.ValueString())
	if err != nil {
		// error reporting done inside GetProtocolsCIFSShareACL
		return
	}

	data.UserOrGroup = types.StringValue(restInfo.UserOrGroup)
	data.Type = types.StringValue(restInfo.Type)
	data.Permission = types.StringValue(restInfo.Permission)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
