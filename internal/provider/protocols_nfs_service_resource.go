package provider

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you resource (should match internal/provider/protocols_nfs_service_resource.go)
// replace ProtocolsNfsService with the name of the resource, following go conventions, eg IPInterface
// replace protocols_nfs_service with the name of the resource, for logging purposes, eg ip_interface
// make sure to create internal/interfaces/protocols_nfs_service.go too)
// delete these 5 lines

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsNfsServiceResource{}
var _ resource.ResourceWithImportState = &ProtocolsNfsServiceResource{}

// NewProtocolsNfsServiceResource is a helper function to simplify the provider implementation.
func NewProtocolsNfsServiceResource() resource.Resource {
	return &ProtocolsNfsServiceResource{
		config: resourceOrDataSourceConfig{
			name: "protocols_nfs_service_resource",
		},
	}
}

// ProtocolsNfsServiceResource defines the resource implementation.
type ProtocolsNfsServiceResource struct {
	config resourceOrDataSourceConfig
}

// ProtocolsNfsServiceResourceModel describes the resource data model.
type ProtocolsNfsServiceResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	// Protocols Nfs Services specific
	Enabled          types.Bool              `tfsdk:"enabled"`
	Protocol         *ProtocolResourceModel  `tfsdk:"protocol"`
	Root             *RootResourceModel      `tfsdk:"root"`
	Security         *SecurityResourceModel  `tfsdk:"security"`
	ShowmountEnabled types.Bool              `tfsdk:"showmount_enabled"`
	Transport        *TransportResourceModel `tfsdk:"transport"`
	VstorageEnabled  types.Bool              `tfsdk:"vstorage_enabled"`
	Windows          *WindowsResourceModel   `tfsdk:"windows"`
}

// ProtocolResourceModel describes the data source of Protocols
type ProtocolResourceModel struct {
	V3Enabled   types.Bool                `tfsdk:"v3_enabled"`
	V4IdDomain  types.String              `tfsdk:"v4_id_domain"`
	V40Enabled  types.Bool                `tfsdk:"v40_enabled"`
	V40Features *V40FeaturesResourceModel `tfsdk:"v40_features"`
	V41Enabled  types.Bool                `tfsdk:"v41_enabled"`
	V41Features *V41FeaturesResourceModel `tfsdk:"v41_features"`
}

// V40FeaturesResourceModel describes the data source of V40 Features
type V40FeaturesResourceModel struct {
	ACLEnabled             types.Bool `tfsdk:"acl_enabled"`
	ReadDelegationEnabled  types.Bool `tfsdk:"read_delegation_enabled"`
	WriteDelegationEnabled types.Bool `tfsdk:"write_delegation_enabled"`
}

// V41FeaturesResourceModel describes the data source of V41 Features
type V41FeaturesResourceModel struct {
	ACLEnabled             types.Bool `tfsdk:"acl_enabled"`
	PnfsEnabled            types.Bool `tfsdk:"pnfs_enabled"`
	ReadDelegationEnabled  types.Bool `tfsdk:"read_delegation_enabled"`
	WriteDelegationEnabled types.Bool `tfsdk:"write_delegation_enabled"`
}

// TransportResourceModel describes the data source of Transport
type TransportResourceModel struct {
	TCPEnabled     types.Bool  `tfsdk:"tcp_enabled"`
	TCPMaxXferSize types.Int64 `tfsdk:"tcp_max_transfer_size"`
	UDPEnabled     types.Bool  `tfsdk:"udp_enabled"`
}

// RootResourceModel describes the data source of Root
type RootResourceModel struct {
	IgnoreNtACL              types.Bool `tfsdk:"ignore_nt_acl"`
	SkipWritePermissionCheck types.Bool `tfsdk:"skip_write_permission_check"`
}

// WindowsResourceModel describes the data source of Windows
type WindowsResourceModel struct {
	DefaultUser                types.String `tfsdk:"default_user"`
	MapUnknownUIDToDefaultUser types.Bool   `tfsdk:"map_unknown_uid_to_default_user"`
	V3MsDosClientEnabled       types.Bool   `tfsdk:"v3_ms_dos_client_enabled"`
}

