package provider

import (
	"context"
	"fmt"

	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/cluster"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/name_services"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/networking"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/protocols"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/security"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/snapmirror"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/storage"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/svm"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure ONTAPProvider satisfies various provider interfaces.
var _ provider.Provider = &ONTAPProvider{}

//var _ provider.ProviderWithMetadata = &ONTAPProvider{}

// ONTAPProvider defines the provider implementation.
type ONTAPProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// ConnectionProfileModel associate a connection profile with a name
// TODO: augment address with hostname, ...
type ConnectionProfileModel struct {
	Name                  types.String `tfsdk:"name"`
	Hostname              types.String `tfsdk:"hostname"`
	Username              types.String `tfsdk:"username"`
	Password              types.String `tfsdk:"password"`
	ValidateCerts         types.Bool   `tfsdk:"validate_certs"`
	ONTAPProviderAWSModel types.Object `tfsdk:"aws_lambda"`
}

// ONTAPProviderModel describes the provider data model.
type ONTAPProviderModel struct {
	Endpoint             types.String `tfsdk:"endpoint"`
	JobCompletionTimeOut types.Int64  `tfsdk:"job_completion_timeout"`
	ConnectionProfiles   types.List   `tfsdk:"connection_profiles"`
}

type ONTAPProviderAWSLambdaModel struct {
	Region              types.String `tfsdk:"region"`
	SharedConfigProfile types.String `tfsdk:"shared_config_profile"`
	FunctionName        types.String `tfsdk:"function_name"`
}

// Metadata defines the provider type name for inclusion in each data source and resource type name
func (p *ONTAPProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "netapp-ontap"
	resp.Version = p.version

}

// Schema defines the schema for provider-level configuration.
func (p *ONTAPProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
			},
			"job_completion_timeout": schema.Int64Attribute{
				MarkdownDescription: "Time in seconds to wait for completion. Default to 600 seconds",
				Optional:            true,
			},
			"connection_profiles": schema.ListNestedAttribute{
				MarkdownDescription: "Define connection and credentials",
				Required:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "Profile name",
							Required:            true,
						},
						"hostname": schema.StringAttribute{
							MarkdownDescription: "ONTAP management interface IP address or name. For AWS Lambda, the management endpoints for the FSxN system.",
							Required:            true,
						},
						"username": schema.StringAttribute{
							MarkdownDescription: "ONTAP management user name (cluster or svm)",
							Required:            true,
						},
						"password": schema.StringAttribute{
							MarkdownDescription: "ONTAP management password for username",
							Required:            true,
							Sensitive:           true,
						},
						"validate_certs": schema.BoolAttribute{
							MarkdownDescription: "Whether to enforce SSL certificate validation, defaults to true. Not applicable for AWS Lambda",
							Optional:            true,
						},
						"aws_lambda": schema.SingleNestedAttribute{
							MarkdownDescription: "AWS configuration for Lambda",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"region": schema.StringAttribute{
									MarkdownDescription: "AWS region.",
									Optional:            true,
								},
								"function_name": schema.StringAttribute{
									MarkdownDescription: "AWS Lambda function name",
									Optional:            true,
								},
								"shared_config_profile": schema.StringAttribute{
									MarkdownDescription: "AWS shared config profile. Region set in the profile will be ignored it it's different from the region set in Terraform. aws_access_key_id and aws_secret_access_key are required to be set in credentials",
									Required:            true,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure shared clients for data source and resource implementations.
func (p *ONTAPProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data ONTAPProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		tflog.Error(ctx, fmt.Sprintf("unable to read data from req: %#v", req))
		return
	}
	// Required attributes
	// For optional values we can use data.Endpoint.IsNull(), ...

	if data.ConnectionProfiles.IsUnknown() {
		resp.Diagnostics.AddError("no connection profiles", "At least one connection profile must be defined.")
		return
	}
	if len(data.ConnectionProfiles.Elements()) == 0 {
		resp.Diagnostics.AddError("no connection profile", "At least one connection profile must be defined.")
		return
	}
	connectionProfilesElements := make([]types.Object, 0, len(data.ConnectionProfiles.Elements()))
	diags := data.ConnectionProfiles.ElementsAs(ctx, &connectionProfilesElements, false)
	if diags.HasError() {
		resp.Diagnostics.Append(diags...)
		return
	}

	connectionProfiles := make(map[string]connection.Profile, len(data.ConnectionProfiles.Elements()))

	for _, profile := range connectionProfilesElements {
		var connectionProfile ConnectionProfileModel
		diags := profile.As(ctx, &connectionProfile, basetypes.ObjectAsOptions{})
		if diags.HasError() {
			resp.Diagnostics.Append(diags...)
			return
		}
		var validateCerts bool
		if connectionProfile.ValidateCerts.IsNull() {
			validateCerts = true
		} else {
			validateCerts = connectionProfile.ValidateCerts.ValueBool()
		}
		connectionProfiles[connectionProfile.Name.ValueString()] = connection.Profile{
			Hostname:              connectionProfile.Hostname.ValueString(),
			Username:              connectionProfile.Username.ValueString(),
			Password:              connectionProfile.Password.ValueString(),
			ValidateCerts:         validateCerts,
			MaxConcurrentRequests: 0,
		}
		if !connectionProfile.ONTAPProviderAWSModel.IsNull() {
			var lambdaConfig ONTAPProviderAWSLambdaModel
			diags := connectionProfile.ONTAPProviderAWSModel.As(ctx, &lambdaConfig, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			currentProfile := connectionProfiles[connectionProfile.Name.ValueString()]
			currentProfile.UseAWSLambda = true
			currentProfile.AWS = connection.AWSConfig{
				Region:              lambdaConfig.Region.ValueString(),
				SharedConfigProfile: lambdaConfig.SharedConfigProfile.ValueString(),
				FunctionName:        lambdaConfig.FunctionName.ValueString(),
			}
			connectionProfiles[connectionProfile.Name.ValueString()] = currentProfile

		}
	}
	jobCompletionTimeOut := data.JobCompletionTimeOut.ValueInt64()
	if data.JobCompletionTimeOut.IsNull() {
		jobCompletionTimeOut = 600
	}
	config := connection.Config{
		ConnectionProfiles:   connectionProfiles,
		JobCompletionTimeOut: int(jobCompletionTimeOut),
		Version:              p.version,
	}
	resp.DataSourceData = config
	resp.ResourceData = config

}

// Resources defines the provider's resources.
func (p *ONTAPProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		storage.NewAggregateResource,
		protocols.NewCifsLocalGroupResource,
		protocols.NewCifsLocalGroupMemberResource,
		protocols.NewCifsLocalUserResource,
		protocols.NewCifsServiceResource,
		protocols.NewCifsUserGroupPrivilegeResource,
		cluster.NewClusterResource,
		cluster.NewClusterLicensingLicenseResource,
		cluster.NewClusterPeersResource,
		cluster.NewClusterScheduleResource,
		NewExampleResource,
		protocols.NewExportPolicyResource,
		protocols.NewExportPolicyRuleResource,
		networking.NewIPInterfaceResource,
		networking.NewIPRouteResource,
		name_services.NewNameServicesDNSResource,
		name_services.NewNameServicesLDAPResource,
		protocols.NewProtocolsCIFSShareResource,
		protocols.NewProtocolsNfsServiceResource,
		protocols.NewProtocolsSanIgroupResource,
		protocols.NewProtocolsSanLunMapsResource,
		security.NewSecurityAccountResource,
		security.NewSecurityRolesResource,
		security.NewSecurityLoginMessageResource,
		snapmirror.NewSnapmirrorResource,
		snapmirror.NewSnapmirrorPolicyResource,
		storage.NewQOSPoliciesResource,
		storage.NewStorageLunResource,
		storage.NewSnapshotPolicyResource,
		storage.NewStorageQuotaRulesResource,
		storage.NewStorageFlexcacheRsource,
		storage.NewStorageVolumeResource,
		storage.NewVolumesFilesResource,
		storage.NewStorageQtreeResource,
		storage.NewStorageVolumeEfficiencyPoliciesResource,
		storage.NewStorageVolumeSnapshotResource,
		svm.NewSVMPeersResource,
		svm.NewSvmResource,
	}
}

// DataSources defines the provider's data sources.
func (p *ONTAPProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		protocols.NewCifsLocalGroupDataSource,
		protocols.NewCifsLocalGroupsDataSource,
		protocols.NewCifsLocalGroupMemberDataSource,
		protocols.NewCifsLocalGroupMembersDataSource,
		protocols.NewCifsLocalUserDataSource,
		protocols.NewCifsLocalUsersDataSource,
		protocols.NewCifsServiceDataSource,
		protocols.NewCifsServicesDataSource,
		protocols.NewCifsUserGroupPrivilegeDataSource,
		protocols.NewCifsUserGroupPrivilegesDataSource,
		cluster.NewClusterDataSource,
		cluster.NewClusterLicensingLicenseDataSource,
		cluster.NewClusterLicensingLicensesDataSource,
		cluster.NewClusterPeerDataSource,
		cluster.NewClusterPeersDataSource,
		cluster.NewClusterScheduleDataSource,
		cluster.NewClusterSchedulesDataSource,
		NewExampleDataSource,
		protocols.NewExportPolicyDataSource,
		protocols.NewExportPoliciesDataSource,
		protocols.NewExportPolicyRuleDataSource,
		protocols.NewExportPolicyRulesDataSource,
		networking.NewIPInterfaceDataSource,
		networking.NewIPInterfacesDataSource,
		networking.NewIPRouteDataSource,
		networking.NewIPRoutesDataSource,
		name_services.NewNameServicesDNSDataSource,
		name_services.NewNameServicesDNSsDataSource,
		name_services.NewNameServicesLDAPDataSource,
		name_services.NewNameServicesLDAPsDataSource,
		protocols.NewProtocolsCIFSShareDataSource,
		protocols.NewProtocolsCIFSSharesDataSource,
		protocols.NewProtocolsNfsServiceDataSource,
		protocols.NewProtocolsNfsServicesDataSource,
		protocols.NewProtocolsSanIgroupDataSource,
		protocols.NewProtocolsSanIgroupsDataSource,
		protocols.NewProtocolsSanLunMapDataSource,
		protocols.NewProtocolsSanLunMapsDataSource,
		security.NewSecurityAccountDataSource,
		security.NewSecurityAccountsDataSource,
		security.NewSecurityCertificateDataSource,
		security.NewSecurityCertificatesDataSource,
		security.NewSecurityRoleDataSource,
		security.NewSecurityRolesDataSource,
		security.NewSecurityLoginMessageDataSource,
		security.NewSecurityLoginMessagesDataSource,
		snapmirror.NewSnapmirrorDataSource,
		snapmirror.NewSnapmirrorsDataSource,
		storage.NewSnapshotPoliciesDataSource,
		storage.NewSnapshotPolicyDataSource,
		snapmirror.NewSnapmirrorPolicyDataSource,
		snapmirror.NewSnapmirrorPoliciesDataSource,
		storage.NewStorageAggregateDataSource,
		storage.NewStorageAggregatesDataSource,
		storage.NewStorageFlexcacheDataSource,
		storage.NewStorageFlexcachesDataSource,
		storage.NewStorageLunDataSource,
		storage.NewStorageLunsDataSource,
		storage.NewStorageQOSPolicyDataSource,
		storage.NewStorageQOSPoliciesDataSource,
		storage.NewStorageQuotaRuleDataSource,
		storage.NewStorageQuotaRulesDataSource,
		storage.NewStorageQtreeDataSource,
		storage.NewStorageQtreesDataSource,
		storage.NewStorageVolumeSnapshotDataSource,
		storage.NewStorageVolumeSnapshotsDataSource,
		storage.NewVolumeEfficiencyPolicyDataSource,
		storage.NewVolumeEfficiencyPoliciesDataSource,
		storage.NewStorageVolumeDataSource,
		storage.NewStorageVolumesDataSource,
		storage.NewStorageVolumesFilesDataSource,
		svm.NewSvmDataSource,
		svm.NewSvmsDataSource,
		svm.NewSVMPeerDataSource,
		svm.NewSVMPeersDataSource,
	}
}

// New creates a provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &ONTAPProvider{
			version: version,
		}
	}
}
