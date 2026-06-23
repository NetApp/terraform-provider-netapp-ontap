package storage

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

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageNvmeNamespacesDataSource{}

// NewStorageNvmeNamespacesDataSource is a helper function to simplify the provider implementation.
func NewStorageNvmeNamespacesDataSource() datasource.DataSource {
	return &StorageNvmeNamespacesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "nvme_namespaces",
		},
	}
}

// NewStorageNvmeNamespacesDataSourceAlias is a helper function to simplify the provider implementation.
func NewStorageNvmeNamespacesDataSourceAlias() datasource.DataSource {
	return &StorageNvmeNamespacesDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "storage_nvme_namespaces_data_source",
		},
	}
}

// StorageNvmeNamespacesDataSource defines the data source implementation.
type StorageNvmeNamespacesDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageNvmeNamespacesDataSourceModel describes the data source data model.
type StorageNvmeNamespacesDataSourceModel struct {
	CxProfileName         types.String                                `tfsdk:"cx_profile_name"`
	StorageNvmeNamespaces []StorageNvmeNamespaceDataSourceModel       `tfsdk:"storage_nvme_namespaces"`
	Filter                *StorageNvmeNamespacesDataSourceFilterModel `tfsdk:"filter"`
}

// StorageNvmeNamespacesDataSourceFilterModel describes the data source data model for queries.
type StorageNvmeNamespacesDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the data source type name.
func (d *StorageNvmeNamespacesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *StorageNvmeNamespacesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "StorageNvmeNamespaces data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "StorageNvmeNamespace name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "StorageNvmeNamespace svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"storage_nvme_namespaces": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "StorageNvmeNamespace name",
							Computed:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "StorageNvmeNamespace svm name",
							Computed:            true,
						},
						"os_type": schema.StringAttribute{
							MarkdownDescription: "OS type for namespace. Possible values: aix, linux, vmware, windows",
							Computed:            true,
						},
						"space": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"size": schema.Int64Attribute{
									MarkdownDescription: "Size of namespace in bytes",
									Computed:            true,
								},
								"block_size": schema.Int64Attribute{
									MarkdownDescription: "Block size of namespace in bytes",
									Computed:            true,
								},
							},
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "StorageNvmeNamespace UUID",
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
func (d *StorageNvmeNamespacesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *StorageNvmeNamespacesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageNvmeNamespacesDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		return
	}

	var filter *interfaces.StorageNvmeNamespaceDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.StorageNvmeNamespaceDataSourceFilterModel{
			Name:    data.Filter.Name.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}

	restInfo, err := interfaces.GetStorageNvmeNamespaces(errorHandler, *client, filter)
	if err != nil {
		return
	}

	data.StorageNvmeNamespaces = make([]StorageNvmeNamespaceDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.StorageNvmeNamespaces[index] = StorageNvmeNamespaceDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
			SVMName:       types.StringValue(record.SVM.Name),
			OSType:        types.StringValue(record.OSType),
			Space: &StorageNvmeNamespaceDataSourceSpaceModel{
				Size:      types.Int64Value(record.Space.Size),
				BlockSize: types.Int64Value(record.Space.BlockSize),
			},
			ID: types.StringValue(record.UUID),
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
