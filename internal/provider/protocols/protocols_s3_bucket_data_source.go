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
var _ datasource.DataSource = &ProtocolsS3BucketDataSource{}

// NewProtocolsS3BucketDataSource is a helper function to simplify the provider implementation.
func NewProtocolsS3BucketDataSource() datasource.DataSource {
	return &ProtocolsS3BucketDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "s3_bucket",
		},
	}
}

// ProtocolsS3BucketDataSource defines the data source implementation.
type ProtocolsS3BucketDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsS3BucketDataSourceModel describes the data source data model.
type ProtocolsS3BucketDataSourceModel struct {
	CxProfileName        types.String                        `tfsdk:"cx_profile_name"`
	SVMName              types.String                        `tfsdk:"svm_name"`
	Name                 types.String                        `tfsdk:"name"`
	Size		         types.Int64                         `tfsdk:"size"`
	Comment              types.String                        `tfsdk:"comment"`
	Type 		         types.String                        `tfsdk:"type"`	
	NASPath	             types.String                        `tfsdk:"nas_path"`
	VersioningState      types.String                        `tfsdk:"versioning_state"`
	Policy              *PolicyDataSourceModel              `tfsdk:"policy"`
	QoSPolicy           *QoSPolicyDataSourceModel           `tfsdk:"qos_policy"`
	SnapshotPolicy       types.String                        `tfsdk:"snapshot_policy"`
	AuditEventSelector  *AuditEventSelectorDataSourceModel  `tfsdk:"audit_event_selector"`
	UUID                 types.String                        `tfsdk:"uuid"`
}

// PolicyDataSourceModel describes the policy data model using go types for mapping.
type PolicyDataSourceModel struct {
	Statements  []StatementsDataSourceModel  `tfsdk:"statements"`
}

// StatementsDataSourceModel describes the statements data model using go types for mapping.
type StatementsDataSourceModel struct {
	SID           types.String               `tfsdk:"sid"`
	Resources     types.Set                  `tfsdk:"resources"`
	Actions       types.Set                  `tfsdk:"actions"`
	Effect        types.String               `tfsdk:"effect"`
	Principals    types.Set                  `tfsdk:"principals"`
	Conditions  []ConditionsDataSourceModel  `tfsdk:"conditions"`
}

// ConditionsDataSourceModel describes the conditions data model using go types for mapping.
type ConditionsDataSourceModel struct {
	Operator    types.String  `tfsdk:"operator"`
	MaxKeys     types.Set     `tfsdk:"max_keys"`
	Delimiters  types.Set     `tfsdk:"delimiters"`
	SourceIPs   types.Set     `tfsdk:"source_ips"`
	Prefixes    types.Set     `tfsdk:"prefixes"`
	Usernames   types.Set     `tfsdk:"usernames"`
}

// QoSPolicyDataSourceModel describes the qos_policy data model using go types for mapping.
type QoSPolicyDataSourceModel struct {
	Name               types.String  `tfsdk:"name"`
	MaxThroughputIops  types.Int64   `tfsdk:"max_throughput_iops"`
	MaxThroughputMbps  types.Int64   `tfsdk:"max_throughput_mbps"`
	MinThroughputIops  types.Int64   `tfsdk:"min_throughput_iops"`
	MinThroughputMbps  types.Int64   `tfsdk:"min_throughput_mbps"`
}

// AuditEventSelectorDataSourceModel describes the audit_event_selector data model using go types for mapping.
type AuditEventSelectorDataSourceModel struct {
	Access      types.String  `tfsdk:"access"`
	Permission  types.String  `tfsdk:"permission"`
}

