package security

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
var _ datasource.DataSource = &SecurityLoginMessageDataSource{}

// NewSecurityLoginMessageDataSource is a helper function to simplify the provider implementation.
func NewSecurityLoginMessageDataSource() datasource.DataSource {
	return &SecurityLoginMessageDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "security_login_message",
		},
	}
}

// SecurityLoginMessageDataSource defines the data source implementation.
type SecurityLoginMessageDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// SecurityLoginMessageDataSourceModel describes the data source data model.
type SecurityLoginMessageDataSourceModel struct {
	CxProfileName      types.String `tfsdk:"cx_profile_name"`
	Banner             types.String `tfsdk:"banner"`
	Message            types.String `tfsdk:"message"`
	ShowClusterMessage types.Bool   `tfsdk:"show_cluster_message"`
	Scope              types.String `tfsdk:"scope"`
	SVMName            types.String `tfsdk:"svm_name"`
	ID                 types.String `tfsdk:"id"`
}

// Metadata returns the data source type name.
func (d *SecurityLoginMessageDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *SecurityLoginMessageDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityLoginMessage data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"message": schema.StringAttribute{
				MarkdownDescription: "SecurityLoginMessage message",
				Optional:            true,
				Computed:            true,
			},
			"banner": schema.StringAttribute{
				MarkdownDescription: "SecurityLoginMessage banner",
				Optional:            true,
				Computed:            true,
			},
			"show_cluster_message": schema.BoolAttribute{
				MarkdownDescription: "Show cluster message",
				Optional:            true,
				Computed:            true,
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "SecurityLoginMessage scope",
				Optional:            true,
				Computed:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "SecurityLoginMessage svm name",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "SecurityAccount id",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *SecurityLoginMessageDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SecurityLoginMessageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityLoginMessageDataSourceModel

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

	restInfo, err := interfaces.GetSecurityLoginMessage(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSecurityLoginMessage
		return
	}

	data.Message = types.StringValue(restInfo.Message)
	data.Banner = types.StringValue(restInfo.Banner)
	data.ShowClusterMessage = types.BoolValue(restInfo.ShowClusterMessage)
	data.Scope = types.StringValue(restInfo.Scope)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.ID = types.StringValue(restInfo.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
