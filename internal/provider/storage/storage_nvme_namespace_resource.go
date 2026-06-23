package storage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &StorageNvmeNamespaceResource{}

// StorageNvmeNamespaceResource defines the resource implementation.
type StorageNvmeNamespaceResource struct {
	config connection.ResourceOrDataSourceConfig
}

// StorageNvmeNamespaceResourceSpaceModel describes the space nested data model.
type StorageNvmeNamespaceResourceSpaceModel struct {
	Size      types.Int64 `tfsdk:"size"`
	SizeUnit  types.String `tfsdk:"size_unit"`
	BlockSize types.Int64 `tfsdk:"block_size"`
}

// StorageNvmeNamespaceResourceModel describes the resource data model.
type StorageNvmeNamespaceResourceModel struct {
	CxProfileName types.String                            `tfsdk:"cx_profile_name"`
	Name          types.String                            `tfsdk:"name"`
	SVMName       types.String                            `tfsdk:"svm_name"`
	OSType        types.String                            `tfsdk:"ostype"`
	Space         *StorageNvmeNamespaceResourceSpaceModel `tfsdk:"space"`
	ID            types.String                            `tfsdk:"id"`
}

func getNVMeSpaceSizeBytes(size types.Int64, sizeUnit types.String) (int64, bool) {
	if size.IsNull() || size.IsUnknown() {
		return 0, false
	}

	baseUnit := int64(1)
	if !sizeUnit.IsNull() && !sizeUnit.IsUnknown() {
		unit, ok := interfaces.POW2BYTEMAP[sizeUnit.ValueString()]
		if !ok {
			return 0, false
		}
		baseUnit = int64(unit)
	}

	return size.ValueInt64() * baseUnit, true
}

func getNVMeNamespaceStateSize(requestedSize types.Int64, requestedSizeUnit types.String, namespaceSize int64, blockSize int64) int64 {
	stateSize := namespaceSize
	if requestedSizeBytes, ok := getNVMeSpaceSizeBytes(requestedSize, requestedSizeUnit); ok && blockSize > 0 {
		roundedSize := requestedSizeBytes
		remainder := roundedSize % blockSize
		if remainder != 0 {
			roundedSize += blockSize - remainder
		}
		if roundedSize == namespaceSize {
			stateSize = requestedSize.ValueInt64()
		}
	}

	return stateSize
}

// NewStorageNvmeNamespaceResource is a helper function to simplify the provider implementation.
func NewStorageNvmeNamespaceResource() resource.Resource {
	return &StorageNvmeNamespaceResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "nvme_namespace",
		},
	}
}

