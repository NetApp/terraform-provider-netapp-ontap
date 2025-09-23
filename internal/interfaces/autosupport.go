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

// removeDefaultPortFromMailHosts removes default SMTP port (:25) from mail host addresses
// This handles the case where ONTAP API returns "mail.example.com:25" but we want to show "mail.example.com" 
// to match user configuration. Custom ports like :587, :465 are preserved.
func removeDefaultPortFromMailHosts(mailHosts []string) []string {
	normalized := make([]string, len(mailHosts))
	for i, host := range mailHosts {
		// Remove default SMTP port :25 from the host to match config
		if strings.HasSuffix(host, ":25") {
			normalized[i] = strings.TrimSuffix(host, ":25")
		} else {
			// Custom port specified, keep as-is
			normalized[i] = host
		}
	}
	return normalized
}

// AutoSupportGetDataModelONTAP describes the GET data model using go types for mapping.
type AutoSupportGetDataModelONTAP struct {
	Enabled                       bool               `mapstructure:"enabled,omitempty"`
	Transport                     string             `mapstructure:"transport,omitempty"`
	To                            []string           `mapstructure:"to,omitempty"`
	From                          string             `mapstructure:"from,omitempty"`
	ContactSupport                bool               `mapstructure:"contact_support,omitempty"`
	PartnerAddresses              []string           `mapstructure:"partner_addresses,omitempty"`
	ProxyURL                      string             `mapstructure:"proxy_url,omitempty"`
	MailHosts                     []string           `mapstructure:"mail_hosts,omitempty"`
	IsMinimal                     bool               `mapstructure:"is_minimal,omitempty"`
	OndemandEnabled               bool               `mapstructure:"ondemand_enabled,omitempty"`
	SmtpEncryption                string             `mapstructure:"smtp_encryption,omitempty"`
}

// AutoSupportDataSourceModel describes the data source data model.
type AutoSupportDataSourceModel struct {
	CxProfileName                 types.String `tfsdk:"cx_profile_name"`
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

// GetAutoSupport to get AutoSupport info
func GetAutoSupport(errorHandler *utils.ErrorHandler, r restclient.RestClient) (*AutoSupportGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Fields([]string{"enabled", "transport", "to", "from", "contact_support", "partner_addresses", 
		"proxy_url", "mail_hosts", "is_minimal", "ondemand_enabled", "smtp_encryption"})

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

// GetAutoSupports to get multiple AutoSupport configs
func GetAutoSupports(errorHandler *utils.ErrorHandler, r restclient.RestClient) ([]AutoSupportGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Fields([]string{"enabled", "transport", "to", "from", "contact_support", "partner_addresses", 
		"proxy_url", "mail_hosts", "is_minimal", "ondemand_enabled", "smtp_encryption"})

statusCode, response, err := r.GetZeroOrMoreRecords("support/autosupport", query, nil)
if err != nil {
return nil, errorHandler.MakeAndReportError("error reading AutoSupport configs", 
fmt.Sprintf("error on GET support/autosupport: %s, statusCode: %d", err, statusCode))
}

var dataONTAP []AutoSupportGetDataModelONTAP
for _, info := range response {
var record AutoSupportGetDataModelONTAP
if err := mapstructure.Decode(info, &record); err != nil {
return nil, errorHandler.MakeAndReportError("error decoding AutoSupport info",
fmt.Sprintf("statusCode %d, response %#v", statusCode, info))
}
// Normalize mail hosts to remove default port :25 for state consistency
record.MailHosts = removeDefaultPortFromMailHosts(record.MailHosts)
dataONTAP = append(dataONTAP, record)
}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read AutoSupport data sources: %#v", dataONTAP))
	return dataONTAP, nil
}

// AutoSupportResourceBodyDataModelONTAP describes the body data model for requests using go types for mapping.
type AutoSupportResourceBodyDataModelONTAP struct {
	Enabled                       *bool     `mapstructure:"enabled,omitempty"`
	Transport                     *string   `mapstructure:"transport,omitempty"`
	To                            []string  `mapstructure:"to,omitempty"`
	From                          *string   `mapstructure:"from,omitempty"`
	ContactSupport                *bool     `mapstructure:"contact_support,omitempty"`
	PartnerAddresses              []string  `mapstructure:"partner_addresses,omitempty"`
	ProxyURL                      *string   `mapstructure:"proxy_url,omitempty"`
	MailHosts                     []string  `mapstructure:"mail_hosts,omitempty"`
	IsMinimal                     *bool     `mapstructure:"is_minimal,omitempty"`
	OndemandEnabled               *bool     `mapstructure:"ondemand_enabled,omitempty"`
	SmtpEncryption                *string   `mapstructure:"smtp_encryption,omitempty"`
}

// UpdateAutoSupport to update AutoSupport configuration using PATCH
func UpdateAutoSupport(errorHandler *utils.ErrorHandler, r restclient.RestClient, body AutoSupportResourceBodyDataModelONTAP) error {
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding AutoSupport body", 
			fmt.Sprintf("error on encoding support/autosupport body: %s, body: %#v", err, body))
	}

	statusCode, _, err := r.CallUpdateMethod("support/autosupport", nil, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating AutoSupport", 
			fmt.Sprintf("error on PATCH support/autosupport: %s, statusCode: %d", err, statusCode))
	}
	return nil
}
