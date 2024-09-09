package name_services

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &NameServicesLDAPResource{}
var _ resource.ResourceWithImportState = &NameServicesLDAPResource{}

// NewNameServicesLDAPResource is a helper function to simplify the provider implementation.
func NewNameServicesLDAPResource() resource.Resource {
	return &NameServicesLDAPResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "name_services_ldap",
		},
	}
}

// NameServicesLDAPResource defines the resource implementation.
type NameServicesLDAPResource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesLDAPResourceModel describes the resource data model.
type NameServicesLDAPResourceModel struct {
	CxProfileName        types.String   `tfsdk:"cx_profile_name"`
	SVMName              types.String   `tfsdk:"svm_name"`
	Servers              []types.String `tfsdk:"servers"`
	Schema               types.String   `tfsdk:"schema"`
	AdDomain             types.String   `tfsdk:"ad_domain"`
	BaseDN               types.String   `tfsdk:"base_dn"`
	BaseScope            types.String   `tfsdk:"base_scope"`
	BindDN               types.String   `tfsdk:"bind_dn"`
	BindAsCIFSServer     types.Bool     `tfsdk:"bind_as_cifs_server"`
	PreferredADServers   []types.String `tfsdk:"preferred_ad_servers"`
	Port                 types.Int64    `tfsdk:"port"`
	QueryTimeout         types.Int64    `tfsdk:"query_timeout"`
	MinBindLevel         types.String   `tfsdk:"min_bind_level"`
	UseStartTLS          types.Bool     `tfsdk:"use_start_tls"`
	ReferralEnabled      types.Bool     `tfsdk:"referral_enabled"`
	SessionSecurity      types.String   `tfsdk:"session_security"`
	LDAPSEnabled         types.Bool     `tfsdk:"ldaps_enabled"`
	BindPassword         types.String   `tfsdk:"bind_password"`
	SkipConfigValidation types.Bool     `tfsdk:"skip_config_validation"`
	ID                   types.String   `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *NameServicesLDAPResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *NameServicesLDAPResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesLDAP resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "NameServicesLDAP svm name",
				Required:            true,
			},
			"servers": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of LDAP servers used for this client configuration",
				Optional:            true,
			},
			"schema": schema.StringAttribute{
				MarkdownDescription: "The name of the schema template used by the SVM",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"base_dn": schema.StringAttribute{
				MarkdownDescription: "Specifies the default base DN for all searches",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ldaps_enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not LDAPS is enabled (9.9)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"min_bind_level": schema.StringAttribute{
				MarkdownDescription: "The minimum bind authentication level",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"anonymous", "simple", "sasl"}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bind_password": schema.StringAttribute{
				MarkdownDescription: "Specifies the bind password for the LDAP servers",
				Optional:            true,
				Sensitive:           true,
			},
			"bind_dn": schema.StringAttribute{
				MarkdownDescription: "Specifies the user that binds to the LDAP servers",
				Optional:            true,
			},
			"preferred_ad_servers": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "This parameter specifies a list of LDAP servers preferred over discovered servers",
				Optional:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port used to connect to the LDAP Servers",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"session_security": schema.StringAttribute{
				MarkdownDescription: "Specifies the level of security to be used for LDAP communications",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"none", "sign", "seal"}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"use_start_tls": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not to use Start TLS over LDAP connections",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ad_domain": schema.StringAttribute{
				MarkdownDescription: "Specifies the name of the Active Directory domain used to discover LDAP servers for use by this client",
				Optional:            true,
				Validators: []validator.String{
					stringvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("servers"),
					}...),
				},
			},
			"base_scope": schema.StringAttribute{
				MarkdownDescription: "Specifies the default search scope for LDAP queries",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"base", "onelevel", "subtree"}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"referral_enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not LDAP referral is enabled (9.9)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"bind_as_cifs_server": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not CIFS server's credentials are used to bind to the LDAP server (9.9)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"query_timeout": schema.Int64Attribute{
				MarkdownDescription: "Specifies the timeout for LDAP queries in seconds (9.9)",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"skip_config_validation": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not to skip the validation of the LDAP configuration (9.9)",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "NameServicesLDAP ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *NameServicesLDAPResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *NameServicesLDAPResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NameServicesLDAPResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// import
	if data.ID.IsNull() {
		// Get SVM info
		svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
		if err != nil {
			// error reporting done inside GetSvmByName
			errorHandler.MakeAndReportError("invalid svm name", fmt.Sprintf("protocols_cifs_local_group_members svm_name %s is invalid", data.SVMName.ValueString()))
			return
		}
		// use SVM uuid as ID since each SVM can have one LDAP configuration
		data.ID = types.StringValue(svm.UUID)
	}
	restInfo, err := interfaces.GetNameServicesLDAPBySVMID(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		// error reporting done inside GetNameServicesLDAP
		return
	}
	if !data.AdDomain.IsNull() {
		data.AdDomain = types.StringValue(restInfo.AdDomain)
	}
	if restInfo.BindDN != "" {
		data.BindDN = types.StringValue(restInfo.BindDN)
	}
	if restInfo.Servers != nil {
		data.Servers = make([]types.String, len(restInfo.Servers))
		for index, server := range restInfo.Servers {
			data.Servers[index] = types.StringValue(server)
		}
	}
	if restInfo.PreferredADServers != nil {
		for _, adserver := range restInfo.PreferredADServers {
			data.PreferredADServers = append(data.PreferredADServers, types.StringValue(adserver))
		}
	}
	// update computed fields
	data.Schema = types.StringValue(restInfo.Schema)
	data.BaseDN = types.StringValue(restInfo.BaseDN)
	data.BaseScope = types.StringValue(restInfo.BaseScope)
	data.BindAsCIFSServer = types.BoolValue(restInfo.BindAsCIFSServer)
	data.Port = types.Int64Value(restInfo.Port)
	data.QueryTimeout = types.Int64Value(restInfo.QueryTimeout)
	data.MinBindLevel = types.StringValue(restInfo.MinBindLevel)
	data.UseStartTLS = types.BoolValue(restInfo.UseStartTLS)
	data.ReferralEnabled = types.BoolValue(restInfo.ReferralEnabled)
	data.SessionSecurity = types.StringValue(restInfo.SessionSecurity)
	data.LDAPSEnabled = types.BoolValue(restInfo.LDAPSEnabled)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *NameServicesLDAPResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *NameServicesLDAPResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.NameServicesLDAPResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
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
	var errors []string

	body.SVM.Name = data.SVMName.ValueString()
	if data.Servers != nil {
		for _, server := range data.Servers {
			body.Servers = append(body.Servers, server.ValueString())
		}
	}
	if !data.Schema.IsNull() {
		body.Schema = data.Schema.ValueString()
	}
	if !data.AdDomain.IsNull() {
		body.AdDomain = data.AdDomain.ValueString()
	}
	if !data.BaseDN.IsNull() {
		body.BaseDN = data.BaseDN.ValueString()
	}
	if !data.BaseScope.IsNull() {
		body.BaseScope = data.BaseScope.ValueString()
	}
	if !data.BindDN.IsNull() {
		body.BindDN = data.BindDN.ValueString()
	}
	if !data.BindAsCIFSServer.IsNull() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 9 {
			body.BindAsCIFSServer = data.BindAsCIFSServer.ValueBool()
		} else {
			errors = append(errors, "bind_as_cifs_server")
		}
	}
	if !data.BindPassword.IsNull() {
		body.BindPassword = data.BindPassword.ValueString()
	}
	if data.PreferredADServers != nil {
		for _, adserver := range data.PreferredADServers {
			body.PreferredADServers = append(body.PreferredADServers, adserver.ValueString())
		}
	}
	if !data.Port.IsNull() {
		body.Port = data.Port.ValueInt64()
	}
	if !data.QueryTimeout.IsNull() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 9 {
			body.QueryTimeout = data.QueryTimeout.ValueInt64()
		} else {
			errors = append(errors, "query_timeout")
		}
	}
	if !data.MinBindLevel.IsNull() {
		body.MinBindLevel = data.MinBindLevel.ValueString()
	}
	if !data.UseStartTLS.IsNull() {
		body.UseStartTLS = data.UseStartTLS.ValueBool()
	}
	if !data.ReferralEnabled.IsNull() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 9 {
			body.ReferralEnabled = data.ReferralEnabled.ValueBool()
		} else {
			errors = append(errors, "referral_enabled")
		}
	}
	if !data.SessionSecurity.IsNull() {
		body.SessionSecurity = data.SessionSecurity.ValueString()
	}
	if !data.LDAPSEnabled.IsNull() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 9 {
			body.LDAPSEnabled = data.LDAPSEnabled.ValueBool()
		} else {
			errors = append(errors, "ldaps_enabled")
		}
	}
	if !data.SkipConfigValidation.IsNull() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 9 {
			body.SkipConfigValidation = data.SkipConfigValidation.ValueBool()
		} else {
			errors = append(errors, "skip_config_validation")
		}
	}

	if len(errors) > 0 {
		errorsString := strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("The following Variables are not supported with current version: %#v", errorsString))
		return
	}

	resource, err := interfaces.CreateNameServicesLDAP(errorHandler, *client, body)
	if err != nil {
		return
	}

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, resource.SVM.Name)
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("invalid svm name", fmt.Sprintf("protocols_cifs_local_group_members svm_name %s is invalid", data.SVMName.ValueString()))
		return
	}

	// Read the Ldap configuration
	restInfo, err := interfaces.GetNameServicesLDAPBySVMID(errorHandler, *client, svm.UUID)
	if err != nil {
		return
	}

	// Save computed data into Terraform state
	data.MinBindLevel = types.StringValue(restInfo.MinBindLevel)
	data.Schema = types.StringValue(restInfo.Schema)
	data.SessionSecurity = types.StringValue(restInfo.SessionSecurity)
	data.BaseScope = types.StringValue(restInfo.BaseScope)
	data.Port = types.Int64Value(restInfo.Port)
	data.QueryTimeout = types.Int64Value(restInfo.QueryTimeout)
	data.LDAPSEnabled = types.BoolValue(restInfo.LDAPSEnabled)
	data.ReferralEnabled = types.BoolValue(restInfo.ReferralEnabled)
	data.UseStartTLS = types.BoolValue(restInfo.UseStartTLS)
	data.BindAsCIFSServer = types.BoolValue(restInfo.BindAsCIFSServer)
	data.BaseDN = types.StringValue(restInfo.BaseDN)

	// use SVM uuid as ID since each SVM can have one LDAP configuration
	data.ID = types.StringValue(svm.UUID)
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *NameServicesLDAPResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *NameServicesLDAPResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
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
	var errors []string
	var request interfaces.NameServicesLDAPResourceBodyDataModelONTAP
	// The update API body can include all the fields, so set all the fields
	for _, server := range data.Servers {
		request.Servers = append(request.Servers, server.ValueString())
	}
	request.Schema = data.Schema.ValueString()
	request.AdDomain = data.AdDomain.ValueString()
	request.BaseDN = data.BaseDN.ValueString()
	request.BaseScope = data.BaseScope.ValueString()
	request.BindDN = data.BindDN.ValueString()
	request.BindAsCIFSServer = data.BindAsCIFSServer.ValueBool()
	request.BindPassword = data.BindPassword.ValueString()
	for _, adserver := range data.PreferredADServers {
		request.PreferredADServers = append(request.PreferredADServers, adserver.ValueString())
	}
	request.Port = data.Port.ValueInt64()
	if data.QueryTimeout.IsNull() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 9 {
			request.QueryTimeout = data.QueryTimeout.ValueInt64()
		} else {
			errors = append(errors, "query_timeout")
		}
	}
	request.MinBindLevel = data.MinBindLevel.ValueString()
	request.UseStartTLS = data.UseStartTLS.ValueBool()
	if data.ReferralEnabled.IsNull() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 9 {
			request.ReferralEnabled = data.ReferralEnabled.ValueBool()
		} else {
			errors = append(errors, "referral_enabled")
		}
	}
	request.SessionSecurity = data.SessionSecurity.ValueString()
	if data.LDAPSEnabled.IsNull() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 9 {
			request.LDAPSEnabled = data.LDAPSEnabled.ValueBool()
		} else {
			errors = append(errors, "ldaps_enabled")
		}
	}
	if !data.SkipConfigValidation.IsNull() {
		if cluster.Version.Generation == 9 && cluster.Version.Major >= 9 {
			request.SkipConfigValidation = data.SkipConfigValidation.ValueBool()
		} else {
			errors = append(errors, "skip_config_validation")
		}
	}
	if len(errors) > 0 {
		errorsString := strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("The following Variables are not supported with current version: %#v", errorsString))
		return
	}
	// Update the resource
	err = interfaces.UpdateNameServicesLDAP(errorHandler, *client, request, data.ID.ValueString())
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *NameServicesLDAPResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *NameServicesLDAPResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	err = interfaces.DeleteNameServicesLDAP(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *NameServicesLDAPResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[1])...)
}
