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

// TODO:
// copy this file to match you data source (should match internal/provider/protocols_cifs_user_group_privilege_data_source.go)
// replace CifsUserGroupPrivilege with the name of the resource, following go conventions, eg IPInterface
// replace protocols_cifs_user_group_privilege with the name of the resource, for logging purposes, eg ip_interface
// make sure to create internal/interfaces/protocols_cifs_user_group_privilege.go too)
// delete these 5 lines

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &CifsUserGroupPrivilegeDataSource{}

// NewCifsUserGroupPrivilegeDataSource is a helper function to simplify the provider implementation.
func NewCifsUserGroupPrivilegeDataSource() datasource.DataSource {
	return &CifsUserGroupPrivilegeDataSource{
		config: resourceOrDataSourceConfig{
			name: "protocols_cifs_user_group_privilege_data_source",
		},
	}
}

// CifsUserGroupPrivilegeDataSource defines the data source implementation.
type CifsUserGroupPrivilegeDataSource struct {
	config resourceOrDataSourceConfig
}

// CifsUserGroupPrivilegeDataSourceModel describes the data source data model.
type CifsUserGroupPrivilegeDataSourceModel struct {
	CxProfileName types.String   `tfsdk:"cx_profile_name"`
	Name          types.String   `tfsdk:"name"`
	SVMName       types.String   `tfsdk:"svm_name"`
	Privileges    []types.String `tfsdk:"privileges"`
}

// Metadata returns the data source type name.
func (d *CifsUserGroupPrivilegeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *CifsUserGroupPrivilegeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsUserGroupPrivilege data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "CifsUserGroupPrivilege name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface svm name",
				Required:            true,
			},
			"privileges": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of privileges",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CifsUserGroupPrivilegeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *CifsUserGroupPrivilegeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CifsUserGroupPrivilegeDataSourceModel

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

	restInfo, err := interfaces.GetCifsUserGroupPrivilegeByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetCifsUserGroupPrivilege
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.Privileges = make([]types.String, len(restInfo.Privileges))
	for index, privilege := range restInfo.Privileges {
		data.Privileges[index] = types.StringValue(privilege)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
