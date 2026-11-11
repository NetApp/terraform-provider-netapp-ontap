package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// StorageNvmeNamespaceGetDataModelONTAP describes the GET record data model using go types for mapping.
type StorageNvmeNamespaceGetDataModelONTAP struct {
	Name         string       `mapstructure:"name"`
	UUID         string       `mapstructure:"uuid"`
	SVM          svm          `mapstructure:"svm,omitempty"`
	OSType       string       `mapstructure:"os_type,omitempty"`
	Space        NvmeSpace     `mapstructure:"space,omitempty"`
}

// NvmeSpace describes the data model for space.
type NvmeSpace struct {
	Size       	int64 `mapstructure:"size,omitempty"`
	BlockSize   int64 `mapstructure:"block_size,omitempty"`
}

// StorageNvmeNamespaceResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type StorageNvmeNamespaceResourceBodyDataModelONTAP struct {
	Name      string    `mapstructure:"name,omitempty"`
	SVM       svm       `mapstructure:"svm,omitempty"`
	OsType    string    `mapstructure:"os_type,omitempty"`
	Space     NvmeSpace `mapstructure:"space,omitempty"`
}

// StorageNvmeNamespaceDataSourceFilterModel describes the data source data model for queries.
type StorageNvmeNamespaceDataSourceFilterModel struct {
	Name       string `mapstructure:"name"`
	SVMName    string `mapstructure:"svm.name"`
}

// GetStorageNvmeNamespaceByName gets a storage namespace by name and SVM.
func GetStorageNvmeNamespaceByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*StorageNvmeNamespaceGetDataModelONTAP, error) {
	api := "storage/namespaces"
	query := r.NewQuery()
	query.Set("name", name)
	query.Set("svm.name", svmName)
	query.Fields([]string{"name", "uuid", "svm.name", "os_type", "space"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_namespace info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageNvmeNamespaceGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_namespace data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetStorageNvmeNamespaceByUUID gets a storage namespace by UUID.
func GetStorageNvmeNamespaceByUUID(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) (*StorageNvmeNamespaceGetDataModelONTAP, error) {
	api := "storage/namespaces/" + uuid
	query := r.NewQuery()
	query.Fields([]string{"name", "uuid", "svm.name", "os_type", "space"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_namespace info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageNvmeNamespaceGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_namespace data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetStorageNvmeNamespaces gets storage namespaces for all resources matching a filter.
func GetStorageNvmeNamespaces(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *StorageNvmeNamespaceDataSourceFilterModel) ([]StorageNvmeNamespaceGetDataModelONTAP, error) {
	api := "storage/namespaces"
	query := r.NewQuery()
	query.Fields([]string{"name", "uuid", "svm.name", "os_type", "space"})
	if filter != nil {
		if filter.Name != "" {
			query.Add("name", filter.Name)
		}
		if filter.SVMName != "" {
			query.Add("svm.name", filter.SVMName)
		}
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_namespaces info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []StorageNvmeNamespaceGetDataModelONTAP
	for _, info := range response {
		var record StorageNvmeNamespaceGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_namespaces data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateStorageNvmeNamespace creates a storage namespace (POST /storage/namespaces).
func CreateStorageNvmeNamespace(errorHandler *utils.ErrorHandler, r restclient.RestClient, body StorageNvmeNamespaceResourceBodyDataModelONTAP) (*StorageNvmeNamespaceGetDataModelONTAP, error) {
	api := "storage/namespaces"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding storage_namespace body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create storage_namespace source - body: %#v", bodyMap))
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating storage_namespace", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageNvmeNamespaceGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding storage_namespace info", fmt.Sprintf("error on decode storage/namespaces info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create storage_namespace source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteStorageNvmeNamespace deletes a storage namespace (DELETE /storage/namespaces/{uuid}).
func DeleteStorageNvmeNamespace(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "storage/namespaces"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting storage_namespace", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// PatchStorageNvmeNamespace updates a storage namespace (PATCH /storage/namespaces/{uuid}).
func PatchStorageNvmeNamespace(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string, body StorageNvmeNamespaceResourceBodyDataModelONTAP) error {
	api := "storage/namespaces"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding storage_namespace body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api+"/"+uuid, query, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error patching storage_namespace", fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
