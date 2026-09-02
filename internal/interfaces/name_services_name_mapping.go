package interfaces

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// NameServicesNameMappingGetDataModelONTAP describes the GET record data model using go types for mapping.
type NameServicesNameMappingGetDataModelONTAP struct {
	SVM         SvmDataModelONTAP `mapstructure:"svm"`
	Direction   string            `mapstructure:"direction"`
	Pattern     string            `mapstructure:"pattern"`
	Index       int               `mapstructure:"index"`
	ClientMatch string            `mapstructure:"client_match"`
	Replacement string            `mapstructure:"replacement"`
}

// NameServicesNameMappingResourceBodyDataModelONTAP describes the request body data model using go types for mapping.
type NameServicesNameMappingResourceBodyDataModelONTAP struct {
	SVM         svm    `mapstructure:"svm"`
	Direction   string `mapstructure:"direction,omitempty"`
	Pattern     string `mapstructure:"pattern,omitempty"`
	Index       int    `mapstructure:"index,omitempty"`
	ClientMatch string `mapstructure:"client_match,omitempty"`
	Replacement string `mapstructure:"replacement,omitempty"`
}

// NameServicesNameMappingDataSourceFilterModel describes the data source filter model for queries.
type NameServicesNameMappingDataSourceFilterModel struct {
	SVMName   string `mapstructure:"svm.name"`
	Direction string `mapstructure:"direction"`
}

// GetNameServicesNameMappingByIndex gets a single name mapping identified by svm name, direction, and index position.
func GetNameServicesNameMappingByIndex(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string, direction string, index int) (*NameServicesNameMappingGetDataModelONTAP, error) {
	api := "name-services/name-mappings"
	query := r.NewQuery()
	query.Set("svm.name", svmName)
	query.Set("direction", direction)
	query.Set("index", strconv.Itoa(index))
	query.Fields([]string{"direction", "pattern", "index", "client_match", "replacement", "svm.name", "svm.uuid"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading name_services_name_mapping info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP NameServicesNameMappingGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read name_services_name_mapping: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetNameServicesNameMappings gets all name mappings matching an optional filter.
func GetNameServicesNameMappings(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *NameServicesNameMappingDataSourceFilterModel) ([]NameServicesNameMappingGetDataModelONTAP, error) {
	api := "name-services/name-mappings"
	query := r.NewQuery()
	query.Fields([]string{"direction", "pattern", "index", "client_match", "replacement", "svm.name", "svm.uuid"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding name_services_name_mapping filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading name_services_name_mappings info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []NameServicesNameMappingGetDataModelONTAP
	for _, info := range response {
		var record NameServicesNameMappingGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read name_services_name_mappings: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateNameServicesNameMapping creates a name mapping.
func CreateNameServicesNameMapping(errorHandler *utils.ErrorHandler, r restclient.RestClient, body NameServicesNameMappingResourceBodyDataModelONTAP) (*NameServicesNameMappingGetDataModelONTAP, error) {
	api := "name-services/name-mappings"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding name_services_name_mapping body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating name_services_name_mapping", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP NameServicesNameMappingGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding name_services_name_mapping info", fmt.Sprintf("error on decode name_services_name_mappings info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create name_services_name_mapping - data: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateNameServicesNameMapping updates a name mapping.
// id is the composite path "{svm_uuid}/{direction}/{index}".
func UpdateNameServicesNameMapping(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string, body NameServicesNameMappingResourceBodyDataModelONTAP) error {
	api := "name-services/name-mappings"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding name_services_name_mapping body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api+"/"+id, query, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating name_services_name_mapping", fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// DeleteNameServicesNameMapping deletes a name mapping.
// id is the composite path "{svm_uuid}/{direction}/{index}".
func DeleteNameServicesNameMapping(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) error {
	api := "name-services/name-mappings"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+id, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting name_services_name_mapping", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
