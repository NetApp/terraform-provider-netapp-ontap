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
var _ datasource.DataSource = &SVMPeersDataSource{}

// NewSVMPeersDataSource is a helper function to simplify the provider implementation.
func NewSVMPeersDataSource() datasource.DataSource {
	return &SVMPeersDataSource{
		config: resourceOrDataSourceConfig{
			name: "svm_peers_data_source",
		},
	}
}

// SVMPeersDataSource defines the data source implementation.
type SVMPeersDataSource struct {
	config resourceOrDataSourceConfig
}

// SVMPeersDataSourceModel describes the data source data model.
type SVMPeersDataSourceModel struct {
	CxProfileName types.String                   `tfsdk:"cx_profile_name"`
	SVMPeers      []SVMPeerDataSourceModel       `tfsdk:"svm_peers"`
	Filter        *SVMPeersDataSourceFilterModel `tfsdk:"filter"`
}

// SVMPeersDataSourceFilterModel describes the data source data model for queries.
type SVMPeersDataSourceFilterModel struct {
	SVM  *SVM      `tfsdk:"svm"`
	Peer *PeerData `tfsdk:"peer"`
}

// Metadata returns the data source type name.
func (d *SVMPeersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *SVMPeersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SVMPeers data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm": schema.SingleNestedAttribute{
						MarkdownDescription: "SVM details for SVMPeer",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "name of the SVM",
								Optional:            true,
							},
						},
					},
					"peer": schema.SingleNestedAttribute{
						MarkdownDescription: "Peer details for SVMPeer",
						Optional:            true,
						Attributes: map[string]schema.Attribute{
							"svm": schema.SingleNestedAttribute{
								MarkdownDescription: "peer SVM details for SVMPeer",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										MarkdownDescription: "name of the peer SVM",
										Optional:            true,
									},
								},
							},
							"cluster": schema.SingleNestedAttribute{
								MarkdownDescription: "peer Cluster details for SVMPeer",
								Optional:            true,
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										MarkdownDescription: "name of the peer Cluster",
										Optional:            true,
									},
								},
							},
						},
					},
				},
				Optional: true,
			},
			"svm_peers": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Computed:            true,
						},
						"applications": schema.SetAttribute{
							ElementType:         types.StringType,
							MarkdownDescription: "SVMPeering applications",
							Computed:            true,
						},
						"svm": schema.SingleNestedAttribute{
							MarkdownDescription: "SVM details for SVMPeer",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "name of the SVM",
									Required:            true,
								},
							},
						},
						"peer": schema.SingleNestedAttribute{
							MarkdownDescription: "Peer details for SVMPeer",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"svm": schema.SingleNestedAttribute{
									MarkdownDescription: "peer SVM details for SVMPeer",
									Required:            true,
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											MarkdownDescription: "name of the peer SVM",
											Required:            true,
										},
									},
								},
								"cluster": schema.SingleNestedAttribute{
									MarkdownDescription: "peer Cluster details for SVMPeer",
									Computed:            true,
									Attributes: map[string]schema.Attribute{
										"name": schema.StringAttribute{
											MarkdownDescription: "name of the peer Cluster",
											Computed:            true,
										},
									},
								},
							},
						},
						"state": schema.StringAttribute{
							Computed: true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "SVMPeers UUID",
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
func (d *SVMPeersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SVMPeersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SVMPeersDataSourceModel

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

	var filter *interfaces.SVMPeerDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.SVMPeerDataSourceFilterModel{
			SVM: interfaces.SVM{},
			Peer: interfaces.Peer{
				SVM:     interfaces.SVM{},
				Cluster: interfaces.Cluster{},
			},
		}
		if data.Filter.Peer != nil {
			if data.Filter.Peer.Cluster != nil {
				filter.Peer.Cluster.Name = data.Filter.Peer.Cluster.Name.ValueString()
			}
			if data.Filter.Peer.SVM != nil {
				filter.Peer.SVM.Name = data.Filter.Peer.SVM.Name.ValueString()
			}
		}
		if data.Filter.SVM != nil {
			filter.SVM.Name = data.Filter.SVM.Name.ValueString()
		}
	}
	restInfo, err := interfaces.GetSvmPeersByName(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetSVMPeers
		return
	}

	data.SVMPeers = make([]SVMPeerDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		var applications []types.String
		for _, e := range record.Applications {
			applications = append(applications, types.StringValue(e))
		}
		data.SVMPeers[index] = SVMPeerDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Applications:  applications,
			SVM:           &SVM{Name: types.StringValue(record.SVM.Name)},
			Peer: &PeerData{
				SVM:     &SVM{Name: types.StringValue(record.Peer.SVM.Name)},
				Cluster: &Cluster{Name: types.StringValue(record.Peer.Cluster.Name)},
			},
			State: types.StringValue(record.State),
			ID:    types.StringValue(record.UUID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
