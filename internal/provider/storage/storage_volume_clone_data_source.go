package storage

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
	
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageVolumeCloneDataSource{}

// NewStorageVolumeCloneDataSource is a helper function to simplify the provider implementation.
func NewStorageVolumeCloneDataSource() datasource.DataSource {
	return &StorageVolumeCloneDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "volume_clone",
		},
	}
}

// StorageVolumeCloneDataSource defines the data source implementation.
type StorageVolumeCloneDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageVolumeCloneDataSourceModel describes the data source data model.
type StorageVolumeCloneDataSourceModel struct {
	CxProfileName types.String                       `tfsdk:"cx_profile_name"`
	SVMName       types.String                       `tfsdk:"svm_name"`
	Name          types.String                       `tfsdk:"name"`
	Type          types.String                       `tfsdk:"type"`
	Clone         *StorageVolumeCloneDataSourceClone `tfsdk:"clone"`
	NAS           *StorageVolumeCloneDataSourceNAS   `tfsdk:"nas"`
	ID            types.String                       `tfsdk:"uuid"`
}

// StorageVolumeCloneDataSourceClone describes the clone model.
type StorageVolumeCloneDataSourceClone struct {
	IsFlexclone    types.Bool   `tfsdk:"is_flexclone"`
	Split 		   types.Bool   `tfsdk:"split"`
	ParentVolume   types.String `tfsdk:"parent_volume"`
	ParentSnapshot types.String `tfsdk:"parent_snapshot"`
	ParentSVM      types.String `tfsdk:"parent_svm"`
}

// StorageVolumeCloneDataSourceNAS describes the NAS model.
type StorageVolumeCloneDataSourceNAS struct {
	JunctionPath types.String `tfsdk:"junction_path"`
	GroupID      types.Int64  `tfsdk:"group_id"`
	UserID       types.Int64  `tfsdk:"user_id"`
}

// StorageVolumeCloneDataSourceName describes the resource name model.
type StorageVolumeCloneDataSourceName struct {
	Name types.String `tfsdk:"name"`
}

// StorageVolumeCloneDataSourceFilterModel describes the data source data model for queries.
type StorageVolumeCloneDataSourceFilterModel struct {
	SVMName types.String `tfsdk:"svm_name"`
	Name    types.String `tfsdk:"name"`
}

// Metadata returns the data source type name.
func (d *StorageVolumeCloneDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *StorageVolumeCloneDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageVolumeClone data source",

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
				MarkdownDescription: "Specifies the name of the volume clone.",
				Required:            true,
			},
			"clone": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"is_flexclone": schema.BoolAttribute{
						MarkdownDescription: "Specifies if this volume is a normal FlexVol or FlexClone.",
						Computed:            true,
					},
					"split": schema.BoolAttribute{
						MarkdownDescription: "This field is set when split is executed on any FlexClone.",
						Computed:            true,
					},
					"parent_svm": schema.StringAttribute{
						MarkdownDescription: "SVM of parent volume in which clone is created off.",
						Computed:            true,
					},
					"parent_volume": schema.StringAttribute{
						MarkdownDescription: "Parent volume of the clone.",
						Computed:            true,
					},
					"parent_snapshot": schema.StringAttribute{
						MarkdownDescription: "Parent snapshot of the clone.",
						Computed:            true,
					},
				},
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The type of the volume clone",
				Computed:            true,
			},
			"nas": schema.SingleNestedAttribute{
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"junction_path": schema.StringAttribute{
						MarkdownDescription: "Junction path of the clone.",
						Computed:            true,
					},
					"group_id": schema.Int64Attribute{
						MarkdownDescription: "The UNIX group ID for the clone volume.",
						Computed:            true,
					},
					"user_id": schema.Int64Attribute{
						MarkdownDescription: "The UNIX user ID for the clone volume.",
						Computed:            true,
					},
				},
			},
			"uuid": schema.StringAttribute{
				MarkdownDescription: "UUID of the volume clone.",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *StorageVolumeCloneDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *StorageVolumeCloneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageVolumeCloneDataSourceModel

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

	restInfo, err := interfaces.GetVolumeClone(errorHandler, *client, data.SVMName.ValueString(), data.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetVolumeClone
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("No volume clone found", fmt.Sprintf("Volume clone named %s not found", data.Name.ValueString()))
		return
	}
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	data.Name = types.StringValue(restInfo.Name)
	data.Type = types.StringValue(restInfo.Type)

	// Clone
	data.Clone = &StorageVolumeCloneDataSourceClone{
		IsFlexclone:    types.BoolValue(restInfo.Clone.IsFlexclone),
		Split:          types.BoolValue(!restInfo.Clone.IsFlexclone),
		ParentVolume:   types.StringNull(),
		ParentSnapshot: types.StringNull(),
		ParentSVM:      types.StringNull(),
	}
	if restInfo.Clone.ParentVolume.Name != "" {
		data.Clone.ParentVolume = types.StringValue(restInfo.Clone.ParentVolume.Name)
	}
	if restInfo.Clone.ParentSnapshot != nil && restInfo.Clone.ParentSnapshot.Name != "" {
		data.Clone.ParentSnapshot = types.StringValue(restInfo.Clone.ParentSnapshot.Name)
	}
	if restInfo.Clone.ParentSVM.Name != "" {
		data.Clone.ParentSVM = types.StringValue(restInfo.Clone.ParentSVM.Name)
	}

	// NAS
	data.NAS = &StorageVolumeCloneDataSourceNAS{
		JunctionPath: types.StringNull(),
		GroupID:      types.Int64Null(),
		UserID:       types.Int64Null(),
	}
	if restInfo.NAS.JunctionPath != nil {
		data.NAS.JunctionPath = types.StringValue(*restInfo.NAS.JunctionPath)
	}
	if restInfo.NAS.GroupID != nil {
		data.NAS.GroupID = types.Int64Value(int64(*restInfo.NAS.GroupID))
	}
	if restInfo.NAS.UserID != nil {
		data.NAS.UserID = types.Int64Value(int64(*restInfo.NAS.UserID))
	}

	data.ID = types.StringValue(restInfo.ID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
