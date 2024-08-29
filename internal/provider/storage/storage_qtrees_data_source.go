package storage

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/provider/storage_qtree_data_source.go)
// replace StorageQtrees with the name of the resource, following go conventions, eg IPInterfaces
// replace storage_qtrees with the name of the resource, for logging purposes, eg ip_interfaces
// make sure to create internal/interfaces/storage_qtree.go too)
// delete these 5 lines

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageQtreesDataSource{}

// NewStorageQtreesDataSource is a helper function to simplify the provider implementation.
func NewStorageQtreesDataSource() datasource.DataSource {
	return &StorageQtreesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "storage_qtrees",
		},
	}
}

// StorageQtreesDataSource defines the data source implementation.
type StorageQtreesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageQtreesDataSourceModel describes the data source data model.
type StorageQtreesDataSourceModel struct {
	CxProfileName types.String                        `tfsdk:"cx_profile_name"`
	StorageQtrees []StorageQtreeDataSourceModel       `tfsdk:"storage_qtrees"`
	Filter        *StorageQtreesDataSourceFilterModel `tfsdk:"filter"`
}

// StorageQtreesDataSourceFilterModel describes the data source data model for queries.
type StorageQtreesDataSourceFilterModel struct {
	Name       types.String `tfsdk:"name"`
	SVMName    types.String `tfsdk:"svm_name"`
	VolumeName types.String `tfsdk:"volume_name"`
}

// Metadata returns the data source type name.
func (d *StorageQtreesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *StorageQtreesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageQtrees data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "StorageQtree name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "StorageQtree svm name",
						Optional:            true,
					},
					"volume_name": schema.StringAttribute{
						MarkdownDescription: "The volume that contains the qtree.",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"storage_qtrees": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "StorageQtree name",
							Computed:            true,
						},
						"volume_name": schema.StringAttribute{
							MarkdownDescription: "The volume that contains the qtree.",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "IPInterface svm name",
							Computed:            true,
						},
						"id": schema.Int64Attribute{
							MarkdownDescription: "The ID of the qtree.",
							Computed:            true,
						},
						"unix_permissions": schema.Int64Attribute{
							MarkdownDescription: "The UNIX permissions for the qtree.",
							Computed:            true,
						},
						"security_style": schema.StringAttribute{
							MarkdownDescription: "StorageQtree security style",
							Computed:            true,
							Validators: []validator.String{
								stringvalidator.OneOf("unix", "ntfs", "mixed"),
							},
						},
						"nas": schema.SingleNestedAttribute{
							MarkdownDescription: "NAS settings",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"path": schema.StringAttribute{
									MarkdownDescription: "Client visible path to the qtree. This field is not available if the volume does not have a junction-path configured.",
									Computed:            true,
								},
							},
						},
						"user": schema.SingleNestedAttribute{
							MarkdownDescription: "The user set as owner of the qtree.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Alphanumeric username of user who owns the qtree.",
									Computed:            true,
								},
							},
						},
						"group": schema.SingleNestedAttribute{
							MarkdownDescription: "The group set as owner of the qtree.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Alphanumeric group name of group who owns the qtree.",
									Computed:            true,
								},
							},
						},
						"export_policy": schema.SingleNestedAttribute{
							MarkdownDescription: "The export policy for the qtree.",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "The name of the export policy.",
									Computed:            true,
								},
								"id": schema.Int64Attribute{
									MarkdownDescription: "The ID of the export policy.",
									Computed:            true,
								},
							},
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
func (d *StorageQtreesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *StorageQtreesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageQtreesDataSourceModel

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

	var filter *interfaces.StorageQtreeDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.StorageQtreeDataSourceFilterModel{
			Name:    data.Filter.Name.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetStorageQtrees(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetStorageQtrees
		return
	}
	data.StorageQtrees = make([]StorageQtreeDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.StorageQtrees[index] = StorageQtreeDataSourceModel{
			CxProfileName:   types.String(data.CxProfileName),
			Name:            types.StringValue(record.Name),
			SVMName:         types.StringValue(record.SVM.Name),
			SecurityStyle:   types.StringValue(record.SecurityStyle),
			Volume:          types.StringValue(record.Volume.Name),
			ID:              types.Int64Value(int64(record.ID)),
			UnixPermissions: types.Int64Value(int64(record.UnixPermissions)),
		}
		elementTypes := map[string]attr.Type{
			"path": types.StringType,
		}
		elements := map[string]attr.Value{
			"path": types.StringValue(record.NAS.Path),
		}
		objectValue, diags := types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.StorageQtrees[index].NAS = objectValue

		elementTypes = map[string]attr.Type{
			"name": types.StringType,
		}
		elements = map[string]attr.Value{
			"name": types.StringValue(record.User.Name),
		}
		objectValue, diags = types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.StorageQtrees[index].User = objectValue

		elements = map[string]attr.Value{
			"name": types.StringValue(record.Group.Name),
		}
		objectValue, diags = types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.StorageQtrees[index].Group = objectValue

		elementTypes = map[string]attr.Type{
			"name": types.StringType,
			"id":   types.Int64Type,
		}
		elements = map[string]attr.Value{
			"name": types.StringValue(record.ExportPolicy.Name),
			"id":   types.Int64Value(int64(record.ExportPolicy.ID)),
		}
		objectValue, diags = types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.StorageQtrees[index].ExportPolicy = objectValue

	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