// Metadata returns the resource type name.
func (r *StorageNvmeNamespaceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *StorageNvmeNamespaceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "StorageNvmeNamespace resource",
		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the NVMe namespace.",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "The name of the SVM.",
				Required:            true,
			},
			"ostype": schema.StringAttribute{
				MarkdownDescription: "The operating system type of the NVMe namespace. Choices: aix, linux, vmware, windows.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"space": schema.SingleNestedAttribute{
				Required:            true,
				MarkdownDescription: "The storage space related properties of the NVMe namespace.",
				Attributes: map[string]schema.Attribute{
					"size": schema.Int64Attribute{
						Required:            true,
						MarkdownDescription: "The size of the namespace in bytes.",
					},
					"size_unit": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Unit for size conversion. Supported values: bytes, b, k, kb, m, mb, g, gb, t, tb, p, pb, e, eb, z, zb, y, yb.",
					},
					"block_size": schema.Int64Attribute{
						Optional:            true,
						Computed:            true,
						MarkdownDescription: "The block size of the namespace in bytes.",
						Validators: []validator.Int64{
							int64validator.OneOf(512, 4096),
						},
						PlanModifiers: []planmodifier.Int64{
							int64planmodifier.UseStateForUnknown(),
							int64planmodifier.RequiresReplace(),
						},
					},
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the NVMe namespace.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *StorageNvmeNamespaceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected connection.Config")
		return
	}
	r.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *StorageNvmeNamespaceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *StorageNvmeNamespaceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	namespace, err := interfaces.GetStorageNvmeNamespaceByUUID(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

	data.Name = types.StringValue(namespace.Name)
	data.SVMName = types.StringValue(namespace.SVM.Name)
	data.OSType = types.StringValue(namespace.OSType)
	tflog.Debug(ctx, fmt.Sprintf("Read storage nvme namespace resource: %#v", namespace))
	data.ID = types.StringValue(namespace.UUID)
	if data.Space != nil {
		if data.Space.SizeUnit.IsNull() || data.Space.SizeUnit.IsUnknown() {
			data.Space.Size = types.Int64Value(namespace.Space.Size)
		}
		stateSize := getNVMeNamespaceStateSize(data.Space.Size, data.Space.SizeUnit, namespace.Space.Size, namespace.Space.BlockSize)
		data.Space.Size = types.Int64Value(stateSize)
		data.Space.BlockSize = types.Int64Value(namespace.Space.BlockSize)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create creates the resource and sets the initial Terraform state.
func (r *StorageNvmeNamespaceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *StorageNvmeNamespaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	var body interfaces.StorageNvmeNamespaceResourceBodyDataModelONTAP
	body.Name = data.Name.ValueString()
	body.SVM.Name = data.SVMName.ValueString()
	if !data.OSType.IsNull() && !data.OSType.IsUnknown() {
		body.OsType = data.OSType.ValueString()
	}
	if data.Space != nil {
		if !data.Space.SizeUnit.IsNull() {
			if _, ok := interfaces.POW2BYTEMAP[data.Space.SizeUnit.ValueString()]; !ok {
				errorHandler.MakeAndReportError("error creating nvme namespace", fmt.Sprintf("invalid input for size_unit: %s, required one of: bytes, b, kb, mb, gb, tb, pb, eb, zb, yb", data.Space.SizeUnit.ValueString()))
				return
			}
			body.Space.Size = data.Space.Size.ValueInt64() * int64(interfaces.POW2BYTEMAP[data.Space.SizeUnit.ValueString()])
		} else {
			body.Space.Size = data.Space.Size.ValueInt64()
		}
		if !data.Space.BlockSize.IsNull() && !data.Space.BlockSize.IsUnknown() {
			body.Space.BlockSize = data.Space.BlockSize.ValueInt64()
		}
	}
	namespace, err := interfaces.CreateStorageNvmeNamespace(errorHandler, *client, body)
	if err != nil {
		return
	}

	namespace, err = interfaces.GetStorageNvmeNamespaceByUUID(errorHandler, *client, namespace.UUID)
	if err != nil {
		return
	}
	tflog.Debug(ctx, fmt.Sprintf("my resource: %#v", namespace))
	data.Name = types.StringValue(namespace.Name)
	data.SVMName = types.StringValue(namespace.SVM.Name)
	data.OSType = types.StringValue(namespace.OSType)
	data.ID = types.StringValue(namespace.UUID)
	if data.Space != nil {
		stateSize := getNVMeNamespaceStateSize(data.Space.Size, data.Space.SizeUnit, namespace.Space.Size, namespace.Space.BlockSize)
		data.Space.Size = types.Int64Value(stateSize)
		data.Space.BlockSize = types.Int64Value(namespace.Space.BlockSize)
	}

	tflog.Trace(ctx, "created a storage nvme namespace resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *StorageNvmeNamespaceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *StorageNvmeNamespaceResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, plan.CxProfileName)
	if err != nil {
		return
	}

	var body interfaces.StorageNvmeNamespaceResourceBodyDataModelONTAP
	if !plan.Name.Equal(state.Name) {
		body.Name = plan.Name.ValueString()
	}
	if !plan.SVMName.Equal(state.SVMName) {
		body.SVM.Name = plan.SVMName.ValueString()
	}
	if plan.Space != nil {
		planBytes, planOK := getNVMeSpaceSizeBytes(plan.Space.Size, plan.Space.SizeUnit)
		stateBytes, stateOK := int64(0), false
		if state.Space != nil {
			stateBytes, stateOK = getNVMeSpaceSizeBytes(state.Space.Size, state.Space.SizeUnit)
		}

		if !planOK {
			errorHandler.MakeAndReportError("error updating nvme namespace size", fmt.Sprintf("invalid input for size_unit: %s, required one of: bytes, b, kb, mb, gb, tb, pb, eb, zb, yb", plan.Space.SizeUnit.ValueString()))
			return
		}

		if !stateOK || planBytes != stateBytes {
			body.Space.Size = planBytes
		}
	}
	err = interfaces.PatchStorageNvmeNamespace(errorHandler, *client, state.ID.ValueString(), body)
	if err != nil {
		return
	}

	namespace, err := interfaces.GetStorageNvmeNamespaceByUUID(errorHandler, *client, state.ID.ValueString())
	if err != nil {
		return
	}

	plan.Name = types.StringValue(namespace.Name)
	plan.SVMName = types.StringValue(namespace.SVM.Name)
	plan.OSType = types.StringValue(namespace.OSType)
	plan.ID = types.StringValue(namespace.UUID)
	if plan.Space != nil {
		stateSize := getNVMeNamespaceStateSize(plan.Space.Size, plan.Space.SizeUnit, namespace.Space.Size, namespace.Space.BlockSize)
		plan.Space.Size = types.Int64Value(stateSize)
		plan.Space.BlockSize = types.Int64Value(namespace.Space.BlockSize)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *StorageNvmeNamespaceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *StorageNvmeNamespaceResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		return
	}

	err = interfaces.DeleteStorageNvmeNamespace(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}
}
