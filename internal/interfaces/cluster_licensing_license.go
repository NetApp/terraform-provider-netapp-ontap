package interfaces

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ClusterLicensingLicenseKeyDataModelONTAP a single record from cluster/licensing/licenses
type ClusterLicensingLicenseKeyDataModelONTAP struct {
	Name     string                                          `mapstructure:"name"`
	Scope    string                                          `mapstructure:"scope"`
	State    string                                          `mapstructure:"state"`
	Licenses []ClusterLicensingLicenseLicensesDataModelONTAP `mapstructure:"licenses"`
}

// ClusterLicensingLicenseLicensesDataModelONTAP a single serial number
type ClusterLicensingLicenseLicensesDataModelONTAP struct {
	SerialNumber string `mapstructure:"serial_number"`
}

// ClusterLicensingLicenseResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ClusterLicensingLicenseResourceBodyDataModelONTAP struct {
	Keys []string `mapstructure:"keys"`
}

// GetClusterLicensingLicenses to get /cluster/licensing/licenses info
func GetClusterLicensingLicenses(errorHandler *utils.ErrorHandler, r restclient.RestClient) ([]ClusterLicensingLicenseKeyDataModelONTAP, error) {
	api := "/cluster/licensing/licenses"
	query := r.NewQuery()
	query.Fields([]string{"name", "state", "licenses"})
	statusCode, records, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && records == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading /cluster/licensing/licenses info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ClusterLicensingLicenseKeyDataModelONTAP
	keys := []ClusterLicensingLicenseKeyDataModelONTAP{}
	for _, record := range records {
		if err := mapstructure.Decode(record, &dataONTAP); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, record))
		}
		keys = append(keys, dataONTAP)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read /cluster/licensing/licenses data source: %#v", dataONTAP))
	return keys, nil
}

// CreateClusterLicensingLicense to create /cluster/licensing/licenses
func CreateClusterLicensingLicense(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ClusterLicensingLicenseResourceBodyDataModelONTAP) (*ClusterLicensingLicenseKeyDataModelONTAP, error) {
	api := "/cluster/licensing/licenses"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding /cluster/licensing/licenses body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating /cluster/licensing/licenses", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ClusterLicensingLicenseKeyDataModelONTAP
	// TODO: Fix it may be possible for a Key to unlock mutiple keys
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding /cluster/licensing/licenses info", fmt.Sprintf("error on decode storage//cluster/licensing/licensess info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create /cluster/licensing/licenses source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteClusterLicensingLicense to delete /cluster/licensing/licenses
func DeleteClusterLicensingLicense(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, serialNumber string) error {
	api := "/cluster/licensing/licenses"
	query := r.NewQuery()
	query.Add("serial_number", serialNumber)
	statusCode, _, err := r.CallDeleteMethod(api+"/"+name, query, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting /cluster/licensing/licenses", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
