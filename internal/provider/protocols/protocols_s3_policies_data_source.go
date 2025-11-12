package protocols

import (
    "context"
    "fmt"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-log/tflog"
    "github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
    "github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
    "github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined interface
var _ datasource.DataSource = &ProtocolsS3PoliciesDataSource{}

// ProtocolsS3PoliciesDataSource is the data source implementation.
type ProtocolsS3PoliciesDataSource struct {
    config connection.ResourceOrDataSourceConfig
}

// NewProtocolsS3PoliciesDataSource is a helper function to simplify the provider implementation.
func NewProtocolsS3PoliciesDataSource() datasource.DataSource {
    return &ProtocolsS3PoliciesDataSource{
        config: connection.ResourceOrDataSourceConfig{
            Name: "s3_policies",
        },
    }
}

// ProtocolsS3PoliciesDataSourceModel describes the data source data model.
type ProtocolsS3PoliciesDataSourceModel struct {
    CxProfileName types.String                       `tfsdk:"cx_profile_name"`
    SVMName       types.String                       `tfsdk:"svm_name"`
    Filter        *ProtocolsS3PolicyDataSourceFilterModel    `tfsdk:"filter"`
    Policies      []ProtocolsS3PolicyDataModel       `tfsdk:"protocols_s3_policies"`
    ID            types.String                       `tfsdk:"id"`
}

// ProtocolsS3PolicyDataModel describes individual policy data model.
type ProtocolsS3PolicyDataModel struct {
    SVMName    types.String                          `tfsdk:"svm_name"`
    Name       types.String                          `tfsdk:"name"`
    Comment    types.String                          `tfsdk:"comment"`
    ReadOnly   types.Bool                            `tfsdk:"read_only"`
    Statements []ProtocolsS3PolicyStatementModel    `tfsdk:"statements"`
}

// Metadata returns the data source type name.
func (d *ProtocolsS3PoliciesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsS3PoliciesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "ProtocolsS3Policies data source",

        Attributes: map[string]schema.Attribute{
            "cx_profile_name": schema.StringAttribute{
                MarkdownDescription: "Connection profile name",
                Required:            true,
            },
            "svm_name": schema.StringAttribute{
                MarkdownDescription: "Name of the SVM",
                Computed:            true,
            },
            "id": schema.StringAttribute{
                MarkdownDescription: "Identifier",
                Computed:            true,
            },
            "filter": schema.SingleNestedAttribute{
                MarkdownDescription: "Filter for S3 policies",
                Optional:            true,
                Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "The name of the SVM",
						Optional:            true,
					},
					"name": schema.StringAttribute{
						MarkdownDescription: "The name of the S3 policy",
						Optional:            true,
					},
                },
            },
            "protocols_s3_policies": schema.ListNestedAttribute{
                Computed:            true,
                MarkdownDescription: "List of S3 policies",
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "The name of the SVM",
							Required:            true,
						},
                        "name": schema.StringAttribute{
                            MarkdownDescription: "Name of the S3 policy",
                            Required:            true,
                        },
                        "comment": schema.StringAttribute{
                            MarkdownDescription: "Optional comment for the S3 policy",
                            Computed:            true,
                        },
                        "read_only": schema.BoolAttribute{
                            MarkdownDescription: "Indicates if the policy is read-only",
                            Computed:            true,
                        },
                        "statements": schema.ListNestedAttribute{
                            Computed:            true,
                            MarkdownDescription: "Policy statements",
                            NestedObject: schema.NestedAttributeObject{
                                Attributes: map[string]schema.Attribute{
                                    "sid": schema.StringAttribute{
                                        MarkdownDescription: "Statement ID",
                                        Computed:            true,
                                    },
                                    "effect": schema.StringAttribute{
                                        MarkdownDescription: "Statement effect (Allow or Deny)",
                                        Computed:            true,
                                    },
                                    "actions": schema.ListAttribute{
                                        ElementType:         types.StringType,
                                        MarkdownDescription: "List of S3 actions",
                                        Computed:            true,
                                    },
                                    "resources": schema.ListAttribute{
                                        ElementType:         types.StringType,
                                        MarkdownDescription: "List of S3 resources",
                                        Computed:            true,
                                    },

                                },
                            },
                        },
                    },
                },
            },
        },
    }
}

// Read refreshes the Terraform state with the latest data.
func (d *ProtocolsS3PoliciesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data ProtocolsS3PoliciesDataSourceModel

    // Read Terraform configuration data into the model
    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
    
    // Get REST client
    client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
    if err != nil {
        // error reporting done inside GetRestClient
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

    restInfos, err := interfaces.GetProtocolsS3Policies(errorHandler, *client, svm.UUID, cluster.Version, data.Filter.Name.ValueString())
    if err != nil {
        // error reporting done inside GetProtocolsS3Policies
        return
    }

    // Convert policies
    policies := make([]ProtocolsS3PolicyDataModel, len(restInfos))
    for i, restInfo := range restInfos {
        // Convert statements
        statements := make([]ProtocolsS3PolicyStatementModel, len(restInfo.Statements))
        for j, stmt := range restInfo.Statements {
            // Convert actions
            actions := make([]types.String, len(stmt.Actions))
            for k, action := range stmt.Actions {
                actions[k] = types.StringValue(action)
            }

            // Convert resources
            resources := make([]types.String, len(stmt.Resources))
            for k, resource := range stmt.Resources {
                resources[k] = types.StringValue(resource)
            }

            statements[j] = ProtocolsS3PolicyStatementModel{
                Effect:      types.StringValue(stmt.Effect),
                Actions:     actions,
                Resources:   resources,
                SID:        types.StringValue(interfaces.ConvertSIDToString(stmt.SID)),
            }
        }

        // Handle empty values as null for consistency
        comment := types.StringValue(restInfo.Comment)
        if restInfo.Comment == "" {
            comment = types.StringNull()
        }

        policies[i] = ProtocolsS3PolicyDataModel{
            SVMName:  types.StringValue(data.Filter.SVMName.ValueString()),
            Name:     types.StringValue(restInfo.Name),
            Comment:  comment,
            ReadOnly: types.BoolValue(restInfo.ReadOnly),
            Statements: statements,
        }
    }

    data.Policies = policies
    
    // Generate ID using the SVM name
    svmNameForID := data.Filter.SVMName.ValueString()
    data.ID = types.StringValue(fmt.Sprintf("svm_%s", svmNameForID))

    // Log data
    tflog.Debug(ctx, fmt.Sprintf("read S3 policies data source: %#v", data))

    // Save data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsS3PoliciesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    config, ok := req.ProviderData.(connection.Config)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Data Source Configure Type",
            fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
        )
        return
    }

    d.config.ProviderConfig = config
}
