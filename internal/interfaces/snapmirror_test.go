package interfaces

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
	"reflect"
	"testing"
)

var snapmirrorRecord = SnapmirrorDataSourceModel{
	Source: Source{
		Cluster: SnapmirrorCluster{
			Name: "string",
			UUID: "string",
		},
		Path: "string",
		Svm: SvmDataModelONTAP{
			Name: "string",
			UUID: "string",
		},
	},
	Destination: Destination{
		Path: "string",
		Svm: SvmDataModelONTAP{
			Name: "string",
			UUID: "string",
		},
	},
	Healthy: false,
	Restore: false,
	UUID:    "string",
	State:   "string",
	Policy: SnapmirrorPolicy{
		UUID: "string",
	},
}

var record911Snapmirror = SnapmirrorDataSourceModel{
	Source: Source{
		Cluster: SnapmirrorCluster{
			Name: "string",
			UUID: "string",
		},
		Path: "string",
		Svm: SvmDataModelONTAP{
			Name: "string",
			UUID: "string",
		},
	},
	Destination: Destination{
		Path: "string",
		Svm: SvmDataModelONTAP{
			Name: "string",
			UUID: "string",
		},
	},
	Healthy: false,
	Restore: false,
	UUID:    "string",
	State:   "string",
	Policy: SnapmirrorPolicy{
		UUID: "string",
	},
	GroupType: "string",
	Throttle:  0,
}

var badRecordSnapmirror = struct{ Healthy int }{123}

func TestGetSnapmirrorByDestinationPath(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(snapmirrorRecord, &recordInterface)
	if err != nil {
		panic(err)
	}

	var record911Interface map[string]any
	err = mapstructure.Decode(record911Snapmirror, &record911Interface)
	if err != nil {
		panic(err)
	}

	var badRecordInterface map[string]any
	err = mapstructure.Decode(badRecordSnapmirror, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecordsResponse := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	oneRecord911Response := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{record911Interface}}
	badRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/relationships", StatusCode: 200, Response: noRecordsResponse, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/relationships", StatusCode: 200, Response: oneRecordResponse, Err: nil},
		},
		"test_one_911_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/relationships", StatusCode: 200, Response: oneRecord911Response, Err: nil},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/relationships", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *SnapmirrorDataSourceModel
		wantErr bool
		gen     int
		maj     int
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true, gen: 9, maj: 11},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &snapmirrorRecord, wantErr: false, gen: 9, maj: 11},
		{name: "test_one_911_record_1", responses: responses["test_one_911_record_1"], want: &record911Snapmirror, wantErr: false, gen: 9, maj: 10},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true, gen: 9, maj: 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetSnapmirrorByDestinationPath(errorHandler, *r, "", versionModelONTAP{Generation: tt.gen, Major: tt.maj})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSnapmirrors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSnapmirrors() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSnapmirrors(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(snapmirrorRecord, &recordInterface)
	if err != nil {
		panic(err)
	}

	var record911Interface map[string]any
	err = mapstructure.Decode(record911Snapmirror, &record911Interface)
	if err != nil {
		panic(err)
	}

	var badRecordInterface map[string]any
	err = mapstructure.Decode(badRecordSnapmirror, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecordsResponse := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	oneRecord911Response := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{record911Interface}}
	twoRecordsResponse := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	badRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}

	var wantOneRecord = []SnapmirrorDataSourceModel{snapmirrorRecord}
	var wantTwoRecords = []SnapmirrorDataSourceModel{snapmirrorRecord, snapmirrorRecord}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/relationships", StatusCode: 200, Response: noRecordsResponse, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/relationships", StatusCode: 200, Response: oneRecordResponse, Err: nil},
		},
		"test_one_911_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/relationships", StatusCode: 200, Response: oneRecord911Response, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/relationships", StatusCode: 200, Response: twoRecordsResponse, Err: nil},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/relationships", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    []SnapmirrorDataSourceModel
		wantErr bool
		gen     int
		maj     int
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: false, gen: 9, maj: 11},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: wantOneRecord, wantErr: false, gen: 9, maj: 11},
		{name: "test_one_911_record_1", responses: responses["test_one_911_record_1"], want: []SnapmirrorDataSourceModel{record911Snapmirror}, wantErr: false, gen: 9, maj: 10},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: wantTwoRecords, wantErr: false, gen: 9, maj: 11},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true, gen: 9, maj: 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetSnapmirrors(errorHandler, *r, &SnapmirrorFilterModel{}, versionModelONTAP{Generation: tt.gen, Major: tt.maj})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSnapmirrors() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSnapmirrors() = %v, want %v", got, tt.want)
			}
		})
	}
}
