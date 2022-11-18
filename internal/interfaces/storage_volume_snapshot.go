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
		tflog.Debug(ctx, fmt.Sprintf("No Sanpshots found carchi"))
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
