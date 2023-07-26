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

var basicCronRecord = ClusterScheduleGetDataModelONTAP{
	Name: "string",
	UUID: "string",
	Cron: CronSchedule{
		Minutes:  []int64{1, 2},
		Hours:    []int64{10},
		Days:     []int64{1},
		Months:   []int64{6, 7},
		Weekdays: []int64{1, 3},
	},
	Type:  "cron",
	Scope: "cluster",
}

var intervalRecord = ClusterScheduleGetDataModelONTAP{
	Name:     "string",
	UUID:     "string",
	Type:     "interval",
	Interval: "string",
	Scope:    "cluster",
}

var badRecord = struct{ Name int }{123}

var basicCronBody = ClusterScheduleResourceBodyDataModelONTAP{
	Name: "string",
	Cron: CronSchedule{
		Minutes:  []int64{2, 3},
		Hours:    []int64{10},
		Days:     []int64{1},
		Months:   []int64{6, 7},
		Weekdays: []int64{1, 3},
	},
}

var basicIntervalBody = ClusterScheduleResourceBodyDataModelONTAP{
	Name:     "string",
	Interval: "string",
}

var badBody = ClusterScheduleResourceBodyDataModelONTAP{
	Name: "string",
}

func TestGetClusterSchedule(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var basicCronRecordInterface map[string]any
	err := mapstructure.Decode(basicCronRecord, &basicCronRecordInterface)
	if err != nil {
		panic(err)
	}
	var intervalRecordInterface map[string]any
	err = mapstructure.Decode(intervalRecord, &intervalRecordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	oneCronRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicCronRecordInterface}}
	oneIntervalRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{intervalRecordInterface}}
	twoCronRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{basicCronRecordInterface, basicCronRecordInterface}}
	genericError := errors.New("generic error for UT")
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_no_records_1": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster/schedules", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_one_cron_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster/schedules", StatusCode: 200, Response: oneCronRecord, Err: nil},
		},
		"test_one_interval_record_1": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster/schedules", StatusCode: 200, Response: oneIntervalRecord, Err: nil},
		},
		"test_two_cron_records_error": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster/schedules", StatusCode: 200, Response: twoCronRecords, Err: genericError},
		},
		"test_error_1": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster/schedules", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		want      *ClusterScheduleGetDataModelONTAP
		wantErr   bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_cron_record_1", responses: responses["test_one_cron_record_1"], want: &basicCronRecord, wantErr: false},
		{name: "test_one_interval_record_1", responses: responses["test_one_interval_record_1"], want: &intervalRecord, wantErr: false},
		{name: "test_two_cron_records_error", responses: responses["test_two_cron_records_error"], want: nil, wantErr: true},
		{name: "test_error_1", responses: responses["test_error_1"], want: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := GetClusterSchedule(errorHandler, *r, "string")
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("GetClusterSchedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetClusterSchedule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateClusterSchedule(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	// cron type
	var basicCronRecordInterface map[string]any
	err := mapstructure.Decode(basicCronRecord, &basicCronRecordInterface)
	if err != nil {
		panic(err)
	}
	// interval
	var intervalRecordInterface map[string]any
	err = mapstructure.Decode(intervalRecord, &intervalRecordInterface)
	if err != nil {
		panic(err)
	}
	// bad record
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	oneBasicCronRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{basicCronRecordInterface}}
	oneIntervalRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{intervalRecordInterface}}
	decodeError := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{badRecordInterface}}
	responses := map[string][]restclient.MockResponse{
		"test_create_basic_corn_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "cluster/schedules", StatusCode: 200, Response: oneBasicCronRecord, Err: nil},
		},
		"test_create_interval_record_1": {
			{ExpectedMethod: "POST", ExpectedURL: "cluster/schedules", StatusCode: 200, Response: oneIntervalRecord, Err: nil},
		},
		"test_create_error": {
			{ExpectedMethod: "POST", ExpectedURL: "cluster/schedules", StatusCode: 200, Response: decodeError, Err: nil},
		},
	}

	tests := []struct {
		name        string
		responses   []restclient.MockResponse
		requestbody ClusterScheduleResourceBodyDataModelONTAP
		want        *ClusterScheduleGetDataModelONTAP
		wantErr     bool
	}{
		{name: "test_create_basic_corn_record_1", responses: responses["test_create_basic_corn_record_1"], requestbody: basicCronBody, want: &basicCronRecord, wantErr: false},
		{name: "test_create_interval_record_1", responses: responses["test_create_interval_record_1"], requestbody: basicIntervalBody, want: &intervalRecord, wantErr: false},
		{name: "test_create_error", responses: responses["test_create_error"], requestbody: badBody, want: nil, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, err := CreateClusterSchedule(errorHandler, *r, tt.requestbody)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateClusterSchedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateClusterSchedule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDeleteClusterSchedule(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	var recordInterface map[string]any
	err := mapstructure.Decode(basicCronRecord, &recordInterface)
	if err != nil {
		panic(err)
	}
	var badRecordInterface map[string]any
	err = mapstructure.Decode(badRecord, &badRecordInterface)
	if err != nil {
		panic(err)
	}
	noRecords := restclient.RestResponse{NumRecords: 0, Records: []map[string]any{}}
	genericError := errors.New("generic error for UT")
	responses := map[string][]restclient.MockResponse{
		"test_delete_one_record": {
			{ExpectedMethod: "DELETE", ExpectedURL: "cluster/schedules/1234", StatusCode: 200, Response: noRecords, Err: nil},
		},
		"test_delete_error": {
			{ExpectedMethod: "GET", ExpectedURL: "cluster/scheduless/1234", StatusCode: 200, Response: noRecords, Err: genericError},
		},
	}
	tests := []struct {
		name      string
		responses []restclient.MockResponse
		wantErr   bool
	}{
		{name: "test_delete_one_record", responses: responses["test_delete_one_record"], wantErr: false},
		{name: "test_delete_error", responses: responses["test_delete_error"], wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := restclient.NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			err2 := DeleteClusterSchedule(errorHandler, *r, "1234")
			if err2 != nil {
				fmt.Printf("err2: %s\n", err)
			}
			if (err2 != nil) != tt.wantErr {
				t.Errorf("DeleteClusterSchedule() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
