package interfaces

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
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
func GetStorageVolume(ctx context.Context, diags diag.Diagnostics, r restclient.RestClient, uuid string) (*StorageVolumeGetDataModelONTAP, error) {
	statusCode, response, err := r.GetNilOrOneRecord("storage/volumes/"+uuid, nil, nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Read volume data - error: %s", err))
		// TODO: diags.Error is not reporting anything here.  Works in the caller.
		diags.AddError(err.Error(), fmt.Sprintf("statusCode %d, error %s", statusCode, err))
		return nil, err
	}

	var dataONTAP *StorageVolumeGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Read volume data - decode error: %s, data: %#v", err, response))
		diags.AddError("failed to unmarshall response from GET storage/volume/ - UDATA", fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("Read volume source - udata: %#v", dataONTAP))
	return dataONTAP, nil

}

// CreateStorageVolume to create volume
func CreateStorageVolume(ctx context.Context, diags diag.Diagnostics, r restclient.RestClient, data StorageVolumeResourceModel) (*StorageVolumeGetDataModelONTAP, error) {
	var volumeData map[string]interface{}
	if err := mapstructure.Decode(data, &volumeData); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Create volume - encode error: %s, data: %#v", err, data))
		return nil, err
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod("storage/volumes", query, volumeData)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Create volume - error: %s", err))
		// TODO: diags.Error is not reporting anything here.  Works in the caller.
		diags.AddError(err.Error(), fmt.Sprintf("statusCode %d, error %s", statusCode, err))
		return nil, err
	}

	var dataONTAP StorageVolumeGetDataModelONTAP
	if err := mapstructure.Decode(response.Records[0], &dataONTAP); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Create volume - decode error: %s, data: %#v", err, response))
		diags.AddError("failed to unmarshall response from POST storage/volume/ - UDATA", fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("Create volume source - udata: %#v", dataONTAP))
	return &dataONTAP, nil

}

// DeleteStorageVolume to delete volume
func DeleteStorageVolume(ctx context.Context, diags diag.Diagnostics, r restclient.RestClient, uuid string) error {
	statusCode, _, err := r.CallDeleteMethod("storage/volumes/"+uuid, nil, nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Delete volume - error: %s", err))
		// TODO: diags.Error is not reporting anything here.  Works in the caller.
		diags.AddError(err.Error(), fmt.Sprintf("statusCode %d, error %s", statusCode, err))
		return err
	}

	return nil

}
