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
var basicSnapmirrorPolicyRecord = SnapmirrorPolicyGetRawDataModelONTAP{
	Name: "string",
	SVM: svm{
		Name: "string",
	},
	UUID:    "string",
	Comment: "string",
	Type:    "async",
}

// snapmirror policy with retention
var basicSnapmirrorPolicyRetentionRecord = SnapmirrorPolicyGetRawDataModelONTAP{
	Name: "string",
	SVM: svm{
		Name: "string",
	},
	UUID:    "string",
	Comment: "string",
	Type:    "async",
	Retention: []RetentionGetRawDataModel{
		{
			Count: "5",
			Label: "string",
		},
		{
			Count: "2",
			Label: "string",
		},
	},
}

// snapmirror policy with sync type
var basicSnapmirrorPolicySyncRecord = SnapmirrorPolicyGetRawDataModelONTAP{
	Name: "string",
	SVM: svm{
		Name: "string",
	},
	UUID:     "string",
	Comment:  "string",
	Type:     "sync",
	SyncType: "automated_failover",
}

// snapmirror policy with sync type and retention
var basicSnapmirrorPolicySyncRetentionRecord = SnapmirrorPolicyGetRawDataModelONTAP{
	Name: "string",
	SVM: svm{
		Name: "string",
	},
	UUID:     "string",
	Comment:  "string",
	Type:     "sync",
	SyncType: "sync",
	Retention: []RetentionGetRawDataModel{
		{
			Count: "1",
			Label: "string",
		},
	},
}

// bad decode
var badSnapmirrorPolicyRecord = struct{ Name int }{123}

// bad request body
var badSnapmirrorPolicyBody = SnapmirrorPolicyResourceBodyDataModelONTAP{
	Name: "string",
}

// create with requried parameters
var basicSnapmirrorPolicyBody = SnapmirrorPolicyResourceBodyDataModelONTAP{
	Name: "string",
	SVM: svm{
		Name: "string",
	},
	Comment: "string",
}

// create with retention parameters
var basicSnapmirrorPolicyRetentionBody = SnapmirrorPolicyResourceBodyDataModelONTAP{
	Name: "string",
	SVM: svm{
		Name: "string",
	},
	Comment: "string",
	Retention: []map[string]any{
		{
			"count": 5,
			"label": "string",
		},
		{
			"count": 2,
			"label": "string",
		},
	},
}

// create sync type snapmirror policy
var basicSnapmirrorPolicySyncBody = SnapmirrorPolicyResourceBodyDataModelONTAP{
	Name: "string",
	SVM: svm{
		Name: "string",
	},
	Comment:  "string",
	Type:     "sync",
	SyncType: "automated_failover",
}

// create sync type snapmirror policy with retention
var basicSnapmirrorPolicySyncRetentionBody = SnapmirrorPolicyResourceBodyDataModelONTAP{
	Name: "string",
	SVM: svm{
		Name: "string",
	},
	Comment:  "string",
	Type:     "sync",
	SyncType: "sync",
	Retention: []map[string]any{
		{
			"count": 1,
			"label": "string",
		},
	},
}

// update snapmirror policy with adding a retention
var updateSnapmirrorPolicyRetentionBody = UpdateSnapmirrorPolicyResourceBodyDataModelONTAP{
	Comment: "string",
	Retention: []map[string]any{
		{
			"count": 1,
			"label": "string",
		},
	},
}

// update snapmirror policy comment
var updateSnapmirrorPolicyCommentBody = UpdateSnapmirrorPolicyResourceBodyDataModelONTAP{
	Comment: "new comment",
}

// update snapmirror policy with wrong values
var updateSnapmirrorPolicyErrorBody = struct{ Name int }{123}

