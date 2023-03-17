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

func TestGetClusterSchedule(t *testing.T) {
	errorHandler := utils.NewErrorHandler(context.Background(), &diag.Diagnostics{})
	cronRecord := ClusterScheduleGetDataModelONTAP{
		Name: "string",
		UUID: "string",
		Cron: CronSchedule{
			Minutes: []int64{1, 2},
		},
		Type:  "cron",
		Scope: "cluster",
	}
	intervalRecord := ClusterScheduleGetDataModelONTAP{
		Name:     "string",
		UUID:     "string",
		Type:     "interval",
		Interval: "string",
		Scope:    "cluster",
	}
	badRecord := struct{ Name int }{123}
	var cronRecordInterface map[string]any
	err := mapstructure.Decode(cronRecord, &cronRecordInterface)
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
	oneCronRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{cronRecordInterface}}
	oneIntervalRecord := restclient.RestResponse{NumRecords: 1, Records: []map[string]any{intervalRecordInterface}}
	twoCronRecords := restclient.RestResponse{NumRecords: 2, Records: []map[string]any{cronRecordInterface, cronRecordInterface}}
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
		// args      args
		want    *ClusterScheduleGetDataModelONTAP
		wantErr bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], want: nil, wantErr: true},
		{name: "test_one_cron_record_1", responses: responses["test_one_cron_record_1"], want: &cronRecord, wantErr: false},
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
