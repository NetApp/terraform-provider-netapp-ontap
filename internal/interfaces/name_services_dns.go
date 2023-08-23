package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/interfaces/name_services_dns.go)
// replace NameServicesDNS with the name of the resource, following go conventions, eg IPInterface
// replace name_services_dns with the name of the resource, for logging purposes, eg ip_interface
// replace api_url with API, eg ip/interfaces
// delete these 5 lines

// NameServicesDNSGetDataModelONTAP describes the GET record data model using go types for mapping.
type NameServicesDNSGetDataModelONTAP struct {
	Domains []string          `mapstructure:"domains"`
	Servers []string          `mapstructure:"servers"`
	SVM     SvmDataModelONTAP `mapstructure:"svm"`
}

// GetNameServicesDNS to get name_services_dns info
func GetNameServicesDNS(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string) (*NameServicesDNSGetDataModelONTAP, error) {
	api := "name-services/dns"
	query := r.NewQuery()
	query.Set("svm.name", svmName)
	query.Fields([]string{"domains", "servers"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading name_services_dns info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP NameServicesDNSGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read name_services_dns data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// CreateNameServicesDNS Create a new DNS service
func CreateNameServicesDNS(errorHandler *utils.ErrorHandler, r restclient.RestClient, data NameServicesDNSGetDataModelONTAP) (*NameServicesDNSGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding DNS body", fmt.Sprintf("error on encoding name-services/dns body: %s, body: %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("carchi8py body is : %#v", body))
	statusCode, response, err := r.CallCreateMethod("name-services/dns", query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating DNS", fmt.Sprintf("error on POST name-services/dns: %s, statusCode %d", err, statusCode))
	}
	var dataONTAP NameServicesDNSGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding DNS info", fmt.Sprintf("error on decode name-services/dns info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create volume source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteNameServicesDNS deletes a DNS
func DeleteNameServicesDNS(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	statusCode, _, err := r.CallDeleteMethod("name-services/dns"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting DNS", fmt.Sprintf("error on DELETE name-services/dns: %s, statusCode %d", err, statusCode))
	}
	return nil
}
