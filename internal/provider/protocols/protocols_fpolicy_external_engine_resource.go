package protocols

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsFpolicyExternalEngineResource{}
var _ resource.ResourceWithImportState = &ProtocolsFpolicyExternalEngineResource{}

// NewProtocolsFpolicyExternalEngineResource is a helper function to simplify the provider implementation.
func NewProtocolsFpolicyExternalEngineResource() resource.Resource {
	return &ProtocolsFpolicyExternalEngineResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_fpolicy_external_engine",
		},
	}
}

// ProtocolsFpolicyExternalEngineResource defines the resource implementation.
type ProtocolsFpolicyExternalEngineResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsFpolicyExternalEngineResourceModel describes the resource data model.
type ProtocolsFpolicyExternalEngineResourceModel struct {
	CxProfileName         types.String `tfsdk:"cx_profile_name"`
	Name                  types.String `tfsdk:"name"`
	SVMName               types.String `tfsdk:"svm_name"`
	KeepAliveInterval     types.String `tfsdk:"keep_alive_interval"`
	RequestCancelTimeout  types.String `tfsdk:"request_cancel_timeout"`
	Certificate           types.Object `tfsdk:"certificate"`
	SessionTimeout        types.String `tfsdk:"session_timeout"`
	RequestAbortTimeout   types.String `tfsdk:"request_abort_timeout"`
	StatusRequestInterval types.String `tfsdk:"status_request_interval"`
	SSLOption             types.String `tfsdk:"ssl_option"`
	PrimaryServers        types.List   `tfsdk:"primary_servers"`
	BufferSize            types.Object `tfsdk:"buffer_size"`
	SecondaryServers      types.List   `tfsdk:"secondary_servers"`
	Port                  types.Int64    `tfsdk:"port"`
	ServerProgressTimeout types.String   `tfsdk:"server_progress_timeout"`
	Format                types.String   `tfsdk:"format"`
	Type                  types.String   `tfsdk:"type"`
	Resiliency            types.Object `tfsdk:"resiliency"`
	MaxServerRequests     types.Int64    `tfsdk:"max_server_requests"`
	MaxConnectionRetries  types.Int64    `tfsdk:"max_connection_retries"`
	ID                    types.String `tfsdk:"id"`
}

// ProtocolsFpolicyExternalEngineCertificateResourceModel describes the certificate resource model
type ProtocolsFpolicyExternalEngineCertificateResourceModel struct {
	SerialNumber types.String `tfsdk:"serial_number"`
	Name         types.String `tfsdk:"name"`
	Ca           types.String `tfsdk:"ca"`
}

// ProtocolsFpolicyExternalEngineBufferSizeResourceModel describes the buffer size resource model
type ProtocolsFpolicyExternalEngineBufferSizeResourceModel struct {
	SendBuffer types.Int64 `tfsdk:"send_buffer"`
	RecvBuffer types.Int64 `tfsdk:"recv_buffer"`
}

// ProtocolsFpolicyExternalEngineResiliencyResourceModel describes the resiliency resource model
type ProtocolsFpolicyExternalEngineResiliencyResourceModel struct {
	DirectoryPath     types.String `tfsdk:"directory_path"`
	RetentionDuration types.String `tfsdk:"retention_duration"`
	Enabled           types.Bool   `tfsdk:"enabled"`
}

