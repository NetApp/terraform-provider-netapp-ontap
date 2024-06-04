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

// TODO:
// copy this file to match you data source (should match internal/provider/security_role_data_source.go)
// replace SecurityRules with the name of the resource, following go conventions, eg IPInterfaces
// replace security_roles with the name of the resource, for logging purposes, eg ip_interfaces
// make sure to create internal/interfaces/security_role.go too)
// delete these 5 lines

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SecurityRolesDataSource{}

// NewSecurityRolesDataSource is a helper function to simplify the provider implementation.
func NewSecurityRolesDataSource() datasource.DataSource {
	return &SecurityRolesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "security_roles_data_source",
		},
	}
}

// SecurityRulesDataSource defines the data source implementation.
type SecurityRolesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// SecurityRolesDataSourceModel describes the data source data model.
type SecurityRolesDataSourceModel struct {
	CxProfileName types.String                        `tfsdk:"cx_profile_name"`
	SecurityRules []SecurityRoleDataSourceModel       `tfsdk:"security_roles"`
	Filter        *SecurityRulesDataSourceFilterModel `tfsdk:"filter"`
}

// SecurityRulesDataSourceFilterModel describes the data source data model for queries.
type SecurityRulesDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm_name"`
	Scope   types.String `tfsdk:"scope"`
}

// Metadata returns the data source type name.
func (d *SecurityRolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *SecurityRolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityRules data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "SecurityRule name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "SecurityRule svm name",
						Optional:            true,
					},
					"scope": schema.StringAttribute{
						MarkdownDescription: "Scope of the entity. Set to 'cluster' for cluster owned objects and to 'svm' for SVM owned objects.",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"security_roles": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "SecurityRule name",
							Required:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "IPInterface svm name",
							Required:            true,
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
								},
							},
						},
						"builtin": schema.BoolAttribute{
							MarkdownDescription: "Indicates if this is a built-in (pre-defined) role which cannot be modified or deleted.",
							Optional:            true,
						},
						"scope": schema.StringAttribute{
							MarkdownDescription: "Scope of the entity. Set to 'cluster' for cluster owned objects and to 'svm' for SVM owned objects.",
							Optional:            true,
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
func (d *SecurityRolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SecurityRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityRolesDataSourceModel

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

	var filter *interfaces.SecurityRoleDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.SecurityRoleDataSourceFilterModel{
			Name:    data.Filter.Name.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
			Scope:   data.Filter.Scope.ValueString(),
		}
	}
	restInfo, err := interfaces.GetSecurityRoles(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetSecurityRoles
		return
	}

	data.SecurityRules = make([]SecurityRoleDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.SecurityRules[index] = SecurityRoleDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
		}
		data.SecurityRules[index].Scope = types.StringValue(record.Scope)
		data.SecurityRules[index].Builtin = types.BoolValue(record.Builtin)

		setElements := []attr.Value{}
		for _, privilege := range record.Privileges {
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
		data.SecurityRules[index].Privileges = setValue
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
