package interfaces

import (
"fmt"
"strings"

"github.com/hashicorp/terraform-plugin-framework/types"
"github.com/hashicorp/terraform-plugin-log/tflog"
"github.com/mitchellh/mapstructure"
"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// removeDefaultPortFromMailHosts removes common SMTP ports from mail host addresses
// This handles the case where ONTAP API adds default ports but we want to show just the hostname
// to match user configuration when no explicit port was specified.
// Removes: :25 (default SMTP), :587 (SMTP-TLS), :465 (SMTPS) 
func removeDefaultPortFromMailHosts(mailHosts []string) []string {
	normalized := make([]string, len(mailHosts))
	commonPorts := []string{":25", ":587", ":465"}
	
	for i, host := range mailHosts {
		normalized[i] = host
		// Remove common SMTP ports to match user config when no port specified
		for _, port := range commonPorts {
			if strings.HasSuffix(host, port) {
				normalized[i] = strings.TrimSuffix(host, port)
				break
			}
		}
	}
	return normalized
}

// AutoSupportGetDataModelONTAP describes the GET data model using go types for mapping.
type AutoSupportGetDataModelONTAP struct {
	Enabled                       bool               `mapstructure:"enabled"`
	Transport                     string             `mapstructure:"transport,omitempty"`
	To                            []string           `mapstructure:"to,omitempty"`
	From                          string             `mapstructure:"from,omitempty"`
	ContactSupport                bool               `mapstructure:"contact_support"`
	PartnerAddresses              []string           `mapstructure:"partner_addresses,omitempty"`
	ProxyURL                      string             `mapstructure:"proxy_url,omitempty"`
	MailHosts                     []string           `mapstructure:"mail_hosts,omitempty"`
	IsMinimal                     bool               `mapstructure:"is_minimal"`
	OndemandEnabled               bool               `mapstructure:"ondemand_enabled"`
	SmtpEncryption                string             `mapstructure:"smtp_encryption,omitempty"`
}

// AutoSupportDataSourceModel describes the data source data model.
type AutoSupportDataSourceModel struct {
	CxProfileName                 types.String `tfsdk:"cx_profile_name"`
	ID                            types.String `tfsdk:"id"`
	Enabled                       types.Bool   `tfsdk:"enabled"`
	Transport                     types.String `tfsdk:"transport"`
	To                            types.Set    `tfsdk:"to_addresses"`
	From                          types.String `tfsdk:"from"`
	ContactSupport                types.Bool   `tfsdk:"contact_support"`
	PartnerAddresses              types.Set    `tfsdk:"partner_addresses"`
	ProxyURL                      types.String `tfsdk:"proxy_url"`
	MailHosts                     types.Set    `tfsdk:"mail_hosts"`
	IsMinimal                     types.Bool   `tfsdk:"is_minimal"`
	OndemandEnabled               types.Bool   `tfsdk:"ondemand_enabled"`
	SmtpEncryption                types.String `tfsdk:"smtp_encryption"`
}

// AutoSupportResourceBodyDataModelONTAP describes the body data model for requests using go types for mapping.
type AutoSupportResourceBodyDataModelONTAP struct {
	Enabled                       bool     `mapstructure:"enabled"`
	Transport                     string   `mapstructure:"transport,omitempty"`
	To                            []string `mapstructure:"to,omitempty"`
	From                          string   `mapstructure:"from,omitempty"`
	ContactSupport                bool     `mapstructure:"contact_support"`
	PartnerAddresses              []string `mapstructure:"partner_addresses,omitempty"`
	ProxyURL                      string   `mapstructure:"proxy_url,omitempty"`
	MailHosts                     []string `mapstructure:"mail_hosts,omitempty"`
	IsMinimal                     bool     `mapstructure:"is_minimal"`
	OndemandEnabled               bool     `mapstructure:"ondemand_enabled"`
	SmtpEncryption                string   `mapstructure:"smtp_encryption,omitempty"`
}

// GetAutoSupport to get AutoSupport info
func GetAutoSupport(errorHandler *utils.ErrorHandler, r restclient.RestClient, version versionModelONTAP) (*AutoSupportGetDataModelONTAP, error) {
	query := r.NewQuery()
	fields := []string{"enabled", "transport", "to", "from", "contact_support", "partner_addresses", 
		"proxy_url", "mail_hosts", "is_minimal"}

	// Add smtp_encryption field for ONTAP 9.15.0 or higher
	if version.Generation == 9 && version.Major >= 15 {
		fields = append(fields, "smtp_encryption")
	}

	// Add ondemand_enabled field for ONTAP 9.16.1 or higher
	if version.Generation == 9 && (version.Major > 16 || (version.Major == 16 && version.Minor >= 1)) {
		fields = append(fields, "ondemand_enabled")
	}

	query.Fields(fields)

	statusCode, response, err := r.GetNilOrOneRecord("support/autosupport", query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading AutoSupport info", 
			fmt.Sprintf("error on GET support/autosupport: %s, statusCode: %d", err, statusCode))
	}

	if response == nil {
		tflog.Debug(errorHandler.Ctx, "AutoSupport configuration not found")
		return nil, errorHandler.MakeAndReportError("error reading AutoSupport info",
			fmt.Sprintf("AutoSupport configuration is not found, statusCode %d", statusCode))
	}

	var dataONTAP AutoSupportGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding AutoSupport info",
			fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
	}

	// Normalize mail hosts to remove default port :25 for state consistency
	dataONTAP.MailHosts = removeDefaultPortFromMailHosts(dataONTAP.MailHosts)

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read AutoSupport data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateAutoSupport to update AutoSupport configuration using PATCH
func UpdateAutoSupport(errorHandler *utils.ErrorHandler, r restclient.RestClient, force bool, bodyMap map[string]interface{}) error {
	query := r.NewQuery()
	
	// Note: force parameter requires ONTAP 9.16.0 or higher
	if force {
		query.Add("force", "true")
	}

	statusCode, _, err := r.CallUpdateMethod("support/autosupport", query, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating AutoSupport", 
			fmt.Sprintf("error on PATCH support/autosupport: %s, statusCode: %d", err, statusCode))
	}
	return nil
}
