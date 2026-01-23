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
var _ datasource.DataSource = &IPServicePoliciesDataSource{}

// NewIPServicePoliciesDataSource is a helper function to simplify the provider implementation.
func NewIPServicePoliciesDataSource() datasource.DataSource {
	return &IPServicePoliciesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "network_ip_service_policies",
		},
	}
}

// IPServicePoliciesDataSource defines the data source implementation.
type IPServicePoliciesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// IPServicePoliciesDataSourceModel describes the data source data model.
type IPServicePoliciesDataSourceModel struct {
	CxProfileName     types.String                          `tfsdk:"cx_profile_name"`
	IPServicePolicies []IPServicePolicyDataSourceModel      `tfsdk:"ip_service_policies"`
	Filter            *IPServicePolicyDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *IPServicePoliciesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *IPServicePoliciesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "IPServicePolicies data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"scope": schema.StringAttribute{
						MarkdownDescription: "The name of the SVM",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "The name of the SVM",
						Optional:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "The name of the service policy",
						Optional:            true,
					},
				},
				Required: true,
			},
			"ip_service_policies": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
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
							Computed:            true,
						},
						"scope": schema.StringAttribute{
							MarkdownDescription: "Set to \"svm\" for service policies owned by an SVM. Otherwise, set to \"cluster\".",
							Computed:            true,
						},
						"ipspace": schema.StringAttribute{
							Computed:            true,
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
				},
				Computed:            true,
				MarkdownDescription: "",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *IPServicePoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *IPServicePoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data IPServicePoliciesDataSourceModel

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

	var filter *interfaces.IPServicePolicyDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.IPServicePolicyDataSourceFilterModel{
			Scope:   data.Filter.Scope.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
			Name:    data.Filter.Name.ValueString(),
		}
	}
	restInfo, err := interfaces.GetServicePolicies(errorHandler, *client, filter, cluster.Version)
	if err != nil {
		// error reporting done inside GetServicePolicies
		return
	}

	data.IPServicePolicies = make([]IPServicePolicyDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		//ipspace
		var ipspaceValue types.String
		if record.IPspace != nil {
			ipspaceValue = types.StringValue(record.IPspace.Name)
		} else {
			ipspaceValue = types.StringNull()
		}

		// services - map list to set
		var services = make([]string, len(record.Services))
		copy(services, record.Services)
		servicesSet, diags := types.SetValueFrom(ctx, types.StringType, services)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.IPServicePolicies[index] = IPServicePolicyDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
			SVMName:       types.StringValue(record.SVM.Name),
			Scope:		   types.StringValue(record.Scope),
			IPspace:       ipspaceValue,
			Services:      servicesSet,
			ID:            types.StringValue(record.UUID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
