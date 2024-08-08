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
