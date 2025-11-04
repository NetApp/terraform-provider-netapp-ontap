package object_store

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
    Filter        *ProtocolsS3PoliciesFilterModel    `tfsdk:"filter"`
    Policies      []ProtocolsS3PolicyDataModel       `tfsdk:"protocols_s3_policies"`
    ID            types.String                       `tfsdk:"id"`
}

// ProtocolsS3PoliciesFilterModel describes the filter data model.
type ProtocolsS3PoliciesFilterModel struct {
    Name types.String `tfsdk:"name"`
}

// ProtocolsS3PolicyDataModel describes individual policy data model.
type ProtocolsS3PolicyDataModel struct {
    Name       types.String                          `tfsdk:"name"`
    Comment    types.String                          `tfsdk:"comment"`
    SVM        *ProtocolsS3PolicySVMModel            `tfsdk:"svm"`
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
                Required:            true,
            },
            "id": schema.StringAttribute{
                MarkdownDescription: "Identifier",
                Computed:            true,
            },
            "filter": schema.SingleNestedAttribute{
                MarkdownDescription: "Filter for S3 policies",
                Optional:            true,
                Attributes: map[string]schema.Attribute{
                    "name": schema.StringAttribute{
                        MarkdownDescription: "Filter by policy name",
                        Optional:            true,
                    },
                },
            },
            "protocols_s3_policies": schema.ListNestedAttribute{
                Computed:            true,
                MarkdownDescription: "List of S3 policies",
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "name": schema.StringAttribute{
                            MarkdownDescription: "Name of the S3 policy",
                            Computed:            true,
                        },
                        "comment": schema.StringAttribute{
                            MarkdownDescription: "Optional comment for the S3 policy",
                            Computed:            true,
                        },
                        "svm": schema.SingleNestedAttribute{
                            Computed:            true,
                            MarkdownDescription: "SVM details",
                            Attributes: map[string]schema.Attribute{
                                "name": schema.StringAttribute{
                                    MarkdownDescription: "SVM name",
                                    Computed:            true,
                                },
                                "uuid": schema.StringAttribute{
                                    MarkdownDescription: "SVM UUID",
                                    Computed:            true,
                                },
                            },
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
        return
    }

    // Prepare filter
    var filter interfaces.ProtocolsS3PolicyDataSourceFilterModel
    filter.SVMName = data.SVMName
    if data.Filter != nil && !data.Filter.Name.IsNull() {
        filter.Name = data.Filter.Name
    }

    // Make API call to get S3 policies
    restInfos, err := interfaces.GetProtocolsS3Policies(errorHandler, *client, &filter)
    if err != nil {
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
                SID:        types.StringValue(stmt.SID),
                Effect:     types.StringValue(stmt.Effect),
                Actions:    actions,
                Resources:  resources,
            }
        }

        // Handle empty values as null for consistency
        comment := types.StringValue(restInfo.Comment)
        if restInfo.Comment == "" {
            comment = types.StringNull()
        }
        
        svmUUID := types.StringValue(restInfo.SVM.UUID)
        if restInfo.SVM.UUID == "" {
            svmUUID = types.StringNull()
        }

        policies[i] = ProtocolsS3PolicyDataModel{
            Name:    types.StringValue(restInfo.Name),
            Comment: comment,
            SVM: &ProtocolsS3PolicySVMModel{
                Name: types.StringValue(restInfo.SVM.Name),
                UUID: svmUUID,
            },
            Statements: statements,
        }
    }

    data.Policies = policies
    data.ID = types.StringValue(fmt.Sprintf("svm_%s", data.SVMName.ValueString()))

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