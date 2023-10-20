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

var ipInterfaceRecord = IPInterfaceGetDataModelONTAP{
	Name: "string",
	SVM: IPInterfaceSvmName{
		Name: "string",
	},
	Scope: "string",
	UUID:  "string",
	IP: IPInterfaceGetIP{
		Address: "string",
		Netmask: "string",
	},
	Location: IPInterfaceResourceLocation{
		HomeNode: &IPInterfaceResourceHomeNode{
			Name: "string",
		},
		HomePort: &IPInterfaceResourceHomePort{
			Name: "string",
			Node: IPInterfaceResourceHomeNode{
				Name: "string",
			},
		},
	},
}

func TestGetIPInterface(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var recordInterface map[string]any
	err := mapstructure.Decode(ipInterfaceRecord, &recordInterface)
	if err != nil {
		panic(err)
	}

	var badRecordInterface map[string]any
	badRecord := struct{ Name int }{123}
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
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/interfaces", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/interfaces", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/interfaces", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/interfaces", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *IPInterfaceGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &ipInterfaceRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetIPInterface(errorHandler, *r, "name", "svmName")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetIPInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetIPInterface() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetListIPInterfaces(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})

	var recordInterface map[string]any
	err := mapstructure.Decode(ipInterfaceRecord, &recordInterface)
	if err != nil {
		panic(err)
	}

	var badRecordInterface map[string]any
	badRecord := struct{ Name int }{123}
	err = mapstructure.Decode(badRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}

	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	badRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}

	var ipInterfacesOneRecord = []IPInterfaceGetDataModelONTAP{ipInterfaceRecord}
	var ipInterfacesTwoRecords = []IPInterfaceGetDataModelONTAP{ipInterfaceRecord, ipInterfaceRecord}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/interfaces", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/interfaces", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/interfaces", StatusCode: 200, Response: twoRecords, Err: nil},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "network/ip/interfaces", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    []IPInterfaceGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: false},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: ipInterfacesOneRecord, wantErr: false},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: ipInterfacesTwoRecords, wantErr: false},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetListIPInterfaces(errorHandler, *r, &IPInterfaceDataSourceFilterModel{})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetListIPInterfaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetListIPInterfaces() = %v, want %v", got, tt.want)
			}
		})
	}
}
