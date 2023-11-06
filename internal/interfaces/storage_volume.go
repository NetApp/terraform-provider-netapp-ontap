package interfaces

import (
	"fmt"
	"log"
	"math"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// StorageVolumeGetDataModelONTAP describes the GET record data model using go types for mapping.
type StorageVolumeGetDataModelONTAP struct {
	Name           string
	SVM            svm
	Space          Space
	State          string
	Type           string
	Comment        string
	SpaceGuarantee Guarantee `mapstructure:"guarantee"`
	NAS            NAS
	QOS            QOS
	Encryption     Encryption
	Efficiency     Efficiency
	SnapshotPolicy SnapshotPolicy `mapstructure:"snapshot_policy,omitempty"`
	TieringPolicy  TieringPolicy  `mapstructure:"tiering,omitempty"`
	Snaplock       Snaplock
	Analytics      Analytics
	Language       string
	Aggregates     []Aggregate
	UUID           string
}

// StorageVolumeResourceModel describes the resource data model.
type StorageVolumeResourceModel struct {
	Name           string                   `mapstructure:"name,omitempty"`
	SVM            svm                      `mapstructure:"svm,omitempty"`
	Space          Space                    `mapstructure:"space,omitempty"`
	State          string                   `mapstructure:"state,omitempty"`
	Type           string                   `mapstructure:"type,omitempty"`
	Comment        string                   `mapstructure:"comment,omitempty"`
	SpaceGuarantee Guarantee                `mapstructure:"guarantee,omitempty"`
	NAS            NAS                      `mapstructure:"nas,omitempty"`
	QOS            QOS                      `mapstructure:"qos,omitempty"`
	Encryption     Encryption               `mapstructure:"encryption,omitempty"`
	Efficiency     Efficiency               `mapstructure:"efficiency,omitempty"`
	SnapshotPolicy SnapshotPolicy           `mapstructure:"snapshot_policy,omitempty"`
	TieringPolicy  TieringPolicy            `mapstructure:"tiering,omitempty"`
	Snaplock       Snaplock                 `mapstructure:"snaplock,omitempty"`
	Analytics      Analytics                `mapstructure:"analytics,omitempty"`
	Language       string                   `mapstructure:"language,omitempty"`
	Aggregates     []map[string]interface{} `mapstructure:"aggregates,omitempty"`
}

// Aggregate describes the resource data model.
type Aggregate struct {
	Name string `mapstructure:"name"`
}

// Analytics describes the resource data model.
type Analytics struct {
	State string `mapstructure:"state,omitempty"`
}

// Space describes the resource data model.
type Space struct {
	Size         int          `mapstructure:"size,omitempty"`
	Snapshot     Snapshot     `mapstructure:"snapshot,omitempty"`
	LogicalSpace LogicalSpace `mapstructure:"logical_space,omitempty"`
}

// LogicalSpace describes the resource data model.
type LogicalSpace struct {
	Enforcement bool `mapstructure:"enforcement,omitempty"`
	Reporting   bool `mapstructure:"reporting,omitempty"`
}

// Efficiency describes the resource data model.
type Efficiency struct {
	Policy      Policy `mapstructure:"policy,omitempty"`
	Compression string `mapstructure:"compression,omitempty"`
}

// Snaplock describes the resource data model.
type Snaplock struct {
	Type string `mapstructure:"type,omitempty"`
}

// Policy describes the resource data model.
type Policy struct {
	Name string `mapstructure:"name,omitempty"`
}

// TieringPolicy describes the resource data model.
type TieringPolicy struct {
	Policy         string `mapstructure:"policy,omitempty"`
	MinCoolingDays int    `mapstructure:"min_cooling_days,omitempty"`
}

// Snapshot describes the resource data model.
type Snapshot struct {
	ReservePercent int `mapstructure:"reserve_percent,omitempty"`
}

// Guarantee describes the resource data model.
type Guarantee struct {
	Type string `mapstructure:"type,omitempty"`
}

// QOS describes the resource data model.
type QOS struct {
	Policy Policy `mapstructure:"policy,omitempty"`
}

// NAS describes the resource data model.
type NAS struct {
	ExportPolicy    ExportPolicy `mapstructure:"export_policy,omitempty"`
	JunctionPath    string       `mapstructure:"path,omitempty"`
	SecurityStyle   string       `mapstructure:"security_style,omitempty"`
	UnixPermissions int          `mapstructure:"unix_permissions,omitempty"`
	GroupID         int          `mapstructure:"gid"`
	UserID          int          `mapstructure:"uid"`
}

// NASData describes the data source model.
// type NASData struct {
// 	ExportPolicy    ExportPolicy `mapstructure:"export_policy,omitempty"`
// 	JunctionPath    string       `mapstructure:"path,omitempty"`
// 	SecurityStyle   string       `mapstructure:"security_style,omitempty"`
// 	UnixPermissions int          `mapstructure:"unix_permissions,omitempty"`
// 	GroupID         int          `mapstructure:"gid,omitempty"`
// 	UserID          int          `mapstructure:"uid,omitempty"`
// }

// Encryption describes the resource data model.
type Encryption struct {
	Enabled bool `mapstructure:"enabled,omitempty"`
}

// ExportPolicy describes the resource data model.
type ExportPolicy struct {
	Name string `mapstructure:"name,omitempty"`
}

// svm describes the resource data model.
type svm struct {
	Name string `mapstructure:"name,omitempty"`
}

// POW2BYTEMAP coverts size based on size unit.
var POW2BYTEMAP = map[string]int{
	// Here, 1 kb = 1024
	"bytes": 1,
	"b":     1,
	"k":     1024,
	"m":     int(math.Pow(1024, 2)),
	"g":     int(math.Pow(1024, 3)),
	"t":     int(math.Pow(1024, 4)),
	"p":     int(math.Pow(1024, 5)),
	"e":     int(math.Pow(1024, 6)),
	"z":     int(math.Pow(1024, 7)),
	"y":     int(math.Pow(1024, 8)),
	"kb":    1024,
	"mb":    int(math.Pow(1024, 2)),
	"gb":    int(math.Pow(1024, 3)),
	"tb":    int(math.Pow(1024, 4)),
	"pb":    int(math.Pow(1024, 5)),
	"eb":    int(math.Pow(1024, 6)),
	"zb":    int(math.Pow(1024, 7)),
	"yb":    int(math.Pow(1024, 8)),
}

// StorageVolumeDataSourceFilterModel describes the data source data model for queries.
type StorageVolumeDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// GetUUIDVolumeByName get a volumes UUID by volume name
func GetUUIDVolumeByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmUUID string, name string) (*NameDataModel, error) {
	query := r.NewQuery()
	query.Add("name", name)
	query.Add("svm.uuid", svmUUID)
	query.Fields([]string{"name", "uuid"})
	api := "storage/volumes/"
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading volume info",
			fmt.Sprintf("error on GET %s: %s, statuscode: %d", api, err, statusCode))
	}

	if response == nil {
		tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Volume %s not found", name))
		return nil, nil
	}
	var dataONTAP NameDataModel
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding volume info",
			fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage/volumes data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetStorageVolume to get volume info by uuid
