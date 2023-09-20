package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &SnapmirrorDataSource{}

// NewSnapmirrorDataSource is a helper function to simplify the provider implementation.
func NewSnapmirrorDataSource() datasource.DataSource {
	return &SnapmirrorDataSource{
		config: resourceOrDataSourceConfig{
			name: "snapmirror_data_source",
		},
	}
}

// SnapmirrorDataSource defines the data source implementation.
type SnapmirrorDataSource struct {
	config resourceOrDataSourceConfig
}

// SnapmirrorDataSourceModel describes the data source data model.
type SnapmirrorDataSourceModel struct {
	CxProfileName types.String      `tfsdk:"cx_profile_name"`
	Source        *Source           `tfsdk:"source"`
	Destination   *Destination      `tfsdk:"destination"`
	Healthy       types.Bool        `tfsdk:"healthy"`
	Restore       types.Bool        `tfsdk:"restore"`
	UUID          types.String      `tfsdk:"uuid"`
	State         types.String      `tfsdk:"state"`
	Policy        *SnapmirrorPolicy `tfsdk:"policy"`
	GroupType     types.String      `tfsdk:"group_type"`
	Throttle      types.Int64       `tfsdk:"throttle"`
}

// Source describes data source model
type Source struct {
	Cluster *SnapmirrorCluster `tfsdk:"cluster"`
	Path    types.String       `tfsdk:"path"`
	Svm     *Svm               `tfsdk:"svm"`
}

// Destination describes data source model
type Destination struct {
	Path types.String `tfsdk:"path"`
	Svm  *Svm         `tfsdk:"svm"`
}

// Svm describes data source model
type Svm struct {
	Name types.String `tfsdk:"name"`
	UUID types.String `tfsdk:"uuid"`
}

// SnapmirrorCluster describes data source model
type SnapmirrorCluster struct {
	Name types.String `tfsdk:"name"`
	UUID types.String `tfsdk:"uuid"`
}

// SnapmirrorPolicy describes data source model
type SnapmirrorPolicy struct {
	UUID types.String `tfsdk:"uuid"`
}

// Metadata returns the data source type name.
func (d *SnapmirrorDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *SnapmirrorDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Snapmirror data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
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
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"path": schema.StringAttribute{
						MarkdownDescription: "Path to the destination endpoint of the SnapMirror relationship",
						Required:            true,
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
			"uuid": schema.StringAttribute{
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *SnapmirrorDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	d.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *SnapmirrorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SnapmirrorDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, d.config, data.CxProfileName)
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
		errorHandler.MakeAndReportError("No cluster found", fmt.Sprintf("cluster not found"))
		return
	}

	restInfo, err := interfaces.GetSnapmirrorByDestinationPath(errorHandler, *client, data.Destination.Path.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetSnapmirror
		return
	}

	data = SnapmirrorDataSourceModel{
		CxProfileName: data.CxProfileName,
		Source: &Source{
			Cluster: &SnapmirrorCluster{
				Name: types.StringValue(restInfo.Source.Cluster.Name),
				UUID: types.StringValue(restInfo.Source.Cluster.UUID),
			},
			Path: types.StringValue(restInfo.Source.Path),
			Svm: &Svm{
				Name: types.StringValue(restInfo.Source.Svm.Name),
				UUID: types.StringValue(restInfo.Source.Svm.UUID),
			},
		},
		Destination: &Destination{
			Path: types.StringValue(restInfo.Destination.Path),
			Svm: &Svm{
				Name: types.StringValue(restInfo.Destination.Svm.Name),
				UUID: types.StringValue(restInfo.Destination.Svm.UUID),
			},
		},
		Healthy: types.BoolValue(restInfo.Healthy),
		Restore: types.BoolValue(restInfo.Restore),
		UUID:    types.StringValue(restInfo.UUID),
		State:   types.StringValue(restInfo.State),
	}

	if cluster.Version.Generation == 9 && cluster.Version.Major > 10 {
		data.Throttle = types.Int64Value(int64(restInfo.Throttle))
		data.GroupType = types.StringValue(restInfo.GroupType)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
