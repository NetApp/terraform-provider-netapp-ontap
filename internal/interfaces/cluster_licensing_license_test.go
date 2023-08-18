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

var basicClusterLicensingLicenseRecord = ClusterLicensingLicenseDataSourceModelONTAP{
	Name: "string",
	Licenses: []LicensesModel{
		{
			SerialNumber: "string",
			Owner:        "string",
			Compliance: Compliance{
				State: "string",
			},
			Active:           false,
			Evaluation:       false,
			InstalledLicense: "string",
		},
	},

	State: "string",
	Scope: "string",
}

var basicClusterLicensingLicenseKeyRecord = ClusterLicensingLicenseKeyDataModelONTAP{
	Name: "string",
	Licenses: []ClusterLicensingLicenseLicensesDataModelONTAP{
		{
			SerialNumber: "string",
		},
	},

	State: "string",
	Scope: "string",
}

var basicClusterLicensingLicenseResourceBodyDataModelONTAP = ClusterLicensingLicenseResourceBodyDataModelONTAP{
	Keys: []string{"string"},
}

// bad record
var badClusterLicensingLicenseRecord = struct{ Name int }{123}

func TestGetClusterLicensingLicenseByName(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(basicClusterLicensingLicenseRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badClusterLicensingLicenseRecord, &badRecordInterface)
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
			{ExpectedMethod: "GET", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *ClusterLicensingLicenseDataSourceModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &basicClusterLicensingLicenseRecord, wantErr: false},
		{name: "test_two_records_error", responses: responses["test_two_records_error"], want: nil, wantErr: true},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetClusterLicensingLicenseByName(errorHandler, *r, "name")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGetClusterLicensingLicenseByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetGetClusterLicensingLicenseByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetListClusterLicensingLicenses(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(basicClusterLicensingLicenseRecord, &recordInterface)
	if err != nil {
		panic(err)
	}

	var badRecordInterface map[string]any
	err = mapstructure.Decode(badClusterLicensingLicenseRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	//genericError := errors.New("generic error for UT")
	badRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}

	var storageVolumeOneRecord = []ClusterLicensingLicenseDataSourceModelONTAP{basicClusterLicensingLicenseRecord}
	var storageVolumeTwoRecords = []ClusterLicensingLicenseDataSourceModelONTAP{basicClusterLicensingLicenseRecord, basicClusterLicensingLicenseRecord}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: twoRecords, Err: nil},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    []ClusterLicensingLicenseDataSourceModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: false},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: storageVolumeOneRecord, wantErr: false},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: storageVolumeTwoRecords, wantErr: false},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetListClusterLicensingLicenses(errorHandler, *r, &ClusterLicensingLicenseFilterModel{Name: ""})
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetGetListClusterLicensingLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetGetListClusterLicensingLicenses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetClusterLicensingLicenses(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(basicClusterLicensingLicenseKeyRecord, &recordInterface)
	if err != nil {
		panic(err)
	}

	var badRecordInterface map[string]any
	err = mapstructure.Decode(badClusterLicensingLicenseRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{recordInterface}}
	twoRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{recordInterface, recordInterface}}
	//genericError := errors.New("generic error for UT")
	badRecordResponse := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}

	var storageVolumeOneRecord = []ClusterLicensingLicenseKeyDataModelONTAP{basicClusterLicensingLicenseKeyRecord}
	var storageVolumeTwoRecords = []ClusterLicensingLicenseKeyDataModelONTAP{basicClusterLicensingLicenseKeyRecord, basicClusterLicensingLicenseKeyRecord}

	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "GET", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: twoRecords, Err: nil},
		},
		"test_decode_error": {
			{ExpectedMethod: "GET", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    []ClusterLicensingLicenseKeyDataModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: []ClusterLicensingLicenseKeyDataModelONTAP{}, wantErr: false},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: storageVolumeOneRecord, wantErr: false},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: storageVolumeTwoRecords, wantErr: false},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetClusterLicensingLicenses(errorHandler, *r)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClusterLicensingLicenses() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetClusterLicensingLicenses() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateClusterLicensingLicense(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(basicClusterLicensingLicenseKeyRecord, &recordInterface)
	if err != nil {
		panic(err)
	}

	var badRecordInterface map[string]any
	err = mapstructure.Decode(badClusterLicensingLicenseRecord, &badRecordInterface)
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
			{ExpectedMethod: "POST", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_record_1": {

			{ExpectedMethod: "POST", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: oneRecord, Err: nil},
		},
		"test_two_records_1": {
			{ExpectedMethod: "POST", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: twoRecords, Err: genericError},
		},
		"test_decode_error": {
			{ExpectedMethod: "POST", ExpectedURL: "/cluster/licensing/licenses", StatusCode: 200, Response: badRecordResponse, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		// args      args
		want    *ClusterLicensingLicenseKeyDataModelONTAP
		wantErr bool
	}{
		{name: "test_one_record_1", responses: responses["test_one_record_1"], want: &basicClusterLicensingLicenseKeyRecord, wantErr: false},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], want: nil, wantErr: true},
		{name: "test_decode_error", responses: responses["test_decode_error"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := CreateClusterLicensingLicense(errorHandler, *r, basicClusterLicensingLicenseResourceBodyDataModelONTAP)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateClusterLicensingLicense() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateClusterLicensingLicense() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteClusterLicensingLicense(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_delete": {
			{ExpectedMethod: "DELETE", ExpectedURL: "cluster/licensing/licenses/license_name", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_error_2": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster/licensing/licenses/license_name", StatusCode: 200, Response: noRecords, Err: genericError},
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
			err2 := DeleteClusterLicensingLicense(errorHandler, *r, "license_name", "serial_number")
			if err2 != nil {
				fmt.Printf("err2: %s\n", err)
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("DeleteClusterLicensingLicense() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
