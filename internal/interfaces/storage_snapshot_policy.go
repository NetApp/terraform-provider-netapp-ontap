package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// SnapshotPolicyGetDataModelONTAP describes the GET record data model using go types for mapping.
type SnapshotPolicyGetDataModelONTAP struct {
	Name    string     `mapstructure:"name"`
	SVM     Vserver    `mapstructure:"svm"`
	UUID    string     `mapstructure:"uuid,omitempty"`
	Copies  []CopyType `mapstructure:"copies"`
	Comment string     `mapstructure:"comment,omitempty"`
	Enabled bool       `mapstructure:"enabled"`
}

// SnapshotPolicyResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type SnapshotPolicyResourceBodyDataModelONTAP struct {
	Name    string                   `mapstructure:"name"`
	SVM     Vserver                  `mapstructure:"svm"`
	Copies  []map[string]interface{} `mapstructure:"copies"`
	Comment string                   `mapstructure:"comment,omitempty"`
	Enabled bool                     `mapstructure:"enabled"`
}

// SnapshotPolicyResourceUpdateRequestONTAP describe the PATCH body data model using go types for mapping
type SnapshotPolicyResourceUpdateRequestONTAP struct {
	Comment string `mapstructure:"comment,omitempty"`
	Enabled bool   `mapstructure:"enabled,omitempty"`
}

// CopyType describes the copy resouce data model
type CopyType struct {
	Schedule        Schedule `mapstructure:"schedule"`
	Count           int64    `mapstructure:"count"`
	Prefix          string   `mapstructure:"prefix,omitempty"`
	RetentionPeriod string   `mapstructure:"retention_period,omitempty"`
	SnapmirrorLabel string   `mapstructure:"snapmirror_label,omitempty"`
}

// Schedule describes the schedule resource data model
type Schedule struct {
	Name string `mapstructure:"name"`
}

// GetSnapshotPolicy to get storage_snapshot_policy info
func GetSnapshotPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) (*SnapshotPolicyGetDataModelONTAP, error) {
	api := "storage/snapshot-policies"
	query := r.NewQuery()
	query.Set("uuid", id)
	query.Fields([]string{"name", "svm.name", "copies", "scope", "enabled"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_snapshot_policy info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SnapshotPolicyGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_snapshot_policy data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetSnapshotPolicies to get storage_snapshot_policy info for all resources matching a filter
func GetSnapshotPolicies(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *SnapshotPolicyGetDataModelONTAP) ([]SnapshotPolicyGetDataModelONTAP, error) {
	api := "storage/snapshot-policies"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "scope", "enabled"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding storage_snapshot_policy filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_snapshot_policy info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []SnapshotPolicyGetDataModelONTAP
	for _, info := range response {
		var record SnapshotPolicyGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_snapshot_policy data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateSnapshotPolicy to create storage_snapshot_policy
func CreateSnapshotPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, body SnapshotPolicyResourceBodyDataModelONTAP) (*SnapshotPolicyGetDataModelONTAP, error) {
	api := "storage/snapshot-policies"
	var bodyMap map[string]interface{}

	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding storage_snapshot_policy body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}

	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating storage_snapshot_policy", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SnapshotPolicyGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding storage_snapshot_policy info", fmt.Sprintf("error on decode storage/storage_snapshot_policys info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create storage_snapshot_policy source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteSnapshotPolicy to delete storage_snapshot_policy
func DeleteSnapshotPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "storage/snapshot-policies"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting storage_snapshot_policy", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// UpdateSnapshotPolicy to update a Snapshot copy policy
func UpdateSnapshotPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, data SnapshotPolicyResourceUpdateRequestONTAP, id string) error {
	api := "storage/snapshot-policies"
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding snapshot policy body", fmt.Sprintf("error on encoding snapshot policy body: %s, body: %#v", err, data))
	}

	statusCode, _, err := r.CallUpdateMethod(api+"/"+id, nil, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating snapshot policy", fmt.Sprintf("error on PATCH storage/snapshot-policies: %s, statusCode %d", err, statusCode))
	}
	return nil
}
