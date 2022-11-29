package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// StorageVolumeSnapshotGetDataModelONTAP describes the GET record data model using go types for mapping
type StorageVolumeSnapshotGetDataModelONTAP struct {
	Name   string
	Volume struct {
		UUID string
		Name string
	}
	CreateTime string `mapstructure:"create_time"`
	ExpiryTime string `mapstructure:"expiry_time"`
	State      string
	Size       float64
	Comment    string
	UUID       string
}

// StorageVolumeSnapshotResourceModel describes the resource data model.
type StorageVolumeSnapshotResourceModel struct {
	Name string `mapstructure:"name"`
}

// GetStorageVolumeSnapshots to get a single snapshot info
func GetStorageVolumeSnapshots(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, uuid string) (*StorageVolumeSnapshotGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Add("name", name)
	query.Fields([]string{"name", "create_time", "expiry_time", "state", "size", "comment", "volume", "volume.uuid"})
	api := "storage/volumes/" + uuid + "/snapshots"
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading snapshot info",
			fmt.Sprintf("error on GET %s: %s, statuscode: %d", api, err, statusCode))
	}

	if response == nil {
		tflog.Debug(errorHandler.Ctx, fmt.Sprintf("snapshot %s not found for volume UUID %s", name, uuid))
		return nil, nil
	}

	var dataONTAP StorageVolumeSnapshotGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding snapshot info",
			fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage/volumes/snapshots data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// CreateStorageVolumeSnapshot to create a snapshot
func CreateStorageVolumeSnapshot(errorHandler *utils.ErrorHandler, r restclient.RestClient, data StorageVolumeSnapshotResourceModel, uuid string) (*StorageVolumeSnapshotGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding snapshot body",
			fmt.Sprintf("err: %s, body %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	api := "storage/volumes/" + uuid + "/snapshots"
	statusCode, response, err := r.CallCreateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating snapshot",
			fmt.Sprintf("error on POST %s: %s, statuscode: %d", api, err, statusCode))
	}

	var dataONTAP StorageVolumeSnapshotGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding snapshot info",
			fmt.Sprintf("err: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create volume source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteStorageVolumeSnapshot to delete a snapshot
func DeleteStorageVolumeSnapshot(errorHandler *utils.ErrorHandler, r restclient.RestClient, volumeUUID string, uuid string) (*StorageVolumeSnapshotGetDataModelONTAP, error) {
	api := "storage/volumes/" + volumeUUID + "/snapshots/" + uuid
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error deleting snapshot info",
			fmt.Sprintf("error on DELETE %s: %s, statuscode: %d", api, err, statusCode))
	}
	return nil, nil
}
