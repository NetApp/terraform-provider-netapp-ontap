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
	Name               string
	Volume             NameDataModel `mapstructure:"volume"`
	SVM                NameDataModel `mapstructure:"svm"`
	CreateTime         string        `mapstructure:"create_time"`
	ExpiryTime         string        `mapstructure:"expiry_time"`
	SnaplockExpiryTime string        `mapstructure:"snaplock_expiry_time"`
	State              string
	Size               float64
	Comment            string
	UUID               string
	SnapmirrorLabel    string `mapstructure:"snapmirror_label"`
}

// StorageVolumeSnapshotResourceModel describes the resource data model.
type StorageVolumeSnapshotResourceModel struct {
	Name               string `mapstructure:"name,omitempty"` // not set name if modificaiton is not rename
	ExpiryTime         string `mapstructure:"expiry_time,omitempty"`
	SnaplockExpiryTime string `mapstructure:"snaplock_expiry_time,omitempty"`
	Comment            string `mapstructure:"comment,omitempty"`
	SnapmirrorLabel    string `mapstructure:"snapmirror_label,omitempty"`
}

// StorageVolumeSnapshotDataSourceFilterModel describes filter model
type StorageVolumeSnapshotDataSourceFilterModel struct {
	Name string `tfsdk:"name"`
}

// GetUUIDStorageVolumeSnapshotsByName get a snapshot UUID based off name
func GetUUIDStorageVolumeSnapshotsByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, volumeUUID string) (*NameDataModel, error) {
	query := r.NewQuery()
	query.Add("name", name)
	query.Fields([]string{"name", "uuid"})
	api := "storage/volumes/" + volumeUUID + "/snapshots"
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading snapshot info",
			fmt.Sprintf("error on GET %s: %s, statuscode: %d", api, err, statusCode))
	}

	if response == nil {
		tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Snapshot %s not found", name))
		return nil, nil
	}
	var dataONTAP NameDataModel
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding snapshot info",
			fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage/volumes data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetStorageVolumeSnapshot to get snapshot info by uuid
func GetStorageVolumeSnapshot(errorHandler *utils.ErrorHandler, r restclient.RestClient, volumeUUID string, UUID string) (*StorageVolumeSnapshotGetDataModelONTAP, error) {
	api := fmt.Sprintf("storage/volumes/%s/snapshots/%s", volumeUUID, UUID)
	statusCode, response, err := r.GetNilOrOneRecord(api, nil, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading snapshot info", fmt.Sprintf("error on GET storage/volumes/%s/snapshots/%s: %s", volumeUUID, UUID, err))
	}

	var dataONTAP StorageVolumeSnapshotGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding snapshot info", fmt.Sprintf("error on decode storage/volumes/%s/snapshots/%s: %s, statusCode %d, response %#v", volumeUUID, UUID, err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read snapshot source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetStorageVolumeSnapshots to get a single snapshot info
func GetStorageVolumeSnapshots(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, volumeUUID string) (*StorageVolumeSnapshotGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Add("name", name)
	query.Fields([]string{"name", "create_time", "expiry_time", "state", "size", "comment", "volume", "volume.uuid", "snapmirror_label"})
	api := "storage/volumes/" + volumeUUID + "/snapshots"
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading snapshot info",
			fmt.Sprintf("error on GET %s: %s, statuscode: %d", api, err, statusCode))
	}

	if response == nil {
		tflog.Debug(errorHandler.Ctx, fmt.Sprintf("snapshot %s not found for volume UUID %s", name, volumeUUID))
		return nil, errorHandler.MakeAndReportError("error reading snapshot info",
			fmt.Sprintf("snapshot %s not found for volume UUID %s", name, volumeUUID))
	}

	var dataONTAP StorageVolumeSnapshotGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding snapshot info",
			fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage/volumes/snapshots data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetListStorageVolumeSnapshots to get snapshots info for all resources matching a filter
func GetListStorageVolumeSnapshots(errorHandler *utils.ErrorHandler, r restclient.RestClient, volumeUUID string, filter *StorageVolumeSnapshotDataSourceFilterModel) ([]StorageVolumeSnapshotGetDataModelONTAP, error) {
	query := r.NewQuery()

	if filter != nil {
		if filter.Name != "" {
			query.Add("name", filter.Name)
		}
	}

	query.Fields([]string{"name", "svm.name", "create_time", "expiry_time", "state", "size", "comment", "volume", "volume.uuid", "snapmirror_label"})
	api := "storage/volumes/" + volumeUUID + "/snapshots"
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading snapshots info",
			fmt.Sprintf("error on GET %s: %s, statuscode: %d", api, err, statusCode))
	}

	if response == nil {
		tflog.Debug(errorHandler.Ctx, fmt.Sprintf("snapshots not found for volume UUID %s", volumeUUID))
		return nil, nil
	}

	var dataONTAP []StorageVolumeSnapshotGetDataModelONTAP
	for _, info := range response {
		var record StorageVolumeSnapshotGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage/volumes/snapshots data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateStorageVolumeSnapshot to create a snapshot
func CreateStorageVolumeSnapshot(errorHandler *utils.ErrorHandler, r restclient.RestClient, data StorageVolumeSnapshotResourceModel, volumeUUID string) (*StorageVolumeSnapshotGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding snapshot body",
			fmt.Sprintf("err: %s, body %#v", err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	api := "storage/volumes/" + volumeUUID + "/snapshots"
	statusCode, response, err := r.CallCreateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating snapshot",
			fmt.Sprintf("error on POST %s: %s, statuscode: %d", api, err, statusCode))
	}

	var dataONTAP StorageVolumeSnapshotGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding snapshot info",
			fmt.Sprintf("err: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create volume source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateStorageVolumeSnapshot updates snapshot
func UpdateStorageVolumeSnapshot(errorHandler *utils.ErrorHandler, r restclient.RestClient, data StorageVolumeSnapshotResourceModel, volumeUUID string, UUID string) error {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding snapshot body", fmt.Sprintf("error on encoding storage/volumes/%s/snapshots/%s body: %s, body: %#v", volumeUUID, UUID, err, data))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")

	// API has no option to return records
	api := fmt.Sprintf("storage/volumes/%s/snapshots/%s", volumeUUID, UUID)
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating snapshot", fmt.Sprintf("error on PATCH storage/volumes/%s/snapshots/%s: %s, statusCode %d", volumeUUID, UUID, err, statusCode))
	}
	return nil
}

// DeleteStorageVolumeSnapshot to delete a snapshot
func DeleteStorageVolumeSnapshot(errorHandler *utils.ErrorHandler, r restclient.RestClient, volumeUUID string, uuid string) error {
	api := "storage/volumes/" + volumeUUID + "/snapshots/" + uuid
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting snapshot info",
			fmt.Sprintf("error on DELETE %s: %s, statuscode: %d", api, err, statusCode))
	}
	return nil
}
