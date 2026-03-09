package protocols

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &ProtocolsS3BucketsDataSource{}

// NewProtocolsS3BucketsDataSource is a helper function to simplify the provider implementation.
func NewProtocolsS3BucketsDataSource() datasource.DataSource {
	return &ProtocolsS3BucketsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "s3_buckets",
		},
	}
}

// ProtocolsS3BucketsDataSource defines the data source implementation.
type ProtocolsS3BucketsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsS3BucketsDataSourceModel describes the data source data model.
type ProtocolsS3BucketsDataSourceModel struct {
	CxProfileName         types.String                            `tfsdk:"cx_profile_name"`
	ProtocolsS3Buckets  []ProtocolsS3BucketDataSourceModel        `tfsdk:"protocols_s3_buckets"`
	Filter               *ProtocolsS3BucketDataSourceFilterModel  `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *ProtocolsS3BucketsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsS3BucketsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsS3Buckets data source",

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
					"name": schema.StringAttribute{
						MarkdownDescription: "The name of the S3 bucket",
						Optional:            true,
					},
				},
				Required: true,
			},
			"protocols_s3_buckets": schema.ListNestedAttribute{
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
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the S3 or NAS bucket.",
							Required:            true,
						},
						"size": schema.Int64Attribute{
							MarkdownDescription: "The size of the S3 bucket in bytes.",
							Computed:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Additional information about the bucket.",
							Computed:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "The type of the bucket.",
							Computed:            true,
						},
						"nas_path": schema.StringAttribute{
							MarkdownDescription: "The NAS path to which the NAS bucket corresponds.",
							Computed:            true,
						},
						"versioning_state": schema.StringAttribute{
							MarkdownDescription: "The versioning state of the bucket.",
							Computed:            true,
						},
						"policy": schema.SingleNestedAttribute{
							MarkdownDescription: "Access policy configuration for the bucket.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"statements": schema.ListNestedAttribute{
									MarkdownDescription: "The list of policy statements.",
									Computed:            true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"sid": schema.StringAttribute{
												MarkdownDescription: "Statement ID",
												Computed:            true,
											},
											"resources": schema.SetAttribute{
												MarkdownDescription: "The bucket and any object it contains.",
												Computed:            true,
												ElementType:         types.StringType,
											},
											"actions": schema.SetAttribute{
												MarkdownDescription: "The list of actions.",
												Computed:            true,
												ElementType:         types.StringType,
											},
											"effect": schema.StringAttribute{
												MarkdownDescription: "The effect of the statement.",
												Computed:            true,
											},
											"principals": schema.SetAttribute{
													MarkdownDescription: "The list of S3 users or groups.",
													Computed:            true,
													ElementType:         types.StringType,
											},
											"conditions": schema.ListNestedAttribute{
												MarkdownDescription: "Conditions for when a policy is in effect.",
												Computed:            true,
												NestedObject: schema.NestedAttributeObject{
													Attributes: map[string]schema.Attribute{
														"operator": schema.StringAttribute{
															MarkdownDescription: "The operator for the condition.",
															Computed:            true,
														},
														"max_keys": schema.SetAttribute{
															MarkdownDescription: "The maximum number of keys that can be returned in a request.",
															Computed:            true,
															ElementType:         types.StringType,
														},
														"delimiters": schema.SetAttribute{
															MarkdownDescription: "The delimiter used to identify a prefix in a list of objects.",
															Computed:            true,
															ElementType:         types.StringType,
														},
														"source_ips": schema.SetAttribute{
															MarkdownDescription: "The source IP addresses of the request.",
															Computed:            true,
															ElementType:         types.StringType,
														},
														"prefixes": schema.SetAttribute{
															MarkdownDescription: "The prefixes of the objects to be listed.",
															Computed:            true,
															ElementType:         types.StringType,
														},
														"usernames": schema.SetAttribute{
															MarkdownDescription: "The user names that are allowed to access the bucket.",
															Computed:            true,
															ElementType:         types.StringType,
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"qos_policy": schema.SingleNestedAttribute{
							MarkdownDescription: "The QoS policy configuration for the bucket.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "The QoS policy group name.",
									Computed:            true,
								},
								"max_throughput_iops": schema.Int64Attribute{
									MarkdownDescription: "The maximum throughput in IOPS.",
									Computed:            true,
								},
								"max_throughput_mbps": schema.Int64Attribute{
									MarkdownDescription: "The maximum throughput in MBPS.",
									Computed:            true,
								},
								"min_throughput_iops": schema.Int64Attribute{
									MarkdownDescription: "The minimum throughput in IOPS.",
									Computed:            true,
								},
								"min_throughput_mbps": schema.Int64Attribute{
									MarkdownDescription: "The minimum throughput in MBPS.",
									Computed:            true,
								},
							},
						},
						"snapshot_policy": schema.StringAttribute{
							MarkdownDescription: "The snapshot policy for the bucket.",
							Computed:            true,
						},
						"audit_event_selector": schema.SingleNestedAttribute{
							MarkdownDescription: "The audit event selector configuration for the bucket.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"access": schema.StringAttribute{
									MarkdownDescription: "The type of event access to be audited.",
									Computed:            true,
								},
								"permission": schema.StringAttribute{
									MarkdownDescription: "The type of event permission to be audited.",
									Computed:            true,
								},
							},
						},
						"uuid": schema.StringAttribute{
							MarkdownDescription: "The UUID of the S3 bucket.",
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
func (d *ProtocolsS3BucketsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsS3BucketsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsS3BucketsDataSourceModel

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

	var filter *interfaces.S3BucketsFilterModel
	if data.Filter != nil {
		filter = &interfaces.S3BucketsFilterModel{
			SVMName: data.Filter.SVMName.ValueString(),
			Name:    data.Filter.Name.ValueString(),
		}
	}
	restInfo, err := interfaces.GetProtocolsS3Buckets(errorHandler, *client, filter, cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsS3Buckets
		return
	}

	data.ProtocolsS3Buckets = make([]ProtocolsS3BucketDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		// policy
		var policy *PolicyDataSourceModel
		if record.Policy != nil && len(record.Policy.Statements) > 0 {
			policyModel := PolicyDataSourceModel{Statements: []StatementsDataSourceModel{}}
			for _, item := range record.Policy.Statements {
				var statement StatementsDataSourceModel
				statement.SID = types.StringValue(item.SID)
				statement.Effect = types.StringValue(item.Effect)

				resourcesSet, diags := types.SetValueFrom(ctx, types.StringType, item.Resources)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				statement.Resources = resourcesSet

				actionsSet, diags := types.SetValueFrom(ctx, types.StringType, item.Actions)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				statement.Actions = actionsSet

				principalsSet, diags := types.SetValueFrom(ctx, types.StringType, item.Principals)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				statement.Principals = principalsSet

				var conditions []ConditionsDataSourceModel
				for _, cond := range item.Conditions {
					var condition ConditionsDataSourceModel
					condition.Operator = types.StringValue(cond.Operator)

					maxKeysSet, diags := types.SetValueFrom(ctx, types.StringType, cond.MaxKeys)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
					condition.MaxKeys = maxKeysSet

					delimitersSet, diags := types.SetValueFrom(ctx, types.StringType, cond.Delimiters)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
					condition.Delimiters = delimitersSet

					sourceIPsSet, diags := types.SetValueFrom(ctx, types.StringType, cond.SourceIPs)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
					condition.SourceIPs = sourceIPsSet

					prefixesSet, diags := types.SetValueFrom(ctx, types.StringType, cond.Prefixes)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
					condition.Prefixes = prefixesSet

					usernamesSet, diags := types.SetValueFrom(ctx, types.StringType, cond.Usernames)
					resp.Diagnostics.Append(diags...)
					if resp.Diagnostics.HasError() {
						return
					}
					condition.Usernames = usernamesSet

					conditions = append(conditions, condition)
				}
				statement.Conditions = conditions

				policyModel.Statements = append(policyModel.Statements, statement)
			}
			policy = &policyModel
		}

		// qos_policy
		var qosPolicy *QoSPolicyDataSourceModel
		if record.QoSPolicy != nil {
			qos := record.QoSPolicy
			if qos.Name != "" || qos.MaxThroughputIops != 0 || qos.MaxThroughputMbps != 0 || qos.MinThroughputIops != 0 || qos.MinThroughputMbps != 0 {
				name := types.StringNull()
				if qos.Name != "" {
					name = types.StringValue(qos.Name)
				}
				maxIops := types.Int64Null()
				if qos.MaxThroughputIops != 0 {
					maxIops = types.Int64Value(int64(qos.MaxThroughputIops))
				}
				maxMbps := types.Int64Null()
				if qos.MaxThroughputMbps != 0 {
					maxMbps = types.Int64Value(int64(qos.MaxThroughputMbps))
				}
				minIops := types.Int64Null()
				if qos.MinThroughputIops != 0 {
					minIops = types.Int64Value(int64(qos.MinThroughputIops))
				}
				minMbps := types.Int64Null()
				if qos.MinThroughputMbps != 0 {
					minMbps = types.Int64Value(int64(qos.MinThroughputMbps))
				}
				qosPolicy = &QoSPolicyDataSourceModel{
					Name:              name,
					MaxThroughputIops: maxIops,
					MaxThroughputMbps: maxMbps,
					MinThroughputIops: minIops,
					MinThroughputMbps: minMbps,
				}
			}
		}

		// audit_event_selector
		var auditEventSelector *AuditEventSelectorDataSourceModel
		if record.AuditEventSelector != nil && (record.AuditEventSelector.Access != "" || record.AuditEventSelector.Permission != "") {
			auditEventSelector = &AuditEventSelectorDataSourceModel{
				Access:     types.StringValue(record.AuditEventSelector.Access),
				Permission: types.StringValue(record.AuditEventSelector.Permission),
			}
		}

		// snapshot_policy
		snapshotPolicy := types.StringNull()
		if record.SnapshotPolicy.Name != "" {
			snapshotPolicy = types.StringValue(record.SnapshotPolicy.Name)
		}

		data.ProtocolsS3Buckets[index] = ProtocolsS3BucketDataSourceModel{
			CxProfileName:      types.String(data.CxProfileName),
			SVMName:            types.StringValue(record.SVM.Name),
			Name:               types.StringValue(record.Name),
			Size:               types.Int64Value(record.Size),
			Comment:            types.StringValue(record.Comment),
			Type:               types.StringValue(record.Type),
			NASPath:            types.StringValue(record.NASPath),
			VersioningState:    types.StringValue(record.VersioningState),
			Policy:             policy,
			QoSPolicy:          qosPolicy,
			SnapshotPolicy:     snapshotPolicy,
			AuditEventSelector: auditEventSelector,
			UUID:               types.StringValue(record.UUID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
