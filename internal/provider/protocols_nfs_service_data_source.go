package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsNfsServiceDataSource{}

// NewProtocolsNfsServiceDataSource is a helper function to simplify the provider implementation.
func NewProtocolsNfsServiceDataSource() datasource.DataSource {
	return &ProtocolsNfsServiceDataSource{
		config: resourceOrDataSourceConfig{
			name: "protcols_nfs_service_data_source",
		},
	}
}

// ProtocolsNfsServiceDataSource defines the data source implementation.
type ProtocolsNfsServiceDataSource struct {
	config resourceOrDataSourceConfig
}

// ProtocolsNfsServiceDataSourceModel describes the data source data model.
type ProtocolsNfsServiceDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	// Protocols Nfs Services specific
	Enabled          types.Bool                `tfsdk:"enabled"`
	Protocol         *ProtocolDataSourceModel  `tfsdk:"protocol"`
	Root             *RootDataSourceModel      `tfsdk:"root"`
	Security         *SecurityDataSourceModel  `tfsdk:"security"`
	ShowmountEnabled types.Bool                `tfsdk:"showmount_enabled"`
	Transport        *TransportDataSourceModel `tfsdk:"transport"`
	VstorageEnabled  types.Bool                `tfsdk:"vstorage_enabled"`
	Windows          *WindowsDataSourceModel   `tfsdk:"windows"`
}

// ProtocolDataSourceModel describes the data source of Protocols
type ProtocolDataSourceModel struct {
	V3Enabled   types.Bool                  `tfsdk:"v3_enabled"`
	V4IdDomain  types.String                `tfsdk:"v4_id_domain"`
	V40Enabled  types.Bool                  `tfsdk:"v40_enabled"`
	V40Features *V40FeaturesDataSourceModel `tfsdk:"v40_features"`
	V41Enabled  types.Bool                  `tfsdk:"v41_enabled"`
	V41Features *V41FeaturesDataSourceModel `tfsdk:"v41_features"`
}

// V40FeaturesDataSourceModel describes the data source of V40 Features
type V40FeaturesDataSourceModel struct {
	ACLEnabled             types.Bool `tfsdk:"acl_enabled"`
	ReadDelegationEnabled  types.Bool `tfsdk:"read_delegation_enabled"`
	WriteDelegationEnabled types.Bool `tfsdk:"write_delegation_enabled"`
}

// V41FeaturesDataSourceModel describes the data source of V41 Features
type V41FeaturesDataSourceModel struct {
	ACLEnabled             types.Bool `tfsdk:"acl_enabled"`
	PnfsEnabled            types.Bool `tfsdk:"pnfs_enabled"`
	ReadDelegationEnabled  types.Bool `tfsdk:"read_delegation_enabled"`
	WriteDelegationEnabled types.Bool `tfsdk:"write_delegation_enabled"`
}

// TransportDataSourceModel describes the data source of Transport
type TransportDataSourceModel struct {
	TCPEnabled     types.Bool  `tfsdk:"tcp_enabled"`
	TCPMaxXferSize types.Int64 `tfsdk:"tcp_max_transfer_size"`
	UDPEnabled     types.Bool  `tfsdk:"udp_enabled"`
}

// RootDataSourceModel describes the data source of Root
type RootDataSourceModel struct {
	IgnoreNtACL              types.Bool `tfsdk:"ignore_nt_acl"`
	SkipWritePermissionCheck types.Bool `tfsdk:"skip_write_permission_check"`
}

// WindowsDataSourceModel describes the data source of Windows
type WindowsDataSourceModel struct {
	DefaultUser                types.String `tfsdk:"default_user"`
	MapUnknownUIDToDefaultUser types.Bool   `tfsdk:"map_unknown_uid_to_default_user"`
	V3MsDosClientEnabled       types.Bool   `tfsdk:"v3_ms_dos_client_enabled"`
}

// SecurityDataSourceModel describes the data source of Security
type SecurityDataSourceModel struct {
	ChownMode               types.String   `tfsdk:"chown_mode"`
	NtACLDisplayPermission  types.Bool     `tfsdk:"nt_acl_display_permission"`
	NtfsUnixSecurity        types.String   `tfsdk:"ntfs_unix_security"`
	PermittedEncrptionTypes []types.String `tfsdk:"permitted_encryption_types"`
	RpcsecContextIdel       types.Int64    `tfsdk:"rpcsec_context_idle"`
}

