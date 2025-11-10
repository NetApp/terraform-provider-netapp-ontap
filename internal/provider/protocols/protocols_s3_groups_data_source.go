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
var _ datasource.DataSource = &ProtocolsS3GroupsDataSource{}

// NewProtocolsS3GroupsDataSource is a helper function to simplify the provider implementation.
func NewProtocolsS3GroupsDataSource() datasource.DataSource {
	return &ProtocolsS3GroupsDataSource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "s3_groups",
		},
	}
}

// ProtocolsS3GroupsDataSource defines the data source implementation.
type ProtocolsS3GroupsDataSource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsS3GroupsDataSourceModel describes the data source data model.
type ProtocolsS3GroupsDataSourceModel struct {
	CxProfileName types.String                           `tfsdk:"cx_profile_name"`
	S3Groups      []ProtocolsS3GroupDataSourceModel      `tfsdk:"s3_groups"`
	Filter        *ProtocolsS3GroupDataSourceFilterModel `tfsdk:"filter"`
}

// Metadata returns the data source type name.
func (d *ProtocolsS3GroupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + d.config.Name
}

// Schema defines the schema for the data source.
func (d *ProtocolsS3GroupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsS3Groups data source",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				Attributes: map[string]schema.Attribute{
					"svm_name": schema.StringAttribute{
						MarkdownDescription: "The name of the SVM",
						Optional:            true,
					},
				},
				Optional: true,
			},
			"s3_groups": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"cx_profile_name": schema.StringAttribute{
							MarkdownDescription: "Connection profile name",
							Required:            true,
						},
						"svm_name": schema.StringAttribute{
							MarkdownDescription: "The name of the SVM",
							Required:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the S3 group",
							Required:            true,
						},
						"comment": schema.StringAttribute{
							MarkdownDescription: "Additional information about the group",
							Computed:            true,
						},
						"users": schema.SetAttribute{
							Computed:            true,
							MarkdownDescription: "The list of users who belong to the group",
							ElementType:         types.StringType,
						},
						"policies": schema.SetAttribute{
							Computed:            true,
							MarkdownDescription: "The list of policies that are attached to the group",
							ElementType:         types.StringType,
						},
						"id": schema.Int64Attribute{
							MarkdownDescription: "S3 Group id",
							Computed:            true,
						},
					},
				},
				Computed:            true,
				MarkdownDescription: "",
			},
		},
	}
}

// Configure adds the provider configured client to the data source.
func (d *ProtocolsS3GroupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
func (d *ProtocolsS3GroupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProtocolsS3GroupsDataSourceModel

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

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	if data.Filter == nil {
		errorHandler.MakeAndReportError("No SVM specified", "either svm_name or filter must be specified")
		return
	}

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.Filter.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	restInfo, err := interfaces.GetProtocolsS3Groups(errorHandler, *client, svm.UUID, cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsS3Groups
		return
	}

	data.S3Groups = make([]ProtocolsS3GroupDataSourceModel, len(restInfo))
	for index, record := range restInfo {
		// users - map to set
		var users = make([]string, len(record.Users))
		for i, user := range record.Users {
			users[i] = user.Name
		}
		usersSet, diags := types.SetValueFrom(ctx, types.StringType, users)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		// policies - map to set
		var policies = make([]string, len(record.Policies))
		for i, policy := range record.Policies {
			policies[i] = policy.Name
		}
		policiesSet, diags := types.SetValueFrom(ctx, types.StringType, policies)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		data.S3Groups[index] = ProtocolsS3GroupDataSourceModel{
			CxProfileName: types.String(data.CxProfileName),
			Name:          types.StringValue(record.Name),
			Comment:       types.StringValue(record.Comment),
			SVMName:       types.StringValue(record.SVM.Name),
			Users:         usersSet,
			Policies:      policiesSet,
			ID:            types.Int64Value(record.ID),
		}
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a data source: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
