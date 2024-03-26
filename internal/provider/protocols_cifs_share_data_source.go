package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/provider/protocols_cifs_share_data_source.go)
// replace ProtocolsCIFSShare with the name of the resource, following go conventions, eg IPInterface
// replace protocols_cifs_share with the name of the resource, for logging purposes, eg ip_interface
// make sure to create internal/interfaces/protocols_cifs_share.go too)
// delete these 5 lines

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsCIFSShareDataSource{}

// NewProtocolsCIFSShareDataSource is a helper function to simplify the provider implementation.
func NewProtocolsCIFSShareDataSource() datasource.DataSource {
	return &ProtocolsCIFSShareDataSource{
		config: resourceOrDataSourceConfig{
			name: "protocols_cifs_share_data_source",
		},
	}
}

// ProtocolsCIFSShareDataSource defines the data source implementation.
type ProtocolsCIFSShareDataSource struct {
	config resourceOrDataSourceConfig
}

// ProtocolsCIFSShareDataSourceModel describes the data source data model.
type ProtocolsCIFSShareDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	SVMName       types.String `tfsdk:"svm_name"`

	Acls                  types.Set    `tfsdk:"acls"`
	ChangeNotify          types.Bool   `tfsdk:"change_notify"`
	Comment               types.String `tfsdk:"comment"`
	ContinuouslyAvailable types.Bool   `tfsdk:"continuously_available"`
	DirUmask              types.Int64  `tfsdk:"dir_umask"`
	Encryption            types.Bool   `tfsdk:"encryption"`
	FileUmask             types.Int64  `tfsdk:"file_umask"`
	ForceGroupForCreate   types.String `tfsdk:"force_group_for_create"`
	HomeDirectory         types.Bool   `tfsdk:"home_directory"`
	NamespaceCaching      types.Bool   `tfsdk:"namespace_caching"`
	NoStrictSecurity      types.Bool   `tfsdk:"no_strict_security"`
	OfflineFiles          types.String `tfsdk:"offline_files"`
	Oplocks               types.Bool   `tfsdk:"oplocks"`
	Path                  types.String `tfsdk:"path"`
	ShowSnapshot          types.Bool   `tfsdk:"show_snapshot"`
	UnixSymlink           types.String `tfsdk:"unix_symlink"`
	VscanProfile          types.String `tfsdk:"vscan_profile"`
}

