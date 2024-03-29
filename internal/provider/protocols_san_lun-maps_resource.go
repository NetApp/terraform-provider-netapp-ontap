package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsSanLunMapsResource{}
var _ resource.ResourceWithImportState = &ProtocolsSanLunMapsResource{}

// NewProtocolsSanLunMapsResource is a helper function to simplify the provider implementation.
func NewProtocolsSanLunMapsResource() resource.Resource {
	return &ProtocolsSanLunMapsResource{
		config: resourceOrDataSourceConfig{
			name: "protocols_san_lun-maps_resource",
		},
	}
}

// ProtocolsSanLunMapsResource defines the resource implementation.
type ProtocolsSanLunMapsResource struct {
	config resourceOrDataSourceConfig
}

// ProtocolsSanLunMapsResourceModel describes the resource data model.
type ProtocolsSanLunMapsResourceModel struct {
	CxProfileName     types.String `tfsdk:"cx_profile_name"`
	SVM               SVM          `tfsdk:"svm"`
	Lun               Lun          `tfsdk:"lun"`
	IGroup            IGroup       `tfsdk:"igroup"`
	LogicalUnitNumber types.Int64  `tfsdk:"logical_unit_number"`
	ID                types.String `tfsdk:"id"`
}

// Lun describes Lun data model.
type Lun struct {
	Name types.String `tfsdk:"name"`
}

// IGroup describes IGroup data model.
type IGroup struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the resource type name.
func (r *ProtocolsSanLunMapsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *ProtocolsSanLunMapsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsSanLunMaps resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"svm": schema.SingleNestedAttribute{
				MarkdownDescription: "SVM details for ProtocolsSanLunMaps",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "name of the SVM",
						Required:            true,
					},
				},
			},
			"igroup": schema.SingleNestedAttribute{
				MarkdownDescription: "SVM details for ProtocolsSanLunMaps",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "name of the igroup",
						Required:            true,
					},
				},
			},
			"lun": schema.SingleNestedAttribute{
				MarkdownDescription: "SVM details for ProtocolsSanLunMaps",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "name of the lun",
						Required:            true,
					},
				},
			},
			"logical_unit_number": schema.Int64Attribute{
				MarkdownDescription: "If no value is provided, ONTAP assigns the lowest available value",
				Optional:            true,
				Computed:            true,
				// Default:             int64default.StaticInt64(0),
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "ProtocolsSanLunMaps igroup and lun UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsSanLunMapsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ProtocolsSanLunMapsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsSanLunMapsResourceModel

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

	restInfo, err := interfaces.GetProtocolsSanLunMapsByName(errorHandler, *client, data.IGroup.Name.ValueString(), data.Lun.Name.ValueString(), data.SVM.Name.ValueString())
	if err != nil {
		// error reporting done inside GetProtocolsSanLunMaps
		return
	}

	id := restInfo.IGroup.UUID + "," + restInfo.Lun.UUID
	data.ID = types.StringValue(id)

	data.LogicalUnitNumber = types.Int64Value(int64(restInfo.LogicalUnitNumber))

	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *ProtocolsSanLunMapsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProtocolsSanLunMapsResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ProtocolsSanLunMapsResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.IGroup.Name = data.IGroup.Name.ValueString()
	body.SVM.Name = data.SVM.Name.ValueString()
	body.Lun.Name = data.Lun.Name.ValueString()
	if !data.LogicalUnitNumber.IsUnknown() {
		body.LogicalUnitNumber = int(data.LogicalUnitNumber.ValueInt64())
	}

	client, err := getRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateProtocolsSanLunMaps(errorHandler, *client, body)
	if err != nil {
		return
	}

	id := resource.IGroup.UUID + "," + resource.Lun.UUID
	data.ID = types.StringValue(id)
	data.LogicalUnitNumber = types.Int64Value(int64(resource.LogicalUnitNumber))

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsSanLunMapsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *ProtocolsSanLunMapsResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsSanLunMapsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsSanLunMapsResourceModel

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
		errorHandler.MakeAndReportError("UUID is null", "protocols_san_lun-maps igroup or lun UUID is null")
		return
	}

	idParts := strings.Split(data.ID.ValueString(), ",")

	err = interfaces.DeleteProtocolsSanLunMaps(errorHandler, *client, idParts[0], idParts[1])
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsSanLunMapsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 4 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" || idParts[3] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: svm_name,igroup_name,lun_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm").AtName("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("igroup").AtName("name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("lun").AtName("name"), idParts[2])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[3])...)
}
