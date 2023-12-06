package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &CifsLocalUserDataSource{}

// NewCifsLocalUserDataSource is a helper function to simplify the provider implementation.
func NewCifsLocalUserDataSource() datasource.DataSource {
	return &CifsLocalUserDataSource{
		config: resourceOrDataSourceConfig{
			name: "protocols_cifs_local_user_data_source",
		},
	}
}

// CifsLocalUserDataSource defines the data source implementation.
type CifsLocalUserDataSource struct {
	config resourceOrDataSourceConfig
}

// CifsLocalUserDataSourceModel describes the data source data model.
type CifsLocalUserDataSourceModel struct {
	CxProfileName   types.String `tfsdk:"cx_profile_name"`
	Name            types.String `tfsdk:"name"`
	SVMName         types.String `tfsdk:"svm_name"`
	FullName        types.String `tfsdk:"full_name"`
	Description     types.String `tfsdk:"description"`
	Membership      []Membership `tfsdk:"membership"`
	AccountDisabled types.Bool   `tfsdk:"account_disabled"`
	ID              types.String `tfsdk:"id"`
}

// Membership describes the membership data model.
type Membership struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *CifsLocalUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *CifsLocalUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsLocalUser data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "CifsLocalUser name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface svm name",
				Required:            true,
			},
			"full_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "CifsLocalUser full name",
			},
			"description": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "CifsLocalUser description",
			},
			"membership": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "CifsLocalUser membership",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "CifsLocalUser membership name",
						},
					},
				},
			},
			"account_disabled": schema.BoolAttribute{
				Computed:            true,
				MarkdownDescription: "CifsLocalUser account disabled",
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "CifsLocalUser id",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CifsLocalUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *CifsLocalUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CifsLocalUserDataSourceModel

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

	restInfo, err := interfaces.GetCifsLocalUserByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetCifsLocalUser
		return
	}
	data.ID = types.StringValue(restInfo.SID)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.FullName = types.StringValue(restInfo.FullName)
	data.Description = types.StringValue(restInfo.Description)
	data.AccountDisabled = types.BoolValue(restInfo.AccountDisabled)
	data.Membership = make([]Membership, len(restInfo.Membership))
	for i, m := range restInfo.Membership {
		data.Membership[i].Name = types.StringValue(m.Name)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
