package cluster

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/boolvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ClusterPeersResource{}
var _ resource.ResourceWithImportState = &ClusterPeersResource{}

// NewClusterPeersResource is a helper function to simplify the provider implementation.
func NewClusterPeersResource() resource.Resource {
	return &ClusterPeersResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "cluster_peers",
		},
	}
}

// ClusterPeersResource defines the resource implementation.
type ClusterPeersResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ClusterPeersResourceModel describes the resource data model.
type ClusterPeersResourceModel struct {
	CxProfileName      types.String   `tfsdk:"cx_profile_name"`
	Passphrase         types.String   `tfsdk:"passphrase"`
	Name               types.String   `tfsdk:"name"`
	Remote             *Remote        `tfsdk:"remote"`
	SourceDetails      *Remote        `tfsdk:"source_details"`
	PeerCxProfileName  types.String   `tfsdk:"peer_cx_profile_name"`
	GeneratePassphrase types.Bool     `tfsdk:"generate_passphrase"`
	PeerApplications   []types.String `tfsdk:"peer_applications"`
	State              types.String   `tfsdk:"state"`
	PeerID             types.String   `tfsdk:"peer_id"`
	ID                 types.String   `tfsdk:"id"`
}

// Remote describes Remote data model.
type Remote struct {
	IPAddresses []types.String `tfsdk:"ip_addresses"`
}

// Status describes the status data model.
type Status struct {
	State types.String `tfsdk:"state"`
}

