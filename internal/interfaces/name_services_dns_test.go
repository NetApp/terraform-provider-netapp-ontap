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

func TestGetNameServicesDNS(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	record := NameServicesDNSGetDataModelONTAP{
		SVM: SvmDataModelONTAP{
			Name: "svmname",
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
	badRecord := struct{ Servers int }{123}
	var recordInterface map[string]any
	err := mapstructure.Decode(record, &recordInterface)
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
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "protocols/nfs/services", StatusCode: 200, Response: oneRecord, Err: nil},
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
		want    *NameServicesDNSGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &record, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_error_3", responses: responses["test_error_3"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetNameServicesDNS(errorHandler, *r, "svmname", "svm")
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
