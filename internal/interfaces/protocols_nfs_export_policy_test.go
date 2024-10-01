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
var basicExportPolicyRecord = ExportPolicyGetDataModelONTAP{
	Name: "string",
	Svm:  "string",
}

// bad record
var badExportPolicyRecord = struct{ Name int }{123}

// create export policy with basic request body
var basicExportPolicyBody = ExportpolicyResourceBodyDataModelONTAP{
	Name: "string",
	Svm: SvmDataModelONTAP{
		Name: "string",
		UUID: "string",
	},
}

// create export policy with empty comment
var badExportPolicyBody = ExportpolicyResourceBodyDataModelONTAP{
	Name: "",
}

// update export policy name
var renameExportPolicyBody = ExportpolicyResourceBodyDataModelONTAP{
	Name: "newname",
	Svm: SvmDataModelONTAP{
		Name: "string",
		UUID: "string",
	},
}

// update export policy with basic request body
var updateExportPolicyErrorBody = ExportpolicyResourceBodyDataModelONTAP{
	Name: "string",
	Svm: SvmDataModelONTAP{
		Name: "newsvm",
		UUID: "string",
	},
}

func TestGetExportPolicy(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicExportPolicyRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badExportPolicyRecord, &badRecordInterface)
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
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_get_error_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		want      *ExportPolicyGetDataModelONTAP
		wantErr   bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &basicExportPolicyRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_get_error_1", responses: responses["test_get_error_1"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetExportPolicy(errorHandler, *r, "string")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetExportPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetExportPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateExportPolicy(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicExportPolicyRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badExportPolicyRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	onebasicExportPolicyRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicRecordInterface}}
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_create_basic_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: onebasicExportPolicyRecord, Err: nil},
		},
		"test_create_error_1": {
			{ExpectedMethod: "POST", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name        string
		responses   []restclient.MockResponse
		requestbody ExportpolicyResourceBodyDataModelONTAP
		want        *ExportPolicyGetDataModelONTAP
		wantErr     bool
	}{
		{name: "test_create_basic_record_1", responses: responses["test_create_basic_record_1"], requestbody: basicExportPolicyBody, want: &basicExportPolicyRecord, wantErr: false},
		{name: "test_create_error_1", responses: responses["test_create_error_1"], requestbody: badExportPolicyBody, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := CreateExportPolicy(errorHandler, *r, tt.requestbody)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateExportPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateExportPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteExportPolicy(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_delete_1": {
			{ExpectedMethod: "DELETE", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_delete_error_1": {
			{ExpectedMethod: "DELETE", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: noRecords, Err: genericError},
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
			err2 := DeleteExportPolicy(errorHandler, *r, "string")
			if err2 != nil {
				fmt.Printf("err2: %s\n", err)
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("DeleteExportPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUpdateExportPolicy(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_update_rename_export_policy": {
			{ExpectedMethod: "PATCH", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_update_error_1": {
			{ExpectedMethod: "PATCH", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: noRecords, Err: genericError},
		},
	}
	tests := []struct {
		name        string
		responses   []restclient.MockResponse
		requestbody ExportpolicyResourceBodyDataModelONTAP
		wantErr     bool
	}{
		{name: "test_update_rename_export_policy", responses: responses["test_update_rename_export_policy"], requestbody: renameExportPolicyBody, wantErr: false},
		{name: "test_update_error_1", responses: responses["test_update_error_1"], requestbody: updateExportPolicyErrorBody, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			err = UpdateExportPolicy(errorHandler, *r, tt.requestbody, "string")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateExportPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetExportPoliciesList(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	badRecord := struct{ Name int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(basicExportPolicyBody, &recordInterface)
	if err != nil {
		panic(err)
	}

	var badRecordInterface map[string]any
	err = mapstructure.Decode(badRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecordsResponse := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	twoRecordsResponse := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	badRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}

	var wantOneRecord = []ExportPolicyGetDataModelONTAP{basicExportPolicyRecord}
	var wantTwoRecords = []ExportPolicyGetDataModelONTAP{basicExportPolicyRecord, basicExportPolicyRecord}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: noRecordsResponse, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: oneRecordResponse, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: twoRecordsResponse, Err: nil},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    []ExportPolicyGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: false},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: wantOneRecord, wantErr: false},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: wantTwoRecords, wantErr: false},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetExportPoliciesList(errorHandler, *r, &ExportPolicyGetDataFilterModel{Name: ""})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetExportPoliciesList() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetExportPoliciesList() = %v, want %v", got, tt.want)
			}
		})
	}
}
