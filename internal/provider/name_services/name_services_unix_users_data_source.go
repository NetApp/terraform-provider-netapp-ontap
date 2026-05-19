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
var _ datasource.DataSource = &NameServicesUnixUsersDataSource{}

// NewNameServicesUnixUsersDataSource is a helper function to simplify the provider implementation.
func NewNameServicesUnixUsersDataSource() datasource.DataSource {
	return &NameServicesUnixUsersDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "unix_users",
		},
	}
}

// NameServicesUnixUsersDataSource defines the data source implementation.
type NameServicesUnixUsersDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesUnixUsersDataSourceModel describes the data source data model.
type NameServicesUnixUsersDataSourceModel struct {
	CxProfileName         types.String                               `tfsdk:"cx_profile_name"`
	NameServicesUnixUsers []NameServicesUnixUserDataSourceModel      `tfsdk:"name_services_unix_users"`
	Filter                *NameServicesUnixUserDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *NameServicesUnixUsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *NameServicesUnixUsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesUnixUsers data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "The name of the SVM",
						Optional:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "The name of the UNIX user",
						Optional:            true,
					},
				},
				Required: true,
			},
			"name_services_unix_users": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
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
				},
				Computed:            true,
				MarkdownDescription: "List of UNIX users",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NameServicesUnixUsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *NameServicesUnixUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NameServicesUnixUsersDataSourceModel

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

	if data.Filter == nil || data.Filter.SVMName.IsNull() {
		errorHandler.MakeAndReportError("No SVM specified", "svm_name must be specified in filter")
		return
	}

	var filter *interfaces.NameServicesUnixUsersFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.NameServicesUnixUsersFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
			Name:    data.Filter.Name.ValueString(),
		}
	}
	
	restInfo, err := interfaces.GetNameServicesUnixUsers(errorHandler, *client, filter, cluster.Version)
	if err != nil {
		// error reporting done inside GetNameServicesUnixUsers
		return
	}

	data.NameServicesUnixUsers = make([]NameServicesUnixUserDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.NameServicesUnixUsers[index] = NameServicesUnixUserDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVMName:       types.StringValue(record.SVM.Name),
			Name:          types.StringValue(record.Name),
			FullName:      types.StringValue(record.FullName),
			PrimaryGID:    types.Int64Value(record.PrimaryGID),
			ID:            types.Int64Value(record.ID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
