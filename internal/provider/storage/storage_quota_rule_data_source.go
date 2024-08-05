package storage

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/svm"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageQuotaRuleDataSource{}

// NewStorageQuotaRuleDataSource is a helper function to simplify the provider implementation.
func NewStorageQuotaRuleDataSource() datasource.DataSource {
	return &StorageQuotaRuleDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "quota_rule",
		},
	}
}

// StorageQuotaRuleDataSource defines the data source implementation.
type StorageQuotaRuleDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageQuotaRuleDataSourceModel describes the resource data model.
type StorageQuotaRuleDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	SVM           svm.SVM      `tfsdk:"svm"`
	Volume        Volume       `tfsdk:"volume"`
	Users         *[]User      `tfsdk:"users"`
	Group         types.Object `tfsdk:"group"`
	Qtree         *Qtree       `tfsdk:"qtree"`
	Type          types.String `tfsdk:"type"`
	Files         types.Object `tfsdk:"files"`
	UserMapping   types.Bool   `tfsdk:"user_mapping"`
	ID            types.String `tfsdk:"id"`
}

// Metadata returns the data source type name.
func (d *StorageQuotaRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *StorageQuotaRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageQuotaRule data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Quota type for the rule. This type can be user, group, or tree",
				Required:            true,
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
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "name of the qtree",
						Required:            true,
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *StorageQuotaRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *StorageQuotaRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageQuotaRuleDataSourceModel

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

	restInfo, err := interfaces.GetStorageQuotaRules(errorHandler, *client, data.Volume.Name.ValueString(), data.SVM.Name.ValueString(), data.Type.ValueString(), data.Qtree.Name.ValueString())
	if err != nil {
		// error reporting done inside GetStorageQuotaRule
		return
	}

	data.Volume.Name = types.StringValue(restInfo.Volume.Name)
	data.SVM.Name = types.StringValue(restInfo.SVM.Name)
	data.Type = types.StringValue(restInfo.Type)
	data.Qtree.Name = types.StringValue(restInfo.Qtree.Name)
	//Users
	data.Users = &[]User{}
	for _, user := range restInfo.Users {
		*data.Users = append(*data.Users, User{Name: types.StringValue(user.Name)})
	}
	//Group
	elementTypes := map[string]attr.Type{
		"name": types.StringType,
	}
	elements := map[string]attr.Value{
		"name": types.StringValue(restInfo.Group.Name),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Group = objectValue
	//Files
	elementTypes = map[string]attr.Type{
		"hard_limit": types.Int64Type,
		"soft_limit": types.Int64Type,
	}
	elements = map[string]attr.Value{
		"hard_limit": types.Int64Value(restInfo.Files.HardLimit),
		"soft_limit": types.Int64Value(restInfo.Files.SoftLimit),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Files = objectValue
	data.UserMapping = types.BoolValue(restInfo.UserMapping)
	data.ID = types.StringValue(restInfo.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
