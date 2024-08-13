package security

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SecurityRoleDataSource{}

// NewSecurityRoleDataSource is a helper function to simplify the provider implementation.
func NewSecurityRoleDataSource() datasource.DataSource {
	return &SecurityRoleDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "security_role",
		},
	}
}

// SecurityRoleDataSource defines the data source implementation.
type SecurityRoleDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// SecurityRoleDataSourceModel describes the data source data model.
type SecurityRoleDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Privileges    types.Set    `tfsdk:"privileges"`
	Builtin       types.Bool   `tfsdk:"builtin"`
	Scope         types.String `tfsdk:"scope"`
}

type SecurityRoleOwnerSourceModel struct {
	Name types.Int64  `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

// Metadata returns the data source type name.
func (d *SecurityRoleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *SecurityRoleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityRole data source",

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
				MarkdownDescription: "IPInterface svm name",
				Required:            true,
			},
			"privileges": schema.SetNestedAttribute{
				MarkdownDescription: "The list of privileges that this role has been granted.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path": schema.StringAttribute{
							MarkdownDescription: "Either of REST URI/endpoint OR command/command directory path.",
							Computed:            true,
						},
						"access": schema.StringAttribute{
							MarkdownDescription: "Access level for the REST endpoint or command/command directory path. If it denotes the access level for a command/command directory path, the only supported enum values are 'none','readonly' and 'all'.",
							Computed:            true,
						},
					},
				},
			},
			"builtin": schema.BoolAttribute{
				MarkdownDescription: "Indicates if this is a built-in (pre-defined) role which cannot be modified or deleted.",
				Computed:            true,
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "Scope of the entity. Set to 'cluster' for cluster owned objects and to 'svm' for SVM owned objects.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *SecurityRoleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SecurityRoleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityRoleDataSourceModel

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

	restInfos, err := interfaces.GetSecurityRoles(errorHandler, *client, &interfaces.SecurityRoleDataSourceFilterModel{
		Name: data.Name.ValueString(),
	})

	if err != nil {
		// error reporting done inside GetSecurityRoles
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
		}
		nestedElements := map[string]attr.Value{
			"access": types.StringValue(privilege.Access),
			"path":   types.StringValue(privilege.Path),
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
		},
	}, setElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Privileges = setValue

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
