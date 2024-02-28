package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// CifsServiceGetDataModelONTAP describes the GET record data model using go types for mapping.
type CifsServiceGetDataModelONTAP struct {
	Name            string            `mapstructure:"name"`
	SVM             svm               `mapstructure:"svm"`
	DefaultUnixUser string            `mapstructure:"default_unix_user"`
	Enabled         bool              `mapstructure:"enabled,omitempty"`
	Comment         string            `mapstructure:"comment,omitempty"`
	AdDomain        AdDomainDataModel `mapstructure:"ad_domain,omitempty"`
	Netbios         NetbiosDataModel  `mapstructure:"netbios,omitempty"`
	Security        SecurityDataModel `mapstructure:"security,omitempty"`
	//UUID string `mapstructure:"uuid"`
}

// AdDomainDataModel describes the ad_domain data model using go types for mapping.
type AdDomainDataModel struct {
	OrganizationalUnit string `mapstructure:"organizational_unit,omitempty"`
	User               string `mapstructure:"user,omitempty"`
	Password           string `mapstructure:"password,omitempty"`
	Fqdn               string `mapstructure:"fqdn,omitempty"`
}

// NetbiosDataModel describes the netbios data model using go types for mapping.
type NetbiosDataModel struct {
	Enabled     bool     `mapstructure:"enabled"`
	Aliases     []string `mapstructure:"aliases,omitempty"`
	WinsServers []string `mapstructure:"wins_servers,omitempty"`
}

// SecurityDataModel describes the security data model using go types for mapping.
type SecurityDataModel struct {
	RestrictAnonymous        string   `mapstructure:"restrict_anonymous,omitempty"`
	SmbSigning               bool     `mapstructure:"smb_signing"`
	SmbEncryption            bool     `mapstructure:"smb_encryption"`
	AdvertisedKdcEncryptions []string `mapstructure:"advertised_kdc_encryptions,omitempty"`
	LmCompatibilityLevel     string   `mapstructure:"lm_compatibility_level,omitempty"`
	AesNetlogonEnabled       bool     `mapstructure:"aes_netlogon_enabled" `
	TryLdapChannelBinding    bool     `mapstructure:"try_ldap_channel_binding"`
	LdapReferralEnabled      bool     `mapstructure:"ldap_referral_enabled"`
	EncryptDcConnection      bool     `mapstructure:"encrypt_dc_connection"`
	UseStartTLS              bool     `mapstructure:"use_start_tls"`
	SessionSecurity          string   `mapstructure:"session_security,omitempty"`
	UseLdaps                 bool     `mapstructure:"use_ldaps"`
}

// CifsServiceResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type CifsServiceResourceBodyDataModelONTAP struct {
	Name            string            `mapstructure:"name,omitempty"`
	SVM             svm               `mapstructure:"svm"`
	AdDomain        AdDomainDataModel `mapstructure:"ad_domain,omitempty"`
	Netbios         NetbiosDataModel  `mapstructure:"netbios,omitempty"`
	Comment         string            `mapstructure:"comment,omitempty"`
	Enabled         bool              `mapstructure:"enabled"`
	Security        SecurityDataModel `mapstructure:"security,omitempty"`
	DefaultUnixUser string            `mapstructure:"default_unix_user,omitempty"`
}

// CifsServiceDataSourceFilterModel describes the data source data model for queries.
type CifsServiceDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// GetCifsServiceByName to get protocols_cifs_service info
func GetCifsServiceByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string) (*CifsServiceGetDataModelONTAP, error) {
	api := "protocols/cifs/services"
	query := r.NewQuery()
	query.Set("name", name)

	query.Fields([]string{"name", "svm.name", "default_unix_user", "comment", "enabled", "security", "ad_domain", "netbios"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_service info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP CifsServiceGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_service data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetCifsServices to get protocols_cifs_service info for all resources matching a filter
func GetCifsServices(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *CifsServiceDataSourceFilterModel) ([]CifsServiceGetDataModelONTAP, error) {
	api := "protocols/cifs/services"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "default_unix_user", "comment", "enabled", "security", "ad_domain", "netbios"})

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding protocols_cifs_services filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_cifs_services info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []CifsServiceGetDataModelONTAP
	for _, info := range response {
		var record CifsServiceGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_cifs_services data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateCifsService to create protocols_cifs_service
func CreateCifsService(errorHandler *utils.ErrorHandler, r restclient.RestClient, force bool, body CifsServiceResourceBodyDataModelONTAP) (*CifsServiceGetDataModelONTAP, error) {
	api := "protocols/cifs/services"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_cifs_service body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	if force {
		query.Add("force", "true")
	}
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating protocols_cifs_service", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP CifsServiceGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols_cifs_service info", fmt.Sprintf("error on decode storage/protocols_cifs_services info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create protocols_cifs_service source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteCifsService to delete protocols_cifs_service
func DeleteCifsService(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmid string, force bool, body AdDomainDataModel) error {
	api := "protocols/cifs/services"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding protocols_cifs_service body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	if force {
		query.Add("force", "true")
	}
	statusCode, _, err := r.CallDeleteMethod(api+"/"+svmid, query, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting protocols_cifs_service", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// UpdateCifsService to update protocols_cifs_service
func UpdateCifsService(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmid string, force bool, body CifsServiceResourceBodyDataModelONTAP) error {
	api := "protocols/cifs/services" + "/" + svmid
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding protocols_cifs_service body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	if force {
		query.Add("force", "true")
	}
	statusCode, _, err := r.CallUpdateMethod(api, query, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating protocols_cifs_service", fmt.Sprintf("error on PUT %s: %s, statusCode %d", api, err, statusCode))
	}

	return nil
}
