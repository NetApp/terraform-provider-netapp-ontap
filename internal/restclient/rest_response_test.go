package restclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/mitchellh/mapstructure"
)

func TestRestClient_unmarshalResponse(t *testing.T) {
	type args struct {
		statusCode    int
		responseJSON  []byte
		httpClientErr error
	}
	responseForJSON := map[string]any{
		"num_records": 1,
		"records": []map[string]any{
			{"option": "value"},
		},
		"statuscode": 200}
	response := RestResponse{
		NumRecords: 1,
		Records: []map[string]any{
			{"option": "value"},
		},
		StatusCode: 200}
	responseOthers := RestResponse{
		NumRecords: 1,
		Records: []map[string]any{
			{"_link": "somelink", "option": "value"},
		},
		StatusCode: 200}
	restError := RestError{"123", "", ""}
	responseRestError := RestResponse{
		NumRecords: 0,
		Records:    []map[string]any(nil),
		RestError:  restError,
		StatusCode: 400,
		HTTPError:  "",
		ErrorType:  "rest_error",
	}
	responseStatusCodeError := RestResponse{
		NumRecords: 0,
		Records:    []map[string]any(nil),
		StatusCode: 400,
		ErrorType:  "statuscode_error",
	}
	rawResponseRestError := struct {
		Error RestError
	}{
		Error: restError,
	}
	responseOther := map[string]any{"_link": "somelink", "option": "value"}

	rawEmpty := any(nil)
	emptyJSON, err := json.Marshal(rawEmpty)
	if err != nil {
		panic(err)
	}
	var responseInterfaceOther map[string]any
	responseJSON, err := json.Marshal(responseForJSON)
	if err != nil {
		panic(err)
	}
	responseJSONOther, err := json.Marshal(responseOther)
	if err != nil {
		panic(err)
	}
	responseJSONRestError, err := json.Marshal(rawResponseRestError)
	if err != nil {
		panic(err)
	}
	err = mapstructure.Decode(responseOther, &responseInterfaceOther)
	if err != nil {
		panic(err)
	}
	badData := map[string]string{"num_records": "123"}
	badJSON, err := json.Marshal(badData)
	if err != nil {
		panic(err)
	}
	genericError := errors.New("generic error for UT")

	tests := []struct {
		name    string
		args    args
		want    int
		want1   RestResponse
		wantErr bool
	}{
		{name: "error_no_json", args: args{}, want: 0, want1: RestResponse{ErrorType: "bad_response_decode_json", Records: []map[string]any{}}, wantErr: true},
		{name: "error_mismatch_json", args: args{statusCode: 200, responseJSON: badJSON}, want: 200, want1: RestResponse{ErrorType: "bad_response_decode_interface", Records: []map[string]any{}, StatusCode: 200}, wantErr: true},
		{name: "error_http_error", args: args{httpClientErr: genericError}, want: 0, want1: RestResponse{HTTPError: genericError.Error(), ErrorType: "http", Records: []map[string]any{}}, wantErr: true},
		{name: "json_unmarshalled", args: args{statusCode: 200, responseJSON: responseJSON}, want: 200, want1: response, wantErr: false},
		{name: "json_unmarshalled_other", args: args{statusCode: 200, responseJSON: responseJSONOther}, want: 200, want1: responseOthers, wantErr: false},
		{name: "rest_error", args: args{statusCode: 400, responseJSON: responseJSONRestError}, want: 400, want1: responseRestError, wantErr: true},
		{name: "status_code_error_1", args: args{statusCode: 400, responseJSON: responseJSONRestError}, want: 400, want1: responseRestError, wantErr: true},
		{name: "status_code_error_2", args: args{statusCode: 400, responseJSON: emptyJSON}, want: 400, want1: responseStatusCodeError, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &RestClient{
				ctx: context.Background(),
			}
			got, got1, err := c.unmarshalResponse(tt.args.statusCode, tt.args.responseJSON, tt.args.httpClientErr)
			if err != nil {
				fmt.Printf("err: %s\n", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("RestClient.unmarshalResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("RestClient.unmarshalResponse() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("RestClient.unmarshalResponse() got1 = %#v, want %#v", got1, tt.want1)
			}
		})
	}
}
