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
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &StorageVolumeResource{}
var _ resource.ResourceWithImportState = &StorageVolumeResource{}

// NewStorageVolumeResource is a helper function to simplify the provider implementation.
func NewStorageVolumeResource() resource.Resource {
	return &StorageVolumeResource{
		name: "storage_volume_resource",
	}
}

// StorageVolumeResource defines the resource implementation.
type StorageVolumeResource struct {
	client *restclient.RestClient
	config Config
	name   string
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
	resp.TypeName = req.ProviderTypeName + "_" + r.name
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
	r.config = config
}

// Create creates the resource and sets the initial Terraform state.
func (r *StorageVolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageVolumeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var request interfaces.StorageVolumeResourceModel

	aggregates := []interfaces.Aggregate{}
	for _, v := range data.Aggregates {
		aggr := interfaces.Aggregate{}
		aggr.Name = v.ValueString()
		aggregates = append(aggregates, aggr)
	}
	err := mapstructure.Decode(aggregates, &request.Aggregates)
	if err != nil {
		msg := fmt.Sprintf("error decode data - error: %s", err)
		tflog.Error(ctx, msg)
		// TODO: diags.Error is not reporting anything here.  Works in the caller.
		resp.Diagnostics.AddError("error creating storage/volumes", msg)
		return
	}
	request.Name = data.Name.ValueString()
	request.SVM.Name = data.Vserver.ValueString()

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := r.getClient(ctx, resp.Diagnostics, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	volume, err := interfaces.CreateStorageVolume(ctx, resp.Diagnostics, *client, request)
	if err != nil {
		msg := fmt.Sprintf("error creating storage/volumes: %s", err)
		tflog.Error(ctx, msg)
		resp.Diagnostics.AddError("error creating storage/volumes", msg)
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

	client, err := r.getClient(ctx, resp.Diagnostics, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.UUID.IsNull() {
		msg := "UUID is null"
		tflog.Error(ctx, msg)
		return
	}

	_, err = interfaces.GetStorageVolume(ctx, resp.Diagnostics, *client, data.UUID.ValueString())
	if err != nil {
		msg := fmt.Sprintf("error reading storage/volumes: %s", err)
		tflog.Error(ctx, msg)
		resp.Diagnostics.AddError("error reading storage/volumes", msg)
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

	client, err := r.getClient(ctx, resp.Diagnostics, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	if data.UUID.IsNull() {
		msg := "UUID is null"
		tflog.Error(ctx, msg)
		return
	}

	err = interfaces.DeleteStorageVolume(ctx, resp.Diagnostics, *client, data.UUID.ValueString())
	if err != nil {
		msg := fmt.Sprintf("error deleting storage/volumes: %s", err)
		tflog.Error(ctx, msg)
		resp.Diagnostics.AddError("error deleting storage/volumes", msg)
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *StorageVolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// getClient will use existing client r.client or create one if it's not set
func (r *StorageVolumeResource) getClient(ctx context.Context, diags diag.Diagnostics, cxProfileName types.String) (*restclient.RestClient, error) {
	if r.client == nil {
		client, err := r.config.NewClient(ctx, diags, cxProfileName.ValueString(), r.name)
		if err != nil {
			return nil, err
		}
		r.client = client
	}
	return r.client, nil
}
