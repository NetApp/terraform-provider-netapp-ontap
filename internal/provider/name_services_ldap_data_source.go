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
var _ datasource.DataSource = &NameServicesLDAPDataSource{}

// NewNameServicesLDAPDataSource is a helper function to simplify the provider implementation.
func NewNameServicesLDAPDataSource() datasource.DataSource {
	return &NameServicesLDAPDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "name_services_ldap_data_source",
		},
	}
}

// NameServicesLDAPDataSource defines the data source implementation.
type NameServicesLDAPDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesLDAPDataSourceModel describes the data source data model.
type NameServicesLDAPDataSourceModel struct {
	CxProfileName      types.String   `tfsdk:"cx_profile_name"`
	SVMName            types.String   `tfsdk:"svm_name"`
	Servers            []types.String `tfsdk:"servers"`
	Schema             types.String   `tfsdk:"schema"`
	AdDomain           types.String   `tfsdk:"ad_domain"`
	BaseDN             types.String   `tfsdk:"base_dn"`
	BaseScope          types.String   `tfsdk:"base_scope"`
	BindDN             types.String   `tfsdk:"bind_dn"`
	BindAsCIFSServer   types.Bool     `tfsdk:"bind_as_cifs_server"`
	PreferredADServers []types.String `tfsdk:"preferred_ad_servers"`
	Port               types.Int64    `tfsdk:"port"`
	QueryTimeout       types.Int64    `tfsdk:"query_timeout"`
	MinBindLevel       types.String   `tfsdk:"min_bind_level"`
	UseStartTLS        types.Bool     `tfsdk:"use_start_tls"`
	ReferralEnabled    types.Bool     `tfsdk:"referral_enabled"`
	SessionSecurity    types.String   `tfsdk:"session_security"`
	LDAPSEnabled       types.Bool     `tfsdk:"ldaps_enabled"`
}

// Metadata returns the data source type name.
func (d *NameServicesLDAPDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *NameServicesLDAPDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesLDAP data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface svm name",
				Required:            true,
			},
			"servers": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of LDAP servers used for this client configuration",
				Computed:            true,
			},
			"schema": schema.StringAttribute{
				MarkdownDescription: "The name of the schema template used by the SVM",
				Computed:            true,
			},
			"base_dn": schema.StringAttribute{
				MarkdownDescription: "Specifies the default base DN for all searches",
				Computed:            true,
			},
			"ldaps_enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not LDAPS is enabled",
				Computed:            true,
			},
			"min_bind_level": schema.StringAttribute{
				MarkdownDescription: "The minimum bind authentication level",
				Computed:            true,
			},
			"bind_dn": schema.StringAttribute{
				MarkdownDescription: "Specifies the user that binds to the LDAP servers",
				Computed:            true,
			},
			"preferred_ad_servers": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "This parameter specifies a list of LDAP servers preferred over discovered servers",
				Computed:            true,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "The port used to connect to the LDAP Servers",
				Computed:            true,
			},
			"session_security": schema.StringAttribute{
				MarkdownDescription: "Specifies the level of security to be used for LDAP communications",
				Computed:            true,
			},
			"use_start_tls": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not to use Start TLS over LDAP connections",
				Computed:            true,
			},
			"referral_enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not LDAP referral is enabled",
				Computed:            true,
			},
			"ad_domain": schema.StringAttribute{
				MarkdownDescription: "Specifies the name of the Active Directory domain used to discover LDAP servers for use by this client",
				Computed:            true,
			},
			"bind_as_cifs_server": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not CIFS server's credentials are used to bind to the LDAP server",
				Computed:            true,
			},
			"base_scope": schema.StringAttribute{
				MarkdownDescription: "Specifies the default search scope for LDAP queries",
				Computed:            true,
			},
			"query_timeout": schema.Int64Attribute{
				MarkdownDescription: "Specifies the timeout for LDAP queries",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NameServicesLDAPDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *NameServicesLDAPDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NameServicesLDAPDataSourceModel

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

	restInfo, err := interfaces.GetNameServicesLDAPBySVMName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetNameServicesLDAP
		return
	}

	data.Schema = types.StringValue(restInfo.Schema)
	data.AdDomain = types.StringValue(restInfo.AdDomain)
	data.BaseDN = types.StringValue(restInfo.BaseDN)
	data.BaseScope = types.StringValue(restInfo.BaseScope)
	data.BindDN = types.StringValue(restInfo.BindDN)
	data.BindAsCIFSServer = types.BoolValue(restInfo.BindAsCIFSServer)
	data.Servers = make([]types.String, len(restInfo.Servers))
	for index, server := range restInfo.Servers {
		data.Servers[index] = types.StringValue(server)
	}
	data.PreferredADServers = make([]types.String, len(restInfo.PreferredADServers))
	for index, adserver := range restInfo.PreferredADServers {
		data.PreferredADServers[index] = types.StringValue(adserver)
	}
	data.Port = types.Int64Value(restInfo.Port)
	data.QueryTimeout = types.Int64Value(restInfo.QueryTimeout)
	data.MinBindLevel = types.StringValue(restInfo.MinBindLevel)
	data.UseStartTLS = types.BoolValue(restInfo.UseStartTLS)
	data.ReferralEnabled = types.BoolValue(restInfo.ReferralEnabled)
	data.SessionSecurity = types.StringValue(restInfo.SessionSecurity)
	data.LDAPSEnabled = types.BoolValue(restInfo.LDAPSEnabled)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
