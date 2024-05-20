package security

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SecurityAccountDataSource{}

// NewSecurityAccountDataSource is a helper function to simplify the provider implementation.
func NewSecurityAccountDataSource() datasource.DataSource {
	return &SecurityAccountDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "security_account",
		},
	}
}

// SecurityAccountDataSource defines the data source implementation.
type SecurityAccountDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// SecurityAccountDataSourceModel describes the data source data model.
type SecurityAccountDataSourceModel struct {
	CxProfileName types.String                  `tfsdk:"cx_profile_name"`
	Name          types.String                  `tfsdk:"name"`
	Owner         *OwnerDataSourceModel         `tfsdk:"owner"`
	Locked        types.Bool                    `tfsdk:"locked"`
	Comment       types.String                  `tfsdk:"comment"`
	Role          *RoleDataSourceModel          `tfsdk:"role"`
	Scope         types.String                  `tfsdk:"scope"`
	Applications  []ApplicationsDataSourceModel `tfsdk:"applications"`
	ID            types.String                  `tfsdk:"id"`
}

// ApplicationsDataSourceModel describes the data source data model.
type ApplicationsDataSourceModel struct {
	Application                types.String    `tfsdk:"application"`
	SecondAuthentiactionMethod types.String    `tfsdk:"second_authentication_method"`
	AuthenticationMethods      *[]types.String `tfsdk:"authentication_methods"`
}

// RoleDataSourceModel describes the data source data model.
type RoleDataSourceModel struct {
	Name types.String `tfsdk:"name"`
}

// OwnerDataSourceModel describes the data source data model.
type OwnerDataSourceModel struct {
	Name    types.String `tfsdk:"name"`
	OwnerID types.String `tfsdk:"uuid"`
}

// Metadata returns the data source type name.
func (d *SecurityAccountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *SecurityAccountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityAccount data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "SecurityAccount name",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "SecurityAccount id",
				Computed:            true,
			},
			"owner": schema.SingleNestedAttribute{
				MarkdownDescription: "SecurityAccount owner",
				Computed:            true,
				Optional:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "SecurityAccount owner name",
						Required:            true,
					},
					"uuid": schema.StringAttribute{
						MarkdownDescription: "SecurityAccount owner uuid",
						Computed:            true,
					},
				},
			},
			"locked": schema.BoolAttribute{
				MarkdownDescription: "SecurityAccount locked",
				Computed:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "SecurityAccount comment",
				Computed:            true,
			},
			"role": schema.SingleNestedAttribute{
				MarkdownDescription: "SecurityAccount role",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "SecurityAccount role name",
						Computed:            true,
					},
				},
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "SecurityAccount scope",
				Computed:            true,
				Optional:            true,
			},
			"applications": schema.ListNestedAttribute{
				MarkdownDescription: "SecurityAccount applications",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"application": schema.StringAttribute{
							MarkdownDescription: "SecurityAccount application",
							Computed:            true,
						},
						"second_authentication_method": schema.StringAttribute{
							MarkdownDescription: "SecurityAccount second authentication method",
							Computed:            true,
						},
						"authentication_methods": schema.ListAttribute{
							MarkdownDescription: "SecurityAccount authentication methods",
							Computed:            true,
							ElementType:         types.StringType,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *SecurityAccountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SecurityAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityAccountDataSourceModel

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
	var svm *interfaces.SvmGetDataSourceModel
	if data.Owner != nil {
		svm, err = interfaces.GetSvmByName(errorHandler, *client, data.Owner.Name.ValueString())
		if err != nil {
			// error reporting done inside GetSvmByName
			return
		}
	}
	var restInfo *interfaces.SecurityAccountGetDataModelONTAP
	if svm == nil {
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

	data.Name = types.StringValue(restInfo.Name)
	// There is no ID in the REST response, so we use the name as ID
	data.ID = types.StringValue(restInfo.Name)
	data.Owner = &OwnerDataSourceModel{
		Name:    types.StringValue(restInfo.Owner.Name),
		OwnerID: types.StringValue(restInfo.Owner.UUID),
	}
	data.Locked = types.BoolValue(restInfo.Locked)
	data.Comment = types.StringValue(restInfo.Comment)
	data.Role = &RoleDataSourceModel{
		Name: types.StringValue(restInfo.Role.Name),
	}
	data.Scope = types.StringValue(restInfo.Scope)
	data.Applications = make([]ApplicationsDataSourceModel, len(restInfo.Applications))
	for index, application := range restInfo.Applications {
		data.Applications[index] = ApplicationsDataSourceModel{
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
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