func GetStorageVolume(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) (*StorageVolumeGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "aggregates", "space.size", "state", "type", "nas.export_policy.name", "nas.path", "guarantee.type", "space.snapshot.reserve_percent",
		"nas.security_style", "encryption.enabled", "efficiency.policy.name", "nas.unix_permissions", "nas.gid", "nas.uid", "snapshot_policy.name", "language", "qos.policy.name",
		"tiering.policy", "comment", "efficiency.compression", "tiering.min_cooling_days", "space.logical_space.enforcement", "space.logical_space.reporting", "snaplock.type", "analytics.state"})
	statusCode, response, err := r.GetNilOrOneRecord("storage/volumes/"+uuid, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading volume info", fmt.Sprintf("error on GET storage/volumes: %s", err))
	}
	log.Printf("raw is: %#v", response)
	var dataONTAP *StorageVolumeGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding volume info", fmt.Sprintf("error on decode storage/volumes: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read volume source - udata: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetStorageVolumeByName to get volume info by name and svm_name
func GetStorageVolumeByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name, svmName string) (*StorageVolumeGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Add("name", name)
	query.Add("svm.name", svmName)
	query.Add("return_records", "true")
	query.Fields([]string{"name", "svm.name", "aggregates", "space.size", "state", "type", "nas.export_policy.name", "nas.path", "guarantee.type", "space.snapshot.reserve_percent",
		"nas.security_style", "encryption.enabled", "efficiency.policy.name", "nas.unix_permissions", "nas.gid", "nas.uid", "snapshot_policy.name", "language", "qos.policy.name",
		"tiering.policy", "comment", "efficiency.compression", "tiering.min_cooling_days", "space.logical_space.enforcement", "space.logical_space.reporting", "snaplock.type", "analytics.state"})
	statusCode, response, err := r.GetNilOrOneRecord("storage/volumes", query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading volume info by name", fmt.Sprintf("error on GET storage/volumes: %s", err))
	}

	if response == nil {
		return nil, errorHandler.MakeAndReportError("no volume found", fmt.Sprintf("no volume found by name %s", name))
	}
	log.Printf("raw is: %#v", response)
	var dataONTAP *StorageVolumeGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding volume info by name", fmt.Sprintf("error on decode storage/volumes: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read volume source - udata: %#v", dataONTAP))
	return dataONTAP, nil
}

