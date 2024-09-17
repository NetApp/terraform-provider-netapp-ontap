package security

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &SecurityLoginMessageResource{}
var _ resource.ResourceWithImportState = &SecurityLoginMessageResource{}

// NewSecurityLoginMessageResource is a helper function to simplify the provider implementation.
func NewSecurityLoginMessageResource() resource.Resource {
	return &SecurityLoginMessageResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "security_login_message",
		},
	}
}

// SecurityLoginMessageResource defines the resource implementation.
type SecurityLoginMessageResource struct {
	config connection.ResourceOrDataSourceConfig
}

// SecurityLoginMessageResourceModel describes the resource data model.
type SecurityLoginMessageResourceModel struct {
	CxProfileName      types.String `tfsdk:"cx_profile_name"`
	Banner             types.String `tfsdk:"banner"`
	Message            types.String `tfsdk:"message"`
	ShowClusterMessage types.Bool   `tfsdk:"show_cluster_message"`
	Scope              types.String `tfsdk:"scope"`
	SVMName            types.String `tfsdk:"svm_name"`
	ID                 types.String `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *SecurityLoginMessageResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *SecurityLoginMessageResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityLoginMessages resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "SecurityLoginMessage svm name",
				Optional:            true,
			},
			"message": schema.StringAttribute{
				MarkdownDescription: "SecurityLoginMessage the message of the day (MOTD). This message appears just before the clustershell prompt after a successful login.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"banner": schema.StringAttribute{
				MarkdownDescription: "SecurityLoginMessage login banner",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"show_cluster_message": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether to show a cluster-level message before the SVM message when logging in as an SVM administrator",
				Optional:            true,
				Computed:            true,
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "SecurityLoginMessage network scope",
				Optional:            true,
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "SecurityLoginMessage ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SecurityLoginMessageResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SecurityLoginMessageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityLoginMessageResourceModel

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

	restInfo, err := interfaces.GetSecurityLoginMessage(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSecurityLoginMessage
		return
	}
	// Remove trailing newline characters from Message and Banner since the newline always is added at the end
	cleanMessage := strings.TrimSuffix(restInfo.Message, "\n")
	cleanBanner := strings.TrimSuffix(restInfo.Banner, "\n")
	// Set the values from the response into the data model
	data.Message = types.StringValue(cleanMessage)
	data.Banner = types.StringValue(cleanBanner)
	data.ShowClusterMessage = types.BoolValue(restInfo.ShowClusterMessage)
	data.Scope = types.StringValue(restInfo.Scope)
	data.ID = types.StringValue(restInfo.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *SecurityLoginMessageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SecurityLoginMessageResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Add an error diagnostic indicating that Create is not supported
	resp.Diagnostics.AddError(
		"Create Not Supported",
		"The create operation is not supported for the security_login_message resource. Please import an existing resource instead.",
	)

	tflog.Error(ctx, "Create not supported for resource security_login_message")
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SecurityLoginMessageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *SecurityLoginMessageResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	client, err := connection.GetRestClient(utils.NewErrorHandler(ctx, &resp.Diagnostics), r.config, plan.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	var request interfaces.SecurityLoginMessageResourceBodyDataModelONTAP

	request.Banner = plan.Banner.ValueString()
	request.Message = plan.Message.ValueString()
	request.ShowClusterMessage = plan.ShowClusterMessage.ValueBool()

	if request.Message == "" && request.ShowClusterMessage {
		resp.Diagnostics.AddError("Invalid Input", "Message must be set when ShowClusterMessage is true")
		return
	}
	err = interfaces.UpdateSecurityLoginMessage(errorHandler, *client, state.ID.ValueString(), request)
	if err != nil {
		// error reporting done inside UpdateSecurityLoginMessage
		return
	}
	// Read the updated data from the API
	restInfo, err := interfaces.GetSecurityLoginMessage(errorHandler, *client, plan.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSecurityLoginMessage
		return
	}
	plan.ShowClusterMessage = types.BoolValue(restInfo.ShowClusterMessage)
	plan.Scope = types.StringValue(restInfo.Scope)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SecurityLoginMessageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SecurityLoginMessageResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Add an error diagnostic indicating that Delete is not supported
	resp.Diagnostics.AddError(
		"Delete Not Supported",
		"The update operation is not supported for the security_login_message resource.",
	)

	tflog.Error(ctx, "Delete not supported for resource security_login_message")
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SecurityLoginMessageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req security login message resource: %#v", req))
	idParts := strings.Split(req.ID, ",")

	// import cx_profile only
	if len(idParts) == 1 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[0])...)
		return
	}
	// import svm and cx_profile
	if len(idParts) == 2 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[1])...)
		return
	}

	resp.Diagnostics.AddError(
		"Unexpected Import Identifier",
		fmt.Sprintf("Expected import identifier with format: svm_name,cx_profile_name. Got: %q", req.ID),
	)
}
