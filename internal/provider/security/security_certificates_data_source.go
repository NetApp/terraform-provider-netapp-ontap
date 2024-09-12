package security

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
var _ datasource.DataSource = &SecurityCertificatesDataSource{}

// NewSecurityCertificatesDataSource is a helper function to simplify the provider implementation.
func NewSecurityCertificatesDataSource() datasource.DataSource {
	return &SecurityCertificatesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "security_certificates",
		},
	}
}

// SecurityCertificatesDataSource defines the data source implementation.
type SecurityCertificatesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// SecurityCertificatesDataSourceModel describes the data source data model.
type SecurityCertificatesDataSourceModel struct {
	CxProfileName        types.String                               `tfsdk:"cx_profile_name"`
	SecurityCertificates []SecurityCertificateDataSourceModel       `tfsdk:"security_certificates"`
	Filter               *SecurityCertificatesDataSourceFilterModel `tfsdk:"filter"`
}

// SecurityCertificatesDataSourceFilterModel describes the data source data model for queries.
type SecurityCertificatesDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
	Scope   types.String `tfsdk:"scope"`
}

// Metadata returns the data source type name.
func (d *SecurityCertificatesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *SecurityCertificatesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityCertificates data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "SVM name in which the certificate is installed.",
						Optional:            true,
					},
					"scope": schema.StringAttribute{
						MarkdownDescription: "Set to 'svm' for certificates installed in a SVM. Otherwise, set to 'cluster'.",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"security_certificates": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
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
						"public_certificate": schema.StringAttribute{
							MarkdownDescription: "Public key Certificate in PEM format.",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "Certificate uuid.",
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
func (d *SecurityCertificatesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SecurityCertificatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityCertificatesDataSourceModel

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

	var filter *interfaces.SecurityCertificateDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.SecurityCertificateDataSourceFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
			Scope:   data.Filter.Scope.ValueString(),
		}
	}
	restInfo, err := interfaces.GetSecurityCertificates(errorHandler, *client, cluster.Version, filter)
	if err != nil {
		// error reporting done inside GetSecurityCertificates
		return
	}

	data.SecurityCertificates = make([]SecurityCertificateDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.SecurityCertificates[index] = SecurityCertificateDataSourceModel{
			CxProfileName:     types.String(data.CxProfileName),
			Name:              types.StringValue(record.Name),
			CommonName:        types.StringValue(record.CommonName),
			SVMName:           types.StringValue(record.SVM.Name),
			Scope:             types.StringValue(record.Scope),
			Type:              types.StringValue(record.Type),
			SerialNumber:      types.StringValue(record.SerialNumber),
			CA:                types.StringValue(record.CA),
			HashFunction:      types.StringValue(record.HashFunction),
			KeySize:           types.Int64Value(record.KeySize),
			ExpiryTime:        types.StringValue(record.ExpiryTime),
			PublicCertificate: types.StringValue(record.PublicCertificate),
			ID:            	   types.StringValue(record.UUID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
