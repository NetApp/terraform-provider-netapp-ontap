package protocols

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsFpolicyExternalEngineDataSource{}

// NewProtocolsFpolicyExternalEngineDataSource is a helper function to simplify the provider implementation.
func NewProtocolsFpolicyExternalEngineDataSource() datasource.DataSource {
	return &ProtocolsFpolicyExternalEngineDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_fpolicy_external_engine",
		},
	}
}

// ProtocolsFpolicyExternalEngineDataSource defines the data source implementation.
type ProtocolsFpolicyExternalEngineDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsFpolicyExternalEngineDataSourceModel describes the data source data model.
type ProtocolsFpolicyExternalEngineDataSourceModel struct {
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
	PrimaryServers        types.Set    `tfsdk:"primary_servers"`
	BufferSize            types.Object `tfsdk:"buffer_size"`
	SecondaryServers      types.Set    `tfsdk:"secondary_servers"`
	Port                  types.Int64  `tfsdk:"port"`
	ServerProgressTimeout types.String `tfsdk:"server_progress_timeout"`
	Format                types.String `tfsdk:"format"`
	Type                  types.String `tfsdk:"type"`
	Resiliency            types.Object `tfsdk:"resiliency"`
	MaxServerRequests     types.Int64  `tfsdk:"max_server_requests"`
	MaxConnectionRetries  types.Int64  `tfsdk:"max_connection_retries"`
}

// ProtocolsFpolicyExternalEngineCertificateDataSourceModel describes the certificate data model
type ProtocolsFpolicyExternalEngineCertificateDataSourceModel struct {
	SerialNumber types.String `tfsdk:"serial_number"`
	Name         types.String `tfsdk:"name"`
	Ca           types.String `tfsdk:"ca"`
}

// ProtocolsFpolicyExternalEngineBufferSizeDataSourceModel describes the buffer size data model
type ProtocolsFpolicyExternalEngineBufferSizeDataSourceModel struct {
	SendBuffer types.Int64 `tfsdk:"send_buffer"`
	RecvBuffer types.Int64 `tfsdk:"recv_buffer"`
}

// ProtocolsFpolicyExternalEngineResiliencyDataSourceModel describes the resiliency data model
type ProtocolsFpolicyExternalEngineResiliencyDataSourceModel struct {
	DirectoryPath     types.String `tfsdk:"directory_path"`
	RetentionDuration types.String `tfsdk:"retention_duration"`
	Enabled           types.Bool   `tfsdk:"enabled"`
}

