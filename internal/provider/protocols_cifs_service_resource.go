package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &CifsServiceResource{}
var _ resource.ResourceWithImportState = &CifsServiceResource{}

// NewCifsServiceResource is a helper function to simplify the provider implementation.
func NewCifsServiceResource() resource.Resource {
	return &CifsServiceResource{
		config: resourceOrDataSourceConfig{
			name: "protocols_cifs_service_resource",
		},
	}
}

// CifsServiceResource defines the resource implementation.
type CifsServiceResource struct {
	config resourceOrDataSourceConfig
}

// CifsServiceResourceModel describes the resource data model.
type CifsServiceResourceModel struct {
	CxProfileName   types.String               `tfsdk:"cx_profile_name"`
	Name            types.String               `tfsdk:"name"`
	SVMName         types.String               `tfsdk:"svm_name"`
	AdDomain        *AdDomainResourceModel     `tfsdk:"ad_domain"`
	Netbios         *CifsNetbiosResourceModel  `tfsdk:"netbios"`
	Security        *CifsSecurityResourceModel `tfsdk:"security"`
	Comment         types.String               `tfsdk:"comment"`
	DefaultUnixUser types.String               `tfsdk:"default_unix_user"`
	Enabled         types.Bool                 `tfsdk:"enabled"`
	Force           types.Bool                 `tfsdk:"force"`
	ID              types.String               `tfsdk:"id"`
}

// AdDomainResourceModel describes the ad_domain data model using go types for mapping.
type AdDomainResourceModel struct {
	OrganizationalUnit types.String `tfsdk:"organizational_unit"`
	User               types.String `tfsdk:"user"`
	Password           types.String `tfsdk:"password"`
	Fqdn               types.String `tfsdk:"fqdn"`
}

// CifsNetbiosResourceModel describes the netbios resource model using go types for mapping.
type CifsNetbiosResourceModel struct {
	Enabled     types.Bool `tfsdk:"enabled"`
	Aliases     types.Set  `tfsdk:"aliases"`
	WinsServers types.Set  `tfsdk:"wins_servers"`
}

// CifsSecurityResourceModel is the model for CIFS security.
type CifsSecurityResourceModel struct {
	RestrictAnonymous        types.String `tfsdk:"restrict_anonymous"`
	SmbSigning               types.Bool   `tfsdk:"smb_signing"`
	SmbEncryption            types.Bool   `tfsdk:"smb_encryption"`
	KdcEncryption            types.Bool   `tfsdk:"kdc_encryption"`
	LmCompatibilityLevel     types.String `tfsdk:"lm_compatibility_level"`
	AesNetlogonEnabled       types.Bool   `tfsdk:"aes_netlogon_enabled"`
	TryLdapChannelBinding    types.Bool   `tfsdk:"try_ldap_channel_binding"`
	LdapReferralEnabled      types.Bool   `tfsdk:"ldap_referral_enabled"`
	EncryptDcConnection      types.Bool   `tfsdk:"encrypt_dc_connection"`
	UseStartTLS              types.Bool   `tfsdk:"use_start_tls"`
	SessionSecurity          types.String `tfsdk:"session_security"`
	UseLdaps                 types.Bool   `tfsdk:"use_ldaps"`
	AdvertisedKdcEncryptions types.Set    `tfsdk:"advertised_kdc_encryptions"`
}

