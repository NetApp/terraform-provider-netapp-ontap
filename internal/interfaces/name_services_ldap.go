package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// NameServicesLDAPGetDataModelONTAP describes the GET record data model using go types for mapping.
type NameServicesLDAPGetDataModelONTAP struct {
	SVMName            string   `mapstructure:"svm.name"`
	Servers            []string `mapstructure:"servers"`
	Schema             string   `mapstructure:"schema"`
	AdDomain           string   `mapstructure:"ad_domain,omitempty"`
	BaseDN             string   `mapstructure:"base_dn,omitempty"`
	BaseScope          string   `mapstructure:"base_scope,omitempty"`
	BindDN             string   `mapstructure:"bind_dn,omitempty"`
	BindAsCIFSServer   bool     `mapstructure:"bind_as_cifs_server,omitempty"`
	PreferredADServers []string `mapstructure:"preferred_ad_servers,omitempty"`
	Port               int64    `mapstructure:"port,omitempty"`
	QueryTimeout       int64    `mapstructure:"query_timeout,omitempty"`
	MinBindLevel       string   `mapstructure:"min_bind_level,omitempty"`
	UseStartTLS        bool     `mapstructure:"use_start_tls,omitempty"`
	ReferralEnabled    bool     `mapstructure:"referral_enabled,omitempty"`
	SessionSecurity    string   `mapstructure:"session_security,omitempty"`
	LDAPSEnabled       bool     `mapstructure:"ldaps_enabled,omitempty"`
}

// NameServicesLDAPResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type NameServicesLDAPResourceBodyDataModelONTAP struct {
	SVM svm `mapstructure:"svm"`
}

// NameServicesLDAPDataSourceFilterModel describes the data source data model for queries.
type NameServicesLDAPDataSourceFilterModel struct {
	SVMName string `mapstructure:"svm.name"`
}

// GetNameServicesLDAPByName to get name_services_ldap info
func GetNameServicesLDAPBySVMName(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string) (*NameServicesLDAPGetDataModelONTAP, error) {
	api := "name-services/ldap"
	query := r.NewQuery()
	query.Set("svm.name", svmName)

	query.Fields([]string{"svm.name", "servers", "schema", "ad_domain", "base_dn", "base_scope", "bind_dn",
		"bind_as_cifs_server", "preferred_ad_servers", "port", "query_timeout", "min_bind_level",
		"bind_dn", "use_start_tls", "referral_enabled", "session_security",
		"ldaps_enabled"})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading name_services_ldap info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP NameServicesLDAPGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read name_services_ldap data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetNameServicesLDAPs to get name_services_ldap info for all resources matching a filter
func GetNameServicesLDAPs(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *NameServicesLDAPDataSourceFilterModel) ([]NameServicesLDAPGetDataModelONTAP, error) {
	api := "name-services/ldap"
	query := r.NewQuery()
	query.Fields([]string{"svm.name", "servers", "schema", "ad_domain", "base_dn", "base_scope", "bind_dn",
		"bind_as_cifs_server", "preferred_ad_servers", "port", "query_timeout", "min_bind_level",
		"bind_dn", "use_start_tls", "referral_enabled", "session_security",
		"ldaps_enabled"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding name_services_ldaps filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading name_services_ldaps info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []NameServicesLDAPGetDataModelONTAP
	for _, info := range response {
		var record NameServicesLDAPGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read name_services_ldaps data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateNameServicesLDAP to create name_services_ldap
func CreateNameServicesLDAP(errorHandler *utils.ErrorHandler, r restclient.RestClient, body NameServicesLDAPResourceBodyDataModelONTAP) (*NameServicesLDAPGetDataModelONTAP, error) {
	api := "api_url"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding name_services_ldap body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating name_services_ldap", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP NameServicesLDAPGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding name_services_ldap info", fmt.Sprintf("error on decode storage/name_services_ldaps info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create name_services_ldap source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteNameServicesLDAP to delete name_services_ldap
func DeleteNameServicesLDAP(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "api_url"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting name_services_ldap", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
