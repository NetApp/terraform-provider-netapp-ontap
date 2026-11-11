package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// StorageVolumeCloneGetDataModelONTAP describes the GET record data model using go types for mapping
type StorageVolumeCloneGetDataModelONTAP struct {
	SVM   svm         `mapstructure:"svm"`
	Name  string      `mapstructure:"name"`
	Type  string      `mapstructure:"type"`
	Clone CloneFields `mapstructure:"clone"`
	NAS   NASFields   `mapstructure:"nas"`
	ID    string      `mapstructure:"uuid"`
}

type CloneFields struct {
	IsFlexclone    bool  `mapstructure:"is_flexclone"`
	Split 		   bool  `mapstructure:"split,omitempty"`
	ParentVolume   Name  `mapstructure:"parent_volume,omitempty"`
	ParentSnapshot *Name `mapstructure:"parent_snapshot,omitempty"`
	ParentSVM      svm   `mapstructure:"parent_svm,omitempty"`
}

// NASFields describes the GET record data model using go types for mapping
type NASFields struct {
	JunctionPath *string `mapstructure:"path,omitempty"`
	GroupID      *int64  `mapstructure:"gid,omitempty"`
	UserID       *int64  `mapstructure:"uid,omitempty"`
}

// StorageVolumeCloneResourceBodyDataModel describes the resource body data model using go types for mapping
type StorageVolumeCloneResourceBodyDataModel struct {
	SVM   svm          `mapstructure:"svm,omitempty"`
	Name  *string      `mapstructure:"name,omitempty"`
	Type  *string      `mapstructure:"type,omitempty"`
	Clone *CloneFields `mapstructure:"clone,omitempty"`
	NAS   *NASFields   `mapstructure:"nas,omitempty"`
	ID    *string      `mapstructure:"uuid,omitempty"`
}

// Name describes the GET record and resource body data model using go types for mapping
type Name struct {
	Name string `mapstructure:"name"`
}

// StorageVolumeClonesFilterModel describes the data source filter model for storage volume clones
type StorageVolumeClonesFilterModel struct {
	SVMName string `mapstructure:"svm.name"`
	Name    string `mapstructure:"name"`
}

// GetVolumeClone to get volume info given its name
func GetVolumeClone(errorHandler *utils.ErrorHandler, r restclient.RestClient, svmName string, volumeName string, version versionModelONTAP) (*StorageVolumeCloneGetDataModelONTAP, error) {
	api := "storage/volumes"

	query := r.NewQuery()
	query.Set("svm.name", svmName)
	query.Set("name", volumeName)
	query.Fields([]string{
		"svm",
		"name",
		"type",
		"clone",
		"nas",
		"uuid",
	})

	statusCode, response, err := r.GetNilOrOneRecord(api, query, nil)
	if err == nil && response == nil {
		err = fmt.Errorf("no response for GET %s", api)
	}
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error reading volume info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP StorageVolumeCloneGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			fmt.Sprintf("failed to decode response from GET %s", api),
			fmt.Sprintf("error: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read volume_clone data source: %#v", dataONTAP))
	return &dataONTAP, nil
}

// GetVolumeClones to get a list of volume clones info for given filter options
func GetVolumeClones(errorHandler *utils.ErrorHandler, r restclient.RestClient, filter *StorageVolumeClonesFilterModel, version versionModelONTAP) ([]StorageVolumeCloneGetDataModelONTAP, error) {
	api := "storage/volumes"
	query := r.NewQuery()
	query.Set("clone.is_flexclone", "true")
	query.Fields([]string{
		"svm",
		"name",
		"type",
		"clone",
		"nas",
		"uuid",
	})

	if filter != nil {
		var filterMap map[string]interface{}
		if err := mapstructure.Decode(filter, &filterMap); err != nil {
			return nil, errorHandler.MakeAndReportError(
				"error encoding volume clones filter info",
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
			"error reading volume clones info",
			fmt.Sprintf("error on GET %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP []StorageVolumeCloneGetDataModelONTAP
	for _, info := range response {
		var record StorageVolumeCloneGetDataModelONTAP
		if err := mapstructure.Decode(info, &record); err != nil {
			return nil, errorHandler.MakeAndReportError(
				fmt.Sprintf("failed to decode response from GET %s", api),
				fmt.Sprintf("error: %s, statusCode %d, info %#v", err, statusCode, info),
			)
		}
		dataONTAP = append(dataONTAP, record)
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read volume_clones data source: %#v", dataONTAP))
	return dataONTAP, nil
}

// CreateVolumeClone to create a clone volume from a source volume/ snapshot
func CreateVolumeClone(errorHandler *utils.ErrorHandler, r restclient.RestClient, request StorageVolumeCloneResourceBodyDataModel) (*StorageVolumeCloneGetDataModelONTAP, error) {
	api := "storage/volumes"
	
	var body map[string]interface{}
	if err := mapstructure.Decode(request, &body); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error encoding storage/volumes body",
			fmt.Sprintf("error on encoding %s body: %s, request: %#v", api, err, request),
		)
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod(api, query, body)
	if err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error creating clone volume",
			fmt.Sprintf("error on POST %s: %s, statusCode %d", api, err, statusCode),
		)
	}

	var dataONTAP StorageVolumeCloneGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError(
			"error decoding clone volume info",
			fmt.Sprintf("error on decoding storage/volumes info: %s, statusCode %d, response %#v", err, statusCode, response),
		)
	}

	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Created volume clone resource: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteStorageVolume defined in storage_volume.go

// SplitVolumeClone splits a volume clone by sending clone.split: true in PATCH.
func SplitVolumeClone(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) error {
	api := fmt.Sprintf("storage/volumes/%s", uuid)

	body := map[string]interface{}{
		"clone": map[string]interface{}{
			"split_initiated": true,
		},
	}
	
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, _, err := r.CallUpdateMethod(api, query, body)
	if err != nil {
		return errorHandler.MakeAndReportError(
			"error splitting volume clone",
			fmt.Sprintf("error on PATCH %s: %s, statusCode %d", api, err, statusCode),
		)
	}
	return nil
}
