package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/cluster"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/connection"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/name_services"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/networking"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/protocols"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/security"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/snapmirror"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/storage"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/support"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/provider/svm"
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
	Name                   types.String `tfsdk:"name"`
	Hostname               types.String `tfsdk:"hostname"`
	Username               types.String `tfsdk:"username"`
	Password               types.String `tfsdk:"password"`
	ValidateCerts          types.Bool   `tfsdk:"validate_certs"`
	ONTAPProviderAWSModel  types.Object `tfsdk:"aws_lambda"`
	ONTAPProviderGCNVModel types.Object `tfsdk:"gcnv"`
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

type ONTAPProviderGCNVDataModel struct {
	ProjectID             types.String `tfsdk:"project_id"`
	Location              types.String `tfsdk:"location"`
	StoragePool           types.String `tfsdk:"storage_pool"`
	ServiceAccountKeyPath types.String `tfsdk:"service_account_key_path"`
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
							MarkdownDescription: "ONTAP management interface IP address or name. For AWS Lambda, the management endpoints for the FSxN system. Not required when using Google Cloud NetApp Volumes (gcnv).",
							Optional:            true,
						},
						"username": schema.StringAttribute{
							MarkdownDescription: "ONTAP management user name (cluster or svm). Not required when using Google Cloud NetApp Volumes (gcnv).",
							Optional:            true,
						},
						"password": schema.StringAttribute{
							MarkdownDescription: "ONTAP management password for username. Not required when using Google Cloud NetApp Volumes (gcnv).",
							Optional:            true,
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
									Required:            true,
								},
								"shared_config_profile": schema.StringAttribute{
									MarkdownDescription: "AWS shared config profile. Region set in the profile will be ignored it it's different from the region set in Terraform. aws_access_key_id and aws_secret_access_key are required to be set in credentials",
									Required:            true,
								},
							},
						},
						"gcnv": schema.SingleNestedAttribute{
							MarkdownDescription: "Google Cloud NetApp Volumes configuration",
							Optional:            true,
							Attributes: map[string]schema.Attribute{
								"project_id": schema.StringAttribute{
									MarkdownDescription: "Google Cloud project ID",
									Required:            true,
								},
								"location": schema.StringAttribute{
									MarkdownDescription: "Google Cloud location",
									Required:            true,
								},
								"storage_pool": schema.StringAttribute{
									MarkdownDescription: "Storage pool name",
									Required:            true,
								},
								"service_account_key_path": schema.StringAttribute{
									MarkdownDescription: "Path to the Google Cloud service account key file (JSON). Required for authentication.",
									Required:            true,
									Sensitive:           false,
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

		// Validate that hostname, username, password are provided when gcnv is not used
		if connectionProfile.ONTAPProviderGCNVModel.IsNull() {
			if connectionProfile.Hostname.IsNull() || connectionProfile.Hostname.ValueString() == "" {
				resp.Diagnostics.AddError(
					"Missing required attribute",
					fmt.Sprintf("hostname is required for connection profile '%s' when gcnv is not configured", connectionProfile.Name.ValueString()),
				)
				return
			}
			if connectionProfile.Username.IsNull() || connectionProfile.Username.ValueString() == "" {
				resp.Diagnostics.AddError(
					"Missing required attribute",
					fmt.Sprintf("username is required for connection profile '%s' when gcnv is not configured", connectionProfile.Name.ValueString()),
				)
				return
			}
			if connectionProfile.Password.IsNull() || connectionProfile.Password.ValueString() == "" {
				resp.Diagnostics.AddError(
					"Missing required attribute",
					fmt.Sprintf("password is required for connection profile '%s' when gcnv is not configured", connectionProfile.Name.ValueString()),
				)
				return
			}
		}

		var validateCerts bool
		if connectionProfile.ValidateCerts.IsNull() {
			validateCerts = true
		} else {
			validateCerts = connectionProfile.ValidateCerts.ValueBool()
		}

		// Use empty string for hostname, username, password when gcnv is used and they're not provided
		hostname := ""
		username := ""
		password := ""
		if !connectionProfile.Hostname.IsNull() {
			hostname = connectionProfile.Hostname.ValueString()
		}
		if !connectionProfile.Username.IsNull() {
			username = connectionProfile.Username.ValueString()
		}
		if !connectionProfile.Password.IsNull() {
			password = connectionProfile.Password.ValueString()
		}

		connectionProfiles[connectionProfile.Name.ValueString()] = connection.Profile{
			Hostname:              hostname,
			Username:              username,
			Password:              password,
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
		if !connectionProfile.ONTAPProviderGCNVModel.IsNull() {
			var gcnvConfig ONTAPProviderGCNVDataModel
			diags := connectionProfile.ONTAPProviderGCNVModel.As(ctx, &gcnvConfig, basetypes.ObjectAsOptions{})
			if diags.HasError() {
				resp.Diagnostics.Append(diags...)
				return
			}
			currentProfile := connectionProfiles[connectionProfile.Name.ValueString()]
			currentProfile.UseGCNV = true
			currentProfile.GCNV = connection.GCNVConfig{
				ProjectID:             gcnvConfig.ProjectID.ValueString(),
				Location:              gcnvConfig.Location.ValueString(),
				StoragePool:           gcnvConfig.StoragePool.ValueString(),
				ServiceAccountKeyPath: gcnvConfig.ServiceAccountKeyPath.ValueString(),
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
		cluster.NewClusterLicensingLicenseResource,
		cluster.NewClusterPeerResource,
		cluster.NewClusterResource,
		cluster.NewClusterScheduleResource,
		name_services.NewNameServicesDNSResource,
		name_services.NewNameServicesLDAPResource,
		networking.NewBroadcastDomainResource,
		networking.NewEthernetPortResource,
		networking.NewIPInterfaceResource,
		networking.NewIPRouteResource,
		NewExampleResource,
		protocols.NewCifsLocalGroupMembersResource,
		protocols.NewCifsLocalGroupResource,
		protocols.NewCifsLocalUserResource,
		protocols.NewCifsServiceResource,
		protocols.NewCifsUserGroupPrivilegeResource,
		protocols.NewExportPolicyResource,
		protocols.NewExportPolicyRuleResource,
		protocols.NewProtocolsCIFSShareResource,
		protocols.NewProtocolsIscsiServiceResource,
		protocols.NewProtocolsNfsServiceResource,
		protocols.NewProtocolsSanIgroupResource,
		protocols.NewProtocolsSanLunMapResource,
		protocols.NewProtocolsS3PolicyResource,
		protocols.NewProtocolsS3GroupResource,
		security.NewSecurityAccountResource,
		security.NewSecurityCertificateResource,
		security.NewSecurityLoginMessageResource,
		security.NewSecurityRoleResource,
		snapmirror.NewSnapmirrorPolicyResource,
		snapmirror.NewSnapmirrorResource,
		storage.NewAggregateResource,
		storage.NewQOSPolicyResource,
		storage.NewSnapshotPolicyResource,
		storage.NewStorageFlexcacheRsource,
		storage.NewStorageLunResource,
		storage.NewStorageQuotaRuleResource,
		storage.NewStorageQtreeResource,
		storage.NewStorageVolumeEfficiencyPolicyResource,
		storage.NewStorageVolumeResource,
		storage.NewStorageVolumeSnapshotResource,
		storage.NewVolumeFileResource,
		support.NewAutoSupportResource,
		svm.NewSVMPeerResource,
		svm.NewSvmResource,
		svm.NewSvmQosPolicyActivationResource,
		// The following resources are Alias for the version 1 names
		cluster.NewClusterLicensingLicenseResourceAlias,
		cluster.NewClusterPeerResourceAlias,
		cluster.NewClusterResourceAlias,
		cluster.NewClusterScheduleResourceAlias,
		name_services.NewNameServicesDNSResourceAlias,
		name_services.NewNameServicesLDAPResourceAlias,
		networking.NewIPInterfaceResourceAlias,
		networking.NewIPRouteResourceAlias,
		protocols.NewCifsLocalGroupMembersResourceAlias,
		protocols.NewCifsLocalGroupResourceAlias,
		protocols.NewCifsLocalUserResourcAlias,
		protocols.NewCifsServiceResourceAlias,
		protocols.NewProtocolsCIFSShareResourceAlias,
		protocols.NewCifsUserGroupPrivilegeResourceAlias,
		protocols.NewExportPolicyResourceAlias,
		protocols.NewExportPolicyRuleResourceAlias,
		protocols.NewProtocolsNfsServiceResourceAlias,
		protocols.NewProtocolsSanIgroupResourceAlias,
		protocols.NewProtocolsSanLunMapResourceAlias,
		security.NewSecurityAccountResourceAlias,
		snapmirror.NewSnapmirrorPolicyResourceAlias,
		snapmirror.NewSnapmirrorResourceAlias,
		storage.NewAggregateResourceAlias,
		storage.NewStorageFlexcacheRsourceAlias,
		storage.NewStorageLunResourceAlias,
		storage.NewSnapshotPolicyResourceAlias,
		storage.NewStorageVolumeResourceAlias,
		storage.NewStorageVolumeSnapshotResourceAlias,
		svm.NewSVMPeerResourceAlias,
		svm.NewSvmResourceAlias,
	}
}

func (p *ONTAPProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		cluster.NewClusterDataSource,
		cluster.NewClusterLicensingLicenseDataSource,
		cluster.NewClusterLicensingLicensesDataSource,
		cluster.NewClusterPeerDataSource,
		cluster.NewClusterPeersDataSource,
		cluster.NewClusterScheduleDataSource,
		cluster.NewClusterSchedulesDataSource,
		name_services.NewNameServicesDNSDataSource,
		name_services.NewNameServicesDNSsDataSource,
		name_services.NewNameServicesLDAPDataSource,
		name_services.NewNameServicesLDAPsDataSource,
		networking.NewBroadcastDomainDataSource,
		networking.NewBroadcastDomainsDataSource,
		networking.NewEthernetPortDataSource,
		networking.NewEthernetPortsDataSource,
		networking.NewIPInterfaceDataSource,
		networking.NewIPInterfacesDataSource,
		networking.NewIPRouteDataSource,
		networking.NewIPRoutesDataSource,
		NewExampleDataSource,
		protocols.NewCifsLocalGroupDataSource,
		protocols.NewCifsLocalGroupMemberDataSource,
		protocols.NewCifsLocalGroupMembersDataSource,
		protocols.NewCifsLocalGroupsDataSource,
		protocols.NewCifsLocalUserDataSource,
		protocols.NewCifsLocalUsersDataSource,
		protocols.NewCifsServiceDataSource,
		protocols.NewCifsServicesDataSource,
		protocols.NewCifsUserGroupPrivilegeDataSource,
		protocols.NewCifsUserGroupPrivilegesDataSource,
		protocols.NewExportPoliciesDataSource,
		protocols.NewExportPolicyDataSource,
		protocols.NewExportPolicyRuleDataSource,
		protocols.NewExportPolicyRulesDataSource,
		protocols.NewProtocolsCIFSShareDataSource,
		protocols.NewProtocolsCIFSSharesDataSource,
		protocols.NewProtocolsIscsiServiceDataSource,
		protocols.NewProtocolsIscsiServicesDataSource,
		protocols.NewProtocolsNfsServiceDataSource,
		protocols.NewProtocolsNfsServicesDataSource,
		protocols.NewProtocolsSanIgroupDataSource,
		protocols.NewProtocolsSanIgroupsDataSource,
		protocols.NewProtocolsSanLunMapDataSource,
		protocols.NewProtocolsSanLunMapsDataSource,
		protocols.NewProtocolsS3PolicyDataSource,
		protocols.NewProtocolsS3PoliciesDataSource,
		protocols.NewProtocolsS3GroupDataSource,
		protocols.NewProtocolsS3GroupsDataSource,
		security.NewSecurityAccountDataSource,
		security.NewSecurityAccountsDataSource,
		security.NewSecurityCertificateDataSource,
		security.NewSecurityCertificatesDataSource,
		security.NewSecurityLoginMessageDataSource,
		security.NewSecurityLoginMessagesDataSource,
		security.NewSecurityRoleDataSource,
		security.NewSecurityRolesDataSource,
		snapmirror.NewSnapmirrorDataSource,
		snapmirror.NewSnapmirrorPolicyDataSource,
		snapmirror.NewSnapmirrorPoliciesDataSource,
		snapmirror.NewSnapmirrorsDataSource,
		storage.NewSnapshotPoliciesDataSource,
		storage.NewSnapshotPolicyDataSource,
		storage.NewStorageAggregateDataSource,
		storage.NewStorageAggregatesDataSource,
		storage.NewStorageFlexcacheDataSource,
		storage.NewStorageFlexcachesDataSource,
		storage.NewStorageLunDataSource,
		storage.NewStorageLunsDataSource,
		storage.NewStorageQOSPoliciesDataSource,
		storage.NewStorageQOSPolicyDataSource,
		storage.NewStorageQuotaRuleDataSource,
		storage.NewStorageQuotaRulesDataSource,
		storage.NewStorageQtreeDataSource,
		storage.NewStorageQtreesDataSource,
		storage.NewStorageVolumeDataSource,
		storage.NewStorageVolumeSnapshotDataSource,
		storage.NewStorageVolumeSnapshotsDataSource,
		storage.NewStorageVolumesDataSource,
		storage.NewStorageVolumesFilesDataSource,
		storage.NewVolumeEfficiencyPoliciesDataSource,
		storage.NewVolumeEfficiencyPolicyDataSource,
		support.NewAutoSupportDataSource,
		svm.NewSVMPeerDataSource,
		svm.NewSVMPeersDataSource,
		svm.NewSvmDataSource,
		svm.NewSvmsDataSource,
		// The following datasource are Alias for the version 1 names
		cluster.NewClusterDataSourceAlias,
		cluster.NewClusterLicensingLicensesDataSourceAlias,
		cluster.NewClusterLicensingLicenseDataSourceAlias,
		cluster.NewClusterPeerDataSourceAlias,
		cluster.NewClusterPeersDataSourceAlias,
		cluster.NewClusterScheduleDataSourceAlias,
		cluster.NewClusterSchedulesDataSourceAlias,
		name_services.NewNameServicesDNSDataSourceAlias,
		name_services.NewNameServicesDNSsDataSourceAlias,
		name_services.NewNameServicesLDAPDataSourceAlias,
		name_services.NewNameServicesLDAPsDataSourceAlias,
		networking.NewIPInterfaceDataSourceAlias,
		networking.NewIPInterfacesDataSourceAlias,
		networking.NewIPRouteDataSourceAlias,
		networking.NewIPRoutesDataSourceAlias,
		protocols.NewCifsLocalGroupDataSourceAlias,
		protocols.NewCifsLocalGroupMemberDataSourceAlias,
		protocols.NewCifsLocalGroupMembersDataSourceAlias,
		protocols.NewCifsLocalGroupsDataSourceAlias,
		protocols.NewCifsLocalUserDataSourceAlias,
		protocols.NewCifsLocalUsersDataSourceAlias,
		protocols.NewCifsServiceDataSourceAlias,
		protocols.NewCifsServicesDataSourceAlias,
		protocols.NewProtocolsCIFSShareDataSourceAlias,
		protocols.NewProtocolsCIFSSharesDataSourceAlias,
		protocols.NewCifsUserGroupPrivilegeDataSourceAlias,
		protocols.NewCifsUserGroupPrivilegesDataSourceAlias,
		protocols.NewExportPoliciesDataSourceAlias,
		protocols.NewExportPolicyDataSourceAlias,
		protocols.NewExportPolicyRuleDataSourceAlias,
		protocols.NewExportPolicyRulesDataSourceAlias,
		protocols.NewProtocolsNfsServiceDataSourceAlias,
		protocols.NewProtocolsNfsServicesDataSourceAlias,
		protocols.NewProtocolsSanIgroupDataSourceAlias,
		protocols.NewProtocolsSanIgroupsDataSourceAlias,
		protocols.NewProtocolsSanLunMapDataSourceAlias,
		protocols.NewProtocolsSanLunMapsDataSourceAlias,
		security.NewSecurityAccountDataSourceAlias,
		security.NewSecurityAccountsDataSourceAlias,
		snapmirror.NewSnapmirrorDataSourceAlias,
		snapmirror.NewSnapmirrorPoliciesDataSourceAlias,
		snapmirror.NewSnapmirrorPolicyDataSourceAlias,
		snapmirror.NewSnapmirrorsDataSourceAlias,
		storage.NewStorageAggregateDataSourceAlias,
		storage.NewStorageAggregatesDataSourceAlias,
		storage.NewStorageFlexcacheDataSourceAlias,
		storage.NewStorageFlexcachesDataSourceAlias,
		storage.NewStorageLunDataSourceAlias,
		storage.NewStorageLunsDataSourceAlias,
		storage.NewSnapshotPoliciesDataSourceAlias,
		storage.NewSnapshotPolicyDataSourceAlias,
		storage.NewStorageVolumeDataSourceAlias,
		storage.NewStorageVolumeSnapshotDataSourceAlias,
		storage.NewStorageVolumeSnapshotsDataSourceAlias,
		storage.NewStorageVolumesDataSourceAlias,
		svm.NewSvmDataSourceAlias,
		svm.NewSVMPeerDataSourceAlias,
		svm.NewSVMPeersDataSourceAlias,
		svm.NewSvmsDataSourceAlias,
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
