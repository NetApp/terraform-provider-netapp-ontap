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
var basicExportPolicyRuleRecord = ExportPolicyRuleGetDataModelONTAP{
	RoRule:              []string{"krb5i", "krb5"},
	RwRule:              []string{"any"},
	Protocols:           []string{"any"},
	Superuser:           []string{"any"},
	AllowDeviceCreation: true,
	AllowSuid:           true,
	AnonymousUser:       "65534",
	ChownMode:           "restricted",
	ClientsMatch: []ClientMatch{
		{
			Match: "0.0.0.0/0",
		},
	},
	NtfsUnixSecurity: "fail",
	Index:            8,
}

// bad record with wrong type
var badExportPolicyRuleRecord = struct{ Index string }{"123"}

// create export policy rule with basic request body
var basicExportPolicyRuleBody = ExportpolicyRuleResourceBodyDataModelONTAP{
	RoRule:              []string{"krb5i", "krb5"},
	RwRule:              []string{"any"},
	Protocols:           []string{"any"},
	Superuser:           []string{"any"},
	AllowDeviceCreation: true,
	AllowSuid:           true,
	AnonymousUser:       "65534",
	ChownMode:           "restricted",
	ClientsMatch: []map[string]string{
		{
			"match": "0.0.0.0/0",
		},
	},
	NtfsUnixSecurity: "fail",
}

// create export policy with empty comment
var badExportPolicyRuleBody = ExportpolicyRuleResourceBodyDataModelONTAP{
	AnonymousUser: "65534",
}

// update export policy rule on protocols
var updateProtocolsExportPolicyRuleBody = ExportpolicyRuleResourceBodyDataModelONTAP{
	RoRule:              []string{"krb5i", "krb5"},
	RwRule:              []string{"any"},
	Protocols:           []string{"nfs3", "nfs"},
	Superuser:           []string{"any"},
	AllowDeviceCreation: true,
	AllowSuid:           true,
	AnonymousUser:       "65534",
	ChownMode:           "restricted",
	ClientsMatch: []map[string]string{
		{
			"match": "0.0.0.0/0",
		},
	},
	NtfsUnixSecurity: "fail",
}

// update export policy rule with basic request body
var updateExportPolicyRuleErrorBody = ExportpolicyRuleResourceBodyDataModelONTAP{
	RoRule:              []string{"krb5i", "krb5"},
	RwRule:              []string{"any"},
	Protocols:           []string{"nfs3", "nfs"},
	Superuser:           []string{"any"},
	AllowDeviceCreation: true,
	AllowSuid:           true,
	AnonymousUser:       "65534",
	ChownMode:           "restricted",
	ClientsMatch: []map[string]string{
		{
			"match": "0.0.0.0/0",
		},
	},
	NtfsUnixSecurity: "fail",
	Index:            9,
}

