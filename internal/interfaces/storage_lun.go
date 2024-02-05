package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// StorageLunGetDataModelONTAP describes the GET record data model using go types for mapping.
type StorageLunGetDataModelONTAP struct {
	Name       string       `mapstructure:"name"`
	UUID       string       `mapstructure:"uuid"`
	SVM        svm          `mapstructure:"svm,omitempty"`
	CreateTime string       `mapstructure:"create_time,omitempty"`
	Location   LunLocation  `mapstructure:"location,omitempty"`
	OSType     string       `mapstructure:"os_type,omitempty"`
	QoSPolicy  LunQoSPolicy `mapstructure:"qos_policy,omitempty"`
	Space      LunSpace     `mapstructure:"space,omitempty"`
}

// LunLocation describes the data model for location.
type LunLocation struct {
	LogicalUnit string    `mapstructure:"logical_unit,omitempty"`
	Volume      LunVolume `mapstructure:"volume,omitempty"`
}

// LunVolume describes the data model for volume.
type LunVolume struct {
	Name string `mapstructure:"name,omitempty"`
	UUID string `mapstructure:"uuid,omitempty"`
}

// LunQoSPolicy describes the data model for QoS policy.
type LunQoSPolicy struct {
	Name string `mapstructure:"name,omitempty"`
	UUID string `mapstructure:"uuid,omitempty"`
}

// LunSpace describes the data model for space.
type LunSpace struct {
	Size int64 `mapstructure:"size,omitempty"`
	Used int64 `mapstructure:"used,omitempty"`
}

// StorageLunResourceBodyDataModelONTAP describes the body data model using go types for mapping.
type StorageLunResourceBodyDataModelONTAP struct {
	SVM       svm      `mapstructure:"svm"`
	Locations location `mapstructure:"location"`
	OsType    string   `mapstructure:"os_type"`
	Space     space    `mapstructure:"space"`
	QosPolicy string   `mapstructure:"qos_policy"`
}

type location struct {
	Volume      volume `mapstructure:"volume"`
	LogicalUnit string `mapstructure:"logical_unit"`
}

type volume struct {
	Name string `mapstructure:"name"`
}

type space struct {
	Size int64 `mapstructure:"size"`
}

// StorageLunDataSourceFilterModel describes the data source data model for queries.
type StorageLunDataSourceFilterModel struct {
	Name       string `mapstructure:"name"`
	SVMName    string `mapstructure:"svm.name"`
	VolumeName string `mapstructure:"location.volume.name"`
}

// GetStorageLunByName to get storage_lun info
func GetStorageLunByName(errorHandler *utils.ErrorHandler, r restclient.RestClient, name string, svmName string, volumeName string) (*StorageLunGetDataModelONTAP, error) {
	api := "storage/luns"
	query := r.NewQuery()
	query.Set("location.logical_unit", name)
	query.Set("svm.name", svmName)
	query.Set("location.volume.name", volumeName)
	query.Fields([]string{"name", "svm.name", "create_time", "location", "os_type", "qos_policy", "space", "uuid"})
	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_lun info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageLunGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_lun data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetStorageLuns to get storage_lun info for all resources matching a filter
func GetStorageLuns(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *StorageLunDataSourceFilterModel) ([]StorageLunGetDataModelONTAP, error) {
	api := "storage/luns"
	query := r.NewQuery()
	query.Fields([]string{"name", "svm.name", "create_time", "location", "os_type", "qos_policy", "space", "uuid"})
	if filter != nil {
		if filter.Name != "" {
			query.Add("name", filter.Name)
		}
		if filter.SVMName != "" {
			query.Add("svm.name", filter.SVMName)
		}
		if filter.VolumeName != "" {
			query.Add("location.volume.name", filter.VolumeName)
		}
	}
	statusCode, response, err := r.GetZeroOrMoreRecords(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading storage_luns info", fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP []StorageLunGetDataModelONTAP
	for _, info := range response {
		var record StorageLunGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info))
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read storage_luns data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateStorageLun to create storage_lun
func CreateStorageLun(errorHandler *utils.ErrorHandler, r restclient.RestClient, body StorageLunResourceBodyDataModelONTAP) (*StorageLunGetDataModelONTAP, error) {
	api := "storage/luns"
	var bodyMap map[string]interface{}
	if err := mapstructure.Decode(body, &bodyMap); err != nil {
		return nil, errorHandler.MakeAndReportError("error encoding storage_lun body", fmt.Sprintf("error on encoding %s body: %s, body: %#v", api, err, body))
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create storage_lun source - body: %#v", bodyMap))
	statusCode, response, err := r.CallCreateMethod(api, query, bodyMap)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error creating storage_lun", fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode))
	}

	var dataONTAP StorageLunGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding storage_lun info", fmt.Sprintf("error on decode storage/storage_luns info: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Create storage_lun source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteStorageLun to delete storage_lun
func DeleteStorageLun(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := "storage/luns"
	statusCode, _, err := r.CallDeleteMethod(api+"/"+uuid, nil, nil)
	if err != nil {
		return errorHandler.MakeAndReportError("error deleting storage_lun", fmt.Sprintf("error on DELETE %s: %s, statusCode %d", api, err, statusCode))
	}
	return nil
}
