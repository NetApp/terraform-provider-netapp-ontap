package security

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &SecurityAccountResource{}
var _ resource.ResourceWithImportState = &SecurityAccountResource{}

// NewSecurityAccountResource is a helper function to simplify the provider implementation.
func NewSecurityAccountResource() resource.Resource {
	return &SecurityAccountResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "security_account",
		},
	}
}

// SecurityAccountResource defines the resource implementation.
type SecurityAccountResource struct {
	config connection.ResourceOrDataSourceConfig
}

// SecurityAccountResourceModel describes the resource data model.
type SecurityAccountResourceModel struct {
	CxProfileName              types.String                `tfsdk:"cx_profile_name"`
	Name                       types.String                `tfsdk:"name"`
	ID                         types.String                `tfsdk:"id"`
	OwnerID                    types.String                `tfsdk:"owner_id"`
	Applications               []ApplicationsResourceModel `tfsdk:"applications"`
	Owner                      types.Object                `tfsdk:"owner"`
	Role                       types.Object                `tfsdk:"role"`
	Password                   types.String                `tfsdk:"password"`
	SecondAuthenticationMethod types.String                `tfsdk:"second_authentication_method"`
	Comment                    types.String                `tfsdk:"comment"`
	Locked                     types.Bool                  `tfsdk:"locked"`
}

// ApplicationsResourceModel describes the resource data model.
type ApplicationsResourceModel struct {
	Application                types.String    `tfsdk:"application"`
	SecondAuthentiactionMethod types.String    `tfsdk:"second_authentication_method"`
	AuthenticationMethods      *[]types.String `tfsdk:"authentication_methods"`
}

// OwnerResourceModel describes the resource data model.
type OwnerResourceModel struct {
	Name types.String `tfsdk:"name"`
}

