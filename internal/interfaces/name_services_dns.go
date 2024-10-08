package interfaces

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// NameServicesDNSGetDataModelONTAP describes the GET record data model using go types for mapping.
type NameServicesDNSGetDataModelONTAP struct {
	Domains              []string          `mapstructure:"domains"`
	Servers              []string          `mapstructure:"servers"`
	SVM                  SvmDataModelONTAP `mapstructure:"svm"`
	SkipConfigValidation bool              `mapstructure:"skip_config_validation"`
}

// NameServicesDNSDataSourceFilterModel describes filter model.
type NameServicesDNSDataSourceFilterModel struct {
	SVMName string `tfsdk:"svm_name"`
	Domains string `tfsdk:"domains"`
	Servers string `tfsdk:"servers"`
}

// GetNameServicesDNS to get name_services_dns info
func GetNameServicesDNS(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string) (*NameServicesDNSGetDataModelONTAP, error) {
	api := "name-services/dns"
	query := r.NewQuery()
	query.Add("svm.name", svmName)
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

// GetListNameServicesDNSs to get name_services_dnss info
func GetListNameServicesDNSs(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *NameServicesDNSDataSourceFilterModel) ([]NameServicesDNSGetDataModelONTAP, error) {
	api := "name-services/dns"
	query := r.NewQuery()

	if filter != nil {
		if filter.SVMName != "" {
			query.Add("svm.name", strings.ToLower(filter.SVMName))
		}
		if filter.Domains != "" {
			query.Add("domains", strings.ToLower(filter.Domains))
		}
		if filter.Servers != "" {
			query.Add("servers", strings.ToLower(filter.Servers))
		}
	}

	query.Fields([]string{"svm.name", "domains", "servers"})

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading name_services_dns info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []NameServicesDNSGetDataModelONTAP
	for _, info := range response {
		var record NameServicesDNSGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read name_services_dns data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateNameServicesDNS Create a new DNS service
func CreateNameServicesDNS(errorHandler *utils.ErrorHandler, r restclient.RestClient, data NameServicesDNSGetDataModelONTAP) (*NameServicesDNSGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding DNS body", fmt.Sprintf("error on encoding name-services/dns body: %s, body: %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("name-services/dns body is : %#v", body))
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
	statusCode, _, err := r.CallDeleteMethod("name-services/dns/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting DNS", fmt.Sprintf("error on DELETE name-services/dns: %s, statusCode %d", err, statusCode))
	}
	return nil
}
