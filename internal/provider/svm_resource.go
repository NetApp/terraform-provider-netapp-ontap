package provider

import (
	"context"
	"fmt"

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
		config: resourceOrDataSourceConfig{
			name: "svm_resource",
		},
	}
}

// SvmResource defines the resource implementation.
type SvmResource struct {
	config resourceOrDataSourceConfig
}

// SvmResourceModel describes the resource data model.
type SvmResourceModel struct {
	CxProfileName  types.String   `tfsdk:"cx_profile_name"`
	Name           types.String   `tfsdk:"name"`
	UUID           types.String   `tfsdk:"uuid"`
	Ipspace        types.String   `tfsdk:"ipspace"`
	SnapshotPolicy types.String   `tfsdk:"snapshot_policy"`
	SubType        types.String   `tfsdk:"subtype"`
	Comment        types.String   `tfsdk:"comment"`
	Language       types.String   `tfsdk:"language"`
	Aggregates     []types.String `tfsdk:"aggregates"`
	MaxVolumes     types.String   `tfsdk:"max_volumes"`
	ID             types.String   `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *SvmResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
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
			"uuid": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "svm identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
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
			"aggregates": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Aggregates to be assigned use for svm",
				Optional:            true,
			},
			"max_volumes": schema.StringAttribute{
				MarkdownDescription: "Maximum number of volumes that can be created on the svm. Expects an integer or unlimited",
				Optional:            true,
			},
			"id": schema.StringAttribute{
				Computed: true,
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
	config, ok := req.ProviderData.(Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected  Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.providerConfig = config
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

	if !data.Comment.IsNull() {
		request.Comment = data.Comment.ValueString()
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

	if len(data.Aggregates) != 0 {
		aggregates := []interfaces.Aggregate{}
		for _, v := range data.Aggregates {
			aggr := interfaces.Aggregate{}
			aggr.Name = v.ValueString()
			aggregates = append(aggregates, aggr)
		}
		err := mapstructure.Decode(aggregates, &request.Aggregates)
		if err != nil {
			errorHandler.MakeAndReportError("error creating svm", fmt.Sprintf("error on encoding aggregates info: %s, aggregates %#v", err, aggregates))
			return
		}

	}

	errorHandler = utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	svm, err := interfaces.CreateSvm(errorHandler, *client, request)
	if err != nil {
		return
	}
	data.UUID = types.StringValue(svm.UUID)
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
	if data.UUID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "svm UUID is null")
		return
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	svm, err := interfaces.GetSvm(errorHandler, *client, data.UUID.ValueString())
	if err != nil {
		return
	}
	if svm == nil {
		errorHandler.MakeAndReportError("No Svm found", fmt.Sprintf("No SVM found"))
	}
	data.Name = types.StringValue(svm.Name)
	data.UUID = types.StringValue(svm.UUID)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SvmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SvmResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.UUID.IsNull() {
		errorHandler.MakeAndReportError("UUID is Null", "UUID is NUll")
	}

	var request interfaces.SvmResourceModel
	request.Name = data.Name.ValueString()
	// TODO: Ipspace can not be modify on SVM/patch. We can't fail or maybe a warning should be sent?
	//if !data.Ipspace.IsNull() {
	//	request.Ipspace.Name = data.Ipspace.ValueString()
	//}

	if !data.SnapshotPolicy.IsNull() {
		request.SnapshotPolicy.Name = data.SnapshotPolicy.ValueString()
	}

	if !data.SubType.IsNull() {
		request.SubType = data.SubType.ValueString()
	}

	if !data.Comment.IsNull() {
		request.Comment = data.Comment.ValueString()
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

	if len(data.Aggregates) != 0 {
		aggregates := []interfaces.Aggregate{}
		for _, v := range data.Aggregates {
			aggr := interfaces.Aggregate{}
			aggr.Name = v.ValueString()
			aggregates = append(aggregates, aggr)
		}
		err := mapstructure.Decode(aggregates, &request.Aggregates)
		if err != nil {
			errorHandler.MakeAndReportError("error creating svm", fmt.Sprintf("error on encoding aggregates info: %s, aggregates %#v", err, aggregates))
			return
		}
	}
	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}
	// We need to check to see if the name has changed for a rename
	rename := false
	existingSvm, err := interfaces.GetSvm(errorHandler, *client, data.UUID.ValueString())
	if err != nil {
		return
	}
	if existingSvm.Name != request.Name {
		rename = true
	}
	err = interfaces.UpdateSvm(errorHandler, *client, request, data.UUID.ValueString(), rename)
	if err != nil {
		return
	}
	data.ID = types.StringValue(data.UUID.ValueString())

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
	if data.UUID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "svm UUID is null")
		return
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	err = interfaces.DeleteSvm(errorHandler, *client, data.UUID.ValueString())
	if err != nil {
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SvmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
