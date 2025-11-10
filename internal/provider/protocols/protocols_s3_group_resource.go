package protocols

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &ProtocolsS3GroupResource{}
var _ resource.ResourceWithImportState = &ProtocolsS3GroupResource{}

// NewProtocolsS3GroupResource is a helper function to simplify the provider implementation.
func NewProtocolsS3GroupResource() resource.Resource {
	return &ProtocolsS3GroupResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "s3_group",
		},
	}
}

// ProtocolsS3GroupResource defines the resource implementation.
type ProtocolsS3GroupResource struct {
	config connection.ResourceOrDataSourceConfig
}

// ProtocolsS3GroupResourceModel describes the resource data model.
type ProtocolsS3GroupResourceModel struct {
	CxProfileName types.String `tfsdk:"cx_profile_name"`
	Name          types.String `tfsdk:"name"`
	Comment       types.String `tfsdk:"comment"`
	Users	      types.Set    `tfsdk:"users"`
	Policies      types.Set    `tfsdk:"policies"`
	SVMName       types.String `tfsdk:"svm_name"`
	ID            types.Int64  `tfsdk:"id"`
}

// Metadata returns the resource type name
func (r *ProtocolsS3GroupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *ProtocolsS3GroupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "ProtocolsS3Group resource",
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
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"users": schema.SetAttribute{
				MarkdownDescription: "List of users in the S3 group",
				Required:            true,
				ElementType:         types.StringType,
				Validators: []validator.Set{
					setvalidator.SizeAtLeast(1),
				},
			},
			"policies": schema.SetAttribute{
				MarkdownDescription: "List of policies in the S3 group",
				Optional:            true,
				Computed:            true,
				ElementType:         types.StringType,
				PlanModifiers: []planmodifier.Set{
					setplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *ProtocolsS3GroupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}
	config, ok := req.ProviderData.(connection.Config)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected Config, got: %T. Please resport this issue to the provider developers.", req.ProviderData),
		)
	}
	r.config.ProviderConfig = config
}

