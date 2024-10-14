package protocols

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/svm"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsSanIgroupResource{}
var _ resource.ResourceWithImportState = &ProtocolsSanIgroupResource{}

// NewProtocolsSanIgroupResource is a helper function to simplify the provider implementation.
func NewProtocolsSanIgroupResource() resource.Resource {
	return &ProtocolsSanIgroupResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "san_igroup",
		},
	}
}

// NewProtocolsSanIgroupResourceAlias is a helper function to simplify the provider implementation.
func NewProtocolsSanIgroupResourceAlias() resource.Resource {
	return &ProtocolsSanIgroupResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_san_igroup_resource",
		},
	}
}

// ProtocolsSanIgroupResource defines the resource implementation.
type ProtocolsSanIgroupResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsSanIgroupResourceModel describes the resource data model.
type ProtocolsSanIgroupResourceModel struct {
	CxProfileName types.String                               `tfsdk:"cx_profile_name"`
	Name          types.String                               `tfsdk:"name"`
	SVM           svm.SVM                                    `tfsdk:"svm"`
	Comment       types.String                               `tfsdk:"comment"`
	Igroups       []ProtocolsSanIgroupResourceIgroupModel    `tfsdk:"igroups"`
	Initiators    []ProtocolsSanIgroupResourceInitiatorModel `tfsdk:"initiators"`
	OsType        types.String                               `tfsdk:"os_type"`
	Portset       types.Object                               `tfsdk:"portset"`
	Protocol      types.String                               `tfsdk:"protocol"`
	ID            types.String                               `tfsdk:"id"`
}

// ProtocolsSanIgroupResourceIgroupModel describes the data source data model.
type ProtocolsSanIgroupResourceIgroupModel struct {
	Name types.String `tfsdk:"name"`
}

// ProtocolsSanIgroupResourceInitiatorModel describes the data source data model.
type ProtocolsSanIgroupResourceInitiatorModel struct {
	Name types.String `tfsdk:"name"`
}

// ProtocolsSanIgroupResourceLunMapModel describes the data source data model.
type ProtocolsSanIgroupResourceLunMapModel struct {
	LogicalUnitNumber types.Int64 `tfsdk:"logical_unit_number"`
	Lun               Lun         `tfsdk:"lun"`
}

// ProtocolsSanIgroupResourcePortsetModel describes the data source data model.
type ProtocolsSanIgroupResourcePortsetModel struct {
	Name types.String `tfsdk:"name"`
}

// Metadata returns the resource type name.
func (r *ProtocolsSanIgroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsSanIgroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsSanIgroup resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Existing SVM in which to create the initiator group.",
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
			"comment": schema.StringAttribute{
				MarkdownDescription: "Comment",
				Optional:            true,
			},
			"igroups": schema.SetNestedAttribute{
				MarkdownDescription: "List of initiator groups",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Initiator group name",
							Required:            true,
						},
					},
				},
			},
			"initiators": schema.SetNestedAttribute{
				MarkdownDescription: "List of initiators",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Initiator name",
							Required:            true,
						},
					},
				},
			},
			"os_type": schema.StringAttribute{
				MarkdownDescription: "Operating system of the initiator group's initiators.\n",
				Required:            true,
			},
			"portset": schema.SingleNestedAttribute{
				MarkdownDescription: "Required ONTAP 9.9 or greater. The portset to which the initiator group is bound. Binding the initiator group to a portset restricts the initiators of the group to accessing mapped LUNs only through network interfaces in the portset.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Object{objectplanmodifier.UseStateForUnknown()},
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Portset name",
						Required:            true,
					},
				},
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "If not specified, the default protocol is mixed.",
				Default:             stringdefault.StaticString("mixed"),
				Computed:            true,
				Optional:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "Igroup UUID",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsSanIgroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *ProtocolsSanIgroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsSanIgroupResourceModel

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
	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}

	restInfo, err := interfaces.GetProtocolsSanIgroupByName(errorHandler, *client, data.Name.ValueString(), data.SVM.Name.ValueString(), cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsSanIgroupByName
		return

	}

	data.Name = types.StringValue(restInfo.Name)
	data.SVM.Name = types.StringValue(restInfo.SVM.Name)
	data.Comment = types.StringValue(restInfo.Comment)
	data.OsType = types.StringValue(restInfo.OsType)
	data.Protocol = types.StringValue(restInfo.Protocol)
	data.ID = types.StringValue(restInfo.UUID)
	elementTypes := map[string]attr.Type{
		"name": types.StringType,
	}
	elements := map[string]attr.Value{
		"name": types.StringValue(restInfo.Portset.Name),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Portset = objectValue

	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *ProtocolsSanIgroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProtocolsSanIgroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ProtocolsSanIgroupResourceBodyDataModelONTAP
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVM.Name.ValueString()
	if !data.Comment.IsUnknown() {
		body.Comment = data.Comment.ValueString()
	}

	if data.Igroups != nil {
		igroups := []interfaces.IgroupLun{}
		for _, igroup := range data.Igroups {
			var ig interfaces.IgroupLun
			ig.Name = igroup.Name.ValueString()
			igroups = append(igroups, ig)
		}
		err := mapstructure.Decode(igroups, &body.Igroups)
		if err != nil {
			errorHandler.MakeAndReportError("error creating igroups", fmt.Sprintf("error on encoding copies info: %s, copies %#v", err, igroups))
			return
		}
	}

	if data.Initiators != nil {
		initiators := []interfaces.IgroupInitiator{}
		for _, v := range data.Igroups {
			var initiator interfaces.IgroupInitiator
			initiator.Name = v.Name.ValueString()
			initiators = append(initiators, initiator)
		}
		err := mapstructure.Decode(initiators, &body.Initiators)
		if err != nil {
			errorHandler.MakeAndReportError("error creating igroups", fmt.Sprintf("error on encoding copies info: %s, copies %#v", err, initiators))
			return
		}
	}

	body.OsType = data.OsType.ValueString()
	if !data.Portset.IsUnknown() {
		var portset ProtocolsSanIgroupResourcePortsetModel
		diags := data.Portset.As(ctx, &portset, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		if !portset.Name.IsUnknown() {
			body.Portset.Name = portset.Name.ValueString()
		}
	}
	body.Protocol = data.Protocol.ValueString()

	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	resource, err := interfaces.CreateProtocolsSanIgroup(errorHandler, *client, body)
	if err != nil {
		return
	}

	data.ID = types.StringValue(resource.UUID)
	elementTypes := map[string]attr.Type{
		"name": types.StringType,
	}
	elements := map[string]attr.Value{
		"name": types.StringValue(resource.Portset.Name),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
	}
	data.Portset = objectValue

	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsSanIgroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data, state *ProtocolsSanIgroupResourceModel

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

	var request interfaces.UpdateProtocolsSanIgroupResourceBodyDataModelONTAP
	if !data.Comment.Equal(state.Comment) {
		request.Comment = data.Comment.ValueString()
	}
	if !data.OsType.Equal(state.OsType) {
		request.OsType = data.OsType.ValueString()
	}

	tflog.Debug(ctx, fmt.Sprintf("update an igroup resource: %#v", data))
	err = interfaces.UpdateProtocolsSanIgroup(errorHandler, *client, request, state.ID.ValueString())
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
func (r *ProtocolsSanIgroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsSanIgroupResourceModel

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
		errorHandler.MakeAndReportError("UUID is null", "protocols_san_igroup UUID is null")
		return
	}

	err = interfaces.DeleteProtocolsSanIgroup(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsSanIgroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm").AtName("name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
}
