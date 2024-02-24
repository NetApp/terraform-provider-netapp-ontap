package provider

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &StorageFlexcachesDataSource{}

// NewStorageFlexcachesDataSource is a helper function to simplify the provider implementation.
func NewStorageFlexcachesDataSource() datasource.DataSource {
	return &StorageFlexcachesDataSource{
		config: resourceOrDataSourceConfig{
			name: "storage_flexcaches_data_source",
		},
	}
}

// StorageFlexcachesDataSource defines the resource implementation.
type StorageFlexcachesDataSource struct {
	config resourceOrDataSourceConfig
}

// StorageFlexcachesDataSourceModel describes the resource data model.
type StorageFlexcachesDataSourceModel struct {
	CxProfileName     types.String                           `tfsdk:"cx_profile_name"`
	StorageFlexcaches []StorageFlexcacheDataSourceModel      `tfsdk:"storage_flexcaches"`
	Filter            *StorageFlexcacheDataSourceFilterModel `tfsdk:"filter"`
}

// StorageFlexcacheDataSourceModel describes the data source data model for queries.
type StorageFlexcacheDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm_name"`
}

// Metadata returns the resource type name.
func (r *StorageFlexcachesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.name
}

// Schema defines the schema for the resource.
func (r *StorageFlexcachesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Flexcache resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "StorageFlexcache name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "StorageFlexcache svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"storage_flexcaches": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the flexcache volume to manage",
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
										MarkdownDescription: "ID of the aggregate",
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
												MarkdownDescription: "Name of the origin volume SVM",
												Computed:            true,
											},
											"id": schema.StringAttribute{
												MarkdownDescription: "ID of the origin volume SVM",
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
									MarkdownDescription: "The type of the guarantee",
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
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *StorageFlexcachesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (r *StorageFlexcachesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data StorageFlexcachesDataSourceModel

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
	var filter *interfaces.StorageFlexcacheDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.StorageFlexcacheDataSourceFilterModel{
			Name:    data.Filter.Name.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}

	restInfo, err := interfaces.GetStorageFlexcaches(errorHandler, *client, filter)
	if err != nil {
		return
	}
	data.StorageFlexcaches = make([]StorageFlexcacheDataSourceModel, len(restInfo))
	for index, record := range restInfo {

		vsize, vunits := interfaces.ByteFormat(int64(record.Size))

		data.StorageFlexcaches[index] = StorageFlexcacheDataSourceModel{}
		data.StorageFlexcaches[index].CxProfileName = data.CxProfileName
		data.StorageFlexcaches[index].Name = types.StringValue(record.Name)
		data.StorageFlexcaches[index].SvmName = types.StringValue(record.SVM.Name)
		data.StorageFlexcaches[index].Size = types.Int64Value(int64(vsize))
		data.StorageFlexcaches[index].SizeUnit = types.StringValue(vunits)
		data.StorageFlexcaches[index].JunctionPath = types.StringValue(record.JunctionPath)
		data.StorageFlexcaches[index].ConstituentsPerAggregate = types.Int64Value(int64(record.ConstituentsPerAggregate))
		data.StorageFlexcaches[index].DrCache = types.BoolValue(record.DrCache)
		data.StorageFlexcaches[index].GlobalFileLockingEnabled = types.BoolValue(record.GlobalFileLockingEnabled)
		data.StorageFlexcaches[index].UseTieredAggregate = types.BoolValue(record.UseTieredAggregate)
		data.StorageFlexcaches[index].ID = types.StringValue(record.UUID)

		//guarantee
		elementTypes := map[string]attr.Type{
			"type": types.StringType,
		}
		elements := map[string]attr.Value{
			"type": types.StringValue(record.Guarantee.Type),
		}
		objectValue, diags := types.ObjectValue(elementTypes, elements)
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		data.StorageFlexcaches[index].Guarantee = objectValue

		//origin
		setElements := []attr.Value{}
		for _, origin := range record.Origins {
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
			log.Printf("objectValue is: %#v", objectValue)
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
		data.StorageFlexcaches[index].Origins = setValue

		//aggregate
		setElements = []attr.Value{}
		log.Printf("flexcache.Aggregates is: %#v", record.Aggregates)
		for _, aggregate := range record.Aggregates {
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
		data.StorageFlexcaches[index].Aggregates = setValue
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
