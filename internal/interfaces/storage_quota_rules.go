package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// StorageQuotaRulesGetDataModelONTAP describes the GET record data model using go types for mapping.
type StorageQuotaRulesGetDataModelONTAP struct {
	SVM    svm    `mapstructure:"svm"`
	Volume volume `mapstructure:"volume"`
	Users  []User `mapstructure:"users,omitempty"`
	Group  Group  `mapstructure:"group,omitempty"`
	Qtree  Qtree  `mapstructure:"qtree,omitempty"`
	Type   string `mapstructure:"type"`
	Files  Files  `mapstructure:"files,omitempty"`
	UUID   string `mapstructure:"uuid"`
}

// StorageQuotaRulesResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type StorageQuotaRulesResourceBodyDataModelONTAP struct {
	SVM    svm      `mapstructure:"svm,omitempty"`
	Volume volume   `mapstructure:"volume,omitempty"`
	Users  []string `mapstructure:"users,omitempty"`
	Group  Group    `mapstructure:"group,omitempty"`
	Qtree  Qtree    `mapstructure:"qtree"`
	Type   string   `mapstructure:"type,omitempty"`
	Files  Files    `mapstructure:"files,omitempty"`
}

// StorageQuotaRulesResourceBodyUpdateModelONTAP describes the body data model using go types for mapping.
type StorageQuotaRulesResourceBodyUpdateModelONTAP struct {
	Files Files `mapstructure:"files,omitempty"`
}

// StorageQuotaRulesCreateResponse describes the Create record data model using go types for mapping.
type StorageQuotaRulesCreateResponse struct {
	UUID string `mapstructure:"uuid"`
}

type Qtree struct {
	Name string `mapstructure:"name"`
}

type Group struct {
	Name string `mapstructure:"name"`
}

type User struct {
	Name string `mapstructure:"name"`
}

type Files struct {
	SoftLimit int64 `mapstructure:"soft_limit"`
	HardLimit int64 `mapstructure:"hard_limit"`
}

// StorageQuotaRulesDataSourceFilterModel describes the data source data model for queries.
type StorageQuotaRulesDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// GetStorageQuotaRules to get storage_quota_rules info
func GetStorageQuotaRules(errorHandler *utils.ErrorHandler, r restclient.RestClient, volumeName string, svmName string, quotaType string, qtree string) (*StorageQuotaRulesGetDataModelONTAP, error) {
	api := "storage/quota/rules"
	query := r.NewQuery()
	query.Set("volume.name", volumeName)
	query.Set("type", quotaType)
	query.Set("qtree.name", qtree)
	query.Set("svm.name", svmName)
	query.Fields([]string{"volume", "svm", "type", "qtree", "users", "group", "files", "uuid"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_quota_rules info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageQuotaRulesGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_quota_rules data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetStorageQuotaRulesByUUID to get storage_lun info
func GetStorageQuotaRulesByUUID(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) (*StorageQuotaRulesGetDataModelONTAP, error) {
	api := "storage/quota/rules/" + uuid
	query := r.NewQuery()
	query.Fields([]string{"svm.name", "volume.name", "users", "group", "qtree", "type", "files", "uuid"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading quota_rules info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageQuotaRulesGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read quota_rules: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetOneORMoreStorageQuotaRules to get storage_quota_rules info for all resources matching a filter
func GetOneORMoreStorageQuotaRules(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *StorageQuotaRulesDataSourceFilterModel) ([]StorageQuotaRulesGetDataModelONTAP, error) {
	api := "storage/quota/rules"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "scope"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding tag_all_prefix filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading tag_all_prefix info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []StorageQuotaRulesGetDataModelONTAP
	for _, info := range response {
		var record StorageQuotaRulesGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read tag_all_prefix data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateStorageQuotaRules to create storage_quota_rules
func CreateStorageQuotaRules(errorHandler *utils.ErrorHandler, r restclient.RestClient, body StorageQuotaRulesResourceBodyDataModelONTAP) (*StorageQuotaRulesCreateResponse, error) {
	api := "storage/quota/rules"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding storage_quota_rules body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	if bodyMap["type"] == "user" {
		delete(bodyMap, "group")
	} else if bodyMap["type"] == "group" {
		delete(bodyMap, "users")
	} else if bodyMap["type"] == "tree" {
		delete(bodyMap, "users")
		delete(bodyMap, "group")
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating storage_quota_rules", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageQuotaRulesCreateResponse
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding storage_quota_rules info", fmt.Sprintf("error on decode storage/storage_quota_ruless info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create storage_quota_rules source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteStorageQuotaRules to delete storage_quota_rules
func DeleteStorageQuotaRules(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := fmt.Sprintf("storage/quota/rules/%s", uuid)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting storage_quota_rules", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// UpdateQuotaRules to update quota rules
func UpdateQuotaRules(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string, body StorageQuotaRulesResourceBodyUpdateModelONTAP) error {
	api := fmt.Sprintf("storage/quota/rules/%s", uuid)
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding storage_quota_rules body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	delete(bodyMap, "users")
	delete(bodyMap, "group")
	delete(bodyMap, "qtree")
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api, query, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating storage_quota_rules", fmt.Sprintf("error on Update %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