// ProtocolsNfsServiceDataSourceFilterModel describes the data source data model for queries.
type ProtocolsNfsServiceDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm.name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsNfsServiceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *ProtocolsNfsServiceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsNfsService data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface vserver name",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "NFS should be enabled or disabled",
				Computed:            true,
			},
			"protocol": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Protocol",
				Attributes: map[string]schema.Attribute{
					"v3_enabled": schema.BoolAttribute{
						MarkdownDescription: "NFSv3 enabled",
						Computed:            true,
					},
					"v4_id_domain": schema.StringAttribute{
						MarkdownDescription: "User ID domain for NFSv4",
						Computed:            true,
					},
					"v40_enabled": schema.BoolAttribute{
						MarkdownDescription: "NFSv4.0 enabled",
						Computed:            true,
					},
					"v40_features": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "NFSv4.0 features",
						Attributes: map[string]schema.Attribute{
							"acl_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable ACL for NFSv4.0",
								Computed:            true,
							},
							"read_delegation_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable Read File Delegation for NFSv4.0",
								Computed:            true,
							},
							"write_delegation_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable Write File Delegation for NFSv4.0",
								Computed:            true,
							},
						},
					},
					"v41_enabled": schema.BoolAttribute{
						MarkdownDescription: "NFSv4.1 enabled",
						Computed:            true,
					},
					"v41_features": schema.SingleNestedAttribute{
						Computed:            true,
						MarkdownDescription: "NFSv4.1 features",
						Attributes: map[string]schema.Attribute{
							"acl_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable ACL for NFSv4.1",
								Computed:            true,
							},
							"pnfs_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enabled pNFS (parallel NFS) for NFSv4.1",
								Computed:            true,
							},
							"read_delegation_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable Read File Delegation for NFSv4.1",
								Computed:            true,
							},
							"write_delegation_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable Write File Delegation for NFSv4.1",
								Computed:            true,
							},
						},
					},
				},
			},
			"root": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "Specific Root user options",
				Attributes: map[string]schema.Attribute{
					"ignore_nt_acl": schema.BoolAttribute{
						MarkdownDescription: "Ignore NTFS ACL for root user",
						Computed:            true,
					},
					"skip_write_permission_check": schema.BoolAttribute{
						MarkdownDescription: "Skip write permissions check for root user",
						Computed:            true,
					},
				},
			},
			"security": schema.SingleNestedAttribute{
				Computed:            true,
				MarkdownDescription: "NFS Security options",
				Attributes: map[string]schema.Attribute{
					"chown_mode": schema.StringAttribute{
						MarkdownDescription: "Specifies whether file ownership can be changed only by the superuser, or if a non-root user can also change file ownership",
						Computed:            true,
					},
					"nt_acl_display_permission": schema.BoolAttribute{
						MarkdownDescription: "Controls the permissions that are displayed to NFSv3 and NFSv4 clients on a file or directory that has an NT ACL set",
						Computed:            true,
					},
					"ntfs_unix_security": schema.StringAttribute{
						MarkdownDescription: "Specifies how NFSv3 security changes affect NTFS volumes",
						Computed:            true,
					},
					"permitted_encryption_types": schema.ListAttribute{
						ElementType:         types.StringType,
						Computed:            true,
						MarkdownDescription: "Specifies the permitted encryption types for Kerberos over NFS.",
					},
					"rpcsec_context_idle": schema.Int64Attribute{
						MarkdownDescription: "Specifies, in seconds, the amount of time a RPCSEC_GSS context is permitted to remain unused before it is deleted",
						Computed:            true,
					},
				},
			},
			"showmount_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether SVM allows showmount",
				Computed:            true,
			},
			"transport": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"tcp_enabled": schema.BoolAttribute{
						MarkdownDescription: "tcp enabled",
						Computed:            true,
					},
					"tcp_max_transfer_size": schema.Int64Attribute{
						MarkdownDescription: "Max tcp transfter size",
						Computed:            true,
					},
					"udp_enabled": schema.BoolAttribute{
						MarkdownDescription: "udp enabled",
						Computed:            true,
					},
				},
			},
			"vstorage_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Vstorage is enabled",
				Computed:            true,
			},
			"windows": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"default_user": schema.StringAttribute{
						MarkdownDescription: "default Windows user for the NFS server",
						Computed:            true,
					},
					"map_unknown_uid_to_default_user": schema.BoolAttribute{
						MarkdownDescription: "whether or not the mapping of an unknown UID to the default Windows user is enabled",
						Computed:            true,
					},
					"v3_ms_dos_client_enabled": schema.BoolAttribute{
						MarkdownDescription: "if permission checks are to be skipped for NFS WRITE calls from root/owner.",
						Computed:            true,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsNfsServiceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	d.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *ProtocolsNfsServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsNfsServiceDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}

	restInfo, err := interfaces.GetProtocolsNfsService(errorHandler, *client, data.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsNfsService
		return
	}
	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.Protocol = &ProtocolDataSourceModel{
		V3Enabled:  types.BoolValue(restInfo.Protocol.V3Enabled),
		V4IdDomain: types.StringValue(restInfo.Protocol.V4IdDomain),
		V40Enabled: types.BoolValue(restInfo.Protocol.V40Enabled),
		V40Features: &V40FeaturesDataSourceModel{
			ACLEnabled:             types.BoolValue(restInfo.Protocol.V40Features.ACLEnabled),
			ReadDelegationEnabled:  types.BoolValue(restInfo.Protocol.V40Features.ReadDelegationEnabled),
			WriteDelegationEnabled: types.BoolValue(restInfo.Protocol.V40Features.WriteDelegationEnabled),
		},
		V41Enabled: types.BoolValue(restInfo.Protocol.V41Enabled),
		V41Features: &V41FeaturesDataSourceModel{
			ACLEnabled:             types.BoolValue(restInfo.Protocol.V41Features.ACLEnabled),
			PnfsEnabled:            types.BoolValue(restInfo.Protocol.V41Features.PnfsEnabled),
			ReadDelegationEnabled:  types.BoolValue(restInfo.Protocol.V41Features.ReadDelegationEnabled),
			WriteDelegationEnabled: types.BoolValue(restInfo.Protocol.V41Features.WriteDelegationEnabled),
		},
	}
	data.Root = &RootDataSourceModel{
		IgnoreNtACL:              types.BoolValue(restInfo.Root.IgnoreNtACL),
		SkipWritePermissionCheck: types.BoolValue(restInfo.Root.SkipWritePermissionCheck),
	}
	data.Security = &SecurityDataSourceModel{
		ChownMode:              types.StringValue(restInfo.Security.ChownMode),
		NtACLDisplayPermission: types.BoolValue(restInfo.Security.NtACLDisplayPermission),
		NtfsUnixSecurity:       types.StringValue(restInfo.Security.NtfsUnixSecurity),
		RpcsecContextIdel:      types.Int64Value(restInfo.Security.RpcsecContextIdel),
	}
	var ptypes []types.String
	for _, v := range restInfo.Security.PermittedEncrptionTypes {
		ptypes = append(ptypes, types.StringValue(v))
	}
	data.Security.PermittedEncrptionTypes = ptypes

	data.ShowmountEnabled = types.BoolValue(restInfo.ShowmountEnabled)
	data.Transport = &TransportDataSourceModel{
		TCPEnabled:     types.BoolValue(restInfo.Transport.TCP),
		TCPMaxXferSize: types.Int64Value(restInfo.Transport.TCPMaxXferSize),
		UDPEnabled:     types.BoolValue(restInfo.Transport.UDP),
	}
	data.VstorageEnabled = types.BoolValue(restInfo.VstorageEnabled)
	data.Windows = &WindowsDataSourceModel{
		DefaultUser:                types.StringValue(restInfo.Windows.DefaultUser),
		MapUnknownUIDToDefaultUser: types.BoolValue(restInfo.Windows.MapUnknownUIDToDefaultUser),
		V3MsDosClientEnabled:       types.BoolValue(restInfo.Windows.V3MsDosClientEnabled),
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
