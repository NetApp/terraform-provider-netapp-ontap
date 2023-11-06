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

var svmRecord = SvmGetDataSourceModel{
	Name: "string",
	UUID: "string",
	Ipspace: Ipspace{
		Name: "string",
	},
	SnapshotPolicy: SnapshotPolicy{
		Name: "string",
	},
	SubType:  "string",
	Comment:  "string",
	Language: "string",
	Aggregates: []Aggregate{
		{
			Name: "string",
		},
	},
	MaxVolumes: "string",
}

var svmSecondRecord = SvmGetDataSourceModel{
	Name: "string",
	UUID: "string",
	Ipspace: Ipspace{
		Name: "string",
	},
	SnapshotPolicy: SnapshotPolicy{
		Name: "string",
	},
	SubType:  "string",
	Comment:  "string",
	Language: "string",
	Aggregates: []Aggregate{
		{
			Name: "string",
		},
	},
	MaxVolumes: "string",
}

func TestGetSvmByNameDataSource(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	badRecord := struct{ Comment int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(svmRecord, &recordInterface)
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
	genericError := errors.New("generic error for UT")
	badRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "svm", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "svm", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "svm", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "svm/svms", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *SvmGetDataSourceModel
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &svmRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetSvmByNameDataSource(errorHandler, *r, "svmname")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSvmByNameDataSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSvmByNameDataSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSvmsByName(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	badRecord := struct{ Comment int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(svmRecord, &recordInterface)
	if err != nil {
		panic(err)
	}

	var secondRecordInterface map[string]any
	err = mapstructure.Decode(svmSecondRecord, &secondRecordInterface)
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
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, secondRecordInterface}}
	//genericError := errors.New("generic error for UT")
	badRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}

	var cvmOneRecord = []SvmGetDataSourceModel{svmRecord}
	var cvmTwoRecords = []SvmGetDataSourceModel{svmRecord, svmSecondRecord}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "svm", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "svm", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "svm", StatusCode: 200, Response: twoRecords, Err: nil},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "svm", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    []SvmGetDataSourceModel
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: false},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: cvmOneRecord, wantErr: false},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: cvmTwoRecords, wantErr: false},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetSvmsByName(errorHandler, *r, &SvmDataSourceFilterModel{Name: ""})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSvmsByNameDataSource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSvmsByNameDataSource() = %v, want %v", got, tt.want)
			}
		})
	}
}
