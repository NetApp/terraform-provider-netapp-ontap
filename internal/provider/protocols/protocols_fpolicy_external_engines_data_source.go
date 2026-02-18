package protocols

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsFpolicyExternalEnginesDataSource{}

// NewProtocolsFpolicyExternalEnginesDataSource is a helper function to simplify the provider implementation.
func NewProtocolsFpolicyExternalEnginesDataSource() datasource.DataSource {
	return &ProtocolsFpolicyExternalEnginesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "fpolicy_external_engines",
		},
	}
}

// ProtocolsFpolicyExternalEnginesDataSource defines the data source implementation.
type ProtocolsFpolicyExternalEnginesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsFpolicyExternalEnginesDataSourceModel describes the data source data model.
type ProtocolsFpolicyExternalEnginesDataSourceModel struct {
	CxProfileName                   types.String                                          `tfsdk:"cx_profile_name"`
	ProtocolsFpolicyExternalEngines []ProtocolsFpolicyExternalEngineDataSourceModel      `tfsdk:"protocols_fpolicy_external_engines"`
	Filter                          *ProtocolsFpolicyExternalEngineDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *ProtocolsFpolicyExternalEnginesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsFpolicyExternalEnginesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsFpolicyExternalEngines data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "The name of the SVM",
						Required:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "FPolicy external engine name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_fpolicy_external_engines": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "FPolicy external engine name",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "SVM name",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "FPolicy external engine UUID",
							Computed:            true,
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
				},
				Computed:            true,
				MarkdownDescription: "List of FPolicy external engines",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsFpolicyExternalEnginesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsFpolicyExternalEnginesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsFpolicyExternalEnginesDataSourceModel

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

	if data.Filter == nil || data.Filter.SVMName.IsNull() {
		errorHandler.MakeAndReportError("No SVM specified", "svm_name must be specified in filter")
		return
	}

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.Filter.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	var filter *interfaces.ProtocolsFpolicyExternalEngineDataSourceFilterModel = nil
	filter = &interfaces.ProtocolsFpolicyExternalEngineDataSourceFilterModel{
		Name:    data.Filter.Name.ValueString(),
		SVMUUID: svm.UUID,
	}
	restInfo, err := interfaces.GetProtocolsFpolicyExternalEngines(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetProtocolsFpolicyExternalEngines
		return
	}

	data.ProtocolsFpolicyExternalEngines = make([]ProtocolsFpolicyExternalEngineDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		// For the list data source, we'll only include basic information
		data.ProtocolsFpolicyExternalEngines[index] = ProtocolsFpolicyExternalEngineDataSourceModel{
			CxProfileName:         data.CxProfileName,
			Name:                  types.StringValue(record.Name),
			SVMName:               types.StringValue(data.Filter.SVMName.ValueString()),
			ID:                    types.StringValue(fmt.Sprintf("%s/%s", record.SVM.UUID, record.Name)),
			KeepAliveInterval:     types.StringValue(record.KeepAliveInterval),
			RequestCancelTimeout:  types.StringValue(record.RequestCancelTimeout),
			SessionTimeout:        types.StringValue(record.SessionTimeout),
			RequestAbortTimeout:   types.StringValue(record.RequestAbortTimeout),
			StatusRequestInterval: types.StringValue(record.StatusRequestInterval),
			SSLOption:             types.StringValue(record.SSLOption),
			Port:                  types.Int64Value(record.Port),
			ServerProgressTimeout: types.StringValue(record.ServerProgressTimeout),
			Format:                types.StringValue(record.Format),
			Type:                  types.StringValue(record.Type),
			MaxServerRequests:     types.Int64Value(record.MaxServerRequests),
			MaxConnectionRetries:  types.Int64Value(record.MaxConnectionRetries),
		}
		
		// Set certificate
		if record.Certificate.Name != "" || record.Certificate.SerialNumber != "" || record.Certificate.Ca != "" {
			certificateAttrs := map[string]attr.Value{
				"serial_number": types.StringValue(record.Certificate.SerialNumber),
				"name":          types.StringValue(record.Certificate.Name),
				"ca":            types.StringValue(record.Certificate.Ca),
			}
			certificateAttrTypes := map[string]attr.Type{
				"serial_number": types.StringType,
				"name":          types.StringType,
				"ca":            types.StringType,
			}
			data.ProtocolsFpolicyExternalEngines[index].Certificate, _ = types.ObjectValue(certificateAttrTypes, certificateAttrs)
		} else {
			data.ProtocolsFpolicyExternalEngines[index].Certificate = types.ObjectNull(map[string]attr.Type{
				"serial_number": types.StringType,
				"name":          types.StringType,
				"ca":            types.StringType,
			})
		}

		// Set buffer size
		if record.BufferSize.SendBuffer != 0 || record.BufferSize.RecvBuffer != 0 {
			bufferSizeAttrs := map[string]attr.Value{
				"send_buffer": types.Int64Value(int64(record.BufferSize.SendBuffer)),
				"recv_buffer": types.Int64Value(int64(record.BufferSize.RecvBuffer)),
			}
			bufferSizeAttrTypes := map[string]attr.Type{
				"send_buffer": types.Int64Type,
				"recv_buffer": types.Int64Type,
			}
			data.ProtocolsFpolicyExternalEngines[index].BufferSize, _ = types.ObjectValue(bufferSizeAttrTypes, bufferSizeAttrs)
		} else {
			data.ProtocolsFpolicyExternalEngines[index].BufferSize = types.ObjectNull(map[string]attr.Type{
				"send_buffer": types.Int64Type,
				"recv_buffer": types.Int64Type,
			})
		}

		// Set resiliency
		if record.Resiliency.DirectoryPath != "" || record.Resiliency.RetentionDuration != "" || record.Resiliency.Enabled {
			resiliencyAttrs := map[string]attr.Value{
				"directory_path":     types.StringValue(record.Resiliency.DirectoryPath),
				"retention_duration": types.StringValue(record.Resiliency.RetentionDuration),
				"enabled":            types.BoolValue(record.Resiliency.Enabled),
			}
			resiliencyAttrTypes := map[string]attr.Type{
				"directory_path":     types.StringType,
				"retention_duration": types.StringType,
				"enabled":            types.BoolType,
			}
			data.ProtocolsFpolicyExternalEngines[index].Resiliency, _ = types.ObjectValue(resiliencyAttrTypes, resiliencyAttrs)
		} else {
			data.ProtocolsFpolicyExternalEngines[index].Resiliency = types.ObjectNull(map[string]attr.Type{
				"directory_path":     types.StringType,
				"retention_duration": types.StringType,
				"enabled":            types.BoolType,
			})
		}

		// Set primary servers
		if len(record.PrimaryServers) > 0 {
			primaryServersSet, diags := types.SetValueFrom(ctx, types.StringType, record.PrimaryServers)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.ProtocolsFpolicyExternalEngines[index].PrimaryServers = primaryServersSet
		} else {
			data.ProtocolsFpolicyExternalEngines[index].PrimaryServers, _ = types.SetValue(types.StringType, []attr.Value{})
		}

		// Set secondary servers
		if len(record.SecondaryServers) > 0 {
			secondaryServersSet, diags := types.SetValueFrom(ctx, types.StringType, record.SecondaryServers)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			data.ProtocolsFpolicyExternalEngines[index].SecondaryServers = secondaryServersSet
		} else {
			data.ProtocolsFpolicyExternalEngines[index].SecondaryServers, _ = types.SetValue(types.StringType, []attr.Value{})
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
