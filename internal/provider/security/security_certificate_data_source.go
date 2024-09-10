package security

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework-validators/datasourcevalidator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SecurityCertificateDataSource{}

// NewSecurityCertificateDataSource is a helper function to simplify the provider implementation.
func NewSecurityCertificateDataSource() datasource.DataSource {
	return &SecurityCertificateDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "security_certificate",
		},
	}
}

// SecurityCertificateDataSource defines the data source implementation.
type SecurityCertificateDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// SecurityCertificateDataSourceModel describes the data source data model.
type SecurityCertificateDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	CommonName    types.String `tfsdk:"common_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Scope         types.String `tfsdk:"scope"`
	Type          types.String `tfsdk:"type"`
	SerialNumber  types.String `tfsdk:"serial_number"`
	CA            types.String `tfsdk:"ca"`
	HashFunction  types.String `tfsdk:"hash_function"`
	KeySize       types.Int64  `tfsdk:"key_size"`
	ExpiryTime    types.String `tfsdk:"expiry_time"`
	ID            types.String `tfsdk:"id"`
}

// Metadata returns the data source type name.
func (d *SecurityCertificateDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *SecurityCertificateDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityCertificate data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The unique name of the security certificate per SVM.",
				Optional:            true,
			},
			"common_name": schema.StringAttribute{
				MarkdownDescription: "Common name of the certificate.",
				Optional:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of Certificate.",
				Optional:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "SVM name in which the certificate is installed.",
				Optional:            true,
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "Set to 'svm' for certificates installed in a SVM. Otherwise, set to 'cluster'.",
				Computed:            true,
			},
			"serial_number": schema.StringAttribute{
				MarkdownDescription: "Serial number of certificate.",
				Computed:            true,
			},
			"ca": schema.StringAttribute{
				MarkdownDescription: "Certificate authority.",
				Computed:            true,
			},
			"hash_function": schema.StringAttribute{
				MarkdownDescription: "Hashing function.",
				Computed:            true,
			},
			"key_size": schema.Int64Attribute{
				MarkdownDescription: "Key size of the certificate in bits.",
				Computed:            true,
			},
			"expiry_time": schema.StringAttribute{
				MarkdownDescription: "Certificate expiration time, in ISO 8601 duration format or date and time format.",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Certificate uuid.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *SecurityCertificateDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// ConfigValidators validates entire data source configurations
func (d *SecurityCertificateDataSource) ConfigValidators(ctx context.Context) []datasource.ConfigValidator {
    return []datasource.ConfigValidator{
        datasourcevalidator.AtLeastOneOf(
            path.MatchRoot("name"),
            path.MatchRoot("common_name"),
        ),
		datasourcevalidator.RequiredTogether(
            path.MatchRoot("common_name"),
			path.MatchRoot("type"),
        ),
    }
}

// Read refreshes the Terraform state with the latest data.
func (d *SecurityCertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityCertificateDataSourceModel

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

	restInfo, err := interfaces.GetSecurityCertificate(errorHandler, *client, cluster.Version, data.Name.ValueString(), data.CommonName.ValueString(), data.Type.ValueString())
	if err != nil {
		// error reporting done inside GetSecurityCertificate
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.CommonName = types.StringValue(restInfo.CommonName)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Scope = types.StringValue(restInfo.Scope)
	data.Type = types.StringValue(restInfo.Type)
	data.SerialNumber = types.StringValue(restInfo.SerialNumber)
	data.CA = types.StringValue(restInfo.CA)
	data.HashFunction = types.StringValue(restInfo.HashFunction)
	data.KeySize = types.Int64Value(restInfo.KeySize)
	data.ExpiryTime = types.StringValue(restInfo.ExpiryTime)
	data.ID = types.StringValue(restInfo.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
