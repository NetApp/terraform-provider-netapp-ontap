package restclient

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
)

// RestError maps the REST error structure
type RestError struct {
	Code    string
	Message string
	Target  string
}

// RestResponse to return a list of records (can be empty) and/or errors.
type RestResponse struct {
	NumRecords int `mapstructure:"num_records"`
	Records    []map[string]interface{}
	RestError  RestError `mapstructure:"error"`
	StatusCode int
	HTTPError  string
	ErrorType  string
}

// unmarshalResponse converts the REST response into a structure with a list of 0 or more records.
// we're doing it in two phases:
// unmarshall to intermediate structure, as records may or may not present
// adjust intermediate structure, and decode to final structure
func (c *RestClient) unmarshalResponse(statusCode int, responseJSON []byte, httpClientErr error) (int, RestResponse, error) {
	emptyResponse := RestResponse{
		NumRecords: 0,
		Records:    []map[string]interface{}{},
		RestError:  RestError{},
		StatusCode: statusCode,
		HTTPError:  "",
		ErrorType:  "",
	}
	if httpClientErr != nil {
		emptyResponse.HTTPError = httpClientErr.Error()
		emptyResponse.ErrorType = "http"
		return statusCode, emptyResponse, httpClientErr
	}

	// We don't know which fields are present or not, and fields may not be in a record, so just use interface{}
	var dataMap map[string]interface{}
	if err := json.Unmarshal(responseJSON, &dataMap); err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("unable to unmarshall response, this may be expected when statusCode %d >= 300, unmarshall error=%s, response=%#v", statusCode, err, responseJSON))
		emptyResponse.ErrorType = "bad_response_decode_json"
		return statusCode, emptyResponse, err
	}
	tflog.Debug(c.ctx, fmt.Sprintf("dataMap %#v", dataMap))

	// The returned REST response may or may not contain records.
	// If records is not present, the contents will show in Other.
	type restStagedResponse struct {
		NumRecords int `mapstructure:"num_records"`
		Records    []map[string]interface{}
		Error      RestError
		Other      map[string]interface{} `mapstructure:",remain"`
	}

	var rawResponse restStagedResponse
	var metadata mapstructure.Metadata
	if err := mapstructure.DecodeMetadata(dataMap, &rawResponse, &metadata); err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("unable to format raw response, this may be expected when statusCode %d >= 300, unmarshall error=%s, response=%#v", statusCode, err, dataMap))
		emptyResponse.ErrorType = "bad_response_decode_interface"
		return statusCode, emptyResponse, err
	}

	tflog.Debug(c.ctx, fmt.Sprintf("rawResponse %#v, metadata %#v", rawResponse, metadata))

	// If Other is present, add it to records.
	// But ignore it if we already have some records.
	if rawResponse.NumRecords == 0 && len(rawResponse.Records) == 0 && len(rawResponse.Other) > 0 {
		rawResponse.NumRecords = 1
		rawResponse.Records = append(rawResponse.Records, rawResponse.Other)
	}

	var finalResponse RestResponse
	if err := mapstructure.DecodeMetadata(rawResponse, &finalResponse, &metadata); err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("unable to format final response - statusCode %d, http err=%#v, decode error=%s, response=%#v", statusCode, httpClientErr, err, rawResponse))
		emptyResponse.ErrorType = "bad_response_decode_raw"
		return statusCode, emptyResponse, err
	}

	// If we reached this point, the only possible errors are a bad HTTP status code and/or a REST error encoded in the paybload
	finalResponse.StatusCode = statusCode
	finalResponse, err := c.checkRestErrors(statusCode, finalResponse)
	tflog.Debug(c.ctx, fmt.Sprintf("finalResponse %#v, metadata %#v", finalResponse, metadata))
	return statusCode, finalResponse, err
}

// check for statusCode and RestError
func (c *RestClient) checkRestErrors(statusCode int, response RestResponse) (RestResponse, error) {
	var err error
	if response.RestError.Code != "0" && response.RestError.Code != "" {
		response.ErrorType = "rest_error"
		err = fmt.Errorf("REST reported error %#v, statusCode: %d", response.RestError, statusCode)
	} else if err = c.checkStatusCode(statusCode); err != nil {
		response.ErrorType = "statuscode_error"
	}
	if err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("checkRestError: %s, statusCode %d, response: %#v", err, statusCode, response))
	}
	return response, err
}

// check for statusCode
func (c *RestClient) checkStatusCode(statusCode int) error {
	if statusCode >= 300 || statusCode < 200 {
		return fmt.Errorf("statusCode indicates error, without details: %d", statusCode)
	}
	return nil
}
