package snapmirror

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

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SnapmirrorsDataSource{}

// NewSnapmirrorsDataSource is a helper function to simplify the provider implementation.
func NewSnapmirrorsDataSource() datasource.DataSource {
	return &SnapmirrorsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "snapmirrors",
		},
	}
}

// SnapmirrorsDataSource defines the data source implementation.
type SnapmirrorsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// SnapmirrorDataSourceFilterModel describes the data source model.
type SnapmirrorDataSourceFilterModel struct {
	DestinantionPath types.String `tfsdk:"destination_path"`
}

// SnapmirrorsDataSourceModel describes the data source data model.
type SnapmirrorsDataSourceModel struct {
	CxProfileName types.String                     `tfsdk:"cx_profile_name"`
	Snapmirrors   []SnapmirrorDataSourceModel      `tfsdk:"snapmirrors"`
	Filter        *SnapmirrorDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *SnapmirrorsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *SnapmirrorsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Snapmirrors data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"destination_path": schema.StringAttribute{
						MarkdownDescription: "Destination path",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"snapmirrors": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"source": schema.SingleNestedAttribute{
							MarkdownDescription: "Snapmirror source endpoint",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"cluster": schema.SingleNestedAttribute{
									MarkdownDescription: "Cluster details",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											MarkdownDescription: "cluster name",
											Computed:            true,
										},
										"uuid": schema.StringAttribute{
											MarkdownDescription: "cluster UUID",
											Computed:            true,
										},
									},
								},
								"path": schema.StringAttribute{
									MarkdownDescription: "Path to the source endpoint of the SnapMirror relationship",
									Computed:            true,
								},
								"svm": schema.SingleNestedAttribute{
									MarkdownDescription: "Cluster details",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											MarkdownDescription: "svm name",
											Computed:            true,
										},
										"uuid": schema.StringAttribute{
											MarkdownDescription: "svm UUID",
											Computed:            true,
										},
									},
								},
							},
						},
						"destination": schema.SingleNestedAttribute{
							MarkdownDescription: "Snapmirror destination endpoint",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"path": schema.StringAttribute{
									MarkdownDescription: "Path to the destination endpoint of the SnapMirror relationship",
									Computed:            true,
								},
								"svm": schema.SingleNestedAttribute{
									MarkdownDescription: "Cluster details",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											MarkdownDescription: "cluster name",
											Computed:            true,
										},
										"uuid": schema.StringAttribute{
											MarkdownDescription: "cluster UUID",
											Computed:            true,
										},
									},
								},
							},
						},
						"healthy": schema.BoolAttribute{
							MarkdownDescription: "healthy of the relationship",
							Computed:            true,
						},
						"restore": schema.BoolAttribute{
							MarkdownDescription: "restore of the relationship",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "uuid of the relationship",
							Computed:            true,
						},
						"state": schema.StringAttribute{
							MarkdownDescription: "state of the relationship",
							Computed:            true,
						},
						"policy": schema.SingleNestedAttribute{
							MarkdownDescription: "policy of the relationship",
							Computed:            true,
							Attributes: map[string]schema.Attribute{
								"uuid": schema.StringAttribute{
									MarkdownDescription: "Policy UUID",
									Computed:            true,
								},
							},
						},
						"group_type": schema.StringAttribute{
							MarkdownDescription: "group_type of the relationship",
							Computed:            true,
						},
						"throttle": schema.Int64Attribute{
							MarkdownDescription: "throttle of the relationship",
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
func (d *SnapmirrorsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SnapmirrorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SnapmirrorsDataSourceModel

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

	var filter *interfaces.SnapmirrorFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.SnapmirrorFilterModel{
			DestinationPath: data.Filter.DestinantionPath.ValueString(),
		}
	}
	restInfo, err := interfaces.GetSnapmirrors(errorHandler, *client, filter, cluster.Version)
	if err != nil {
		// error reporting done inside GetSnapmirrors
		return
	}

	data.Snapmirrors = make([]SnapmirrorDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		data.Snapmirrors[index] = SnapmirrorDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Source: &Source{
				Cluster: &SnapmirrorCluster{
					Name: types.StringValue(record.Source.Cluster.Name),
					UUID: types.StringValue(record.Source.Cluster.UUID),
				},
				Path: types.StringValue(record.Source.Path),
				Svm: &Svm{
					Name: types.StringValue(record.Source.Svm.Name),
					UUID: types.StringValue(record.Source.Svm.UUID),
				},
			},
			Destination: &Destination{
				Path: types.StringValue(record.Destination.Path),
				Svm: &Svm{
					Name: types.StringValue(record.Destination.Svm.Name),
					UUID: types.StringValue(record.Destination.Svm.UUID),
				},
			},
			Healthy: types.BoolValue(record.Healthy),
			Restore: types.BoolValue(record.Restore),
			ID:      types.StringValue(record.UUID),
			State:   types.StringValue(record.State),
		}

		if cluster.Version.Generation == 9 && cluster.Version.Major > 10 {
			data.Snapmirrors[index].Throttle = types.Int64Value(int64(record.Throttle))
			data.Snapmirrors[index].GroupType = types.StringValue(record.GroupType)
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
