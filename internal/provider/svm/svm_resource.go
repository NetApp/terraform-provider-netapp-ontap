package svm

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

var _ resource.Resource = &SvmResource{}
var _ resource.ResourceWithImportState = &SvmResource{}

// NewSvmResource is a helper function to simplify the provider implementation.
func NewSvmResource() resource.Resource {
	return &SvmResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "svm",
		},
	}
}

// SvmResource defines the resource implementation.
type SvmResource struct {
	config connection.ResourceOrDataSourceConfig
}

// SvmResourceModel describes the resource data model.
type SvmResourceModel struct {
	CxProfileName  types.String `tfsdk:"cx_profile_name"`
	Name           types.String `tfsdk:"name"`
	Ipspace        types.String `tfsdk:"ipspace"`
	SnapshotPolicy types.String `tfsdk:"snapshot_policy"`
	SubType        types.String `tfsdk:"subtype"`
	Comment        types.String `tfsdk:"comment"`
	Language       types.String `tfsdk:"language"`
	Aggregates     []Aggregate  `tfsdk:"aggregates"`
	MaxVolumes     types.String `tfsdk:"max_volumes"`
	ID             types.String `tfsdk:"id"`
}

// Aggregate describes the resource data model.
type Aggregate struct {
	Name string `tfsdk:"name"`
}

// Metadata returns the resource type name.
func (r *SvmResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *SvmResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Svm resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the svm to manage",
				Required:            true,
			},
			"ipspace": schema.StringAttribute{
				MarkdownDescription: "The name of the ipspace to manage",
				Optional:            true,
			},
			"snapshot_policy": schema.StringAttribute{
				MarkdownDescription: "The name of the snapshot policy to manage",
				Optional:            true,
			},
			"subtype": schema.StringAttribute{
				MarkdownDescription: "The subtype for svm to be created",
				Optional:            true,
			},
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment for svm to be created",
				Optional:            true,
			},
			"language": schema.StringAttribute{
				MarkdownDescription: "Language to use for svm",
				Optional:            true,
			},
			"aggregates": schema.SetNestedAttribute{
				Required:            true,
				MarkdownDescription: "List of Aggregates to be assigned use for svm",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the aggregate",
							Required:            true,
						},
					},
				},
			},
			"max_volumes": schema.StringAttribute{
				MarkdownDescription: "Maximum number of volumes that can be created on the svm. Expects an integer or unlimited",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "SVM identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SvmResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected  Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.ProviderConfig = config
}

