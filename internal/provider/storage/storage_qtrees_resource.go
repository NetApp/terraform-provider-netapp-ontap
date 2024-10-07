package storage

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &StorageQtreeResource{}
var _ resource.ResourceWithImportState = &StorageQtreeResource{}

// NewStorageQtreeResource is a helper function to simplify the provider implementation.
func NewStorageQtreeResource() resource.Resource {
	return &StorageQtreeResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "storage_qtrees",
		},
	}
}

// StorageQtreeResource defines the resource implementation.
type StorageQtreeResource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageQtreeResourceModel describes the resource data model.
type StorageQtreeResourceModel struct {
	CxProfileName   types.String `tfsdk:"cx_profile_name"`
	Name            types.String `tfsdk:"name"`
	SVMName         types.String `tfsdk:"svm_name"` // if needed or relevant
	ID              types.Int64  `tfsdk:"id"`
	SecurityStyle   types.String `tfsdk:"security_style"`
	NAS             types.Object `tfsdk:"nas"`
	User            types.Object `tfsdk:"user"`
	Group           types.Object `tfsdk:"group"`
	Volume          types.String `tfsdk:"volume_name"`
	ExportPolicy    types.Object `tfsdk:"export_policy"`
	UnixPermissions types.Int64  `tfsdk:"unix_permissions"`
}

type StorageQtreeResourceNASModel struct {
	Path types.String `tfsdk:"path"`
}

type StorageQtreeResourceNameModel struct {
	Name types.String `tfsdk:"name"`
}

