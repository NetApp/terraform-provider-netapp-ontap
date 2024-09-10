package interfaces

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// SecurityCertificateGetDataModelONTAP describes the GET record data model using go types for mapping.
type SecurityCertificateGetDataModelONTAP struct {
	Name         string `mapstructure:"name"`
	UUID         string `mapstructure:"uuid"`
	CommonName   string `mapstructure:"common_name"`
	SVM          svm    `mapstructure:"svm"`
	Scope        string `mapstructure:"scope"`
	Type         string `mapstructure:"type"`
	SerialNumber string `mapstructure:"serial_number"`
	CA           string `mapstructure:"ca"`
	HashFunction string `mapstructure:"hash_function"`
	KeySize      int64  `mapstructure:"key_size"`
	ExpiryTime   string `mapstructure:"expiry_time"`
}

// SecurityCertificateDataSourceFilterModel describes the data source data model for queries.
type SecurityCertificateDataSourceFilterModel struct {
	SVMName string `mapstructure:"svm.name"`
	Scope   string `mapstructure:"scope"`
}

// GetSecurityCertificate to get security_certificate info
func GetSecurityCertificate(errorHandler *utils.ErrorHandler, r restclient.RestClient, version versionModelONTAP, name string, common_name string, type_ string) (*SecurityCertificateGetDataModelONTAP, error) {
	api := "security/certificates"
	query := r.NewQuery()
	if name != "" {
		query.Set("name", name)
	} else {
		query.Set("common_name", common_name)
		query.Set("type", type_)
	}
	var fields = []string{"uuid", "common_name", "svm.name", "scope", "type", "serial_number", "ca", "hash_function", "key_size", "expiry_time"}
	if version.Generation == 9 && version.Major >= 8 {
		fields = append(fields, "name")
	}
	query.Fields(fields)

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		if strings.Contains(err.Error(), "or more records when only one is expected") {
			return nil, errorHandler.MakeAndReportError("error reading security_certificate info", "Duplicate records found with the same common_name.")
		}
		return nil, errorHandler.MakeAndReportError("error reading security_certificate info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SecurityCertificateGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read security_certificate data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetSecurityCertificateByUUID to get security_certificate info
func GetSecurityCertificateByUUID(errorHandler *utils.ErrorHandler, r restclient.RestClient, version versionModelONTAP, uuid string) (*SecurityCertificateGetDataModelONTAP, error) {
	api := "security/certificates/" + uuid
	query := r.NewQuery()
	var fields = []string{"uuid", "common_name", "svm.name", "scope", "type", "serial_number", "ca", "hash_function", "key_size", "expiry_time"}
	if version.Generation == 9 && version.Major >= 8 {
		fields = append(fields, "name")
	}
	query.Fields(fields)
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading security_certificate info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SecurityCertificateGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read security_certificate data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetSecurityCertificates to get security_certificate info for all resources matching a filter
func GetSecurityCertificates(errorHandler *utils.ErrorHandler, r restclient.RestClient, version versionModelONTAP, filter *SecurityCertificateDataSourceFilterModel) ([]SecurityCertificateGetDataModelONTAP, error) {
	api := "security/certificates"
	query := r.NewQuery()
	var fields = []string{"uuid", "common_name", "svm.name", "scope", "type", "serial_number", "ca", "hash_function", "key_size", "expiry_time"}
	if version.Generation == 9 && version.Major >= 8 {
		fields = append(fields, "name")
	}
	query.Fields(fields)
	if filter != nil {
		if filter.SVMName != "" {
			query.Add("svm.name", filter.SVMName)
			query.Add("scope", "svm")
		} else {
			// set scope to cluster
			query.Add("scope", strings.ToLower(filter.Scope))
		}
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("security account filter: %+v", query))
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading security_certificates info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []SecurityCertificateGetDataModelONTAP
	for _, info := range response {
		var record SecurityCertificateGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read security_certificates data source: %#v", dataONTAP))
	return dataONTAP, nil
}