// GetStorageVolumes to get volumes info for all resources matching a filter
func GetStorageVolumes(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *StorageVolumeDataSourceFilterModel) ([]StorageVolumeGetDataModelONTAP, error) {
	api := "storage/volumes"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "aggregates", "space.size", "state", "type", "nas.export_policy.name", "nas.path", "guarantee.type", "space.snapshot.reserve_percent",
		"nas.security_style", "encryption.enabled", "efficiency.policy.name", "nas.unix_permissions", "nas.gid", "nas.uid", "snapshot_policy.name", "language", "qos.policy.name",
		"tiering.policy", "comment", "efficiency.compression", "tiering.min_cooling_days", "space.logical_space.enforcement", "space.logical_space.reporting", "snaplock.type", "analytics.state"})
	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError("error encoding storage volume filter info", fmt.Sprintf("error on filter %#v: %s", filter, err))
		}
		query.SetValues(filterMap)
	}

	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage volume info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []StorageVolumeGetDataModelONTAP
	for _, info := range response {
		var record StorageVolumeGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage volume data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateStorageVolume to create volume
func CreateStorageVolume(errorHandler *utils.ErrorHandler, r restclient.RestClient, data StorageVolumeResourceModel) (*StorageVolumeGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding volume body", fmt.Sprintf("error on encoding storage/volumes body: %s, body: %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod("storage/volumes", query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating volume", fmt.Sprintf("error on POST storage/volumes: %s, statusCode %d", err, statusCode))
	}

	var dataONTAP StorageVolumeGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding volume info", fmt.Sprintf("error on decode storage/volumes info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create volume source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteStorageVolume to delete volume
func DeleteStorageVolume(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	statusCode, _, err := r.CallDeleteMethod("storage/volumes/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting volume", fmt.Sprintf("error on DELETE storage/volumes: %s, statusCode %d", err, statusCode))
	}
	return nil
}

// UpddateStorageVolume to update volume
func UpddateStorageVolume(errorHandler *utils.ErrorHandler, r restclient.RestClient, data StorageVolumeResourceModel, ID string) error {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding volume body", fmt.Sprintf("error on encoding storage/volumes body: %s, body: %#v", err, data))
	}
	log.Printf("body body: %#v", body)
	statusCode, _, err := r.CallUpdateMethod("storage/volumes/"+ID, nil, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating volume", fmt.Sprintf("error on POST storage/volumes: %s, statusCode %d", err, statusCode))
	}
	return nil
}

// BoolToOnline converts bool to online or offline
func BoolToOnline(value bool) string {
	if value {
		return "online"
	}
	return "offline"
}

// OnlineToBool converts online or offline to bool value
func OnlineToBool(value string) bool {
	var boolValue bool
	if value == "online" {
		boolValue = true
	} else if value == "offline" {
		boolValue = false
	}
	return boolValue
}

// GetCompression gets values to compression and inlineCompression parameters
func GetCompression(compression bool, inlineCompression bool) string {
	if compression && inlineCompression {
		return "both"
	}
	if compression {
		return "background"
	}
	if inlineCompression {
		return "inline"
	}
	if !compression && !inlineCompression {
		return "none"
	}
	return ""
}

// ByteFormat converts bytes to respective byte size
func ByteFormat(value int64) (int64, string) {
	var number int64
	var unit string
	if value >= int64(math.Pow(1024, 6)) {
		number = value / int64(math.Pow(1024, 6))
		unit = "eb"
	} else if value >= int64(math.Pow(1024, 5)) {
		number = value / int64(math.Pow(1024, 5))
		unit = "pb"
	} else if value >= int64(math.Pow(1024, 4)) {
		number = value / int64(math.Pow(1024, 4))
		unit = "tb"
	} else if value >= int64(math.Pow(1024, 3)) {
		number = value / int64(math.Pow(1024, 3))
		unit = "gb"
	} else if value >= int64(math.Pow(1024, 2)) {
		number = value / int64(math.Pow(1024, 2))
		unit = "mb"
	} else if value >= 1024 {
		number = value / 1024
		unit = "kb"
	} else {
		number = value
		unit = "bytes"
	}

	return number, unit
}
