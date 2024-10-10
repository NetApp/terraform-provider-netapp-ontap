package security

import (
	"context"
	"fmt"
	"strings"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/interfaces"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &SecurityCertificateResource{}
var _ resource.ResourceWithImportState = &SecurityCertificateResource{}

// NewSecurityCertificateResource is a helper function to simplify the provider implementation.
func NewSecurityCertificateResource() resource.Resource {
	return &SecurityCertificateResource{
		config: connection.ResourceOrDataSourceConfig{
			Name: "security_certificate",
		},
	}
}

// SecurityCertificateResource defines the resource implementation.
type SecurityCertificateResource struct {
	config connection.ResourceOrDataSourceConfig
}

// SecurityCertificateResourceModel describes the resource data model.
type SecurityCertificateResourceModel struct {
	CxProfileName      types.String `tfsdk:"cx_profile_name"`
	Name               types.String `tfsdk:"name"`
	CommonName         types.String `tfsdk:"common_name"`
	Type               types.String `tfsdk:"type"`
	SVMName            types.String `tfsdk:"svm_name"`
	Scope              types.String `tfsdk:"scope"`
	SerialNumber       types.String `tfsdk:"serial_number"`
	CA                 types.String `tfsdk:"ca"`
	PublicCertificate  types.String `tfsdk:"public_certificate"`
	SignedCertificate  types.String `tfsdk:"signed_certificate"`
	PrivateKey         types.String `tfsdk:"private_key"`
	SigningRequest     types.String `tfsdk:"signing_request"`
	HashFunction       types.String `tfsdk:"hash_function"`
	KeySize            types.Int64  `tfsdk:"key_size"`
	ExpiryTime         types.String `tfsdk:"expiry_time"`
	ID                 types.String `tfsdk:"id"`
}

// Metadata returns the resource type name.
func (r *SecurityCertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + r.config.Name
}

// Schema defines the schema for the resource.
func (r *SecurityCertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "SecurityCertificate resource",

		Attributes: map[string]schema.Attribute{
			"cx_profile_name": schema.StringAttribute{
				MarkdownDescription: "Connection profile name.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The unique name of the security certificate per SVM.",
				Optional:            true,
				Computed:            true,
			},
			"common_name": schema.StringAttribute{
				MarkdownDescription: "Common name of the certificate.",
				Required:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "Type of certificate.",
				Required:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("client", "server", "client_ca", "server_ca", "root_ca"),
				},
			},
			"svm_name": schema.StringAttribute{
				MarkdownDescription: "Name of the SVM in which the certificate is created or installed or the SVM on which the signed certificate will exist.",
				Optional:            true,
				Computed:            true,
			},
			"scope": schema.StringAttribute{
				MarkdownDescription: "Set to 'svm' for certificates installed in a SVM. Otherwise, set to 'cluster'.",
				Optional:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"serial_number": schema.StringAttribute{
				MarkdownDescription: "Serial number of the certificate.",
				Optional:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ca": schema.StringAttribute{
				MarkdownDescription: "Certificate authority.",
				Optional:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"public_certificate": schema.StringAttribute{
				MarkdownDescription: "Public key Certificate in PEM format. If this is not provided during create action, a self-signed certificate is created.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"signed_certificate": schema.StringAttribute{
				MarkdownDescription: "Signed public key Certificate in PEM format that is returned while signing a certificate.",
				Optional:            false,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"private_key": schema.StringAttribute{
				MarkdownDescription: "Private key Certificate in PEM format. Only valid when installing a CA-signed certificate.",
				Optional:            true,
				Sensitive:           true,
			},
			"signing_request": schema.StringAttribute{
				MarkdownDescription: "Certificate signing request to be signed by the given certificate authority. Request should be in X509 PEM format.",
				Optional:            true,
			},
			"hash_function": schema.StringAttribute{
				MarkdownDescription: "Hashing function.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.String{
					stringvalidator.OneOf("sha1", "sha256", "md5", "sha224", "sha384", "sha512"),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"key_size": schema.Int64Attribute{
				MarkdownDescription: "Key size of the certificate in bits.",
				Optional:            true,
				Computed:            true,
				Validators: []validator.Int64{
					int64validator.OneOf(512, 1024, 1536, 2048, 3072),
				},
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"expiry_time": schema.StringAttribute{
				MarkdownDescription: "Certificate expiration time, in ISO 8601 duration format or date and time format.",
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "UUID of the certificate.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *SecurityCertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *SecurityCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SecurityCertificateResourceModel

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
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "cluster not found")
		return
	}

	var restInfo *interfaces.SecurityCertificateGetDataModelONTAP
	if data.ID.ValueString() != "" {
		restInfo, err = interfaces.GetSecurityCertificateByUUID(errorHandler, *client, cluster.Version, data.ID.ValueString())
		if err != nil {
			// error reporting done inside GetSecurityCertificateByUUID
			return
		}
	} else {
		restInfo, err = interfaces.GetSecurityCertificate(errorHandler, *client, cluster.Version, data.Name.ValueString(), data.CommonName.ValueString(), data.Type.ValueString())
		if err != nil {
			// error reporting done inside GetSecurityCertificate
			return
		}
	}

	if restInfo == nil {
		errorHandler.MakeAndReportError("error reading info", "No Certificate found")
		return
	}

	// Set the values from the response into the data model
	data.ID = types.StringValue(restInfo.UUID)
	data.Name = types.StringValue(restInfo.Name)
	data.CommonName = types.StringValue(restInfo.CommonName)
	data.Scope = types.StringValue(restInfo.Scope)
	data.Type = types.StringValue(restInfo.Type)
	data.SerialNumber = types.StringValue(restInfo.SerialNumber)
	data.CA = types.StringValue(restInfo.CA)
	data.HashFunction = types.StringValue(restInfo.HashFunction)
	data.KeySize = types.Int64Value(restInfo.KeySize)
	data.PublicCertificate = types.StringValue(restInfo.PublicCertificate)
	if data.ExpiryTime.IsNull() {
		data.ExpiryTime = types.StringValue(restInfo.ExpiryTime)
	}
	if data.SVMName.IsUnknown() {
		data.SVMName = types.StringValue(restInfo.SVM.Name)
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Debug(ctx, fmt.Sprintf("read a resource: %#v", data))

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Create a resource and retrieve UUID
func (r *SecurityCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *SecurityCertificateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
	
	errorHandler := utils.NewErrorHandler(ctx, &resp.Diagnostics)
	client, err := connection.GetRestClient(errorHandler, r.config, data.CxProfileName)
	if err != nil {
		// error reporting done inside NewClient
		return
	}

	cluster, err := interfaces.GetCluster(errorHandler, *client)
	if cluster == nil {
		errorHandler.MakeAndReportError("No cluster found", "Cluster not found.")
		return
	}
	if err != nil {
		// error reporting done inside GetCluster
		return
	}
	
	name_supported := false
	if !data.Name.IsNull() && data.Name.ValueString() != "" {
		if cluster.Version.Generation == 9 && cluster.Version.Major < 8 {
			tflog.Error(ctx, "'name' is supported with ONTAP 9.8 or higher.")
			errorHandler.MakeAndReportError("Unsupported parameter", "'name' is supported with ONTAP 9.8 or higher.")
			return
		} else {
			name_supported = true
		}
	}

	if !data.SigningRequest.IsNull() {
		// this if block is for signing security certificate
		var body interfaces.SecurityCertificateResourceSignBodyDataModelONTAP

		// Read the updated data from the API
		restInfo, err := interfaces.GetSecurityCertificate(errorHandler, *client, cluster.Version, data.Name.ValueString(), data.CommonName.ValueString(), data.Type.ValueString())
		if err != nil {
			// error reporting done inside GetSecurityCertificate
			return
		}

		data.ID = types.StringValue(restInfo.UUID)
		data.Name = types.StringValue(restInfo.Name)
		data.CommonName = types.StringValue(restInfo.CommonName)
		data.Type = types.StringValue(restInfo.Type)
		data.Scope = types.StringValue(restInfo.Scope)
		data.SerialNumber = types.StringValue(restInfo.SerialNumber)
		data.CA = types.StringValue(restInfo.CA)
		data.PublicCertificate = types.StringValue(restInfo.PublicCertificate)
		data.HashFunction = types.StringValue(restInfo.HashFunction)
		data.KeySize = types.Int64Value(restInfo.KeySize)
		if data.SVMName.IsUnknown() {
			data.SVMName = types.StringValue(restInfo.SVM.Name)
		}

		body.SigningRequest = data.SigningRequest.ValueString()
		if !data.HashFunction.IsUnknown() {
			body.HashFunction = data.HashFunction.ValueString()
		}
		if !data.ExpiryTime.IsUnknown() {
			body.ExpiryTime = data.ExpiryTime.ValueString()
		}
		
		resource, err := interfaces.SignSecurityCertificate(errorHandler, *client, restInfo.UUID, body)
		if err != nil {
			// error reporting done inside SignSecurityCertificate
			return
		}

		// Save public_certificate returned while signing certificate into Terraform state
		data.SignedCertificate = types.StringValue(resource.SignedCertificate)
		
		tflog.Trace(ctx, "signed a resource")
	} else {
		// This else block is for creating or installing security certificate
		var body interfaces.SecurityCertificateResourceCreateBodyDataModelONTAP

		if !data.Name.IsNull() &&  data.Name.ValueString() != "" {
			if name_supported {
				body.Name = data.Name.ValueString()
			}
		}
		body.CommonName = data.CommonName.ValueString()
		body.Type = data.Type.ValueString()
		if !data.SVMName.IsUnknown() {
			body.SVM.Name = data.SVMName.ValueString()
		}
		if !data.PublicCertificate.IsUnknown() {
			body.PublicCertificate = data.PublicCertificate.ValueString()
		}
		if !data.PrivateKey.IsUnknown() {
			body.PrivateKey = data.PrivateKey.ValueString()
		}
		if !data.HashFunction.IsUnknown() {
			body.HashFunction = data.HashFunction.ValueString()
		}
		if !data.KeySize.IsUnknown() {
			body.KeySize = data.KeySize.ValueInt64()
		}
		if !data.ExpiryTime.IsUnknown() {
			body.ExpiryTime = data.ExpiryTime.ValueString()
		}

		var operation string
		if !data.PublicCertificate.IsUnknown() || !data.PrivateKey.IsNull() {
			operation = "installing"
		} else {
			operation = "creating"
		}
		resource, err := interfaces.CreateOrInstallSecurityCertificate(errorHandler, *client, body, operation)
		if err != nil {
			// error reporting done inside CreateOrInstallSecurityCertificate
			return
		}
		tflog.Trace(ctx, "created/ installed a resource")
		data.ID = types.StringValue(resource.UUID)

		// Read the updated data from the API
		restInfo, err := interfaces.GetSecurityCertificateByUUID(errorHandler, *client, cluster.Version, resource.UUID)
		if err != nil {
			// error reporting done inside GetSecurityCertificateByUUID
			return
		}

		data.Name = types.StringValue(restInfo.Name)
		data.CommonName = types.StringValue(restInfo.CommonName)
		data.Type = types.StringValue(restInfo.Type)
		data.SVMName = types.StringValue(restInfo.SVM.Name)
		data.Scope = types.StringValue(restInfo.Scope)
		data.PublicCertificate = types.StringValue(restInfo.PublicCertificate)
		data.SerialNumber = types.StringValue(restInfo.SerialNumber)
		data.CA = types.StringValue(restInfo.CA)
		data.HashFunction = types.StringValue(restInfo.HashFunction)
		data.KeySize = types.Int64Value(restInfo.KeySize)
		if data.ExpiryTime.IsUnknown() {
			data.ExpiryTime = types.StringValue(restInfo.ExpiryTime)
		}
		// SignedCertificate would be available only while signing a certificate
		data.SignedCertificate = types.StringValue("NA")

		tflog.Trace(ctx, "read newly created resource")
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *SecurityCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *SecurityCertificateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Add an error diagnostic indicating that Update is not supported
	resp.Diagnostics.AddError(
		"Update Not Supported",
		"The update operation is not supported for the security_certificate resource.",
	)
	tflog.Error(ctx, "Update not supported for resource security_certificate")
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *SecurityCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *SecurityCertificateResourceModel

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

	if data.ID.IsUnknown() {
		errorHandler.MakeAndReportError("UUID is null", "security certificate UUID is null")
		return
	}

	err = interfaces.DeleteSecurityCertificate(errorHandler, *client, data.ID.ValueString())
	if err != nil {
		return
	}

}

// ImportState imports a resource using ID from terraform import command by calling the Read method.
func (r *SecurityCertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	tflog.Debug(ctx, fmt.Sprintf("import req security certificate resource: %#v", req))
	// Parse the ID
	idParts := strings.Split(req.ID, ",")

	// import name, common_name, type and cx_profile
	if len(idParts) == 4 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("common_name"), idParts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), idParts[2])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[3])...)
		return
	}

	// import common_name, type, and cx_profile
	if len(idParts) == 3 {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("common_name"), idParts[0])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), idParts[1])...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cx_profile_name"), idParts[2])...)
		return
	}

	resp.Diagnostics.AddError(
		"Unexpected Import Identifier",
		fmt.Sprintf("Expected import identifier with format: name,common_name,type,cx_profile_name or common_name,type,cx_profile_name. Got: %q", req.ID),
	)
}
