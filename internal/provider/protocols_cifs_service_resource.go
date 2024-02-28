package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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
	CxProfileName   types.String                 `tfsdk:"cx_profile_name"`
	Name            types.String                 `tfsdk:"name"`
	SVMName         types.String                 `tfsdk:"svm_name"`
	Force           types.Bool                   `tfsdk:"force"`
	Enabled         types.Bool                   `tfsdk:"enabled"`
	DefaultUnixUser types.String                 `tfsdk:"default_unix_user"`
	Comment         types.String                 `tfsdk:"comment"`
	AdDomain        *AdDomainDataSourceModel     `tfsdk:"ad_domain"`
	Netbios         *NetbiosDataSourceModel      `tfsdk:"netbios"`
	Security        *CifsSecurityDataSourceModel `tfsdk:"security"`
	ID              types.String                 `tfsdk:"id"`
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
			"enabled": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Specifies if the CIFS service is administratively enabled",
			},
			"default_unix_user": schema.StringAttribute{
				Computed: true,
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Default unix user",
			},
			"comment": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "text comment of up to 48 characters about the CIFS server",
			},
			"force": schema.BoolAttribute{
				Computed: true,
				Optional: true,
				// If this is set and a machine account with the same name as specified in 'cifs-server name' exists
				// in the Active Directory, existing machine account will be overwritten and reused
				// The default value for this field is false.
				Default: booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				MarkdownDescription: "Specifies if the CIFS service is administratively enabled",
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
						MarkdownDescription: "Account password used to add this CIFS server to the Active Directory",
					},
					"fqdn": schema.StringAttribute{
						Required:            true,
						MarkdownDescription: " Fully qualified domain name of the Windows Active Directory to which this CIFS server belongs",
					},
				},
			},
			"netbios": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Netbios",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Computed:            true,
						Optional:            true,
						MarkdownDescription: "NetBios name service (NBNS) is enabled for the CIFS",
					},
					"aliases": schema.SetAttribute{
						Optional:            true,
						MarkdownDescription: "list of one or more NetBIOS aliases for the CIFS server",
						ElementType:         types.StringType,
					},
					"wins_servers": schema.SetAttribute{
						Optional:            true,
						MarkdownDescription: "list of Windows Internet Name Server (WINS) addresses that manage and map the NetBIOS name of the CIFS server to their network IP addresses. The IP addresses must be IPv4 addresses.",
						ElementType:         types.StringType,
					},
				},
			},
			"security": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Security",
				Attributes: map[string]schema.Attribute{
					"advertised_kdc_encryptions": schema.SetAttribute{
						Optional:            true,
						MarkdownDescription: "Specify the encryption type to use",
						ElementType:         types.StringType,
					},
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
					"lm_compatibility_level": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "CIFS server minimum security level",
						Validators: []validator.String{
							stringvalidator.OneOf("lm_ntlm_ntlmv2_krb", "lm_ntlm_ntlmv2_krb", "lm_ntlm_ntlmv2_krb", "krb"),
						},
					},
					"aes_netlogon_enabled": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "An AES session key is enabled for the Netlogon channel",
					},
					"try_ldap_channel_binding": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Specifies whether or not channel binding is attempted in the case of TLS/LDAPS",
					},
					"ldap_referral_enabled": schema.BoolAttribute{
						Computed: true,
						Optional: true,
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
						MarkdownDescription: "Specifies if LDAP referral chasing is enabled for AD LDAP connections",
					},
					"encrypt_dc_connection": schema.BoolAttribute{
						Computed:            true,
						Optional:            true,
						MarkdownDescription: "Encryption is required for domain controller connections",
						PlanModifiers: []planmodifier.Bool{
							boolplanmodifier.UseStateForUnknown(),
						},
					},
					"use_start_tls": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Specifies whether or not to use SSL/TLS for allowing secure LDAP communication with Active Directory LDAP servers",
					},
					"session_security": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Client session security for AD LDAP connections",
						Validators: []validator.String{
							stringvalidator.OneOf("none", "sign", "seal"),
						},
					},
					"use_ldaps": schema.BoolAttribute{
						Optional:            true,
						MarkdownDescription: "Specifies whether or not to use use LDAPS for secure Active Directory LDAP connections by using the TLS/SSL protocols",
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

	//data.Name = types.StringValue(restInfo.Name)
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

	aliases := make([]types.String, len(restInfo.Netbios.Aliases))
	for i, alias := range restInfo.Netbios.Aliases {
		aliases[i] = types.StringValue(alias)
	}
	winsServers := make([]types.String, len(restInfo.Netbios.WinsServers))
	for i, winsServer := range restInfo.Netbios.WinsServers {
		winsServers[i] = types.StringValue(winsServer)
	}
	data.Netbios = &NetbiosDataSourceModel{
		Enabled:     types.BoolValue(restInfo.Netbios.Enabled),
		Aliases:     aliases,
		WinsServers: winsServers,
	}
	advertisedKdcEncryptions := make([]types.String, len(restInfo.Security.AdvertisedKdcEncryptions))
	for i, encryption := range restInfo.Security.AdvertisedKdcEncryptions {
		advertisedKdcEncryptions[i] = types.StringValue(encryption)
	}
	data.Security = &CifsSecurityDataSourceModel{
		RestrictAnonymous:        types.StringValue(restInfo.Security.RestrictAnonymous),
		SmbSigning:               types.BoolValue(restInfo.Security.SmbSigning),
		SmbEncryption:            types.BoolValue(restInfo.Security.SmbEncryption),
		AdvertisedKdcEncryptions: advertisedKdcEncryptions,
		LmCompatibilityLevel:     types.StringValue(restInfo.Security.LmCompatibilityLevel),
		AesNetlogonEnabled:       types.BoolValue(restInfo.Security.AesNetlogonEnabled),
		TryLdapChannelBinding:    types.BoolValue(restInfo.Security.TryLdapChannelBinding),
		LdapReferralEnabled:      types.BoolValue(restInfo.Security.LdapReferralEnabled),
		EncryptDcConnection:      types.BoolValue(restInfo.Security.EncryptDcConnection),
		UseStartTLS:              types.BoolValue(restInfo.Security.UseStartTLS),
		SessionSecurity:          types.StringValue(restInfo.Security.SessionSecurity),
		UseLdaps:                 types.BoolValue(restInfo.Security.UseLdaps),
	}
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
	if !data.Comment.IsNull() {
		body.Comment = data.Comment.ValueString()
	}
	if !data.Enabled.IsNull() {
		body.Enabled = data.Enabled.ValueBool()
	}
	if !data.DefaultUnixUser.IsNull() {
		body.DefaultUnixUser = data.DefaultUnixUser.ValueString()
	}
	var aliases, winservers []string
	if data.Netbios != nil {
		if !data.Netbios.Enabled.IsNull() {
			body.Netbios.Enabled = data.Netbios.Enabled.ValueBool()
		}
		for _, e := range data.Netbios.Aliases {
			aliases = append(aliases, e.ValueString())
		}
		body.Netbios.Aliases = aliases
		for _, e := range data.Netbios.WinsServers {
			winservers = append(winservers, e.ValueString())
		}
		body.Netbios.WinsServers = winservers
	}

	if data.Security != nil {
		var adEncryptions []string
		for _, e := range data.Security.AdvertisedKdcEncryptions {
			adEncryptions = append(adEncryptions, e.ValueString())
		}
		body.Security.AdvertisedKdcEncryptions = adEncryptions
		if !data.Security.RestrictAnonymous.IsNull() {
			body.Security.RestrictAnonymous = data.Security.RestrictAnonymous.ValueString()
		}
		if !data.Security.SmbSigning.IsNull() {
			body.Security.SmbSigning = data.Security.SmbSigning.ValueBool()
		}
		if !data.Security.SmbEncryption.IsNull() {
			body.Security.SmbEncryption = data.Security.SmbEncryption.ValueBool()
		}
		if !data.Security.LmCompatibilityLevel.IsNull() {
			body.Security.LmCompatibilityLevel = data.Security.LmCompatibilityLevel.ValueString()
		}
		if !data.Security.AesNetlogonEnabled.IsNull() {
			body.Security.AesNetlogonEnabled = data.Security.AesNetlogonEnabled.ValueBool()
		}
		if !data.Security.TryLdapChannelBinding.IsNull() {
			body.Security.TryLdapChannelBinding = data.Security.TryLdapChannelBinding.ValueBool()
		}
		if !data.Security.LdapReferralEnabled.IsNull() {
			body.Security.LdapReferralEnabled = data.Security.LdapReferralEnabled.ValueBool()
		}
		if !data.Security.EncryptDcConnection.IsNull() {
			body.Security.EncryptDcConnection = data.Security.EncryptDcConnection.ValueBool()
		}
		if !data.Security.UseStartTLS.IsNull() {
			body.Security.UseStartTLS = data.Security.UseStartTLS.ValueBool()
		}
		if !data.Security.SessionSecurity.IsNull() {
			body.Security.SessionSecurity = data.Security.SessionSecurity.ValueString()
		}
		if !data.Security.UseLdaps.IsNull() {
			body.Security.UseLdaps = data.Security.UseLdaps.ValueBool()
		}
	}
	resource, err := interfaces.CreateCifsService(errorHandler, *client, data.Force.ValueBool(), body)
	if err != nil {
		return
	}

	// Set the ID
	data.ID = types.StringValue(fmt.Sprintf("%s_%s_%s", data.CxProfileName.ValueString(), data.SVMName.ValueString(), data.Name.ValueString()))

	// update fields have default values
	data.AdDomain.OrganizationalUnit = types.StringValue(resource.AdDomain.OrganizationalUnit)
	data.Enabled = types.BoolValue(resource.Enabled)
	data.Security.RestrictAnonymous = types.StringValue(resource.Security.RestrictAnonymous)
	data.Security.SmbSigning = types.BoolValue(resource.Security.SmbSigning)
	data.Security.SmbEncryption = types.BoolValue(resource.Security.SmbEncryption)
	data.Security.EncryptDcConnection = types.BoolValue(resource.Security.EncryptDcConnection)
	data.DefaultUnixUser = types.StringValue(resource.DefaultUnixUser)
	data.Netbios.Enabled = types.BoolValue(resource.Netbios.Enabled)
	data.Security.AesNetlogonEnabled = types.BoolValue(resource.Security.AesNetlogonEnabled)
	data.Security.TryLdapChannelBinding = types.BoolValue(resource.Security.TryLdapChannelBinding)
	data.Security.LdapReferralEnabled = types.BoolValue(resource.Security.LdapReferralEnabled)

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
	var aliases, winservers []string

	if !data.Netbios.Enabled.Equal(dataOld.Netbios.Enabled) {
		body.Netbios.Enabled = data.Netbios.Enabled.ValueBool()
	}
	for _, e := range data.Netbios.Aliases {
		aliases = append(aliases, e.ValueString())
	}
	body.Netbios.Aliases = aliases
	for _, e := range data.Netbios.WinsServers {
		winservers = append(winservers, e.ValueString())
	}
	body.Netbios.WinsServers = winservers

	body.Netbios.Enabled = data.Netbios.Enabled.ValueBool()

	if data.Security != nil {

		var adEncryptions, oldAdEncryptions []string
		for _, e := range data.Security.AdvertisedKdcEncryptions {
			adEncryptions = append(adEncryptions, e.ValueString())
		}
		for _, e := range dataOld.Security.AdvertisedKdcEncryptions {
			oldAdEncryptions = append(oldAdEncryptions, e.ValueString())
		}
		if !stringSlicesEqual(adEncryptions, oldAdEncryptions) {
			body.Security.AdvertisedKdcEncryptions = adEncryptions
		}

		if !data.Security.RestrictAnonymous.Equal(dataOld.Security.RestrictAnonymous) {
			body.Security.RestrictAnonymous = data.Security.RestrictAnonymous.ValueString()
		}
		if !data.Security.SmbSigning.Equal(dataOld.Security.SmbSigning) {
			body.Security.SmbSigning = data.Security.SmbSigning.ValueBool()
		}
		if !data.Security.SmbEncryption.Equal(dataOld.Security.SmbEncryption) {
			body.Security.SmbEncryption = data.Security.SmbEncryption.ValueBool()
		}
		if !data.Security.LmCompatibilityLevel.Equal(dataOld.Security.LmCompatibilityLevel) {
			body.Security.LmCompatibilityLevel = data.Security.LmCompatibilityLevel.ValueString()
		}
		if !data.Security.AesNetlogonEnabled.Equal(dataOld.Security.AesNetlogonEnabled) {
			body.Security.AesNetlogonEnabled = data.Security.AesNetlogonEnabled.ValueBool()
		}
		if !data.Security.TryLdapChannelBinding.Equal(dataOld.Security.TryLdapChannelBinding) {
			body.Security.TryLdapChannelBinding = data.Security.TryLdapChannelBinding.ValueBool()
		}
		if !data.Security.LdapReferralEnabled.Equal(dataOld.Security.LdapReferralEnabled) {
			body.Security.LdapReferralEnabled = data.Security.LdapReferralEnabled.ValueBool()
		}
		if !data.Security.EncryptDcConnection.Equal(dataOld.Security.EncryptDcConnection) {
			body.Security.EncryptDcConnection = data.Security.EncryptDcConnection.ValueBool()
		}
		if !data.Security.UseStartTLS.Equal(dataOld.Security.UseStartTLS) {
			body.Security.UseStartTLS = data.Security.UseStartTLS.ValueBool()
		}
		if !data.Security.SessionSecurity.Equal(dataOld.Security.SessionSecurity) {
			body.Security.SessionSecurity = data.Security.SessionSecurity.ValueString()
		}
		if !data.Security.UseLdaps.Equal(dataOld.Security.UseLdaps) {
			body.Security.UseLdaps = data.Security.UseLdaps.ValueBool()
		}
	}

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

	var body interfaces.AdDomainDataModel
	body.Fqdn = data.AdDomain.Fqdn.ValueString()
	body.User = data.AdDomain.User.ValueString()
	body.Password = data.AdDomain.Password.ValueString()
	// optional fields
	if !data.AdDomain.OrganizationalUnit.IsNull() {
		body.OrganizationalUnit = data.AdDomain.OrganizationalUnit.ValueString()
	}

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
