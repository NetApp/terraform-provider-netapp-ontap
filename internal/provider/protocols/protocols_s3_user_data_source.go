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
var _ datasource.DataSource = &ProtocolsS3UserDataSource{}

// NewProtocolsS3UserDataSource is a helper function to simplify the provider implementation.
func NewProtocolsS3UserDataSource() datasource.DataSource {
	return &ProtocolsS3UserDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "s3_user",
		},
	}
}

// ProtocolsS3UserDataSource defines the data source implementation.
type ProtocolsS3UserDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsS3UserDataSourceModel describes the data source data model.
type ProtocolsS3UserDataSourceModel struct {
	CxProfileName  types.String  `tfsdk:"cx_profile_name"`
	SVMName        types.String  `tfsdk:"svm_name"`
	Name           types.String  `tfsdk:"name"`
	Comment        types.String  `tfsdk:"comment"`
	AccessKey      types.String  `tfsdk:"access_key"`
	KeyExpiryTime  types.String  `tfsdk:"key_expiry_time"`
	KeyTimeToLive  types.String  `tfsdk:"key_time_to_live"`
	ID             types.Int64   `tfsdk:"id"`
}

// ProtocolsS3UserDataSourceFilterModel describes the data source data model for queries.
type ProtocolsS3UserDataSourceFilterModel struct {
	SVMName  types.String  `tfsdk:"svm_name"`
	Name     types.String  `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsS3UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsS3UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsS3User data source",

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
				MarkdownDescription: "The name of the S3 user.",
				Required:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Additional information about the user.",
				Computed:            true,
			},
			"access_key": schema.StringAttribute{
				MarkdownDescription: "Access key for the S3 user.",
				Computed:            true,
			},
			"key_expiry_time": schema.StringAttribute{
				MarkdownDescription: "The date and time after which keys expire and are no longer valid.",
				Computed:            true,
			},
			"key_time_to_live": schema.StringAttribute{
				MarkdownDescription: "The time period in ISO 8601 format for which the keys are valid.",
				Computed:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "The UUID of the S3 user.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsS3UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsS3UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsS3UserDataSourceModel

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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", fmt.Sprintf("SVM named %s not found", data.SVMName.ValueString()))
		return
	}

	restInfo, err := interfaces.GetProtocolsS3User(errorHandler, *client, svm.UUID, data.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsS3User
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No S3 user found", fmt.Sprintf("S3 user named %s not found", data.Name.ValueString()))
		return
	}
	data.Name = types.StringValue(restInfo.Name)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Comment = types.StringValue(restInfo.Comment)
	data.AccessKey = types.StringValue(restInfo.AccessKey)
	data.KeyExpiryTime = types.StringValue(restInfo.KeyExpiryTime)
	data.KeyTimeToLive = types.StringValue(restInfo.KeyTimeToLive)
	data.ID = types.Int64Value(restInfo.ID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
