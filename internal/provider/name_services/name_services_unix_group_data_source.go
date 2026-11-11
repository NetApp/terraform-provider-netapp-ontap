package name_services

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
var _ datasource.DataSource = &NameServicesUnixGroupDataSource{}

// NewNameServicesUnixGroupDataSource is a helper function to simplify the provider implementation.
func NewNameServicesUnixGroupDataSource() datasource.DataSource {
	return &NameServicesUnixGroupDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "unix_group",
		},
	}
}

// NameServicesUnixGroupDataSource defines the data source implementation.
type NameServicesUnixGroupDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesUnixGroupDataSourceModel describes the data source data model.
type NameServicesUnixGroupDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Name          types.String `tfsdk:"name"`
	Users         types.Set    `tfsdk:"users"`
	ID            types.Int64  `tfsdk:"id"`
}

// UserGetDataModel describes the GET record data model using go types for mapping.
type UserGetDataModel struct {
	Name types.String `tfsdk:"name"`
}

// NameServicesUnixGroupDataSourceFilterModel describes the data source data model for queries.
type NameServicesUnixGroupDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
	Name    types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *NameServicesUnixGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *NameServicesUnixGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesUnixGroup data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "The name of the SVM.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the UNIX group.",
				Required:            true,
			},
			"users": schema.SetAttribute{				
				MarkdownDescription: "The list of users associated with the UNIX group.",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "UNIX group ID of the specified group.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NameServicesUnixGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *NameServicesUnixGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NameServicesUnixGroupDataSourceModel

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

	restInfo, err := interfaces.GetNameServicesUnixGroup(errorHandler, *client, data.SVMName.ValueString(), data.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetNameServicesUnixGroup
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No UNIX group found", fmt.Sprintf("UNIX group named %s not found", data.Name.ValueString()))
		return
	}
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.ID = types.Int64Value(restInfo.ID)
	// users - map to set
	var users = make([]string, len(restInfo.Users))
	for i, user := range restInfo.Users {
		users[i] = user.Name
	}
	usersSet, diags := types.SetValueFrom(ctx, types.StringType, users)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Users = usersSet

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
