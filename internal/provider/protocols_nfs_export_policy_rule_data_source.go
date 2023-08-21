package provider

import (
	"context"
	"fmt"
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
		config: resourceOrDataSourceConfig{
			name: "protocols_nfs_export_policy_rule_data_source",
		},
	}
}

// ExportPolicyRuleDataSource defines the source implementation.
type ExportPolicyRuleDataSource struct {
	config resourceOrDataSourceConfig
}

// ExportPolicyRuleDataSourceModel describes the source data model.
type ExportPolicyRuleDataSourceModel struct {
	CxProfileName       types.String   `tfsdk:"cx_profile_name"`
	ExportPolicyID      types.String   `tfsdk:"export_policy_id"`
	SVMName             types.String   `tfsdk:"svm_name"`
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
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
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
func (d *ExportPolicyRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data *ExportPolicyRuleResourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
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

	restInfo, err := interfaces.GetExportPolicyRule(errorHandler, *client, exportPolicyID, data.Index.ValueInt64())
	if err != nil {
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("error reading export policy rule", "rule not found")
		return
	}
	var roRule, rwRule, protocols, superuser, clientsMatch []types.String
	for _, e := range restInfo.RoRule {
		roRule = append(roRule, types.StringValue(e))
	}

	for _, e := range restInfo.RwRule {
		rwRule = append(rwRule, types.StringValue(e))
	}
	for _, e := range restInfo.Protocols {
		protocols = append(protocols, types.StringValue(e))
	}
	for _, e := range restInfo.Superuser {
		superuser = append(superuser, types.StringValue(e))
	}
	data.RoRule = roRule
	data.RwRule = rwRule
	data.Protocols = protocols
	data.Superuser = superuser

	for _, e := range restInfo.ClientsMatch {
		clientsMatch = append(clientsMatch, types.StringValue(e.Match))
	}
	data.ClientsMatch = clientsMatch

	if !data.AllowDeviceCreation.IsNull() {
		data.AllowDeviceCreation = types.BoolValue(restInfo.AllowDeviceCreation)
	}

	if !data.AllowSuid.IsNull() {
		data.AllowSuid = types.BoolValue(restInfo.AllowSuid)
	}

	if !data.ChownMode.IsNull() {
		data.ChownMode = types.StringValue(restInfo.ChownMode)
	}
	if !data.NtfsUnixSecurity.IsNull() {
		data.NtfsUnixSecurity = types.StringValue(restInfo.NtfsUnixSecurity)
	}
	if !data.AnonymousUser.IsNull() {
		data.AnonymousUser = types.StringValue(restInfo.AnonymousUser)
	}
	if data.ExportPolicyID.IsNull() {
		data.ExportPolicyID = types.StringValue(exportPolicyID)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