// Metadata returns the resource type name.
func (r *CifsServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *CifsServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsService resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "CifsService name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "CifsService svm name",
				Required:            true,
			},
			"ad_domain": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "Ad domain",
				Attributes: map[string]schema.Attribute{
					"organizational_unit": schema.StringAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Organizational unit",
					},
					"user": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: "User account with the access to add the CIFS server to the Active Directory",
					},
					"password": schema.StringAttribute{
						Required:            true,
						Sensitive:           true,
						MarkdownDescription: "Account password used to add this CIFS server to the Active Directory",
					},
					"fqdn": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: " Fully qualified domain name of the Windows Active Directory to which this CIFS server belongs",
					},
				},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies if the CIFS service is administratively enabled",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"default_unix_user": schema.StringAttribute{
				MarkdownDescription: "Default unix user",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Text comment of up to 48 characters about the CIFS server",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"force": schema.BoolAttribute{
				MarkdownDescription: "Specifies if the CIFS service is administratively enabled (9.11)",
				Computed:            true,
				Optional:            true,
				// If this is set and a machine account with the same name as specified in 'cifs-server name' exists
				// in the Active Directory, existing machine account will be overwritten and reused
				// The default value for this field is false.
				Default: booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"netbios": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Netbios",
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"enabled": types.BoolType,
					},
					map[string]attr.Value{
						"enabled": types.BoolValue(false),
					},
				)),
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						Default:  booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "NetBios name service (NBNS) is enabled for the CIFS",
					},
					"aliases": schema.SetAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "list of one or more NetBIOS aliases for the CIFS server",
						ElementType:         types.StringType,
					},
					"wins_servers": schema.SetAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "list of Windows Internet Name Server (WINS) addresses that manage and map the NetBIOS name of the CIFS server to their network IP addresses. The IP addresses must be IPv4 addresses.",
						ElementType:         types.StringType,
					},
				},
			},
			"security": schema.SingleNestedAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Security",
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"restrict_anonymous":       types.StringType,
						"smb_signing":              types.BoolType,
						"smb_encryption":           types.BoolType,
						"kdc_encryption":           types.BoolType,
						"lm_compatibility_level":   types.StringType,
						"try_ldap_channel_binding": types.BoolType,
						"ldap_referral_enabled":    types.BoolType,
						"encrypt_dc_connection":    types.BoolType,
						"session_security":         types.StringType,
					},
					map[string]attr.Value{
						"restrict_anonymous":       types.StringValue("no_enumeration"),
						"smb_signing":              types.BoolValue(false),
						"smb_encryption":           types.BoolValue(false),
						"kdc_encryption":           types.BoolValue(false),
						"lm_compatibility_level":   types.StringValue("lm_ntlm_ntlmv2_krb"),
						"try_ldap_channel_binding": types.BoolValue(true),
						"ldap_referral_enabled":    types.BoolValue(false),
						"encrypt_dc_connection":    types.BoolValue(false),
						"session_security":         types.StringValue("none"),
					},
				)),
				Attributes: map[string]schema.Attribute{
					"restrict_anonymous": schema.StringAttribute{
						Computed:            true,
						Optional:            true,
						Default:             stringdefault.StaticString("no_enumeration"),
						MarkdownDescription: "Specifies what level of access an anonymous user is granted",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("no_restriction", "no_enumeration", "no_access"),
						},
					},
					"smb_signing": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						Default:  booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Specifies if signing is required for incoming CIFS traffic",
					},
					"smb_encryption": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						Default:  booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Specifies if encryption is required for incoming CIFS traffic",
					},
					"kdc_encryption": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						Default:  booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.Bool{
							boolvalidator.ConflictsWith(path.Expressions{
								path.MatchRoot("advertised_kdc_encryptions"),
							}...),
						},
						MarkdownDescription: "Specifies whether AES-128 and AES-256 encryption is enabled for all Kerberos-based communication with the Active Directory KDC",
					},
					"lm_compatibility_level": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("lm_ntlm_ntlmv2_krb"),
						MarkdownDescription: "CIFS server minimum security level",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("lm_ntlm_ntlmv2_krb", "lm_ntlm_ntlmv2_krb", "lm_ntlm_ntlmv2_krb", "krb"),
						},
					},
					"aes_netlogon_enabled": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						Default:  booldefault.StaticBool(false),
						// PlanModifiers: []planmodifier.Bool{
						// 	boolplanmodifier.UseStateForUnknown(),
						// },
						MarkdownDescription: "An AES session key is enabled for the Netlogon channel (9.10)",
					},
					"try_ldap_channel_binding": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						Default:  booldefault.StaticBool(true),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Specifies whether or not channel binding is attempted in the case of TLS/LDAPS (9.10)",
					},
					"ldap_referral_enabled": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						Default:  booldefault.StaticBool(false),
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Specifies if LDAP referral chasing is enabled for AD LDAP connections (9.10)",
					},
					"encrypt_dc_connection": schema.BoolAttribute{
						Computed:            true,
						Optional:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Encryption is required for domain controller connections (9.8)",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"use_start_tls": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Specifies whether or not to use SSL/TLS for allowing secure LDAP communication with Active Directory LDAP servers (9.10)",
					},
					"session_security": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						Default:             stringdefault.StaticString("none"),
						MarkdownDescription: "Client session security for AD LDAP connections (9.10)",
						// PlanModifiers: []planmodifier.String{
						// 	stringplanmodifier.UseStateForUnknown(),
						// },
						Validators: []validator.String{
							stringvalidator.OneOf("none", "sign", "seal"),
						},
					},
					"use_ldaps": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Specifies whether or not to use use LDAPS for secure Active Directory LDAP connections by using the TLS/SSL protocols (9.10)",
					},
					"advertised_kdc_encryptions": schema.SetAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Set{
							setplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "List of advertised KDC encryptions",
						ElementType:         types.StringType,
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "CifsService ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *CifsServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
}

