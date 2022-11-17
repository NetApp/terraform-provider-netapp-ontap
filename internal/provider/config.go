package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"golang.org/x/exp/maps"
)

// ConnectionProfile describes how to reach a cluster or vserver
type ConnectionProfile struct {
	// TODO: add certs in addition to basic authentication
	// TODO: Add Timeout (currently hardcoded to 10 seconds)
	Hostname              string
	Username              string
	Password              string
	ValidateCerts         bool
	MaxConcurrentRequests int
}

// Config is created by the provide configure method
type Config struct {
	ConnectionProfiles map[string]ConnectionProfile
}

// GetConnectionProfile retrieves a connection profile based on name
// If name is empty and only one profile is defined, it is returned
func (c *Config) GetConnectionProfile(name string) (*ConnectionProfile, error) {
	if c == nil {
		return nil, fmt.Errorf("internal error, config is not initialized")
	}
	if len(c.ConnectionProfiles) == 0 {
		return nil, fmt.Errorf("error, at least one connection profile is required to connect to ONTAP")
	}
	if name == "" && len(c.ConnectionProfiles) == 1 {
		name = maps.Keys(c.ConnectionProfiles)[0]
	}
	if name == "" {
		return nil, fmt.Errorf("error, connection profile name is required if more than one profile is defined")
	}
	if profile, ok := c.ConnectionProfiles[name]; ok {
		return &profile, nil
	}
	return nil, fmt.Errorf("connection profile wiuth name %s is not defined", name)
}

// NewClient creates a RestClient based on the connection profile identified by cxProfileName
func (c *Config) NewClient(ctx context.Context, diags diag.Diagnostics, cxProfileName string) (*restclient.RestClient, error) {
	connectionProfile, err := c.GetConnectionProfile(cxProfileName)
	if err != nil {
		tflog.Error(ctx, err.Error())
		diags.AddError("failed to set connection profile", err.Error())
		return nil, err
	}
	var profile restclient.ConnectionProfile
	err = mapstructure.Decode(connectionProfile, &profile)
	if err != nil {
		msg := fmt.Sprintf("decode error on ConnectionProfile %#v to restclient.ConnectionProfile", connectionProfile)
		tflog.Error(ctx, msg)
		diags.AddError("unable to create REST client", msg)
		return nil, err
	}

	// TODO: get credentials from connection_profiles, using req.ProviderData
	client, err := restclient.NewClient(ctx, profile)
	if err != nil {
		msg := fmt.Sprintf("error creating REST client: %s", err)
		tflog.Error(ctx, msg)
		diags.AddError("unable to create REST client", msg)
		return nil, err
	}
	return client, err
}
