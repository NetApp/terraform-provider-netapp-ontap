package storage

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageQtreeDataSource{}

// NewStorageQtreeDataSource is a helper function to simplify the provider implementation.
func NewStorageQtreeDataSource() datasource.DataSource {
	return &StorageQtreeDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "storage_qtree",
		},
	}
}

// StorageQtreeDataSource defines the data source implementation.
type StorageQtreeDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageQtreeDataSourceModel describes the data source data model.
type StorageQtreeDataSourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	SVMName       types.String `tfsdk:"svm_name"`
	SecurityStyle types.String `tfsdk:"security_style"`
	NAS           types.Object `tfsdk:"nas"`
	User          types.Object `tfsdk:"user"`
	Volume        types.String `tfsdk:"volume_name"`
}

type StorageQtreeDataSourceNASModel struct {
	Path types.String `tfsdk:"path"`
}

// Metadata returns the data source type name.
func (d *StorageQtreeDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *StorageQtreeDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "StorageQtree data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "StorageQtree name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "IPInterface svm name",
				Required:            true,
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
					"id": schema.StringAttribute{
						MarkdownDescription: "The numeric ID of the user who owns the qtree.",
						Computed:            true,
					},
				},
			},
			"volume_name": schema.StringAttribute{
				MarkdownDescription: "The volume that contains the qtree.",
				Required:            true,
			},
			// qos_policy gives error in the API swagger when this option is specified.
			// Commented out for now, further investigation needed.

			// "qos_policy": schema.SingleNestedAttribute{
			// 	MarkdownDescription: "When min_throughput_iops, min_throughput_mbps, max_throughput_iops or max_throughput_mbps attributes are specified, " +
			// 		"the storage object is assigned to an auto-generated QoS policy group. If the attributes are later modified, the auto-generated QoS policy-group attributes are modified." +
			// 		"Attributes can be removed by specifying 0 and policy group by specifying none. Upon deletion of the storage object or if the attributes are removed, then the QoS policy-group is also removed.",
			// 	Computed: true,
			// 	Attributes: map[string]schema.Attribute{
			// 		"max_throughput_iops": schema.Int64Attribute{
			// 			MarkdownDescription: "Specifies the maximum throughput in IOPS, 0 means none.",
			// 			Computed:            true,
			// 		},
			// 		"max_throughput_mbps": schema.Int64Attribute{
			// 			MarkdownDescription: "Specifies the maximum throughput in Megabytes per sec, 0 means none.",
			// 			Computed:            true,
			// 		},
			// 		"min_throughput_iops": schema.Int64Attribute{
			// 			MarkdownDescription: "Specifies the minimum throughput in IOPS, 0 means none. Setting min_throughput is supported on AFF platforms only, unless FabricPool tiering policies are set.",
			// 			Computed:            true,
			// 		},
			// 		"min_throughput_mbps": schema.Int64Attribute{
			// 			MarkdownDescription: "Specifies the minimum throughput in Megabytes per sec, 0 means none.",
			// 			Computed:            true,
			// 		},
			// 		"name": schema.StringAttribute{
			// 			MarkdownDescription: "The QoS policy group name.",
			// 			Computed:            true,
			// 		},
			// 		"id": schema.StringAttribute{
			// 			MarkdownDescription: "The QoS policy group UUID.",
			// 			Computed:            true,
			// 		},
			// 	},
			// },
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *StorageQtreeDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *StorageQtreeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageQtreeDataSourceModel

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

	restInfo, err := interfaces.GetStorageQtreeByName(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString(), data.Volume.ValueString())
	if err != nil {
		// error reporting done inside GetStorageQtree
		return
	}

	data.Name = types.StringValue(restInfo.Name)

	data.SecurityStyle = types.StringValue(restInfo.SecurityStyle)
	data.Volume = types.StringValue(restInfo.Volume.Name)

	elementTypes := map[string]attr.Type{
		"path": types.StringType,
	}
	elements := map[string]attr.Value{
		"path": types.StringValue(restInfo.NAS.Path),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.NAS = objectValue

	elementTypes = map[string]attr.Type{
		"name": types.StringType,
		"id":   types.StringType,
	}
	elements = map[string]attr.Value{
		"name": types.StringValue(restInfo.User.Name),
		"id":   types.StringValue(restInfo.User.ID),
	}
	objectValue, diags = types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.User = objectValue

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
