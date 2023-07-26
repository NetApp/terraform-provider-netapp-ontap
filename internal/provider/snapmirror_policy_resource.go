package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &SnapmirrorPolicyResource{}
var _ resource.ResourceWithImportState = &SnapmirrorPolicyResource{}

// NewSnapmirrorPolicyResource is a helper function to simplify the provider implementation.
func NewSnapmirrorPolicyResource() resource.Resource {
	return &SnapmirrorPolicyResource{
		config: resourceOrDataSourceConfig{
			name: "snapmirror_policy_resource",
		},
	}
}

// SnapmirrorPolicyResource defines the resource implementation.
type SnapmirrorPolicyResource struct {
	config resourceOrDataSourceConfig
}

// SnapmirrorPolicyResourceModel describes the resource data model.
type SnapmirrorPolicyResourceModel struct {
	CxProfileName             types.String     `tfsdk:"cx_profile_name"`
	Name                      types.String     `tfsdk:"name"`
	SVMName                   types.String     `tfsdk:"svm_name"`
	Type                      types.String     `tfsdk:"type"`
	Comment                   types.String     `tfsdk:"comment"`
	TransferSchedule          types.String     `tfsdk:"transfer_schedule"`
	NetworkCompressionEnabled types.Bool       `tfsdk:"network_compression_enabled"`
	Retention                 []RetentionModel `tfsdk:"retention"`
	IdentityPreservation      types.String     `tfsdk:"identity_preservation"`
	CopyAllSourceSnapshots    types.Bool       `tfsdk:"copy_all_source_snapshots"`
	ID                        types.String     `tfsdk:"id"`
}

// RetentionModel describes retention data model.
type RetentionModel struct {
	CreationSchedule types.String `tfsdk:"creation_schedule"`
	Count            types.Int64  `tfsdk:"count"`
	Label            types.String `tfsdk:"label"`
	Prefix           types.String `tfsdk:"prefix"`
}

// Metadata returns the resource type name
func (r *SnapmirrorPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *SnapmirrorPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SnapmirrorPolicy resource",
		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "SnapmirrorPolicy name",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "SnapmirrorPolicy vserver name",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "SnapmirrorPolicy type",
				Optional:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment associated with the policy.",
				Optional:            true,
			},
			"transfer_schedule": schema.StringAttribute{
				MarkdownDescription: "The schedule used to update asynchronous relationships",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"network_compression_enabled": schema.BoolAttribute{
				MarkdownDescription: "Specifies whether network compression is enabled for transfers",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"retention": schema.ListNestedAttribute{
				Optional:            true,
				MarkdownDescription: "Rules for Snapshot copy retention.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"creation_schedule": schema.StringAttribute{
							MarkdownDescription: "Schedule used to create Snapshot copies on the destination for long term retention.",
							Optional:            true,
						},
						"count": schema.Int64Attribute{
							MarkdownDescription: "Number of Snapshot copies to be kept for retention.",
							Optional:            true,
						},
						"label": schema.StringAttribute{
							MarkdownDescription: "Snapshot copy label",
							Required:            true,
						},
						"prefix": schema.StringAttribute{
							MarkdownDescription: "Specifies the prefix for the Snapshot copy name to be created as per the schedule",
							Optional:            true,
						},
					},
				},
			},
			"identity_preservation": schema.StringAttribute{
				MarkdownDescription: "Specifies which configuration of the source SVM is replicated to the destination SVM.",
				Optional:            true,
			},
			"copy_all_source_snapshots": schema.BoolAttribute{
				MarkdownDescription: "Specifies that all the source Snapshot copies (including the one created by SnapMirror before the transfer begins) should be copied to the destination on a transfer.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SnapmirrorPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please resport this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *SnapmirrorPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SnapmirrorPolicyResourceModel

	// Read Terraform prior state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside New Client
		return
	}

	restInfo, err := interfaces.GetSnapmirrorPolicy(errorHandler, *client, data.Name.ValueString(), data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GETSnapmirrorPolicy
		return
	}

	data.Name = types.StringValue(restInfo.Name)
	data.Comment = types.StringValue(restInfo.Comment)
	data.CopyAllSourceSnapshots = types.BoolValue(restInfo.CopyAllSourceSnapshots)
	data.IdentityPreservation = types.StringValue(restInfo.IdentityPreservation)
	data.Type = types.StringValue(restInfo.Type)
	data.NetworkCompressionEnabled = types.BoolValue(restInfo.NetworkCompressionEnabled)
	if restInfo.Retention == nil {
		data.Retention = nil
	} else {
		data.Retention = []RetentionModel{}
		for _, item := range restInfo.Retention {
			data.Retention = append(data.Retention, RetentionModel{
				CreationSchedule: types.StringValue(item.CreationSchedule.Name),
				Count:            types.Int64Value(int64(item.Count)),
				Label:            types.StringValue(item.Label),
				Prefix:           types.StringValue(item.Prefix),
			})
		}
	}
	data.ID = types.StringValue(restInfo.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *SnapmirrorPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SnapmirrorPolicyResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.SnapmirrorPolicyResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	body.SVMName.Name = data.SVMName.ValueString()
	if !data.IdentityPreservation.IsNull() {
		body.IdentityPreservation = data.IdentityPreservation.ValueString()
	}
	if !data.Comment.IsNull() {
		body.Comment = data.Comment.ValueString()
	}
	if !data.CopyAllSourceSnapshots.IsNull() {
		body.CopyAllSourceSnapshots = data.CopyAllSourceSnapshots.ValueBool()
	}
	if !data.NetworkCompressionEnabled.IsNull() {
		body.NetworkCompressionEnabled = data.NetworkCompressionEnabled.ValueBool()
	}
	if !data.TransferSchedule.IsNull() {
		body.TransferSchedule = data.TransferSchedule.ValueString()
	}
	if !data.Type.IsNull() {
		body.Type = data.Type.ValueString()
	}
	var Retention []interfaces.RetentionGetDataModel
	for _, item := range data.Retention {
		Retention = append(Retention, interfaces.RetentionGetDataModel{
			CreationSchedule: interfaces.CreationScheduleModel{
				Name: item.CreationSchedule.ValueString(),
			},
			Count:  item.Count.ValueInt64(),
			Label:  item.Label.ValueString(),
			Prefix: item.Prefix.ValueString(),
		})
	}
	body.Retention = Retention

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateSnapmirrorPolicy(errorHandler, *client, body)
	if err != nil {
		return
	}
	data.ID = types.StringValue(resource.UUID)

	tflog.Trace(ctx, fmt.Sprintf("created a resource, UUID=%s", data.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SnapmirrorPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *IPInterfaceResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SnapmirrorPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SnapmirrorPolicyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "snapmirror_policy UUID is null")
		return
	}

	err = interfaces.DeleteSnapmirrorPolicy(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SnapmirrorPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
