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
	SVMName                   Vserver                 `mapstructure:"svm"`
	Type                      string                  `mapstructure:"type"`
	Comment                   string                  `mapstructure:"comment"`
	TransferSchedule          string                  `mapstructure:"transfer_schedule"`
	NetworkCompressionEnabled bool                    `mapstructure:"network_compression_enabled"`
	Retention                 []RetentionGetDataModel `mapstructure:"retention"`
	IdentityPreservation      string                  `mapstructure:"identity_preservation"`
	CopyAllSourceSnapshots    bool                    `mapstructure:"copy_all_source_snapshots"`
	UUID                      string                  `mapstructure:"uuid"`
}

// RetentionGetDataModel defines the resource get retention model
type RetentionGetDataModel struct {
	CreationSchedule CreationScheduleModel `json:"creation_schedule"`
	Count            int64                 `json:"count"`
	Label            string                `json:"label"`
	Prefix           string                `json:"prefix"`
}

// CreationScheduleModel defines the resource creationschedule model
type CreationScheduleModel struct {
	Name string `json:"name"`
}

// SnapmirrorPolicyResourceBodyDataModelONTAP defines the resource data model
type SnapmirrorPolicyResourceBodyDataModelONTAP struct {
	Name                      string                  `mapstructure:"name"`
	SVMName                   Vserver                 `mapstructure:"svm"`
	Type                      string                  `mapstructure:"type,omitempty"`
	Comment                   string                  `mapstructure:"comment,omitempty"`
	TransferSchedule          string                  `mapstructure:"transfer_schedule,omitempty"`
	NetworkCompressionEnabled bool                    `mapstructure:"network_compression_enabled,omitempty"`
	Retention                 []RetentionGetDataModel `mapstructure:"retention,omitempty"`
	IdentityPreservation      string                  `mapstructure:"identity_preservation,omitempty"`
	CopyAllSourceSnapshots    bool                    `mapstructure:"copy_all_source_snapshots,omitempty"`
}

// GetSnapmirrorPolicy to get snapmirror policy info
func GetSnapmirrorPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string) (*SnapmirrorPolicyGetDataModelONTAP, error) {
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

// CreateSnapmirrorPolicy to create snapmirror policy
func CreateSnapmirrorPolicy(errorHandler *utils.ErrorHandler, r restclient.RestClient, body SnapmirrorPolicyResourceBodyDataModelONTAP) (*SnapmirrorPolicyGetDataModelONTAP, error) {
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

	var dataONTAP SnapmirrorPolicyGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding snapmirror/policies info", fmt.Sprintf("error on decode snapmirror/policies info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create snapmirror/policies source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
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