// ProtocolsFpolicyExternalEngineDataSourceFilterModel describes the data source data model for queries.
type ProtocolsFpolicyExternalEngineDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
	Name    types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsFpolicyExternalEngineDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsFpolicyExternalEngineDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsFpolicyExternalEngine data source",

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
				Computed:            true,
			},
			"request_cancel_timeout": schema.StringAttribute{
				MarkdownDescription: "Specifies the ISO-8601 timeout duration for a screen request to be processed by an FPolicy server",
				Computed:            true,
			},
			"certificate": schema.SingleNestedAttribute{
				MarkdownDescription: "Provides details about certificate used to authenticate the FPolicy server",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"serial_number": schema.StringAttribute{
						MarkdownDescription: "Serial number",
						Computed:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "Certificate name",
						Computed:            true,
					},
					"ca": schema.StringAttribute{
						MarkdownDescription: "Certificate authority",
						Computed:            true,
					},
				},
			},
			"session_timeout": schema.StringAttribute{
				MarkdownDescription: "Specifies the interval after which a new session ID is sent to the FPolicy server during reconnection attempts",
				Computed:            true,
			},
			"request_abort_timeout": schema.StringAttribute{
				MarkdownDescription: "Specifies the ISO-8601 timeout duration for a screen request to be aborted by a storage appliance",
				Computed:            true,
			},
			"status_request_interval": schema.StringAttribute{
				MarkdownDescription: "Specifies the ISO-8601 interval time for a storage appliance to query a status request from an FPolicy server",
				Computed:            true,
			},
			"ssl_option": schema.StringAttribute{
				MarkdownDescription: "The SSL option for external communication with the FPolicy server",
				Computed:            true,
			},
			"primary_servers": schema.SetAttribute{
				MarkdownDescription: "The primary FPolicy servers to which the node sends",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"buffer_size": schema.SingleNestedAttribute{
				MarkdownDescription: "Specifies the send and receive buffer size of the connected socket for the FPolicy server",
				Optional:            true,
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"send_buffer": schema.Int64Attribute{
						MarkdownDescription: "Send buffer size",
						Computed:            true,
					},
					"recv_buffer": schema.Int64Attribute{
						MarkdownDescription: "Receive buffer size",
						Computed:            true,
					},
				},
			},
			"secondary_servers": schema.SetAttribute{
				MarkdownDescription: "Send file access events for a given FPolicy policy",
				Computed:            true,
				ElementType:         types.StringType,
			},
			"port": schema.Int64Attribute{
				MarkdownDescription: "Port number of the FPolicy server application",
				Computed:            true,
			},
			"server_progress_timeout": schema.StringAttribute{
				MarkdownDescription: "Specifies the ISO-8601 timeout duration in which a throttled FPolicy server must complete at least one screen request",
				Computed:            true,
			},
			"format": schema.StringAttribute{
				MarkdownDescription: "The format for the notification messages sent to the FPolicy servers",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Determines what ONTAP does after sending notifications to FPolicy servers",
				Computed:            true,
			},
			"resiliency": schema.SingleNestedAttribute{
				MarkdownDescription: "When primary and secondary servers are down or unresponsive, file access events are stored in the storage controller under the specified resiliency-directory-path",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"directory_path": schema.StringAttribute{
						MarkdownDescription: "Directory path",
						Computed:            true,
					},
					"retention_duration": schema.StringAttribute{
						MarkdownDescription: "Retention duration",
						Computed:            true,
					},
					"enabled": schema.BoolAttribute{
						MarkdownDescription: "Enabled status",
						Computed:            true,
					},
				},
			},
			"max_server_requests": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of outstanding requests for the FPolicy server",
				Computed:            true,
			},
			"max_connection_retries": schema.Int64Attribute{
				MarkdownDescription: "The maximum number of attempts to reconnect to the FPolicy server",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsFpolicyExternalEngineDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsFpolicyExternalEngineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsFpolicyExternalEngineDataSourceModel

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

	// Set certificate
	if restInfo.Certificate.Name != "" || restInfo.Certificate.SerialNumber != "" || restInfo.Certificate.Ca != "" {
		certificateAttrs := map[string]attr.Value{
			"serial_number": types.StringValue(restInfo.Certificate.SerialNumber),
			"name":          types.StringValue(restInfo.Certificate.Name),
			"ca":            types.StringValue(restInfo.Certificate.Ca),
		}
		certificateAttrTypes := map[string]attr.Type{
			"serial_number": types.StringType,
			"name":          types.StringType,
			"ca":            types.StringType,
		}
		data.Certificate, _ = types.ObjectValue(certificateAttrTypes, certificateAttrs)
	} else {
		data.Certificate = types.ObjectNull(map[string]attr.Type{
			"serial_number": types.StringType,
			"name":          types.StringType,
			"ca":            types.StringType,
		})
	}

	// Set buffer size
	if restInfo.BufferSize.SendBuffer != 0 || restInfo.BufferSize.RecvBuffer != 0 {
		bufferSizeAttrs := map[string]attr.Value{
			"send_buffer": types.Int64Value(int64(restInfo.BufferSize.SendBuffer)),
			"recv_buffer": types.Int64Value(int64(restInfo.BufferSize.RecvBuffer)),
		}
		bufferSizeAttrTypes := map[string]attr.Type{
			"send_buffer": types.Int64Type,
			"recv_buffer": types.Int64Type,
		}
		data.BufferSize, _ = types.ObjectValue(bufferSizeAttrTypes, bufferSizeAttrs)
	} else {
		data.BufferSize = types.ObjectNull(map[string]attr.Type{
			"send_buffer": types.Int64Type,
			"recv_buffer": types.Int64Type,
		})
	}

	// Set resiliency
	if restInfo.Resiliency.DirectoryPath != "" || restInfo.Resiliency.RetentionDuration != "" || restInfo.Resiliency.Enabled {
		resiliencyAttrs := map[string]attr.Value{
			"directory_path":     types.StringValue(restInfo.Resiliency.DirectoryPath),
			"retention_duration": types.StringValue(restInfo.Resiliency.RetentionDuration),
			"enabled":            types.BoolValue(restInfo.Resiliency.Enabled),
		}
		resiliencyAttrTypes := map[string]attr.Type{
			"directory_path":     types.StringType,
			"retention_duration": types.StringType,
			"enabled":            types.BoolType,
		}
		data.Resiliency, _ = types.ObjectValue(resiliencyAttrTypes, resiliencyAttrs)
	} else {
		data.Resiliency = types.ObjectNull(map[string]attr.Type{
			"directory_path":     types.StringType,
			"retention_duration": types.StringType,
			"enabled":            types.BoolType,
		})
	}

	// Set primary servers
	if len(restInfo.PrimaryServers) > 0 {
		primaryServersList := make([]types.String, len(restInfo.PrimaryServers))
		for i, server := range restInfo.PrimaryServers {
			primaryServersList[i] = types.StringValue(server)
		}
		var diagsTemp diag.Diagnostics
		data.PrimaryServers, diagsTemp = types.SetValueFrom(ctx, types.StringType, primaryServersList)
		if diagsTemp.HasError() {
			resp.Diagnostics.Append(diagsTemp...)
			return
		}
	}

	// Set secondary servers
	if len(restInfo.SecondaryServers) > 0 {
		secondaryServersList := make([]types.String, len(restInfo.SecondaryServers))
		for i, server := range restInfo.SecondaryServers {
			secondaryServersList[i] = types.StringValue(server)
		}
		data.SecondaryServers, _ = types.SetValue(types.StringType, basetypes.ListValue{}.Elements())
		secondaryServersSet, diags := types.SetValueFrom(ctx, types.StringType, secondaryServersList)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.SecondaryServers = secondaryServersSet
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
