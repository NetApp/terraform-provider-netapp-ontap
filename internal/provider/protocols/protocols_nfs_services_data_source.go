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
var _ datasource.DataSource = &ProtocolsNfsServicesDataSource{}

// NewProtocolsNfsServicesDataSource is a helper function to simplify the provider implementation.
func NewProtocolsNfsServicesDataSource() datasource.DataSource {
	return &ProtocolsNfsServicesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "nfs_services",
		},
	}
}

// NewProtocolsNfsServicesDataSource is a helper function to simplify the provider implementation.
func NewProtocolsNfsServicesDataSourceAlias() datasource.DataSource {
	return &ProtocolsNfsServicesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_nfs_services_data_source",
		},
	}
}

// ProtocolsNfsServicesDataSource defines the data source implementation.
type ProtocolsNfsServicesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsNfsServicesDataSourceModel describes the data source data model.
type ProtocolsNfsServicesDataSourceModel struct {
	CxProfileName        types.String                              `tfsdk:"cx_profile_name"`
	ProtocolsNfsServices []ProtocolsNfsServiceDataSourceModel      `tfsdk:"protocols_nfs_services"`
	Filter               *ProtocolsNfsServiceDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *ProtocolsNfsServicesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsNfsServicesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsNfsServices data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "ProtocolsNfsService svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_nfs_services": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "IPInterface svm name",
							Computed:            true,
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
				},
				Computed:            true,
				MarkdownDescription: "",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsNfsServicesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsNfsServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsNfsServicesDataSourceModel

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

	var filter *interfaces.NfsServicesFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.NfsServicesFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}

	restInfo, err := interfaces.GetProtocolsNfsServices(errorHandler, *client, filter, cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsNfsServices
		return
	}

	data.ProtocolsNfsServices = make([]ProtocolsNfsServiceDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		var permittedEncryptionTypesArray []types.String
		for _, v := range record.Security.PermittedEncrptionTypes {
			permittedEncryptionTypesArray = append(permittedEncryptionTypesArray, types.StringValue(v))
		}

		data.ProtocolsNfsServices[index] = ProtocolsNfsServiceDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVMName:       types.StringValue(record.SVM.Name),
			Enabled:       types.BoolValue(record.Enabled),
			Protocol: &ProtocolDataSourceModel{
				V3Enabled:  types.BoolValue(record.Protocol.V3Enabled),
				V4IdDomain: types.StringValue(record.Protocol.V4IdDomain),
				V40Enabled: types.BoolValue(record.Protocol.V40Enabled),
				V40Features: &V40FeaturesDataSourceModel{
					ACLEnabled:             types.BoolValue(record.Protocol.V40Features.ACLEnabled),
					ReadDelegationEnabled:  types.BoolValue(record.Protocol.V40Features.ReadDelegationEnabled),
					WriteDelegationEnabled: types.BoolValue(record.Protocol.V40Features.WriteDelegationEnabled),
				},
				V41Enabled: types.BoolValue(record.Protocol.V41Enabled),
				V41Features: &V41FeaturesDataSourceModel{
					ACLEnabled:             types.BoolValue(record.Protocol.V41Features.ACLEnabled),
					PnfsEnabled:            types.BoolValue(record.Protocol.V41Features.PnfsEnabled),
					ReadDelegationEnabled:  types.BoolValue(record.Protocol.V41Features.ReadDelegationEnabled),
					WriteDelegationEnabled: types.BoolValue(record.Protocol.V41Features.WriteDelegationEnabled),
				},
			},
			Root: &RootDataSourceModel{
				IgnoreNtACL:              types.BoolValue(record.Root.IgnoreNtACL),
				SkipWritePermissionCheck: types.BoolValue(record.Root.SkipWritePermissionCheck),
			},
			Security: &SecurityDataSourceModel{
				ChownMode:               types.StringValue(record.Security.ChownMode),
				NtACLDisplayPermission:  types.BoolValue(record.Security.NtACLDisplayPermission),
				NtfsUnixSecurity:        types.StringValue(record.Security.NtfsUnixSecurity),
				PermittedEncrptionTypes: permittedEncryptionTypesArray,
				RpcsecContextIdel:       types.Int64Value(record.Security.RpcsecContextIdel),
			},
			ShowmountEnabled: types.Bool{},
			Transport: &TransportDataSourceModel{
				TCPEnabled:     types.BoolValue(record.Transport.TCP),
				TCPMaxXferSize: types.Int64Value(record.Transport.TCPMaxXferSize),
				UDPEnabled:     types.BoolValue(record.Transport.UDP),
			},
			VstorageEnabled: types.BoolValue(record.VstorageEnabled),
			Windows: &WindowsDataSourceModel{
				DefaultUser:                types.StringValue(record.Windows.DefaultUser),
				MapUnknownUIDToDefaultUser: types.BoolValue(record.Windows.MapUnknownUIDToDefaultUser),
				V3MsDosClientEnabled:       types.BoolValue(record.Windows.V3MsDosClientEnabled),
			},
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
