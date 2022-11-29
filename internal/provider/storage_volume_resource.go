package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &StorageVolumeResource{}
var _ resource.ResourceWithImportState = &StorageVolumeResource{}

// NewStorageVolumeResource is a helper function to simplify the provider implementation.
func NewStorageVolumeResource() resource.Resource {
	return &StorageVolumeResource{
		config: resourceOrDataSourceConfig{
			name: "storage_volume_resource",
		},
	}
}

// StorageVolumeResource defines the resource implementation.
type StorageVolumeResource struct {
	config resourceOrDataSourceConfig
}

// StorageVolumeResourceModel describes the resource data model.
type StorageVolumeResourceModel struct {
	CxProfileName types.String   `tfsdk:"cx_profile_name"`
	Name          types.String   `tfsdk:"name"`
	Vserver       types.String   `tfsdk:"vserver"`
	Aggregates    []types.String `tfsdk:"aggregates"`
	UUID          types.String   `tfsdk:"uuid"`
}

// Metadata returns the resource type name.
func (r *StorageVolumeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// GetSchema defines the schema for the resource.
func (r *StorageVolumeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Volume resource",

		Attributes: map[string]tfsdk.Attribute{
			"cx_profile_name": {
				MarkdownDescription: "Connection profile name",
				Type:                types.StringType,
				Required:            true,
			},
			"name": {
				MarkdownDescription: "The name of the volume to manage",
				Required:            true,
				Type:                types.StringType,
			},
			"vserver": {
				MarkdownDescription: "Name of the vserver to use",
				Required:            true,
				Type:                types.StringType,
			},
			"aggregates": {
				MarkdownDescription: "List of aggregates in which to create the volume",
				Required:            true,
				Type: types.SetType{
					ElemType: types.StringType,
				},
			},
			"uuid": {
				Computed:            true,
				MarkdownDescription: "Volume identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
}

// Configure adds the provider configured client to the resource.
func (r *StorageVolumeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (r *StorageVolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageVolumeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var request interfaces.StorageVolumeResourceModel
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	aggregates := []interfaces.Aggregate{}
	for _, v := range data.Aggregates {
		aggr := interfaces.Aggregate{}
		aggr.Name = v.ValueString()
		aggregates = append(aggregates, aggr)
	}
	err := mapstructure.Decode(aggregates, &request.Aggregates)
	if err != nil {
		errorHandler.MakeAndReportError("error creating volume", fmt.Sprintf("error on encoding aggregates info: %s, aggregates %#v", err, aggregates))
		return
	}
	request.Name = data.Name.ValueString()
	request.SVM.Name = data.Vserver.ValueString()

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	volume, err := interfaces.CreateStorageVolume(errorHandler, *client, request)
	if err != nil {
		return
	}

	data.UUID = types.StringValue(volume.UUID)

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Read refreshes the Terraform state with the latest data.
func (r *StorageVolumeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *StorageVolumeResourceModel

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

	if data.UUID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "VOlume UUID is null")
		return
	}

	_, err = interfaces.GetStorageVolume(errorHandler, *client, data.UUID.ValueString())
	if err != nil {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *StorageVolumeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *StorageVolumeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *StorageVolumeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *StorageVolumeResourceModel

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

	if data.UUID.IsNull() {
		errorHandler.MakeAndReportError("UUID is null", "VOlume UUID is null")
		return
	}

	err = interfaces.DeleteStorageVolume(errorHandler, *client, data.UUID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *StorageVolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
