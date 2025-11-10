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

// Ensure provider defined interface
var _ datasource.DataSource = &ProtocolsS3PolicyDataSource{}

// ProtocolsS3PolicyDataSource is the data source implementation.
type ProtocolsS3PolicyDataSource struct {
    config connection.ResourceOrDataSourceConfig
}

// NewProtocolsS3PolicyDataSource is a helper function to simplify the provider implementation.
func NewProtocolsS3PolicyDataSource() datasource.DataSource {
    return &ProtocolsS3PolicyDataSource{
        config: connection.ResourceOrDataSourceConfig{
            Name: "s3_policy",
        },
    }
}

// ProtocolsS3PolicyDataSourceModel describes the data source data model.
type ProtocolsS3PolicyDataSourceModel struct {
    CxProfileName types.String                    `tfsdk:"cx_profile_name"`
    ID           types.String                    `tfsdk:"id"`
    Name         types.String                    `tfsdk:"name"`
    Comment      types.String                    `tfsdk:"comment"`
    ReadOnly     types.Bool                      `tfsdk:"read_only"`
    SVMName      types.String                    `tfsdk:"svm_name"`
    Filter       *ProtocolsS3PolicyFilterModel   `tfsdk:"filter"`
    SVM          *ProtocolsS3PolicySVMModel      `tfsdk:"svm"`
    Statements   []ProtocolsS3PolicyStatementModel `tfsdk:"statements"`
}

// ProtocolsS3PolicyFilterModel describes the filter data model.
type ProtocolsS3PolicyFilterModel struct {
    Name    types.String `tfsdk:"name"`
    SVMName types.String `tfsdk:"svm_name"`
}

// ProtocolsS3PolicySVMModel describes the SVM data model.
type ProtocolsS3PolicySVMModel struct {
    Name types.String `tfsdk:"name"`
}

// ProtocolsS3PolicyStatementModel describes the statement data model.
type ProtocolsS3PolicyStatementModel struct {
    SID        types.String   `tfsdk:"sid"`
    Effect     types.String   `tfsdk:"effect"`
    Actions    []types.String `tfsdk:"actions"`
    Resources  []types.String `tfsdk:"resources"`
}

// Metadata returns the data source type name.
func (d *ProtocolsS3PolicyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsS3PolicyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "ProtocolsS3Policy data source",

        Attributes: map[string]schema.Attribute{
            "cx_profile_name": schema.StringAttribute{
                MarkdownDescription: "Connection profile name",
                Required:            true,
            },
            "name": schema.StringAttribute{
                MarkdownDescription: "Name of the S3 policy",
                Optional:            true,
            },
            "svm_name": schema.StringAttribute{
                MarkdownDescription: "Name of the SVM",
                Optional:            true,
            },
            "filter": schema.SingleNestedAttribute{
                MarkdownDescription: "Filter for S3 policy",
                Optional:            true,
                Attributes: map[string]schema.Attribute{
                    "name": schema.StringAttribute{
                        MarkdownDescription: "Filter by policy name",
                        Optional:            true,
                    },
                    "svm_name": schema.StringAttribute{
                        MarkdownDescription: "Filter by SVM name", 
                        Optional:            true,
                    },
                },
            },
            "id": schema.StringAttribute{
                MarkdownDescription: "Identifier",
                Computed:            true,
            },
            "comment": schema.StringAttribute{
                MarkdownDescription: "Optional comment for the S3 policy",
                Computed:            true,
            },
            "read_only": schema.BoolAttribute{
                MarkdownDescription: "Indicates if the policy is read-only",
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
    }
}

// Read refreshes the Terraform state with the latest data.
func (d *ProtocolsS3PolicyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data ProtocolsS3PolicyDataSourceModel

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

    // Prepare filter from user input - prefer filter over individual fields
    filter := &interfaces.ProtocolsS3PolicyDataSourceFilterModel{}
    
    // Use filter values if provided, otherwise fall back to top-level fields
    if data.Filter != nil && !data.Filter.SVMName.IsNull() {
        filter.SVMName = data.Filter.SVMName.ValueString()
    } else if !data.SVMName.IsNull() {
        filter.SVMName = data.SVMName.ValueString()
    } else {
        resp.Diagnostics.AddError("Missing required parameter", "Either svm_name or filter.svm_name must be provided")
        return
    }
    
    if data.Filter != nil && !data.Filter.Name.IsNull() {
        filter.Name = data.Filter.Name.ValueString()
    } else if !data.Name.IsNull() {
        filter.Name = data.Name.ValueString()
    } else {
        resp.Diagnostics.AddError("Missing required parameter", "Either name or filter.name must be provided")
        return
    }

    // Make API call to get S3 policy using filter
    restInfo, err := interfaces.GetProtocolsS3PolicyByFilter(errorHandler, *client, filter)
    if err != nil {
        return
    }

    // Map response to model
    data.Name = types.StringValue(restInfo.Name)
    data.Comment = types.StringValue(restInfo.Comment)
    data.ReadOnly = types.BoolValue(restInfo.ReadOnly)
    data.SVMName = types.StringValue(restInfo.SVM.Name)
    
    // Set SVM details
    data.SVM = &ProtocolsS3PolicySVMModel{
        Name: types.StringValue(restInfo.SVM.Name),
    }

    // Convert statements
    statements := make([]ProtocolsS3PolicyStatementModel, len(restInfo.Statements))
    for i, stmt := range restInfo.Statements {
        // Convert actions
        actions := make([]types.String, len(stmt.Actions))
        for j, action := range stmt.Actions {
            actions[j] = types.StringValue(action)
        }

        // Convert resources
        resources := make([]types.String, len(stmt.Resources))
        for j, resource := range stmt.Resources {
            resources[j] = types.StringValue(resource)
        }

        statements[i] = ProtocolsS3PolicyStatementModel{
            Effect:      types.StringValue(stmt.Effect),
            Actions:     actions,
            Resources:   resources,
            SID:        types.StringValue(interfaces.ConvertSIDToString(stmt.SID)),
        }
    }

    data.Statements = statements
    data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.SVMName.ValueString(), data.Name.ValueString()))

    // Log data
    tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

    // Save data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsS3PolicyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
