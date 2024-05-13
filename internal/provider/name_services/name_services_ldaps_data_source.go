package name_services

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
var _ datasource.DataSource = &NameServicesLDAPsDataSource{}

// NewNameServicesLDAPsDataSource is a helper function to simplify the provider implementation.
func NewNameServicesLDAPsDataSource() datasource.DataSource {
	return &NameServicesLDAPsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "name_services_ldaps_data_source",
		},
	}
}

// NameServicesLDAPsDataSource defines the data source implementation.
type NameServicesLDAPsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesLDAPsDataSourceModel describes the data source data model.
type NameServicesLDAPsDataSourceModel struct {
	CxProfileName     types.String                            `tfsdk:"cx_profile_name"`
	NameServicesLDAPs []NameServicesLDAPDataSourceModel       `tfsdk:"name_services_ldaps"`
	Filter            *NameServicesLDAPsDataSourceFilterModel `tfsdk:"filter"`
}

// NameServicesLDAPsDataSourceFilterModel describes the data source data model for queries.
type NameServicesLDAPsDataSourceFilterModel struct {
	SVMName      types.String `tfsdk:"svm_name"`
	BaseScope    types.String `tfsdk:"base_scope"`
	MinBindLevel types.String `tfsdk:"min_bind_level"`
}

// Metadata returns the data source type name.
func (d *NameServicesLDAPsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *NameServicesLDAPsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesLDAPs data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "NameServicesLDAP svm name",
						Optional:            true,
					},
					"min_bind_level": schema.StringAttribute{
						MarkdownDescription: "The minimum bind authentication level",
						Optional:            true,
					},
					"base_scope": schema.StringAttribute{
						MarkdownDescription: "Specifies the default search scope for LDAP queries",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"name_services_ldaps": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "IPInterface svm name",
							Computed:            true,
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
				},
				Computed:            true,
				MarkdownDescription: "",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *NameServicesLDAPsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *NameServicesLDAPsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data NameServicesLDAPsDataSourceModel

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

	var filter *interfaces.NameServicesLDAPDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.NameServicesLDAPDataSourceFilterModel{
			SVMName:      data.Filter.SVMName.ValueString(),
			MinBindLevel: data.Filter.MinBindLevel.ValueString(),
			BaseScope:    data.Filter.BaseScope.ValueString(),
		}
	}
	restInfo, err := interfaces.GetNameServicesLDAPs(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetNameServicesLDAPs
		return
	}

	data.NameServicesLDAPs = make([]NameServicesLDAPDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		adServers := make([]types.String, len(record.PreferredADServers))
		for i, server := range record.PreferredADServers {
			adServers[i] = types.StringValue(server)
		}
		servers := make([]types.String, len(record.Servers))
		for i, server := range record.Servers {
			servers[i] = types.StringValue(server)
		}
		data.NameServicesLDAPs[index] = NameServicesLDAPDataSourceModel{
			CxProfileName:      types.String(data.CxProfileName),
			AdDomain:           types.StringValue(record.AdDomain),
			BaseDN:             types.StringValue(record.BaseDN),
			BaseScope:          types.StringValue(record.BaseScope),
			BindDN:             types.StringValue(record.BindDN),
			BindAsCIFSServer:   types.BoolValue(record.BindAsCIFSServer),
			PreferredADServers: adServers,
			Port:               types.Int64Value(record.Port),
			QueryTimeout:       types.Int64Value(record.QueryTimeout),
			MinBindLevel:       types.StringValue(record.MinBindLevel),
			UseStartTLS:        types.BoolValue(record.UseStartTLS),
			ReferralEnabled:    types.BoolValue(record.ReferralEnabled),
			SessionSecurity:    types.StringValue(record.SessionSecurity),
			LDAPSEnabled:       types.BoolValue(record.LDAPSEnabled),
			Servers:            servers,
			Schema:             types.StringValue(record.Schema),
			SVMName:            types.StringValue(record.SVM.Name),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
