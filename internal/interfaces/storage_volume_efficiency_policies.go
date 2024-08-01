package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// StorageVolumeEfficiencyPoliciesGetDataModelONTAP describes the GET record data model using go types for mapping.
type StorageVolumeEfficiencyPoliciesGetDataModelONTAP struct {
	Name                  string   `mapstructure:"name"`
	UUID                  string   `mapstructure:"uuid"`
	SVM                   svm      `mapstructure:"svm"`
	Type                  string   `mapstructure:"type,omitempty"`
	Schedule              schedule `mapstructure:"schedule,omitempty"`
	Duration              int64    `mapstructure:"duration,omitempty"`
	StartThresholdPercent int64    `mapstructure:"start_threshold_percent,omitempty"`
	QOSPolicy             string   `mapstructure:"qos_policy,omitempty"`
	Comment               string   `mapstructure:"comment,omitempty"`
	Enabled               bool     `mapstructure:"enabled,omitempty"`
}

// StorageVolumeEfficiencyPoliciesResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type StorageVolumeEfficiencyPoliciesResourceBodyDataModelONTAP struct {
	Name                  string   `mapstructure:"name"`
	SVM                   svm      `mapstructure:"svm"`
	Type                  string   `mapstructure:"type,omitempty"`
	Schedule              schedule `mapstructure:"schedule,omitempty"`
	Duration              int64    `mapstructure:"duration,omitempty"`
	StartThresholdPercent int64    `mapstructure:"start_threshold_percent,omitempty"`
	QOSPolicy             string   `mapstructure:"qos_policy,omitempty"`
	Comment               string   `mapstructure:"comment,omitempty"`
	Enabled               bool     `mapstructure:"enabled,omitempty"`
}

// UpdateStorageVolumeEfficiencyPoliciesResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type UpdateStorageVolumeEfficiencyPoliciesResourceBodyDataModelONTAP struct {
	Type                  string   `mapstructure:"type,omitempty"`
	Schedule              schedule `mapstructure:"schedule,omitempty"`
	Duration              int64    `mapstructure:"duration,omitempty"`
	StartThresholdPercent int64    `mapstructure:"start_threshold_percent,omitempty"`
	QOSPolicy             string   `mapstructure:"qos_policy,omitempty"`
	Comment               string   `mapstructure:"comment,omitempty"`
	Enabled               bool     `mapstructure:"enabled,omitempty"`
}

// StorageVolumeEfficiencyPoliciesDataSourceFilterModel describes the data source data model for queries.
type StorageVolumeEfficiencyPoliciesDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

type schedule struct {
	Name string `mapstructure:"name,omitempty"`
}

// GetStorageVolumeEfficiencyPoliciesByUUID to get VolumeEfficiencyPoliciy info
func GetStorageVolumeEfficiencyPoliciesByUUID(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) (*StorageVolumeEfficiencyPoliciesGetDataModelONTAP, error) {
	api := "storage/volume-efficiency-policies/" + uuid
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "type", "qos_policy", "comment", "enabled", "schedule", "duration", "start_threshold_percent", "uuid"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_volume_efficiency_policies info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageVolumeEfficiencyPoliciesGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_lun data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetStorageVolumeEfficiencyPoliciesByName to get storage_volume_efficiency_policies info
func GetStorageVolumeEfficiencyPoliciesByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*StorageVolumeEfficiencyPoliciesGetDataModelONTAP, error) {
	api := "storage/volume-efficiency-policies"
	query := r.NewQuery()
	query.Set("name", name)
	query.Set("svm.name", svmName)
	query.Fields([]string{"name", "svm.name", "type", "qos_policy", "comment", "enabled", "schedule", "duration", "start_threshold_percent", "uuid"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_volume_efficiency_policies info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageVolumeEfficiencyPoliciesGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_volume_efficiency_policies data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetStorageVolumeEfficiencyPoliciess to get storage_volume_efficiency_policies info for all resources matching a filter
func GetStorageVolumeEfficiencyPoliciess(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *StorageVolumeEfficiencyPoliciesDataSourceFilterModel) ([]StorageVolumeEfficiencyPoliciesGetDataModelONTAP, error) {
	api := "api_url"
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

	var dataONTAP []StorageVolumeEfficiencyPoliciesGetDataModelONTAP
	for _, info := range response {
		var record StorageVolumeEfficiencyPoliciesGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read tag_all_prefix data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateStorageVolumeEfficiencyPolicies to create storage_volume_efficiency_policies
func CreateStorageVolumeEfficiencyPolicies(errorHandler *utils.ErrorHandler, r restclient.RestClient, body StorageVolumeEfficiencyPoliciesResourceBodyDataModelONTAP) (*StorageVolumeEfficiencyPoliciesGetDataModelONTAP, error) {
	api := "storage/volume-efficiency-policies"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding storage_volume_efficiency_policies body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating storage_volume_efficiency_policies", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageVolumeEfficiencyPoliciesGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding storage_volume_efficiency_policies info", fmt.Sprintf("error on decode storage/storage_volume_efficiency_policiess info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create storage_volume_efficiency_policies source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteStorageVolumeEfficiencyPolicies to delete storage_volume_efficiency_policies
func DeleteStorageVolumeEfficiencyPolicies(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "storage/volume-efficiency-policies"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting storage_volume_efficiency_policies", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// UpdateStorageVolumeEfficiencyPolicies to update storage_volume_efficiency_policies
func UpdateStorageVolumeEfficiencyPolicies(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string, body UpdateStorageVolumeEfficiencyPoliciesResourceBodyDataModelONTAP) error {
	api := "storage/volume-efficiency-policies"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding storage_volume_efficiency_policies body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api+"/"+uuid, query, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating storage_volume_efficiency_policies", fmt.Sprintf("error on Update %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