type StorageQtreeResourceExportPolicyModel struct {
	Name types.String `tfsdk:"name"`
	ID   types.Int64  `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *StorageQtreeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *StorageQtreeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageQtree resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "StorageQtree name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "StorageQtree svm name",
				Optional:            true,
			},
			"id": schema.Int64Attribute{
				MarkdownDescription: "StorageQtree UUID",
				Computed:            true,
				Optional:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"security_style": schema.StringAttribute{
				MarkdownDescription: "StorageQtree security style",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("unix", "ntfs", "mixed"),
				},
			},
			"nas": schema.SingleNestedAttribute{
				MarkdownDescription: "NAS settings",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						MarkdownDescription: "Client visible path to the qtree. This field is not available if the volume does not have a junction-path configured.",
						Computed:            true,
					},
				},
			},
			"user": schema.SingleNestedAttribute{
				MarkdownDescription: "The user set as owner of the qtree.",
				Computed:            true,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Alphanumeric username of user who owns the qtree.",
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"group": schema.SingleNestedAttribute{
				MarkdownDescription: "The user set as owner of the qtree.",
				Computed:            true,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Alphanumeric group name of group that owns the qtree.",
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"volume_name": schema.StringAttribute{
				MarkdownDescription: "The volume that contains the qtree.",
				Required:            true,
			},
			"export_policy": schema.SingleNestedAttribute{
				MarkdownDescription: "The export policy that controls access to the qtree.",
				Computed:            true,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "The name of the export policy.",
						Computed:            true,
						Optional:            true,
					},
					"id": schema.Int64Attribute{
						MarkdownDescription: "The UUID of the export policy.",
						Computed:            true,
						Optional:            true,
					},
				},
			},
			"unix_permissions": schema.Int64Attribute{
				MarkdownDescription: "The UNIX permissions for the qtree.",
				Optional:            true,
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *StorageQtreeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *StorageQtreeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StorageQtreeResourceModel

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

	restInfo, err := interfaces.GetStorageQtreeByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString(), data.Volume.ValueString())
	if err != nil {
		// error reporting done inside GetStorageQtree
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.SecurityStyle = types.StringValue(restInfo.SecurityStyle)
	data.Volume = types.StringValue(restInfo.Volume.Name)
	data.ID = types.Int64Value(int64(restInfo.ID))
	data.UnixPermissions = types.Int64Value(int64(restInfo.UnixPermissions))

	// NAS
	elementTypes := map[string]attr.Type{
		"path": types.StringType,
	}
	elements := map[string]attr.Value{
		"path": types.StringValue(restInfo.NAS.Path),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.NAS = objectValue

	//User
	elementTypes = map[string]attr.Type{
		"name": types.StringType,
	}
	elements = map[string]attr.Value{
		"name": types.StringValue(restInfo.User.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.User = objectValue

	// Group
	elements = map[string]attr.Value{
		"name": types.StringValue(restInfo.Group.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Group = objectValue

	// Export Policy
	elementTypes = map[string]attr.Type{
		"name": types.StringType,
		"id":   types.Int64Type,
	}
	elements = map[string]attr.Value{
		"name": types.StringValue(restInfo.ExportPolicy.Name),
		"id":   types.Int64Value(restInfo.ExportPolicy.ID),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.ExportPolicy = objectValue

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *StorageQtreeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageQtreeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.StorageQtreeResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}

	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()
	body.Volume.Name = data.Volume.ValueString()

	if !data.ID.IsUnknown() {
		body.ID = int(data.ID.ValueInt64())
	}
	if !data.SecurityStyle.IsUnknown() {
		body.SecurityStyle = data.SecurityStyle.ValueString()
	}
	if !data.UnixPermissions.IsUnknown() {
		body.UnixPermissions = int(data.UnixPermissions.ValueInt64())
	}

	if !data.User.IsUnknown() {
		var user StorageQtreeResourceNameModel
		diags := data.User.As(ctx, &user, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !user.Name.IsUnknown() {
			if interfaces.GetUnixUserByName(errorHandler, *client, svm.UUID, user.Name.ValueString()) == nil {
				body.User.Name = user.Name.ValueString()
			} else {
				return
			}
		}
	}

	if !data.Group.IsUnknown() {
		var group StorageQtreeResourceNameModel
		diags := data.Group.As(ctx, &group, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !group.Name.IsUnknown() {
			if interfaces.GetUnixGroupByName(errorHandler, *client, svm.UUID, group.Name.ValueString()) == nil {
				body.Group.Name = group.Name.ValueString()
			} else {
				return
			}
		}
	}

	if !data.ExportPolicy.IsUnknown() {
		var exportPolicy StorageQtreeResourceExportPolicyModel
		diags := data.ExportPolicy.As(ctx, &exportPolicy, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !exportPolicy.Name.IsUnknown() {
			body.ExportPolicy.Name = exportPolicy.Name.ValueString()
		}
		if !exportPolicy.ID.IsUnknown() {
			body.ExportPolicy.ID = exportPolicy.ID.ValueInt64()
		}
	}

	_, err = interfaces.CreateStorageQtree(errorHandler, *client, body)
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetStorageQtreeByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString(), data.Volume.ValueString())
	if err != nil {
		// error reporting done inside GetStorageQtree
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.SecurityStyle = types.StringValue(restInfo.SecurityStyle)
	data.Volume = types.StringValue(restInfo.Volume.Name)
	data.ID = types.Int64Value(int64(restInfo.ID))
	data.UnixPermissions = types.Int64Value(int64(restInfo.UnixPermissions))

	// NAS
	elementTypes := map[string]attr.Type{
		"path": types.StringType,
	}
	elements := map[string]attr.Value{
		"path": types.StringValue(restInfo.NAS.Path),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.NAS = objectValue

	//User
	elementTypes = map[string]attr.Type{
		"name": types.StringType,
	}
	elements = map[string]attr.Value{
		"name": types.StringValue(restInfo.User.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.User = objectValue

	// Group
	elements = map[string]attr.Value{
		"name": types.StringValue(restInfo.Group.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Group = objectValue

	// Export Policy
	elementTypes = map[string]attr.Type{
		"name": types.StringType,
		"id":   types.Int64Type,
	}

	elements = map[string]attr.Value{
		"name": types.StringValue(restInfo.ExportPolicy.Name),
		"id":   types.Int64Value(restInfo.ExportPolicy.ID),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.ExportPolicy = objectValue

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *StorageQtreeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *StorageQtreeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	vol, err := interfaces.GetStorageVolumeByName(errorHandler, *client, data.Volume.ValueString(), data.SVMName.ValueString())
	if err != nil {
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}

	var body interfaces.StorageQtreeResourceBodyDataModelONTAP
	body.Name = data.Name.ValueString()
	body.SecurityStyle = data.SecurityStyle.ValueString()
	body.UnixPermissions = int(data.UnixPermissions.ValueInt64())

	// user
	if !data.User.IsUnknown() {
		var user StorageQtreeResourceNameModel
		diags := data.User.As(ctx, &user, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !user.Name.IsUnknown() && !user.Name.IsNull() {
			if interfaces.GetUnixUserByName(errorHandler, *client, svm.UUID, user.Name.ValueString()) == nil {
				body.User.Name = user.Name.ValueString()
			} else {
				return
			}
		}
	}

	// group
	if !data.Group.IsUnknown() {
		var group StorageQtreeResourceNameModel
		diags := data.Group.As(ctx, &group, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !group.Name.IsUnknown() && !group.Name.IsNull() {
			if interfaces.GetUnixGroupByName(errorHandler, *client, svm.UUID, group.Name.ValueString()) == nil {
				body.Group.Name = group.Name.ValueString()
			} else {
				return
			}
		}
	}

	// export policy
	if !data.ExportPolicy.IsUnknown() {
		var exportPolicy StorageQtreeResourceExportPolicyModel
		diags := data.ExportPolicy.As(ctx, &exportPolicy, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !exportPolicy.Name.IsUnknown() && !exportPolicy.Name.IsNull() {
			body.ExportPolicy.Name = exportPolicy.Name.ValueString()
		}
		if !exportPolicy.ID.IsUnknown() && !exportPolicy.ID.IsNull() {
			body.ExportPolicy.ID = exportPolicy.ID.ValueInt64()
		}
	}

	err = interfaces.UpdateStorageQtree(errorHandler, *client, body, vol.UUID, strconv.Itoa(int(data.ID.ValueInt64())))
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetStorageQtreeByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString(), data.Volume.ValueString())
	if err != nil {
		// error reporting done inside GetStorageQtree
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.SecurityStyle = types.StringValue(restInfo.SecurityStyle)
	data.Volume = types.StringValue(restInfo.Volume.Name)
	data.ID = types.Int64Value(int64(restInfo.ID))
	data.UnixPermissions = types.Int64Value(int64(restInfo.UnixPermissions))

	// NAS
	elementTypes := map[string]attr.Type{
		"path": types.StringType,
	}
	elements := map[string]attr.Value{
		"path": types.StringValue(restInfo.NAS.Path),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.NAS = objectValue

	//User
	elementTypes = map[string]attr.Type{
		"name": types.StringType,
	}
	elements = map[string]attr.Value{
		"name": types.StringValue(restInfo.User.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.User = objectValue

	// Group
	elements = map[string]attr.Value{
		"name": types.StringValue(restInfo.Group.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Group = objectValue

	// Export Policy
	elementTypes = map[string]attr.Type{
		"name": types.StringType,
		"id":   types.Int64Type,
	}

	elements = map[string]attr.Value{
		"name": types.StringValue(restInfo.ExportPolicy.Name),
		"id":   types.Int64Value(restInfo.ExportPolicy.ID),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.ExportPolicy = objectValue

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *StorageQtreeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *StorageQtreeResourceModel

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
		errorHandler.MakeAndReportError("UUID is null", "storage_qtree UUID is null")
		return
	}

	vol, err := interfaces.GetStorageVolumeByName(errorHandler, *client, data.Volume.ValueString(), data.SVMName.ValueString())
	if err != nil {
		return
	}
	id := strconv.Itoa(int(data.ID.ValueInt64()))
	err = interfaces.DeleteStorageQtree(errorHandler, *client, vol.UUID, id)
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *StorageQtreeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,volume_name,svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("volume_name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[3])...)
}
