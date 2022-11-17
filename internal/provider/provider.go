package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure ONTAPProvider satisfies various provider interfaces.
var _ provider.Provider = &ONTAPProvider{}
var _ provider.ProviderWithMetadata = &ONTAPProvider{}

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
	Name          types.String `tfsdk:"name"`
	Hostname      types.String `tfsdk:"hostname"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	ValidateCerts types.Bool   `tfsdk:"validate_certs"`
}

// ONTAPProviderModel describes the provider data model.
type ONTAPProviderModel struct {
	Endpoint           types.String             `tfsdk:"endpoint"`
	ConnectionProfiles []ConnectionProfileModel `tfsdk:"connection_profiles"`
}

// Metadata defines the provider type name for inclusion in each data source and resource type name
func (p *ONTAPProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "netapp-ontap"
	resp.Version = p.version
}

// GetSchema defines the schema for provider-level configuration.
func (p *ONTAPProvider) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"endpoint": {
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
				Type:                types.StringType,
			},
			"connection_profiles": {
				MarkdownDescription: "Define connection and credentials",
				Required:            true,
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"name": {
						MarkdownDescription: "Profile name",
						Required:            true,
						Type:                types.StringType,
					},
					"hostname": {
						MarkdownDescription: "ONTAP management interface IP address or name",
						Required:            true,
						Type:                types.StringType,
					},
					"username": {
						MarkdownDescription: "ONTAP management user name (cluster or vserver)",
						Required:            true,
						Type:                types.StringType,
					},
					"password": {
						MarkdownDescription: "ONTAP management password for username",
						Required:            true,
						Type:                types.StringType,
						Sensitive:           true,
					},
					"validate_certs": {
						MarkdownDescription: "Whether to enforce SSL certificate validation, defaults to true",
						Type:                types.BoolType,
						Attributes:          nil,
						Optional:            true,
					},
				}),
			},
		},
	}, nil
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
	if len(data.ConnectionProfiles) == 0 {
		resp.Diagnostics.AddError("no connection profile", "At least one connection profile must be defined.")
		return
	}
	connectionProfiles := make(map[string]ConnectionProfile, len(data.ConnectionProfiles))
	for _, profile := range data.ConnectionProfiles {
		var validateCerts bool
		if profile.ValidateCerts.IsNull() {
			validateCerts = true
		} else {
			validateCerts = profile.ValidateCerts.ValueBool()
		}
		connectionProfiles[profile.Name.ValueString()] = ConnectionProfile{
			Hostname:              profile.Hostname.ValueString(),
			Username:              profile.Username.ValueString(),
			Password:              profile.Password.ValueString(),
			ValidateCerts:         validateCerts,
			MaxConcurrentRequests: 0,
		}
	}

	config := Config{
		ConnectionProfiles: connectionProfiles,
	}
	resp.DataSourceData = config
	resp.ResourceData = config
}

// Resources defines the provider's resources.
func (p *ONTAPProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewExampleResource,
	}
}

// DataSources defines the provider's data sources.
func (p *ONTAPProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewClusterDataSource,
		NewExampleDataSource,
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
