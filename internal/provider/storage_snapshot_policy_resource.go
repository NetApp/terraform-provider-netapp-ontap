package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
	"strings"
)

// TODO:
// copy this file to match you resource (should match internal/provider/storage_snapshot_policy_resource.go)
// replace SnapshotPolicy with the name of the resource, following go conventions, eg IPInterface
// replace storage_snapshot_policy with the name of the resource, for logging purposes, eg ip_interface
// make sure to create internal/interfaces/storage_snapshot_policy.go too)
// delete these 5 lines

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &SnapshotPolicyResource{}
var _ resource.ResourceWithImportState = &SnapshotPolicyResource{}

// NewSnapshotPolicyResource is a helper function to simplify the provider implementation.
func NewSnapshotPolicyResource() resource.Resource {
	return &SnapshotPolicyResource{
		config: resourceOrDataSourceConfig{
			name: "storage_snapshot_policy_resource",
		},
	}
}

// SnapshotPolicyResource defines the resource implementation.
type SnapshotPolicyResource struct {
	config resourceOrDataSourceConfig
}

// ScheduleResourceModel describes the schedule data source
type ScheduleResourceModel struct {
	Name types.String `tfsdk:"name"`
}

// CopyResourceModel describe the snapshot copies data model
type CopyResourceModel struct {
	Count           types.Int64           `tfsdk:"count"`
	Schedule        ScheduleResourceModel `tfsdk:"schedule"`
	RetentionPeriod types.String          `tfsdk:"retention_period"`
	SnapmirrorLabel types.String          `tfsdk:"snapmirror_label"`
	Prefix          types.String          `tfsdk:"prefix"`
}

// SnapshotPolicyResourceModel describes the resource data model.
type SnapshotPolicyResourceModel struct {
	CxProfileName types.String        `tfsdk:"cx_profile_name"`
	Name          types.String        `tfsdk:"name"`
	SVMName       types.String        `tfsdk:"svm_name"` // if needed or relevant
	ID            types.String        `tfsdk:"id"`
	Copies        []CopyResourceModel `tfsdk:"copies"`
	Comment       types.String        `tfsdk:"comment"`
	Enabled       types.Bool          `tfsdk:"enabled"`
}

// Metadata returns the resource type name.
func (r *SnapshotPolicyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *SnapshotPolicyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SnapshotPolicy resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "SnapshotPolicy name",
				Required:            true,
			},
			"copies": schema.SetNestedAttribute{
				MarkdownDescription: "Snapshot copy",
				Required:            true,
				PlanModifiers:       []planmodifier.Set{setplanmodifier.RequiresReplace()},
				NestedObject: schema.NestedAttributeObject{
					PlanModifiers: []planmodifier.Object{objectplanmodifier.RequiresReplace()},
					Attributes: map[string]schema.Attribute{
						"count": schema.Int64Attribute{
							MarkdownDescription: "The number of Snapshot copies to maintain for this schedule",
							Required:            true,
							PlanModifiers:       []planmodifier.Int64{int64planmodifier.RequiresReplace()},
						},
						"schedule": schema.SingleNestedAttribute{
							MarkdownDescription: "Schedule at which Snapshot copies are captured on the volume",
							Required:            true,
							PlanModifiers:       []planmodifier.Object{objectplanmodifier.RequiresReplace()},
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Some common schedules already defined in the system are hourly, daily, weekly, at 15 minute intervals, and at 5 minute intervals. Snapshot copy policies with custom schedules can be referenced",
									Required:            true,
									PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
								},
							},
						},
						"retention_period": schema.StringAttribute{
							MarkdownDescription: "The retention period of Snapshot copies for this schedule",
							Optional:            true,
							PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"snapmirror_label": schema.StringAttribute{
							MarkdownDescription: "Label for SnapMirror operations",
							Optional:            true,
							PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
						"prefix": schema.StringAttribute{
							MarkdownDescription: "The prefix to use while creating Snapshot copies at regular intervals",
							Optional:            true,
							PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
						},
					},
				},
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "A comment associated with the Snapshot copy policy",
				Optional:            true,
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Is the Snapshot copy policy enabled?",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(true),
				// not suport update, so force recreate if changes
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.RequiresReplace()},
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "SnapshotPolicy svm name",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "SnapshotPolicy ID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SnapshotPolicyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *SnapshotPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SnapshotPolicyResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	var restInfo *interfaces.SnapshotPolicyGetDataModelONTAP
	if data.ID.ValueString() == "" {
		restInfo, err = interfaces.GetSnapshotPolicyByName(errorHandler, *client, data.Name.ValueString())
		if err != nil {
			return
		}
		if restInfo == nil {
			errorHandler.MakeAndReportError("No snapshot policy found", fmt.Sprintf("snapshot policy  %s not found.", data.Name.ValueString()))
			return
		}
	} else {
		restInfo, err = interfaces.GetSnapshotPolicy(errorHandler, *client, data.ID.ValueString())
		if err != nil {
			// error reporting done inside GetSnapshotPolicy
			return
		}
		if restInfo == nil {
			errorHandler.MakeAndReportError("No snapshot policy found", fmt.Sprintf("snapshot policy  %s not found.", data.Name.ValueString()))
			return
		}
	}
	data.Name = types.StringValue(restInfo.Name)
	data.ID = types.StringValue(restInfo.UUID)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve ID
func (r *SnapshotPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SnapshotPolicyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.SnapshotPolicyResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()

	copies := []interfaces.CopyType{}
	for _, v := range data.Copies {
		onecopy := interfaces.CopyType{}
		onecopy.Count = v.Count.ValueInt64()
		onecopy.Schedule.Name = v.Schedule.Name.ValueString()
		if !v.Prefix.IsNull() {
			onecopy.Prefix = v.Prefix.ValueString()
		}
		if !v.RetentionPeriod.IsNull() {
			onecopy.RetentionPeriod = v.RetentionPeriod.ValueString()
		}
		if !v.SnapmirrorLabel.IsNull() {
			onecopy.SnapmirrorLabel = v.SnapmirrorLabel.ValueString()
		}
		copies = append(copies, onecopy)
	}
	err := mapstructure.Decode(copies, &body.Copies)
	if err != nil {
		errorHandler.MakeAndReportError("error creating snapshot policies", fmt.Sprintf("error on encoding copies info: %s, copies %#v", err, copies))
		return
	}

	if !data.Comment.IsNull() {
		body.Comment = data.Comment.ValueString()
	}
	if !data.Enabled.IsNull() {
		body.Enabled = data.Enabled.ValueBool()
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateSnapshotPolicy(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.ID = types.StringValue(resource.UUID)
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SnapshotPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SnapshotPolicyResourceModel
	var state *SnapshotPolicyResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	// Read Terraform state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

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
		errorHandler.MakeAndReportError("ID is null", "storage_snapshot_policy ID is null")
		return
	}

	var body interfaces.SnapshotPolicyResourceUpdateRequestONTAP
	if !data.Comment.Equal(state.Comment) {
		body.Comment = data.Comment.ValueString()
	}
	if !data.Enabled.Equal(state.Enabled) {
		body.Enabled = data.Enabled.ValueBool()
	}

	err = interfaces.UpdateSnapshotPolicy(errorHandler, *client, body, data.ID.ValueString())
	if err != nil {
		return
	}
	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SnapshotPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SnapshotPolicyResourceModel

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
		errorHandler.MakeAndReportError("ID is null", "storage_snapshot_policy ID is null")
		return
	}

	err = interfaces.DeleteSnapshotPolicy(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SnapshotPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")
	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
}