// Read refreshes the Terraform state with the latest data.
func (r *ProtocolsS3GroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProtocolsS3GroupResourceModel

	// Read Terraform prior state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	// we need to defer setting the client until we can read the connection profile name
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside New Client
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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	var restInfo *interfaces.ProtocolsS3GroupGetDataModelONTAP
	// var err error
	if data.ID.ValueInt64() != 0 {
		restInfo, err = interfaces.GetProtocolsS3GroupbyID(errorHandler, *client, svm.UUID, data.ID.ValueInt64(), cluster.Version)
	} else {
		restInfo, err = interfaces.GetProtocolsS3Group(errorHandler, *client, data.Name.ValueString(), svm.UUID, cluster.Version)
	}
	if err != nil {
		// error reporting done inside GetProtocolsS3GroupbyID, GetProtocolsS3Group
		return
	}

	data.ID = types.Int64Value(restInfo.ID)
	data.Name = types.StringValue(restInfo.Name)
	data.Comment = types.StringValue(restInfo.Comment)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	
	// Users - convert to types.Set
	var userElements []attr.Value
	for _, user := range restInfo.Users {
		userElements = append(userElements, types.StringValue(user.Name))
	}
	usersSet, diags := types.SetValue(types.StringType, userElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Users = usersSet
	
	// Policies - convert to types.Set
	if len(restInfo.Policies) == 0 || restInfo.Policies == nil {
		// Set to empty set when API returns no policies
		emptySet, diags := types.SetValue(types.StringType, []attr.Value{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Policies = emptySet
	} else {
		var policyElements []attr.Value
		for _, policy := range restInfo.Policies {
			policyElements = append(policyElements, types.StringValue(policy.Name))
		}
		policiesSet, diags := types.SetValue(types.StringType, policyElements)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Policies = policiesSet
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *ProtocolsS3GroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ProtocolsS3GroupResourceModel

	// Read Terraform plan data into the model.
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	var body interfaces.ProtocolsS3GroupResourceBodyDataModel
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	body.Name = data.Name.ValueString()
	if !data.Comment.IsNull() {
		commentValue := data.Comment.ValueString()
		body.Comment = &commentValue
	}
	// users
	users := []interfaces.UserGetDataModel{}
	userItems := make([]types.String, 0, len(data.Users.Elements()))
	diags := data.Users.ElementsAs(ctx, &userItems, false)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	for _, userName := range userItems {
		var aUser interfaces.UserGetDataModel
		aUser.Name = userName.ValueString()
		users = append(users, aUser)
	}
	err := mapstructure.Decode(users, &body.Users)
	if err != nil {
		errorHandler.MakeAndReportError("error creating S3 group", fmt.Sprintf("error on encoding users info: %s, users %#v", err, users))
		return
	}

	// policies
	if data.Policies.IsNull() || data.Policies.IsUnknown() {
		body.Policies = nil
	} else if len(data.Policies.Elements()) > 0 {
		policies := []interfaces.PolicyGetDataModel{}
		policyElements := make([]types.String, 0, len(data.Policies.Elements()))
		diags := data.Policies.ElementsAs(ctx, &policyElements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, policyName := range policyElements {
			var aPolicy interfaces.PolicyGetDataModel
			aPolicy.Name = policyName.ValueString()
			policies = append(policies, aPolicy)
		}
		var mappedPolicies []map[string]interface{}
		err := mapstructure.Decode(policies, &mappedPolicies)
		if err != nil {
			errorHandler.MakeAndReportError("error creating S3 group", fmt.Sprintf("error on encoding policies info: %s, policies %#v", err, policies))
			return
		}
		body.Policies = &mappedPolicies
	} else {
		// Explicitly set empty policies
		emptyPolicies := []map[string]interface{}{}
		body.Policies = &emptyPolicies
	}

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
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	resource, err := interfaces.CreateProtocolS3Group(errorHandler, *client, svm.UUID, body)
	if err != nil {
		return
	}

	data.ID = types.Int64Value(resource.ID)
	restInfo, err := interfaces.GetProtocolsS3GroupbyID(errorHandler, *client, svm.UUID, data.ID.ValueInt64(), cluster.Version)
	if err != nil {
		// error reporting done inside GetProtocolsS3GroupbyID
		return
	}
	
	// Update the computed parameters
	data.Name = types.StringValue(restInfo.Name)
	data.Comment = types.StringValue(restInfo.Comment)
	data.SVMName = types.StringValue(restInfo.SVM.Name)
	
	// Users - convert to types.Set
	var userElements []attr.Value
	for _, user := range restInfo.Users {
		userElements = append(userElements, types.StringValue(user.Name))
	}
	usersSet, diags := types.SetValue(types.StringType, userElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Users = usersSet
	
	// Policies - convert to types.Set
	if len(restInfo.Policies) == 0 || restInfo.Policies == nil {
		// Set to empty set when API returns no policies  
		emptySet, diags := types.SetValue(types.StringType, []attr.Value{})
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Policies = emptySet
	} else {
		var policyElements []attr.Value
		for _, policy := range restInfo.Policies {
			policyElements = append(policyElements, types.StringValue(policy.Name))
		}
		policiesSet, diags := types.SetValue(types.StringType, policyElements)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		data.Policies = policiesSet
	}

	tflog.Trace(ctx, fmt.Sprintf("created S3 group resource, ID=%s", data.ID))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *ProtocolsS3GroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *ProtocolsS3GroupResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	// Read Terraform state data in to the model
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := connection.GetRestClient(errorHandler, r.config, state.CxProfileName)
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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, state.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	// Update the resource
	var body interfaces.ProtocolsS3GroupResourceBodyDataModel

	body.Name = plan.Name.ValueString()
	if !plan.Comment.Equal(state.Comment) {
		commentValue := plan.Comment.ValueString()
		body.Comment = &commentValue
	}
	//users update
	if plan.Users.IsNull() {
		body.Users = nil
	} else {
		users := []interfaces.UserGetDataModel{}
		userElements := make([]types.String, 0, len(plan.Users.Elements()))
		diags := plan.Users.ElementsAs(ctx, &userElements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, userName := range userElements {
			var aUser interfaces.UserGetDataModel
			aUser.Name = userName.ValueString()
			users = append(users, aUser)
		}
		err := mapstructure.Decode(users, &body.Users)
		if err != nil {
			errorHandler.MakeAndReportError("error updating S3 group", fmt.Sprintf("error on encoding users info: %s, users %#v", err, users))
			return
		}
	}

	//policies update
	if !plan.Policies.IsNull() && !plan.Policies.IsUnknown() && len(plan.Policies.Elements()) > 0 {
		policies := []interfaces.PolicyGetDataModel{}
		policyElements := make([]types.String, 0, len(plan.Policies.Elements()))
		diags := plan.Policies.ElementsAs(ctx, &policyElements, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		for _, policyName := range policyElements {
			var aPolicy interfaces.PolicyGetDataModel
			aPolicy.Name = policyName.ValueString()
			policies = append(policies, aPolicy)
		}
		var mappedPolicies []map[string]interface{}
		err := mapstructure.Decode(policies, &mappedPolicies)
		if err != nil {
			errorHandler.MakeAndReportError("error updating S3 group", fmt.Sprintf("error on encoding policies info: %s, policies %#v", err, policies))
			return
		}
		body.Policies = &mappedPolicies
	} else if !plan.Policies.IsNull() && !plan.Policies.IsUnknown() {
		// Explicitly set empty policies when user specifies empty array
		emptyPolicies := []map[string]interface{}{}
		body.Policies = &emptyPolicies
	}

	err = interfaces.UpdateProtocolsS3Group(errorHandler, *client, body, plan.ID.ValueInt64(), svm.UUID)
	if err != nil {
		return
	}

	restInfo, err := interfaces.GetProtocolsS3GroupbyID(errorHandler, *client, svm.UUID, plan.ID.ValueInt64(), cluster.Version)
	if err != nil {
		return
	}
	
	// Update the computed parameters
	plan.Name = types.StringValue(restInfo.Name)
	plan.Comment = types.StringValue(restInfo.Comment)
	plan.SVMName = types.StringValue(restInfo.SVM.Name)
	
	// users
	// convert to types.Set
	var userElements []attr.Value
	for _, user := range restInfo.Users {
		userElements = append(userElements, types.StringValue(user.Name))
	}
	usersSet, diags := types.SetValue(types.StringType, userElements)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.Users = usersSet
	
	// policies
	// only update if they were explicitly specified in the plan
	if !plan.Policies.IsNull() {
		if len(restInfo.Policies) == 0 || restInfo.Policies == nil {
			emptySet, diags := types.SetValue(types.StringType, []attr.Value{})
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			plan.Policies = emptySet
		} else {
			var policyElements []attr.Value
			for _, policy := range restInfo.Policies {
				policyElements = append(policyElements, types.StringValue(policy.Name))
			}
			policiesSet, diags := types.SetValue(types.StringType, policyElements)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}
			plan.Policies = policiesSet
		}
	}

	tflog.Debug(ctx, fmt.Sprintf("updated S3 group resource: ID=%s", plan.ID))

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (r *ProtocolsS3GroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *ProtocolsS3GroupResourceModel

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

	// Get SVM info
	svm, err := interfaces.GetSvmByName(errorHandler, *client, data.SVMName.ValueString())
	if err != nil {
		// error reporting done inside GetSvmByName
		errorHandler.MakeAndReportError("No SVM found", "SVM not found")
		return
	}

	if data.ID.IsNull() {
		errorHandler.MakeAndReportError("ID is null", "S3 group ID is null")
		return
	}

	err = interfaces.DeleteProtocolsS3Group(errorHandler, *client, data.ID.ValueInt64(), svm.UUID)
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *ProtocolsS3GroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ",")

	if len(idParts) != 3 || idParts[0] == "" || idParts[1] == "" || idParts[2] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: name,svm_name,cx_profile_name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("svm_name"), idParts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
}
