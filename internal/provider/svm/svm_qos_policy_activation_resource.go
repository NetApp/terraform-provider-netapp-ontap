package svm

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
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
	_ resource.Resource                = &SvmQosPolicyActivationResource{}
	_ resource.ResourceWithConfigure   = &SvmQosPolicyActivationResource{}
	_ resource.ResourceWithImportState = &SvmQosPolicyActivationResource{}
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
	CxProfileName types.String                                  `tfsdk:"cx_profile_name"`
	Svm           *SvmQosPolicyActivationSvmResourceModel       `tfsdk:"svm"`
	QoSPolicy     *SvmQosPolicyActivationQosPolicyResourceModel `tfsdk:"qos_policy"`
	ID            types.String                                  `tfsdk:"id"`
}

type SvmQosPolicyActivationSvmResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type SvmQosPolicyActivationQosPolicyResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
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
			"svm": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "The name of the SVM",
						Optional:            true,
						Computed:            true,
					},
					"id": schema.StringAttribute{
						MarkdownDescription: "The unique identifier of the SVM",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
				},
				MarkdownDescription: "SVM to manage (name or UUID)",
				Required:            true,
				PlanModifiers: []planmodifier.Object{
					objectplanmodifier.RequiresReplace(),
				},
			},
			"qos_policy": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "The QoS policy group name",
						Optional:            true,
						Computed:            true,
					},
					"id": schema.StringAttribute{
						MarkdownDescription: "The QoS policy group UUID",
						Optional:            true,
						Computed:            true,
					},
				},
				MarkdownDescription: "QoS policy group to apply to the SVM (name or UUID)",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the SVM QoS Policy Activation",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("read a svm resource: %#v", data))

	// If no svm ID specified...
	if data.Svm.ID.ValueString() == "" {

		// ...call ONTAP REST API for reading svm info
		svm, err := interfaces.GetSvmByNameDataSource(
			errorHandler,
			*client,
			data.Svm.Name.ValueString(),
			cluster.Version,
		)
		if err != nil {
			return
		}

		// Copy svm info to resource model
		data.Svm.ID = types.StringValue(svm.UUID)
	}

	// Call ONTAP REST API for reading svm info
	svm, err := interfaces.GetSvm(
		errorHandler,
		*client,
		data.Svm.ID.ValueString(),
	)
	if err != nil {
		return
	}

	// Copy svm info to resource model
	data.Svm.Name = types.StringValue(svm.Name)
	if data.QoSPolicy == nil {
		data.QoSPolicy = &SvmQosPolicyActivationQosPolicyResourceModel{}
	}
	data.QoSPolicy.ID = types.StringValue(svm.QoSPolicy.UUID)
	data.ID = types.StringValue(svm.QoSPolicy.UUID)
	data.QoSPolicy.Name = types.StringValue(svm.QoSPolicy.Name)

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

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	// If no svm ID specified...
	if data.Svm.ID.ValueString() == "" {

		// ...call ONTAP REST API for reading svm info
		svm, err := interfaces.GetSvmByNameDataSource(
			errorHandler,
			*client,
			data.Svm.Name.ValueString(),
			cluster.Version,
		)
		if err != nil {
			return
		}

		// Copy svm info to resource model
		data.Svm.ID = types.StringValue(svm.UUID)
	}

	// Copy QoS policy info to request body
	var body interfaces.SvmResourceModel
	body.QoSPolicy.UUID = data.QoSPolicy.ID.ValueString()
	body.QoSPolicy.Name = data.QoSPolicy.Name.ValueString()

	// Call ONTAP REST API for updating svm
	tflog.Debug(ctx, fmt.Sprintf("update a svm resource: %#v", data))
	err = interfaces.UpdateSvm(
		errorHandler,
		*client,
		body,
		data.Svm.ID.ValueString(),
		true,
		true,
	)
	if err != nil {
		return
	}

	// Call ONTAP REST API again for reading svm info
	svm, err := interfaces.GetSvm(
		errorHandler,
		*client,
		data.Svm.ID.ValueString(),
	)
	if err != nil {
		return
	}

	// Copy remaining svm info to resource model
	data.Svm.Name = types.StringValue(svm.Name)
	data.QoSPolicy.ID = types.StringValue(svm.QoSPolicy.UUID)
	data.ID = types.StringValue(svm.QoSPolicy.UUID)
	data.QoSPolicy.Name = types.StringValue(svm.QoSPolicy.Name)

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
	body.QoSPolicy.UUID = data.QoSPolicy.ID.ValueString()
	body.QoSPolicy.Name = data.QoSPolicy.Name.ValueString()

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
		data.Svm.ID.ValueString(),
		true,
		true,
	)
	if err != nil {
		return
	}

	// Call ONTAP REST API again for reading svm info
	svm, err := interfaces.GetSvm(
		errorHandler,
		*client,
		data.Svm.ID.ValueString(),
	)
	if err != nil {
		return
	}

	// Copy remaining svm info to resource model
	data.Svm.Name = types.StringValue(svm.Name)
	data.QoSPolicy.ID = types.StringValue(svm.QoSPolicy.UUID)
	data.ID = types.StringValue(svm.QoSPolicy.UUID)
	data.QoSPolicy.Name = types.StringValue(svm.QoSPolicy.Name)

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
		data.Svm.ID.ValueString(),
		true,
		true,
	)
	if err != nil {
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SvmQosPolicyActivationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm").AtName("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[1])...)
}
