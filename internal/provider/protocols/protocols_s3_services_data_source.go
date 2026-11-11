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
var _ datasource.DataSource = &ProtocolsS3ServicesDataSource{}

// NewProtocolsSVMAuditsDataSource is a helper function to simplify the provider implementation.
func NewProtocolsS3ServicesDataSource() datasource.DataSource {
	return &ProtocolsS3ServicesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "s3_services",
		},
	}
}

// ProtocolsS3ServicesDataSource defines the data source implementation.
type ProtocolsS3ServicesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsS3ServicesDataSourceModel describes the data source data model.
type ProtocolsS3ServicesDataSourceModel struct {
	CxProfileName       types.String                             `tfsdk:"cx_profile_name"`
	ProtocolsS3Services []ProtocolsS3ServiceDataSourceModel      `tfsdk:"protocols_s3_services"`
	Filter              *ProtocolsS3ServiceDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *ProtocolsS3ServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsS3ServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsS3Services data source",

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
						MarkdownDescription: "The name of the S3 service.",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_s3_services": schema.ListNestedAttribute{
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
				},
				Computed:            true,
				MarkdownDescription: "",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsS3ServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsS3ServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsS3ServicesDataSourceModel

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

	var filter *interfaces.ProtocolsS3ServicesFilterModel
	if data.Filter != nil {
		filter = &interfaces.ProtocolsS3ServicesFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
			Name:    data.Filter.Name.ValueString(),
		}
	}
	restInfo, err := interfaces.GetS3Servers(errorHandler, *client, filter, cluster.Version)
	if err != nil {
		// error reporting done inside GetS3Servers
		return
	}

	data.ProtocolsS3Services = make([]ProtocolsS3ServiceDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		// certificate_name
		certificateName := types.StringNull()
		if record.Certificate.Name != "" {
			certificateName = types.StringValue(record.Certificate.Name)
		}
		data.ProtocolsS3Services[index] = ProtocolsS3ServiceDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVMName:       types.StringValue(record.SVM.Name),
			Name:          types.StringValue(record.Name),
			Enabled:       types.BoolValue(record.Enabled),
			Comment:       types.StringValue(record.Comment),
			CertificateName: certificateName,
			IsHTTPEnabled:   types.BoolValue(record.IsHTTPEnabled),
			IsHTTPSEnabled:  types.BoolValue(record.IsHTTPSEnabled),
			Port:            types.Int64Value(record.Port),
			SecurePort:      types.Int64Value(record.SecurePort),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
