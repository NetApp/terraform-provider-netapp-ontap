package interfaces

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// StorageAggregateGetDataModelONTAP describes the GET record data model using go types for mapping.
type StorageAggregateGetDataModelONTAP struct {
	Name           string                  `mapstructure:"name"`
	UUID           string                  `mapstructure:"uuid"`
	BlockStorage   AggregateBlockStorage   `mapstructure:"block_storage"`
	DataEncryption AggregateDataEncryption `mapstructure:"data_encryption"`
	SnaplockType   string                  `mapstructure:"snaplock_type"`
	State          string                  `mapstructure:"state"`
}

// AggregateDataEncryption describes data_encryption within StorageAggregateGetDataModelONTAP
type AggregateDataEncryption struct {
	SoftwareEncryptionEnabled bool `mapstructure:"software_encryption_enabled"`
}

// AggregateBlockStorage describes block_storage within StorageAggregateGetDataModelONTAP
type AggregateBlockStorage struct {
	Primary AggregateBlockStoragePrimary `mapstructure:"primary"`
	Mirror  AggregateBlockStorageMirror  `mapstructure:"mirror"`
}

// AggregateBlockStorageMirror describes mirror within AggregateBlockStorage
type AggregateBlockStorageMirror struct {
	Enabled bool `mapstructure:"enabled"`
}

// StorageAggregateResourceModel describes the resource data model.
type StorageAggregateResourceModel struct {
	Name           string                 `mapstructure:"name,omitempty"`
	State          string                 `mapstructure:"state,omitempty"`
	Node           map[string]string      `mapstructure:"node,omitempty"`
	BlockStorage   map[string]interface{} `mapstructure:"block_storage,omitempty"`
	SnaplockType   string                 `mapstructure:"snaplock_type,omitempty"`
	DataEncryption map[string]bool        `mapstructure:"data_encryption,omitempty"`
}

// AggregateBlockStoragePrimary describes primary within AggregateBlockStorage
type AggregateBlockStoragePrimary struct {
	DiskClass string `mapstructure:"disk_class,omitempty"`
	DiskCount int64  `mapstructure:"disk_count,omitempty"`
	RaidSize  int64  `mapstructure:"raid_size,omitempty"`
	RaidType  string `mapstructure:"raid_type,omitempty"`
}

// GetStorageAggregate to get aggregate info by uuid
func GetStorageAggregate(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) (*StorageAggregateGetDataModelONTAP, error) {
	statusCode, response, err := r.GetNilOrOneRecord("storage/aggregates/"+uuid, nil, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading aggregate info", fmt.Sprintf("error on GET storage/aggregates/%s: %s", uuid, err))
	}

	var dataONTAP *StorageAggregateGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding aggregate info", fmt.Sprintf("error on decode storage/aggregates: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read aggregate source - udata: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateStorageAggregate to create aggregate
func CreateStorageAggregate(errorHandler *utils.ErrorHandler, r restclient.RestClient, data StorageAggregateResourceModel, diskSize int) (*StorageAggregateGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding aggregate body", fmt.Sprintf("error on encoding storage/aggregates body: %s, body: %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	if diskSize > 0 {
		query.Add("disk_size", strconv.Itoa(diskSize))
	}
	statusCode, response, err := r.CallCreateMethod("storage/aggregates", query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating aggregate", fmt.Sprintf("error on POST storage/aggregates: %s, statusCode %d", err, statusCode))
	}

	var dataONTAP StorageAggregateGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding aggregate info", fmt.Sprintf("error on decode storage/aggregates info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create aggregate source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateStorageAggregate updates aggregate
func UpdateStorageAggregate(errorHandler *utils.ErrorHandler, r restclient.RestClient, data StorageAggregateResourceModel, diskSize int, uuid string) error {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding aggregate body", fmt.Sprintf("error on encoding storage/aggregates body: %s, body: %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	if diskSize > 0 {
		query.Add("disk_size", strconv.Itoa(diskSize))
	}
	// API has no option to return records
	statusCode, _, err := r.CallUpdateMethod(fmt.Sprintf("storage/aggregates/%s", uuid), query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating aggregate", fmt.Sprintf("error on PATCH storage/aggregates: %s, statusCode %d", err, statusCode))
	}
	return nil
}

// DeleteStorageAggregate to delete aggregate
func DeleteStorageAggregate(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	statusCode, _, err := r.CallDeleteMethod("storage/aggregates/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting aggregate", fmt.Sprintf("error on DELETE storage/aggregates: %s, statusCode %d", err, statusCode))
	}
	return nil
}
