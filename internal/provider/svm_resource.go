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
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
)

var _ resource.Resource = &SvmResource{}
var _ resource.ResourceWithImportState = &SvmResource{}

// NewSvmResource is a helper function to simplify the provider implementation.
func NewSvmResource() resource.Resource {
	return &SvmResource{name: "svm_resource"}
}

// SvmResource defines the resource implementation.
type SvmResource struct {
	client *restclient.RestClient
	config Config
	name   string
}

// SvmResourceModel describes the resource data model.
type SvmResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	UUID          types.String `tfsdk:"uuid"`
}

// Metadata returns the resource type name.
func (r *SvmResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.name
}

// GetSchema defines the schema for the resource.
func (r *SvmResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Svm resource",

		Attributes: map[string]tfsdk.Attribute{
			"cx_profile_name": {
				MarkdownDescription: "Connection profile name",
				Type:                types.StringType,
				Required:            true,
			},
			"name": {
				MarkdownDescription: "The name of the svm to manage",
				Required:            true,
				Type:                types.StringType,
			},
			"uuid": {
				Computed:            true,
				MarkdownDescription: "Vserver identifier",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
		},
	}, nil
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
	r.config = config
	// we need to defer setting the client until we can read the connection profile name
	r.client = nil
}

// Create the resource and sets the initial Terraform state.
func (r *SvmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SvmResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var request interfaces.SvmResourceModel
	request.Name = data.Name.ValueString()
	client, err := r.getClient(ctx, resp.Diagnostics, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	svm, err := interfaces.CreateSvm(ctx, resp.Diagnostics, *client, request)
	if err != nil {
		msg := fmt.Sprintf("error creating svm/svms: %s", err)
		tflog.Error(ctx, msg)
		resp.Diagnostics.AddError("error creating svm", msg)
		return
	}
	data.UUID = types.StringValue(svm.UUID)
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

	if data.UUID.IsNull() {
		msg := "UUID is null"
		tflog.Error(ctx, msg)
		return
	}

	client, err := r.getClient(ctx, resp.Diagnostics, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	_, err = interfaces.GetSvm(ctx, resp.Diagnostics, *client, data.UUID.ValueString())
	if err != nil {
		msg := fmt.Sprintf("error reading svm/svms: %s", err)
		tflog.Error(ctx, msg)
		resp.Diagnostics.AddError("error getting svm", msg)
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SvmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SvmResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SvmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SvmResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	if data.UUID.IsNull() {
		msg := "UUID is null"
		tflog.Error(ctx, msg)
		return
	}
	// TODO: Uncomment when the DeleteSvm is ready
	client, err := r.getClient(ctx, resp.Diagnostics, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}
	err = interfaces.DeleteSvm(ctx, resp.Diagnostics, *client, data.UUID.ValueString())
	if err != nil {
		msg := fmt.Sprintf("error deleting svm/svms: %s", err)
		tflog.Error(ctx, msg)
		resp.Diagnostics.AddError("error deleting svm", msg)
		return
	}
}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SvmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// getClient will use existing client r.client or create one if it's not set
func (r *SvmResource) getClient(ctx context.Context, diags diag.Diagnostics, cxProfileName types.String) (*restclient.RestClient, error) {
	if r.client == nil {
		client, err := r.config.NewClient(ctx, diags, cxProfileName.ValueString(), r.name)
		if err != nil {
			return nil, err
		}
		r.client = client
	}
	return r.client, nil
}
