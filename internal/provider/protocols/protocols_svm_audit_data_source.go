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
var _ datasource.DataSource = &ProtocolsSVMAuditDataSource{}

// NewProtocolsSVMAuditDataSource is a helper function to simplify the provider implementation.
func NewProtocolsSVMAuditDataSource() datasource.DataSource {
	return &ProtocolsSVMAuditDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "svm_audit",
		},
	}
}

// ProtocolsSVMAuditDataSource defines the data source implementation.
type ProtocolsSVMAuditDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsSVMAuditDataSourceModel describes the data source data model.
type ProtocolsSVMAuditDataSourceModel struct {
	CxProfileName types.String           `tfsdk:"cx_profile_name"`
	SVMName       types.String           `tfsdk:"svm_name"`
	Enabled       types.Bool             `tfsdk:"enabled"`
	Events		  *EventsDataSourceModel `tfsdk:"events"`
	Guarantee	  types.Bool             `tfsdk:"guarantee"`
	Log           *LogDataSourceModel    `tfsdk:"log"`
	LogPath 	  types.String           `tfsdk:"log_path"`
}

// EventsDataSourceModel describes the events data model using go types for mapping.
type EventsDataSourceModel struct {
	AuthorizationPolicy types.Bool `tfsdk:"authorization_policy"`
	CapStaging 		    types.Bool `tfsdk:"cap_staging"`
	CIFSLogonLogoff     types.Bool `tfsdk:"cifs_logon_logoff"`
	FileOperations      types.Bool `tfsdk:"file_operations"`
	FileShare           types.Bool `tfsdk:"file_share"`
	SecurityGroup		types.Bool `tfsdk:"security_group"`
	UserAccount         types.Bool `tfsdk:"user_account"`
}

// LogDataSourceModel describes the log data model using go types for mapping.
type LogDataSourceModel struct {
	Format    types.String             `tfsdk:"format"`
	Retention RetentionDataSourceModel `tfsdk:"retention"`
	Rotation  RotationDataSourceModel  `tfsdk:"rotation"`
}

// RetentionDataSourceModel describes the retention data model using go types for mapping.
type RetentionDataSourceModel struct {
	Count    types.Int64  `tfsdk:"count"`
	Duration types.String `tfsdk:"duration"`
}

// RotationDataSourceModel describes the rotation data model using go types for mapping.
type RotationDataSourceModel struct {
	Size     types.Int64             `tfsdk:"size"`
	Schedule ScheduleDataSourceModel `tfsdk:"schedule"`
}

// ScheduleDataSourceModel describes the schedule data model using go types for mapping.
type ScheduleDataSourceModel struct {
	Days     []types.Int64 `tfsdk:"days"`
	Hours    []types.Int64 `tfsdk:"hours"`
	Minutes  []types.Int64 `tfsdk:"minutes"`
	Months   []types.Int64 `tfsdk:"months"`
	Weekdays []types.Int64 `tfsdk:"weekdays"`
}

// ProtocolsSVMAuditDataSourceFilterModel describes the data source data model for queries.
type ProtocolsSVMAuditDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsSVMAuditDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsSVMAuditDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsSVMAudit data source",

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
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsSVMAuditDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsSVMAuditDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsSVMAuditDataSourceModel

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

	restInfo, err := interfaces.GetSVMAuditConfig(errorHandler, *client, data.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetSVMAuditConfig
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No SVM audit found", "SVM audit not found")
		return
	}
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Enabled = types.BoolValue(restInfo.Enabled)
	// events
	if restInfo.Events != nil {
		data.Events = &EventsDataSourceModel{
			AuthorizationPolicy: types.BoolValue(restInfo.Events.AuthorizationPolicy),
			CapStaging:          types.BoolValue(restInfo.Events.CapStaging),
			CIFSLogonLogoff:     types.BoolValue(restInfo.Events.CIFSLogonLogoff),
			FileOperations:      types.BoolValue(restInfo.Events.FileOperations),
			FileShare:           types.BoolValue(restInfo.Events.FileShare),
			SecurityGroup:	     types.BoolValue(restInfo.Events.SecurityGroup),
			UserAccount:         types.BoolValue(restInfo.Events.UserAccount),
		}
	} else {
		data.Events = nil
	}
	data.Guarantee = types.BoolValue(restInfo.Guarantee)
	// log
	if restInfo.Log != nil {
		days := make([]types.Int64, 0)
		hours := make([]types.Int64, 0)
		minutes := make([]types.Int64, 0)
		months := make([]types.Int64, 0)
		weekdays := make([]types.Int64, 0)

		if restInfo.Log.Rotation != nil && restInfo.Log.Rotation.Schedule != nil {
			for _, value := range restInfo.Log.Rotation.Schedule.Days {
				days = append(days, types.Int64Value(value))
			}
			for _, value := range restInfo.Log.Rotation.Schedule.Hours {
				hours = append(hours, types.Int64Value(value))
			}
			for _, value := range restInfo.Log.Rotation.Schedule.Minutes {
				minutes = append(minutes, types.Int64Value(value))
			}
			for _, value := range restInfo.Log.Rotation.Schedule.Months {
				months = append(months, types.Int64Value(value))
			}
			for _, value := range restInfo.Log.Rotation.Schedule.Weekdays {
				weekdays = append(weekdays, types.Int64Value(value))
			}
		}

		data.Log = &LogDataSourceModel{
			Format:    types.StringValue(restInfo.Log.Format),
			Retention: RetentionDataSourceModel{},
			Rotation:  RotationDataSourceModel{},
		}
		if restInfo.Log.Retention != nil {
			data.Log.Retention = RetentionDataSourceModel{
				Count:    types.Int64Value(int64(restInfo.Log.Retention.Count)),
				Duration: types.StringValue(restInfo.Log.Retention.Duration),
			}
		}
		if restInfo.Log.Rotation != nil {
			data.Log.Rotation = RotationDataSourceModel{
				Size: types.Int64Value(int64(restInfo.Log.Rotation.Size)),
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
		data.Log = nil
	}
	data.LogPath = types.StringValue(restInfo.LogPath)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