// RoleResourceModel describes the resource data model.
type RoleResourceModel struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the resource type name.
func (r *SecurityAccountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *SecurityAccountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityAccount resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "SecurityAccount name",
				Required:            true,
			},
			"applications": schema.ListNestedAttribute{
				MarkdownDescription: "List of applications",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"application": schema.StringAttribute{
							MarkdownDescription: "Application name",
							Required:            true,
						},
						"second_authentication_method": schema.StringAttribute{
							MarkdownDescription: "Second authentication method",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString("none"),
							PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"authentication_methods": schema.ListAttribute{
							MarkdownDescription: "List of authentication methods",
							Optional:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
			"owner": schema.SingleNestedAttribute{
				MarkdownDescription: "Account owner",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Account owner name",
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"owner_id": schema.StringAttribute{
				MarkdownDescription: "Account owner uuid",
				Computed:            true,
			},
			"role": schema.SingleNestedAttribute{
				MarkdownDescription: "Account role",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Account role name",
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Account password",
				Optional:            true,
				Sensitive:           true,
			},
			"second_authentication_method": schema.StringAttribute{
				MarkdownDescription: "Second authentication method",
				Optional:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Account comment",
				Optional:            true,
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "Account locked",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "SecurityAccount id",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SecurityAccountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SecurityAccountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityAccountResourceModel

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
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))
	var restInfo *interfaces.SecurityAccountGetDataModelONTAP
	if data.Owner.IsUnknown() {
		restInfo, err = interfaces.GetSecurityAccountByName(errorHandler, *client, data.Name.ValueString(), "")
		if err != nil {
			// error reporting done inside GetSecurityAccount
			return
		}
	} else {
		svm, err := interfaces.GetSvmByNameIgnoreNotFound(errorHandler, *client, data.Owner.Attributes()["name"].String())
		if err == nil && svm == nil {
			// reset errorHandler so we don't fail in this case
			restInfo, err = interfaces.GetSecurityAccountByName(errorHandler, *client, data.Name.ValueString(), "")
			if err != nil {
				// error reporting done inside GetSecurityAccount
				return
			}
		} else {
			restInfo, err = interfaces.GetSecurityAccountByName(errorHandler, *client, data.Name.ValueString(), svm.UUID)
			if err != nil {
				// error reporting done inside GetSecurityAccount
				return
			}
		}
	}

	data.Name = types.StringValue(restInfo.Name)
	// There is no ID in the REST response, so we use the name as ID
	data.ID = types.StringValue(restInfo.Name)
	elementTypes := map[string]attr.Type{
		"name": types.StringType,
	}
	elements := map[string]attr.Value{
		"name": types.StringValue(restInfo.Owner.Name),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}

	data.Owner = objectValue
	data.OwnerID = types.StringValue(restInfo.Owner.UUID)
	data.Locked = types.BoolValue(restInfo.Locked)
	if restInfo.Comment != "" {
		data.Comment = types.StringValue(restInfo.Comment)
	}
	elementTypes = map[string]attr.Type{
		"name": types.StringType,
	}
	elements = map[string]attr.Value{
		"name": types.StringValue(restInfo.Role.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Role = objectValue

	data.Applications = make([]ApplicationsResourceModel, len(restInfo.Applications))
	for index, application := range restInfo.Applications {
		data.Applications[index] = ApplicationsResourceModel{
			Application:                types.StringValue(application.Application),
			SecondAuthentiactionMethod: types.StringValue(application.SecondAuthenticationMethod),
		}
		var authenticationMethods []types.String
		for _, authenticationMethod := range application.AuthenticationMethods {
			authenticationMethods = append(authenticationMethods, types.StringValue(authenticationMethod))
		}
		data.Applications[index].AuthenticationMethods = &authenticationMethods
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *SecurityAccountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SecurityAccountResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.SecurityAccountResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	applications := []interfaces.SecurityAccountApplication{}
	for _, item := range data.Applications {
		var application interfaces.SecurityAccountApplication
		application.Application = item.Application.ValueString()
		if item.SecondAuthentiactionMethod.IsNull() {
			application.SecondAuthenticationMethod = item.SecondAuthentiactionMethod.ValueString()
		}
		if item.AuthenticationMethods != nil {
			application.AuthenticationMethods = make([]string, len(*item.AuthenticationMethods))
			for index, authenticationMethod := range *item.AuthenticationMethods {
				application.AuthenticationMethods[index] = authenticationMethod.ValueString()
			}
		}
		applications = append(applications, application)
	}
	err := mapstructure.Decode(applications, &body.Applications)
	if err != nil {
		errorHandler.MakeAndReportError("error creating User applications", fmt.Sprintf("error on encoding copies info: %s, copies %#v", err, body.Applications))
		return
	}
	if !data.Owner.IsUnknown() {
		var owner OwnerResourceModel
		diags := data.Owner.As(ctx, &owner, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.Owner.Name = owner.Name.ValueString()
	}
	if !data.Role.IsUnknown() {
		var role RoleResourceModel
		diags := data.Role.As(ctx, &role, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.Role.Name = role.Name.ValueString()
	}
	if data.Password.IsNull() {
		body.Password = data.Password.ValueString()
	}
	if data.SecondAuthenticationMethod.IsNull() {
		body.SecondAuthenticationMethod = data.SecondAuthenticationMethod.ValueString()
	}
	if data.Comment.IsNull() {
		body.Comment = data.Comment.ValueString()
	}
	if data.Locked.IsNull() {
		body.Locked = data.Locked.ValueBool()
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateSecurityAccount(errorHandler, *client, body)
	if err != nil {
		return
	}
	// As some field are set by the API, we need to read the resource again
	var restInfo *interfaces.SecurityAccountGetDataModelONTAP
	if data.Owner.IsUnknown() {
		restInfo, err = interfaces.GetSecurityAccountByName(errorHandler, *client, data.Name.ValueString(), "")
		if err != nil {
			// error reporting done inside GetSecurityAccount
			return
		}
	} else {
		svm, err := interfaces.GetSvmByNameIgnoreNotFound(errorHandler, *client, data.Owner.Attributes()["name"].String())
		if err == nil && svm == nil {
			// reset errorHandler so we don't fail in this case
			restInfo, err = interfaces.GetSecurityAccountByName(errorHandler, *client, data.Name.ValueString(), "")
			if err != nil {
				// error reporting done inside GetSecurityAccount
				return
			}
		} else {
			restInfo, err = interfaces.GetSecurityAccountByName(errorHandler, *client, data.Name.ValueString(), svm.UUID)
			if err != nil {
				// error reporting done inside GetSecurityAccount
				return
			}
		}
	}
	elementTypes := map[string]attr.Type{
		"name": types.StringType,
	}
	elements := map[string]attr.Value{
		"name": types.StringValue(resource.Owner.Name),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Owner = objectValue

	elementTypes = map[string]attr.Type{
		"name": types.StringType,
	}
	elements = map[string]attr.Value{
		"name": types.StringValue(restInfo.Role.Name),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Role = objectValue

	data.ID = types.StringValue(resource.Name)
	data.OwnerID = types.StringValue(resource.Owner.UUID)

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SecurityAccountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SecurityAccountResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SecurityAccountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SecurityAccountResourceModel

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

	if data.OwnerID.IsNull() {
		errorHandler.MakeAndReportError("Owner UUID is null", "security_account Owner UUID is null")
		return
	}

	err = interfaces.DeleteSecurityAccount(errorHandler, *client, data.Name.ValueString(), data.OwnerID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SecurityAccountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
