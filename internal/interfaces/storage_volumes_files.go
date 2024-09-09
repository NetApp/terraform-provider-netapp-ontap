package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// StorageVolumesFilesGetDataModelONTAP describes the GET record data model using go types for mapping.
type StorageVolumesFilesGetDataModelONTAP struct {
	Name             string `mapstructure:"name"`
	Path             string `mapstructure:"path"`
	Type             string `mapstructure:"type"`
	Volume           volume `mapstructure:"volume"`
	FillEnabled      bool   `mapstructure:"fill_enabled"`
	Size             int    `mapstructure:"size"`
	OverwriteEnabled bool   `mapstructure:"overwrite_enabled"`
	GroupID          int    `mapstructure:"group_id"`
	HardLinksCount   int    `mapstructure:"hard_links_count"`
	BytesUsed        int    `mapstructure:"bytes_used"`
	OwnerID          int    `mapstructure:"owner_id"`
	InodeNumber      int    `mapstructure:"inode_number"`
	IsEmpty          bool   `mapstructure:"is_empty"`
	Target           string `mapstructure:"target"`
}

// VolumesFilesGetDataModelONTAP describes the GET record data model using go types for mapping.
type VolumesFilesGetDataModelONTAP struct {
	UUID            string `mapstructure:"uuid"`
	Path            string `mapstructure:"path"`
	Type            string `mapstructure:"type"`
	UnixPermissions int64  `mapstructure:"unix_permissions"`
}

// VolumesFilesResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type VolumesFilesResourceBodyDataModelONTAP struct {
	Name            string `mapstructure:"name,omitempty"`
	Path            string `mapstructure:"path"`
	ByteOffset      int64  `mapstructure:"byte_offset,omitempty"`
	Overwrite       bool   `mapstructure:"overwrite,omitempty"`
	Type            string `mapstructure:"type,omitempty"`
	UnixPermissions int64  `mapstructure:"unix_permissions,omitempty"`
	Size            int    `mapstructure:"size,omitempty"`
}

// QOSPolicy describes the resource data model.
type QOSPolicy struct {
	Name string `mapstructure:"name,omitempty"`
}

// VolumesFilesDataSourceFilterModel describes the data source data model for queries.
type VolumesFilesDataSourceFilterModel struct {
	Name    string `mapstructure:"name"`
	SVMName string `mapstructure:"svm.name"`
}

// GetStorageVolumesFiles to get storage_volumes_files info
func GetStorageVolumesFiles(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string, path string) ([]StorageVolumesFilesGetDataModelONTAP, error) {
	api := "storage/volumes/" + uuid + "/files/" + path
	query := r.NewQuery()
	query.Set("volume.uuid", uuid)
	query.Set("path", path)
	query.Fields([]string{"path", "name", "type", "volume", "fill_enabled", "size", "overwrite_enabled", "type", "group_id", "hard_links_count",
		"bytes_used", "owner_id", "inode_number", "is_empty", "target"})
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_volumes_filess info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []StorageVolumesFilesGetDataModelONTAP
	for _, info := range response {
		var record StorageVolumesFilesGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_volumes_filess data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateVolumesFiles to create volumes_files
func CreateVolumesFiles(errorHandler *utils.ErrorHandler, r restclient.RestClient, body VolumesFilesResourceBodyDataModelONTAP, volUUID string) (*VolumesFilesGetDataModelONTAP, error) {
	api := "storage/volumes/" + volUUID + "/files/" + body.Path
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding volumes_files body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating volumes_files", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP VolumesFilesGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding volumes_files info", fmt.Sprintf("error on decode storage/volumes_filess info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create volumes_files source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// UpdateVolumesFiles updates volumes_files
func UpdateVolumesFiles(errorHandler *utils.ErrorHandler, r restclient.RestClient, data VolumesFilesResourceBodyDataModelONTAP, volumeUUID string, path string) error {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		return errorHandler.MakeAndReportError("error encoding volumes_files body", fmt.Sprintf("error on encoding storage/volumes/%s/files/%s body: %s, body: %#v", volumeUUID, path, err, data))
	}
	query := r.NewQuery()

	api := fmt.Sprintf("storage/volumes/%s/files/%s", volumeUUID, path)
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError("error updating volumes_files", fmt.Sprintf("error on PATCH storage/volumes/%s/files/%s: %s, statusCode %d", volumeUUID, path, err, statusCode))
	}
	return nil
}

// DeleteVolumesFiles to delete volumes_files
func DeleteVolumesFiles(errorHandler *utils.ErrorHandler, r restclient.RestClient, path string, volUUID string) error {
	api := "storage/volumes/" + volUUID + "/files/" + path
	statusCode, _, err := r.CallDeleteMethod(api, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting volumes_files", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
