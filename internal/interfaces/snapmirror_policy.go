package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// SnapmirrorPolicyGetDataModelONTAP defines the resource get data model
type SnapmirrorPolicyGetDataModelONTAP struct {
	Name                      string                  `mapstructure:"name"`
	SVM                       svm                     `mapstructure:"svm"`
	Type                      string                  `mapstructure:"type,omitempty"`
	SyncType                  string                  `mapstructure:"sync_type,omitempty"`
	Comment                   string                  `mapstructure:"comment"`
	TransferSchedule          TransferScheduleType    `mapstructure:"transfer_schedule"`
	NetworkCompressionEnabled bool                    `mapstructure:"network_compression_enabled"`
	Retention                 []RetentionGetDataModel `mapstructure:"retention,omitempty"`
	IdentityPreservation      string                  `mapstructure:"identity_preservation,omitempty"`
	CopyAllSourceSnapshots    bool                    `mapstructure:"copy_all_source_snapshots,omitempty"`
	CopyLatestSourceSnapshot  bool                    `mapstructure:"copy_latest_source_snapshot,omitempty"`
	UUID                      string                  `mapstructure:"uuid"`
}

// SnapmirrorPolicyGetRawDataModelONTAP defines the resource get data model
type SnapmirrorPolicyGetRawDataModelONTAP struct {
	Name                      string                     `mapstructure:"name"`
	SVM                       svm                        `mapstructure:"svm"`
	Type                      string                     `mapstructure:"type,omitempty"`
	SyncType                  string                     `mapstructure:"sync_type,omitempty"`
	Comment                   string                     `mapstructure:"comment"`
	TransferSchedule          TransferScheduleType       `mapstructure:"transfer_schedule"`
	NetworkCompressionEnabled bool                       `mapstructure:"network_compression_enabled"`
	Retention                 []RetentionGetRawDataModel `mapstructure:"retention"`
	IdentityPreservation      string                     `mapstructure:"identity_preservation,omitempty"`
	CopyAllSourceSnapshots    bool                       `mapstructure:"copy_all_source_snapshots,omitempty"`
	CopyLatestSourceSnapshot  bool                       `mapstructure:"copy_latest_source_snapshot,omitempty"`
	CreateSnapshotOnSource    bool                       `mapstructure:"create_snapshot_on_source,omitempty"`
	UUID                      string                     `mapstructure:"uuid"`
}

// RetentionGetRawDataModel defines the resource get retention model
type RetentionGetRawDataModel struct {
	CreationSchedule CreationScheduleModel `mapstructure:"creation_schedule,omitempty"`
	Count            string                `mapstructure:"count"`
	Label            string                `mapstructure:"label"`
	Prefix           string                `mapstructure:"prefix,omitempty"`
}

// RetentionGetDataModel defines the resource get retention model
type RetentionGetDataModel struct {
	CreationSchedule CreationScheduleModel `mapstructure:"creation_schedule,omitempty"`
	Count            int64                 `mapstructure:"count"`
	Label            string                `mapstructure:"label"`
	Prefix           string                `mapstructure:"prefix,omitempty"`
}

// CreationScheduleModel defines the resource creationschedule model
type CreationScheduleModel struct {
	Name string `mapstructure:"name"`
}

// SnapmirrorPolicyResourceBodyDataModelONTAP defines the resource data model
type SnapmirrorPolicyResourceBodyDataModelONTAP struct {
	Name                      string                   `mapstructure:"name"`
	SVM                       svm                      `mapstructure:"svm"`
	Type                      string                   `mapstructure:"type,omitempty"`
	SyncType                  string                   `mapstructure:"sync_type,omitempty"`
	Comment                   string                   `mapstructure:"comment"`
	TransferSchedule          TransferScheduleType     `mapstructure:"transfer_schedule,omitempty"`
	NetworkCompressionEnabled bool                     `mapstructure:"network_compression_enabled,omitempty"`
	Retention                 []map[string]interface{} `mapstructure:"retention,omitempty"`
	IdentityPreservation      string                   `mapstructure:"identity_preservation,omitempty"`
	CopyAllSourceSnapshots    bool                     `mapstructure:"copy_all_source_snapshots,omitempty"`
	CopyLatestSourceSnapshot  bool                     `mapstructure:"copy_latest_source_snapshot,omitempty"`
	CreateSnapshotOnSource    bool                     `mapstructure:"create_snapshot_on_source,omitempty"`
}

// TransferScheduleType describes the transfer_schedule
type TransferScheduleType struct {
	Name string `mapstructure:"name,omitempty"`
}

// UpdateSnapmirrorPolicyResourceBodyDataModelONTAP defines the resource update request body
type UpdateSnapmirrorPolicyResourceBodyDataModelONTAP struct {
	Comment                   string                   `mapstructure:"comment"`
	TransferSchedule          map[string]interface{}   `mapstructure:"transfer_schedule"`
	NetworkCompressionEnabled bool                     `mapstructure:"network_compression_enabled"`
	Retention                 []map[string]interface{} `mapstructure:"retention,omitempty"`
	IdentityPreservation      string                   `mapstructure:"identity_preservation,omitempty"`
}

// UpdateSyncSnapmirrorPolicyResourceBodyDataModelONTAP defins the sync type snapmirror policy update request body
type UpdateSyncSnapmirrorPolicyResourceBodyDataModelONTAP struct {
	Comment                   string                   `mapstructure:"comment"`
	NetworkCompressionEnabled bool                     `mapstructure:"network_compression_enabled"`
	Retention                 []map[string]interface{} `mapstructure:"retention,omitempty"`
}

