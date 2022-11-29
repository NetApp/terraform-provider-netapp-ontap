package provider

import (
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
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
	Version            string
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
	return nil, fmt.Errorf("connection profile with name %s is not defined", name)
}

// NewClient creates a RestClient based on the connection profile identified by cxProfileName
func (c *Config) NewClient(errorHandler *utils.ErrorHandler, cxProfileName string, resName string) (*restclient.RestClient, error) {
	connectionProfile, err := c.GetConnectionProfile(cxProfileName)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("failed to set connection profile", err.Error())
	}
	var profile restclient.ConnectionProfile
	err = mapstructure.Decode(connectionProfile, &profile)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("unable to create REST client",
			fmt.Sprintf("decode error on ConnectionProfile %#v to restclient.ConnectionProfile", connectionProfile))
	}
	// the tag resource_name/version will be used for telemetry
	client, err := restclient.NewClient(errorHandler.Ctx, profile, strings.Join([]string{resName, c.Version}, "/"))
	if err != nil {
		return nil, errorHandler.MakeAndReportError("unable to create REST client",
			fmt.Sprintf("error creating REST client: %s", err))
	}
	return client, err
}
