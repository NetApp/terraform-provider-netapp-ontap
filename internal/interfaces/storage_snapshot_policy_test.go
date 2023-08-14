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

// Only requried parameters
var basicSnapshotPolicyRecord = SnapshotPolicyGetDataModelONTAP{
	Name: "string",
	UUID: "string",
	Copies: []CopyType{
		{
			Count: 1,
			Schedule: Schedule{
				Name: "Daily",
			},
		},
	},
	Comment: "string",
	Enabled: true,
}

// two copies snapshot policy
var twoSnapshotPolicyCopiesRecord = SnapshotPolicyGetDataModelONTAP{
	Name: "string",
	UUID: "string",
	Copies: []CopyType{
		{
			Count: 1,
			Schedule: Schedule{
				Name: "Weekly",
			},
		},
		{
			Count: 3,
			Schedule: Schedule{
				Name: "Monthly",
			},
		},
	},
	Comment: "string",
	Enabled: true,
}

// set enabled to false
var notEnabledSnapshotPolicyRecord = SnapshotPolicyGetDataModelONTAP{
	Name: "string",
	UUID: "string",
	Copies: []CopyType{
		{
			Count: 1,
			Schedule: Schedule{
				Name: "Hourly",
			},
		},
	},
	Comment: "string",
	Enabled: false,
}

// bad decode
var badSnapshotPolicyRecord = struct{ Name int }{123}

var badSnapshotPolicyBody = SnapshotPolicyResourceBodyDataModelONTAP{
	Name: "string",
}

// create with requried parameters
var basicSnapshotPolicyBody = SnapshotPolicyResourceBodyDataModelONTAP{
	Name: "string",
	SVM: svm{
		Name: "string",
	},
	Copies: []map[string]any{
		{
			"count": 1,
			"schedule": map[string]any{
				"name": "Daily",
			},
		},
	},
}

// create with two copies snapshot policy
var twoSnapshotPolicyCopiesBody = SnapshotPolicyResourceBodyDataModelONTAP{
	Name: "string",
	SVM: svm{
		Name: "string",
	},
	Copies: []map[string]any{
		{
			"count": 1,
			"schedule": map[string]any{
				"name": "Weekly",
			},
		},
		{
			"count": 3,
			"schedule": map[string]any{
				"name": "Monthly",
			},
		},
	},
}

// create with not enabled parameters
var notEnabledSnapshotPolicyBody = SnapshotPolicyResourceBodyDataModelONTAP{
	Name: "string",
	SVM: svm{
		Name: "string",
	},
	Copies: []map[string]any{
		{
			"count": 1,
			"schedule": map[string]any{
				"name": "Horly",
			},
		},
	},
	Enabled: false,
}

func TestGetSnapshotPolicy(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicSnapshotPolicyRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var twoCopiesRecordInterface map[string]any
	err = mapstructure.Decode(twoSnapshotPolicyCopiesRecord, &twoCopiesRecordInterface)
	if err != nil {
		panic(err)
	}
	var notEnabledRecordInterface map[string]any
	err = mapstructure.Decode(notEnabledSnapshotPolicyRecord, &notEnabledRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badSnapshotPolicyRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicRecordInterface}}
	oneTwoCopiesRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{twoCopiesRecordInterface}}
	oneNotEnabledRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{notEnabledRecordInterface}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{basicRecordInterface, basicRecordInterface}}
	genericError := errors.New("generic error for UT")
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/snapshot-policies", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/snapshot-policies", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_copies_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/snapshot-policies", StatusCode: 200, Response: oneTwoCopiesRecord, Err: nil},
		},
		"test_one_not_enabled_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/snapshot-policies", StatusCode: 200, Response: oneNotEnabledRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/snapshot-policies", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_error_1": {
			{ExpectedMethod: "GET", ExpectedURL: "storage/snapshot-policies", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		want      *SnapshotPolicyGetDataModelONTAP
		wantErr   bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &basicSnapshotPolicyRecord, wantErr: false},
		{name: "test_two_copies_record_1", responses: responses["test_two_copies_record_1"], want: &twoSnapshotPolicyCopiesRecord, wantErr: false},
		{name: "test_one_not_enabled_record_1", responses: responses["test_one_not_enabled_record_1"], want: &notEnabledSnapshotPolicyRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_error_1", responses: responses["test_error_1"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetSnapshotPolicy(errorHandler, *r, "string")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSnapshotPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSnapshotPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateSnapshotPolicy(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicSnapshotPolicyRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var twoCopiesRecordInterface map[string]any
	err = mapstructure.Decode(twoSnapshotPolicyCopiesRecord, &twoCopiesRecordInterface)
	if err != nil {
		panic(err)
	}
	var notEnabledRecordInterface map[string]any
	err = mapstructure.Decode(notEnabledSnapshotPolicyRecord, &notEnabledRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badSnapshotPolicyRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	onebasicSnapshotPolicyRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicRecordInterface}}
	oneTwoCopiesRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{twoCopiesRecordInterface}}
	oneNotEnabledRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{notEnabledRecordInterface}}
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_create_basic_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "storage/snapshot-policies", StatusCode: 200, Response: onebasicSnapshotPolicyRecord, Err: nil},
		},
		"test_create_two_copies_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "storage/snapshot-policies", StatusCode: 200, Response: oneTwoCopiesRecord, Err: nil},
		},
		"test_create_not_enabled_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "storage/snapshot-policies", StatusCode: 200, Response: oneNotEnabledRecord, Err: nil},
		},
		"test_error_3": {
			{ExpectedMethod: "POST", ExpectedURL: "storage/snapshot-policies", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name        string
		responses   []restclient.MockResponse
		requestbody SnapshotPolicyResourceBodyDataModelONTAP
		want        *SnapshotPolicyGetDataModelONTAP
		wantErr     bool
	}{
		{name: "test_create_basic_record_1", responses: responses["test_create_basic_record_1"], requestbody: basicSnapshotPolicyBody, want: &basicSnapshotPolicyRecord, wantErr: false},
		{name: "test_create_two_copies_record_1", responses: responses["test_create_two_copies_record_1"], requestbody: twoSnapshotPolicyCopiesBody, want: &twoSnapshotPolicyCopiesRecord, wantErr: false},
		{name: "test_create_not_enabled_record_1", responses: responses["test_create_not_enabled_record_1"], requestbody: notEnabledSnapshotPolicyBody, want: &notEnabledSnapshotPolicyRecord, wantErr: false},
		{name: "test_error_3", responses: responses["test_error_3"], requestbody: badSnapshotPolicyBody, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := CreateSnapshotPolicy(errorHandler, *r, tt.requestbody)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSnapshotPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateSnapshotPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteSnapshotPolicy(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(basicSnapshotPolicyRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badSnapshotPolicyRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_delete": {
			{ExpectedMethod: "DELETE", ExpectedURL: "storage/snapshot-policies/1234", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_error_2": {
			{ExpectedMethod: "DELETE", ExpectedURL: "storage/snapshot-policies/1234", StatusCode: 200, Response: noRecords, Err: genericError},
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
			err2 := DeleteSnapshotPolicy(errorHandler, *r, "1234")
			if err2 != nil {
				fmt.Printf("err2: %s\n", err)
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("DeleteSnapshotPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
