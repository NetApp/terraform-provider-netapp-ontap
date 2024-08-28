package storage

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/svm"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageQuotaRulesDataSource{}

// NewStorageQuotaRulesDataSource is a helper function to simplify the provider implementation.
func NewStorageQuotaRulesDataSource() datasource.DataSource {
	return &StorageQuotaRulesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "quota_rules",
		},
	}
}

// StorageQuotaRulesDataSource defines the data source implementation.
type StorageQuotaRulesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageQuotaRulesDataSourceModel describes the data source data model.
type StorageQuotaRulesDataSourceModel struct {
	CxProfileName     types.String                            `tfsdk:"cx_profile_name"`
	StorageQuotaRules []StorageQuotaRuleDataSourceModel       `tfsdk:"storage_quota_rules"`
	Filter            *StorageQuotaRulesDataSourceFilterModel `tfsdk:"filter"`
}

// StorageQuotaRulesDataSourceFilterModel describes the data source data model for queries.
type StorageQuotaRulesDataSourceFilterModel struct {
	Type   types.String `tfsdk:"type"`
	SVM    types.String `tfsdk:"svm_name"`
	Volume types.String `tfsdk:"volume_name"`
	Qtree  types.String `tfsdk:"qtree_name"`
}

// Metadata returns the data source type name.
func (d *StorageQuotaRulesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *StorageQuotaRulesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageQuotaRules data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "StorageQuotaRule type",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "StorageQuotaRule svm name",
						Optional:            true,
					},
					"volume_name": schema.StringAttribute{
						MarkdownDescription: "StorageQuotaRule volume name",
						Optional:            true,
					},
					"qtree_name": schema.StringAttribute{
						MarkdownDescription: "StorageQuotaRule qtree name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"storage_quota_rules": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"type": schema.StringAttribute{
							MarkdownDescription: "Quota type for the rule. This type can be user, group, or tree",
							Optional:            true,
							Computed:            true,
						},
						"svm": schema.SingleNestedAttribute{
							MarkdownDescription: "Existing SVM",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "name of the SVM",
									Required:            true,
								},
							},
						},
						"volume": schema.SingleNestedAttribute{
							MarkdownDescription: "Existing volume",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "name of the volume",
									Required:            true,
								},
							},
						},
						"qtree": schema.SingleNestedAttribute{
							MarkdownDescription: "Qtree name for the rule",
							Optional:            true,
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "name of the qtree",
									Optional:            true,
									Computed:            true,
								},
							},
						},
						"users": schema.SetNestedAttribute{
							MarkdownDescription: "user to which the user quota policy rule applies",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										MarkdownDescription: "name of the user",
										Computed:            true,
									},
								},
							},
						},
						"group": schema.SingleNestedAttribute{
							MarkdownDescription: "group to which the group quota policy rule applies",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "name of the group",
									Computed:            true,
								},
							},
						},
						"files": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"hard_limit": schema.Int64Attribute{
									MarkdownDescription: "Specifies the hard limit for files",
									Computed:            true,
								},
								"soft_limit": schema.Int64Attribute{
									MarkdownDescription: "Specifies the soft limit for files",
									Computed:            true,
								},
							},
						},
						"user_mapping": schema.BoolAttribute{
							MarkdownDescription: "user mapping for user quota policy rules",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							Computed: true,
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
func (d *StorageQuotaRulesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *StorageQuotaRulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageQuotaRulesDataSourceModel

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

	var filter *interfaces.StorageQuotaRulesDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.StorageQuotaRulesDataSourceFilterModel{
			Type:       data.Filter.Type.ValueString(),
			SVMName:    data.Filter.SVM.ValueString(),
			VolumeName: data.Filter.Volume.ValueString(),
			QtreeName:  data.Filter.Qtree.ValueString(),
		}
	}
	restInfo, err := interfaces.GetOneORMoreStorageQuotaRules(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetStorageQuotaRules
		return
	}

	data.StorageQuotaRules = make([]StorageQuotaRuleDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.StorageQuotaRules[index] = StorageQuotaRuleDataSourceModel{}
		data.StorageQuotaRules[index].CxProfileName = types.String(data.CxProfileName)
		data.StorageQuotaRules[index].Type = types.StringValue(record.Type)
		data.StorageQuotaRules[index].SVM = svm.SVM{
			Name: types.StringValue(record.SVM.Name),
		}
		data.StorageQuotaRules[index].Volume = Volume{
			Name: types.StringValue(record.Volume.Name),
		}
		data.StorageQuotaRules[index].Qtree = &Qtree{
			Name: types.StringValue(record.Qtree.Name),
		}
		//Users
		data.StorageQuotaRules[index].Users = &[]User{}
		for _, user := range record.Users {
			*data.StorageQuotaRules[index].Users = append(*data.StorageQuotaRules[index].Users, User{Name: types.StringValue(user.Name)})
		}
		//Group
		elementTypes := map[string]attr.Type{
			"name": types.StringType,
		}
		elements := map[string]attr.Value{
			"name": types.StringValue(record.Group.Name),
		}
		objectValue, diags := types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
		data.StorageQuotaRules[index].Group = objectValue
		//Files
		elementTypes = map[string]attr.Type{
			"hard_limit": types.Int64Type,
			"soft_limit": types.Int64Type,
		}
		elements = map[string]attr.Value{
			"hard_limit": types.Int64Value(record.Files.HardLimit),
			"soft_limit": types.Int64Value(record.Files.SoftLimit),
		}
		objectValue, diags = types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
		}
		data.StorageQuotaRules[index].Files = objectValue
		data.StorageQuotaRules[index].UserMapping = types.BoolValue(record.UserMapping)
		data.StorageQuotaRules[index].ID = types.StringValue(record.UUID)

	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
