package svm

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/snapmirror"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &SVMPeersResource{}
var _ resource.ResourceWithImportState = &SVMPeersResource{}

// NewSVMPeersResource is a helper function to simplify the provider implementation.
func NewSVMPeerResource() resource.Resource {
	return &SVMPeersResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "svm_peer",
		},
	}
}

// SVMPeersResource defines the resource implementation.
type SVMPeersResource struct {
	config connection.ResourceOrDataSourceConfig
}

// SVMPeersResourceModel describes the resource data model.
type SVMPeersResourceModel struct {
	CxProfileName types.String   `tfsdk:"cx_profile_name"`
	Applications  []types.String `tfsdk:"applications"`
	SVM           SVM            `tfsdk:"svm"`
	Peer          Peer           `tfsdk:"peer"`
	ID            types.String   `tfsdk:"id"`
	State         types.String   `tfsdk:"state"`
}

// SVM describes SVM data model.
type SVM struct {
	Name types.String `tfsdk:"name"`
}

// Peer describes Peer data model.
type Peer struct {
	SVM               SVM                `tfsdk:"svm"`
	Cluster           snapmirror.Cluster `tfsdk:"cluster"`
	PeerCxProfileName types.String       `tfsdk:"peer_cx_profile_name"`
}

// Metadata returns the resource type name.
func (r *SVMPeersResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *SVMPeersResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SVMPeers resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"applications": schema.SetAttribute{
				MarkdownDescription: "SVMPeering applications",
				Required:            true,
				ElementType:         types.StringType,
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
						Required:            true,
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								MarkdownDescription: "name of the peer Cluster",
								Required:            true,
							},
						},
					},
					"peer_cx_profile_name": schema.StringAttribute{
						MarkdownDescription: "Peer connection profile name, if not provided, status will be only initiated and need to be accepted from peer side to make the status peered",
						Optional:            true,
					},
				},
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "SVMPeers UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SVMPeersResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *SVMPeersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SVMPeersResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("read a SVMPeer resource: %#v", data))
	var restInfo *interfaces.SVMPeerDataSourceModel
	if data.ID.ValueString() != "" {
		restInfo, err = interfaces.GetSVMPeer(errorHandler, *client, data.ID.ValueString())
		if err != nil {
			// error reporting done inside GetSVMPeer
			return
		}
	} else {
		restInfo, err = interfaces.GetSVMPeersBySVMNameAndPeerSvmName(errorHandler, *client, data.SVM.Name.ValueString(), data.Peer.SVM.Name.ValueString())
		if err != nil {
			// error reporting done inside GetSVMPeersBySVMNameAndPeerSvmName
			return
		}
		data.Peer.Cluster.Name = types.StringValue(restInfo.Peer.Cluster.Name)
	}

	if restInfo == nil {
		errorHandler.MakeAndReportError("error reading info", "No SVM Peer found")
		return
	}

	data.ID = types.StringValue(restInfo.UUID)
	data.State = types.StringValue(restInfo.State)
	var applications []types.String
	for _, e := range restInfo.Applications {
		applications = append(applications, types.StringValue(e))
	}
	data.Applications = applications

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *SVMPeersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SVMPeersResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	var request interfaces.SVMPeerResourceModel
	var applications []string
	for _, e := range data.Applications {
		applications = append(applications, e.ValueString())
	}
	request.Applications = applications
	request.SVM.Name = data.SVM.Name.ValueString()
	request.Peer.SVM.Name = data.Peer.SVM.Name.ValueString()
	request.Peer.Cluster.Name = data.Peer.Cluster.Name.ValueString()

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateSVMPeers(errorHandler, *client, request)
	if err != nil {
		return
	}

	data.ID = types.StringValue(resource.UUID)
	var restInfo *interfaces.SVMPeerDataSourceModel
	if !data.Peer.PeerCxProfileName.IsNull() {
		peerClient, err := connection.GetRestClient(errorHandler, r.config, data.Peer.PeerCxProfileName)
		if err != nil {
			// error reporting done inside NewClient
			return
		}
		var body interfaces.SVMPeerAcceptResourceModel
		body.State = "peered"
		err = interfaces.UpdateSVMPeers(errorHandler, *peerClient, body, data.ID.ValueString())
		if err != nil {
			return
		}
		restInfo, err = interfaces.GetSVMPeer(errorHandler, *client, data.ID.ValueString())
		if err != nil {
			// error reporting done inside GetSVMPeers
			return
		}
	}

	if restInfo != nil {
		data.State = types.StringValue(restInfo.State)
	} else {
		data.State = types.StringValue(resource.State)
	}

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SVMPeersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *SVMPeersResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Read Terraform state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, state.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if !plan.Peer.PeerCxProfileName.IsNull() && state.State.ValueString() == "initiated" {
		peerClient, err := connection.GetRestClient(errorHandler, r.config, plan.Peer.PeerCxProfileName)
		if err != nil {
			// error reporting done inside NewClient
			return
		}
		var bodyPeer interfaces.SVMPeerAcceptResourceModel
		bodyPeer.State = "peered"
		err = interfaces.UpdateSVMPeers(errorHandler, *peerClient, bodyPeer, state.ID.ValueString())
		if err != nil {
			return
		}
	}

	isEqual := reflect.DeepEqual(plan.Applications, state.Applications)

	if plan.Applications != nil && !isEqual {
		var applications []string
		for _, e := range plan.Applications {
			applications = append(applications, e.ValueString())
		}
		var body interfaces.SVMPeerUpdateResourceModel
		body.Applications = applications
		err = interfaces.UpdateSVMPeers(errorHandler, *client, body, plan.ID.ValueString())
		if err != nil {
			return
		}
	}

	restInfo, err := interfaces.GetSVMPeer(errorHandler, *client, plan.ID.ValueString())
	if err != nil {
		// error reporting done inside GetSVMPeers
		return
	}

	plan.State = types.StringValue(restInfo.State)

	tflog.Debug(ctx, fmt.Sprintf("updated svm peer resource: UUID=%s", plan.ID))
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SVMPeersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SVMPeersResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "svm_peers UUID is null")
		return
	}

	err = interfaces.DeleteSVMPeers(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SVMPeersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: svm_name,peer_svm_name,peer_cluster_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm").AtName("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("peer").AtName("svm").AtName("name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("peer").AtName("cluster").AtName("name"), idParts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[3])...)
}
