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

func TestGetCluster(t *testing.T) {

	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	record := ClusterGetDataModelONTAP{
		Name: "cluster1",
		Version: versionModelONTAP{
			Full: "ONTAP 1.2.3",
		},
	}
	badRecord := struct{ Name int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(record, &recordInterface)
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
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster", StatusCode: 200, Response: twoRecords, Err: nil},
		},
		"test_error_1": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_error_2": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *ClusterGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &record, wantErr: false},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: nil, wantErr: true},
		{name: "test_error_1", responses: responses["test_error_1"], want: nil, wantErr: true},
		{name: "test_error_2", responses: responses["test_error_2"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetCluster(errorHandler, *r)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetClusterNodes(t *testing.T) {

	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	record := ClusterNodeGetDataModelONTAP{
		Name: "cluster1-01",
		ManagementInterfaces: []noddMgmtInterface{
			{ipAddress{"10.11.12.13"}},
			{ipAddress{"10.11.12.14"}},
		},
	}
	var recordInterface, badRecordInterface map[string]any
	err := mapstructure.Decode(record, &recordInterface)
	if err != nil {
		panic(err)
	}
	badRecord := struct{ Name int }{123}
	err = mapstructure.Decode(badRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	genericError := errors.New("generic error for UT")
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}

	expectedOneRecord := []ClusterNodeGetDataModelONTAP{record}
	expectedTwoRecords := []ClusterNodeGetDataModelONTAP{record, record}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster", StatusCode: 200, Response: twoRecords, Err: nil},
		},
		"test_error_1": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_error_2": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    []ClusterNodeGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: false},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: expectedOneRecord, wantErr: false},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: expectedTwoRecords, wantErr: false},
		{name: "test_error_1", responses: responses["test_error_1"], want: nil, wantErr: true},
		{name: "test_error_2", responses: responses["test_error_2"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if err != nil {
				panic(err)
			}
			got, err := GetClusterNodes(errorHandler, *r)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}
