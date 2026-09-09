package protocols

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsS3ServiceDataSource{}

// NewProtocolsS3ServiceDataSource is a helper function to simplify the provider implementation.
func NewProtocolsS3ServiceDataSource() datasource.DataSource {
	return &ProtocolsS3ServiceDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "s3_service",
		},
	}
}

// ProtocolsS3ServiceDataSource defines the data source implementation.
type ProtocolsS3ServiceDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsS3ServiceDataSourceModel describes the data source data model.
type ProtocolsS3ServiceDataSourceModel struct {
	CxProfileName   types.String `tfsdk:"cx_profile_name"`
	SVMName         types.String `tfsdk:"svm_name"`
	Name            types.String `tfsdk:"name"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Comment         types.String `tfsdk:"comment"`
	CertificateName types.String `tfsdk:"certificate_name"`
	IsHTTPEnabled   types.Bool   `tfsdk:"is_http_enabled"`
	IsHTTPSEnabled  types.Bool   `tfsdk:"is_https_enabled"`
	Port            types.Int64  `tfsdk:"port"`
	SecurePort      types.Int64  `tfsdk:"secure_port"`
}

// ProtocolsS3ServiceDataSourceFilterModel describes the data source data model for queries.
type ProtocolsS3ServiceDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
	Name    types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsS3ServiceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsS3ServiceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsS3Service data source",

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
				MarkdownDescription: "The name of the S3 server.",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether or not the S3 server is enabled.",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Additional information about the S3 server.",
				Computed:            true,
			},
			"certificate_name": schema.StringAttribute{
				MarkdownDescription: "Indicates the certificate being used for creating HTTPS connections to the S3 server.",
				Computed:            true,
			},
			"is_http_enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether HTTP is enabled on the S3 server.",
				Computed:            true,
			},
			"is_https_enabled": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether HTTPS is enabled on the S3 server.",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Indicates the HTTP listener port for the S3 server.",
				Computed:            true,
			},
			"secure_port": schema.Int64Attribute{
				MarkdownDescription: "Indicates the HTTPS listener port for the S3 server.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsS3ServiceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsS3ServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsS3ServiceDataSourceModel

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

	restInfo, err := interfaces.GetS3Server(errorHandler, *client, data.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetS3Server
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No SVM audit found", "SVM audit not found")
		return
	}
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.Comment = types.StringValue(restInfo.Comment)
	if restInfo.Certificate.Name != "" {
		data.CertificateName = types.StringValue(restInfo.Certificate.Name)
	} else {
		data.CertificateName = types.StringNull()
	}
	data.IsHTTPEnabled = types.BoolValue(restInfo.IsHTTPEnabled)
	data.IsHTTPSEnabled = types.BoolValue(restInfo.IsHTTPSEnabled)
	data.Port = types.Int64Value(restInfo.Port)
	data.SecurePort = types.Int64Value(restInfo.SecurePort)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
