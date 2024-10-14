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
var _ datasource.DataSource = &CifsServiceDataSource{}

// NewCifsServiceDataSource is a helper function to simplify the provider implementation.
func NewCifsServiceDataSource() datasource.DataSource {
	return &CifsServiceDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cifs_service",
		},
	}
}

// NewCifsServiceDataSourceAlias is a helper function to simplify the provider implementation.
func NewCifsServiceDataSourceAlias() datasource.DataSource {
	return &CifsServiceDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_cifs_service_data_source",
		},
	}
}

// CifsServiceDataSource defines the data source implementation.
type CifsServiceDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// CifsServiceDataSourceModel describes the data source data model.
type CifsServiceDataSourceModel struct {
	CxProfileName   types.String                 `tfsdk:"cx_profile_name"`
	Name            types.String                 `tfsdk:"name"`
	SVMName         types.String                 `tfsdk:"svm_name"`
	Enabled         types.Bool                   `tfsdk:"enabled"`
	DefaultUnixUser types.String                 `tfsdk:"default_unix_user"`
	Comment         types.String                 `tfsdk:"comment"`
	AdDomain        *AdDomainDataSourceModel     `tfsdk:"ad_domain"`
	Netbios         *NetbiosDataSourceModel      `tfsdk:"netbios"`
	Security        *CifsSecurityDataSourceModel `tfsdk:"security"`
}

// AdDomainDataSourceModel describes the ad_domain data model using go types for mapping.
type AdDomainDataSourceModel struct {
	OrganizationalUnit types.String `tfsdk:"organizational_unit"`
	User               types.String `tfsdk:"user"`
	Password           types.String `tfsdk:"password"`
	Fqdn               types.String `tfsdk:"fqdn"`
}

// NetbiosDataSourceModel describes the netbios data model using go types for mapping.
type NetbiosDataSourceModel struct {
	Enabled     types.Bool     `tfsdk:"enabled"`
	Aliases     []types.String `tfsdk:"aliases"`
	WinsServers []types.String `tfsdk:"wins_servers"`
}

// CifsSecurityDataSourceModel describes the security data model using go types for mapping.
type CifsSecurityDataSourceModel struct {
	RestrictAnonymous        types.String   `tfsdk:"restrict_anonymous"`
	SmbSigning               types.Bool     `tfsdk:"smb_signing"`
	SmbEncryption            types.Bool     `tfsdk:"smb_encryption"`
	KdcEncryption            types.Bool     `tfsdk:"kdc_encryption"`
	LmCompatibilityLevel     types.String   `tfsdk:"lm_compatibility_level"`
	AesNetlogonEnabled       types.Bool     `tfsdk:"aes_netlogon_enabled"`
	TryLdapChannelBinding    types.Bool     `tfsdk:"try_ldap_channel_binding"`
	LdapReferralEnabled      types.Bool     `tfsdk:"ldap_referral_enabled"`
	EncryptDcConnection      types.Bool     `tfsdk:"encrypt_dc_connection"`
	UseStartTLS              types.Bool     `tfsdk:"use_start_tls"`
	SessionSecurity          types.String   `tfsdk:"session_security"`
	UseLdaps                 types.Bool     `tfsdk:"use_ldaps"`
	AdvertisedKdcEncryptions []types.String `tfsdk:"advertised_kdc_encryptions"`
}

// Metadata returns the data source type name.
func (d *CifsServiceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *CifsServiceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsService data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the CIFS server",
				Required:            true,
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
					"aliases": schema.ListAttribute{
						Computed:            true,
						MarkdownDescription: "list of one or more NetBIOS aliases for the CIFS server",
						ElementType:         types.StringType,
					},
					"wins_servers": schema.ListAttribute{
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *CifsServiceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *CifsServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CifsServiceDataSourceModel

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

	restInfo, err := interfaces.GetCifsServiceByName(errorHandler, *client, data.Name.ValueString())
	if err != nil {
		// error reporting done inside GetCifsService
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	if len(restInfo.Comment) != 0 {
		data.Comment = types.StringValue(restInfo.Comment)
	}
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.DefaultUnixUser = types.StringValue(restInfo.DefaultUnixUser)
	data.AdDomain = &AdDomainDataSourceModel{
		OrganizationalUnit: types.StringValue(restInfo.AdDomain.OrganizationalUnit),
		User:               types.StringValue(restInfo.AdDomain.User),
		Password:           types.StringValue(restInfo.AdDomain.Password),
		Fqdn:               types.StringValue(restInfo.AdDomain.Fqdn),
	}

	// aliases := make([]types.String, len(restInfo.Netbios.Aliases))
	// for i, alias := range restInfo.Netbios.Aliases {
	// 	aliases[i] = types.StringValue(alias)
	// }
	// winsServers := make([]types.String, len(restInfo.Netbios.WinsServers))
	// for i, winsServer := range restInfo.Netbios.WinsServers {
	// 	winsServers[i] = types.StringValue(winsServer)
	// }
	data.Netbios = &NetbiosDataSourceModel{
		Enabled:     types.BoolValue(restInfo.Netbios.Enabled),
		Aliases:     connection.FlattenTypesStringList(restInfo.Netbios.Aliases),
		WinsServers: connection.FlattenTypesStringList(restInfo.Netbios.WinsServers),
	}

	data.Security = &CifsSecurityDataSourceModel{
		RestrictAnonymous:        types.StringValue(restInfo.Security.RestrictAnonymous),
		SmbSigning:               types.BoolValue(restInfo.Security.SmbSigning),
		SmbEncryption:            types.BoolValue(restInfo.Security.SmbEncryption),
		KdcEncryption:            types.BoolValue(restInfo.Security.KdcEncryption),
		LmCompatibilityLevel:     types.StringValue(restInfo.Security.LmCompatibilityLevel),
		AesNetlogonEnabled:       types.BoolValue(restInfo.Security.AesNetlogonEnabled),
		TryLdapChannelBinding:    types.BoolValue(restInfo.Security.TryLdapChannelBinding),
		LdapReferralEnabled:      types.BoolValue(restInfo.Security.LdapReferralEnabled),
		EncryptDcConnection:      types.BoolValue(restInfo.Security.EncryptDcConnection),
		UseStartTLS:              types.BoolValue(restInfo.Security.UseStartTLS),
		SessionSecurity:          types.StringValue(restInfo.Security.SessionSecurity),
		UseLdaps:                 types.BoolValue(restInfo.Security.UseLdaps),
		AdvertisedKdcEncryptions: connection.FlattenTypesStringList(restInfo.Security.AdvertisedKdcEncryptions),
	}
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