// Metadata returns the resource type name.
func (r *ProtocolsFpolicyExternalEngineResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsFpolicyExternalEngineResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsFpolicyExternalEngine resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "FPolicy external engine name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "SVM name",
				Required:            true,
			},
			"keep_alive_interval": schema.StringAttribute{
				MarkdownDescription: "Specifies the ISO-8601 interval time for a storage appliance to send Keep Alive message to an FPolicy server",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"request_cancel_timeout": schema.StringAttribute{
				MarkdownDescription: "Specifies the ISO-8601 timeout duration for a screen request to be processed by an FPolicy server",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate": schema.SingleNestedAttribute{
				MarkdownDescription: "Provides details about certificate used to authenticate the FPolicy server",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"serial_number": schema.StringAttribute{
						MarkdownDescription: "Serial number",
						Optional:            true,
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Certificate name",
						Optional:            true,
						Computed:            true,
					},
					"ca": schema.StringAttribute{
						MarkdownDescription: "Certificate authority",
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"session_timeout": schema.StringAttribute{
				MarkdownDescription: "Specifies the interval after which a new session ID is sent to the FPolicy server during reconnection attempts",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"request_abort_timeout": schema.StringAttribute{
				MarkdownDescription: "Specifies the ISO-8601 timeout duration for a screen request to be aborted by a storage appliance",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status_request_interval": schema.StringAttribute{
				MarkdownDescription: "Specifies the ISO-8601 interval time for a storage appliance to query a status request from an FPolicy server",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ssl_option": schema.StringAttribute{
				MarkdownDescription: "The SSL option for external communication with the FPolicy server",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"primary_servers": schema.ListAttribute{
				MarkdownDescription: "The primary FPolicy servers to which the node sends",
				Required:            true,
				ElementType:         types.StringType,
			},
			"buffer_size": schema.SingleNestedAttribute{
				MarkdownDescription: "Specifies the send and receive buffer size of the connected socket for the FPolicy server",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"send_buffer": schema.Int64Attribute{
						MarkdownDescription: "Send buffer size",
						Optional:            true,
						Computed:            true,
					},
					"recv_buffer": schema.Int64Attribute{
						MarkdownDescription: "Receive buffer size",
						Optional:            true,
						Computed:            true,						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},					},
				},
			},
			"secondary_servers": schema.ListAttribute{
				MarkdownDescription: "Send file access events for a given FPolicy policy",
				Optional:            true,
				ElementType:         types.StringType,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port number of the FPolicy server application",
				Required:            true,
			},
			"server_progress_timeout": schema.StringAttribute{
				MarkdownDescription: "Specifies the ISO-8601 timeout duration in which a throttled FPolicy server must complete at least one screen request",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"format": schema.StringAttribute{
				MarkdownDescription: "The format for the notification messages sent to the FPolicy servers",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Determines what ONTAP does after sending notifications to FPolicy servers",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"resiliency": schema.SingleNestedAttribute{
				MarkdownDescription: "When primary and secondary servers are down or unresponsive, file access events are stored in the storage controller under the specified resiliency-directory-path",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"directory_path": schema.StringAttribute{
						MarkdownDescription: "Directory path",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"retention_duration": schema.StringAttribute{
						MarkdownDescription: "Retention duration",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.String{
							stringplanmodifier.UseStateForUnknown(),
						},
					},
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enable resiliency feature",
						Optional:            true,
						Computed:            true,
					},
				},
			},
			"max_server_requests": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of outstanding requests for the FPolicy server",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"max_connection_retries": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of attempts to reconnect to the FPolicy server",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "FPolicy external engine ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsFpolicyExternalEngineResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ProtocolsFpolicyExternalEngineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsFpolicyExternalEngineResourceModel

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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	restInfo, err := interfaces.GetProtocolsFpolicyExternalEngineByName(errorHandler, *client, data.Name.ValueString(), svm.UUID)
	if err != nil {
		// error reporting done inside GetProtocolsFpolicyExternalEngine
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.KeepAliveInterval = types.StringValue(restInfo.KeepAliveInterval)
	data.RequestCancelTimeout = types.StringValue(restInfo.RequestCancelTimeout)
	data.SessionTimeout = types.StringValue(restInfo.SessionTimeout)
	data.RequestAbortTimeout = types.StringValue(restInfo.RequestAbortTimeout)
	data.StatusRequestInterval = types.StringValue(restInfo.StatusRequestInterval)
	data.SSLOption = types.StringValue(restInfo.SSLOption)
	data.Port = types.Int64Value(restInfo.Port)
	data.ServerProgressTimeout = types.StringValue(restInfo.ServerProgressTimeout)
	data.Format = types.StringValue(restInfo.Format)
	data.Type = types.StringValue(restInfo.Type)
	data.MaxServerRequests = types.Int64Value(restInfo.MaxServerRequests)
	data.MaxConnectionRetries = types.Int64Value(restInfo.MaxConnectionRetries)
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", svm.UUID, restInfo.Name))

	// Set certificate using ObjectValue
	certificateElementTypes := map[string]attr.Type{
		"serial_number": types.StringType,
		"name":          types.StringType,
		"ca":            types.StringType,
	}
	certificateElements := map[string]attr.Value{
		"serial_number": types.StringValue(restInfo.Certificate.SerialNumber),
		"name":          types.StringValue(restInfo.Certificate.Name),
		"ca":            types.StringValue(restInfo.Certificate.Ca),
	}
	certificateObjectValue, diags := types.ObjectValue(certificateElementTypes, certificateElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Certificate = certificateObjectValue

	// Set buffer size using ObjectValue
	bufferElementTypes := map[string]attr.Type{
		"send_buffer": types.Int64Type,
		"recv_buffer": types.Int64Type,
	}
	bufferElements := map[string]attr.Value{
		"send_buffer": types.Int64Value(int64(restInfo.BufferSize.SendBuffer)),
		"recv_buffer": types.Int64Value(int64(restInfo.BufferSize.RecvBuffer)),
	}
	bufferObjectValue, diags := types.ObjectValue(bufferElementTypes, bufferElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.BufferSize = bufferObjectValue

	// Set resiliency using ObjectValue
	resiliencyElementTypes := map[string]attr.Type{
		"directory_path":     types.StringType,
		"retention_duration": types.StringType,
		"enabled":            types.BoolType,
	}
	resiliencyElements := map[string]attr.Value{
		"directory_path":     types.StringValue(restInfo.Resiliency.DirectoryPath),
		"retention_duration": types.StringValue(restInfo.Resiliency.RetentionDuration),
		"enabled":            types.BoolValue(restInfo.Resiliency.Enabled),
	}
	resiliencyObjectValue, diags := types.ObjectValue(resiliencyElementTypes, resiliencyElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Resiliency = resiliencyObjectValue

	// Set primary servers
	if len(restInfo.PrimaryServers) > 0 {
		data.PrimaryServers, _ = types.ListValueFrom(ctx, types.StringType, restInfo.PrimaryServers)
	}

	// Set secondary servers
	if len(restInfo.SecondaryServers) > 0 {
		data.SecondaryServers, _ = types.ListValueFrom(ctx, types.StringType, restInfo.SecondaryServers)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *ProtocolsFpolicyExternalEngineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProtocolsFpolicyExternalEngineResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ProtocolsFpolicyExternalEngineResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	body.Name = data.Name.ValueString()
	if !data.KeepAliveInterval.IsNull() {
		body.KeepAliveInterval = data.KeepAliveInterval.ValueString()
	}
	if !data.RequestCancelTimeout.IsNull() {
		body.RequestCancelTimeout = data.RequestCancelTimeout.ValueString()
	}
	if !data.SessionTimeout.IsNull() {
		body.SessionTimeout = data.SessionTimeout.ValueString()
	}
	if !data.RequestAbortTimeout.IsNull() {
		body.RequestAbortTimeout = data.RequestAbortTimeout.ValueString()
	}
	if !data.StatusRequestInterval.IsNull() {
		body.StatusRequestInterval = data.StatusRequestInterval.ValueString()
	}
	if !data.SSLOption.IsNull() {
		body.SSLOption = data.SSLOption.ValueString()
	}
	if !data.Port.IsNull() {
		body.Port = data.Port.ValueInt64()
	}
	if !data.ServerProgressTimeout.IsNull() {
		body.ServerProgressTimeout = data.ServerProgressTimeout.ValueString()
	}
	if !data.Format.IsNull() {
		body.Format = data.Format.ValueString()
	}
	if !data.Type.IsNull() {
		body.Type = data.Type.ValueString()
	}
	if !data.MaxServerRequests.IsNull() {
		body.MaxServerRequests = data.MaxServerRequests.ValueInt64()
	}
	if !data.MaxConnectionRetries.IsNull() {
		body.MaxConnectionRetries = data.MaxConnectionRetries.ValueInt64()
	}

	// Set certificate only if provided
	if !data.Certificate.IsNull() && !data.Certificate.IsUnknown() {
		var certificate ProtocolsFpolicyExternalEngineCertificateResourceModel
		diags := data.Certificate.As(ctx, &certificate, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.Certificate = interfaces.ProtocolsFpolicyExternalEngineCertificate{
			SerialNumber: certificate.SerialNumber.ValueString(),
			Name:         certificate.Name.ValueString(),
			Ca:           certificate.Ca.ValueString(),
		}
	}

	// Set buffer size only if provided
	if !data.BufferSize.IsNull() && !data.BufferSize.IsUnknown() {
		var bufferSize ProtocolsFpolicyExternalEngineBufferSizeResourceModel
		diags := data.BufferSize.As(ctx, &bufferSize, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.BufferSize = interfaces.ProtocolsFpolicyExternalEngineBufferSize{
			SendBuffer: int(bufferSize.SendBuffer.ValueInt64()),
			RecvBuffer: int(bufferSize.RecvBuffer.ValueInt64()),
		}
	}

	// Set resiliency - handle explicit configuration including enabled=false
	if !data.Resiliency.IsNull() && !data.Resiliency.IsUnknown() {
		var resiliency ProtocolsFpolicyExternalEngineResiliencyResourceModel
		diags := data.Resiliency.As(ctx, &resiliency, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		
		// Always send resiliency configuration if explicitly configured
		body.Resiliency = interfaces.ProtocolsFpolicyExternalEngineResiliency{
			DirectoryPath:     resiliency.DirectoryPath.ValueString(),
			RetentionDuration: resiliency.RetentionDuration.ValueString(),
			Enabled:           resiliency.Enabled.ValueBool(),
		}
		
		// Special handling for enabled=false - ensure directory_path is provided
		// ONTAP requires directory_path when any resiliency config is sent
		if !resiliency.Enabled.ValueBool() && resiliency.DirectoryPath.IsNull() {
			// Set a default path when disabling resiliency
			body.Resiliency.DirectoryPath = ""
		}
	}

	// Set primary servers
	if !data.PrimaryServers.IsNull() {
		var primaryServers []string
		resp.Diagnostics.Append(data.PrimaryServers.ElementsAs(ctx, &primaryServers, false)...)
		body.PrimaryServers = primaryServers
	}

	// Set secondary servers
	if !data.SecondaryServers.IsNull() {
		var secondaryServers []string
		resp.Diagnostics.Append(data.SecondaryServers.ElementsAs(ctx, &secondaryServers, false)...)
		body.SecondaryServers = secondaryServers
	}

	resource, err := interfaces.CreateProtocolsFpolicyExternalEngine(errorHandler, *client, body, svm.UUID)
	if err != nil {
		return
	}

	// Read the resource back after creation to get computed fields
	restInfo, err := interfaces.GetProtocolsFpolicyExternalEngineByName(errorHandler, *client, resource.Name, svm.UUID)
	if err != nil {
		// error reporting done inside GetProtocolsFpolicyExternalEngine
		return
	}

	// Update all fields with current values from ONTAP
	data.Name = types.StringValue(restInfo.Name)
	data.KeepAliveInterval = types.StringValue(restInfo.KeepAliveInterval)
	data.RequestCancelTimeout = types.StringValue(restInfo.RequestCancelTimeout)
	data.SessionTimeout = types.StringValue(restInfo.SessionTimeout)
	data.RequestAbortTimeout = types.StringValue(restInfo.RequestAbortTimeout)
	data.StatusRequestInterval = types.StringValue(restInfo.StatusRequestInterval)
	data.SSLOption = types.StringValue(restInfo.SSLOption)
	data.Port = types.Int64Value(restInfo.Port)
	data.ServerProgressTimeout = types.StringValue(restInfo.ServerProgressTimeout)
	data.Format = types.StringValue(restInfo.Format)
	data.Type = types.StringValue(restInfo.Type)
	data.MaxServerRequests = types.Int64Value(restInfo.MaxServerRequests)
	data.MaxConnectionRetries = types.Int64Value(restInfo.MaxConnectionRetries)
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", svm.UUID, restInfo.Name))

	// Set certificate using ObjectValue
	certificateElementTypes := map[string]attr.Type{
		"serial_number": types.StringType,
		"name":          types.StringType,
		"ca":            types.StringType,
	}
	certificateElements := map[string]attr.Value{
		"serial_number": types.StringValue(restInfo.Certificate.SerialNumber),
		"name":          types.StringValue(restInfo.Certificate.Name),
		"ca":            types.StringValue(restInfo.Certificate.Ca),
	}
	certificateObjectValue, diags := types.ObjectValue(certificateElementTypes, certificateElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Certificate = certificateObjectValue

	// Set buffer size using ObjectValue
	bufferElementTypes := map[string]attr.Type{
		"send_buffer": types.Int64Type,
		"recv_buffer": types.Int64Type,
	}
	bufferElements := map[string]attr.Value{
		"send_buffer": types.Int64Value(int64(restInfo.BufferSize.SendBuffer)),
		"recv_buffer": types.Int64Value(int64(restInfo.BufferSize.RecvBuffer)),
	}
	bufferObjectValue, diagsBuffer := types.ObjectValue(bufferElementTypes, bufferElements)
	if diagsBuffer.HasError() {
		resp.Diagnostics.Append(diagsBuffer...)
		return
	}
	data.BufferSize = bufferObjectValue

	// Set resiliency using ObjectValue
	resiliencyElementTypes := map[string]attr.Type{
		"directory_path":     types.StringType,
		"retention_duration": types.StringType,
		"enabled":            types.BoolType,
	}
	resiliencyElements := map[string]attr.Value{
		"directory_path":     types.StringValue(restInfo.Resiliency.DirectoryPath),
		"retention_duration": types.StringValue(restInfo.Resiliency.RetentionDuration),
		"enabled":            types.BoolValue(restInfo.Resiliency.Enabled),
	}
	resiliencyObjectValue, diagsResiliency := types.ObjectValue(resiliencyElementTypes, resiliencyElements)
	if diagsResiliency.HasError() {
		resp.Diagnostics.Append(diagsResiliency...)
		return
	}
	data.Resiliency = resiliencyObjectValue

	// Set primary servers
	if restInfo.PrimaryServers != nil {
		primaryServersList, diagsPrimary := types.ListValueFrom(ctx, types.StringType, restInfo.PrimaryServers)
		if diagsPrimary.HasError() {
			resp.Diagnostics.Append(diagsPrimary...)
			return
		}
		data.PrimaryServers = primaryServersList
	}

	// Set secondary servers
	if restInfo.SecondaryServers != nil {
		secondaryServersList, diagsSecondary := types.ListValueFrom(ctx, types.StringType, restInfo.SecondaryServers)
		if diagsSecondary.HasError() {
			resp.Diagnostics.Append(diagsSecondary...)
			return
		}
		data.SecondaryServers = secondaryServersList
	} else {
		data.SecondaryServers = types.ListNull(types.StringType)
	}

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsFpolicyExternalEngineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ProtocolsFpolicyExternalEngineResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	var body interfaces.ProtocolsFpolicyExternalEngineResourceBodyDataModelONTAP

	if !data.KeepAliveInterval.IsNull() {
		body.KeepAliveInterval = data.KeepAliveInterval.ValueString()
	}
	if !data.RequestCancelTimeout.IsNull() {
		body.RequestCancelTimeout = data.RequestCancelTimeout.ValueString()
	}
	if !data.SessionTimeout.IsNull() {
		body.SessionTimeout = data.SessionTimeout.ValueString()
	}
	if !data.RequestAbortTimeout.IsNull() {
		body.RequestAbortTimeout = data.RequestAbortTimeout.ValueString()
	}
	if !data.StatusRequestInterval.IsNull() {
		body.StatusRequestInterval = data.StatusRequestInterval.ValueString()
	}
	if !data.SSLOption.IsNull() {
		body.SSLOption = data.SSLOption.ValueString()
	}
	if !data.Port.IsNull() {
		body.Port = data.Port.ValueInt64()
	}
	if !data.ServerProgressTimeout.IsNull() {
		body.ServerProgressTimeout = data.ServerProgressTimeout.ValueString()
	}
	if !data.Format.IsNull() {
		body.Format = data.Format.ValueString()
	}
	if !data.Type.IsNull() {
		body.Type = data.Type.ValueString()
	}
	if !data.MaxServerRequests.IsNull() {
		body.MaxServerRequests = data.MaxServerRequests.ValueInt64()
	}
	if !data.MaxConnectionRetries.IsNull() {
		body.MaxConnectionRetries = data.MaxConnectionRetries.ValueInt64()
	}

	// Set certificate only if provided
	if !data.Certificate.IsNull() && !data.Certificate.IsUnknown() {
		var certificate ProtocolsFpolicyExternalEngineCertificateResourceModel
		diags := data.Certificate.As(ctx, &certificate, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.Certificate = interfaces.ProtocolsFpolicyExternalEngineCertificate{
			SerialNumber: certificate.SerialNumber.ValueString(),
			Name:         certificate.Name.ValueString(),
			Ca:           certificate.Ca.ValueString(),
		}
	}

	// Set buffer size only if provided
	if !data.BufferSize.IsNull() && !data.BufferSize.IsUnknown() {
		var bufferSize ProtocolsFpolicyExternalEngineBufferSizeResourceModel
		diags := data.BufferSize.As(ctx, &bufferSize, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		body.BufferSize = interfaces.ProtocolsFpolicyExternalEngineBufferSize{
			SendBuffer: int(bufferSize.SendBuffer.ValueInt64()),
			RecvBuffer: int(bufferSize.RecvBuffer.ValueInt64()),
		}
	}

	// Set resiliency - handle explicit configuration including enabled=false
	if !data.Resiliency.IsNull() && !data.Resiliency.IsUnknown() {
		var resiliency ProtocolsFpolicyExternalEngineResiliencyResourceModel
		diags := data.Resiliency.As(ctx, &resiliency, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		
		// Always send resiliency configuration if explicitly configured
		body.Resiliency = interfaces.ProtocolsFpolicyExternalEngineResiliency{
			DirectoryPath:     resiliency.DirectoryPath.ValueString(),
			RetentionDuration: resiliency.RetentionDuration.ValueString(),
			Enabled:           resiliency.Enabled.ValueBool(),
		}
		
		// Special handling for enabled=false - ensure directory_path is provided
		// ONTAP requires directory_path when any resiliency config is sent
		if !resiliency.Enabled.ValueBool() && resiliency.DirectoryPath.IsNull() {
			// Set a default path when disabling resiliency
			body.Resiliency.DirectoryPath = ""
		}
	}

	// Set primary servers
	if !data.PrimaryServers.IsNull() {
		var primaryServers []string
		resp.Diagnostics.Append(data.PrimaryServers.ElementsAs(ctx, &primaryServers, false)...)
		body.PrimaryServers = primaryServers
	}

	// Set secondary servers
	if !data.SecondaryServers.IsNull() {
		var secondaryServers []string
		resp.Diagnostics.Append(data.SecondaryServers.ElementsAs(ctx, &secondaryServers, false)...)
		body.SecondaryServers = secondaryServers
	}

	err = interfaces.UpdateProtocolsFpolicyExternalEngine(errorHandler, *client, body, svm.UUID, data.Name.ValueString())
	if err != nil {
		return
	}

	// Read the resource back after update to get computed fields
	restInfo, err := interfaces.GetProtocolsFpolicyExternalEngineByName(errorHandler, *client, data.Name.ValueString(), svm.UUID)
	if err != nil {
		// error reporting done inside GetProtocolsFpolicyExternalEngine
		return
	}

	// Update all fields with current values from ONTAP
	data.Name = types.StringValue(restInfo.Name)
	data.KeepAliveInterval = types.StringValue(restInfo.KeepAliveInterval)
	data.RequestCancelTimeout = types.StringValue(restInfo.RequestCancelTimeout)
	data.SessionTimeout = types.StringValue(restInfo.SessionTimeout)
	data.RequestAbortTimeout = types.StringValue(restInfo.RequestAbortTimeout)
	data.StatusRequestInterval = types.StringValue(restInfo.StatusRequestInterval)
	data.SSLOption = types.StringValue(restInfo.SSLOption)
	data.Port = types.Int64Value(restInfo.Port)
	data.ServerProgressTimeout = types.StringValue(restInfo.ServerProgressTimeout)
	data.Format = types.StringValue(restInfo.Format)
	data.Type = types.StringValue(restInfo.Type)
	data.MaxServerRequests = types.Int64Value(restInfo.MaxServerRequests)
	data.MaxConnectionRetries = types.Int64Value(restInfo.MaxConnectionRetries)
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", svm.UUID, restInfo.Name))

	// Set certificate using ObjectValue
	certificateElementTypes := map[string]attr.Type{
		"serial_number": types.StringType,
		"name":          types.StringType,
		"ca":            types.StringType,
	}
	certificateElements := map[string]attr.Value{
		"serial_number": types.StringValue(restInfo.Certificate.SerialNumber),
		"name":          types.StringValue(restInfo.Certificate.Name),
		"ca":            types.StringValue(restInfo.Certificate.Ca),
	}
	certificateObjectValue, diags := types.ObjectValue(certificateElementTypes, certificateElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Certificate = certificateObjectValue

	// Set buffer size using ObjectValue
	bufferElementTypes := map[string]attr.Type{
		"send_buffer": types.Int64Type,
		"recv_buffer": types.Int64Type,
	}
	bufferElements := map[string]attr.Value{
		"send_buffer": types.Int64Value(int64(restInfo.BufferSize.SendBuffer)),
		"recv_buffer": types.Int64Value(int64(restInfo.BufferSize.RecvBuffer)),
	}
	bufferObjectValue, diagsBuffer := types.ObjectValue(bufferElementTypes, bufferElements)
	if diagsBuffer.HasError() {
		resp.Diagnostics.Append(diagsBuffer...)
		return
	}
	data.BufferSize = bufferObjectValue

	// Set resiliency using ObjectValue
	resiliencyElementTypes := map[string]attr.Type{
		"directory_path":     types.StringType,
		"retention_duration": types.StringType,
		"enabled":            types.BoolType,
	}
	resiliencyElements := map[string]attr.Value{
		"directory_path":     types.StringValue(restInfo.Resiliency.DirectoryPath),
		"retention_duration": types.StringValue(restInfo.Resiliency.RetentionDuration),
		"enabled":            types.BoolValue(restInfo.Resiliency.Enabled),
	}
	resiliencyObjectValue, diagsResiliency := types.ObjectValue(resiliencyElementTypes, resiliencyElements)
	if diagsResiliency.HasError() {
		resp.Diagnostics.Append(diagsResiliency...)
		return
	}
	data.Resiliency = resiliencyObjectValue

	// Set primary servers
	if restInfo.PrimaryServers != nil {
		primaryServersList, diagsPrimary := types.ListValueFrom(ctx, types.StringType, restInfo.PrimaryServers)
		if diagsPrimary.HasError() {
			resp.Diagnostics.Append(diagsPrimary...)
			return
		}
		data.PrimaryServers = primaryServersList
	}

	// Set secondary servers
	if restInfo.SecondaryServers != nil {
		secondaryServersList, diagsSecondary := types.ListValueFrom(ctx, types.StringType, restInfo.SecondaryServers)
		if diagsSecondary.HasError() {
			resp.Diagnostics.Append(diagsSecondary...)
			return
		}
		data.SecondaryServers = secondaryServersList
	} else {
		data.SecondaryServers = types.ListNull(types.StringType)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsFpolicyExternalEngineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsFpolicyExternalEngineResourceModel

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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	err = interfaces.DeleteProtocolsFpolicyExternalEngine(errorHandler, *client, svm.UUID, data.Name.ValueString())
	if err != nil {
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsFpolicyExternalEngineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}
	
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
}
