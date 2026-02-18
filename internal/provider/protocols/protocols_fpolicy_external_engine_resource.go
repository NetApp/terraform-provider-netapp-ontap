package protocols

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
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

// normalizeDuration normalizes ISO8601 duration strings to match ONTAP's preferred format.
// ONTAP returns durations like "PT1M" for "PT60S", "PT1M30S" for "PT90S", etc.
func normalizeDuration(duration string) string {
	if duration == "" {
		return ""
	}

	// Match ISO8601 duration pattern: PT{hours}H{minutes}M{seconds}S
	// ONTAP uses PT format
	re := regexp.MustCompile(`^PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?$`)
	matches := re.FindStringSubmatch(duration)
	if matches == nil {
		// If it doesn't match, return as-is
		return duration
	}

	// Extract hours, minutes, seconds
	hours := 0
	minutes := 0
	seconds := 0

	if matches[1] != "" {
		hours, _ = strconv.Atoi(matches[1])
	}
	if matches[2] != "" {
		minutes, _ = strconv.Atoi(matches[2])
	}
	if matches[3] != "" {
		seconds, _ = strconv.Atoi(strings.Split(matches[3], ".")[0]) // Ignore fractional seconds for simplicity
	}

	// Convert everything to total seconds
	totalSeconds := hours*3600 + minutes*60 + seconds

	// Convert back to ONTAP's preferred format
	if totalSeconds == 0 {
		return "PT0S"
	}

	result := "PT"
	remainingSeconds := totalSeconds

	// Add hours if >= 1 hour
	if remainingSeconds >= 3600 {
		hours := remainingSeconds / 3600
		result += fmt.Sprintf("%dH", hours)
		remainingSeconds %= 3600
	}

	// Add minutes if >= 1 minute
	if remainingSeconds >= 60 {
		minutes := remainingSeconds / 60
		result += fmt.Sprintf("%dM", minutes)
		remainingSeconds %= 60
	}

	// Add remaining seconds if any
	if remainingSeconds > 0 {
		result += fmt.Sprintf("%dS", remainingSeconds)
	}

	return result
}

// compareDuration checks if two duration strings are equivalent, handling different normalization formats
func compareDuration(duration1, duration2 string) bool {
	if duration1 == duration2 {
		return true
	}
	// Both should normalize to the same value
	return normalizeDuration(duration1) == normalizeDuration(duration2)
}

// durationNormalizePlanModifier implements the plan modifier.
type durationNormalizePlanModifier struct{}

// Description returns a human-readable description of the plan modifier.
func (m durationNormalizePlanModifier) Description(_ context.Context) string {
	return "Normalizes duration values to match ONTAP format"
}

// MarkdownDescription returns a markdown description of the plan modifier.
func (m durationNormalizePlanModifier) MarkdownDescription(_ context.Context) string {
	return "Normalizes duration values to match ONTAP format"
}

// PlanModifyString implements the plan modification logic.
func (m durationNormalizePlanModifier) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	// This plan modifier preserves semantic equivalence between user config and ONTAP normalized values
	// It doesn't modify values, but helps with diff suppression in future versions
	// For now, we handle equivalence in the CRUD functions
}

// DurationNormalize returns a plan modifier that normalizes duration values to match ONTAP format.
func DurationNormalize() planmodifier.String {
	return durationNormalizePlanModifier{}
}

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsFpolicyExternalEngineResource{}
var _ resource.ResourceWithImportState = &ProtocolsFpolicyExternalEngineResource{}

