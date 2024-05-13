package provider

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ExportPolicyRuleDataSource{}

// var _ resource.ResourceWithImportState = &ExportPolicyRuleResource{}

// NewExportPolicyRuleDataSource is a helper function to simplify the provider implementation.
func NewExportPolicyRuleDataSource() datasource.DataSource {
	return &ExportPolicyRuleDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_nfs_export_policy_rule_data_source",
		},
	}
}

// ExportPolicyRuleDataSource defines the source implementation.
type ExportPolicyRuleDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ExportPolicyRuleDataSourceModel describes the source data model.
type ExportPolicyRuleDataSourceModel struct {
	CxProfileName       types.String   `tfsdk:"cx_profile_name"`
	ExportPolicyID      types.String   `tfsdk:"export_policy_id"`
	SVMName             types.String   `tfsdk:"svm_name"`
	ExportPolicyName    types.String   `tfsdk:"export_policy_name"`
	RoRule              []types.String `tfsdk:"ro_rule"`
	RwRule              []types.String `tfsdk:"rw_rule"`
	Protocols           []types.String `tfsdk:"protocols"`
	AnonymousUser       types.String   `tfsdk:"anonymous_user"`
	Superuser           []types.String `tfsdk:"superuser"`
	AllowDeviceCreation types.Bool     `tfsdk:"allow_device_creation"`
	NtfsUnixSecurity    types.String   `tfsdk:"ntfs_unix_security"`
	ChownMode           types.String   `tfsdk:"chown_mode"`
	AllowSuid           types.Bool     `tfsdk:"allow_suid"`
	ClientsMatch        []types.String `tfsdk:"clients_match"`
	Index               types.Int64    `tfsdk:"index"`
}

// Metadata returns the resource type name.
func (d *ExportPolicyRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the resource.
func (d *ExportPolicyRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Export policy rule resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the svm to use",
				Required:            true,
			},
			"export_policy_name": schema.StringAttribute{
				MarkdownDescription: "Export policy name",
				Required:            true,
			},
			"export_policy_id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Export policy identifier",
			},
			"ro_rule": schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "RO Access Rule",
				ElementType:         types.StringType,
			},
			"rw_rule": schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "RW Access Rule",
				ElementType:         types.StringType,
			},
			"clients_match": schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "List of Client Match Hostnames, IP Addresses, Netgroups, or Domains",
				ElementType:         types.StringType,
			},
			"protocols": schema.ListAttribute{
				Computed:            true,
				MarkdownDescription: "Access Protocol",
				ElementType:         types.StringType,
			},
			"anonymous_user": schema.StringAttribute{
				MarkdownDescription: "User ID To Which Anonymous Users Are Mapped",
				Computed:            true,
			},
			"superuser": schema.ListAttribute{
				MarkdownDescription: "Superuser Security Types",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"allow_device_creation": schema.BoolAttribute{
				MarkdownDescription: "Allow Creation of Devices",
				Computed:            true,
			},
			"ntfs_unix_security": schema.StringAttribute{
				MarkdownDescription: "NTFS export UNIX security options",
				Computed:            true,
			},
			"chown_mode": schema.StringAttribute{
				MarkdownDescription: "Specifies who is authorized to change the ownership mode of a file",
				Computed:            true,
			},
			"allow_suid": schema.BoolAttribute{
				MarkdownDescription: "Honor SetUID Bits in SETATTR",
				Computed:            true,
			},
			"index": schema.Int64Attribute{
				MarkdownDescription: "rule index",
				Required:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ExportPolicyRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ExportPolicyRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *ExportPolicyRuleDataSourceModel

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

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", fmt.Sprintf("cluster not found"))
		return
	}

	if data.Index.IsNull() {
		errorHandler.MakeAndReportError("error reading export policy rule", "rule index is null")
		return
	}

	if data.ExportPolicyName.IsNull() {
		errorHandler.MakeAndReportError("error reading export policy rule", "export policy name is null")
		return
	}

	var exportPolicyID string
	if data.ExportPolicyID.IsNull() {
		filter := map[string]string{
			"name":     data.ExportPolicyName.ValueString(),
			"svm.name": data.SVMName.ValueString(),
		}
		exportPolicy, err := interfaces.GetNfsExportPolicyByName(errorHandler, *client, &filter)
		if err != nil {
			return
		}
		exportPolicyID = strconv.Itoa(exportPolicy.ID)
	} else {
		exportPolicyID = data.ExportPolicyID.ValueString()
	}

	restInfo, err := interfaces.GetExportPolicyRuleSingle(errorHandler, *client, exportPolicyID, data.Index.ValueInt64(), cluster.Version)
	if err != nil {
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("error reading export policy rule", "rule not found")
		return
	}

	data.RoRule = connection.FlattenTypesStringList(restInfo.RoRule)
	data.RwRule = connection.FlattenTypesStringList(restInfo.RwRule)
	data.Protocols = connection.FlattenTypesStringList(restInfo.Protocols)
	data.Superuser = connection.FlattenTypesStringList(restInfo.Superuser)

	var clientsMatch []types.String
	for _, e := range restInfo.ClientsMatch {
		clientsMatch = append(clientsMatch, types.StringValue(e.Match))
	}
	data.ClientsMatch = clientsMatch

	data.AllowDeviceCreation = types.BoolValue(restInfo.AllowDeviceCreation)
	data.AllowSuid = types.BoolValue(restInfo.AllowSuid)
	data.ChownMode = types.StringValue(restInfo.ChownMode)
	data.NtfsUnixSecurity = types.StringValue(restInfo.NtfsUnixSecurity)
	data.AnonymousUser = types.StringValue(restInfo.AnonymousUser)
	data.ExportPolicyID = types.StringValue(exportPolicyID)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
