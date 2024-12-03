package svm

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
	_ resource.Resource              = &SvmQosPolicyActivationResource{}
	_ resource.ResourceWithConfigure = &SvmQosPolicyActivationResource{}
)

// NewSvmQosPolicyActivationResource is a helper function to simplify the provider implementation.
func NewSvmQosPolicyActivationResource() resource.Resource {
	return &SvmQosPolicyActivationResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "svm_qos_policy_activation",
		},
	}
}

// SvmQosPolicyActivationResource defines the resource implementation.
type SvmQosPolicyActivationResource struct {
	config connection.ResourceOrDataSourceConfig
}

// SvmQosPolicyActivationResourceModel describes the resource data model.
type SvmQosPolicyActivationResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SvmID         types.String `tfsdk:"svm_id"`
	QoSPolicyID   types.String `tfsdk:"qos_policy_id"`
}

// Metadata returns the resource type name.
func (r *SvmQosPolicyActivationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *SvmQosPolicyActivationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Resource to activate QoS Policy for SVM",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"svm_id": schema.StringAttribute{
				MarkdownDescription: "SVM UUID",
				Required:            true,
			},
			"qos_policy_id": schema.StringAttribute{
				MarkdownDescription: "QoS Policy UUID",
				Required:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SvmQosPolicyActivationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

		return
	}

	r.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *SvmQosPolicyActivationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SvmQosPolicyActivationResourceModel

	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Use existing-, or create new REST API client
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(
		errorHandler,
		r.config,
		data.CxProfileName,
	)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for reading svm info
	tflog.Debug(ctx, fmt.Sprintf("read a svm resource: %#v", data))
	svm, err := interfaces.GetSvm(
		errorHandler,
		*client,
		data.SvmID.ValueString(),
	)
	if err != nil {
		return
	}
	if svm == nil {
		errorHandler.MakeAndReportError(
			"No SVM Found",
			fmt.Sprintf("No SVM '%s' found.", data.SvmID.ValueString()),
		)

		return
	}

	// Copy svm info to resource model
	data.SvmID = types.StringValue(svm.UUID)
	data.QoSPolicyID = types.StringValue(svm.QoSPolicy.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Create a resource and retrieve UUID
func (r *SvmQosPolicyActivationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SvmQosPolicyActivationResourceModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Copy QoS policy info to request body
	var body interfaces.SvmResourceModel
	body.QoSPolicy.UUID = data.QoSPolicyID.ValueString()

	// Use existing-, or create new REST API client
	client, err := connection.GetRestClient(
		errorHandler,
		r.config,
		data.CxProfileName,
	)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for updating svm
	tflog.Debug(ctx, fmt.Sprintf("update a svm resource: %#v", data))
	err = interfaces.UpdateSvm(
		errorHandler,
		*client,
		body,
		data.SvmID.ValueString(),
		true,
		true,
	)
	if err != nil {
		return
	}

	// Save data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SvmQosPolicyActivationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SvmQosPolicyActivationResourceModel

	// Read Terraform plan data into the model
	diags := req.Plan.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Copy QoS policy info to request body
	var body interfaces.SvmResourceModel
	body.QoSPolicy.UUID = data.QoSPolicyID.ValueString()

	// Use existing-, or create new REST API client
	client, err := connection.GetRestClient(
		errorHandler,
		r.config,
		data.CxProfileName,
	)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for updating svm
	tflog.Debug(ctx, fmt.Sprintf("update a svm resource: %#v", data))
	err = interfaces.UpdateSvm(
		errorHandler,
		*client,
		body,
		data.SvmID.ValueString(),
		true,
		true,
	)
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	diags = resp.State.Set(ctx, &data)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SvmQosPolicyActivationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SvmQosPolicyActivationResourceModel

	// Read Terraform prior state data into the model
	diags := req.State.Get(ctx, &data)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Set QoS policy to 'none' in request body
	var body interfaces.SvmResourceModel
	body.QoSPolicy.Name = "none"

	// Use existing-, or create new REST API client
	client, err := connection.GetRestClient(
		errorHandler,
		r.config,
		data.CxProfileName,
	)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Call ONTAP REST API for updating svm
	tflog.Debug(ctx, fmt.Sprintf("update a svm resource: %#v", data))
	err = interfaces.UpdateSvm(
		errorHandler,
		*client,
		body,
		data.SvmID.ValueString(),
		true,
		true,
	)
	if err != nil {
		return
	}
}
