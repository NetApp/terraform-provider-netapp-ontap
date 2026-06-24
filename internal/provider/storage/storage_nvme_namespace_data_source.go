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
var _ datasource.DataSource = &StorageNvmeNamespaceDataSource{}

// NewStorageNvmeNamespaceDataSource is a helper function to simplify the provider implementation.
func NewStorageNvmeNamespaceDataSource() datasource.DataSource {
	return &StorageNvmeNamespaceDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "nvme_namespace",
		},
	}
}

// NewStorageNvmeNamespaceDataSourceAlias is a helper function to simplify the provider implementation.
func NewStorageNvmeNamespaceDataSourceAlias() datasource.DataSource {
	return &StorageNvmeNamespaceDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "storage_nvme_namespace_data_source",
		},
	}
}

// StorageNvmeNamespaceDataSource defines the data source implementation.
type StorageNvmeNamespaceDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageNvmeNamespaceDataSourceModel describes the data source data model.
type StorageNvmeNamespaceDataSourceModel struct {
	CxProfileName types.String                              `tfsdk:"cx_profile_name"`
	Name          types.String                              `tfsdk:"name"`
	SVMName       types.String                              `tfsdk:"svm_name"`
	OSType        types.String                              `tfsdk:"os_type"`
	Space         *StorageNvmeNamespaceDataSourceSpaceModel `tfsdk:"space"`
	ID            types.String                              `tfsdk:"id"`
}

// StorageNvmeNamespaceDataSourceSpaceModel describes the data source data model for space.
type StorageNvmeNamespaceDataSourceSpaceModel struct {
	Size      types.Int64 `tfsdk:"size"`
	BlockSize types.Int64 `tfsdk:"block_size"`
}

// Metadata returns the data source type name.
func (d *StorageNvmeNamespaceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *StorageNvmeNamespaceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "StorageNvmeNamespace data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name for namespace",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "SVM name for namespace",
				Required:            true,
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
				MarkdownDescription: "Namespace UUID",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *StorageNvmeNamespaceDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *StorageNvmeNamespaceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageNvmeNamespaceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetStorageNvmeNamespaceByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.OSType = types.StringValue(restInfo.OSType)
	data.Space = &StorageNvmeNamespaceDataSourceSpaceModel{
		Size:      types.Int64Value(restInfo.Space.Size),
		BlockSize: types.Int64Value(restInfo.Space.BlockSize),
	}
	data.ID = types.StringValue(restInfo.UUID)

	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
