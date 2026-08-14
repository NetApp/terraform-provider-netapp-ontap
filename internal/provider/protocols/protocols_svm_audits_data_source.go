package protocols

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsSVMAuditsDataSource{}

// NewProtocolsSVMAuditsDataSource is a helper function to simplify the provider implementation.
func NewProtocolsSVMAuditsDataSource() datasource.DataSource {
	return &ProtocolsSVMAuditsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "svm_audits",
		},
	}
}

// ProtocolsSVMAuditsDataSource defines the data source implementation.
type ProtocolsSVMAuditsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsSVMAuditsDataSourceModel describes the data source data model.
type ProtocolsSVMAuditsDataSourceModel struct {
	CxProfileName      types.String                            `tfsdk:"cx_profile_name"`
	ProtocolsSVMAudits []ProtocolsSVMAuditDataSourceModel      `tfsdk:"protocols_svm_audits"`
	Filter             *ProtocolsSVMAuditDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *ProtocolsSVMAuditsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsSVMAuditsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsSVMAudits data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "The name of the SVM",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_svm_audits": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "The name of the SVM.",
							Required:            true,
						},
						"enabled": schema.BoolAttribute{
							MarkdownDescription: "Indicates whether or not auditing is enabled on the SVM.",
							Computed:            true,
						},
						"events": schema.SingleNestedAttribute{
							MarkdownDescription: "Indicates events for which auditing is enabled on the SVM.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"authorization_policy": schema.BoolAttribute{
									MarkdownDescription: "Authorization policy change events.",
									Computed:            true,
								},
								"cap_staging": schema.BoolAttribute{
									MarkdownDescription: "Central access policy staging events.",
									Computed:            true,
								},
								"cifs_logon_logoff": schema.BoolAttribute{
									MarkdownDescription: "CIFS logon and logoff events.",
									Computed:            true,
								},
								"file_operations": schema.BoolAttribute{
									MarkdownDescription: "File operation events.",
									Computed:            true,
								},
								"file_share": schema.BoolAttribute{
									MarkdownDescription: "File share category events.",
									Computed:            true,
								},
								"security_group": schema.BoolAttribute{
									MarkdownDescription: "Local security group management events.",
									Computed:            true,
								},
								"user_account": schema.BoolAttribute{
									MarkdownDescription: "Local user account management events.",
									Computed:            true,
								},
							},
						},
						"guarantee": schema.BoolAttribute{
							MarkdownDescription: `
							Indicates whether there is a strict Guarantee of Auditing.
							Requires ONTAP 9.10.1 or later.
							`,
							Computed:            true,
						},
						"log": schema.SingleNestedAttribute{
							MarkdownDescription: "Indicates the configuration options for audit log files.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"format": schema.StringAttribute{
									MarkdownDescription: `
									Describes the format in which the logs are generated by consolidation process.
									Possible values are,
									xml - Data ONTAP-specific XML log format
									evtx - Microsoft Windows EVTX log format
									`,
									Computed:            true,
								},
								"retention": schema.SingleNestedAttribute{
									MarkdownDescription: "Describes the count and time to retain the audit log file.",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"count": schema.Int64Attribute{
											MarkdownDescription: `
											Determines how many audit log files to retain before rotating the oldest log file out.
											This is mutually exclusive with duration.
											`,
											Computed:            true,
										},
										"duration": schema.StringAttribute{
											MarkdownDescription: `
											Specifies an ISO-8601 format date and time to retain the audit log file.
											The audit log files are deleted once they reach the specified date/time.
											`,
											Computed:            true,
										},
									},
								},
								"rotation": schema.SingleNestedAttribute{
									MarkdownDescription: `
									Audit event log files are rotated when they reach a configured threshold log size or are on a configured schedule.
									When an event log file is rotated, the scheduled consolidation task first renames the active converted file to a time-stamped archive file,
									and then creates a new active converted event log file.
									`,
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"size": schema.Int64Attribute{
											MarkdownDescription: `
											Rotates logs based on log size in bytes.
											Default value is 104857600.
											`,
											Computed:            true,
										},
										"schedule": schema.SingleNestedAttribute{
											MarkdownDescription: `
											Rotates the audit logs based on a schedule by using the time-based rotation parameters in any combination.
											The rotation schedule is calculated by using all the time-related values.
											`,
											Computed:            true,
											Attributes: map[string]schema.Attribute{
												"days": schema.ListAttribute{
													MarkdownDescription: `
													Specifies the day of the month schedule to rotate audit log.
													Specify -1 to rotate the audit logs all days of a month.
													`,
													Computed:            true,
													ElementType:         types.Int64Type,
												},
												"hours": schema.ListAttribute{
													MarkdownDescription: `
													Specifies the hourly schedule to rotate audit log.
													Specify -1 to rotate the audit logs every hour.
													`,
													Computed:            true,
													ElementType:         types.Int64Type,
												},
												"minutes": schema.ListAttribute{
													MarkdownDescription: "Specifies the minutes schedule to rotate the audit log.",
													Computed:            true,
													ElementType:         types.Int64Type,
												},
												"months": schema.ListAttribute{
													MarkdownDescription: `
													Specifies the months schedule to rotate audit log.
													Specify -1 to rotate the audit logs every month.
													`,
													Computed:            true,
													ElementType:         types.Int64Type,
												},
												"weekdays": schema.ListAttribute{
													MarkdownDescription: `
													Specifies the weekdays schedule to rotate audit log.
													Specify -1 to rotate the audit logs every day.
													`,
													Computed:            true,
													ElementType:         types.Int64Type,
												},
											},
										},
									},
								},
							},
						},
						"log_path": schema.StringAttribute{
							MarkdownDescription: "The audit log destination path where consolidated audit logs are stored.",
							Computed:            true,
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsSVMAuditsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsSVMAuditsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsSVMAuditsDataSourceModel

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

	var filter *interfaces.ProtocolsAuditConfigsFilterModel
	if data.Filter != nil {
		filter = &interfaces.ProtocolsAuditConfigsFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetAuditConfigs(errorHandler, *client, filter, cluster.Version)
	if err != nil {
		// error reporting done inside GetAuditConfigs
		return
	}

	data.ProtocolsSVMAudits = make([]ProtocolsSVMAuditDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		// events
		var events *EventsDataSourceModel
		if record.Events != nil {
			events = &EventsDataSourceModel{
				AuthorizationPolicy: types.BoolValue(record.Events.AuthorizationPolicy),
				CapStaging:          types.BoolValue(record.Events.CapStaging),
				CIFSLogonLogoff:     types.BoolValue(record.Events.CIFSLogonLogoff),
				FileOperations:      types.BoolValue(record.Events.FileOperations),
				FileShare:           types.BoolValue(record.Events.FileShare),
				SecurityGroup:	     types.BoolValue(record.Events.SecurityGroup),
				UserAccount:         types.BoolValue(record.Events.UserAccount),
			}
		} else {
			events = nil
		}
		// log
		var log *LogDataSourceModel
		if record.Log != nil {
			days := make([]types.Int64, 0)
			hours := make([]types.Int64, 0)
			minutes := make([]types.Int64, 0)
			months := make([]types.Int64, 0)
			weekdays := make([]types.Int64, 0)

			if record.Log.Rotation != nil && record.Log.Rotation.Schedule != nil {
				for _, value := range record.Log.Rotation.Schedule.Days {
					days = append(days, types.Int64Value(value))
				}
				for _, value := range record.Log.Rotation.Schedule.Hours {
					hours = append(hours, types.Int64Value(value))
				}
				for _, value := range record.Log.Rotation.Schedule.Minutes {
					minutes = append(minutes, types.Int64Value(value))
				}
				for _, value := range record.Log.Rotation.Schedule.Months {
					months = append(months, types.Int64Value(value))
				}
				for _, value := range record.Log.Rotation.Schedule.Weekdays {
					weekdays = append(weekdays, types.Int64Value(value))
				}
			}

			log = &LogDataSourceModel{
				Format:    types.StringValue(record.Log.Format),
				Retention: RetentionDataSourceModel{},
				Rotation:  RotationDataSourceModel{},
			}
			if record.Log.Retention != nil {
				log.Retention = RetentionDataSourceModel{
					Count:    types.Int64Value(int64(record.Log.Retention.Count)),
					Duration: types.StringValue(record.Log.Retention.Duration),
				}
			}
			if record.Log.Rotation != nil {
				log.Rotation = RotationDataSourceModel{
					Size: types.Int64Value(int64(record.Log.Rotation.Size)),
					Schedule: ScheduleDataSourceModel{
						Days:     days,
						Hours:    hours,
						Minutes:  minutes,
						Months:   months,
						Weekdays: weekdays,
					},
				}
			}
		} else {
			log = nil
		}
		data.ProtocolsSVMAudits[index] = ProtocolsSVMAuditDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			SVMName:       types.StringValue(record.SVM.Name),
			Enabled:       types.BoolValue(record.Enabled),
			Events:        events,
			Guarantee:     types.BoolValue(record.Guarantee),
			Log: 		   log,
			LogPath:	   types.StringValue(record.LogPath),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
