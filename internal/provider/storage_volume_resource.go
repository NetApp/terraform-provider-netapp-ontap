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
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &StorageVolumeResource{}
var _ resource.ResourceWithImportState = &StorageVolumeResource{}

// NewStorageVolumeResource is a helper function to simplify the provider implementation.
func NewStorageVolumeResource() resource.Resource {
	return &StorageVolumeResource{}
}

// StorageVolumeResource defines the resource implementation.
type StorageVolumeResource struct {
	client *restclient.RestClient
}

// StorageVolumeResourceModel describes the resource data model.
type StorageVolumeResourceModel struct {
	Name       types.String   `tfsdk:"name"`
	Vserver    types.String   `tfsdk:"vserver"`
	Aggregates []types.String `tfsdk:"aggregates"`
	Size       types.Number   `tfsdk:"size"`
	ID         types.String   `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *StorageVolumeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume"
}

// GetSchema defines the schema for the resource.
func (r *StorageVolumeResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Volume resource",

		Attributes: map[string]tfsdk.Attribute{
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
			"size": {
				MarkdownDescription: "Physical Size of the volume in bytes",
				Optional:            true,
				Type:                types.NumberType,
			},
			"id": {
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
	// if req.ProviderData == nil {
	// 	return
	// }

	// TODO: get credentials from connection_profiles, using req.ProviderData
	client, err := restclient.NewClient(ctx, restclient.ConnectionProfile{Hostname: "10.193.78.222", Username: "admin", Password: "netapp1!"})
	if err != nil {
		msg := fmt.Sprintf("error creating REST client: %s", err)
		tflog.Error(ctx, msg)
		resp.Diagnostics.AddError("unable to create REST client", msg)
		return
	}

	// client, ok := req.ProviderData.(*http.Client)

	// if !ok {
	// 	resp.Diagnostics.AddError(
	// 		"Unexpected Resource Configure Type",
	// 		fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
	// 	)

	// 	return
	// }

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *StorageVolumeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageVolumeResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create volume, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.ID = types.StringValue("id")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
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

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read volume, got error: %s", err))
	//     return
	// }

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

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update volume, got error: %s", err))
	//     return
	// }

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

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := d.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete volume, got error: %s", err))
	//     return
	// }
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *StorageVolumeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