// UpdateTransferScheduleType describes the transfer_schedule data type in update request
type UpdateTransferScheduleType struct {
	Name string `mapstructure:"name"`
	UUID string `mapstructure:"uuid,omitempty"`
}

// SnapmirrorPolicyFilterModel describes filter model
type SnapmirrorPolicyFilterModel struct {
	Name string `mapstructure:"name"`
}

// GetSnapmirrorPolicy by ID
func GetSnapmirrorPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, id string) (*SnapmirrorPolicyGetRawDataModelONTAP, error) {
	api := "snapmirror/policies/" + id
	statusCode, response, err := r.GetNilOrOneRecord(api, nil, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading snapmirror policy info", fmt.Sprintf("error on GET %s: %s", api, err))
	}
	var rawDataONTAP SnapmirrorPolicyGetRawDataModelONTAP
	if err := mapstructure.Decode(response, &rawDataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding snapmirror policy info", fmt.Sprintf("error on decode %s: %s, statusCode %d, response %#v", api, err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("\n777Read snapmirror policy source - udata: %#v", rawDataONTAP))
	return &rawDataONTAP, nil
}

// GetSnapmirrorPolicyByName to get snapmirror policy info
func GetSnapmirrorPolicyByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*SnapmirrorPolicyGetDataModelONTAP, error) {
	api := "snapmirror/policies"
	query := r.NewQuery()
	query.Set("name", name)
	if svmName == "" {
		query.Set("scope", "cluster")
	} else {
		query.Set("svm.name", svmName)
		query.Set("scope", "svm")
	}
	// TODO: copy_all_source_snapshots is 9.10 and up
	query.Fields(([]string{"name", "svm.name", "type", "comment", "transfer_schedule", "network_compression_enabled", "retention", "identity_preservation", "copy_all_source_snapshots", "uuid"}))
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading snapmirror/policies info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SnapmirrorPolicyGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read snapmirror/policies data source: %#v", dataONTAP))
	// If retention is [] we need to convert it to nil for Terrafrom to work correct
	return &dataONTAP, nil
}

// GetSnapmirrorPolicyDataSourceByName to get snapmirror policy data source info by name
func GetSnapmirrorPolicyDataSourceByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, version versionModelONTAP) (*SnapmirrorPolicyGetRawDataModelONTAP, error) {
	api := "snapmirror/policies"
	query := r.NewQuery()
	query.Set("name", name)

	fields := []string{"name", "svm.name", "type", "comment", "transfer_schedule", "network_compression_enabled",
		"retention", "identity_preservation", "uuid", "create_snapshot_on_source", "transfer_schedule.name", "sync_type"}
	if version.Generation == 9 && version.Major > 9 {
		fields = append(fields, "copy_all_source_snapshots")
	}
	if version.Generation == 9 && version.Major > 10 {
		fields = append(fields, "create_snapshot_on_source", "copy_latest_source_snapshot")
	}
	query.Fields(fields)

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading snapmirror/policies info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP SnapmirrorPolicyGetRawDataModelONTAP

	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read snapmirror/policies data source: %#v", dataONTAP))

	return &dataONTAP, nil
}

// GetSnapmirrorPolicies to get list of policies
func GetSnapmirrorPolicies(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *SnapmirrorPolicyFilterModel, version versionModelONTAP) ([]SnapmirrorPolicyGetRawDataModelONTAP, error) {
	api := "snapmirror/policies"
	query := r.NewQuery()

	fields := []string{"name", "svm.name", "type", "comment", "transfer_schedule", "network_compression_enabled",
		"retention", "identity_preservation", "uuid", "create_snapshot_on_source", "transfer_schedule.name", "sync_type"}
	if version.Generation == 9 && version.Major > 9 {
		fields = append(fields, "copy_all_source_snapshots")
	}
	if version.Generation == 9 && version.Major > 10 {
		fields = append(fields, "create_snapshot_on_source", "copy_latest_source_snapshot")
	}
	query.Fields(fields)
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding snapmirror/policies filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading snapmirror/policies info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []SnapmirrorPolicyGetRawDataModelONTAP
	for _, info := range response {
		var record SnapmirrorPolicyGetRawDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protcols_nfs_service data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateSnapmirrorPolicy to create snapmirror policy
func CreateSnapmirrorPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, body SnapmirrorPolicyResourceBodyDataModelONTAP) (*SnapmirrorPolicyGetRawDataModelONTAP, error) {
	api := "snapmirror/policies"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding snapmirror/policies body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating snapmirror/policies", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var rawDataONTAP SnapmirrorPolicyGetRawDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &rawDataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding snapmirror/policies info", fmt.Sprintf("error on decode snapmirror/policies info: %s, statusCode %d, response %#v", err, statusCode, response))
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create snapmirror/policies source - udata: %#v", rawDataONTAP))
	return &rawDataONTAP, nil
}

// UpdateSnapmirrorPolicy to update snapmirror policy
func UpdateSnapmirrorPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, data any, id string) error {
	api := "snapmirror/policies/" + id
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding update snapmirror/policies body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating export policy", fmt.Sprintf("error on PATCH %s: %s, statusCode %d, response %#v", api, err, statusCode, response))
	}

	return nil
}

// DeleteSnapmirrorPolicy to delete ip_interface
func DeleteSnapmirrorPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "snapmirror/policies/"
	statusCode, _, err := r.CallDeleteMethod(api+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting snapmirror/policies", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
