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

var nfsServiceRecord = ProtocolsNfsServiceGetDataModelONTAP{
	Enabled: true,
	Protocol: Protocol{
		V3Enabled:  true,
		V4IdDomain: "string",
		V40Enabled: true,
		V40Features: V40Features{
			ACLEnabled:             false,
			ReadDelegationEnabled:  false,
			WriteDelegationEnabled: false,
		},
		V41Enabled: false,
		V41Features: V41Features{
			ACLEnabled:             false,
			PnfsEnabled:            false,
			ReadDelegationEnabled:  false,
			WriteDelegationEnabled: false,
		},
	},
	Root: Root{
		IgnoreNtACL:              false,
		SkipWritePermissionCheck: false,
	},
	Security: Security{
		ChownMode:              "use_export_policy",
		NtACLDisplayPermission: false,
		NtfsUnixSecurity:       "use_export_policy",
		RpcsecContextIdel:      0,
	},
	ShowmountEnabled: true,
	Transport: Transport{
		TCP:            true,
		UDP:            true,
		TCPMaxXferSize: 16384,
	},
	VstorageEnabled: true,
	Windows: Windows{
		DefaultUser:                "carchi8py",
		MapUnknownUIDToDefaultUser: true,
		V3MsDosClientEnabled:       false,
	},
}

var record910 = ProtocolsNfsServiceGetDataModelONTAP{
	Enabled: true,
	Protocol: Protocol{
		V3Enabled:  true,
		V4IdDomain: "string",
		V40Enabled: true,
		V40Features: V40Features{
			ACLEnabled:             false,
			ReadDelegationEnabled:  false,
			WriteDelegationEnabled: false,
		},
		V41Enabled: false,
		V41Features: V41Features{
			ACLEnabled:             false,
			PnfsEnabled:            false,
			ReadDelegationEnabled:  false,
			WriteDelegationEnabled: false,
		},
	},
	ShowmountEnabled: true,
	Transport: Transport{
		TCP: true,
		UDP: true,
	},
	VstorageEnabled: true,
}

func TestGetProtocolsNfsService(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	badRecord := struct{ Enabled int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(nfsServiceRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var recordInterface910 map[string]any
	err = mapstructure.Decode(record910, &recordInterface910)
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
	one910Record := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface910}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	genericError := errors.New("generic error for UT")
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_one_910_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: one910Record, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_error_3": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *ProtocolsNfsServiceGetDataModelONTAP
		wantErr bool
		gen     int
		maj     int
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true, gen: 9, maj: 11},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &nfsServiceRecord, wantErr: false, gen: 9, maj: 11},
		{name: "test_one_910_record_1", responses: responses["test_one_910_record_1"], want: &record910, wantErr: false, gen: 9, maj: 10},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true, gen: 9, maj: 11},
		{name: "test_error_3", responses: responses["test_error_3"], want: nil, wantErr: true, gen: 9, maj: 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetProtocolsNfsService(errorHandler, *r, "svmname", versionModelONTAP{Generation: tt.gen, Major: tt.maj})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProtocolsNfsServic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateProtocolsNfsService(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	badRecord := struct{ Enabled int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(nfsServiceRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var recordInterface910 map[string]any
	err = mapstructure.Decode(record910, &recordInterface910)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	oneRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	one910Record := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface910}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	genericError := errors.New("generic error for UT")
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_one_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_one_910_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: one910Record, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "POST", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_error_3": {
			{ExpectedMethod: "POST", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		requestBody ProtocolsNfsServiceResourceDataModelONTAP
		want        *ProtocolsNfsServiceGetDataModelONTAP
		wantErr     bool
	}{
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &nfsServiceRecord, wantErr: false},
		{name: "test_one_910_record_1", responses: responses["test_one_910_record_1"], want: &record910, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_error_3", responses: responses["test_error_3"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := CreateProtocolsNfsService(errorHandler, *r, tt.requestBody)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProtocolsNfsServic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetCluster() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteProtocolsNfsService(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_delete": {
			{ExpectedMethod: "DELETE", ExpectedURL: "protocols/nfs/services/1234", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_error_2": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services/1234", StatusCode: 200, Response: noRecords, Err: genericError},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *ProtocolsNfsServiceGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_delete", responses: responses["test_delete"], want: &nfsServiceRecord, wantErr: false},
		{name: "test_error_2", responses: responses["test_error_2"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			err2 := DeleteProtocolsNfsService(errorHandler, *r, "1234")
			if err2 != nil {
				fmt.Printf("err2: %s\n", err)
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("GetProtocolsNfsServic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestGetProtocolsNfsServices(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	badRecord := struct{ Enabled int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(nfsServiceRecord, &recordInterface)
	if err != nil {
		panic(err)
	}

	var record910Interface map[string]any
	err = mapstructure.Decode(record910, &record910Interface)
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
	oneRecord910Response := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{record910Interface}}
	twoRecordsResponse := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	badRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}

	var wantOneRecord = []ProtocolsNfsServiceGetDataModelONTAP{nfsServiceRecord}
	var wantTwoRecords = []ProtocolsNfsServiceGetDataModelONTAP{nfsServiceRecord, nfsServiceRecord}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: noRecordsResponse, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: oneRecordResponse, Err: nil},
		},
		"test_one_910_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: oneRecord910Response, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: twoRecordsResponse, Err: nil},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    []ProtocolsNfsServiceGetDataModelONTAP
		wantErr bool
		gen     int
		maj     int
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: false, gen: 9, maj: 11},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: wantOneRecord, wantErr: false, gen: 9, maj: 11},
		{name: "test_one_910_record_1", responses: responses["test_one_910_record_1"], want: []ProtocolsNfsServiceGetDataModelONTAP{record910}, wantErr: false, gen: 9, maj: 10},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: wantTwoRecords, wantErr: false, gen: 9, maj: 11},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true, gen: 9, maj: 11},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetProtocolsNfsServices(errorHandler, *r, &NfsServicesFilterModel{}, versionModelONTAP{Generation: tt.gen, Major: tt.maj})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProtocolsNfsServices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProtocolsNfsServices() = %v, want %v", got, tt.want)
			}
		})
	}
}
