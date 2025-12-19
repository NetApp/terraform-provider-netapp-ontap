package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// FpolicyEventFileOperations describes file operations model
type FpolicyEventFileOperations struct {
	Access    bool `mapstructure:"access"`
	Close     bool `mapstructure:"close"`
	Create    bool `mapstructure:"create"`
	CreateDir bool `mapstructure:"create_dir"`
	Delete    bool `mapstructure:"delete"`
	DeleteDir bool `mapstructure:"delete_dir"`
	GetAttr   bool `mapstructure:"getattr"`
	Link      bool `mapstructure:"link"`
	Lookup    bool `mapstructure:"lookup"`
	Open      bool `mapstructure:"open"`
	Read      bool `mapstructure:"read"`
	Rename    bool `mapstructure:"rename"`
	RenameDir bool `mapstructure:"rename_dir"`
	SetAttr   bool `mapstructure:"setattr"`
	Symlink   bool `mapstructure:"symlink"`
	Write     bool `mapstructure:"write"`
}

// FpolicyEventFilters describes filters model
type FpolicyEventFilters struct {
	CloseWithModification           bool `mapstructure:"close_with_modification"`
	CloseWithRead                   bool `mapstructure:"close_with_read"`
	CloseWithoutModification        bool `mapstructure:"close_without_modification"`
	ExcludeDirectory                bool `mapstructure:"exclude_directory"`
	FirstRead                       bool `mapstructure:"first_read"`
	FirstWrite                      bool `mapstructure:"first_write"`
	MonitorAds                      bool `mapstructure:"monitor_ads"`
	OfflineBit                      bool `mapstructure:"offline_bit"`
	OpenWithDeleteIntent            bool `mapstructure:"open_with_delete_intent"`
	OpenWithWriteIntent             bool `mapstructure:"open_with_write_intent"`
	SetAttrWithAccessTimeChange     bool `mapstructure:"setattr_with_access_time_change"`
	SetAttrWithAllocationSizeChange bool `mapstructure:"setattr_with_allocation_size_change"`
	SetAttrWithCreationTimeChange   bool `mapstructure:"setattr_with_creation_time_change"`
	SetAttrWithDaclChange           bool `mapstructure:"setattr_with_dacl_change"`
	SetAttrWithGroupChange          bool `mapstructure:"setattr_with_group_change"`
	SetAttrWithModeChange           bool `mapstructure:"setattr_with_mode_change"`
	SetAttrWithModifyTimeChange     bool `mapstructure:"setattr_with_modify_time_change"`
	SetAttrWithOwnerChange          bool `mapstructure:"setattr_with_owner_change"`
	SetAttrWithSaclChange           bool `mapstructure:"setattr_with_sacl_change"`
	SetAttrWithSizeChange           bool `mapstructure:"setattr_with_size_change"`
	WriteWithSizeChange             bool `mapstructure:"write_with_size_change"`
}

// ProtocolsFpolicyEventGetDataModelONTAP describes the GET record data model using go types for mapping.
type ProtocolsFpolicyEventGetDataModelONTAP struct {
	Name             string                     `mapstructure:"name"`
	Protocol         string                     `mapstructure:"protocol"`
	FileOperations   FpolicyEventFileOperations `mapstructure:"file_operations"`
	Filters          FpolicyEventFilters        `mapstructure:"filters"`
	VolumeMonitoring bool                       `mapstructure:"volume_monitoring"`
	SVM              SvmDataModelONTAP          `mapstructure:"svm"`
}

// ProtocolsFpolicyEventResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type ProtocolsFpolicyEventResourceBodyDataModelONTAP struct {
	Name             string                     `mapstructure:"name"`
	Protocol         string                     `mapstructure:"protocol,omitempty"`
	FileOperations   FpolicyEventFileOperations `mapstructure:"file_operations,omitempty"`
	Filters          FpolicyEventFilters        `mapstructure:"filters,omitempty"`
	VolumeMonitoring bool                       `mapstructure:"volume_monitoring,omitempty"`
}

// ProtocolsFpolicyEventResourceUpdateBodyDataModelONTAP describes the update body data model (for PATCH - without name field)
type ProtocolsFpolicyEventResourceUpdateBodyDataModelONTAP struct {
	Protocol         string                     `mapstructure:"protocol,omitempty"`
	FileOperations   FpolicyEventFileOperations `mapstructure:"file_operations,omitempty"`
	Filters          FpolicyEventFilters        `mapstructure:"filters,omitempty"`
	VolumeMonitoring *bool                      `mapstructure:"volume_monitoring,omitempty"`
}

// ProtocolsFpolicyEventDataSourceFilterModel describes filter model
type ProtocolsFpolicyEventDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// GetProtocolsFpolicyEventByName to get protocols_fpolicy_event info
func GetProtocolsFpolicyEventByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmUUID string) (*ProtocolsFpolicyEventGetDataModelONTAP, error) {
	api := fmt.Sprintf("protocols/fpolicy/%s/events/%s", svmUUID, name)
	query := r.NewQuery()
	query.Set("return_timeout", "15")
	query.Fields([]string{"name", "protocol", "file_operations", "filters", "volume_monitoring", "svm"})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_fpolicy_event info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsFpolicyEventGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_fpolicy_event source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetProtocolsFpolicyEvents to get protocols_fpolicy_event info for all resources matching a filter
func GetProtocolsFpolicyEvents(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ProtocolsFpolicyEventDataSourceFilterModel) ([]ProtocolsFpolicyEventGetDataModelONTAP, error) {
	api := "protocols/fpolicy/*/events"
	query := r.NewQuery()
	query.Set("return_timeout", "15")
	query.Fields([]string{"name", "protocol", "file_operations", "filters", "volume_monitoring", "svm"})

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding protocols_fpolicy_events filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading protocols_fpolicy_event info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []ProtocolsFpolicyEventGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsFpolicyEventGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read protocols_fpolicy_events data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateProtocolsFpolicyEvent to create protocols_fpolicy_event
func CreateProtocolsFpolicyEvent(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ProtocolsFpolicyEventResourceBodyDataModelONTAP, svmUUID string) (*ProtocolsFpolicyEventGetDataModelONTAP, error) {
	api := fmt.Sprintf("protocols/fpolicy/%s/events", svmUUID)
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding protocols_fpolicy_event body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Set("return_records", "true")

	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating protocols_fpolicy_event", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP ProtocolsFpolicyEventGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding protocols_fpolicy_event info", fmt.Sprintf("error on decode storage/protocols_fpolicy_events info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create protocols_fpolicy_event source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateProtocolsFpolicyEvent to update protocols_fpolicy_event
func UpdateProtocolsFpolicyEvent(errorHandler *utils.ErrorHandler, r restclient.RestClient, body ProtocolsFpolicyEventResourceUpdateBodyDataModelONTAP, svmUUID string, name string) error {
	api := fmt.Sprintf("protocols/fpolicy/%s/events/%s", svmUUID, name)
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return errorHandler.MakeAndReportError("error encoding protocols_fpolicy_event body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Set("return_timeout", "15")

	statusCode, _, err := r.CallUpdateMethod(api, query, bodyMap)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating protocols_fpolicy_event", fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}

// DeleteProtocolsFpolicyEvent to delete protocols_fpolicy_event
func DeleteProtocolsFpolicyEvent(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, name string) error {
	api := fmt.Sprintf("protocols/fpolicy/%s/events/%s", svmUUID, name)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting protocols_fpolicy_event", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
