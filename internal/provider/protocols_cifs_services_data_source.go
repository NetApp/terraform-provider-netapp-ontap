package provider

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
var _ datasource.DataSource = &CifsServicesDataSource{}

// NewCifsServicesDataSource is a helper function to simplify the provider implementation.
func NewCifsServicesDataSource() datasource.DataSource {
	return &CifsServicesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_cifs_services_data_source",
		},
	}
}

// CifsServicesDataSource defines the data source implementation.
type CifsServicesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// CifsServicesDataSourceModel describes the data source data model.
type CifsServicesDataSourceModel struct {
	CxProfileName types.String                       `tfsdk:"cx_profile_name"`
	CifsServices  []CifsServiceDataSourceModel       `tfsdk:"protocols_cifs_services"`
	Filter        *CifsServicesDataSourceFilterModel `tfsdk:"filter"`
}

// CifsServicesDataSourceFilterModel describes the data source data model for queries.
type CifsServicesDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *CifsServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *CifsServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsServices data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "CifsService name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "CifsService svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_cifs_services": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the CIFS server",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "Svm name",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Specifies if the CIFS service is administratively enabled",
						},
						"default_unix_user": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Default unix user",
						},
						"comment": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "text comment of up to 48 characters about the CIFS server",
						},
						"ad_domain": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Ad domain",
							Attributes: map[string]schema.Attribute{
								"organizational_unit": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Organizational unit",
								},
								"user": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "User account with the access to add the CIFS server to the Active Directory",
								},
								"password": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Account password used to add this CIFS server to the Active Directory",
								},
								"fqdn": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: " Fully qualified domain name of the Windows Active Directory to which this CIFS server belongs",
								},
							},
						},
						"netbios": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Netbios",
							Attributes: map[string]schema.Attribute{
								"enabled": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "NetBios name service (NBNS) is enabled for the CIFS",
								},
								"aliases": schema.SetAttribute{
									Computed:            true,
									MarkdownDescription: "list of one or more NetBIOS aliases for the CIFS server",
									ElementType:         types.StringType,
								},
								"wins_servers": schema.SetAttribute{
									Computed:            true,
									MarkdownDescription: "list of Windows Internet Name Server (WINS) addresses that manage and map the NetBIOS name of the CIFS server to their network IP addresses. The IP addresses must be IPv4 addresses.",
									ElementType:         types.StringType,
								},
							},
						},
						"security": schema.SingleNestedAttribute{
							Computed:            true,
							MarkdownDescription: "Security",
							Attributes: map[string]schema.Attribute{
								"kdc_encryption": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Specifies whether AES-128 and AES-256 encryption is enabled for all Kerberos-based communication with the Active Directory KDC",
								},
								"restrict_anonymous": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Specifies what level of access an anonymous user is granted",
								},
								"smb_signing": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Specifies if signing is required for incoming CIFS traffic",
								},
								"smb_encryption": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Specifies if encryption is required for incoming CIFS traffic",
								},
								"lm_compatibility_level": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "CIFS server minimum security level",
								},
								"aes_netlogon_enabled": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "An AES session key is enabled for the Netlogon channel",
								},
								"try_ldap_channel_binding": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Specifies whether or not channel binding is attempted in the case of TLS/LDAPS",
								},
								"ldap_referral_enabled": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Specifies if LDAP referral chasing is enabled for AD LDAP connections",
								},
								"encrypt_dc_connection": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Encryption is required for domain controller connections",
								},
								"use_start_tls": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Specifies whether or not to use SSL/TLS for allowing secure LDAP communication with Active Directory LDAP servers",
								},
								"session_security": schema.StringAttribute{
									Computed:            true,
									MarkdownDescription: "Client session security for AD LDAP connections",
								},
								"use_ldaps": schema.BoolAttribute{
									Computed:            true,
									MarkdownDescription: "Specifies whether or not to use use LDAPS for secure Active Directory LDAP connections by using the TLS/SSL protocols",
								},
								"advertised_kdc_encryptions": schema.SetAttribute{
									Computed:            true,
									MarkdownDescription: "List of encryption types that are advertised to the KDC",
									ElementType:         types.StringType,
								},
							},
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "Protocols CIFS services",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CifsServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *CifsServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CifsServicesDataSourceModel

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

	var filter *interfaces.CifsServiceDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.CifsServiceDataSourceFilterModel{
			Name:    data.Filter.Name.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetCifsServices(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetCifsServices
		return
	}

	data.CifsServices = make([]CifsServiceDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.CifsServices[index] = CifsServiceDataSourceModel{
			CxProfileName:   types.String(data.CxProfileName),
			Name:            types.StringValue(record.Name),
			SVMName:         types.StringValue(record.SVM.Name),
			Enabled:         types.BoolValue(record.Enabled),
			DefaultUnixUser: types.StringValue(record.DefaultUnixUser),
			Comment:         types.StringValue(record.Comment),
		}
		// aliases := make([]types.String, len(record.Netbios.Aliases))
		// for i, alias := range record.Netbios.Aliases {
		// 	aliases[i] = types.StringValue(alias)
		// }
		// winsServers := make([]types.String, len(record.Netbios.WinsServers))
		// for i, winsServer := range record.Netbios.WinsServers {
		// 	winsServers[i] = types.StringValue(winsServer)
		// }
		data.CifsServices[index].AdDomain = &AdDomainDataSourceModel{
			OrganizationalUnit: types.StringValue(record.AdDomain.OrganizationalUnit),
			User:               types.StringValue(record.AdDomain.User),
			Password:           types.StringValue(record.AdDomain.Password),
			Fqdn:               types.StringValue(record.AdDomain.Fqdn),
		}
		data.CifsServices[index].Netbios = &NetbiosDataSourceModel{
			Enabled:     types.BoolValue(record.Netbios.Enabled),
			Aliases:     connection.FlattenTypesStringList(record.Netbios.Aliases),
			WinsServers: connection.FlattenTypesStringList(record.Netbios.WinsServers),
		}
		data.CifsServices[index].Security = &CifsSecurityDataSourceModel{
			KdcEncryption:            types.BoolValue(record.Security.KdcEncryption),
			RestrictAnonymous:        types.StringValue(record.Security.RestrictAnonymous),
			SmbSigning:               types.BoolValue(record.Security.SmbSigning),
			SmbEncryption:            types.BoolValue(record.Security.SmbEncryption),
			LmCompatibilityLevel:     types.StringValue(record.Security.LmCompatibilityLevel),
			AesNetlogonEnabled:       types.BoolValue(record.Security.AesNetlogonEnabled),
			TryLdapChannelBinding:    types.BoolValue(record.Security.TryLdapChannelBinding),
			LdapReferralEnabled:      types.BoolValue(record.Security.LdapReferralEnabled),
			EncryptDcConnection:      types.BoolValue(record.Security.EncryptDcConnection),
			UseStartTLS:              types.BoolValue(record.Security.UseStartTLS),
			SessionSecurity:          types.StringValue(record.Security.SessionSecurity),
			UseLdaps:                 types.BoolValue(record.Security.UseLdaps),
			AdvertisedKdcEncryptions: connection.FlattenTypesStringList(record.Security.AdvertisedKdcEncryptions),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
