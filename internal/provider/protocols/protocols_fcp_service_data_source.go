package protocols

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
var _ datasource.DataSource = &ProtocolsFcpServiceDataSource{}

// NewProtocolsFcpServiceDataSource is a helper function to simplify the provider implementation.
func NewProtocolsFcpServiceDataSource() datasource.DataSource {
	return &ProtocolsFcpServiceDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "fcp_service",
		},
	}
}

// ProtocolsFcpServiceDataSource defines the data source implementation.
type ProtocolsFcpServiceDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsFcpServiceDataSourceModel describes the data source data model.
type ProtocolsFcpServiceDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVMName       types.String `tfsdk:"svm_name"`
	Enabled       types.Bool   `tfsdk:"enabled"`
	ID            types.String `tfsdk:"id"`
}

// Metadata returns the data source type name.
func (d *ProtocolsFcpServiceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsFcpServiceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Protocols FC Protocol service data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "FC Protocol SVM name",
				Required:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "The administrative state of the FC Protocol service",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "FC Protocol service SVM UUID",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsFcpServiceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

		return
	}

	d.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *ProtocolsFcpServiceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsFcpServiceDataSourceModel

	// Read Terraform configuration data into the model
	diags := req.Config.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Use existing-, or create new REST API client
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	restInfo, err := interfaces.GetProtocolsFcpService(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetProtocolsFcpService
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No FCP service found", "FCP service not found.")
		return
	}

	data.Enabled = types.BoolValue(restInfo.Enabled)
	data.ID = types.StringValue(restInfo.SVM.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}