func TestGetExportPolicyRuleSingle(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicExportPolicyRuleRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badExportPolicyRuleRecord, &badRecordInterface)
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
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules/8", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules/8", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules/8", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_get_error_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules/8", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		want      *ExportPolicyRuleGetDataModelONTAP
		wantErr   bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: false},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &basicExportPolicyRuleRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_get_error_1", responses: responses["test_get_error_1"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetExportPolicyRuleSingle(errorHandler, *r, "string", 8, versionModelONTAP{Generation: 9, Major: 10})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("TestGetExportPolicyRuleSingle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TestGetExportPolicyRuleSingle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetExportPolicyRule(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicExportPolicyRuleRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badExportPolicyRuleRecord, &badRecordInterface)
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
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules/8", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules/8", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules/8", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_get_error_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules/8", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		want      *ExportPolicyRuleGetDataModelONTAP
		wantErr   bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &basicExportPolicyRuleRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_get_error_1", responses: responses["test_get_error_1"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetExportPolicyRule(errorHandler, *r, "string", 8)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetExportPolicyRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetExportPolicyRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetListExportPolicyRules(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	badRecord := struct{ Svm int }{1}
	var recordInterface map[string]any
	err := mapstructure.Decode(basicExportPolicyRuleRecord, &recordInterface)
	if err != nil {
		panic(err)
	}

	var badRecordInterface map[string]interface{}
	err = mapstructure.Decode(badRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecordsResponse := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	twoRecordsResponse := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	decodeErrorResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]interface{}{badRecordInterface}}

	var wantOneRecord = []ExportPolicyRuleGetDataModelONTAP{basicExportPolicyRuleRecord}
	var wantTwoRecords = []ExportPolicyRuleGetDataModelONTAP{basicExportPolicyRuleRecord, basicExportPolicyRuleRecord}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies/1234/rules", StatusCode: 200, Response: noRecordsResponse, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies/1234/rules", StatusCode: 200, Response: oneRecordResponse, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies/1234/rules", StatusCode: 200, Response: twoRecordsResponse, Err: nil},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/export-policies/1234/rules", StatusCode: 200, Response: decodeErrorResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    []ExportPolicyRuleGetDataModelONTAP
		wantErr bool
		gen     int
		maj     int
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: false, gen: 9, maj: 11},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: wantOneRecord, wantErr: false, gen: 9, maj: 11},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: wantTwoRecords, wantErr: false, gen: 9, maj: 10},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true, gen: 9, maj: 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetListExportPolicyRules(errorHandler, *r, "string", nil, versionModelONTAP{Generation: tt.gen, Major: tt.maj})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetListExportPolicyRules() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetListExportPolicyRules() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateExportPolicyRule(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var basicRecordInterface map[string]any
	err := mapstructure.Decode(basicExportPolicyRuleRecord, &basicRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badExportPolicyRuleRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	onebasicExportPolicyRuleRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicRecordInterface}}
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_create_basic_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules", StatusCode: 200, Response: onebasicExportPolicyRuleRecord, Err: nil},
		},
		"test_create_error_1": {
			{ExpectedMethod: "POST", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name        string
		responses   []restclient.MockResponse
		requestbody ExportpolicyRuleResourceBodyDataModelONTAP
		want        *ExportPolicyRuleGetDataModelONTAP
		wantErr     bool
	}{
		{name: "test_create_basic_record_1", responses: responses["test_create_basic_record_1"], requestbody: basicExportPolicyRuleBody, want: &basicExportPolicyRuleRecord, wantErr: false},
		{name: "test_create_error_1", responses: responses["test_create_error_1"], requestbody: badExportPolicyRuleBody, want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := CreateExportPolicyRule(errorHandler, *r, tt.requestbody, "12884901889")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateExportPolicyRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateExportPolicyRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteSnapshotPolicyRule(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(basicExportPolicyRuleRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badExportPolicyRuleRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_delete": {
			{ExpectedMethod: "DELETE", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules/8", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_error_2": {
			{ExpectedMethod: "DELETE", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules/8", StatusCode: 200, Response: noRecords, Err: genericError},
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
			err2 := DeleteExportPolicyRule(errorHandler, *r, "12884901889", 8)
			if err2 != nil {
				fmt.Printf("err2: %s\n", err)
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("DeleteExportPolicyRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestUpdateExportPolicyRule(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_update_protocols_export_policy_rule": {
			{ExpectedMethod: "PATCH", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules/8", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_update_error_1": {
			{ExpectedMethod: "PATCH", ExpectedURL: "protocols/nfs/export-policies/12884901889/rules/8", StatusCode: 200, Response: noRecords, Err: genericError},
		},
	}
	tests := []struct {
		name        string
		responses   []restclient.MockResponse
		requestbody ExportpolicyRuleResourceBodyDataModelONTAP
		wantErr     bool
	}{
		{name: "test_update_update_protocols_export_policy_rule", responses: responses["test_update_protocols_export_policy_rule"], requestbody: updateProtocolsExportPolicyRuleBody, wantErr: false},
		{name: "test_update_error_1", responses: responses["test_update_error_1"], requestbody: updateExportPolicyRuleErrorBody, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			_, err = UpdateExportPolicyRule(errorHandler, *r, tt.requestbody, "12884901889", 8)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateExportPolicyRule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
