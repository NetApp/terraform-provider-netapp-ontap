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
var _ datasource.DataSource = &SecurityLoginMessagesDataSource{}

// NewSecurityLoginMessagesDataSource is a helper function to simplify the provider implementation.
func NewSecurityLoginMessagesDataSource() datasource.DataSource {
	return &SecurityLoginMessagesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "security_login_messages",
		},
	}
}

// SecurityLoginMessagesDataSource defines the data source implementation.
type SecurityLoginMessagesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// SecurityLoginMessagesDataSourceModel describes the data source data model.
type SecurityLoginMessagesDataSourceModel struct {
	CxProfileName         types.String                                `tfsdk:"cx_profile_name"`
	SecurityLoginMessages []SecurityLoginMessageDataSourceModel       `tfsdk:"security_login_messages"`
	Filter                *SecurityLoginMessagesDataSourceFilterModel `tfsdk:"filter"`
}

// SecurityLoginMessagesDataSourceFilterModel describes the data source data model for queries.
type SecurityLoginMessagesDataSourceFilterModel struct {
	Banner  types.String `tfsdk:"banner"`
	Message types.String `tfsdk:"message"`
	Scope   types.String `tfsdk:"scope"`
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *SecurityLoginMessagesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *SecurityLoginMessagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityLoginMessages data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"banner": schema.StringAttribute{
						MarkdownDescription: "SecurityLoginMessage banner",
						Optional:            true,
					},
					"message": schema.StringAttribute{
						MarkdownDescription: "SecurityLoginMessage message",
						Optional:            true,
					},
					"scope": schema.StringAttribute{
						MarkdownDescription: "SecurityLoginMessage scope",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "SecurityLoginMessage svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"security_login_messages": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"message": schema.StringAttribute{
							MarkdownDescription: "SecurityLoginMessage message",
							Computed:            true,
						},
						"banner": schema.StringAttribute{
							MarkdownDescription: "SecurityLoginMessage banner",
							Computed:            true,
						},
						"show_cluster_message": schema.BoolAttribute{
							MarkdownDescription: "Show cluster message",
							Computed:            true,
						},
						"scope": schema.StringAttribute{
							MarkdownDescription: "SecurityLoginMessage scope",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "SecurityLoginMessage svm name",
							Optional:            true,
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "SecurityLoginMessage id",
							Computed:            true,
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
func (d *SecurityLoginMessagesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SecurityLoginMessagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SecurityLoginMessagesDataSourceModel

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

	var filter *interfaces.SecurityLoginMessageDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.SecurityLoginMessageDataSourceFilterModel{
			Banner:  data.Filter.Banner.ValueString(),
			Message: data.Filter.Message.ValueString(),
			Scope:   data.Filter.Scope.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetSecurityLoginMessages(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetSecurityLoginMessages
		return
	}

	data.SecurityLoginMessages = make([]SecurityLoginMessageDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.SecurityLoginMessages[index] = SecurityLoginMessageDataSourceModel{
			CxProfileName:      types.String(data.CxProfileName),
			Message:            types.StringValue(record.Message),
			Banner:             types.StringValue(record.Banner),
			ShowClusterMessage: types.BoolValue(record.ShowClusterMessage),
			Scope:              types.StringValue(record.Scope),
			SVMName:            types.StringValue(record.SVM.Name),
			ID:                 types.StringValue(record.UUID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
