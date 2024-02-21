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
var _ datasource.DataSource = &SVMPeerDataSource{}

// NewSVMPeerDataSource is a helper function to simplify the provider implementation.
func NewSVMPeerDataSource() datasource.DataSource {
	return &SVMPeerDataSource{
		config: resourceOrDataSourceConfig{
			name: "svm_peer_data_source",
		},
	}
}

// SVMPeerDataSource defines the data source implementation.
type SVMPeerDataSource struct {
	config resourceOrDataSourceConfig
}

type SVMPeerDataSourceModel struct {
	CxProfileName types.String   `tfsdk:"cx_profile_name"`
	Applications  []types.String `tfsdk:"applications"`
	SVM           *SVM           `tfsdk:"svm"`
	Peer          *PeerData      `tfsdk:"peer"`
	ID            types.String   `tfsdk:"id"`
	State         types.String   `tfsdk:"state"`
}

// PeerData describes Peer data model.
type PeerData struct {
	SVM     *SVM     `tfsdk:"svm"`
	Cluster *Cluster `tfsdk:"cluster"`
}

// Metadata returns the data source type name.
func (d *SVMPeerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.name
}

// Schema defines the schema for the data source.
func (d *SVMPeerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SVMPeer data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
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
	}
}

// Configure adds the provider configured client to the data source.
func (d *SVMPeerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *SVMPeerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SVMPeerDataSourceModel

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

	restInfo, err := interfaces.GetSVMPeersBySVMNameAndPeerSvmName(errorHandler, *client, data.SVM.Name.ValueString(), data.Peer.SVM.Name.ValueString())
	if err != nil {
		// error reporting done inside GetSVMPeer
		return
	}

	data.ID = types.StringValue(restInfo.UUID)
	data.State = types.StringValue(restInfo.State)
	var applications []types.String
	for _, e := range restInfo.Applications {
		applications = append(applications, types.StringValue(e))
	}
	data.Applications = applications
	data.Peer = &PeerData{
		SVM:     &SVM{Name: types.StringValue(restInfo.Peer.SVM.Name)},
		Cluster: &Cluster{Name: types.StringValue(restInfo.Peer.Cluster.Name)},
	}
	data.SVM = &SVM{Name: types.StringValue(restInfo.SVM.Name)}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
