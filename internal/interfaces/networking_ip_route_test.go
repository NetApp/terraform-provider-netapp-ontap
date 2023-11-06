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

var ipRouteRecord = IPRouteGetDataModelONTAP{
	Destination: DestinationDataSourceModel{
		Address: "string",
		Netmask: "string",
	},
	UUID:    "string",
	Gateway: "string",
	Metric:  0,
	SVMName: svm{
		Name: "string",
	},
}

var badIPRouteRecord = struct{ Destination int }{123}

func TestGetIPRoute(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(ipRouteRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badIPRouteRecord, &badRecordInterface)
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
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/routes", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "network/ip/routes", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/routes", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/routes", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}

	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *IPRouteGetDataModelONTAP
		wantErr bool
		gen     int
		maj     int
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &ipRouteRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetIPRoute(errorHandler, *r, "destination", "svmName", "gateway", versionModelONTAP{Generation: tt.gen, Major: tt.maj})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIPRoute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetIPRoute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetListIPRoutes(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(ipRouteRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]interface{}
	err = mapstructure.Decode(badIPRouteRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecordsResponse := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	twoRecordsResponse := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	genericErrorResponse := errors.New("generic error for UT")
	decodeErrorResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]interface{}{badRecordInterface}}

	var wantOneRecord = []IPRouteGetDataModelONTAP{ipRouteRecord}
	var wantTwoRecords = []IPRouteGetDataModelONTAP{ipRouteRecord, ipRouteRecord}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/routes", StatusCode: 200, Response: noRecordsResponse, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/routes", StatusCode: 200, Response: oneRecordResponse, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/routes", StatusCode: 200, Response: twoRecordsResponse, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/routes", StatusCode: 200, Response: twoRecordsResponse, Err: genericErrorResponse},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/routes", StatusCode: 200, Response: decodeErrorResponse, Err: nil},
		},
	}

	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    []IPRouteGetDataModelONTAP
		wantErr bool
		gen     int
		maj     int
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: false},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: wantOneRecord, wantErr: false},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: wantTwoRecords, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetListIPRoutes(errorHandler, *r, "gateway", &IPRouteDataSourceFilterModel{}, versionModelONTAP{Generation: tt.gen, Major: tt.maj})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIPRoutes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}