// Create the resource and sets the initial Terraform state.
func (r *SvmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SvmResourceModel
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var request interfaces.SvmResourceModel
	request.Name = data.Name.ValueString()
	if !data.Ipspace.IsNull() {
		request.Ipspace.Name = data.Ipspace.ValueString()
	}

	if !data.SnapshotPolicy.IsNull() {
		request.SnapshotPolicy.Name = data.SnapshotPolicy.ValueString()
	}

	if !data.SubType.IsNull() {
		request.SubType = data.SubType.ValueString()
	}

	setCommentEmpty := false
	if !data.Comment.IsNull() {
		request.Comment = data.Comment.ValueString()
	} else {
		setCommentEmpty = true
	}

	if !data.Language.IsNull() {
		request.Language = data.Language.ValueString()
	}

	if !data.MaxVolumes.IsNull() {
		err := interfaces.ValidateIntORString(errorHandler, data.MaxVolumes.ValueString(), "unlimited")
		if err != nil {
			return
		}
		request.MaxVolumes = data.MaxVolumes.ValueString()
	}

	setAggrEmpty := false
	if len(data.Aggregates) != 0 {
		aggregates := []interfaces.Aggregate{}
		for _, v := range data.Aggregates {
			aggr := interfaces.Aggregate{}
			aggr.Name = v.Name
			aggregates = append(aggregates, aggr)
		}
		err := mapstructure.Decode(aggregates, &request.Aggregates)
		if err != nil {
			errorHandler.MakeAndReportError("error creating svm", fmt.Sprintf("error on encoding aggregates info: %s, aggregates %#v", err, aggregates))
			return
		}
	} else {
		setAggrEmpty = true
	}

	errorHandler = utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	svm, err := interfaces.CreateSvm(errorHandler, *client, request, setAggrEmpty, setCommentEmpty)
	if err != nil {
		return
	}
	// data.UUID = types.StringValue(svm.UUID)
	data.ID = types.StringValue(svm.UUID)
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *SvmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *SvmResourceModel

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
	tflog.Debug(ctx, fmt.Sprintf("read a svm resource: %#v", data))
	var svm *interfaces.SvmGetDataSourceModel
	if data.ID.ValueString() != "" {
		svm, err = interfaces.GetSvm(errorHandler, *client, data.ID.ValueString())
	} else {
		svm, err = interfaces.GetSvmByNameDataSource(errorHandler, *client, data.Name.ValueString())
	}
	if err != nil {
		return
	}
	if svm == nil {
		errorHandler.MakeAndReportError("No Svm found", "No SVM found")
		return
	}
	data.Name = types.StringValue(svm.Name)
	data.ID = types.StringValue(svm.UUID)
	aggregates := []Aggregate{}

	if len(svm.Aggregates) != 0 {
		for _, v := range svm.Aggregates {
			aggr := Aggregate{}
			aggr.Name = v.Name
			aggregates = append(aggregates, aggr)
		}
		data.Aggregates = aggregates
	}

	if svm.Comment != "" {
		data.Comment = types.StringValue(svm.Comment)
	}

	if svm.Ipspace.Name != "" {
		data.Ipspace = types.StringValue(svm.Ipspace.Name)
	}

	if svm.SnapshotPolicy.Name != "" {
		data.SnapshotPolicy = types.StringValue(svm.SnapshotPolicy.Name)
	}

	if svm.SubType != "" {
		data.SubType = types.StringValue(svm.SubType)
	}

	if svm.Language != "" {
		data.Language = types.StringValue(svm.Language)
	}

	if svm.MaxVolumes != "" {
		data.MaxVolumes = types.StringValue(svm.MaxVolumes)
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SvmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SvmResourceModel
	var state *SvmResourceModel
	setCommentEmpty := false
	setAggrEmpty := false

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	// Read state file data
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	var request interfaces.SvmResourceModel
	if !data.Name.Equal(state.Name) {
		if data.Name.ValueString() == "" {
			errorHandler.MakeAndReportError("update name", "name cannot be updated with empty string")
			return
		}
		request.Name = data.Name.ValueString()
	}
	// TODO: Ipspace can not be modify on SVM/patch. We can't fail or maybe a warning should be sent?
	//if !data.Ipspace.IsNull() {
	//	request.Ipspace.Name = data.Ipspace.ValueString()
	//}

	if !data.SnapshotPolicy.Equal(state.SnapshotPolicy) {
		if data.SnapshotPolicy.ValueString() == "" {
			// snapshot policy cannot be modifoed as empty name but API does not fail with empty snapshot policy
			errorHandler.MakeAndReportError("update snapshot_policy", "snapshot_policy cannot be updated with empty string")
			return
		}
		request.SnapshotPolicy.Name = data.SnapshotPolicy.ValueString()
	}
	if !data.SubType.Equal(state.SubType) {
		errorHandler.MakeAndReportError("update subtype", "subtype cannot be modified")
		return
	}
	// comment can be modified as empty string
	if !data.Comment.Equal(state.Comment) {
		if data.Comment.ValueString() == "" {
			setCommentEmpty = true
		}
		request.Comment = data.Comment.ValueString()
	}

	if !data.Language.Equal(state.Language) {
		if data.Language.ValueString() == "" {
			errorHandler.MakeAndReportError("update language", "language cannot be updated with empty string")
			return
		}
		request.Language = data.Language.ValueString()
	}

	if !data.MaxVolumes.Equal(state.MaxVolumes) {
		if data.MaxVolumes.ValueString() == "" {
			errorHandler.MakeAndReportError("update max_volumes", "max_volumes cannot be updated with empty string")
			return
		}
		err := interfaces.ValidateIntORString(errorHandler, data.MaxVolumes.ValueString(), "unlimited")
		if err != nil {
			return
		}
		request.MaxVolumes = data.MaxVolumes.ValueString()
	}

	// aggregates can be modified as empty list
	aggregates := []interfaces.Aggregate{}
	if len(data.Aggregates) != 0 {
		for _, v := range data.Aggregates {
			aggr := interfaces.Aggregate{}
			aggr.Name = v.Name
			aggregates = append(aggregates, aggr)
		}
	} else {
		if len(state.Aggregates) != 0 {
			setAggrEmpty = true
		}
	}
	err = mapstructure.Decode(aggregates, &request.Aggregates)
	if err != nil {
		errorHandler.MakeAndReportError("error creating svm", fmt.Sprintf("error on encoding aggregates info: %s, aggregates %#v", err, aggregates))
		return
	}

	tflog.Debug(ctx, fmt.Sprintf("update a svm resource: %#v", data))
	err = interfaces.UpdateSvm(errorHandler, *client, request, state.ID.ValueString(), setAggrEmpty, setCommentEmpty)
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SvmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SvmResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("ID is null", "svm UUID is null")
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	err = interfaces.DeleteSvm(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SvmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
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
