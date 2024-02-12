package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/interfaces/protocols_san_igroup.go)
// replace ProtocolsSanIgroup with the name of the resource, following go conventions, eg IPInterface
// replace protocols_san_igroup with the name of the resource, for logging purposes, eg ip_interface
// replace api_url with API, eg ip/interfaces
// delete these 5 lines

// ProtocolsSanIgroupGetDataModelONTAP describes the GET record data model using go types for mapping.
type ProtocolsSanIgroupGetDataModelONTAP struct {
	Name string `mapstructure:"name"`
	UUID string `mapstructure:"uuid"`
}

// ProtocolsSanIgroupResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ProtocolsSanIgroupResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name"`
	SVM  svm    `mapstructure:"svm"`
}

// ProtocolsSanIgroupDataSourceFilterModel describes the data source data model for queries.
type ProtocolsSanIgroupDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// GetProtocolsSanIgroupByName to get protocols_san_igroup info
func GetProtocolsSanIgroupByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*ProtocolsSanIgroupGetDataModelONTAP, error) {
	api := "api_url"
	query := r.NewQuery()
	query.Set("name", name)
	if svmName == "" {
		query.Set("scope", "cluster")
	} else {
		query.Set("svm.name", svmName)
		query.Set("scope", "svm")
	}
	query.Fields([]string{"name", "svm.name", "ip", "scope"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_san_igroup info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsSanIgroupGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_san_igroup data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetProtocolsSanIgroups to get protocols_san_igroup info for all resources matching a filter
func GetProtocolsSanIgroups(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ProtocolsSanIgroupDataSourceFilterModel) ([]ProtocolsSanIgroupGetDataModelONTAP, error) {
	api := "api_url"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "scope"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding protocols_san_igroups filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_san_igroups info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []ProtocolsSanIgroupGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsSanIgroupGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_san_igroups data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateProtocolsSanIgroup to create protocols_san_igroup
func CreateProtocolsSanIgroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ProtocolsSanIgroupResourceBodyDataModelONTAP) (*ProtocolsSanIgroupGetDataModelONTAP, error) {
	api := "api_url"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_san_igroup body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating protocols_san_igroup", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsSanIgroupGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols_san_igroup info", fmt.Sprintf("error on decode storage/protocols_san_igroups info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create protocols_san_igroup source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteProtocolsSanIgroup to delete protocols_san_igroup
func DeleteProtocolsSanIgroup(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "api_url"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting protocols_san_igroup", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
