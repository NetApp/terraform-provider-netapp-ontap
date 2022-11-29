package interfaces

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// StorageVolumeGetDataModelONTAP describes the GET record data model using go types for mapping.
type StorageVolumeGetDataModelONTAP struct {
	Name       string
	Vserver    string
	Aggregates []Aggregate
	UUID       string
}

// StorageVolumeResourceModel describes the resource data model.
type StorageVolumeResourceModel struct {
	Name       string              `mapstructure:"name"`
	SVM        Vserver             `mapstructure:"svm"`
	Aggregates []map[string]string `mapstructure:"aggregates"`
}

// Aggregate describes the resource data model.
type Aggregate struct {
	Name string `mapstructure:"name"`
}

// Vserver describes the resource data model.
type Vserver struct {
	Name string `mapstructure:"name"`
}

// GetStorageVolume to get volume info by uuid
func GetStorageVolume(errorHandler *utils.ErrorHandler, r restclient.RestClient, uuid string) (*StorageVolumeGetDataModelONTAP, error) {
	statusCode, response, err := r.GetNilOrOneRecord("storage/volumes/"+uuid, nil, nil)
	if err != nil {
		return nil, errorHandler.MakeAndReportError("error reading volume info", fmt.Sprintf("error on GET storage/volumes: %s", err))
	}

	var dataONTAP *StorageVolumeGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		return nil, errorHandler.MakeAndReportError("error decoding volume info", fmt.Sprintf("error on decode storage/volumes: %s, statusCode %d, response %#v", err, statusCode, response))
	}
	tflog.Debug(errorHandler.Ctx, fmt.Sprintf("Read volume source - udata: %#v", dataONTAP))
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