func TestGetSnapmirrorPolicy(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicSnapmirrorPolicyRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var oneRetentionRecordInterface map[string]any
	err = mapstructure.Decode(basicSnapmirrorPolicyRetentionRecord, &oneRetentionRecordInterface)
	if err != nil {
		panic(err)
	}
	var oneSyncRecordInterface map[string]any
	err = mapstructure.Decode(basicSnapmirrorPolicySyncRecord, &oneSyncRecordInterface)
	if err != nil {
		panic(err)
	}
	var oneSyncRetentionRecordInterface map[string]any
	err = mapstructure.Decode(basicSnapmirrorPolicySyncRetentionRecord, &oneSyncRetentionRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badSnapmirrorPolicyRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicRecordInterface}}
	oneRetentionRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{oneRetentionRecordInterface}}
	oneSyncRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{oneSyncRecordInterface}}
	oneSyncRetentionRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{oneSyncRetentionRecordInterface}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{basicRecordInterface, basicRecordInterface}}
	genericError := errors.New("generic error for UT")
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/policies", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/policies", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_one_retention_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/policies", StatusCode: 200, Response: oneRetentionRecord, Err: nil},
		},
		"test_one_sync_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/policies", StatusCode: 200, Response: oneSyncRecord, Err: nil},
		},
		"test_one_sync_retention_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/policies", StatusCode: 200, Response: oneSyncRetentionRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/policies", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_error_1": {
			{ExpectedMethod: "GET", ExpectedURL: "snapmirror/policies", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		want      *SnapmirrorPolicyGetRawDataModelONTAP
		wantErr   bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &basicSnapmirrorPolicyRecord, wantErr: false},
		{name: "test_one_retention_record_1", responses: responses["test_one_retention_record_1"], want: &basicSnapmirrorPolicyRetentionRecord, wantErr: false},
		{name: "test_one_sync_record_1", responses: responses["test_one_sync_record_1"], want: &basicSnapmirrorPolicySyncRecord, wantErr: false},
		{name: "test_one_sync_retention_record_1", responses: responses["test_one_sync_retention_record_1"], want: &basicSnapmirrorPolicySyncRetentionRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_error_1", responses: responses["test_error_1"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetSnapmirrorPolicy(errorHandler, *r, "string")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSnapmirrorPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSnapmirrorPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateSnapmirrorPolicy(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicSnapmirrorPolicyRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var basicRetentionRecordInterface map[string]any
	err = mapstructure.Decode(basicSnapmirrorPolicyRetentionRecord, &basicRetentionRecordInterface)
	if err != nil {
		panic(err)
	}
	var basicSyncRecordInterface map[string]any
	err = mapstructure.Decode(basicSnapmirrorPolicySyncRecord, &basicSyncRecordInterface)
	if err != nil {
		panic(err)
	}
	var basicSyncRetentionRecordInterface map[string]any
	err = mapstructure.Decode(basicSnapmirrorPolicySyncRetentionRecord, &basicSyncRetentionRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badSnapmirrorPolicyRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	onebasicSnapmirrorPolicyRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicRecordInterface}}
	onebasicSnapmirrorPolicyRetentionRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicRetentionRecordInterface}}
	onebasicSnapmirrorPolicySyncRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicSyncRecordInterface}}
	onebasicSnapmirrorPolicySyncRetentionRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicSyncRetentionRecordInterface}}
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_create_basic_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "snapmirror/policies", StatusCode: 200, Response: onebasicSnapmirrorPolicyRecord, Err: nil},
		},
		"test_create_retention_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "snapmirror/policies", StatusCode: 200, Response: onebasicSnapmirrorPolicyRetentionRecord, Err: nil},
		},
		"test_create_sync_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "snapmirror/policies", StatusCode: 200, Response: onebasicSnapmirrorPolicySyncRecord, Err: nil},
		},
		"test_create_sync_retention_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "snapmirror/policies", StatusCode: 200, Response: onebasicSnapmirrorPolicySyncRetentionRecord, Err: nil},
		},
		"test_error_3": {
			{ExpectedMethod: "POST", ExpectedURL: "snapmirror/policies", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name        string
		responses   []restclient.MockResponse
		requestbody SnapmirrorPolicyResourceBodyDataModelONTAP
		want        *SnapmirrorPolicyGetRawDataModelONTAP
		wantErr     bool
	}{
		{name: "test_create_basic_record_1", responses: responses["test_create_basic_record_1"], requestbody: basicSnapmirrorPolicyBody, want: &basicSnapmirrorPolicyRecord, wantErr: false},
		{name: "test_create_retention_record_1", responses: responses["test_create_retention_record_1"], requestbody: basicSnapmirrorPolicyRetentionBody, want: &basicSnapmirrorPolicyRetentionRecord, wantErr: false},
		{name: "test_create_sync_record_1", responses: responses["test_create_sync_record_1"], requestbody: basicSnapmirrorPolicySyncBody, want: &basicSnapmirrorPolicySyncRecord, wantErr: false},
		{name: "test_create_sync_retention_record_1", responses: responses["test_create_sync_retention_record_1"], requestbody: basicSnapmirrorPolicySyncRetentionBody, want: &basicSnapmirrorPolicySyncRetentionRecord, wantErr: false},
		{name: "test_error_3", responses: responses["test_error_3"], requestbody: badSnapmirrorPolicyBody, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := CreateSnapmirrorPolicy(errorHandler, *r, tt.requestbody)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSnapmirrorPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateSnapmirrorPolicy() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteSnapmirrorPolicy(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(basicSnapmirrorPolicyRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badSnapmirrorPolicyRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_delete": {
			{ExpectedMethod: "DELETE", ExpectedURL: "snapmirror/policies/1234", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_error_2": {
			{ExpectedMethod: "DELETE", ExpectedURL: "snapmirror/policies/1234", StatusCode: 200, Response: noRecords, Err: genericError},
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
			err2 := DeleteSnapmirrorPolicy(errorHandler, *r, "1234")
			if err2 != nil {
				fmt.Printf("err2: %s\n", err)
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("DeleteSnapmirrorPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUpdateSnapmirrorPolicy(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_update_add_retention": {
			{ExpectedMethod: "PATCH", ExpectedURL: "snapmirror/policies/1234", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_update_comment_snapmirror_policy": {
			{ExpectedMethod: "PATCH", ExpectedURL: "snapmirror/policies/1234", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_update_error_1": {
			{ExpectedMethod: "PATCH", ExpectedURL: "snapmirror/policies/1234", StatusCode: 200, Response: noRecords, Err: genericError},
		},
	}
	tests := []struct {
		name        string
		responses   []restclient.MockResponse
		requestbody any
		wantErr     bool
	}{
		{name: "test_update_add_retention", responses: responses["test_update_add_retention"], requestbody: updateSnapmirrorPolicyRetentionBody, wantErr: false},
		{name: "test_update_comment_snapmirror_policy", responses: responses["test_update_comment_snapmirror_policy"], requestbody: updateSnapmirrorPolicyCommentBody, wantErr: false},
		{name: "test_update_error_1", responses: responses["test_update_error_1"], requestbody: updateSnapmirrorPolicyErrorBody, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			err = UpdateSnapmirrorPolicy(errorHandler, *r, tt.requestbody, "string")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateSnapmirrorPolicy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
