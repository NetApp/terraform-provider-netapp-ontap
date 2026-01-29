package networking

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
var _ datasource.DataSource = &IPServicePolicyDataSource{}

// NewIPServicePolicyDataSource is a helper function to simplify the provider implementation.
func NewIPServicePolicyDataSource() datasource.DataSource {
	return &IPServicePolicyDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "network_ip_service_policy",
		},
	}
}

// IPServicePolicyDataSource defines the data source implementation.
type IPServicePolicyDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// IPServicePolicyDataSourceModel describes the data source data model.
type IPServicePolicyDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	Scope         types.String `tfsdk:"scope"`
	IPspace 	  types.String `tfsdk:"ipspace"`
	Services      types.Set    `tfsdk:"services"`
	SVMName       types.String `tfsdk:"svm_name"`
	ID            types.String `tfsdk:"id"`
}

// IPServicePolicyDataSourceFilterModel describes the data source data model for queries.
type IPServicePolicyDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
	Scope   types.String `tfsdk:"scope"`
	Name    types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *IPServicePolicyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *IPServicePolicyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "IPServicePolicy data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service policy",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "The name of the SVM on which the service policy exists",
				Optional:            true,
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "Set to \"svm\" for service policies owned by an SVM. Otherwise, set to \"cluster\".",
				Optional:            true,
			},
			"ipspace": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "IP space name",
			},
			"services": schema.SetAttribute{
				Computed:            true,
				MarkdownDescription: "A list of services that should be included in this policy",
				ElementType:         types.StringType,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Service policy UUID",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *IPServicePolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *IPServicePolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IPServicePolicyDataSourceModel

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

	restInfo, err := interfaces.GetServicePolicyByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString(), data.IPspace.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetServicePolicyByName
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No service policy found", "service policy not found")
		return
	}
	data.ID = types.StringValue(restInfo.UUID)
	data.Name = types.StringValue(restInfo.Name)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Scope = types.StringValue(restInfo.Scope)

	// ipspace
	if restInfo.IPspace != nil {
		data.IPspace = types.StringValue(restInfo.IPspace.Name)
	} else {
		data.IPspace = types.StringNull()
	}

	// services - map list to set
	var services = make([]string, len(restInfo.Services))
	copy(services, restInfo.Services)
	servicesSet, diags := types.SetValueFrom(ctx, types.StringType, services)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Services = servicesSet

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
