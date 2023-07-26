package restclient

import (
	"reflect"
	"testing"
)

func TestRestClient_GetNilOrOneRecord(t *testing.T) {
	type args struct {
		baseURL string
		query   *RestQuery
		body    map[string]any
	}
	record := map[string]any{
		"option": "value",
	}
	oneRecord := RestResponse{NumRecords: 1, Records: []map[string]any{record}}
	twoRecords := RestResponse{NumRecords: 2, Records: []map[string]any{record, record}}

	responses := map[string][]MockResponse{
		"test_no_records_1": {
			{"GET", "cluster", 200, RestResponse{}, nil},
		},
		"test_no_records_2": {
			{"GET", "cluster", 200, RestResponse{NumRecords: 0}, nil},
		},
		// "test_no_records_3": {
		// 	{"GET", "cluster", 200, RestResponse{NumRecords: 1, Records: []map[string]interface{}{}}, nil},
		// },
		"test_one_record_1": {
			{"GET", "cluster", 200, oneRecord, nil},
		},
		"test_two_records_1": {
			{"GET", "cluster", 200, twoRecords, nil},
		},
	}
	tests := []struct {
		name      string
		responses []MockResponse
		args      args
		want      int
		want1     map[string]any
		wantErr   bool
	}{
		{name: "test_no_records_1", responses: responses["test_no_records_1"], args: args{baseURL: "cluster"}, want: 200, want1: nil, wantErr: false},
		{name: "test_no_records_2", responses: responses["test_no_records_2"], args: args{baseURL: "cluster"}, want: 200, want1: nil, wantErr: false},
		// {name: "test_no_records_3", responses: responses["test_no_records_3"], args: args{baseURL: "cluster"}, want: 200, want1: nil, wantErr: false},
		{name: "test_one_record_1", responses: responses["test_one_record_1"], args: args{baseURL: "cluster"}, want: 200, want1: record, wantErr: false},
		{name: "test_two_records_1", responses: responses["test_two_records_1"], args: args{baseURL: "cluster"}, want: 200, want1: nil, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c, err := NewMockedRestClient(tt.responses)
			if err != nil {
				panic(err)
			}
			got, got1, err := c.GetNilOrOneRecord(tt.args.baseURL, tt.args.query, tt.args.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("RestClient.GetNilOrOneRecord() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Re(stClient.GetNilOrOneRecord() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("RestClient.GetNilOrOneRecord() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