// NewProtocolsFpolicyExternalEngineResource is a helper function to simplify the provider implementation.
func NewProtocolsFpolicyExternalEngineResource() resource.Resource {
	return &ProtocolsFpolicyExternalEngineResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "fpolicy_external_engine",
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
	PrimaryServers        types.Set   `tfsdk:"primary_servers"`
	BufferSize            types.Object `tfsdk:"buffer_size"`
	SecondaryServers      types.Set   `tfsdk:"secondary_servers"`
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
					DurationNormalize(),
				},
			},
			"request_cancel_timeout": schema.StringAttribute{
				MarkdownDescription: "Specifies the ISO-8601 timeout duration for a screen request to be processed by an FPolicy server",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					DurationNormalize(),
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
					DurationNormalize(),
				},
			},
			"request_abort_timeout": schema.StringAttribute{
				MarkdownDescription: "Specifies the ISO-8601 timeout duration for a screen request to be aborted by a storage appliance",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					DurationNormalize(),
				},
			},
			"status_request_interval": schema.StringAttribute{
				MarkdownDescription: "Specifies the ISO-8601 interval time for a storage appliance to query a status request from an FPolicy server",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					DurationNormalize(),
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
			"primary_servers": schema.SetAttribute{
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
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
					"recv_buffer": schema.Int64Attribute{
						MarkdownDescription: "Receive buffer size",
						Optional:            true,
						Computed:            true,
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
						},
					},
				},
			},
			"secondary_servers": schema.SetAttribute{
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
					DurationNormalize(),
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
	// For duration fields, preserve config values if they exist, otherwise use normalized ONTAP values
	if data.KeepAliveInterval.IsNull() {
		if restInfo.KeepAliveInterval != "" {
			data.KeepAliveInterval = types.StringValue(restInfo.KeepAliveInterval)
		} else {
			data.KeepAliveInterval = types.StringValue("PT2M") // ONTAP default
		}
	}
	if data.RequestCancelTimeout.IsNull() {
		if restInfo.RequestCancelTimeout != "" {
			data.RequestCancelTimeout = types.StringValue(restInfo.RequestCancelTimeout)
		} else {
			data.RequestCancelTimeout = types.StringValue("PT1M") // ONTAP default
		}
	}
	if data.SessionTimeout.IsNull() {
		if restInfo.SessionTimeout != "" {
			data.SessionTimeout = types.StringValue(restInfo.SessionTimeout)
		} else {
			data.SessionTimeout = types.StringValue("PT10S") // ONTAP default
		}
	}
	if data.RequestAbortTimeout.IsNull() {
		if restInfo.RequestAbortTimeout != "" {
			data.RequestAbortTimeout = types.StringValue(restInfo.RequestAbortTimeout)
		} else {
			data.RequestAbortTimeout = types.StringValue("PT40S") // ONTAP default
		}
	}
	if data.StatusRequestInterval.IsNull() {
		if restInfo.StatusRequestInterval != "" {
			data.StatusRequestInterval = types.StringValue(restInfo.StatusRequestInterval)
		} else {
			data.StatusRequestInterval = types.StringValue("PT10S") // ONTAP default
		}
	}
	data.SSLOption = types.StringValue(restInfo.SSLOption)
	data.Port = types.Int64Value(restInfo.Port)
	if data.ServerProgressTimeout.IsNull() {
		if restInfo.ServerProgressTimeout != "" {
			data.ServerProgressTimeout = types.StringValue(restInfo.ServerProgressTimeout)
		} else {
			data.ServerProgressTimeout = types.StringValue("PT1M") // ONTAP default
		}
	}
	data.Format = types.StringValue(restInfo.Format)
	data.Type = types.StringValue(restInfo.Type)
	data.MaxServerRequests = types.Int64Value(restInfo.MaxServerRequests)
	data.MaxConnectionRetries = types.Int64Value(restInfo.MaxConnectionRetries)
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", svm.UUID, restInfo.Name))

	// Set certificate - check if config originally had it by checking if current state is not null
	if !data.Certificate.IsNull() || (restInfo.Certificate.Name != "" || restInfo.Certificate.SerialNumber != "" || restInfo.Certificate.Ca != "") {
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
	} else {
		// No meaningful certificate values, set as null
		data.Certificate = types.ObjectNull(map[string]attr.Type{
			"serial_number": types.StringType,
			"name":          types.StringType,
			"ca":            types.StringType,
		})
	}

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

	// Set resiliency - check if config originally had it by checking if current state is not null
	if !data.Resiliency.IsNull() || (restInfo.Resiliency.Enabled || restInfo.Resiliency.DirectoryPath != "" || restInfo.Resiliency.RetentionDuration != "") {
		resiliencyElementTypes := map[string]attr.Type{
			"directory_path":     types.StringType,
			"retention_duration": types.StringType,
			"enabled":            types.BoolType,
		}

		// Preserve existing retention_duration from state to avoid provider inconsistency
		currentRetentionDuration := func() string {
			if !data.Resiliency.IsNull() && !data.Resiliency.IsUnknown() {
				var currentResiliency ProtocolsFpolicyExternalEngineResiliencyResourceModel
				diags := data.Resiliency.As(ctx, &currentResiliency, basetypes.ObjectAsOptions{})
				if !diags.HasError() && !currentResiliency.RetentionDuration.IsNull() && !currentResiliency.RetentionDuration.IsUnknown() {
					return currentResiliency.RetentionDuration.ValueString()
				}
			}
			return restInfo.Resiliency.RetentionDuration
		}()

		resiliencyElements := map[string]attr.Value{
			"directory_path":     types.StringValue(restInfo.Resiliency.DirectoryPath),
			"retention_duration": types.StringValue(currentRetentionDuration),
			"enabled":            types.BoolValue(restInfo.Resiliency.Enabled),
		}
		resiliencyObjectValue, diags := types.ObjectValue(resiliencyElementTypes, resiliencyElements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.Resiliency = resiliencyObjectValue
	} else {
		// No meaningful resiliency values, set as null
		data.Resiliency = types.ObjectNull(map[string]attr.Type{
			"directory_path":     types.StringType,
			"retention_duration": types.StringType,
			"enabled":            types.BoolType,
		})
	}

	// Set primary servers
	if len(restInfo.PrimaryServers) > 0 {
		var primaryElements []attr.Value
		for _, server := range restInfo.PrimaryServers {
			primaryElements = append(primaryElements, types.StringValue(server))
		}
		primarySet, diags := types.SetValue(types.StringType, primaryElements)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.PrimaryServers = primarySet
	}

	// Set secondary servers
	if len(restInfo.SecondaryServers) > 0 {
		var secondaryElements []attr.Value
		for _, server := range restInfo.SecondaryServers {
			secondaryElements = append(secondaryElements, types.StringValue(server))
		}
		secondarySet, diags := types.SetValue(types.StringType, secondaryElements)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.SecondaryServers = secondarySet
	} else if !data.SecondaryServers.IsNull() {
		// Only set to empty set if secondary_servers was explicitly provided in config
		emptySet, diags := types.SetValue(types.StringType, []attr.Value{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.SecondaryServers = emptySet
	}
	// If data.SecondaryServers.IsNull(), leave it as null (not provided in config)

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

	// Normalize duration values to match ONTAP's expected format BEFORE sending to ONTAP
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
		body.KeepAliveInterval = normalizeDuration(data.KeepAliveInterval.ValueString())
	}
	if !data.RequestCancelTimeout.IsNull() {
		body.RequestCancelTimeout = normalizeDuration(data.RequestCancelTimeout.ValueString())
	}
	if !data.SessionTimeout.IsNull() {
		body.SessionTimeout = normalizeDuration(data.SessionTimeout.ValueString())
	}
	if !data.RequestAbortTimeout.IsNull() {
		body.RequestAbortTimeout = normalizeDuration(data.RequestAbortTimeout.ValueString())
	}
	if !data.StatusRequestInterval.IsNull() {
		body.StatusRequestInterval = normalizeDuration(data.StatusRequestInterval.ValueString())
	}
	if !data.SSLOption.IsNull() {
		body.SSLOption = data.SSLOption.ValueString()
	}
	if !data.Port.IsNull() {
		body.Port = data.Port.ValueInt64()
	}
	if !data.ServerProgressTimeout.IsNull() {
		body.ServerProgressTimeout = normalizeDuration(data.ServerProgressTimeout.ValueString())
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
			SendBuffer: bufferSize.SendBuffer.ValueInt64(),
			RecvBuffer: bufferSize.RecvBuffer.ValueInt64(),
		}
	}


	// Store planned resiliency retention_duration to preserve in state after create
	var plannedRetentionDuration string
	if !data.Resiliency.IsNull() && !data.Resiliency.IsUnknown() {
		var resiliencyPlan ProtocolsFpolicyExternalEngineResiliencyResourceModel
		diags := data.Resiliency.As(ctx, &resiliencyPlan, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		plannedRetentionDuration = resiliencyPlan.RetentionDuration.ValueString()
	}
	// Set resiliency - handle explicit configuration including enabled=false
	if !data.Resiliency.IsNull() && !data.Resiliency.IsUnknown() {
		var resiliency ProtocolsFpolicyExternalEngineResiliencyResourceModel
		diags := data.Resiliency.As(ctx, &resiliency, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// Validate that if resiliency block is provided, enabled must be true
		if resiliency.Enabled.IsNull() || !resiliency.Enabled.ValueBool() {
			resp.Diagnostics.AddError(
				"Invalid resiliency configuration",
				"ONTAP does not support resiliency with enabled=false. When resiliency block is provided, 'enabled' must be true. To disable resiliency, remove the entire resiliency block.",
			)
			return
		}

		// Only send resiliency configuration if enabled=true (ONTAP requirement)
		if resiliency.Enabled.ValueBool() {
			body.Resiliency = interfaces.ProtocolsFpolicyExternalEngineResiliency{
				DirectoryPath:     resiliency.DirectoryPath.ValueString(),
				RetentionDuration: resiliency.RetentionDuration.ValueString(),
				Enabled:           resiliency.Enabled.ValueBool(),
			}
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

	// Update all fields with current values from ONTAP, but preserve planned duration values
	data.Name = types.StringValue(restInfo.Name)
	// For duration fields, keep the planned config values to prevent provider inconsistency
	// The plan modifier and ONTAP normalization ensure semantic correctness
	if data.KeepAliveInterval.IsNull() || data.KeepAliveInterval.IsUnknown() {
		if restInfo.KeepAliveInterval != "" {
			data.KeepAliveInterval = types.StringValue(restInfo.KeepAliveInterval)
		} else {
			data.KeepAliveInterval = types.StringValue("PT2M") // ONTAP default
		}
	}
	if data.RequestCancelTimeout.IsNull() || data.RequestCancelTimeout.IsUnknown() {
		if restInfo.RequestCancelTimeout != "" {
			data.RequestCancelTimeout = types.StringValue(restInfo.RequestCancelTimeout)
		} else {
			data.RequestCancelTimeout = types.StringValue("PT1M") // ONTAP default
		}
	}
	if data.SessionTimeout.IsNull() || data.SessionTimeout.IsUnknown() {
		if restInfo.SessionTimeout != "" {
			data.SessionTimeout = types.StringValue(restInfo.SessionTimeout)
		} else {
			data.SessionTimeout = types.StringValue("PT10S") // ONTAP default
		}
	}
	if data.RequestAbortTimeout.IsNull() || data.RequestAbortTimeout.IsUnknown() {
		if restInfo.RequestAbortTimeout != "" {
			data.RequestAbortTimeout = types.StringValue(restInfo.RequestAbortTimeout)
		} else {
			data.RequestAbortTimeout = types.StringValue("PT40S") // ONTAP default
		}
	}
	if data.StatusRequestInterval.IsNull() || data.StatusRequestInterval.IsUnknown() {
		if restInfo.StatusRequestInterval != "" {
			data.StatusRequestInterval = types.StringValue(restInfo.StatusRequestInterval)
		} else {
			data.StatusRequestInterval = types.StringValue("PT10S") // ONTAP default
		}
	}
	data.SSLOption = types.StringValue(restInfo.SSLOption)
	data.Port = types.Int64Value(restInfo.Port)
	if data.ServerProgressTimeout.IsNull() || data.ServerProgressTimeout.IsUnknown() {
		if restInfo.ServerProgressTimeout != "" {
			data.ServerProgressTimeout = types.StringValue(restInfo.ServerProgressTimeout)
		} else {
			data.ServerProgressTimeout = types.StringValue("PT1M") // ONTAP default
		}
	}
	data.Format = types.StringValue(restInfo.Format)
	data.Type = types.StringValue(restInfo.Type)
	data.MaxServerRequests = types.Int64Value(restInfo.MaxServerRequests)
	data.MaxConnectionRetries = types.Int64Value(restInfo.MaxConnectionRetries)
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", svm.UUID, restInfo.Name))

	// Set certificate - if config provided certificate block, always create object
	if !data.Certificate.IsNull() {
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
		// No certificate block in config, set as null
			return
		}
		data.Certificate = certificateObjectValue
	} else {
		// No meaningful certificate values, set as null
		data.Certificate = types.ObjectNull(map[string]attr.Type{
			"serial_number": types.StringType,
			"name":          types.StringType,
			"ca":            types.StringType,
		})
	}

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
		"retention_duration": types.StringValue(func() string {
			if plannedRetentionDuration != "" {
				return plannedRetentionDuration
			}
			return restInfo.Resiliency.RetentionDuration
		}()),
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
		var primaryElements []attr.Value
		for _, server := range restInfo.PrimaryServers {
			primaryElements = append(primaryElements, types.StringValue(server))
		}
		primarySet, diagsPrimary := types.SetValue(types.StringType, primaryElements)
		if diagsPrimary.HasError() {
			resp.Diagnostics.Append(diagsPrimary...)
			return
		}
		data.PrimaryServers = primarySet
	}

	// Set secondary servers
	if restInfo.SecondaryServers != nil {
		var secondaryElements []attr.Value
		for _, server := range restInfo.SecondaryServers {
			secondaryElements = append(secondaryElements, types.StringValue(server))
		}
		secondarySet, diagsSecondary := types.SetValue(types.StringType, secondaryElements)
		if diagsSecondary.HasError() {
			resp.Diagnostics.Append(diagsSecondary...)
			return
		}
		data.SecondaryServers = secondarySet
	} else if !data.SecondaryServers.IsNull() {
		// Only set to empty set if secondary_servers was explicitly provided in config
		emptySet, diagsSecondary := types.SetValue(types.StringType, []attr.Value{})
		if diagsSecondary.HasError() {
			resp.Diagnostics.Append(diagsSecondary...)
			return
		}
		data.SecondaryServers = emptySet
	}
	// If data.SecondaryServers.IsNull(), leave it as null (not provided in config)

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
		body.KeepAliveInterval = normalizeDuration(data.KeepAliveInterval.ValueString())
	}
	if !data.RequestCancelTimeout.IsNull() {
		body.RequestCancelTimeout = normalizeDuration(data.RequestCancelTimeout.ValueString())
	}
	if !data.SessionTimeout.IsNull() {
		body.SessionTimeout = normalizeDuration(data.SessionTimeout.ValueString())
	}
	if !data.RequestAbortTimeout.IsNull() {
		body.RequestAbortTimeout = normalizeDuration(data.RequestAbortTimeout.ValueString())
	}
	if !data.StatusRequestInterval.IsNull() {
		body.StatusRequestInterval = normalizeDuration(data.StatusRequestInterval.ValueString())
	}
	if !data.SSLOption.IsNull() {
		body.SSLOption = data.SSLOption.ValueString()
	}
	if !data.Port.IsNull() {
		body.Port = data.Port.ValueInt64()
	}
	if !data.ServerProgressTimeout.IsNull() {
		body.ServerProgressTimeout = normalizeDuration(data.ServerProgressTimeout.ValueString())
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
			SendBuffer: bufferSize.SendBuffer.ValueInt64(),
			RecvBuffer: bufferSize.RecvBuffer.ValueInt64(),
		}
	}

	// Set resiliency - handle explicit configuration including enabled=false
	var plannedRetentionDuration string
	if !data.Resiliency.IsNull() && !data.Resiliency.IsUnknown() {
		var resiliency ProtocolsFpolicyExternalEngineResiliencyResourceModel
		diags := data.Resiliency.As(ctx, &resiliency, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		// Validate that if resiliency block is provided, enabled must be true
		if resiliency.Enabled.IsNull() || !resiliency.Enabled.ValueBool() {
			resp.Diagnostics.AddError(
				"Invalid resiliency configuration",
				"ONTAP does not support resiliency with enabled=false. When resiliency block is provided, 'enabled' must be true. To disable resiliency, remove the entire resiliency block.",
			)
			return
		}

		// Store planned retention duration to preserve it after ONTAP normalization
		plannedRetentionDuration = resiliency.RetentionDuration.ValueString()

		// Only send resiliency configuration if enabled=true (ONTAP requirement)
		if resiliency.Enabled.ValueBool() {
			body.Resiliency = interfaces.ProtocolsFpolicyExternalEngineResiliency{
				DirectoryPath:     resiliency.DirectoryPath.ValueString(),
				RetentionDuration: resiliency.RetentionDuration.ValueString(),
				Enabled:           resiliency.Enabled.ValueBool(),
			}
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

	// Update all fields with current values from ONTAP, but preserve planned duration values
	data.Name = types.StringValue(restInfo.Name)
	// For duration fields, keep the planned config values to prevent provider inconsistency
	if data.KeepAliveInterval.IsNull() || data.KeepAliveInterval.IsUnknown() {
		if restInfo.KeepAliveInterval != "" {
			data.KeepAliveInterval = types.StringValue(restInfo.KeepAliveInterval)
		} else {
			data.KeepAliveInterval = types.StringValue("PT2M") // ONTAP default
		}
	}
	if data.RequestCancelTimeout.IsNull() || data.RequestCancelTimeout.IsUnknown() {
		if restInfo.RequestCancelTimeout != "" {
			data.RequestCancelTimeout = types.StringValue(restInfo.RequestCancelTimeout)
		} else {
			data.RequestCancelTimeout = types.StringValue("PT1M") // ONTAP default
		}
	}
	if data.SessionTimeout.IsNull() || data.SessionTimeout.IsUnknown() {
		if restInfo.SessionTimeout != "" {
			data.SessionTimeout = types.StringValue(restInfo.SessionTimeout)
		} else {
			data.SessionTimeout = types.StringValue("PT10S") // ONTAP default
		}
	}
	if data.RequestAbortTimeout.IsNull() || data.RequestAbortTimeout.IsUnknown() {
		if restInfo.RequestAbortTimeout != "" {
			data.RequestAbortTimeout = types.StringValue(restInfo.RequestAbortTimeout)
		} else {
			data.RequestAbortTimeout = types.StringValue("PT40S") // ONTAP default
		}
	}
	if data.StatusRequestInterval.IsNull() || data.StatusRequestInterval.IsUnknown() {
		if restInfo.StatusRequestInterval != "" {
			data.StatusRequestInterval = types.StringValue(restInfo.StatusRequestInterval)
		} else {
			data.StatusRequestInterval = types.StringValue("PT10S") // ONTAP default
		}
	}
	data.SSLOption = types.StringValue(restInfo.SSLOption)
	data.Port = types.Int64Value(restInfo.Port)
	if data.ServerProgressTimeout.IsNull() || data.ServerProgressTimeout.IsUnknown() {
		if restInfo.ServerProgressTimeout != "" {
			data.ServerProgressTimeout = types.StringValue(restInfo.ServerProgressTimeout)
		} else {
			data.ServerProgressTimeout = types.StringValue("PT1M") // ONTAP default
		}
	}
	data.Format = types.StringValue(restInfo.Format)
	data.Type = types.StringValue(restInfo.Type)
	data.MaxServerRequests = types.Int64Value(restInfo.MaxServerRequests)
	data.MaxConnectionRetries = types.Int64Value(restInfo.MaxConnectionRetries)
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", svm.UUID, restInfo.Name))

	// Set certificate - if config provided certificate block, always create object
	if !data.Certificate.IsNull() {
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
	} else {
		// No certificate block in config, set as null
		data.Certificate = types.ObjectNull(map[string]attr.Type{
			"serial_number": types.StringType,
			"name":          types.StringType,
			"ca":            types.StringType,
		})
	}

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
		"retention_duration": types.StringValue(func() string {
			if plannedRetentionDuration != "" {
				return plannedRetentionDuration
			}
			return restInfo.Resiliency.RetentionDuration
		}()),
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
		var primaryElements []attr.Value
		for _, server := range restInfo.PrimaryServers {
			primaryElements = append(primaryElements, types.StringValue(server))
		}
		primarySet, diagsPrimary := types.SetValue(types.StringType, primaryElements)
		if diagsPrimary.HasError() {
			resp.Diagnostics.Append(diagsPrimary...)
			return
		}
		data.PrimaryServers = primarySet
	}

	// Set secondary servers
	if restInfo.SecondaryServers != nil {
		var secondaryElements []attr.Value
		for _, server := range restInfo.SecondaryServers {
			secondaryElements = append(secondaryElements, types.StringValue(server))
		}
		secondarySet, diagsSecondary := types.SetValue(types.StringType, secondaryElements)
		if diagsSecondary.HasError() {
			resp.Diagnostics.Append(diagsSecondary...)
			return
		}
		data.SecondaryServers = secondarySet
	} else if !data.SecondaryServers.IsNull() {
		// Only set to empty set if secondary_servers was explicitly provided in config
		emptySet, diagsSecondary := types.SetValue(types.StringType, []attr.Value{})
		if diagsSecondary.HasError() {
			resp.Diagnostics.Append(diagsSecondary...)
			return
		}
		data.SecondaryServers = emptySet
	}
	// If data.SecondaryServers.IsNull(), leave it as null (not provided in config)

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
