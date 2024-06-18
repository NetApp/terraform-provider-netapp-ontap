package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/svm"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &StorageQuotaRulesResource{}
var _ resource.ResourceWithImportState = &StorageQuotaRulesResource{}

// NewStorageQuotaRulesResource is a helper function to simplify the provider implementation.
func NewStorageQuotaRulesResource() resource.Resource {
	return &StorageQuotaRulesResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "quota_rules",
		},
	}
}

// StorageQuotaRulesResource defines the resource implementation.
type StorageQuotaRulesResource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageQuotaRulesResourceModel describes the resource data model.
type StorageQuotaRulesResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVM           svm.SVM      `tfsdk:"svm"`
	Volume        Volume       `tfsdk:"volume"`
	Users         *[]User      `tfsdk:"users"`
	Group         types.Object `tfsdk:"group"`
	Qtree         *Qtree       `tfsdk:"qtree"`
	Type          types.String `tfsdk:"type"`
	Files         types.Object `tfsdk:"files"`
	ID            types.String `tfsdk:"id"`
}

// Volume describes Volume data model.
type Volume struct {
	Name types.String `tfsdk:"name"`
}

// User describes User data model.
type User struct {
	Name types.String `tfsdk:"name"`
}

// Group describes Group data model.
type Group struct {
	Name types.String `tfsdk:"name"`
}

// Qtree describes Qtree data model.
type Qtree struct {
	Name types.String `tfsdk:"name"`
}

// Files describes Files data model.
type Files struct {
	HardLimit types.Int64 `tfsdk:"hard_limit"`
	SoftLimit types.Int64 `tfsdk:"soft_limit"`
}

// Metadata returns the resource type name.
func (r *StorageQuotaRulesResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *StorageQuotaRulesResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageQuotaRules resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Quota type for the rule. This type can be user, group, or tree",
				Required:            true,
			},
			"svm": schema.SingleNestedAttribute{
				MarkdownDescription: "Existing SVM in which to create the qtree",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "name of the SVM",
						Required:            true,
					},
				},
			},
			"volume": schema.SingleNestedAttribute{
				MarkdownDescription: "Existing volume in which to create the qtree",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "name of the volume",
						Required:            true,
					},
				},
			},
			"users": schema.SetNestedAttribute{
				MarkdownDescription: "If the quota type is user, this property takes the user name. For default user quota rules, the user name must be specified as \"\"",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "name of the user",
							Optional:            true,
						},
					},
				},
			},
			"group": schema.SingleNestedAttribute{
				MarkdownDescription: "If the quota type is group, this property takes the group name. For default group quota rules, the group name must be specified as \"\"",
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "name of the group",
						Required:            true,
					},
				},
			},
			"qtree": schema.SingleNestedAttribute{
				MarkdownDescription: "Qtree for which to create the rule. For default tree rules, the qtree name must be specified as \"\"",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "name of the qtree",
						Required:            true,
					},
				},
			},
			"files": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"hard_limit": schema.Int64Attribute{
						MarkdownDescription: "Specifies the hard limit for files",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"soft_limit": schema.Int64Attribute{
						MarkdownDescription: "Specifies the soft limit for files",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *StorageQuotaRulesResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *StorageQuotaRulesResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data StorageQuotaRulesResourceModel

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

	var restInfo *interfaces.StorageQuotaRulesGetDataModelONTAP
	if data.ID.ValueString() != "" {
		restInfo, err = interfaces.GetStorageQuotaRulesByUUID(errorHandler, *client, data.ID.ValueString())
		if err != nil {
			// error reporting done inside GetStorageQuotaRulesByUUID
			return
		}
	} else {
		restInfo, err = interfaces.GetStorageQuotaRules(errorHandler, *client, data.Volume.Name.ValueString(), data.SVM.Name.ValueString(), data.Type.ValueString(), data.Qtree.Name.ValueString())
		if err != nil {
			// error reporting done inside GetStorageQuotaRules
			return
		}
		data.Type = types.StringValue(restInfo.Type)
		data.SVM.Name = types.StringValue(restInfo.SVM.Name)
		data.Volume.Name = types.StringValue(restInfo.Volume.Name)
		data.Qtree.Name = types.StringValue(restInfo.Qtree.Name)
	}

	// Files
	elementTypes := map[string]attr.Type{
		"hard_limit": types.Int64Type,
		"soft_limit": types.Int64Type,
	}
	elements := map[string]attr.Value{
		"hard_limit": types.Int64Value(restInfo.Files.HardLimit),
		"soft_limit": types.Int64Value(restInfo.Files.SoftLimit),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Files = objectValue
	data.ID = types.StringValue(restInfo.UUID)

	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *StorageQuotaRulesResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageQuotaRulesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.StorageQuotaRulesResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Type = data.Type.ValueString()
	body.SVM.Name = data.SVM.Name.ValueString()
	body.Volume.Name = data.Volume.Name.ValueString()
	body.Qtree.Name = data.Qtree.Name.ValueString()

	if data.Users != nil {
		for _, user := range *data.Users {
			body.Users = append(body.Users, user.Name.ValueString())
		}
	}

	if !data.Group.IsNull() {
		var group Group
		diags := data.Group.As(ctx, &group, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !group.Name.IsUnknown() {
			body.Group.Name = group.Name.ValueString()
		}
	}
	if !data.Files.IsNull() {
		var files Files
		diags := data.Files.As(ctx, &files, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !files.HardLimit.IsUnknown() {
			body.Files.HardLimit = files.HardLimit.ValueInt64()
		}
		if !files.SoftLimit.IsUnknown() {
			body.Files.SoftLimit = files.SoftLimit.ValueInt64()
		}
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateStorageQuotaRules(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.ID = types.StringValue(resource.UUID)

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *StorageQuotaRulesResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *StorageQuotaRulesResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}
	client, err := connection.GetRestClient(utils.NewErrorHandler(ctx, &resp.Diagnostics), r.config, plan.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	var request interfaces.StorageQuotaRulesResourceBodyUpdateModelONTAP
	if !plan.Files.IsUnknown() {
		if !plan.Files.Equal(state.Files) {
			var files Files
			diags := plan.Files.As(ctx, &files, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			if !files.HardLimit.IsUnknown() {
				request.Files.HardLimit = files.HardLimit.ValueInt64()
			}
			if !files.SoftLimit.IsUnknown() {
				request.Files.SoftLimit = files.SoftLimit.ValueInt64()
			}
		}
	}

	err = interfaces.UpdateQuotaRules(errorHandler, *client, state.ID.ValueString(), request)
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *StorageQuotaRulesResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *StorageQuotaRulesResourceModel

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
		errorHandler.MakeAndReportError("UUID is null", "storage_quota_rules UUID is null")
		return
	}

	err = interfaces.DeleteStorageQuotaRules(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *StorageQuotaRulesResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 5 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" || idParts[4] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: volume_name,svm_name,type,qtree,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("volume").AtName("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm").AtName("name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), idParts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("qtree").AtName("name"), idParts[3])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[4])...)
}
