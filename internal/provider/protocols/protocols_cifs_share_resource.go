package protocols

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsCIFSShareResource{}
var _ resource.ResourceWithImportState = &ProtocolsCIFSShareResource{}

// NewProtocolsCIFSShareResource is a helper function to simplify the provider implementation.
func NewProtocolsCIFSShareResource() resource.Resource {
	return &ProtocolsCIFSShareResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cifs_share",
		},
	}
}

// NewProtocolsCIFSShareResourceAlias is a helper function to simplify the provider implementation.
func NewProtocolsCIFSShareResourceAlias() resource.Resource {
	return &ProtocolsCIFSShareResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_cifs_share_resource",
		},
	}
}

// ProtocolsCIFSShareResource defines the resource implementation.
type ProtocolsCIFSShareResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsCIFSShareResourceModel describes the resource data model.
type ProtocolsCIFSShareResourceModel struct {
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
	ID                    types.String `tfsdk:"id"`
}

// ProtocolsCIFSShareResourceAcls describes the acls resource data model.
type ProtocolsCIFSShareResourceAcls struct {
	Permission  string `tfsdk:"permission"`
	Type        string `tfsdk:"type"`
	UserOrGroup string `tfsdk:"user_or_group"`
}

// Metadata returns the resource type name.
func (r *ProtocolsCIFSShareResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsCIFSShareResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsCIFSShare resource",

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
				MarkdownDescription: "svm name",
				Required:            true,
			},
			"acls": schema.SetNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The permissions that users and groups have on a CIFS share.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"permission": schema.StringAttribute{
							MarkdownDescription: "Specifies the access rights that a user or group has on the defined CIFS Share.",
							Optional:            true,
							Computed:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("full_control", "read", "change", "no_access"),
							},
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "string Specifies the type of the user or group to add to the access control list of a CIFS share.",
							Optional:            true,
							Computed:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("windows", "unix_user", "unix_group"),
							},
						},
						"user_or_group": schema.StringAttribute{
							MarkdownDescription: "Specifies the user or group name to add to the access control list of a CIFS share.",
							Optional:            true,
							Computed:            true,
						},
					},
				},
			},
			"change_notify": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether CIFS clients can request for change notifications for directories on this share.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Specify the CIFS share descriptions.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"continuously_available": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not the clients connecting to this share can open files in a persistent manner.Files opened in this way are protected from disruptive events, such as, failover and giveback.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"dir_umask": schema.Int64Attribute{
				MarkdownDescription: "Directory Mode Creation Mask to be viewed as an octal number.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					IntUseStateForUnknown(),
				},
			},
			"encryption": schema.BoolAttribute{
				MarkdownDescription: "Specifies that SMB encryption must be used when accessing this share. Clients that do not support encryption are not able to access this share.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"file_umask": schema.Int64Attribute{
				MarkdownDescription: "File Mode Creation Mask to be viewed as an octal number.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					IntUseStateForUnknown(),
				},
			},
			"force_group_for_create": schema.StringAttribute{
				MarkdownDescription: `Specifies that all files that CIFS users create in a specific share belong to the same group
				(also called the force-group). The force-group must be a predefined group in the UNIX group
				database. This setting has no effect unless the security style of the volume is UNIX or mixed
				security style.`,
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace_caching": schema.BoolAttribute{
				MarkdownDescription: `Specifies whether or not the SMB clients connecting to this share can cache the directory enumeration
				results returned by the CIFS servers.`,
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"no_strict_security": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether or not CIFS clients can follow a unix symlinks outside the share boundaries.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"offline_files": schema.StringAttribute{
				MarkdownDescription: `Offline Files. The supported values are:
				none - Clients are not permitted to cache files for offline access.
				manual - Clients may cache files that are explicitly selected by the user for offline access.
				documents - Clients may automatically cache files that are used by the user for offline access.
				programs - Clients may automatically cache files that are used by the user for offline access
				and may use those files in an offline mode even if the share is available.
				`,
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("none", "manual", "documents", "programs"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"oplocks": schema.BoolAttribute{
				MarkdownDescription: `Specify whether opportunistic locks are enabled on this share. "Oplocks" allow clients to lock files and cache content locally,
				which can increase performance for file operations.
				`,
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"path": schema.StringAttribute{
				MarkdownDescription: `The fully-qualified pathname in the owning SVM namespace that is shared through this share.
				If this is a home directory share then the path should be dynamic by specifying the pattern
				%w (Windows user name), %u (UNIX user name), or %d (domain name) variables in any combination.
				ONTAP generates the path dynamically for the connected user and this path is appended to each
				search path to find the full Home Directory path.
				`,
				Required: true,
			},
			"show_snapshot": schema.BoolAttribute{
				MarkdownDescription: `Specifies whether or not the Snapshot copies can be viewed and traversed by clients.`,
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"unix_symlink": schema.StringAttribute{
				MarkdownDescription: `Controls the access of UNIX symbolic links to CIFS clients.
				The supported values are:
				* local - Enables only local symbolic links which is within the same CIFS share.
				* widelink - Enables both local symlinks and widelinks.
				* disable - Disables local symlinks and widelinks.`,
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("local", "widelink", "disable"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vscan_profile": schema.StringAttribute{
				MarkdownDescription: ` Vscan File-Operations Profile
				The supported values are:
				no_scan - Virus scans are never triggered for accesses to this share.
				standard - Virus scans can be triggered by open, close, and rename operations.
				strict - Virus scans can be triggered by open, read, close, and rename operations.
				writes_only - Virus scans can be triggered only when a file that has been modified is closed.`,
				Optional: true,
				Computed: true,
				Validators: []validator.String{
					stringvalidator.OneOf("no_scan", "standard", "strict", "writes_only"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The UUID of the CIFS share.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsCIFSShareResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ProtocolsCIFSShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsCIFSShareResourceModel

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

	restInfo, err := interfaces.GetProtocolsCIFSShareByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetProtocolsCIFSShare
		return
	}

	data.ID = types.StringValue(restInfo.Name)
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
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *ProtocolsCIFSShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProtocolsCIFSShareResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ProtocolsCIFSShareResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()
	body.Path = data.Path.ValueString()

	configHasDefaultACL := false
	if !data.Acls.IsUnknown() {
		aclsList := []interfaces.Acls{}
		elements := make([]types.Object, 0, len(data.Acls.Elements()))
		diags := data.Acls.ElementsAs(ctx, &elements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		for _, element := range elements {
			var acls ProtocolsCIFSShareResourceAcls
			diags := element.As(ctx, &acls, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			interfacesAcls := interfaces.Acls{}
			interfacesAcls.Permission = acls.Permission
			interfacesAcls.Type = acls.Type
			interfacesAcls.UserOrGroup = acls.UserOrGroup
			if acls.UserOrGroup == "Everyone" && acls.Permission == "full_control" {
				configHasDefaultACL = true
			}
			aclsList = append(aclsList, interfacesAcls)
		}
		body.Acls = aclsList

	}

	if !data.ChangeNotify.IsUnknown() {
		body.ChangeNotify = data.ChangeNotify.ValueBool()
	}
	if !data.Comment.IsUnknown() {
		body.Comment = data.Comment.ValueString()
	}
	if !data.ContinuouslyAvailable.IsUnknown() {
		body.ContinuouslyAvailable = data.ContinuouslyAvailable.ValueBool()
	}
	if !data.DirUmask.IsUnknown() {
		body.DirUmask = data.DirUmask.ValueInt64()
	}
	if !data.Encryption.IsUnknown() {
		body.Encryption = data.Encryption.ValueBool()
	}
	if !data.FileUmask.IsUnknown() {
		body.FileUmask = data.FileUmask.ValueInt64()
	}
	if !data.ForceGroupForCreate.IsUnknown() {
		body.ForceGroupForCreate = data.ForceGroupForCreate.ValueString()
	}
	if !data.HomeDirectory.IsUnknown() {
		body.HomeDirectory = data.HomeDirectory.ValueBool()
	}
	if !data.NamespaceCaching.IsUnknown() {
		body.NamespaceCaching = data.NamespaceCaching.ValueBool()
	}
	if !data.NoStrictSecurity.IsUnknown() {
		body.NoStrictSecurity = data.NoStrictSecurity.ValueBool()
	}
	if !data.OfflineFiles.IsUnknown() {
		body.OfflineFiles = data.OfflineFiles.ValueString()
	}
	if !data.Oplocks.IsUnknown() {
		body.Oplocks = data.Oplocks.ValueBool()
	}
	if !data.ShowSnapshot.IsUnknown() {
		body.ShowSnapshot = data.ShowSnapshot.ValueBool()
	}
	if !data.UnixSymlink.IsUnknown() {
		body.UnixSymlink = data.UnixSymlink.ValueString()
	}
	if !data.VscanProfile.IsUnknown() {
		body.VscanProfile = data.VscanProfile.ValueString()
	}

	_, err = interfaces.CreateProtocolsCIFSShare(errorHandler, *client, body)
	if err != nil {
		return
	}

	// The POST API returns record which does not contains the info of Acls, so we have to do another GET call to get the Acls info.
	restInfo, err := interfaces.GetProtocolsCIFSShareByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetProtocolsCIFSShare
		return
	}

	data.ID = types.StringValue(restInfo.Name)

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
	if restInfo.Acls == nil {
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
		data.Acls = setValue
	} else {
		for _, acls := range restInfo.Acls {
			//If the config file does not have acl set user_or_group as "Everyone / Full Control", the API will create one by default. Need to delete it if user does not want one.
			if acls.UserOrGroup == "Everyone" && acls.Permission == "full_control" && !configHasDefaultACL {
				svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
				if err != nil {
					return
				}
				err = interfaces.DeleteProtocolsCIFSShareACL(errorHandler, *client, svm.UUID, data.Name.ValueString(), acls.UserOrGroup, acls.Type)
				if err != nil {
					// error reporting done inside DeleteProtocolsCIFSShareAcl
					return
				}
				continue
			}

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
	}

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsCIFSShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *ProtocolsCIFSShareResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	var body interfaces.ProtocolsCIFSShareResourceBodyDataModelONTAP

	if !plan.ChangeNotify.IsUnknown() {
		if plan.ChangeNotify != state.ChangeNotify {
			body.ChangeNotify = plan.ChangeNotify.ValueBool()
		}
	}
	if !plan.Comment.IsUnknown() {
		if plan.Comment != state.Comment {
			body.Comment = plan.Comment.ValueString()
		}
	}
	if !plan.ContinuouslyAvailable.IsUnknown() {
		if plan.ContinuouslyAvailable != state.ContinuouslyAvailable {
			body.ContinuouslyAvailable = plan.ContinuouslyAvailable.ValueBool()
		}
	}

	if !plan.DirUmask.IsUnknown() {
		if plan.DirUmask != state.DirUmask {
			body.DirUmask = plan.DirUmask.ValueInt64()
		}
	}
	if !plan.Encryption.IsUnknown() {
		if plan.Encryption != state.Encryption {
			body.Encryption = plan.Encryption.ValueBool()
		}
	}
	if !plan.FileUmask.IsUnknown() {
		if plan.FileUmask != state.FileUmask {
			body.FileUmask = plan.FileUmask.ValueInt64()
		}
	}

	if !plan.ForceGroupForCreate.IsUnknown() {
		if plan.ForceGroupForCreate != state.ForceGroupForCreate {
			body.ForceGroupForCreate = plan.ForceGroupForCreate.ValueString()
		}
	}

	if !plan.NamespaceCaching.IsUnknown() {
		if plan.NamespaceCaching != state.NamespaceCaching {
			body.NamespaceCaching = plan.NamespaceCaching.ValueBool()
		}
	}

	if !plan.NoStrictSecurity.IsUnknown() {
		if plan.NoStrictSecurity != state.NoStrictSecurity {
			body.NoStrictSecurity = plan.NoStrictSecurity.ValueBool()
		}
	}

	if !plan.OfflineFiles.IsUnknown() {
		if plan.OfflineFiles != state.OfflineFiles {
			body.OfflineFiles = plan.OfflineFiles.ValueString()
		}
	}

	if !plan.Oplocks.IsUnknown() {
		if plan.Oplocks != state.Oplocks {
			body.Oplocks = plan.Oplocks.ValueBool()
		}
	}

	if !plan.ShowSnapshot.IsUnknown() {
		if plan.ShowSnapshot != state.ShowSnapshot {
			body.ShowSnapshot = plan.ShowSnapshot.ValueBool()
		}
	}

	if !plan.UnixSymlink.IsUnknown() {
		if plan.UnixSymlink != state.UnixSymlink {
			body.UnixSymlink = plan.UnixSymlink.ValueString()
		}
	}

	if !plan.VscanProfile.IsUnknown() {
		if plan.VscanProfile != state.VscanProfile {
			body.VscanProfile = plan.VscanProfile.ValueString()
		}
	}

	if !plan.ContinuouslyAvailable.IsUnknown() {
		if plan.ContinuouslyAvailable != state.ContinuouslyAvailable {
			body.ContinuouslyAvailable = plan.ContinuouslyAvailable.ValueBool()
		}
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, plan.SVMName.ValueString())
	if err != nil {
		return
	}

	// have no luck updating acls using PATCH cifs/shares API sucessfully, so we have to use the acls set of API.
	if !plan.Acls.IsUnknown() {
		// reading acls from plan
		planeAcls := make([]types.Object, 0, len(plan.Acls.Elements()))
		diags := plan.Acls.ElementsAs(ctx, &planeAcls, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		// reading acls from state
		stateAcls := make([]types.Object, 0, len(state.Acls.Elements()))
		diags = state.Acls.ElementsAs(ctx, &stateAcls, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		// iterate over plan acls and compare with state acls. Make create or update or delete calls accordingly.
		for _, element := range stateAcls {
			var stateACLElement ProtocolsCIFSShareResourceAcls
			diags := element.As(ctx, &stateACLElement, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			for index, planACL := range planeAcls {
				var planACLElement ProtocolsCIFSShareResourceAcls
				diags := planACL.As(ctx, &planACLElement, basetypes.ObjectAsOptions{})
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
				}
				// if 'userOrGroup' and 'type' matches, then we know it's not a create action. If permission is same, then break the loop because nothing to update.
				if stateACLElement.UserOrGroup == planACLElement.UserOrGroup && stateACLElement.Type == planACLElement.Type {
					if stateACLElement.Permission == planACLElement.Permission {
						break
					} else {
						// update the acls since permission is different
						interfacesAcls := interfaces.ProtocolsCIFSShareACLResourceBodyDataModelONTAP{}
						interfacesAcls.Permission = planACLElement.Permission
						err = interfaces.UpdateProtocolsCIFSShareACL(errorHandler, *client, interfacesAcls, svm.UUID, plan.Name.ValueString(), planACLElement.UserOrGroup, planACLElement.Type)
						if err != nil {
							return
						}
						break
					}
				}
				// if we reach the end of stateAcls, then we know it's a delete action because it was not found in plan acls.
				if index == len(planeAcls)-1 {
					err = interfaces.DeleteProtocolsCIFSShareACL(errorHandler, *client, svm.UUID, plan.Name.ValueString(), stateACLElement.UserOrGroup, stateACLElement.Type)
					if err != nil {
						return
					}

				}
			}

		}
		// now handle create action
		for _, planACL := range planeAcls {
			var planACLElement ProtocolsCIFSShareResourceAcls
			diags := planACL.As(ctx, &planACLElement, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}

			for index, element := range stateAcls {
				var stateACLElement ProtocolsCIFSShareResourceAcls
				diags := element.As(ctx, &stateACLElement, basetypes.ObjectAsOptions{})
				if diags.HasError() {
					resp.Diagnostics.Append(diags...)
					return
				}
				if stateACLElement.UserOrGroup == planACLElement.UserOrGroup && stateACLElement.Type == planACLElement.Type {
					if stateACLElement.Permission == planACLElement.Permission {
						break
					} else {
						// update is already handled by above logic, so break
						break
					}
				}
				// if we reach the end of planAcls, then we know it's a create action because it was not found in state acls.
				if index == len(stateAcls)-1 {
					interfacesAcls := interfaces.ProtocolsCIFSShareACLResourceBodyDataModelONTAP{}
					interfacesAcls.Permission = planACLElement.Permission
					interfacesAcls.Type = planACLElement.Type
					interfacesAcls.UserOrGroup = planACLElement.UserOrGroup
					_, err = interfaces.CreateProtocolsCIFSShareACL(errorHandler, *client, interfacesAcls, svm.UUID, plan.Name.ValueString())
					if err != nil {
						return
					}
				}
			}

		}

	}

	err = interfaces.UpdateProtocolsCIFSShare(errorHandler, *client, body, plan.Name.ValueString(), svm.UUID)
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsCIFSShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsCIFSShareResourceModel

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

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "protocols_cifs_share UUID is null")
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}

	err = interfaces.DeleteProtocolsCIFSShare(errorHandler, *client, data.Name.ValueString(), svm.UUID)
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsCIFSShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)

	tflog.Debug(ctx, fmt.Sprintf("import req an protocols cifs share resource: %#v", req))
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprint("Expected ID in the format 'name,svm_name,cx_profile_name', got: ", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
}
