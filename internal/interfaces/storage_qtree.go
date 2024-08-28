package interfaces

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// StorageQtreeGetDataModelONTAP describes the GET record data model using go types for mapping.
type StorageQtreeGetDataModelONTAP struct {
	Name          string         `mapstructure:"name"`
	UUID          string         `mapstructure:"uuid"`
	SVM           svm            `mapstructure:"svm"`
	SecurityStyle string         `mapstructure:"security_style"`
	NAS           qtreeNas       `mapstructure:"nas"`
	User          qtreeUser      `mapstructure:"user"`
	Volume        qtreeVloume    `mapstructure:"volume"`
	QoSPolicy     qtreeQosPolicy `mapstructure:"qos_policy"`
}

type qtreeNas struct {
	Path string `mapstructure:"path"`
}

type qtreeUser struct {
	ID   string `mapstructure:"id"`
	Name string `mapstructure:"name"`
}

type qtreeVloume struct {
	Name string `mapstructure:"name"`
	ID   string `mapstructure:"id"`
}

type qtreeQosPolicy struct {
	Name              string `mapstructure:"name"`
	ID                string `mapstructure:"uuid"`
	MaxThroughputIops int64  `mapstructure:"max_throughput_iops"`
	MaxThroughputMbps int64  `mapstructure:"max_throughput_mbps"`
	MinThroughputIops int64  `mapstructure:"min_throughput_iops"`
	MinThroughputMbps int64  `mapstructure:"min_throughput_mbps"`
}

// StorageQtreeResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type StorageQtreeResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name"`
	SVM  svm    `mapstructure:"svm"`
}

// StorageQtreeDataSourceFilterModel describes the data source data model for queries.
type StorageQtreeDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// GetStorageQtreeByName to get storage_qtree info
func GetStorageQtreeByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string, volumeName string) (*StorageQtreeGetDataModelONTAP, error) {
	api := "storage/qtrees"
	query := r.NewQuery()
	query.Set("name", name)
	query.Set("svm.name", svmName)
	query.Set("volume.name", volumeName)
	query.Fields([]string{"*"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	log.Printf("GetStorageQtreeByName response: %v", response)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_qtree info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageQtreeGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_qtree data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetStorageQtrees to get storage_qtree info for all resources matching a filter
func GetStorageQtrees(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *StorageQtreeDataSourceFilterModel) ([]StorageQtreeGetDataModelONTAP, error) {
	api := "storage/qtrees"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "security_style", "nas", "user", "volume"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding storage_qtrees filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_qtrees info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []StorageQtreeGetDataModelONTAP
	for _, info := range response {
		var record StorageQtreeGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_qtrees data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateStorageQtree to create storage_qtree
func CreateStorageQtree(errorHandler *utils.ErrorHandler, r restclient.RestClient, body StorageQtreeResourceBodyDataModelONTAP) (*StorageQtreeGetDataModelONTAP, error) {
	api := "api_url"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding storage_qtree body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating storage_qtree", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageQtreeGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding storage_qtree info", fmt.Sprintf("error on decode storage/storage_qtrees info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create storage_qtree source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteStorageQtree to delete storage_qtree
func DeleteStorageQtree(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "api_url"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting storage_qtree", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
