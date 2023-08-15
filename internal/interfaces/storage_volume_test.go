package interfaces

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
	"reflect"
	"testing"
)

var basicStorageVolumeRecord = StorageVolumeGetDataModelONTAP{
	Name: "string",
	SVM: svm{
		Name: "string",
	},
	Aggregates: nil,
	UUID:       "string",
	Space: Space{
		Size: 0,
		Snapshot: Snapshot{
			ReservePercent: 0,
		},
		LogicalSpace: LogicalSpace{
			Enforcement: false,
			Reporting:   false,
		},
	},
	State: "string",
	Type:  "string",
	NAS: NASData{
		ExportPolicy: ExportPolicy{
			Name: "string",
		},
		JunctionPath:    "string",
		SecurityStyle:   "string",
		UnixPermissions: 0,
		GroupID:         0,
		UserID:          0,
	},
	SpaceGuarantee: Guarantee{
		Type: "string",
	},
	PercentSnapshotSpace: Snaplock{
		Type: "string",
	},
	Encryption: Encryption{
		Enabled: false,
	},
	Efficiency: Efficiency{
		Policy: Policy{
			Name: "string",
		},
		Compression: "string",
	},
	SnapshotPolicy: SnapshotPolicy{
		Name: "string",
	},
	Language: "string",
	QOS: QOS{
		Policy: Policy{
			Name: "string",
		},
	},
	TieringPolicy: TieringPolicy{
		Policy:         "string",
		MinCoolingDays: 0,
	},
	Comment: "string",
	Snaplock: Snaplock{
		Type: "string",
	},
	Analytics: Analytics{
		State: "string",
	},
}

// bad record
var badStorageVolumeRecord = struct{ Name int }{123}

func TestGetStorageVolumeByName(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(basicStorageVolumeRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badStorageVolumeRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	genericError := errors.New("generic error for UT")
	badRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "/storage/volumes", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "/storage/volumes", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "/storage/volumes", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "/storage/volumes", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *StorageVolumeGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &basicStorageVolumeRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetStorageVolumeByName(errorHandler, *r, "name", "svm")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStorageVolumeByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStorageVolumeByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetStorageVolumes(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(basicStorageVolumeRecord, &recordInterface)
	if err != nil {
		panic(err)
	}

	var badRecordInterface map[string]any
	err = mapstructure.Decode(badRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	//genericError := errors.New("generic error for UT")
	badRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}

	var storageVolumeOneRecord = []StorageVolumeGetDataModelONTAP{basicStorageVolumeRecord}
	var storageVolumeTwoRecords = []StorageVolumeGetDataModelONTAP{basicStorageVolumeRecord, basicStorageVolumeRecord}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "/storage/volumes", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "/storage/volumes", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "/storage/volumes", StatusCode: 200, Response: twoRecords, Err: nil},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "/storage/volumes", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    []StorageVolumeGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: false},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: storageVolumeOneRecord, wantErr: false},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: storageVolumeTwoRecords, wantErr: false},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetStorageVolumes(errorHandler, *r, &StorageVolumeDataSourceFilterModel{Name: ""})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStorageVolumes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStorageVolumes() = %v, want %v", got, tt.want)
			}
		})
	}
}
