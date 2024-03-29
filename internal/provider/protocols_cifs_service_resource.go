package provider

import (
	"context"
	"fmt"
	"strconv"
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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
	CxProfileName   types.String           `tfsdk:"cx_profile_name"`
	Name            types.String           `tfsdk:"name"`
	SVMName         types.String           `tfsdk:"svm_name"`
	AdDomain        *AdDomainResourceModel `tfsdk:"ad_domain"`
	Netbios         types.Object           `tfsdk:"netbios"`
	Security        types.Object           `tfsdk:"security"`
	Comment         types.String           `tfsdk:"comment"`
	DefaultUnixUser types.String           `tfsdk:"default_unix_user"`
	Enabled         types.Bool             `tfsdk:"enabled"`
	Force           types.Bool             `tfsdk:"force"`
	ID              types.String           `tfsdk:"id"`
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
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
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
				Attributes: map[string]schema.Attribute{
					"restrict_anonymous": schema.StringAttribute{
						Computed:            true,
						Optional:            true,
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
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Specifies if signing is required for incoming CIFS traffic",
					},
					"smb_encryption": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Specifies if encryption is required for incoming CIFS traffic",
					},
					"kdc_encryption": schema.BoolAttribute{
						Computed: true,
						Optional: true,
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
						MarkdownDescription: "CIFS server minimum security level",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("lm_ntlm_ntlmv2_krb", "ntlm_ntlmv2_krb", "ntlmv2_krb", "krb"),
						},
					},
					"aes_netlogon_enabled": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "An AES session key is enabled for the Netlogon channel (9.10)",
					},
					"try_ldap_channel_binding": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Specifies whether or not channel binding is attempted in the case of TLS/LDAPS (9.10)",
					},
					"ldap_referral_enabled": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Specifies if LDAP referral chasing is enabled for AD LDAP connections (9.10)",
					},
					"encrypt_dc_connection": schema.BoolAttribute{
						Computed:            true,
						Optional:            true,
						MarkdownDescription: "Encryption is required for domain controller connections (9.8)",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"use_start_tls": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Specifies whether or not to use SSL/TLS for allowing secure LDAP communication with Active Directory LDAP servers (9.10)",
					},
					"session_security": schema.StringAttribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "Client session security for AD LDAP connections (9.10)",
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
						Validators: []validator.String{
							stringvalidator.OneOf("none", "sign", "seal"),
						},
					},
					"use_ldaps": schema.BoolAttribute{
						Optional: true,
						Computed: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
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

// stringSliceToSet converts a slice of strings to a types.Set
func stringSliceToSet(ctx context.Context, stringsSliceIn []string, diags *diag.Diagnostics, toLower bool) types.Set {
	words := stringsSliceIn
	if toLower {
		for i, word := range stringsSliceIn {
			words[i] = strings.ToLower(word)
		}
	}
	keys, d := types.SetValueFrom(ctx, types.StringType, words)
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
		errorHandler.MakeAndReportError("No CIFS service found", fmt.Sprintf("CIFS service %s not found.", data.Name.ValueString()))
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
		User:     types.StringValue(data.AdDomain.User.ValueString()),
		Password: types.StringValue(data.AdDomain.Password.ValueString()),
		Fqdn:     types.StringValue(strings.ToLower(restInfo.AdDomain.Fqdn)),
	}

	elementTypes := map[string]attr.Type{
		"enabled":      types.BoolType,
		"aliases":      types.SetType{ElemType: types.StringType},
		"wins_servers": types.SetType{ElemType: types.StringType},
	}
	elements := map[string]attr.Value{
		"enabled":      types.BoolValue(restInfo.Netbios.Enabled),
		"aliases":      stringSliceToSet(ctx, restInfo.Netbios.Aliases, &resp.Diagnostics, true),
		"wins_servers": stringSliceToSet(ctx, restInfo.Netbios.WinsServers, &resp.Diagnostics, false),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Netbios = objectValue

	elementTypes = map[string]attr.Type{
		"restrict_anonymous":         types.StringType,
		"smb_signing":                types.BoolType,
		"smb_encryption":             types.BoolType,
		"kdc_encryption":             types.BoolType,
		"lm_compatibility_level":     types.StringType,
		"try_ldap_channel_binding":   types.BoolType,
		"ldap_referral_enabled":      types.BoolType,
		"encrypt_dc_connection":      types.BoolType,
		"session_security":           types.StringType,
		"aes_netlogon_enabled":       types.BoolType,
		"use_ldaps":                  types.BoolType,
		"use_start_tls":              types.BoolType,
		"advertised_kdc_encryptions": types.SetType{ElemType: types.StringType},
	}

	elements = map[string]attr.Value{
		"restrict_anonymous":         types.StringValue(restInfo.Security.RestrictAnonymous),
		"smb_signing":                types.BoolValue(restInfo.Security.SmbSigning),
		"smb_encryption":             types.BoolValue(restInfo.Security.SmbEncryption),
		"kdc_encryption":             types.BoolValue(restInfo.Security.KdcEncryption),
		"lm_compatibility_level":     types.StringValue(restInfo.Security.LmCompatibilityLevel),
		"try_ldap_channel_binding":   types.BoolValue(restInfo.Security.TryLdapChannelBinding),
		"ldap_referral_enabled":      types.BoolValue(restInfo.Security.LdapReferralEnabled),
		"encrypt_dc_connection":      types.BoolValue(restInfo.Security.EncryptDcConnection),
		"session_security":           types.StringValue(restInfo.Security.SessionSecurity),
		"aes_netlogon_enabled":       types.BoolValue(restInfo.Security.AesNetlogonEnabled),
		"use_ldaps":                  types.BoolValue(restInfo.Security.UseLdaps),
		"use_start_tls":              types.BoolValue(restInfo.Security.UseStartTLS),
		"advertised_kdc_encryptions": stringSliceToSet(ctx, restInfo.Security.AdvertisedKdcEncryptions, &resp.Diagnostics, false),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Security = objectValue

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
	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", fmt.Sprintf("Cluster %s not found.", data.CxProfileName.ValueString()))
		return
	}
	clusterVersion := strconv.Itoa(cluster.Version.Generation) + "." + strconv.Itoa(cluster.Version.Major)
	var errors []string
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

	if !data.DefaultUnixUser.IsNull() {
		body.DefaultUnixUser = data.DefaultUnixUser.ValueString()
	}
	if !data.Comment.IsNull() {
		body.Comment = data.Comment.ValueString()
	}

	if !data.Netbios.IsUnknown() {
		var netbios CifsNetbiosResourceModel
		diags := data.Netbios.As(ctx, &netbios, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.Netbios.Enabled = netbios.Enabled.ValueBool()
		if !netbios.Aliases.IsNull() {
			for _, e := range netbios.Aliases.Elements() {
				body.Netbios.Aliases = append(body.Netbios.Aliases, e.(basetypes.StringValue).ValueString())
			}
		}
		if !netbios.WinsServers.IsNull() {
			windowServers := netbios.WinsServers.Elements()
			body.Netbios.WinsServers = make([]string, len(windowServers))
			for i, e := range windowServers {
				body.Netbios.WinsServers[i] = e.String()
			}
		}
	}

	if !data.Security.IsUnknown() {
		var security CifsSecurityResourceModel
		diags := data.Security.As(ctx, &security, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.Security.RestrictAnonymous = security.RestrictAnonymous.ValueString()
		body.Security.SmbSigning = security.SmbSigning.ValueBool()
		body.Security.SmbEncryption = security.SmbEncryption.ValueBool()
		// kdc_encryption is only supported in 9.12 and earlier
		if !security.KdcEncryption.IsNull() {
			if CompareVersions(clusterVersion, "9.12") <= 0 {
				body.Security.KdcEncryption = security.KdcEncryption.ValueBool()
			} else {
				errors = append(errors, "kdc_encryption")
			}
		}
		if !security.LmCompatibilityLevel.IsNull() {
			if CompareVersions(clusterVersion, "9.8") >= 0 {
				body.Security.LmCompatibilityLevel = security.LmCompatibilityLevel.ValueString()
			} else {
				errors = append(errors, "lm_compatibility_level")
			}
		}
		if !security.AesNetlogonEnabled.IsNull() {
			if CompareVersions(clusterVersion, "9.10") >= 0 {
				body.Security.AesNetlogonEnabled = security.AesNetlogonEnabled.ValueBool()
			} else {
				errors = append(errors, "aes_netlogon_enabled")
			}
		}
		if !security.TryLdapChannelBinding.IsNull() {
			if CompareVersions(clusterVersion, "9.10") >= 0 {
				body.Security.TryLdapChannelBinding = security.TryLdapChannelBinding.ValueBool()
			} else {
				errors = append(errors, "try_ldap_channel_binding")
			}
		}
		if !security.LdapReferralEnabled.IsNull() {
			if CompareVersions(clusterVersion, "9.10") >= 0 {
				body.Security.LdapReferralEnabled = security.LdapReferralEnabled.ValueBool()
			} else {
				errors = append(errors, "ldap_referral_enabled")
			}
		}
		if !security.EncryptDcConnection.IsNull() {
			if CompareVersions(clusterVersion, "9.8") >= 0 {
				body.Security.EncryptDcConnection = security.EncryptDcConnection.ValueBool()
			} else {
				errors = append(errors, "encrypt_dc_connection")
			}
		}
		if !security.UseStartTLS.IsNull() {
			if CompareVersions(clusterVersion, "9.10") >= 0 {
				body.Security.UseStartTLS = security.UseStartTLS.ValueBool()
			} else {
				errors = append(errors, "use_start_tls")
			}
		}
		if !security.SessionSecurity.IsNull() {
			if CompareVersions(clusterVersion, "9.10") >= 0 {
				body.Security.SessionSecurity = security.SessionSecurity.ValueString()
			} else {
				errors = append(errors, "session_security")
			}
		}
		if !security.UseLdaps.IsNull() {
			if CompareVersions(clusterVersion, "9.10") >= 0 {
				body.Security.UseLdaps = security.UseLdaps.ValueBool()
			} else {
				errors = append(errors, "use_ldaps")
			}
		}
		if !security.AdvertisedKdcEncryptions.IsNull() {
			if CompareVersions(clusterVersion, "9.12") >= 0 {
				advertisedKdcEncryptions := security.AdvertisedKdcEncryptions.Elements()
				body.Security.AdvertisedKdcEncryptions = make([]string, len(advertisedKdcEncryptions))
				for i, e := range advertisedKdcEncryptions {
					body.Security.AdvertisedKdcEncryptions[i] = e.String()
				}
			} else {
				errors = append(errors, "advertised_kdc_encryptions")
			}
		}
	}
	if len(errors) > 0 {
		errorsString := strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("The following Variables are not supported with current version: %#v", errorsString))
		return
	}
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
	if restInfo == nil {
		errorHandler.MakeAndReportError("No CIFS service found", "CIFS service not found.")
		return
	}
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.DefaultUnixUser = types.StringValue(restInfo.DefaultUnixUser)
	var fqdn types.String
	if strings.EqualFold(data.AdDomain.Fqdn.ValueString(), restInfo.AdDomain.Fqdn) {
		fqdn = types.StringValue(data.AdDomain.Fqdn.ValueString())
	}
	data.AdDomain = &AdDomainResourceModel{
		OrganizationalUnit: types.StringValue(restInfo.AdDomain.OrganizationalUnit),
		// use the same values as in the state for both user and password since they cannot be read by API
		User:     data.AdDomain.User,
		Password: data.AdDomain.Password,
		Fqdn:     fqdn,
	}

	elementTypes := map[string]attr.Type{
		"enabled":      types.BoolType,
		"aliases":      types.SetType{ElemType: types.StringType},
		"wins_servers": types.SetType{ElemType: types.StringType},
	}
	elements := map[string]attr.Value{
		"enabled":      types.BoolValue(restInfo.Netbios.Enabled),
		"aliases":      stringSliceToSet(ctx, restInfo.Netbios.Aliases, &resp.Diagnostics, true),
		"wins_servers": stringSliceToSet(ctx, restInfo.Netbios.WinsServers, &resp.Diagnostics, false),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Netbios = objectValue

	elementTypes = map[string]attr.Type{
		"restrict_anonymous":         types.StringType,
		"smb_signing":                types.BoolType,
		"smb_encryption":             types.BoolType,
		"kdc_encryption":             types.BoolType,
		"lm_compatibility_level":     types.StringType,
		"try_ldap_channel_binding":   types.BoolType,
		"ldap_referral_enabled":      types.BoolType,
		"encrypt_dc_connection":      types.BoolType,
		"session_security":           types.StringType,
		"aes_netlogon_enabled":       types.BoolType,
		"use_ldaps":                  types.BoolType,
		"use_start_tls":              types.BoolType,
		"advertised_kdc_encryptions": types.SetType{ElemType: types.StringType},
	}

	elements = map[string]attr.Value{
		"restrict_anonymous":         types.StringValue(restInfo.Security.RestrictAnonymous),
		"smb_signing":                types.BoolValue(restInfo.Security.SmbSigning),
		"smb_encryption":             types.BoolValue(restInfo.Security.SmbEncryption),
		"kdc_encryption":             types.BoolValue(restInfo.Security.KdcEncryption),
		"lm_compatibility_level":     types.StringValue(restInfo.Security.LmCompatibilityLevel),
		"try_ldap_channel_binding":   types.BoolValue(restInfo.Security.TryLdapChannelBinding),
		"ldap_referral_enabled":      types.BoolValue(restInfo.Security.LdapReferralEnabled),
		"encrypt_dc_connection":      types.BoolValue(restInfo.Security.EncryptDcConnection),
		"session_security":           types.StringValue(restInfo.Security.SessionSecurity),
		"aes_netlogon_enabled":       types.BoolValue(restInfo.Security.AesNetlogonEnabled),
		"use_ldaps":                  types.BoolValue(restInfo.Security.UseLdaps),
		"use_start_tls":              types.BoolValue(restInfo.Security.UseStartTLS),
		"advertised_kdc_encryptions": stringSliceToSet(ctx, restInfo.Security.AdvertisedKdcEncryptions, &resp.Diagnostics, false),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Security = objectValue

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *CifsServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *CifsServiceResourceModel
	var state *CifsServiceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}
	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, plan.SVMName.ValueString())
	if err != nil {
		return
	}
	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", fmt.Sprintf("Cluster not found."))
		return
	}
	clusterVersion := strconv.Itoa(cluster.Version.Generation) + "." + strconv.Itoa(cluster.Version.Major)

	var body interfaces.CifsServiceResourceBodyDataModelONTAP
	// check if the name is changed
	if !plan.Name.Equal(state.Name) {
		// rename a server should be in stop state
		body.Name = plan.Name.ValueString()
	}
	body.Enabled = plan.Enabled.ValueBool()

	if !plan.AdDomain.Fqdn.Equal(state.AdDomain.Fqdn) {
		body.AdDomain.Fqdn = plan.AdDomain.Fqdn.ValueString()
	}
	if !plan.AdDomain.User.Equal(state.AdDomain.User) {
		body.AdDomain.User = plan.AdDomain.User.ValueString()
	}
	if !plan.AdDomain.Password.Equal(state.AdDomain.Password) {
		body.AdDomain.Password = plan.AdDomain.Password.ValueString()
	}

	if !plan.AdDomain.OrganizationalUnit.Equal(state.AdDomain.OrganizationalUnit) {
		body.AdDomain.OrganizationalUnit = plan.AdDomain.OrganizationalUnit.ValueString()
	}
	if !plan.Comment.Equal(state.Comment) {
		body.Comment = plan.Comment.ValueString()
	}

	if !plan.DefaultUnixUser.Equal(state.DefaultUnixUser) {
		body.DefaultUnixUser = plan.DefaultUnixUser.ValueString()
	}

	if !plan.Netbios.IsUnknown() {
		if !plan.Netbios.Equal(state.Netbios) {
			var netbios CifsNetbiosResourceModel
			diags := plan.Netbios.As(ctx, &netbios, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if !netbios.Enabled.IsUnknown() {
				body.Netbios.Enabled = netbios.Enabled.ValueBool()
			}
			if !netbios.Aliases.IsUnknown() {
				for _, e := range netbios.Aliases.Elements() {
					body.Netbios.Aliases = append(body.Netbios.Aliases, e.(basetypes.StringValue).ValueString())
				}
			}
			if !netbios.WinsServers.IsUnknown() {
				for _, e := range netbios.WinsServers.Elements() {
					body.Netbios.WinsServers = append(body.Netbios.WinsServers, e.(basetypes.StringValue).ValueString())
				}
			}
		}
	}

	if !plan.Security.IsUnknown() {
		if !plan.Security.Equal(state.Security) {
			var security CifsSecurityResourceModel
			diags := plan.Security.As(ctx, &security, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if !security.KdcEncryption.IsUnknown() && CompareVersions(clusterVersion, "9.12") < 0 {
				body.Security.KdcEncryption = security.KdcEncryption.ValueBool()
			}
			if !security.RestrictAnonymous.IsUnknown() {
				body.Security.RestrictAnonymous = security.RestrictAnonymous.ValueString()
			}
			if !security.SmbSigning.IsUnknown() {
				body.Security.SmbSigning = security.SmbSigning.ValueBool()
			}
			if !security.SmbEncryption.IsUnknown() {
				body.Security.SmbEncryption = security.SmbEncryption.ValueBool()
			}
			if !security.EncryptDcConnection.IsUnknown() && CompareVersions(clusterVersion, "9.8") >= 0 {
				body.Security.EncryptDcConnection = security.EncryptDcConnection.ValueBool()
			}
			if !security.LmCompatibilityLevel.IsUnknown() && CompareVersions(clusterVersion, "9.8") >= 0 {
				body.Security.LmCompatibilityLevel = security.LmCompatibilityLevel.ValueString()
			}
			if !security.AesNetlogonEnabled.IsUnknown() && CompareVersions(clusterVersion, "9.10") >= 0 {
				body.Security.AesNetlogonEnabled = security.AesNetlogonEnabled.ValueBool()
			}
			if !security.LdapReferralEnabled.IsUnknown() && CompareVersions(clusterVersion, "9.10") >= 0 {
				body.Security.LdapReferralEnabled = security.LdapReferralEnabled.ValueBool()
			}
			if !security.SessionSecurity.IsUnknown() && CompareVersions(clusterVersion, "9.10") >= 0 {
				body.Security.SessionSecurity = security.SessionSecurity.ValueString()
			}
			if !security.TryLdapChannelBinding.IsUnknown() && CompareVersions(clusterVersion, "9.10") >= 0 {
				body.Security.TryLdapChannelBinding = security.TryLdapChannelBinding.ValueBool()
			}
			if !security.UseLdaps.IsUnknown() && CompareVersions(clusterVersion, "9.10") >= 0 {
				body.Security.UseLdaps = security.UseLdaps.ValueBool()
			}
			if !security.UseStartTLS.IsUnknown() && CompareVersions(clusterVersion, "9.10") >= 0 {
				body.Security.UseStartTLS = security.UseStartTLS.ValueBool()
			}
			if !security.AdvertisedKdcEncryptions.IsUnknown() && CompareVersions(clusterVersion, "9.12") >= 0 {
				for _, e := range security.AdvertisedKdcEncryptions.Elements() {
					body.Security.AdvertisedKdcEncryptions = append(body.Security.AdvertisedKdcEncryptions, e.(basetypes.StringValue).ValueString())
				}
			}
		}
	}
	tflog.Debug(ctx, fmt.Sprintf("##Update protocols_cifs_service source - body: %#v", body))
	err = interfaces.UpdateCifsService(errorHandler, *client, svm.UUID, plan.Force.ValueBool(), body)
	if err != nil {
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
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
	if len(idParts) != 5 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" || idParts[4] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,svm_name,cx_profile_name,ad_domain.user,ad_domain.password. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ad_domain").AtName("user"), idParts[3])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("ad_domain").AtName("password"), idParts[4])...)
}
