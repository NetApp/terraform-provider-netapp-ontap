package provider

import (
	"context"
	"fmt"
	"log"

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

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsCIFSSharesDataSource{}

// NewProtocolsCIFSSharesDataSource is a helper function to simplify the provider implementation.
func NewProtocolsCIFSSharesDataSource() datasource.DataSource {
	return &ProtocolsCIFSSharesDataSource{
		config: resourceOrDataSourceConfig{
			name: "protocols_cifs_shares_data_source",
		},
	}
}

// ProtocolsCIFSSharesDataSource defines the data source implementation.
type ProtocolsCIFSSharesDataSource struct {
	config resourceOrDataSourceConfig
}

// ProtocolsCIFSSharesDataSourceModel describes the data source data model.
type ProtocolsCIFSSharesDataSourceModel struct {
	CxProfileName       types.String                              `tfsdk:"cx_profile_name"`
	ProtocolsCIFSShares []ProtocolsCIFSShareDataSourceModel       `tfsdk:"protocols_cifs_shares"`
	Filter              *ProtocolsCIFSSharesDataSourceFilterModel `tfsdk:"filter"`
}

// ProtocolsCIFSSharesDataSourceFilterModel describes the data source data model for queries.
type ProtocolsCIFSSharesDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsCIFSSharesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *ProtocolsCIFSSharesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsCIFSShares data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "ProtocolsCIFSShare name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "ProtocolsCIFSShare svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_cifs_shares": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
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
				},
				Computed:            true,
				MarkdownDescription: "",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsCIFSSharesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsCIFSSharesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsCIFSSharesDataSourceModel

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

	var filter *interfaces.ProtocolsCIFSShareDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.ProtocolsCIFSShareDataSourceFilterModel{
			Name: data.Filter.Name.ValueString(),
		}
	}
	restInfo, err := interfaces.GetProtocolsCIFSShares(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetProtocolsCIFSShares
		return
	}
	data.ProtocolsCIFSShares = make([]ProtocolsCIFSShareDataSourceModel, len(restInfo))
	log.Printf("restInfo is: %#v", len(restInfo))
	for index, record := range restInfo {
		data.ProtocolsCIFSShares[index] = ProtocolsCIFSShareDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
		}
		// data.ProtocolsCIFSShares[index].Name = types.StringValue(restInfo.Name)
		data.ProtocolsCIFSShares[index].ChangeNotify = types.BoolValue(record.ChangeNotify)
		data.ProtocolsCIFSShares[index].Comment = types.StringValue(record.Comment)
		data.ProtocolsCIFSShares[index].ContinuouslyAvailable = types.BoolValue(record.ContinuouslyAvailable)
		data.ProtocolsCIFSShares[index].DirUmask = types.Int64Value(record.DirUmask)
		data.ProtocolsCIFSShares[index].Encryption = types.BoolValue(record.Encryption)
		data.ProtocolsCIFSShares[index].FileUmask = types.Int64Value(record.FileUmask)
		data.ProtocolsCIFSShares[index].ForceGroupForCreate = types.StringValue(record.ForceGroupForCreate)
		data.ProtocolsCIFSShares[index].HomeDirectory = types.BoolValue(record.HomeDirectory)
		data.ProtocolsCIFSShares[index].NamespaceCaching = types.BoolValue(record.NamespaceCaching)
		data.ProtocolsCIFSShares[index].NoStrictSecurity = types.BoolValue(record.NoStrictSecurity)
		data.ProtocolsCIFSShares[index].OfflineFiles = types.StringValue(record.OfflineFiles)
		data.ProtocolsCIFSShares[index].Oplocks = types.BoolValue(record.Oplocks)
		data.ProtocolsCIFSShares[index].Path = types.StringValue(record.Path)
		data.ProtocolsCIFSShares[index].ShowSnapshot = types.BoolValue(record.ShowSnapshot)
		data.ProtocolsCIFSShares[index].UnixSymlink = types.StringValue(record.UnixSymlink)
		data.ProtocolsCIFSShares[index].VscanProfile = types.StringValue(record.VscanProfile)

		if len(record.Acls) == 0 {
			setValue, diags := types.SetValue(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"permission":    types.StringType,
					"type":          types.StringType,
					"user_or_group": types.StringType,
				},
			}, []attr.Value{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			data.ProtocolsCIFSShares[index].Acls = setValue
		} else {
			setElements := []attr.Value{}
			for _, acls := range record.Acls {
				log.Printf("acls is: %#v", acls)
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
				data.ProtocolsCIFSShares[index].Acls = setValue
			}
		}
	}
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
