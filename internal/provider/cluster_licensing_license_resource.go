package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ClusterLicensingLicenseResource{}
var _ resource.ResourceWithImportState = &ClusterLicensingLicenseResource{}

// NewClusterLicensingLicenseResource is a helper function to simplify the provider implementation.
func NewClusterLicensingLicenseResource() resource.Resource {
	return &ClusterLicensingLicenseResource{
		config: resourceOrDataSourceConfig{
			name: "cluster_licensing_license_resource",
		},
	}
}

// ClusterLicensingLicenseResource defines the resource implementation.
type ClusterLicensingLicenseResource struct {
	config resourceOrDataSourceConfig
}

// ClusterLicensingLicenseResourceModel describes the resource data model.
type ClusterLicensingLicenseResourceModel struct {
	CxProfileName types.String   `tfsdk:"cx_profile_name"`
	Keys          []types.String `tfsdk:"keys"`
	ID            types.String   `tfsdk:"id"`
	Name          types.String   `tfsdk:"name"`
	Scope         types.String   `tfsdk:"scope"`
	State         types.String   `tfsdk:"state"`
	SerialNumber  types.String   `tfsdk:"serial_number"`
}

// Metadata returns the resource type name.
func (r *ClusterLicensingLicenseResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *ClusterLicensingLicenseResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ClusterLicensingLicense resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"keys": schema.ListAttribute{
				Required:            true,
				MarkdownDescription: "List of NLF or 26-character keys",
				ElementType:         types.StringType,
			},
			"name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the license",
			},
			"scope": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Scope of the license",
			},
			"state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "State of the license",
			},
			"serial_number": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ClusterLicensingLicenseResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
}

// Create a resource and retrieve UUID
func (r *ClusterLicensingLicenseResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ClusterLicensingLicenseResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ClusterLicensingLicenseResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	var keys []string
	for _, v := range data.Keys {
		keys = append(keys, v.ValueString())
	}
	body.Keys = keys
	resource, err := interfaces.CreateClusterLicensingLicense(errorHandler, *client, body)
	if err != nil {
		return
	}
	if resource == nil {
		return // TODO: Add error
	}
	data.Name = types.StringValue(resource.Name)
	data.Scope = types.StringValue(resource.Scope)
	data.State = types.StringValue(resource.State)
	data.ID = types.StringValue(resource.Name)
	data.SerialNumber = types.StringValue(resource.Licenses[0].SerialNumber) // TODO: Double check there is only ever 1

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *ClusterLicensingLicenseResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClusterLicensingLicenseResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	restInfo, err := interfaces.GetClusterLicensingLicenses(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetClusterLicensingLicense
		return
	}
	if restInfo == nil {
		return // TODO: Fix
	}

	var matchingLicense interfaces.ClusterLicensingLicenseKeyDataModelONTAP

	for _, item := range restInfo {
		if data.Name.ValueString() == item.Name {
			matchingLicense = item
		}
	}

	data.Name = types.StringValue(matchingLicense.Name)
	data.State = types.StringValue(matchingLicense.State)
	data.Scope = types.StringValue(matchingLicense.Scope)
	data.SerialNumber = types.StringValue(matchingLicense.Licenses[0].SerialNumber) // TODO: Double check there is only ever 1

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ClusterLicensingLicenseResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ClusterLicensingLicenseResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// License updates are not supported
	err := errorHandler.MakeAndReportError("Update not supported for License", "Update not supported for License")
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ClusterLicensingLicenseResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ClusterLicensingLicenseResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	err = interfaces.DeleteClusterLicensingLicense(errorHandler, *client, data.Name.ValueString(), data.SerialNumber.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ClusterLicensingLicenseResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
