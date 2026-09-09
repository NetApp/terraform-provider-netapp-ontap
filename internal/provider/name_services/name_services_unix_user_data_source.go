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
var _ datasource.DataSource = &NameServicesUnixUserDataSource{}

// NewNameServicesUnixUserDataSource is a helper function to simplify the provider implementation.
func NewNameServicesUnixUserDataSource() datasource.DataSource {
	return &NameServicesUnixUserDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "unix_user",
		},
	}
}

// NameServicesUnixUserDataSource defines the data source implementation.
type NameServicesUnixUserDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesUnixUserDataSourceModel describes the data source data model.
type NameServicesUnixUserDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Name          types.String `tfsdk:"name"`
	FullName      types.String `tfsdk:"full_name"`
	PrimaryGID    types.Int64  `tfsdk:"primary_gid"`
	ID            types.Int64  `tfsdk:"id"`
}

// NameServicesUnixUserDataSourceFilterModel describes the data source data model for queries.
type NameServicesUnixUserDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
	Name    types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *NameServicesUnixUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *NameServicesUnixUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesUnixUser data source",

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
				MarkdownDescription: "The name of the UNIX user.",
				Required:            true,
			},
			"full_name": schema.StringAttribute{
				MarkdownDescription: "Full name of the UNIX user.",
				Computed:            true,
			},
			"primary_gid": schema.Int64Attribute{
				MarkdownDescription: "Primary group ID to which the UNIX user belongs.",
				Computed:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "UNIX user ID of the specified user.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NameServicesUnixUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *NameServicesUnixUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NameServicesUnixUserDataSourceModel

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

	restInfo, err := interfaces.GetNameServicesUnixUser(errorHandler, *client, data.SVMName.ValueString(), data.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetNameServicesUnixUser
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No UNIX user found", fmt.Sprintf("UNIX user named %s not found", data.Name.ValueString()))
		return
	}
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.FullName = types.StringValue(restInfo.FullName)
	data.PrimaryGID = types.Int64Value(restInfo.PrimaryGID)
	data.ID = types.Int64Value(restInfo.ID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
