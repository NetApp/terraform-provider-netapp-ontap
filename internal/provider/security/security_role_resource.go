package security

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &SecurityRoleResource{}
var _ resource.ResourceWithImportState = &SecurityRoleResource{}

// NewSecurityRoleResource is a helper function to simplify the provider implementation.
func NewSecurityRoleResource() resource.Resource {
	return &SecurityRoleResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "security_role",
		},
	}
}

// SecurityRoleResource defines the resource implementation.
type SecurityRoleResource struct {
	config connection.ResourceOrDataSourceConfig
}

// SecurityRoleResourceModel describes the resource data model.
type SecurityRoleResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	SVMName       types.String `tfsdk:"svm_name"` // if needed or relevant
	Privileges    types.Set    `tfsdk:"privileges"`
	Builtin       types.Bool   `tfsdk:"builtin"`
	Scope         types.String `tfsdk:"scope"`
	ID            types.String `tfsdk:"id"`
}

type SecurityRoleResourcePrivilege struct {
	Path   types.String `tfsdk:"path"`
	Access types.String `tfsdk:"access"`
	Query  types.String `tfsdk:"query"`
}

// Metadata returns the resource type name.
func (r *SecurityRoleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *SecurityRoleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityRole resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "SecurityRole name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "SecurityRole svm name",
				Optional:            true,
			},
			"privileges": schema.SetNestedAttribute{
				MarkdownDescription: "The list of privileges that this role has been granted.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							MarkdownDescription: "Either of REST URI/endpoint OR command/command directory path.",
							Optional:            true,
						},
						"access": schema.StringAttribute{
							MarkdownDescription: "Access level for the REST endpoint or command/command directory path. If it denotes the access level for a command/command directory path, the only supported enum values are 'none','readonly' and 'all'.",
							Optional:            true,
						},
						"query": schema.StringAttribute{
							MarkdownDescription: "Optional attribute that can be specified only if the 'path' attribute refers to a command/command directory path. The privilege tuple implicitly defines a set of objects the role can or cannot access at the specified access level. The query further reduces this set of objects to a subset of objects that the role is allowed to access. The query attribute must be applicable to the command/command directory specified by the 'path' attribute. It is defined using one or more parameters of the command/command directory path specified by the 'path' attribute.",
							Optional:            true,
						},
					},
				},
			},
			"builtin": schema.BoolAttribute{
				MarkdownDescription: "Indicates if this is a built-in (pre-defined) role which cannot be modified or deleted.",
				Computed:            true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "Scope of the entity. Set to 'cluster' for cluster owned objects and to 'svm' for SVM owned objects.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the security role.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SecurityRoleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SecurityRoleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityRoleResourceModel

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

	restInfos, err := interfaces.GetSecurityRoles(errorHandler, *client, &interfaces.SecurityRoleDataSourceFilterModel{
		Name:    data.Name.ValueString(),
		SVMName: data.SVMName.ValueString(),
	})

	if err != nil {
		// error reporting done inside GetSecurityRole
		return
	}

	foundRole := false
	restInfo := interfaces.SecurityRoleGetDataModelONTAP{}
	for _, role := range restInfos {
		if role.Name == data.Name.ValueString() {
			foundRole = true
			restInfo = role
			break
		}
	}
	if !foundRole {
		resp.Diagnostics.AddError("SecurityRole not found", fmt.Sprintf("SecurityRole %s not found", data.Name.ValueString()))
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.Builtin = types.BoolValue(restInfo.Builtin)
	data.Scope = types.StringValue(restInfo.Scope)

	// Priviledges
	setElements := []attr.Value{}
	for _, privilege := range restInfo.Privileges {
		nestedElementTypes := map[string]attr.Type{
			"access": types.StringType,
			"path":   types.StringType,
			"query":  types.StringType,
		}
		nestedElements := map[string]attr.Value{
			"access": types.StringValue(privilege.Access),
			"path":   types.StringValue(privilege.Path),
		}
		if privilege.Query != "" {
			nestedElements["query"] = types.StringValue(privilege.Query)
		} else {
			nestedElements["query"] = basetypes.NewStringNull()
		}
		objectValue, diags := types.ObjectValue(nestedElementTypes, nestedElements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		setElements = append(setElements, objectValue)
	}
	setValue, diags := types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"access": types.StringType,
			"path":   types.StringType,
			"query":  types.StringType,
		},
	}, setElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Privileges = setValue
	data.ID = types.StringValue(restInfo.Owner.Id + "/" + restInfo.Name)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource
func (r *SecurityRoleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SecurityRoleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.SecurityRoleResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	body.Owner.Name = data.SVMName.ValueString()

	PrivilegesList := []interfaces.SecurityRolePrivilegesListBodyDataModelONTAP{}
	if !data.Privileges.IsNull() {
		elements := make([]types.Object, 0, len(data.Privileges.Elements()))
		diags := data.Privileges.ElementsAs(ctx, &elements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		for _, element := range elements {
			var privilege SecurityRoleResourcePrivilege
			diags := element.As(ctx, &privilege, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			interfacesPrivilege := interfaces.SecurityRolePrivilegesListBodyDataModelONTAP{}
			interfacesPrivilege.Path = privilege.Path.ValueString()
			interfacesPrivilege.Access = privilege.Access.ValueString()
			interfacesPrivilege.Query = privilege.Query.ValueString()
			PrivilegesList = append(PrivilegesList, interfacesPrivilege)
		}
		body.Privileges = PrivilegesList
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	_, err = interfaces.CreateSecurityRole(errorHandler, *client, body)
	if err != nil {
		return
	}

	tflog.Trace(ctx, "created a resource")

	restInfos, err := interfaces.GetSecurityRoles(errorHandler, *client, &interfaces.SecurityRoleDataSourceFilterModel{
		Name:    data.Name.ValueString(),
		SVMName: data.SVMName.ValueString(),
	})

	if err != nil {
		// error reporting done inside GetSecurityRole
		return
	}

	foundRole := false
	restInfo := interfaces.SecurityRoleGetDataModelONTAP{}
	for _, role := range restInfos {
		if role.Name == data.Name.ValueString() {
			foundRole = true
			restInfo = role
			break
		}
	}
	if !foundRole {
		resp.Diagnostics.AddError("SecurityRole not found", fmt.Sprintf("SecurityRole %s not found", data.Name.ValueString()))
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.Builtin = types.BoolValue(restInfo.Builtin)
	data.Scope = types.StringValue(restInfo.Scope)

	// Priviledges
	setElements := []attr.Value{}
	for _, privilege := range restInfo.Privileges {
		deleteDefaultPrivileges := false
		if privilege.Path == "DEFAULT" && privilege.Access == "none" && privilege.Query == "" {
			for _, planedPrivilege := range PrivilegesList {
				if planedPrivilege.Path == "DEFAULT" && planedPrivilege.Access == "none" && planedPrivilege.Query == "" {
					deleteDefaultPrivileges = false
					break
				}
				deleteDefaultPrivileges = true
			}
		}
		if deleteDefaultPrivileges {
			err = interfaces.DeleteSecurityRolePrivileges(errorHandler, *client, privilege.Path, data.Name.ValueString(), restInfo.Owner.Id)
			if err != nil {
				errorHandler.MakeAndReportError("error deleting default security_role privileges", "error on DELETE API created default privileges: {path: 'DEFAULT', access: 'none', query: ''}")
				return
			}
			continue
		}
		nestedElementTypes := map[string]attr.Type{
			"access": types.StringType,
			"path":   types.StringType,
			"query":  types.StringType,
		}
		nestedElements := map[string]attr.Value{
			"access": types.StringValue(privilege.Access),
			"path":   types.StringValue(privilege.Path),
		}
		if privilege.Query != "" {
			nestedElements["query"] = types.StringValue(privilege.Query)
		} else {
			nestedElements["query"] = basetypes.NewStringNull()
		}
		objectValue, diags := types.ObjectValue(nestedElementTypes, nestedElements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		setElements = append(setElements, objectValue)
	}
	setValue, diags := types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"access": types.StringType,
			"path":   types.StringType,
			"query":  types.StringType,
		},
	}, setElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Privileges = setValue
	data.ID = types.StringValue(restInfo.Owner.Id + "/" + restInfo.Name)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
// Only the privileges can be updated by the API.
func (r *SecurityRoleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *SecurityRoleResourceModel
	var config *SecurityRoleResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	svm, err := interfaces.GetSvmByName(errorHandler, *client, plan.SVMName.ValueString())
	if err != nil {
		return
	}

	PlanPrivilegesList := []interfaces.SecurityRolePrivilegesBodyDataModelONTAP{}
	ConfigPrivilegesList := []interfaces.SecurityRolePrivilegesBodyDataModelONTAP{}

	if !plan.Privileges.IsNull() {

		elements := make([]types.Object, 0, len(plan.Privileges.Elements()))
		diags := plan.Privileges.ElementsAs(ctx, &elements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		for _, element := range elements {
			var privilege SecurityRoleResourcePrivilege
			diags := element.As(ctx, &privilege, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			log.Printf("privilegePlan : %v", privilege)
			interfacesPrivilege := interfaces.SecurityRolePrivilegesBodyDataModelONTAP{}
			interfacesPrivilege.Path = privilege.Path.ValueString()
			interfacesPrivilege.Access = privilege.Access.ValueString()
			interfacesPrivilege.Query = privilege.Query.ValueString()
			PlanPrivilegesList = append(PlanPrivilegesList, interfacesPrivilege)
		}
	}

	if !config.Privileges.IsNull() {
		elements := make([]types.Object, 0, len(config.Privileges.Elements()))
		diags := config.Privileges.ElementsAs(ctx, &elements, false)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		for _, element := range elements {
			var privilege SecurityRoleResourcePrivilege
			diags := element.As(ctx, &privilege, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			log.Printf("privilegeConfig : %v", privilege)
			interfacesPrivilege := interfaces.SecurityRolePrivilegesBodyDataModelONTAP{}
			interfacesPrivilege.Path = privilege.Path.ValueString()
			interfacesPrivilege.Access = privilege.Access.ValueString()
			interfacesPrivilege.Query = privilege.Query.ValueString()
			ConfigPrivilegesList = append(ConfigPrivilegesList, interfacesPrivilege)
		}
	}
	//Find the difference of paths that are in the plan but not in the config. These are the paths that need to be added. POST on these paths: /security/roles/{owner.uuid}/{name}/privileges
	hasDefaultPathInPlan := false
	for _, planPrivilege := range PlanPrivilegesList {
		foundPathInConfig := false
		if planPrivilege.Path == "DEFAULT" && planPrivilege.Access == "none" && planPrivilege.Query == "" {
			hasDefaultPathInPlan = true
		}
		for _, configPrivilege := range ConfigPrivilegesList {
			if planPrivilege.Path == configPrivilege.Path {
				log.Print("hit true")
				foundPathInConfig = true
				// if Path is the same, but others are not. Do a PATCH on this path
				if planPrivilege.Access != configPrivilege.Access || planPrivilege.Query != configPrivilege.Query {
					err = interfaces.UpdateSecurityRolePrivileges(errorHandler, *client, planPrivilege, plan.Name.ValueString(), svm.UUID)
					if err != nil {
						return
					}
				}
			}
		}
		if !foundPathInConfig {
			//POST on this path
			log.Printf("going to create privilege : %v", planPrivilege)
			err = interfaces.CreateSecurityRolePrivileges(errorHandler, *client, planPrivilege, plan.Name.ValueString(), svm.UUID)
			errorHandler.MakeAndReportError("error deleting default security_role privileges", "error on DELETE API created default privileges: {path: 'DEFAULT', access: 'none', query: ''}")
			if err != nil {
				return
			}
			if !hasDefaultPathInPlan {
				// DELETE on this path
				err = interfaces.DeleteSecurityRolePrivileges(errorHandler, *client, "DEFAULT", plan.Name.ValueString(), svm.UUID)
				if err != nil {
					return
				}
			}

		}
	}

	//Find the difference of paths that are in the config but not in the plan. These are the paths that need to be deleted. DELETE on these paths: /security/roles/{owner.uuid}/{name}/privileges/{path}
	for _, configPrivilege := range ConfigPrivilegesList {
		foundPathInPlan := false
		for _, planPrivilege := range PlanPrivilegesList {
			if planPrivilege.Path == configPrivilege.Path {
				foundPathInPlan = true
			}
		}
		if !foundPathInPlan {
			//DELETE on this path
			err = interfaces.DeleteSecurityRolePrivileges(errorHandler, *client, configPrivilege.Path, plan.Name.ValueString(), svm.UUID)
			if err != nil {
				return
			}
		}
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SecurityRoleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SecurityRoleResourceModel

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

	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		return
	}
	err = interfaces.DeleteSecurityRole(errorHandler, *client, data.Name.ValueString(), svm.UUID)
	if err != nil {
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SecurityRoleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req an security role resource: %#v", req))
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