// Metadata returns the resource type name.
func (r *ClusterPeersResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ClusterPeersResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ClusterPeers resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"passphrase": schema.StringAttribute{
				MarkdownDescription: "User generated passphrase for use in authentication",
				Optional:            true,
				Sensitive:           true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the peering relationship or name of the remote peer",
				Optional:            true,
			},
			"generate_passphrase": schema.BoolAttribute{
				MarkdownDescription: "When true, ONTAP automatically generates a passphrase to authenticate cluster peers",
				Optional:            true,
				Validators: []validator.Bool{
					boolvalidator.ConflictsWith(path.Expressions{
						path.MatchRoot("passphrase"),
					}...),
				},
			},
			"remote": schema.SingleNestedAttribute{
				MarkdownDescription: "Remote cluster details for cluster peer",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"ip_addresses": schema.SetAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "list of the remote ip addresses",
						Required:            true,
					},
				},
			},
			"source_details": schema.SingleNestedAttribute{
				MarkdownDescription: "Source cluster details for cluster peer from remote cluster",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"ip_addresses": schema.SetAttribute{
						ElementType:         types.StringType,
						MarkdownDescription: "list of the source ip addresses",
						Required:            true,
					},
				},
			},
			"peer_applications": schema.SetAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "SVM peering applications",
				Optional:            true,
			},
			"peer_cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Peer connection profile name, to be accepted from peer side to make the status OK",
				Required:            true,
			},
			"state": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"peer_id": schema.StringAttribute{
				MarkdownDescription: "ClusterPeers destination UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "ClusterPeers source UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ClusterPeersResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ClusterPeersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ClusterPeersResourceModel

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

	tflog.Debug(ctx, fmt.Sprintf("read a ClusterPeer resource: %#v", data))
	var restInfo *interfaces.ClusterPeerGetDataModelONTAP
	if data.ID.ValueString() != "" {
		restInfo, err = interfaces.GetClusterPeer(errorHandler, *client, data.ID.ValueString())
		if err != nil {
			// error reporting done inside GetClusterPeer
			return
		}
	} else {
		restInfo, err = interfaces.GetClusterPeerByName(errorHandler, *client, data.Name.ValueString())
		if err != nil {
			// error reporting done inside GetClusterPeerByName
			return
		}
	}

	if restInfo == nil {
		errorHandler.MakeAndReportError("error reading info", "No Cluster Peer found")
		return
	}

	data.ID = types.StringValue(restInfo.UUID)
	var ipAddresses []types.String
	for _, e := range restInfo.Remote.IPAddress {
		ipAddresses = append(ipAddresses, types.StringValue(e))
	}
	if data.Remote == nil {
		data.Remote = &Remote{}
	}
	data.Remote.IPAddresses = ipAddresses
	data.State = types.StringValue(restInfo.Authentication.State)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *ClusterPeersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ClusterPeersResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ClusterPeersResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	if !data.Name.IsUnknown() {
		body.Name = data.Name.ValueString()
	}
	if data.PeerApplications != nil {
		var applications []string
		for _, e := range data.PeerApplications {
			applications = append(applications, e.ValueString())
		}
		body.PeerApplications = applications
	}
	var ipAddresses []string
	for _, e := range data.Remote.IPAddresses {
		ipAddresses = append(ipAddresses, e.ValueString())
	}
	body.Remote.IPAddress = ipAddresses
	if !data.GeneratePassphrase.IsUnknown() {
		body.Authentication.GeneratePassphrase = data.GeneratePassphrase.ValueBool()
	}
	if !data.Passphrase.IsUnknown() {
		body.Authentication.Passphrase = data.Passphrase.ValueString()
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateClusterPeers(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.ID = types.StringValue(resource.UUID)
	peerClient, err := connection.GetRestClient(errorHandler, r.config, data.PeerCxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	var bodyPeer interfaces.ClusterPeersResourceBodyDataModelONTAP
	var ipAddressesPeer []string
	for _, e := range data.SourceDetails.IPAddresses {
		ipAddressesPeer = append(ipAddressesPeer, e.ValueString())
	}
	bodyPeer.Remote.IPAddress = ipAddressesPeer
	if data.PeerApplications != nil {
		var applications []string
		for _, e := range data.PeerApplications {
			applications = append(applications, e.ValueString())
		}
		bodyPeer.PeerApplications = applications
	}
	bodyPeer.Authentication.Passphrase = resource.Authentication.Passphrase
	resourcePeer, err := interfaces.CreateClusterPeers(errorHandler, *peerClient, bodyPeer)
	if err != nil {
		return
	}
	data.PeerID = types.StringValue(resourcePeer.UUID)

	var restInfo *interfaces.ClusterPeerGetDataModelONTAP
	restInfo, err = interfaces.GetClusterPeer(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		// error reporting done inside GetSVMPeers
		return
	}
	if restInfo == nil {
		errorHandler.MakeAndReportError("error reading info", "No Cluster Peer found")
		return
	}
	data.State = types.StringValue(restInfo.Authentication.State)

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ClusterPeersResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state, plan *ClusterPeersResourceModel

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

	isEqual := reflect.DeepEqual(plan.Remote.IPAddresses, state.Remote.IPAddresses)

	if plan.Remote.IPAddresses != nil && !isEqual {
		var ipAddresses []string
		for _, e := range plan.Remote.IPAddresses {
			ipAddresses = append(ipAddresses, e.ValueString())
		}
		var body interfaces.ClusterPeersResourceBodyDataModelONTAP
		body.Remote.IPAddress = ipAddresses
		err = interfaces.UpdateClusterPeers(errorHandler, *client, body, plan.ID.ValueString())
		if err != nil {
			return
		}
	}

	restInfo, err := interfaces.GetClusterPeer(errorHandler, *client, plan.ID.ValueString())
	if err != nil {
		// error reporting done inside GetClusterPeer
		return
	}

	plan.State = types.StringValue(restInfo.Authentication.State)
	var ipAddresses []types.String
	for _, e := range restInfo.Remote.IPAddress {
		ipAddresses = append(ipAddresses, types.StringValue(e))
	}
	if plan.Remote == nil {
		plan.Remote = &Remote{}
	}
	plan.Remote.IPAddresses = ipAddresses

	tflog.Debug(ctx, fmt.Sprintf("updated svm peer resource: UUID=%s", plan.ID))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ClusterPeersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ClusterPeersResourceModel

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
		errorHandler.MakeAndReportError("UUID is null", "cluster_peers UUID is null")
		return
	}

	err = interfaces.DeleteClusterPeers(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

	// Delete remote peer
	peerClient, err := connection.GetRestClient(errorHandler, r.config, data.PeerCxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "cluster_peers UUID is null")
		return
	}

	err = interfaces.DeleteClusterPeers(errorHandler, *peerClient, data.PeerID.ValueString())
	if err != nil {
		return
	}

	tflog.Trace(ctx, "deleted a resource")

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ClusterPeersResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[1])...)
}
