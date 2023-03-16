package interfaces

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/utils"
	"reflect"
	"testing"
)

func TestGetProtocolsNfsService(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	record := ProtocolsNfsServiceGetDataModelONTAP{
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
			PermittedEncrptionTypes: []string{
				"aes_256",
				"aes_256",
			},
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
	record910 := ProtocolsNfsServiceGetDataModelONTAP{
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
	badRecord := struct{ Enabled int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(record, &recordInterface)
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
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &record, wantErr: false, gen: 9, maj: 11},
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
