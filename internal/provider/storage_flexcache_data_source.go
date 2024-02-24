package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageFlexcacheDataSource{}

// NewStorageFlexcacheDataSource is a helper function to simplify the provider implementation.
func NewStorageFlexcacheDataSource() datasource.DataSource {
	return &StorageFlexcacheDataSource{
		config: resourceOrDataSourceConfig{
			name: "storage_flexcache_data_source",
		},
	}
}

// StorageFlexcacheDataSource implements the datasource interface and defines the data model for the resource.
type StorageFlexcacheDataSource struct {
	config resourceOrDataSourceConfig
}

// StorageFlexcacheDataSourceModel describes the resource data model.
type StorageFlexcacheDataSourceModel struct {
	CxProfileName            types.String `tfsdk:"cx_profile_name"`
	Name                     types.String `tfsdk:"name"`
	SvmName                  types.String `tfsdk:"svm_name"`
	Origins                  types.Set    `tfsdk:"origins"`
	JunctionPath             types.String `tfsdk:"junction_path"`
	Size                     types.Int64  `tfsdk:"size"`
	SizeUnit                 types.String `tfsdk:"size_unit"`
	ConstituentsPerAggregate types.Int64  `tfsdk:"constituents_per_aggregate"`
	DrCache                  types.Bool   `tfsdk:"dr_cache"`
	Guarantee                types.Object `tfsdk:"guarantee"`
	GlobalFileLockingEnabled types.Bool   `tfsdk:"global_file_locking_enabled"`
	UseTieredAggregate       types.Bool   `tfsdk:"use_tiered_aggregate"`
	Aggregates               types.Set    `tfsdk:"aggregates"`
	ID                       types.String `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *StorageFlexcacheDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *StorageFlexcacheDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Flexcache resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the flexcache volume",
				Required:            true,
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the svm to use",
				Required:            true,
			},
			"aggregates": schema.SetNestedAttribute{
				MarkdownDescription: "",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of the aggregate",
							Computed:            true,
						},
						"id": schema.StringAttribute{
							MarkdownDescription: "UUID of the aggregate",
							Computed:            true,
						},
					},
				},
			},
			"origins": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"volume": schema.SingleNestedAttribute{
							MarkdownDescription: "Origin volume",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Name of the origin volume",
									Computed:            true,
								},
								"id": schema.StringAttribute{
									MarkdownDescription: "ID of the origin volume",
									Computed:            true,
								},
							},
						},
						"svm": schema.SingleNestedAttribute{
							MarkdownDescription: "Origin volume SVM",
							Required:            true,
							Attributes: map[string]schema.Attribute{
								"name": schema.StringAttribute{
									MarkdownDescription: "Name of the origin volume",
									Computed:            true,
								},
								"id": schema.StringAttribute{
									MarkdownDescription: "ID of the origin volume",
									Computed:            true,
								},
							},
						},
					},
				},
				MarkdownDescription: "Set of the origin volumes",
				Computed:            true,
			},
			"junction_path": schema.StringAttribute{
				MarkdownDescription: "Name of the junction path",
				Computed:            true,
			},
			"size": schema.Int64Attribute{
				MarkdownDescription: "The size of the flexcache volume",
				Computed:            true,
			},
			"size_unit": schema.StringAttribute{
				MarkdownDescription: "The unit used to interpret the size parameter",
				Computed:            true,
			},
			"constituents_per_aggregate": schema.Int64Attribute{
				MarkdownDescription: "The number of constituents per aggregate",
				Computed:            true,
			},
			"dr_cache": schema.BoolAttribute{
				MarkdownDescription: "The state of the dr cache",
				Computed:            true,
			},
			"guarantee": schema.SingleNestedAttribute{
				MarkdownDescription: "The guarantee of the volume",
				Computed:            true,
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "The type of guarantee",
						Computed:            true,
					},
				},
			},
			"global_file_locking_enabled": schema.BoolAttribute{
				MarkdownDescription: "The state of the global file locking",
				Computed:            true,
			},
			"use_tiered_aggregate": schema.BoolAttribute{
				MarkdownDescription: "The state of the use tiered aggregates",
				Computed:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The UUID of the flexcache volume",
				Computed:            true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *StorageFlexcacheDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

// Read refreshes the Terraform state with the latest data.
func (r *StorageFlexcacheDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageFlexcacheDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

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
	flexcache, err := interfaces.GetStorageFlexcacheByName(errorHandler, *client, data.Name.ValueString(), data.SvmName.ValueString())
	if err != nil {
		return
	}
	if flexcache == nil {
		errorHandler.MakeAndReportError("No flexcache found", fmt.Sprintf("Flexcache %s not found.", data.Name))
		return
	}

	size, sizeUnit := interfaces.ByteFormat(int64(flexcache.Size))
	data.Size = types.Int64Value(int64(size))
	data.SizeUnit = types.StringValue(sizeUnit)
	data.JunctionPath = types.StringValue(flexcache.JunctionPath)
	data.ConstituentsPerAggregate = types.Int64Value(int64(flexcache.ConstituentsPerAggregate))
	data.DrCache = types.BoolValue(flexcache.DrCache)
	data.GlobalFileLockingEnabled = types.BoolValue(flexcache.GlobalFileLockingEnabled)
	data.UseTieredAggregate = types.BoolValue(flexcache.UseTieredAggregate)

	elementTypes := map[string]attr.Type{
		"type": types.StringType,
	}
	elements := map[string]attr.Value{
		"type": types.StringValue(flexcache.Guarantee.Type),
	}
	objectValue, diags := types.ObjectValue(elementTypes, elements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Guarantee = objectValue

	//Origins
	setElements := []attr.Value{}
	for _, origin := range flexcache.Origins {
		nestedElementTypes := map[string]attr.Type{
			"name": types.StringType,
			"id":   types.StringType,
		}
		nestedVolumeElements := map[string]attr.Value{
			"name": types.StringValue(origin.Volume.Name),
			"id":   types.StringValue(origin.Volume.ID),
		}
		nestedSVMElements := map[string]attr.Value{
			"name": types.StringValue(origin.SVM.Name),
			"id":   types.StringValue(origin.SVM.ID),
		}
		originVolumeObjectValue, diags := types.ObjectValue(nestedElementTypes, nestedVolumeElements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		originSVMObjectValue, _ := types.ObjectValue(nestedElementTypes, nestedSVMElements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}

		elementTypes := map[string]attr.Type{
			"volume": types.ObjectType{AttrTypes: nestedElementTypes},
			"svm":    types.ObjectType{AttrTypes: nestedElementTypes},
		}
		elements := map[string]attr.Value{
			"volume": originVolumeObjectValue,
			"svm":    originSVMObjectValue,
		}
		objectValue, diags := types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		setElements = append(setElements, objectValue)
	}

	setValue, diags := types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"volume": types.ObjectType{AttrTypes: map[string]attr.Type{
				"name": types.StringType,
				"id":   types.StringType,
			}},
			"svm": types.ObjectType{AttrTypes: map[string]attr.Type{
				"name": types.StringType,
				"id":   types.StringType,
			}},
		},
	}, setElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	data.Origins = setValue

	//aggregate
	setElements = []attr.Value{}
	for _, aggregate := range flexcache.Aggregates {
		nestedElementTypes := map[string]attr.Type{
			"name": types.StringType,
			"id":   types.StringType,
		}
		nestedElements := map[string]attr.Value{
			"name": types.StringValue(aggregate.Name),
			"id":   types.StringValue(aggregate.ID),
		}
		objectValue, diags := types.ObjectValue(nestedElementTypes, nestedElements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		setElements = append(setElements, objectValue)
	}
	setValue, diags = types.SetValue(types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name": types.StringType,
			"id":   types.StringType,
		},
	}, setElements)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}
	data.Aggregates = setValue
	data.ID = types.StringValue(flexcache.UUID)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
