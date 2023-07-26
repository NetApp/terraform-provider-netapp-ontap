package interfaces

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
)

// basic get data record
var basicStorageVolumeSnapshotRecord = StorageVolumeSnapshotGetDataModelONTAP{
	Name: "string",
	UUID: "string",
	Volume: NameDataModel{
		UUID: "string",
		Name: "string",
	},
	CreateTime: "MM/DD/YYYY HH:MM:SS",
	State:      "string",
	Size:       122880,
}

var fullStorageVolumeSnapshotRecord = StorageVolumeSnapshotGetDataModelONTAP{
	Name: "string",
	UUID: "string",
	Volume: NameDataModel{
		UUID: "string",
		Name: "string",
	},
	CreateTime:         "MM/DD/YYYY HH:MM:SS",
	ExpiryTime:         "MM/DD/YYYY HH:MM:SS",
	SnaplockExpiryTime: "MM/DD/YYYY HH:MM:SS",
	State:              "string",
	Size:               122880,
	Comment:            "string",
	SnapmirrorLabel:    "string",
}

// bad record
var badStorageVolumeSnapshotRecord = struct{ Name int }{123}

// create snapshot with basic request body
var basicStorageVolumeSnapshotBody = StorageVolumeSnapshotResourceModel{
	Name: "string",
}

// create snapshot with full request body
var fullStorageVolumeSnapshotBody = StorageVolumeSnapshotResourceModel{
	Name:               "string",
	ExpiryTime:         "MM/DD/YYYY HH:MM:SS",
	SnaplockExpiryTime: "MM/DD/YYYY HH:MM:SS",
	Comment:            "string",
	SnapmirrorLabel:    "string",
}

// create snapshot with empty comment
var badStorageVolumeSnapshotBody = StorageVolumeSnapshotResourceModel{
	Comment: "",
}

// update snapshot with new name
var renameStorageVolumeSnapshotBody = StorageVolumeSnapshotResourceModel{
	Name: "newname",
}

// update snapshot comment
var updateStorageVolumeSnapshotCommentBody = StorageVolumeSnapshotResourceModel{
	Comment: "new comment",
}

// update snapshot with wrong values
var updateStorageVolumeSnapshotErrorBody = StorageVolumeSnapshotResourceModel{
	ExpiryTime: "",
	Comment:    "",
}

func TestGetStorageVolumeSnapshot(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicStorageVolumeSnapshotRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badStorageVolumeSnapshotRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicRecordInterface}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{basicRecordInterface, basicRecordInterface}}
	genericError := errors.New("generic error for UT")
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/volumes/1234/snapshots/5678", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/volumes/1234/snapshots/5678", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/volumes/1234/snapshots/5678", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_get_error_1": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/volumes/1234/snapshots/5678", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		want      *StorageVolumeSnapshotGetDataModelONTAP
		wantErr   bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &basicStorageVolumeSnapshotRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_get_error_1", responses: responses["test_get_error_1"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetStorageVolumeSnapshot(errorHandler, *r, "string", "string")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStorageVolumeSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStorageVolumeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateStorageVolumeSnapshot(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicStorageVolumeSnapshotRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var fullRecordInterface map[string]any
	err = mapstructure.Decode(fullStorageVolumeSnapshotRecord, &fullRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badStorageVolumeSnapshotRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	onebasicStorageVolumeSnapshotRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicRecordInterface}}
	onefullStorageVolumeSnapshotRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{fullRecordInterface}}
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_create_basic_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "storage/volumes/1234/snapshots", StatusCode: 200, Response: onebasicStorageVolumeSnapshotRecord, Err: nil},
		},
		"test_create_full_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "storage/volumes/1234/snapshots", StatusCode: 200, Response: onefullStorageVolumeSnapshotRecord, Err: nil},
		},
		"test_create_error_1": {
			{ExpectedMethod: "POST", ExpectedURL: "storage/volumes/1234/snapshots", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name        string
		responses   []restclient.MockResponse
		requestbody StorageVolumeSnapshotResourceModel
		want        *StorageVolumeSnapshotGetDataModelONTAP
		wantErr     bool
	}{
		{name: "test_create_basic_record_1", responses: responses["test_create_basic_record_1"], requestbody: basicStorageVolumeSnapshotBody, want: &basicStorageVolumeSnapshotRecord, wantErr: false},
		{name: "test_create_full_record_1", responses: responses["test_create_full_record_1"], requestbody: fullStorageVolumeSnapshotBody, want: &fullStorageVolumeSnapshotRecord, wantErr: false},
		{name: "test_create_error_1", responses: responses["test_create_error_1"], requestbody: badStorageVolumeSnapshotBody, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := CreateStorageVolumeSnapshot(errorHandler, *r, tt.requestbody, "string")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateStorageVolumeSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateStorageVolumeSnapshot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteStorageVolumeSnapshot(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_delete_1": {
			{ExpectedMethod: "DELETE", ExpectedURL: "storage/volumes/1234/snapshots/5678", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_delete_error_1": {
			{ExpectedMethod: "DELETE", ExpectedURL: "storage/volumes/1234/snapshots/5678", StatusCode: 200, Response: noRecords, Err: genericError},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		wantErr   bool
	}{
		{name: "test_delete_1", responses: responses["test_delete_1"], wantErr: false},
		{name: "test_delete_error_1", responses: responses["test_delete_error_1"], wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			err2 := DeleteStorageVolumeSnapshot(errorHandler, *r, "string", "string")
			if err2 != nil {
				fmt.Printf("err2: %s\n", err)
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("DeleteStorageVolumeSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUpdateStorageVolumeSnapshot(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_update_rename_snapshot": {
			{ExpectedMethod: "PATCH", ExpectedURL: "storage/volumes/1234/snapshots/5678", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_update_comment_snapshot": {
			{ExpectedMethod: "PATCH", ExpectedURL: "storage/volumes/1234/snapshots/5678", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_update_error_1": {
			{ExpectedMethod: "PATCH", ExpectedURL: "storage/volumes/1234/snapshots/5678", StatusCode: 200, Response: noRecords, Err: genericError},
		},
	}
	tests := []struct {
		name        string
		responses   []restclient.MockResponse
		requestbody StorageVolumeSnapshotResourceModel
		wantErr     bool
	}{
		{name: "test_update_rename_snapshot", responses: responses["test_update_rename_snapshot"], requestbody: renameStorageVolumeSnapshotBody, wantErr: false},
		{name: "test_update_comment_snapshot", responses: responses["test_update_comment_snapshot"], requestbody: updateStorageVolumeSnapshotCommentBody, wantErr: false},
		{name: "test_update_error_1", responses: responses["test_update_error_1"], requestbody: updateStorageVolumeSnapshotErrorBody, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			err = UpdateStorageVolumeSnapshot(errorHandler, *r, tt.requestbody, "string", "string")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateStorageVolumeSnapshot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
