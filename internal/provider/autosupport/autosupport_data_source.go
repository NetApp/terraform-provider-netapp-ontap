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
var _ datasource.DataSource = &AutoSupportDataSource{}

// NewAutoSupportDataSource is a helper function to simplify the provider implementation.
func NewAutoSupportDataSource() datasource.DataSource {
	return &AutoSupportDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "autosupport",
		},
	}
}

// AutoSupportDataSource defines the data source implementation.
type AutoSupportDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// Metadata returns the data source type name.
func (d *AutoSupportDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *AutoSupportDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "AutoSupport data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "AutoSupport identifier",
				Computed:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether the AutoSupport daemon is enabled. When enabled, AutoSupport messages are generated",
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *AutoSupportDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *AutoSupportDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data interfaces.AutoSupportDataSourceModel
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

	cluster, err := interfaces.GetAutoSupport(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetAutoSupport
		return
	}

	data.ID = data.CxProfileName // For data sources, use cx_profile_name as ID
	data.Enabled = types.BoolValue(cluster.Enabled)
	data.Transport = types.StringValue(cluster.Transport)
	data.From = types.StringValue(cluster.From)
	data.ContactSupport = types.BoolValue(cluster.ContactSupport)
	data.ProxyURL = types.StringValue(cluster.ProxyURL)
	data.IsMinimal = types.BoolValue(cluster.IsMinimal)
	data.OndemandEnabled = types.BoolValue(cluster.OndemandEnabled)
	data.SmtpEncryption = types.StringValue(cluster.SmtpEncryption)

	// Handle string slices for sets
	if cluster.To != nil {
		elements := []attr.Value{}
		for _, address := range cluster.To {
			elements = append(elements, types.StringValue(address))
		}
		setVal, diags := types.SetValue(types.StringType, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.To = setVal
	} else {
		data.To = types.SetNull(types.StringType)
	}

	if cluster.PartnerAddresses != nil {
		elements := []attr.Value{}
		for _, address := range cluster.PartnerAddresses {
			elements = append(elements, types.StringValue(address))
		}
		setVal, diags := types.SetValue(types.StringType, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.PartnerAddresses = setVal
	} else {
		data.PartnerAddresses = types.SetNull(types.StringType)
	}

	if cluster.MailHosts != nil {
		elements := []attr.Value{}
		for _, host := range cluster.MailHosts {
			elements = append(elements, types.StringValue(host))
		}
		setVal, diags := types.SetValue(types.StringType, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.MailHosts = setVal
	} else {
		data.MailHosts = types.SetNull(types.StringType)
	}

	tflog.Trace(ctx, "read a data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
