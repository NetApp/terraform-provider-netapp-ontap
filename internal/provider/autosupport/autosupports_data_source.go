package autosupport

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
var _ datasource.DataSource = &AutoSupportsDataSource{}

// NewAutoSupportsDataSource is a helper function to simplify the provider implementation.
func NewAutoSupportsDataSource() datasource.DataSource {
	return &AutoSupportsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "autosupports",
		},
	}
}

// AutoSupportsDataSource defines the data source implementation.
type AutoSupportsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// AutoSupportsDataSourceModel describes the data source data model.
type AutoSupportsDataSourceModel struct {
	CxProfileName types.String                            `tfsdk:"cx_profile_name"`
	AutoSupports  []interfaces.AutoSupportDataSourceModel `tfsdk:"autosupports"`
}

// Metadata returns the data source type name.
func (d *AutoSupportsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *AutoSupportsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AutoSupports data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"autosupports": schema.ListNestedAttribute{
				MarkdownDescription: "List of AutoSupport configurations",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Specifies whether the AutoSupport daemon is enabled",
							Computed:            true,
						},
						"transport": schema.StringAttribute{
							MarkdownDescription: "The name of the transport protocol used to deliver AutoSupport messages",
							Computed:            true,
						},
						"to_addresses": schema.SetAttribute{
							MarkdownDescription: "Specifies up to five recipients of full AutoSupport e-mail messages",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"from": schema.StringAttribute{
							MarkdownDescription: "The e-mail address from which the node sends AutoSupport messages",
							Computed:            true,
						},
						"contact_support": schema.BoolAttribute{
							MarkdownDescription: "Specifies whether AutoSupport notification to technical support is enabled",
							Computed:            true,
						},
						"partner_addresses": schema.SetAttribute{
							MarkdownDescription: "Specifies up to five partner vendor recipients of full AutoSupport e-mail messages",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"proxy_url": schema.StringAttribute{
							MarkdownDescription: "HTTP or HTTPS proxy if the transport parameter is set to HTTP or HTTPS",
							Computed:            true,
						},
						"mail_hosts": schema.SetAttribute{
							MarkdownDescription: "List of mail server(s) used to deliver AutoSupport messages via SMTP",
							Computed:            true,
							ElementType:         types.StringType,
						},
						"is_minimal": schema.BoolAttribute{
							MarkdownDescription: "Specifies whether the system information is collected in compliant form to remove private data or in complete form to enhance diagnostics",
							Computed:            true,
						},
						"ondemand_enabled": schema.BoolAttribute{
							MarkdownDescription: "Specifies whether the AutoSupport OnDemand Download feature is enabled",
							Computed:            true,
						},
						"smtp_encryption": schema.StringAttribute{
							MarkdownDescription: "SMTP encryption type for AutoSupport messages",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *AutoSupportsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Helper function to convert string slice to Set
func convertStringSliceToSet(slice []string) (types.Set, error) {
	if slice == nil || len(slice) == 0 {
		return types.SetNull(types.StringType), nil
	}
	
	elements := make([]attr.Value, len(slice))
	for i, s := range slice {
		elements[i] = types.StringValue(s)
	}
	
	set, diags := types.SetValue(types.StringType, elements)
	if diags.HasError() {
		return types.SetNull(types.StringType), fmt.Errorf("error creating set: %v", diags)
	}
	return set, nil
}

// Read refreshes the Terraform state with the latest data.
func (d *AutoSupportsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AutoSupportsDataSourceModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	clusters, err := interfaces.GetAutoSupports(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetAutoSupports
		return
	}

	for _, cluster := range clusters {
		var record interfaces.AutoSupportDataSourceModel
		
		record.CxProfileName = data.CxProfileName
		record.Enabled = types.BoolValue(cluster.Enabled)
		record.Transport = types.StringValue(cluster.Transport)
		record.From = types.StringValue(cluster.From)
		record.ContactSupport = types.BoolValue(cluster.ContactSupport)
		record.ProxyURL = types.StringValue(cluster.ProxyURL)
		record.IsMinimal = types.BoolValue(cluster.IsMinimal)
		record.OndemandEnabled = types.BoolValue(cluster.OndemandEnabled)
		record.SmtpEncryption = types.StringValue(cluster.SmtpEncryption)

		// Handle string slices for sets
		toSet, err := convertStringSliceToSet(cluster.To)
		if err != nil {
			resp.Diagnostics.AddError("Error converting to", err.Error())
			return
		}
		record.To = toSet

		partnerAddressesSet, err := convertStringSliceToSet(cluster.PartnerAddresses)
		if err != nil {
			resp.Diagnostics.AddError("Error converting partner_addresses", err.Error())
			return
		}
		record.PartnerAddresses = partnerAddressesSet

		mailHostsSet, err := convertStringSliceToSet(cluster.MailHosts)
		if err != nil {
			resp.Diagnostics.AddError("Error converting mail_hosts", err.Error())
			return
		}
		record.MailHosts = mailHostsSet

		data.AutoSupports = append(data.AutoSupports, record)
	}

	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
