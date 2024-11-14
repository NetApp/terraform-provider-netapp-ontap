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
var _ datasource.DataSource = &CifsLocalGroupMembersDataSource{}

// NewCifsLocalGroupMembersDataSource is a helper function to simplify the provider implementation.
func NewCifsLocalGroupMembersDataSource() datasource.DataSource {
	return &CifsLocalGroupMembersDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cifs_local_group_members",
		},
	}
}

// NewCifsLocalGroupMembersDataSourceAlias is a helper function to simplify the provider implementation.
func NewCifsLocalGroupMembersDataSourceAlias() datasource.DataSource {
	return &CifsLocalGroupMembersDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_cifs_local_group_members_data_source",
		},
	}
}

// CifsLocalGroupMembersDataSource defines the data source implementation.
type CifsLocalGroupMembersDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// CifsLocalGroupMembersDataSourceModel describes the data source data model.
type CifsLocalGroupMembersDataSourceModel struct {
	CxProfileName types.String   `tfsdk:"cx_profile_name"`
	GroupName     types.String   `tfsdk:"group_name"`
	SVMName       types.String   `tfsdk:"svm_name"`
	Members       []types.String `tfsdk:"members"`
}

// Metadata returns the data source type name.
func (d *CifsLocalGroupMembersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *CifsLocalGroupMembersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsLocalGroupMembers data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"group_name": schema.StringAttribute{
				MarkdownDescription: "Local group name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface svm name",
				Required:            true,
			},
			"members": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of members",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CifsLocalGroupMembersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *CifsLocalGroupMembersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CifsLocalGroupMembersDataSourceModel

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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("invalid svm name", fmt.Sprintf("protocols_cifs_local_group_members svm_name %s is invalid", data.SVMName.ValueString()))
		return
	}

	restInfo, err := interfaces.GetCifsLocalGroupByName(errorHandler, *client, data.GroupName.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetCifsLocalGroup
		errorHandler.MakeAndReportError("invalid group name", fmt.Sprintf("protocols_cifs_local_group_members group_name %s is invalid", data.GroupName.ValueString()))
		return
	}

	restInfoMembers, err := interfaces.GetCifsLocalGroupMembers(errorHandler, *client, svm.UUID, restInfo.SID)
	if err != nil {
		// error reporting done inside GetCifsLocalGroupMember
		return
	}

	data.Members = make([]types.String, len(restInfoMembers))
	for index, record := range restInfoMembers {
		data.Members[index] = types.StringValue(record.Name)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
