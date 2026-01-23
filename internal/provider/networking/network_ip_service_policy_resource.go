package networking

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &IPServicePolicyResource{}
var _ resource.ResourceWithImportState = &IPServicePolicyResource{}

// NewIPServicePolicyResource is a helper function to simplify the provider implementation.
func NewIPServicePolicyResource() resource.Resource {
	return &IPServicePolicyResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "network_ip_service_policy",
		},
	}
}

// IPServicePolicyResource defines the resource implementation.
type IPServicePolicyResource struct {
	config connection.ResourceOrDataSourceConfig
}

// IPServicePolicyResourceModel describes the resource data model.
type IPServicePolicyResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	Scope         types.String `tfsdk:"scope"`
	IPspace       types.String `tfsdk:"ipspace"`
	Services	  types.Set    `tfsdk:"services"`
	SVMName       types.String `tfsdk:"svm_name"`
	ID            types.String `tfsdk:"id"`
}

// Metadata returns the resource type name
func (r *IPServicePolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *IPServicePolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "IPServicePolicy resource",
		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the service policy",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "The name of the SVM",
				Required:            true,
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "Set to \"svm\" for service policies owned by an SVM. Otherwise, set to \"cluster\".",
				Optional:            true,
				Computed:			 true,
			},
			"ipspace": schema.StringAttribute{
				MarkdownDescription: `
				The IPspace that owns the service policy.
				Required for cluster-scoped service policies.
				Optional for SVM-scoped service policies.
				`,
				Optional:            true,
				Computed:            true,
			},
			"services": schema.SetAttribute{
				MarkdownDescription: "A list of services that should be included in this policy",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
				Validators: []validator.Set{
					setvalidator.ValueStringsAre(stringvalidator.OneOf(
						"cluster_core",
						"intercluster_core",
						"management_core",
						"management_autosupport",
						"management_bgp",
						"management_ems",
						"management_https",
						"management_http",
						"management_ssh",
						"management_portmap",
						"data_core",
						"data_nfs",
						"data_cifs",
						"data_flexcache",
						"data_iscsi",
						"data_s3_server",
						"data_dns_server",
						"data_fpolicy_client",
						"management_ntp_client",
						"management_dns_client",
						"management_ad_client",
						"management_ldap_client",
						"management_nis_client",
						"management_snmp_server",
						"management_rsh_server",
						"management_telnet_server",
						"management_ntp_server",
						"data_nvme_tcp",
						"backup_ndmp_control",
						"management_log_forwarding",
					)),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Service policy UUID",
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *IPServicePolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *IPServicePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data IPServicePolicyResourceModel

	// Read Terraform prior state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside New Client
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

	var restInfo *interfaces.IPServicePolicyGetDataModelONTAP
	restInfo, err = interfaces.GetServicePolicyByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString(), data.IPspace.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetServicePolicyByUUID, GetServicePolicyByName
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
	for i, service := range restInfo.Services {
		services[i] = service
	}
	servicesSet, diags := types.SetValueFrom(ctx, types.StringType, services)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Services = servicesSet

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *IPServicePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *IPServicePolicyResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.IPServicePolicyResourceBodyDataModel
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	if !data.SVMName.IsNull() {
		body.SVM.Name = data.SVMName.ValueString()
	} else {
		body.Scope = "cluster"
		if !data.IPspace.IsNull() {
			body.IPspace = data.IPspace.ValueString()
		} else {
			errorHandler.MakeAndReportError("error creating service policy", "ipspace is required when creating a cluster-scoped service policy")
			return
		}
	}
	// services
	if !data.Services.IsNull() && !data.Services.IsUnknown() {
		if len(data.Services.Elements()) > 0 {
			serviceElements := make([]types.String, 0, len(data.Services.Elements()))
			diags := data.Services.ElementsAs(ctx, &serviceElements, false)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			services := make([]string, len(serviceElements))
			for i, service := range serviceElements {
				services[i] = service.ValueString()
			}
			body.Services = services
		} else {
			// If data.Services length is zero, pass empty list in the body
			body.Services = []string{}
		}
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
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

	resource, err := interfaces.CreateIPServicePolicy(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.ID = types.StringValue(resource.UUID)
	restInfo, err := interfaces.GetServicePolicyByUUID(errorHandler, *client, data.ID.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetServicePolicyByUUID
		return
	}
	
	// Update the computed parameters
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
	for i, service := range restInfo.Services {
		services[i] = service
	}
	servicesSet, diags := types.SetValueFrom(ctx, types.StringType, services)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Services = servicesSet

	tflog.Trace(ctx, fmt.Sprintf("created service policy resource, ID=%s", data.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *IPServicePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *IPServicePolicyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Read Terraform state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, state.CxProfileName)
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

	// Update the resource
	var body interfaces.IPServicePolicyResourceBodyDataModel

	if !plan.Name.Equal(state.Name) {
		body.Name = plan.Name.ValueString()
	}
	// services update
	if !plan.Services.Equal(state.Services) {
		if !plan.Services.IsNull() && !plan.Services.IsUnknown() {
			if len(plan.Services.Elements()) > 0 {
				serviceElements := make([]types.String, 0, len(plan.Services.Elements()))
				diags := plan.Services.ElementsAs(ctx, &serviceElements, false)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				services := make([]string, len(serviceElements))
				for i, service := range serviceElements {
					services[i] = service.ValueString()
				}
				body.Services = services
			} else {
				// If plan.Services length is zero, pass empty list in the body
				body.Services = []string{}
			}
		}
	}

	err = interfaces.UpdateIPServicePolicy(errorHandler, *client, body, plan.ID.ValueString())
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetServicePolicyByUUID(errorHandler, *client, plan.ID.ValueString(), cluster.Version)
	if err != nil {
		return
	}
	
	// Update the computed parameters
	plan.Name = types.StringValue(restInfo.Name)
	plan.SVMName = types.StringValue(restInfo.SVM.Name)
	plan.Scope = types.StringValue(restInfo.Scope)
	
	// ipspace
	if restInfo.IPspace != nil {
		plan.IPspace = types.StringValue(restInfo.IPspace.Name)
	} else {
		plan.IPspace = types.StringNull()
	}

	// services - map list to set
	var services = make([]string, len(restInfo.Services))
	for i, service := range restInfo.Services {
		services[i] = service
	}
	servicesSet, diags := types.SetValueFrom(ctx, types.StringType, services)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Services = servicesSet

	tflog.Debug(ctx, fmt.Sprintf("updated service policy resource: ID=%s", plan.ID))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *IPServicePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *IPServicePolicyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("ID is null", "service policy UUID is null")
		return
	}

	err = interfaces.DeleteIPServicePolicy(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *IPServicePolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	// import using name,scope,cx_profile_name (for cluster-scoped policies)
	// or name,svm_name,cx_profile_name (for SVM-scoped policies)
	if len(idParts) == 3 {
		// Check if the second part is "cluster" scope or an SVM name
		if idParts[1] == "cluster" {
			// Cluster-scoped: name,scope=cluster,cx_profile_name
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scope"), idParts[1])...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
			// For cluster-scoped, svm_name should be empty
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), "")...)
		} else {
			// SVM-scoped: name,svm_name,cx_profile_name
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[1])...)
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
			// For SVM-scoped, scope is inferred as "svm"
			resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("scope"), "svm")...)
		}
		return
	}
	
	resp.Diagnostics.AddError(
		"Unexpected Import Identifier",
		fmt.Sprintf("Expected import identifier with format: 'name,scope=cluster,cx_profile_name' for cluster-scoped policies or 'name,svm_name,cx_profile_name' for SVM-scoped policies. Got: %q", req.ID),
	)
}
