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

var dnsRecord = NameServicesDNSGetDataModelONTAP{
	SVM: SvmDataModelONTAP{
		Name: "svmname",
		UUID: "id",
	},
	Servers: []string{
		"10.193.0.250",
		"10.192.0.250",
	},
	Domains: []string{
		"sales.bar.com",
		"sale.bar.com",
	},
}

func TestGetNameServicesDNS(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	badRecord := struct{ Servers int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(dnsRecord, &recordInterface)
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
			{ExpectedMethod: "GET", ExpectedURL: "name-services/dns", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "name-services/dns", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "name-services/dns", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_error_3": {
			{ExpectedMethod: "GET", ExpectedURL: "name-services/dns", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *NameServicesDNSGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &dnsRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_error_3", responses: responses["test_error_3"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetNameServicesDNS(errorHandler, *r, "svmname")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNameServicesDNS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNameServicesDNS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetListNameServicesDNSs(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	badRecord := struct{ Domains int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(dnsRecord, &recordInterface)
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

	var wantOneRecord = []NameServicesDNSGetDataModelONTAP{dnsRecord}
	var wantTwoRecords = []NameServicesDNSGetDataModelONTAP{dnsRecord, dnsRecord}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "name-services/dns", StatusCode: 200, Response: noRecordsResponse, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "name-services/dns", StatusCode: 200, Response: oneRecordResponse, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "name-services/dns", StatusCode: 200, Response: twoRecordsResponse, Err: nil},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "name-services/dns", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    []NameServicesDNSGetDataModelONTAP
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
			got, err := GetListNameServicesDNSs(errorHandler, *r, &NameServicesDNSDataSourceFilterModel{})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetListNameServicesDNSs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetListNameServicesDNSs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateNameServicesDNS(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	badRecord := struct{ Servers int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(dnsRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	oneRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	genericError := errors.New("generic error for UT")
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_one_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "name-services/dns", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "POST", ExpectedURL: "name-services/dns", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_error_3": {
			{ExpectedMethod: "POST", ExpectedURL: "name-services/dns", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *NameServicesDNSGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &dnsRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_error_3", responses: responses["test_error_3"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := CreateNameServicesDNS(errorHandler, *r, dnsRecord)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateNameServicesDNS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateNameServicesDNS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteNameServicesDNS(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	badRecord := struct{ Servers int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(dnsRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	oneRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_one_record_1": {
			{ExpectedMethod: "DELETE", ExpectedURL: "name-services/dns/1234", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "DELETE", ExpectedURL: "name-services/dns/1234", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *NameServicesDNSGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &dnsRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			err2 := DeleteNameServicesDNS(errorHandler, *r, "1234")
			if err2 != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("DeleteNameServicesDNS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
