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
var basicStorageAggregateRecord = StorageAggregateGetDataModelONTAP{
	Name: "string",
	UUID: "string",
	Node: StorageAggregateNode{
		Name: "string",
	},
	BlockStorage: AggregateBlockStorage{
		Primary: AggregateBlockStoragePrimary{
			DiskClass: "string",
			DiskCount: 5,
			RaidSize:  16,
			RaidType:  "string",
		},
		Mirror: AggregateBlockStorageMirror{
			Enabled: false,
		},
	},
	DataEncryption: AggregateDataEncryption{
		SoftwareEncryptionEnabled: false,
	},
	SnaplockType: "non_snaplock",
	State:        "online",
}

// bad record
var badStorageAggregateRecord = struct{ Name int }{123}

// only requried parameters
var basicStorageAggregateCreateRecord = StorageAggregateResourceModel{
	Name: "string",
	Node: map[string]string{
		"name": "string",
	},
	BlockStorage: map[string]any{
		"disk_count": 5,
	},
}

// bed decode
var badStorageAggregateBody = StorageAggregateResourceModel{
	Name: "string",
}

func TestGetStorageAggregate(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicStorageAggregateRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badStorageAggregateRecord, &badRecordInterface)
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
			{ExpectedMethod: "GET", ExpectedURL: "storage/aggregates", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/aggregates", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/aggregates", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_error_1": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/aggregates", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		want      *StorageAggregateGetDataModelONTAP
		wantErr   bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &basicStorageAggregateRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_error_1", responses: responses["test_error_1"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetStorageAggregate(errorHandler, *r, "string")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStorageAggregate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStorageAggregate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateStorageAggregate(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicStorageAggregateRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badStorageAggregateRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}

	onebasicStorageAggregateRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicRecordInterface}}
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_create_basic_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "storage/aggregates", StatusCode: 200, Response: onebasicStorageAggregateRecord, Err: nil},
		},
		"test_error_1": {
			{ExpectedMethod: "POST", ExpectedURL: "storage/aggregates", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name        string
		responses   []restclient.MockResponse
		requestbody StorageAggregateResourceModel
		want        *StorageAggregateGetDataModelONTAP
		wantErr     bool
	}{
		{name: "test_create_basic_record_1", responses: responses["test_create_basic_record_1"], requestbody: basicStorageAggregateCreateRecord, want: &basicStorageAggregateRecord, wantErr: false},
		{name: "test_error_1", responses: responses["test_error_1"], requestbody: badStorageAggregateBody, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := CreateStorageAggregate(errorHandler, *r, tt.requestbody, 0)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateStorageAggregate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateStorageAggregate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteStorageAggregate(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(basicStorageAggregateRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badStorageAggregateRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_delete": {
			{ExpectedMethod: "DELETE", ExpectedURL: "storage/aggregates/1234", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_error_2": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/aggregates/1234", StatusCode: 200, Response: noRecords, Err: genericError},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		wantErr   bool
	}{
		{name: "test_delete", responses: responses["test_delete"], wantErr: false},
		{name: "test_error_2", responses: responses["test_error_2"], wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			err2 := DeleteStorageAggregate(errorHandler, *r, "1234")
			if err2 != nil {
				fmt.Printf("err2: %s\n", err)
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("DeleteStorageAggregate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
