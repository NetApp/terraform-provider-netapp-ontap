package protocols

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsNFSExportPolicyRulesDataSource{}

// NewExportPolicyRulesDataSource is a helper function to simplify the provider implementation.
func NewExportPolicyRulesDataSource() datasource.DataSource {
	return &ProtocolsNFSExportPolicyRulesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_nfs_export_policy_rules_data_source",
		},
	}
}

// ProtocolsNFSExportPolicyRulesDataSource defines the data source implementation.
type ProtocolsNFSExportPolicyRulesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ExportPolicyRulesDataSourceModel describes the data source data model.
type ExportPolicyRulesDataSourceModel struct {
	CxProfileName                 types.String                           `tfsdk:"cx_profile_name"`
	SVMName                       types.String                           `tfsdk:"svm_name"`
	ExportPolicyName              types.String                           `tfsdk:"export_policy_name"`
	ProtocolsNFSExportPolicyRules []ExportPolicyRuleDataSourceModel      `tfsdk:"protocols_nfs_export_policy_rules"`
	Filter                        *ExportPolicyRuleDataSourceFilterModel `tfsdk:"filter"`
}

// ExportPolicyRuleDataSourceFilterModel describes the data source data model for queries.
type ExportPolicyRuleDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsNFSExportPolicyRulesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsNFSExportPolicyRulesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsNFSExportPolicyRules data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the svm to use",
			},
			"export_policy_name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Export policy name",
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "ProtocolsNFSExportPolicyRule svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_nfs_export_policy_rules": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Connection profile name",
						},
						"svm_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Name of the svm to use",
						},
						"export_policy_name": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Export policy name",
						},
						"export_policy_id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Export policy identifier",
						},
						"ro_rule": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "RO Access Rule",
						},
						"rw_rule": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "RW Access Rule",
						},
						"clients_match": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "List of Client Match Hostnames, IP Addresses, Netgroups, or Domains",
						},
						"protocols": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Access Protocol",
						},
						"anonymous_user": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "User ID To Which Anonymous Users Are Mapped",
						},
						"superuser": schema.ListAttribute{
							ElementType:         types.StringType,
							Computed:            true,
							MarkdownDescription: "Superuser Security Types",
						},
						"allow_device_creation": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Allow Creation of Devices",
						},
						"ntfs_unix_security": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "NTFS export UNIX security options",
						},
						"chown_mode": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Specifies who is authorized to change the ownership mode of a file",
						},
						"allow_suid": schema.BoolAttribute{
							Computed:            true,
							MarkdownDescription: "Honor SetUID Bits in SETATTR",
						},
						"index": schema.Int64Attribute{
							Computed:            true,
							MarkdownDescription: "rule index",
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "Export policy rule resource",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsNFSExportPolicyRulesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsNFSExportPolicyRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ExportPolicyRulesDataSourceModel

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

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	var filter *interfaces.ExportPolicyRuleDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.ExportPolicyRuleDataSourceFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}

	var exportPolicyID string
	exportPolicyByNameFilter := map[string]string{
		"name":     data.ExportPolicyName.ValueString(),
		"svm.name": data.SVMName.ValueString(),
	}
	exportPolicy, err := interfaces.GetNfsExportPolicyByName(errorHandler, *client, &exportPolicyByNameFilter)
	if err != nil {
		return
	}
	exportPolicyID = strconv.Itoa(exportPolicy.ID)

	restInfo, err := interfaces.GetListExportPolicyRules(errorHandler, *client, exportPolicyID, filter, cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsNFSExportPolicyRules
		return
	}

	data.ProtocolsNFSExportPolicyRules = make([]ExportPolicyRuleDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.ProtocolsNFSExportPolicyRules[index] = ExportPolicyRuleDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			RoRule:        connection.FlattenTypesStringList(record.RoRule),
			RwRule:        connection.FlattenTypesStringList(record.RwRule),
			Protocols:     connection.FlattenTypesStringList(record.Protocols),
			Superuser:     connection.FlattenTypesStringList(record.Superuser),
		}

		var clientsMatch []types.String
		for _, e := range record.ClientsMatch {
			clientsMatch = append(clientsMatch, types.StringValue(e.Match))
		}

		data.ProtocolsNFSExportPolicyRules[index].ClientsMatch = clientsMatch
		data.ProtocolsNFSExportPolicyRules[index].SVMName = types.StringValue(record.Svm.Name)
		data.ProtocolsNFSExportPolicyRules[index].AllowDeviceCreation = types.BoolValue(record.AllowDeviceCreation)
		data.ProtocolsNFSExportPolicyRules[index].AllowSuid = types.BoolValue(record.AllowSuid)
		data.ProtocolsNFSExportPolicyRules[index].ChownMode = types.StringValue(record.ChownMode)
		data.ProtocolsNFSExportPolicyRules[index].NtfsUnixSecurity = types.StringValue(record.NtfsUnixSecurity)
		data.ProtocolsNFSExportPolicyRules[index].AnonymousUser = types.StringValue(record.AnonymousUser)
		data.ProtocolsNFSExportPolicyRules[index].ExportPolicyID = types.StringValue(exportPolicyID)
		data.ProtocolsNFSExportPolicyRules[index].ExportPolicyName = types.StringValue(record.ExportPolicy.Name)
		data.ProtocolsNFSExportPolicyRules[index].Index = types.Int64Value(record.Index)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
