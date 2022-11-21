package interfaces

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
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
func GetStorageVolumeSnapshots(ctx context.Context, diags diag.Diagnostics, r restclient.RestClient, name string, uuid string) (*StorageVolumeSnapshotGetDataModelONTAP, error) {
	query := r.NewQuery()
	query.Add("name", name)
	query.Fields([]string{"name", "create_time", "expiry_time", "state", "size", "comment", "volume", "volume.uuid"})
	statusCode, response, err := r.GetNilOrOneRecord("storage/volumes/"+uuid+"/snapshots", query, nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Read storage/volumes/snapshots data - error: %s", err))
		// TODO: diags.Error is not reporting anything here.  Works in the caller.
		diags.AddError(err.Error(), fmt.Sprintf("statusCode %d, error %s", statusCode, err))
		return nil, err
	}

	if response == nil {
		tflog.Debug(ctx, fmt.Sprintf("snapshot %s not found for volume UUID %s", name, uuid))
		return nil, nil
	}

	var dataONTAP StorageVolumeSnapshotGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Read storage/volumes/snapshots data - decode error: %s, data: %#v", err, response))
		diags.AddError("failed to unmarshall response from GET storage/volumes/snapshots - UDATA", fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("Read storage/volumes/snapshots data source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// CreateStorageVolumeSnapshot to create a snapshot
func CreateStorageVolumeSnapshot(ctx context.Context, diags diag.Diagnostics, r restclient.RestClient, data StorageVolumeSnapshotResourceModel, uuid string) (*StorageVolumeSnapshotGetDataModelONTAP, error) {
	var body map[string]interface{}
	if err := mapstructure.Decode(data, &body); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Create snapshot - encode error: %s, data: %#v", err, data))
		return nil, err
	}
	query := r.NewQuery()
	query.Add("return_records", "true")
	statusCode, response, err := r.CallCreateMethod("storage/volumes/"+uuid+"/snapshots", query, body)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Create snapshot - error: %s", err))
		// TODO: diags.Error is not reporting anything here.  Works in the caller.
		diags.AddError(err.Error(), fmt.Sprintf("statusCode %d, error %s", statusCode, err))
		return nil, err
	}

	var dataONTAP StorageVolumeSnapshotGetDataModelONTAP
	if err := mapstructure.Decode(response, &dataONTAP); err != nil {
		tflog.Error(ctx, fmt.Sprintf("Create Snapshot - decode error: %s, data: %#v", err, response))
		diags.AddError("failed to unmarshall response from POST storage/volume/snapshot - UDATA", fmt.Sprintf("statusCode %d, response %#v", statusCode, response))
		return nil, err
	}
	tflog.Debug(ctx, fmt.Sprintf("Create volume source - udata: %#v", dataONTAP))
	return &dataONTAP, nil
}

// DeleteStorageVolumeSnapshot to delete a snapshot
func DeleteStorageVolumeSnapshot(ctx context.Context, diags diag.Diagnostics, r restclient.RestClient, volumeUUID string, uuid string) (*StorageVolumeSnapshotGetDataModelONTAP, error) {
	statusCode, _, err := r.CallDeleteMethod("storage/volumes/"+volumeUUID+"/snapshots/"+uuid, nil, nil)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Delete sanpshot - error: %s", err))
		// TODO: diags.Error is not reporting anything here.  Works in the caller.
		diags.AddError(err.Error(), fmt.Sprintf("statusCode %d, error %s", statusCode, err))
		return nil, err
	}
	return nil, nil
}
