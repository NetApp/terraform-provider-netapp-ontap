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
var _ datasource.DataSource = &NameServicesUnixGroupsDataSource{}

// NewNameServicesUnixGroupsDataSource is a helper function to simplify the provider implementation.
func NewNameServicesUnixGroupsDataSource() datasource.DataSource {
	return &NameServicesUnixGroupsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "unix_groups",
		},
	}
}

// NameServicesUnixGroupsDataSource defines the data source implementation.
type NameServicesUnixGroupsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesUnixGroupsDataSourceModel describes the data source data model.
type NameServicesUnixGroupsDataSourceModel struct {
	CxProfileName          types.String                                `tfsdk:"cx_profile_name"`
	NameServicesUnixGroups []NameServicesUnixGroupDataSourceModel      `tfsdk:"name_services_unix_groups"`
	Filter                 *NameServicesUnixGroupDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *NameServicesUnixGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *NameServicesUnixGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesUnixGroups data source",

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
						MarkdownDescription: "The name of the UNIX group",
						Optional:            true,
					},
				},
				Required: true,
			},
			"name_services_unix_groups": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "The name of the SVM.",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the UNIX group.",
							Computed:            true,
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
				},
				Computed:            true,
				MarkdownDescription: "List of UNIX groups",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NameServicesUnixGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *NameServicesUnixGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NameServicesUnixGroupsDataSourceModel

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

	var filter *interfaces.NameServicesUnixGroupsFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.NameServicesUnixGroupsFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
			Name:    data.Filter.Name.ValueString(),
		}
	}
	
	restInfo, err := interfaces.GetNameServicesUnixGroups(errorHandler, *client, filter, cluster.Version)
	if err != nil {
		// error reporting done inside GetNameServicesUnixGroups
		return
	}

	data.NameServicesUnixGroups = make([]NameServicesUnixGroupDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		// users - map to set
		var users = make([]string, len(record.Users))
		for i, user := range record.Users {
			users[i] = user.Name
		}
		usersSet, diags := types.SetValueFrom(ctx, types.StringType, users)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.NameServicesUnixGroups[index] = NameServicesUnixGroupDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVMName:       types.StringValue(record.SVM.Name),
			Name:          types.StringValue(record.Name),
			Users:         usersSet,
			ID:            types.Int64Value(record.ID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
