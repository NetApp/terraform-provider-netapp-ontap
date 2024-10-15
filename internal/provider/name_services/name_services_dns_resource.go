package name_services

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &NameServicesDNSResource{}
var _ resource.ResourceWithImportState = &NameServicesDNSResource{}

// NewNameServicesDNSResource is a helper function to simplify the provider implementation.
func NewNameServicesDNSResource() resource.Resource {
	return &NameServicesDNSResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "dns",
		},
	}
}

// NameServicesDNSResource defines the resource implementation.
type NameServicesDNSResource struct {
	config connection.ResourceOrDataSourceConfig
}

// NameServicesDNSResourceModel describes the resource data model.
type NameServicesDNSResourceModel struct {
	CxProfileName        types.String   `tfsdk:"cx_profile_name"`
	SVMName              types.String   `tfsdk:"svm_name"`
	ID                   types.String   `tfsdk:"id"`
	SkipConfigValidation types.Bool     `tfsdk:"skip_config_validation"`
	Domains              []types.String `tfsdk:"dns_domains"`
	NameServers          []types.String `tfsdk:"name_servers"`
}

// Metadata returns the resource type name.
func (r *NameServicesDNSResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *NameServicesDNSResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "NameServicesDNS resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface svm name",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "UUID of svm",
				Computed:            true,
			},
			"dns_domains": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of DNS domains such as 'sales.bar.com'. The first domain is the one that the svm belongs to",
				Optional:            true,
			},
			"name_servers": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "List of IPv4 addresses of name servers such as '123.123.123.123'.",
				Optional:            true,
			},
			"skip_config_validation": schema.BoolAttribute{
				MarkdownDescription: "Indicates whether or not the validation for the specified DNS configuration is disabled. (9.9)",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// create a function that add 2 numbers togehter

// Configure adds the provider configured client to the resource.
func (r *NameServicesDNSResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *NameServicesDNSResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NameServicesDNSResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	restInfo, err := interfaces.GetNameServicesDNS(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetNameServicesDNS
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No DNS found", fmt.Sprintf("NO DNS on svm %s found.", data.SVMName.ValueString()))
		return
	}

	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.ID = types.StringValue(restInfo.SVM.UUID)

	if restInfo.Servers != nil {
		data.NameServers = make([]types.String, len(restInfo.Servers))
		for index, server := range restInfo.Servers {
			data.NameServers[index] = types.StringValue(server)
		}
	}

	if restInfo.Domains != nil {
		data.Domains = make([]types.String, len(restInfo.Domains))
		for index, domain := range restInfo.Domains {
			data.Domains[index] = types.StringValue(domain)
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *NameServicesDNSResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *NameServicesDNSResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.NameServicesDNSGetDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.SVM.Name = data.SVMName.ValueString()
	body.SVM.UUID = data.ID.ValueString()

	var servers, domains []string
	for _, v := range data.NameServers {
		servers = append(servers, v.ValueString())
	}
	for _, v := range data.Domains {
		domains = append(domains, v.ValueString())
	}
	body.Servers = servers
	body.Domains = domains
	body.SkipConfigValidation = data.SkipConfigValidation.ValueBool()
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	dns, err := interfaces.CreateNameServicesDNS(errorHandler, *client, body)
	if err != nil {
		return
	}
	data.ID = types.StringValue(dns.SVM.UUID)

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *NameServicesDNSResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *NameServicesDNSResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// License updates are not supported
	err := errorHandler.MakeAndReportError("Update not supported for dns", "Update not supported for dns")
	if err != nil {
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *NameServicesDNSResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *NameServicesDNSResourceModel

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
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	err = interfaces.DeleteNameServicesDNS(errorHandler, *client, svm.UUID)
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *NameServicesDNSResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[1])...)
}
