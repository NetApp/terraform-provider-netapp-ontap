package protocols

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &CifsLocalGroupDataSource{}

// NewCifsLocalGroupDataSource is a helper function to simplify the provider implementation.
func NewCifsLocalGroupDataSource() datasource.DataSource {
	return &CifsLocalGroupDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cifs_local_group",
		},
	}
}

// CifsLocalGroupDataSource defines the data source implementation.
type CifsLocalGroupDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// CifsLocalGroupDataSourceModel describes the data source data model.
type CifsLocalGroupDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	SVMName       types.String `tfsdk:"svm_name"`
	ID            types.String `tfsdk:"id"`
	Description   types.String `tfsdk:"description"`
	Members       []Member     `tfsdk:"members"`
}

// Member describes the data source data model.
type Member struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *CifsLocalGroupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *CifsLocalGroupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "CifsLocalGroup data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Cifs Local Group name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Svm name",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Cifs Local Group description",
				Computed:            true,
			},
			"members": schema.ListNestedAttribute{
				MarkdownDescription: "Cifs Local Group members",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Cifs Local Group member",
							Computed:            true,
						},
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Cifs Local Group id",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CifsLocalGroupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *CifsLocalGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CifsLocalGroupDataSourceModel

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

	restInfo, err := interfaces.GetCifsLocalGroupByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetCifsLocalGroup
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.ID = types.StringValue(restInfo.SID)
	data.Description = types.StringValue(restInfo.Description)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Members = make([]Member, len(restInfo.Members))
	for i, member := range restInfo.Members {
		data.Members[i].Name = types.StringValue(member.Name)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