// stringSliceToSet converts a slice of GroupMember to a types.Set
func stringSliceToSet(ctx context.Context, stringsSliceIn []string, diags *diag.Diagnostics) types.Set {
	keys, d := types.SetValueFrom(ctx, types.StringType, stringsSliceIn)
	diags.Append(d...)

	return keys
}

// Read refreshes the Terraform state with the latest data.
func (r *CifsServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CifsServiceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	restInfo, err := interfaces.GetCifsServiceByName(errorHandler, *client, data.Name.ValueString())
	if err != nil {
		// error reporting done inside GetCifsService
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No CIFS service found", "CIFS service not found.")
		return
	}
	data.Name = types.StringValue(strings.ToLower(restInfo.Name))
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	if len(restInfo.Comment) != 0 {
		data.Comment = types.StringValue(restInfo.Comment)
	}
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.DefaultUnixUser = types.StringValue(restInfo.DefaultUnixUser)
	data.AdDomain = &AdDomainResourceModel{
		OrganizationalUnit: types.StringValue(restInfo.AdDomain.OrganizationalUnit),
		// use the same values as in the state for both user and password since they cannot be read by API
		User:     data.AdDomain.User,
		Password: data.AdDomain.Password,
		Fqdn:     types.StringValue(restInfo.AdDomain.Fqdn),
	}

	data.Netbios = &CifsNetbiosResourceModel{
		Enabled:     types.BoolValue(restInfo.Netbios.Enabled),
		Aliases:     stringSliceToSet(ctx, restInfo.Netbios.Aliases, &resp.Diagnostics),
		WinsServers: stringSliceToSet(ctx, restInfo.Netbios.WinsServers, &resp.Diagnostics),
	}

	data.Security = &CifsSecurityResourceModel{
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
		AdvertisedKdcEncryptions: stringSliceToSet(ctx, restInfo.Security.AdvertisedKdcEncryptions, &resp.Diagnostics),
	}

	// Set the ID
	data.ID = types.StringValue(fmt.Sprintf("%s_%s_%s", data.CxProfileName.ValueString(), data.SVMName.ValueString(), data.Name.ValueString()))

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *CifsServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *CifsServiceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.CifsServiceResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Create the resource
	body.AdDomain.Fqdn = data.AdDomain.Fqdn.ValueString()
	body.AdDomain.User = data.AdDomain.User.ValueString()
	body.AdDomain.Password = data.AdDomain.Password.ValueString()
	// optional fields
	if !data.AdDomain.OrganizationalUnit.IsNull() {
		body.AdDomain.OrganizationalUnit = data.AdDomain.OrganizationalUnit.ValueString()
	}

	// default value
	body.Enabled = data.Enabled.ValueBool()
	tflog.Debug(ctx, fmt.Sprintf("\n\n***Create protocols_cifs_service source - body enabled: %#v", body.Enabled))

	if !data.DefaultUnixUser.IsNull() {
		body.DefaultUnixUser = data.DefaultUnixUser.ValueString()
	}
	if !data.Comment.IsNull() {
		body.Comment = data.Comment.ValueString()
	}

	if data.Netbios != nil {
		tflog.Debug(ctx, "\n===netbios root is not nil")
		if !data.Netbios.Enabled.IsNull() {
			tflog.Debug(ctx, "\t====netbios enable has value")
			body.Netbios.Enabled = data.Netbios.Enabled.ValueBool()
		}
		if !data.Netbios.Aliases.IsNull() {
			tflog.Debug(ctx, "\t====netbios aliases has value")
			//tflog.Debug(ctx, fmt.Sprintf("##netbios aliases - body: %#v", data.Netbios.Aliases))
			aliases := data.Netbios.Aliases.Elements()
			body.Netbios.Aliases = make([]string, len(aliases))
			for i, e := range aliases {
				body.Netbios.Aliases[i] = e.String()
			}
		}
		if !data.Netbios.WinsServers.IsNull() {
			tflog.Debug(ctx, "\t====netbios winservers has value")
			//tflog.Debug(ctx, fmt.Sprintf("##netbios aliases - body: %#v", data.Netbios.WinsServers))
			windowServers := data.Netbios.WinsServers.Elements()
			body.Netbios.WinsServers = make([]string, len(windowServers))
			for i, e := range windowServers {
				body.Netbios.WinsServers[i] = e.String()
			}
		}
	}

	if data.Security != nil {
		tflog.Debug(ctx, "\n===security root is not nil")
		body.Security.RestrictAnonymous = data.Security.RestrictAnonymous.ValueString()
		body.Security.SmbSigning = data.Security.SmbSigning.ValueBool()
		body.Security.SmbEncryption = data.Security.SmbEncryption.ValueBool()
		body.Security.KdcEncryption = data.Security.KdcEncryption.ValueBool()
		body.Security.LmCompatibilityLevel = data.Security.LmCompatibilityLevel.ValueString()
		body.Security.AesNetlogonEnabled = data.Security.AesNetlogonEnabled.ValueBool()
		body.Security.TryLdapChannelBinding = data.Security.TryLdapChannelBinding.ValueBool()
		body.Security.LdapReferralEnabled = data.Security.LdapReferralEnabled.ValueBool()
		body.Security.EncryptDcConnection = data.Security.EncryptDcConnection.ValueBool()
		body.Security.UseStartTLS = data.Security.UseStartTLS.ValueBool()
		body.Security.SessionSecurity = data.Security.SessionSecurity.ValueString()
		if !data.Security.UseLdaps.IsNull() {
			tflog.Debug(ctx, "\t====security useldap is not nil")
			body.Security.UseLdaps = data.Security.UseLdaps.ValueBool()
		}
		if !data.Security.AdvertisedKdcEncryptions.IsNull() {
			tflog.Debug(ctx, "\t====security advertise kd encryptions is not nil")
			advertisedKdcEncryptions := data.Security.AdvertisedKdcEncryptions.Elements()
			body.Security.AdvertisedKdcEncryptions = make([]string, len(advertisedKdcEncryptions))
			for i, e := range advertisedKdcEncryptions {
				body.Security.AdvertisedKdcEncryptions[i] = e.String()
			}
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("##Create protocols_cifs_service source - body: %#v", body))
	_, err = interfaces.CreateCifsService(errorHandler, *client, data.Force.ValueBool(), body)
	if err != nil {
		return
	}

	// Set the ID
	data.ID = types.StringValue(fmt.Sprintf("%s_%s_%s", data.CxProfileName.ValueString(), data.SVMName.ValueString(), data.Name.ValueString()))

	restInfo, err := interfaces.GetCifsServiceByName(errorHandler, *client, data.Name.ValueString())
	if err != nil {
		// error reporting done inside GetCifsService
		return
	}

	//data.Name = types.StringValue(strings.ToLower(restInfo.Name))
	//data.SVMName = types.StringValue(restInfo.SVM.Name)
	// if len(restInfo.Comment) != 0 {
	// 	data.Comment = types.StringValue(restInfo.Comment)
	// }
	data.Enabled = types.BoolValue(restInfo.Enabled)
	tflog.Debug(ctx, fmt.Sprintf("**Create protocols_cifs_service source - after query enable: %#v", data))
	data.DefaultUnixUser = types.StringValue(restInfo.DefaultUnixUser)
	data.AdDomain = &AdDomainResourceModel{
		OrganizationalUnit: types.StringValue(restInfo.AdDomain.OrganizationalUnit),
		// use the same values as in the state for both user and password since they cannot be read by API
		User:     data.AdDomain.User,
		Password: data.AdDomain.Password,
		Fqdn:     types.StringValue(restInfo.AdDomain.Fqdn),
	}

	data.Netbios = &CifsNetbiosResourceModel{
		Enabled:     types.BoolValue(restInfo.Netbios.Enabled),
		Aliases:     stringSliceToSet(ctx, restInfo.Netbios.Aliases, &resp.Diagnostics),
		WinsServers: stringSliceToSet(ctx, restInfo.Netbios.WinsServers, &resp.Diagnostics),
	}

	data.Security = &CifsSecurityResourceModel{
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
		AdvertisedKdcEncryptions: stringSliceToSet(ctx, restInfo.Security.AdvertisedKdcEncryptions, &resp.Diagnostics),
	}
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Check if two slices of strings are equal
func stringSlicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *CifsServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *CifsServiceResourceModel
	var dataOld *CifsServiceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &dataOld)...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}

	var body interfaces.CifsServiceResourceBodyDataModelONTAP
	// check if the name is changed
	if !data.Name.Equal(dataOld.Name) {
		// rename a server should be in stop state
		body.Name = data.Name.ValueString()
	}
	body.Enabled = data.Enabled.ValueBool()

	if !data.AdDomain.Fqdn.Equal(dataOld.AdDomain.Fqdn) {
		body.AdDomain.Fqdn = data.AdDomain.Fqdn.ValueString()
	}
	if !data.AdDomain.User.Equal(dataOld.AdDomain.User) {
		body.AdDomain.User = data.AdDomain.User.ValueString()
	}
	if !data.AdDomain.Password.Equal(dataOld.AdDomain.Password) {
		body.AdDomain.Password = data.AdDomain.Password.ValueString()
	}

	if !data.AdDomain.OrganizationalUnit.Equal(dataOld.AdDomain.OrganizationalUnit) {
		body.AdDomain.OrganizationalUnit = data.AdDomain.OrganizationalUnit.ValueString()
	}
	if !data.Comment.Equal(dataOld.Comment) {
		body.Comment = data.Comment.ValueString()
	}

	if !data.DefaultUnixUser.Equal(dataOld.DefaultUnixUser) {
		body.DefaultUnixUser = data.DefaultUnixUser.ValueString()
	}
	// var aliases, winservers []string

	// if !data.Netbios.Enabled.Equal(dataOld.Netbios.Enabled) {
	// 	body.Netbios.Enabled = data.Netbios.Enabled.ValueBool()
	// }
	// for _, e := range data.Netbios.Aliases {
	// 	aliases = append(aliases, e.ValueString())
	// }
	// body.Netbios.Aliases = aliases
	// for _, e := range data.Netbios.WinsServers {
	// 	winservers = append(winservers, e.ValueString())
	// }
	// body.Netbios.WinsServers = winservers

	// body.Netbios.Enabled = data.Netbios.Enabled.ValueBool()

	// if data.Security != nil {
	// 	if !data.Security.KdcEncryption.Equal(dataOld.Security.KdcEncryption) {
	// 		body.Security.KdcEncryption = data.Security.KdcEncryption.ValueBool()
	// 	}

	// 	if !data.Security.RestrictAnonymous.Equal(dataOld.Security.RestrictAnonymous) {
	// 		body.Security.RestrictAnonymous = data.Security.RestrictAnonymous.ValueString()
	// 	}
	// 	if !data.Security.SmbSigning.Equal(dataOld.Security.SmbSigning) {
	// 		body.Security.SmbSigning = data.Security.SmbSigning.ValueBool()
	// 	}
	// 	if !data.Security.SmbEncryption.Equal(dataOld.Security.SmbEncryption) {
	// 		body.Security.SmbEncryption = data.Security.SmbEncryption.ValueBool()
	// 	}
	// 	if !data.Security.LmCompatibilityLevel.Equal(dataOld.Security.LmCompatibilityLevel) {
	// 		body.Security.LmCompatibilityLevel = data.Security.LmCompatibilityLevel.ValueString()
	// 	}
	// 	if !data.Security.AesNetlogonEnabled.Equal(dataOld.Security.AesNetlogonEnabled) {
	// 		body.Security.AesNetlogonEnabled = data.Security.AesNetlogonEnabled.ValueBool()
	// 	}
	// 	if !data.Security.TryLdapChannelBinding.Equal(dataOld.Security.TryLdapChannelBinding) {
	// 		body.Security.TryLdapChannelBinding = data.Security.TryLdapChannelBinding.ValueBool()
	// 	}
	// 	if !data.Security.LdapReferralEnabled.Equal(dataOld.Security.LdapReferralEnabled) {
	// 		body.Security.LdapReferralEnabled = data.Security.LdapReferralEnabled.ValueBool()
	// 	}
	// 	if !data.Security.EncryptDcConnection.Equal(dataOld.Security.EncryptDcConnection) {
	// 		body.Security.EncryptDcConnection = data.Security.EncryptDcConnection.ValueBool()
	// 	}
	// 	if !data.Security.UseStartTLS.Equal(dataOld.Security.UseStartTLS) {
	// 		body.Security.UseStartTLS = data.Security.UseStartTLS.ValueBool()
	// 	}
	// 	if !data.Security.SessionSecurity.Equal(dataOld.Security.SessionSecurity) {
	// 		body.Security.SessionSecurity = data.Security.SessionSecurity.ValueString()
	// 	}
	// 	if !data.Security.UseLdaps.Equal(dataOld.Security.UseLdaps) {
	// 		body.Security.UseLdaps = data.Security.UseLdaps.ValueBool()
	// 	}
	// }

	err = interfaces.UpdateCifsService(errorHandler, *client, svm.UUID, data.Force.ValueBool(), body)
	if err != nil {
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *CifsServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *CifsServiceResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("ID is null", "protocols_cifs_service ID is null")
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}

	var body interfaces.CifsServiceResourceDeleteBodyDataModelONTAP

	body.AdDomain.User = data.AdDomain.User.ValueString()
	body.AdDomain.Password = data.AdDomain.Password.ValueString()

	err = interfaces.DeleteCifsService(errorHandler, *client, svm.UUID, data.Force.ValueBool(), body)
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *CifsServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req a protocols cifs service resource: %#v", req))
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
}
