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
var _ datasource.DataSource = &ProtocolsS3GroupDataSource{}

// NewProtocolsS3GroupDataSource is a helper function to simplify the provider implementation.
func NewProtocolsS3GroupDataSource() datasource.DataSource {
	return &ProtocolsS3GroupDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "s3_group",
		},
	}
}

// ProtocolsS3GroupDataSource defines the data source implementation.
type ProtocolsS3GroupDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsS3GroupDataSourceModel describes the data source data model.
type ProtocolsS3GroupDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	Comment       types.String `tfsdk:"comment"`
	Users         []string     `tfsdk:"users"`
	Policies      []string     `tfsdk:"policies"`
	SVMName       types.String `tfsdk:"svm_name"`
	ID            types.Int64  `tfsdk:"id"`
}

// UserGetDataModel describes the GET record data model using go types for mapping.
type UserGetDataModel struct {
	Name types.String `tfsdk:"name"`
}

// PolicyGetDataModel describes the GET record data model using go types for mapping.
type PolicyGetDataModel struct {
	Name types.String `tfsdk:"name"`
}

// ProtocolsS3GroupDataSourceFilterModel describes the data source data model for queries.
type ProtocolsS3GroupDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsS3GroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsS3GroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsS3Group data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "The name of the SVM",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the S3 group",
				Required:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Additional information about the group",
				Computed:            true,
			},
			"users": schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "The list of users who belong to the group",
				ElementType:         types.StringType,
			},
			"policies": schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "The list of policies that are attached to the group",
				ElementType:         types.StringType,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "S3 Group id",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsS3GroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsS3GroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsS3GroupDataSourceModel

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

	svmUUID, err := interfaces.GetSVMUUID(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSVMUUID
		return
	}
	if svmUUID == "" {
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	restInfo, err := interfaces.GetProtocolsS3Group(errorHandler, *client, data.Name.ValueString(), svmUUID, cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsS3Group
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No S3 group found", "S3 group not found")
		return
	}
	data.Name = types.StringValue(restInfo.Name)
	data.Comment = types.StringValue(restInfo.Comment)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	
	// Users - map to simple string list
	var users = make([]string, len(restInfo.Users))
	for i, user := range restInfo.Users {
		users[i] = user.Name
	}
	data.Users = users
	
	// Policies - map to simple string list
	var policies = make([]string, len(restInfo.Policies))
	for i, policy := range restInfo.Policies {
		policies[i] = policy.Name
	}
	data.Policies = policies
	data.ID = types.Int64Value(restInfo.ID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