// SecurityResourceModel describes the data source of Security
type SecurityResourceModel struct {
	ChownMode               types.String   `tfsdk:"chown_mode"`
	NtACLDisplayPermission  types.Bool     `tfsdk:"nt_acl_display_permission"`
	NtfsUnixSecurity        types.String   `tfsdk:"ntfs_unix_security"`
	PermittedEncrptionTypes []types.String `tfsdk:"permitted_encryption_types"`
	RpcsecContextIdel       types.Int64    `tfsdk:"rpcsec_context_idle"`
}

// Metadata returns the resource type name.
func (r *ProtocolsNfsServiceResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *ProtocolsNfsServiceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsNfsService resource",

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
				Optional:            true,
			},
			"protocol": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Protocol",
				Attributes: map[string]schema.Attribute{
					"v3_enabled": schema.BoolAttribute{
						MarkdownDescription: "NFSv3 enabled",
						Optional:            true,
					},
					"v4_id_domain": schema.StringAttribute{
						MarkdownDescription: "User ID domain for NFSv4",
						Optional:            true,
					},
					"v40_enabled": schema.BoolAttribute{
						MarkdownDescription: "NFSv4.0 enabled",
						Optional:            true,
					},
					"v40_features": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "NFSv4.0 features",
						Attributes: map[string]schema.Attribute{
							"acl_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable ACL for NFSv4.0",
								Optional:            true,
							},
							"read_delegation_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable Read File Delegation for NFSv4.0",
								Optional:            true,
							},
							"write_delegation_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable Write File Delegation for NFSv4.0",
								Optional:            true,
							},
						},
					},
					"v41_enabled": schema.BoolAttribute{
						MarkdownDescription: "NFSv4.1 enabled",
						Optional:            true,
					},
					"v41_features": schema.SingleNestedAttribute{
						Optional:            true,
						MarkdownDescription: "NFSv4.1 features",
						Attributes: map[string]schema.Attribute{
							"acl_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable ACL for NFSv4.1",
								Optional:            true,
							},
							"pnfs_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enabled pNFS (parallel NFS) for NFSv4.1",
								Optional:            true,
							},
							"read_delegation_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable Read File Delegation for NFSv4.1",
								Optional:            true,
							},
							"write_delegation_enabled": schema.BoolAttribute{
								MarkdownDescription: "Enable Write File Delegation for NFSv4.1",
								Optional:            true,
							},
						},
					},
				},
			},
			"root": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Specific Root user options",
				Attributes: map[string]schema.Attribute{
					"ignore_nt_acl": schema.BoolAttribute{
						MarkdownDescription: "Ignore NTFS ACL for root user",
						Optional:            true,
					},
					"skip_write_permission_check": schema.BoolAttribute{
						MarkdownDescription: "Skip write permissions check for root user",
						Optional:            true,
					},
				},
			},
			"security": schema.SingleNestedAttribute{
				Optional:            true,
				MarkdownDescription: "NFS Security options",
				Attributes: map[string]schema.Attribute{
					"chown_mode": schema.StringAttribute{
						MarkdownDescription: "Specifies whether file ownership can be changed only by the superuser, or if a non-root user can also change file ownership",
						Optional:            true,
					},
					"nt_acl_display_permission": schema.BoolAttribute{
						MarkdownDescription: "Controls the permissions that are displayed to NFSv3 and NFSv4 clients on a file or directory that has an NT ACL set",
						Optional:            true,
					},
					"ntfs_unix_security": schema.StringAttribute{
						MarkdownDescription: "Specifies how NFSv3 security changes affect NTFS volumes",
						Optional:            true,
					},
					"permitted_encryption_types": schema.ListAttribute{
						ElementType:         types.StringType,
						Optional:            true,
						MarkdownDescription: "Specifies the permitted encryption types for Kerberos over NFS.",
					},
					"rpcsec_context_idle": schema.Int64Attribute{
						MarkdownDescription: "Specifies, in seconds, the amount of time a RPCSEC_GSS context is permitted to remain unused before it is deleted",
						Optional:            true,
					},
				},
			},
			"showmount_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether SVM allows showmount",
				Optional:            true,
			},
			"transport": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"tcp_enabled": schema.BoolAttribute{
						MarkdownDescription: "tcp enabled",
						Optional:            true,
					},
					"tcp_max_transfer_size": schema.Int64Attribute{
						MarkdownDescription: "Max tcp transfter size",
						Optional:            true,
					},
					"udp_enabled": schema.BoolAttribute{
						MarkdownDescription: "udp enabled",
						Optional:            true,
					},
				},
			},
			"vstorage_enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether Vstorage is enabled",
				Optional:            true,
			},
			"windows": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"default_user": schema.StringAttribute{
						MarkdownDescription: "default Windows user for the NFS server",
						Optional:            true,
					},
					"map_unknown_uid_to_default_user": schema.BoolAttribute{
						MarkdownDescription: "whether or not the mapping of an unknown UID to the default Windows user is enabled",
						Optional:            true,
					},
					"v3_ms_dos_client_enabled": schema.BoolAttribute{
						MarkdownDescription: "if permission checks are to be skipped for NFS WRITE calls from root/owner.",
						Optional:            true,
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsNfsServiceResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ProtocolsNfsServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsNfsServiceResourceModel

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
	data.Protocol = &ProtocolResourceModel{
		V3Enabled:  types.BoolValue(restInfo.Protocol.V3Enabled),
		V4IdDomain: types.StringValue(restInfo.Protocol.V4IdDomain),
		V40Enabled: types.BoolValue(restInfo.Protocol.V40Enabled),
		V40Features: &V40FeaturesResourceModel{
			ACLEnabled:             types.BoolValue(restInfo.Protocol.V40Features.ACLEnabled),
			ReadDelegationEnabled:  types.BoolValue(restInfo.Protocol.V40Features.ReadDelegationEnabled),
			WriteDelegationEnabled: types.BoolValue(restInfo.Protocol.V40Features.WriteDelegationEnabled),
		},
		V41Enabled: types.BoolValue(restInfo.Protocol.V41Enabled),
		V41Features: &V41FeaturesResourceModel{
			ACLEnabled:             types.BoolValue(restInfo.Protocol.V41Features.ACLEnabled),
			PnfsEnabled:            types.BoolValue(restInfo.Protocol.V41Features.PnfsEnabled),
			ReadDelegationEnabled:  types.BoolValue(restInfo.Protocol.V41Features.ReadDelegationEnabled),
			WriteDelegationEnabled: types.BoolValue(restInfo.Protocol.V41Features.WriteDelegationEnabled),
		},
	}
	data.Root = &RootResourceModel{
		IgnoreNtACL:              types.BoolValue(restInfo.Root.IgnoreNtACL),
		SkipWritePermissionCheck: types.BoolValue(restInfo.Root.SkipWritePermissionCheck),
	}
	data.Security = &SecurityResourceModel{
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
	data.Transport = &TransportResourceModel{
		TCPEnabled:     types.BoolValue(restInfo.Transport.TCP),
		TCPMaxXferSize: types.Int64Value(restInfo.Transport.TCPMaxXferSize),
		UDPEnabled:     types.BoolValue(restInfo.Transport.UDP),
	}
	data.VstorageEnabled = types.BoolValue(restInfo.VstorageEnabled)
	data.Windows = &WindowsResourceModel{
		DefaultUser:                types.StringValue(restInfo.Windows.DefaultUser),
		MapUnknownUIDToDefaultUser: types.BoolValue(restInfo.Windows.MapUnknownUIDToDefaultUser),
		V3MsDosClientEnabled:       types.BoolValue(restInfo.Windows.V3MsDosClientEnabled),
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *ProtocolsNfsServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProtocolsNfsServiceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ProtocolsNfsServiceGetDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

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
	clusterVersion := strconv.Itoa(cluster.Version.Generation) + "." + strconv.Itoa(cluster.Version.Major)
	var errors []string

	if !data.Enabled.IsNull() {
		body.Enabled = data.Enabled.ValueBool()
	}
	if !data.Protocol.V3Enabled.IsNull() {
		body.Protocol.V3Enabled = data.Protocol.V3Enabled.ValueBool()
	}
	if !data.Protocol.V4IdDomain.IsNull() {
		body.Protocol.V4IdDomain = data.Protocol.V4IdDomain.ValueString()
	}
	if !data.Protocol.V40Enabled.IsNull() {
		body.Protocol.V40Enabled = data.Protocol.V40Enabled.ValueBool()
	}
	if !data.Protocol.V40Features.ACLEnabled.IsNull() {
		body.Protocol.V40Features.ACLEnabled = data.Protocol.V40Features.ACLEnabled.ValueBool()
	}
	if !data.Protocol.V40Features.ReadDelegationEnabled.IsNull() {
		body.Protocol.V40Features.ReadDelegationEnabled = data.Protocol.V40Features.ReadDelegationEnabled.ValueBool()
	}
	if !data.Protocol.V40Features.WriteDelegationEnabled.IsNull() {
		body.Protocol.V40Features.WriteDelegationEnabled = data.Protocol.V40Features.ReadDelegationEnabled.ValueBool()
	}
	if !data.Protocol.V41Enabled.IsNull() {
		body.Protocol.V41Enabled = data.Protocol.V41Enabled.ValueBool()
	}
	if !data.Protocol.V41Features.ACLEnabled.IsNull() {
		body.Protocol.V41Features.ACLEnabled = data.Protocol.V41Features.ACLEnabled.ValueBool()
	}
	if !data.Protocol.V41Features.PnfsEnabled.IsNull() {
		body.Protocol.V41Features.PnfsEnabled = data.Protocol.V41Features.PnfsEnabled.ValueBool()
	}
	if !data.Protocol.V41Features.ReadDelegationEnabled.IsNull() {
		body.Protocol.V41Features.ReadDelegationEnabled = data.Protocol.V41Features.ReadDelegationEnabled.ValueBool()
	}
	if !data.Protocol.V41Features.WriteDelegationEnabled.IsNull() {
		body.Protocol.V41Features.WriteDelegationEnabled = data.Protocol.V41Features.WriteDelegationEnabled.ValueBool()
	}
	if !data.Root.IgnoreNtACL.IsNull() && clusterVersion > "9.10" {
		body.Root.IgnoreNtACL = data.Root.IgnoreNtACL.ValueBool()
	} else if !data.Root.IgnoreNtACL.IsNull() && clusterVersion <= "9.10" {
		errors = append(errors, "root.ignore_nt_acl")
	}
	if !data.Root.SkipWritePermissionCheck.IsNull() && clusterVersion > "9.10" {
		body.Root.SkipWritePermissionCheck = data.Root.SkipWritePermissionCheck.ValueBool()
	} else if !data.Root.SkipWritePermissionCheck.IsNull() && clusterVersion <= "9.10" {
		errors = append(errors, "root.skip_write_permission_check")
	}
	if !data.Security.ChownMode.IsNull() && clusterVersion > "9.10" {
		body.Security.ChownMode = data.Security.ChownMode.ValueString()
	} else if !data.Security.ChownMode.IsNull() && clusterVersion <= "9.10" {
		errors = append(errors, "security.chown_mode")
	}
	if !data.Security.NtACLDisplayPermission.IsNull() && clusterVersion > "9.10" {
		body.Security.NtACLDisplayPermission = data.Security.NtACLDisplayPermission.ValueBool()
	} else if !data.Security.NtACLDisplayPermission.IsNull() && clusterVersion <= "9.10" {
		errors = append(errors, "security.nt_acl_display_permission")
	}
	if !data.Security.NtfsUnixSecurity.IsNull() && clusterVersion > "9.10" {
		body.Security.NtfsUnixSecurity = data.Security.NtfsUnixSecurity.ValueString()
	} else if !data.Security.NtfsUnixSecurity.IsNull() && clusterVersion <= "9.10" {
		errors = append(errors, "security.ntfs_unix_security")
	}
	if !data.Security.RpcsecContextIdel.IsNull() && clusterVersion > "9.10" {
		body.Security.RpcsecContextIdel = data.Security.RpcsecContextIdel.ValueInt64()
	} else if !data.Security.RpcsecContextIdel.IsNull() && clusterVersion <= "9.10" {
		errors = append(errors, "security.rpcsec_context_idle")
	}
	var ptypes []string
	for _, k := range data.Security.PermittedEncrptionTypes {
		ptypes = append(ptypes, k.String())
	}
	if ptypes != nil && clusterVersion > "9.10" {
		body.Security.PermittedEncrptionTypes = ptypes
	} else if ptypes != nil && clusterVersion <= "9.10" {
		errors = append(errors, "security.permitted_encryption_types")
	}
	if !data.ShowmountEnabled.IsNull() {
		body.ShowmountEnabled = data.ShowmountEnabled.ValueBool()
	}
	if !data.Transport.TCPEnabled.IsNull() {
		body.Transport.TCP = data.Transport.TCPEnabled.ValueBool()
	}
	if !data.Transport.TCPMaxXferSize.IsNull() && clusterVersion > "9.10" {
		body.Transport.TCPMaxXferSize = data.Transport.TCPMaxXferSize.ValueInt64()
	} else if !data.Transport.TCPMaxXferSize.IsNull() && clusterVersion <= "9.10" {
		errors = append(errors, "transport.tcp_max_transfer_size")
	}
	if !data.Transport.UDPEnabled.IsNull() {
		body.Transport.UDP = data.Transport.UDPEnabled.ValueBool()
	}
	if !data.VstorageEnabled.IsNull() {
		body.VstorageEnabled = data.VstorageEnabled.ValueBool()
	}
	if !data.Windows.DefaultUser.IsNull() && clusterVersion > "9.10" {
		body.Windows.DefaultUser = data.Windows.DefaultUser.ValueString()
	} else if !data.Windows.DefaultUser.IsNull() && clusterVersion <= "9.10" {
		errors = append(errors, "windows.default_user")
	}
	if !data.Windows.MapUnknownUIDToDefaultUser.IsNull() && clusterVersion > "9.10" {
		body.Windows.MapUnknownUIDToDefaultUser = data.Windows.MapUnknownUIDToDefaultUser.ValueBool()
	} else if !data.Windows.MapUnknownUIDToDefaultUser.IsNull() && clusterVersion <= "9.10" {
		errors = append(errors, "windows.map_unknown_uid_to_default_user")
	}
	if !data.Windows.V3MsDosClientEnabled.IsNull() && clusterVersion > "9.10" {
		body.Windows.V3MsDosClientEnabled = data.Windows.V3MsDosClientEnabled.ValueBool()
	} else if !data.Windows.V3MsDosClientEnabled.IsNull() && clusterVersion <= "9.10" {
		errors = append(errors, "windows.v3_ms_dos_client_enabled")
	}
	if len(errors) > 0 {
		errorsString := strings.Join(errors, ", ")
		tflog.Error(ctx, fmt.Sprintf("The following Variables are only support with ONTAP 9.11 or higher: %#v", errorsString))
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	_, err = interfaces.CreateProtocolsNfsService(errorHandler, *client, body, svm.Name)
	if err != nil {
		return
	}

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsNfsServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ProtocolsNfsServiceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsNfsServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsNfsServiceResourceModel

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
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	err = interfaces.DeleteProtocolsNfsService(errorHandler, *client, svm.UUID)
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsNfsServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
