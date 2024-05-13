package protocols

import (
	"context"
	"fmt"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ datasource.DataSource = &CifsLocalGroupsDataSource{}

// NewCifsLocalGroupsDataSource is a helper function to simplify the provider implementation.
func NewCifsLocalGroupsDataSource() datasource.DataSource {
	return &CifsLocalGroupsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "protocols_cifs_local_groups_data_source",
		},
	}
}

// CifsLocalGroupsDataSource defines the data source implementation.
type CifsLocalGroupsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// CifsLocalGroupDataSourceFilterModel describes the data source data model for queries.
type CifsLocalGroupDataSourceFilterModel struct {
	Name    types.String `tfsdk:"name"`
	SVMName types.String `tfsdk:"svm_name"`
}

// CifsLocalGroupsDataSourceModel describes the data source data model.
type CifsLocalGroupsDataSourceModel struct {
	CxProfileName   types.String                         `tfsdk:"cx_profile_name"`
	CifsLocalGroups []CifsLocalGroupDataSourceModel      `tfsdk:"protocols_cifs_local_groups"`
	Filter          *CifsLocalGroupDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *CifsLocalGroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *CifsLocalGroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cifs Local Groups data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"name": schema.StringAttribute{
						MarkdownDescription: "Cifs Local Group name",
						Optional:            true,
					},
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "Cifs Local Group svm name",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"protocols_cifs_local_groups": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Cifs Local Group name",
							Required:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "Cifs Local Group svm name",
							Required:            true,
						},
						"description": schema.StringAttribute{
							MarkdownDescription: "Cifs Local Group description",
							Computed:            true,
						},
						"members": schema.ListNestedAttribute{
							MarkdownDescription: "Cifs Local Group members",
							Computed:            true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										MarkdownDescription: "Cifs Local Group member names",
										Computed:            true,
									},
								},
							},
						},
						"id": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Cifs Local Group identifier",
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "Protocols Cifs Local Groups",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *CifsLocalGroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
	}
	d.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (d *CifsLocalGroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CifsLocalGroupsDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, d.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	var filter *interfaces.CifsLocalGroupDataSourceFilterModel = nil
	if data.Filter != nil {
		filter = &interfaces.CifsLocalGroupDataSourceFilterModel{
			Name:    data.Filter.Name.ValueString(),
			SVMName: data.Filter.SVMName.ValueString(),
		}
	}
	restInfo, err := interfaces.GetCifsLocalGroups(errorHandler, *client, filter)
	if err != nil {
		// error reporting done inside GetCifsLocalGroups
		return
	}

	data.CifsLocalGroups = make([]CifsLocalGroupDataSourceModel, len(restInfo))
	for index, record := range restInfo {

		var members = make([]Member, len(record.Members))
		for i, v := range record.Members {
			members[i].Name = types.StringValue(v.Name)
		}
		data.CifsLocalGroups[index] = CifsLocalGroupDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
			SVMName:       types.StringValue(record.SVM.Name),
			Description:   types.StringValue(record.Description),
			ID:            types.StringValue(record.SID),
			Members:       members,
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read data sources: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