// Metadata returns the data source type name.
func (d *ProtocolsCIFSShareDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *ProtocolsCIFSShareDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsCIFSShare data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: `Specifies the name of the CIFS share that you want to create. If this
				is a home directory share then the share name includes the pattern as
				%w (Windows user name), %u (UNIX user name) and %d (Windows domain name)
				variables in any combination with this parameter to generate shares dynamically.
				`,
				Required: true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface svm name",
				Optional:            true,
			},
			"acls": schema.SetNestedAttribute{
				Computed:    true,
				Description: "The permissions that users and groups have on a CIFS share.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"permission": schema.StringAttribute{
							MarkdownDescription: "Specifies the access rights that a user or group has on the defined CIFS Share.",
							Computed:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("full_control", "read", "change", "no_access"),
							},
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "string Specifies the type of the user or group to add to the access control list of a CIFS share.",
							Computed:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("windows", "unix_user", "unix_group"),
							},
						},
						"user_or_group": schema.StringAttribute{
							MarkdownDescription: "Specifies the user or group name to add to the access control list of a CIFS share.",
							Computed:            true,
						},
					},
				},
			},
			"change_notify": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether CIFS clients can request for change notifications for directories on this share.",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Specify the CIFS share descriptions.",
				Computed:            true,
			},
			"continuously_available": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not the clients connecting to this share can open files in a persistent manner.Files opened in this way are protected from disruptive events, such as, failover and giveback.",
				Computed:            true,
			},
			"dir_umask": schema.Int64Attribute{
				MarkdownDescription: "Directory Mode Creation Mask to be viewed as an octal number.",
				Computed:            true,
			},
			"encryption": schema.BoolAttribute{
				MarkdownDescription: "Specifies that SMB encryption must be used when accessing this share. Clients that do not support encryption are not able to access this share.",
				Computed:            true,
			},
			"file_umask": schema.Int64Attribute{
				MarkdownDescription: "File Mode Creation Mask to be viewed as an octal number.",
				Computed:            true,
			},
			"force_group_for_create": schema.StringAttribute{
				MarkdownDescription: `Specifies that all files that CIFS users create in a specific share belong to the same group
				(also called the force-group). The force-group must be a predefined group in the UNIX group
				database. This setting has no effect unless the security style of the volume is UNIX or mixed
				security style.`,
				Computed: true,
			},
			"home_directory": schema.BoolAttribute{
				MarkdownDescription: `Specifies whether or not the share is a home directory share, where the share and path names are dynamic.
				ONTAP home directory functionality automatically offer each user a dynamic share to their home directory without creating an
				individual SMB share for each user.
				The ONTAP CIFS home directory feature enable us to configure a share that maps to
				different directories based on the user that connects to it. Instead of creating a separate shares for each user,
				a single share with a home directory parameters can be created.
				In a home directory share, ONTAP dynamically generates the share-name and share-path by substituting
				%w, %u, and %d variables with the corresponding Windows user name, UNIX user name, and domain name, respectively.`,
				Computed: true,
			},
			"namespace_caching": schema.BoolAttribute{
				MarkdownDescription: `Specifies whether or not the SMB clients connecting to this share can cache the directory enumeration
				results returned by the CIFS servers.`,
				Computed: true,
			},
			"no_strict_security": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not CIFS clients can follow a unix symlinks outside the share boundaries.",
				Computed:            true,
			},
			"offline_files": schema.StringAttribute{
				MarkdownDescription: `Offline Files. The supported values are:
				none - Clients are not permitted to cache files for offline access.
				manual - Clients may cache files that are explicitly selected by the user for offline access.
				documents - Clients may automatically cache files that are used by the user for offline access.
				programs - Clients may automatically cache files that are used by the user for offline access
				and may use those files in an offline mode even if the share is available.
				`,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("none", "manual", "documents", "programs"),
				},
			},
			"oplocks": schema.BoolAttribute{
				MarkdownDescription: `Specify whether opportunistic locks are enabled on this share. "Oplocks" allow clients to lock files and cache content locally,
				which can increase performance for file operations.
				`,
				Computed: true,
			},
			"path": schema.StringAttribute{
				MarkdownDescription: `The fully-qualified pathname in the owning SVM namespace that is shared through this share.
				If this is a home directory share then the path should be dynamic by specifying the pattern
				%w (Windows user name), %u (UNIX user name), or %d (domain name) variables in any combination.
				ONTAP generates the path dynamically for the connected user and this path is appended to each
				search path to find the full Home Directory path.
				`,
				Computed: true,
			},
			"show_snapshot": schema.BoolAttribute{
				MarkdownDescription: `Specifies whether or not the Snapshot copies can be viewed and traversed by clients.`,
				Computed:            true,
			},
			"unix_symlink": schema.StringAttribute{
				MarkdownDescription: `Controls the access of UNIX symbolic links to CIFS clients.
				The supported values are:
				* local - Enables only local symbolic links which is within the same CIFS share.
				* widelink - Enables both local symlinks and widelinks.
				* disable - Disables local symlinks and widelinks.`,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("local", "widelink", "disable"),
				},
			},
			"vscan_profile": schema.StringAttribute{
				MarkdownDescription: ` Vscan File-Operations Profile
				The supported values are:
				no_scan - Virus scans are never triggered for accesses to this share.
				standard - Virus scans can be triggered by open, close, and rename operations.
				strict - Virus scans can be triggered by open, read, close, and rename operations.
				writes_only - Virus scans can be triggered only when a file that has been modified is closed.`,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("no_scan", "standard", "strict", "writes_only"),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsCIFSShareDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsCIFSShareDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsCIFSShareDataSourceModel

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

	restInfo, err := interfaces.GetProtocolsCIFSShareByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetProtocolsCIFSShare
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.ChangeNotify = types.BoolValue(restInfo.ChangeNotify)
	data.Comment = types.StringValue(restInfo.Comment)
	data.ContinuouslyAvailable = types.BoolValue(restInfo.ContinuouslyAvailable)
	data.DirUmask = types.Int64Value(restInfo.DirUmask)
	data.Encryption = types.BoolValue(restInfo.Encryption)
	data.FileUmask = types.Int64Value(restInfo.FileUmask)
	data.ForceGroupForCreate = types.StringValue(restInfo.ForceGroupForCreate)
	data.HomeDirectory = types.BoolValue(restInfo.HomeDirectory)
	data.NamespaceCaching = types.BoolValue(restInfo.NamespaceCaching)
	data.NoStrictSecurity = types.BoolValue(restInfo.NoStrictSecurity)
	data.OfflineFiles = types.StringValue(restInfo.OfflineFiles)
	data.Oplocks = types.BoolValue(restInfo.Oplocks)
	data.Path = types.StringValue(restInfo.Path)
	data.ShowSnapshot = types.BoolValue(restInfo.ShowSnapshot)
	data.UnixSymlink = types.StringValue(restInfo.UnixSymlink)
	data.VscanProfile = types.StringValue(restInfo.VscanProfile)

	// Acls
	setElements := []attr.Value{}
	for _, acls := range restInfo.Acls {
		elementType := map[string]attr.Type{
			"permission":    types.StringType,
			"type":          types.StringType,
			"user_or_group": types.StringType,
		}
		elementValue := map[string]attr.Value{
			"permission":    types.StringValue(acls.Permission),
			"type":          types.StringValue(acls.Type),
			"user_or_group": types.StringValue(acls.UserOrGroup),
		}
		objectValue, diags := types.ObjectValue(elementType, elementValue)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		setElements = append(setElements, objectValue)
		setValue, diags := types.SetValue(types.ObjectType{
			AttrTypes: map[string]attr.Type{
				"permission":    types.StringType,
				"type":          types.StringType,
				"user_or_group": types.StringType,
			},
		}, setElements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.Acls = setValue
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
