package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// TODO:
// copy this file to match you data source (should match internal/interfaces/storage_lun.go)
// replace StorageLun with the name of the resource, following go conventions, eg IPInterface
// replace storage_lun with the name of the resource, for logging purposes, eg ip_interface
// replace api_url with API, eg ip/interfaces
// delete these 5 lines

// StorageLunGetDataModelONTAP describes the GET record data model using go types for mapping.
type StorageLunGetDataModelONTAP struct {
	Name string `mapstructure:"name"`
	UUID string `mapstructure:"uuid"`
}

// StorageLunResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type StorageLunResourceBodyDataModelONTAP struct {
	Name string `mapstructure:"name"`
	SVM  svm    `mapstructure:"svm"`
}

// StorageLunDataSourceFilterModel describes the data source data model for queries.
type StorageLunDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// GetStorageLunByName to get storage_lun info
func GetStorageLunByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string, volumeName string) (*StorageLunGetDataModelONTAP, error) {
	api := "storage/luns"
	query := r.NewQuery()
	query.Set("name", name)
	query.Set("svm.name", svmName)
	query.Set("location.volume.name", volumeName)
	query.Fields([]string{"name", "svm.name", "ip", "scope"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_lun info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageLunGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_lun data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetStorageLuns to get storage_lun info for all resources matching a filter
func GetStorageLuns(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *StorageLunDataSourceFilterModel) ([]StorageLunGetDataModelONTAP, error) {
	api := "api_url"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "scope"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding storage_luns filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_luns info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []StorageLunGetDataModelONTAP
	for _, info := range response {
		var record StorageLunGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_luns data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateStorageLun to create storage_lun
func CreateStorageLun(errorHandler *utils.ErrorHandler, r restclient.RestClient, body StorageLunResourceBodyDataModelONTAP) (*StorageLunGetDataModelONTAP, error) {
	api := "api_url"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding storage_lun body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating storage_lun", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageLunGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding storage_lun info", fmt.Sprintf("error on decode storage/storage_luns info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create storage_lun source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteStorageLun to delete storage_lun
func DeleteStorageLun(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "api_url"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting storage_lun", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
