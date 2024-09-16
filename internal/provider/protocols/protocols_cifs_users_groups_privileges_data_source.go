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
var _ datasource.DataSource = &CifsUserGroupPrivilegesDataSource{}

// NewCifsUserGroupPrivilegesDataSource is a helper function to simplify the provider implementation.
func NewCifsUsersGroupsPrivilegesDataSource() datasource.DataSource {
	return &CifsUserGroupPrivilegesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cifs_users_groups_privileges",
		},
	}
}

// CifsUserGroupPrivilegesDataSource defines the data source implementation.
type CifsUserGroupPrivilegesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// CifsUserGroupPrivilegesDataSourceModel describes the data source data model.
type CifsUserGroupPrivilegesDataSourceModel struct {
	CxProfileName           types.String                                  `tfsdk:"cx_profile_name"`
	CifsUserGroupPrivileges []CifsUserGroupPrivilegeDataSourceModel       `tfsdk:"protocols_cifs_user_group_privileges"`
	Filter                  *CifsUserGroupPrivilegesDataSourceFilterModel `tfsdk:"filter"`
}

// CifsUserGroupPrivilegesDataSourceFilterModel describes the data source data model for queries.
type CifsUserGroupPrivilegesDataSourceFilterModel struct {
	Name       types.String `tfsdk:"name"`
	SVMName    types.String `tfsdk:"svm_name"`
	Privileges types.String `tfsdk:"privileges"` //only support one privilege search
}

// Metadata returns the data source type name.
func (d *CifsUserGroupPrivilegesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *CifsUserGroupPrivilegesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsUserGroupPrivileges data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "CifsUserGroupPrivilege name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "CifsUserGroupPrivilege svm name",
						Optional:            true,
					},
					"privileges": schema.StringAttribute{
						MarkdownDescription: "CifsUserGroupPrivilege privileges",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_cifs_user_group_privileges": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
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
							MarkdownDescription: "CifsUserGroupPrivilege svm name",
							Required:            true,
						},
						"privileges": schema.SetAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "List of privileges",
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
func (d *CifsUserGroupPrivilegesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *CifsUserGroupPrivilegesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CifsUserGroupPrivilegesDataSourceModel

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

	var filter *interfaces.CifsUserGroupPrivilegeDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.CifsUserGroupPrivilegeDataSourceFilterModel{
			Name:       data.Filter.Name.ValueString(),
			SVMName:    data.Filter.SVMName.ValueString(),
			Privileges: data.Filter.Privileges.ValueString(),
		}
	}
	restInfo, err := interfaces.GetCifsUserGroupPrivileges(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetCifsUserGroupPrivileges
		return
	}

	data.CifsUserGroupPrivileges = make([]CifsUserGroupPrivilegeDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.CifsUserGroupPrivileges[index] = CifsUserGroupPrivilegeDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
			Privileges:    make([]types.String, len(record.Privileges)),
			SVMName:       types.StringValue(record.SVM.Name),
		}
		for idx, privilege := range record.Privileges {
			data.CifsUserGroupPrivileges[index].Privileges[idx] = types.StringValue(privilege)
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
