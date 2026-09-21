package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// ProtocolsSVMAuditGetDataModelONTAP describes the GET record data model using go types for mapping
type ProtocolsSVMAuditGetDataModelONTAP struct {
	SVM       svm              `mapstructure:"svm"`
	Enabled   bool             `mapstructure:"enabled"`
	Events    *EventsDataModel `mapstructure:"events"`
	Guarantee bool             `mapstructure:"guarantee"`
	Log       *LogDataModel    `mapstructure:"log"`
	LogPath   string           `mapstructure:"log_path"`
	ID        string           `mapstructure:"ID"`
}

// EventsDataModel describes the GET record data model using go types for mapping
type EventsDataModel struct {
	AuthorizationPolicy bool `mapstructure:"authorization_policy,omitempty"`
	CapStaging 		    bool `mapstructure:"cap_staging,omitempty"`
	CIFSLogonLogoff     bool `mapstructure:"cifs_logon_logoff,omitempty"`
	FileOperations      bool `mapstructure:"file_operations,omitempty"`
	FileShare           bool `mapstructure:"file_share,omitempty"`
	SecurityGroup       bool `mapstructure:"security_group,omitempty"`
	UserAccount         bool `mapstructure:"user_account,omitempty"`
}

// LogDataModel describes the GET record data model using go types for mapping
type LogDataModel struct {
	Format    string             `mapstructure:"format,omitempty"`
	Retention *RetentionDataModel `mapstructure:"retention,omitempty"`
	Rotation  *RotationDataModel  `mapstructure:"rotation,omitempty"`
}

// RetentionDataModel describes the retention data model using go types for mapping
type RetentionDataModel struct {
	Count    int64  `mapstructure:"count,omitempty"    json:"count,omitempty"`
	Duration string `mapstructure:"duration,omitempty" json:"duration,omitempty"`
}
// RotationDataModel describes the rotation data model using go types for mapping
type RotationDataModel struct {
	Size     int64              `mapstructure:"size,omitempty"     json:"size,omitempty"`
	Schedule *ScheduleDataModel `mapstructure:"schedule,omitempty" json:"schedule,omitempty"`
}

// ScheduleDataModel describes the schedule data model using go types for mapping
type ScheduleDataModel struct {
	Days     []int64 `mapstructure:"days,omitempty"     json:"days,omitempty"`
	Hours    []int64 `mapstructure:"hours,omitempty"    json:"hours,omitempty"`
	Minutes  []int64 `mapstructure:"minutes,omitempty"  json:"minutes,omitempty"`
	Months   []int64 `mapstructure:"months,omitempty"   json:"months,omitempty"`
	Weekdays []int64 `mapstructure:"weekdays,omitempty" json:"weekdays,omitempty"`
}

// ProtocolsSVMAuditResourceBodyDataModel describes the resource body data model using go types for mapping
type ProtocolsSVMAuditResourceBodyDataModel struct {
	SVM       svm              `mapstructure:"svm,omitempty"`
	Enabled   *bool            `mapstructure:"enabled,omitempty"`
	Events    *EventsDataModel `mapstructure:"events,omitempty"`
	Guarantee *bool            `mapstructure:"guarantee,omitempty"`
	Log       *LogDataModel    `mapstructure:"log,omitempty"`
	LogPath   *string          `mapstructure:"log_path,omitempty"`
}

// ProtocolsAuditConfigsFilterModel describes the data source filter model for protocols SVM audit
type ProtocolsAuditConfigsFilterModel struct {
	SVMName string `mapstructure:"svm.name"`
}

// GetSVMAuditConfig to get audit configuration given SVM name
func GetSVMAuditConfig(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string, version versionModelONTAP) (*ProtocolsSVMAuditGetDataModelONTAP, error) {
	api := "protocols/audit"

	query := r.NewQuery()
	query.Set("svm.name", svmName)
	var fields = []string{
		"svm",
		"enabled",
		"events",
		"log",
		"log_path",
	}
	if (version.Generation == 9 && version.Major >= 10) || (version.Generation > 9) {
		fields = append(fields, "guarantee")
	}
	query.Fields(fields)

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading audit info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP ProtocolsSVMAuditGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read svm_audit data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetAuditConfigs to get a list of audit configurations for given filter options
func GetAuditConfigs(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *ProtocolsAuditConfigsFilterModel, version versionModelONTAP) ([]ProtocolsSVMAuditGetDataModelONTAP, error) {
	api := "protocols/audit"
	query := r.NewQuery()
	var fields = []string{
		"svm",
		"enabled",
		"events",
		"log",
		"log_path",
	}
	if (version.Generation == 9 && version.Major >= 10) || (version.Generation > 9) {
		fields = append(fields, "guarantee")
	}
	query.Fields(fields)

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError(
				"error encoding audit configs filter info",
				fmt.Sprintf("error on filter %#v: %s", filter, err),
			)
		}
		query.SetValues(filterMap)
	}

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading audit configs info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP []ProtocolsSVMAuditGetDataModelONTAP
	for _, info := range response {
		var record ProtocolsSVMAuditGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(
				fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info),
			)
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read svm_audits data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateSVMAuditConfig creates audit configuration for a SVM
func CreateSVMAuditConfig(errorHandler *utils.ErrorHandler, r restclient.RestClient, request ProtocolsSVMAuditResourceBodyDataModel) (*ProtocolsSVMAuditGetDataModelONTAP, error) {
	api := "protocols/audit"
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error encoding protocols/audit body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error creating audit config",
			fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP ProtocolsSVMAuditGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error decoding audit config response",
			fmt.Sprintf("error on decoding protocols/audit info: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Created SVM audit config resource: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteSVMAuditConfig deletes audit configuration for a SVM
func DeleteSVMAuditConfig(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string) error {
	api := fmt.Sprintf("protocols/audit/%s", svmUUID)
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error deleting SVM audit config",
			fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}

// UpdateSVMAuditConfig modifies audit configuration for a SVM
func UpdateSVMAuditConfig(errorHandler *utils.ErrorHandler, r restclient.RestClient, request ProtocolsSVMAuditResourceBodyDataModel, svmUUID string) error {
	api := fmt.Sprintf("protocols/audit/%s", svmUUID)
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return errorHandler.MakeAndReportError(
			"error encoding protocols/audit/{svm.uuid} body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error modifying SVM audit config",
			fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}

// EnableDisableSVMAudit updates audit enabled state for a SVM by sending enabled in PATCH
func EnableDisableSVMAudit(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, enabled bool) error {
	api := fmt.Sprintf("protocols/audit/%s", svmUUID)

	body := map[string]interface{}{
		"enabled": enabled,
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error updating SVM auditing state",
			fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}