// ProtocolsS3BucketDataSourceFilterModel describes the data source data model for queries.
type ProtocolsS3BucketDataSourceFilterModel struct {
	SVMName  types.String  `tfsdk:"svm_name"`
	Name     types.String  `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *ProtocolsS3BucketDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsS3BucketDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsS3Bucket data source",

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
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsS3BucketDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsS3BucketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsS3BucketDataSourceModel

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

	restInfo, err := interfaces.GetProtocolsS3Bucket(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsS3Bucket
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No S3 bucket found", "S3 bucket not found")
		return
	}
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.Size = types.Int64Value(restInfo.Size)
	data.Comment = types.StringValue(restInfo.Comment)
	data.Type = types.StringValue(restInfo.Type)
	data.NASPath = types.StringValue(restInfo.NASPath)
	data.VersioningState = types.StringValue(restInfo.VersioningState)
	// policy
	if restInfo.Policy != nil && len(restInfo.Policy.Statements) > 0 {
		var policy PolicyDataSourceModel
		policy.Statements = []StatementsDataSourceModel{}
		
		for _, item := range restInfo.Policy.Statements {
			var statement StatementsDataSourceModel
			statement.SID = types.StringValue(item.SID)
			statement.Effect = types.StringValue(item.Effect)
			// resources
			var resources = make([]string, len(item.Resources))
			copy(resources, item.Resources)
			resourcesSet, diags := types.SetValueFrom(ctx, types.StringType, resources)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			statement.Resources = resourcesSet
			//actions
			var actions = make([]string, len(item.Actions))
			copy(actions, item.Actions)
			actionsSet, diags := types.SetValueFrom(ctx, types.StringType, actions)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			statement.Actions = actionsSet
			// principals
			var principals = make([]string, len(item.Principals))
			copy(principals, item.Principals)
			principalsSet, diags := types.SetValueFrom(ctx, types.StringType, principals)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			statement.Principals = principalsSet
			//conditions
			var conditions []ConditionsDataSourceModel
			for _, cond := range item.Conditions {
				var condition ConditionsDataSourceModel
				condition.Operator = types.StringValue(cond.Operator)
				// max_keys
				var maxKeys = make([]string, len(cond.MaxKeys))
				copy(maxKeys, cond.MaxKeys)
				maxKeysSet, diags := types.SetValueFrom(ctx, types.StringType, maxKeys)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				condition.MaxKeys = maxKeysSet
				// delimiters
				var delimiters = make([]string, len(cond.Delimiters))
				copy(delimiters, cond.Delimiters)
				delimitersSet, diags := types.SetValueFrom(ctx, types.StringType, delimiters)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				condition.Delimiters = delimitersSet
				// source_ips
				var sourceIPs = make([]string, len(cond.SourceIPs))
				copy(sourceIPs, cond.SourceIPs)
				sourceIPsSet, diags := types.SetValueFrom(ctx, types.StringType, sourceIPs)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				condition.SourceIPs = sourceIPsSet
				// prefixes
				var prefixes = make([]string, len(cond.Prefixes))
				copy(prefixes, cond.Prefixes)
				prefixesSet, diags := types.SetValueFrom(ctx, types.StringType, prefixes)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				condition.Prefixes = prefixesSet
				// usernames
				var usernames = make([]string, len(cond.Usernames))
				copy(usernames, cond.Usernames)
				usernamesSet, diags := types.SetValueFrom(ctx, types.StringType, usernames)
				resp.Diagnostics.Append(diags...)
				if resp.Diagnostics.HasError() {
					return
				}
				condition.Usernames = usernamesSet
				conditions = append(conditions, condition)
			}
			statement.Conditions = conditions

			policy.Statements = append(policy.Statements, statement)
		}
		data.Policy = &policy
	} else {
		data.Policy = nil
	}
	// qos_policy
	if restInfo.QoSPolicy != nil {
		qos := restInfo.QoSPolicy
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
			data.QoSPolicy = &QoSPolicyDataSourceModel{
				Name:              name,
				MaxThroughputIops: maxIops,
				MaxThroughputMbps: maxMbps,
				MinThroughputIops: minIops,
				MinThroughputMbps: minMbps,
			}
		} else {
			data.QoSPolicy = nil
		}
	} else {
		data.QoSPolicy = nil
	}
	
	if restInfo.SnapshotPolicy.Name != "" {
		data.SnapshotPolicy = types.StringValue(restInfo.SnapshotPolicy.Name)
	} else {
		data.SnapshotPolicy = types.StringNull()
	}
	// audit_event_selector
	if restInfo.AuditEventSelector != nil && (restInfo.AuditEventSelector.Access != "" || restInfo.AuditEventSelector.Permission != "") {
		data.AuditEventSelector = &AuditEventSelectorDataSourceModel{
			Access:      types.StringValue(restInfo.AuditEventSelector.Access),
			Permission:  types.StringValue(restInfo.AuditEventSelector.Permission),
		}
	} else {
		data.AuditEventSelector = nil
	}
	data.UUID = types.StringValue(restInfo.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
