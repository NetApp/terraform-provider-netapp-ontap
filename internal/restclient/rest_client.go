package restclient

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/mitchellh/mapstructure"
	"github.com/netapp/terraform-provider-netapp-ontap/internal/restclient/httpclient"
)

// ConnectionProfile describes out to reach a cluster or vserver
type ConnectionProfile struct {
	// TODO: add certs in addition to basic authentication
	// TODO: Add Timeout (currently hardcoded to 10 seconds)
	Hostname              string
	Username              string
	Password              string
	ValidateCerts         bool
	MaxConcurrentRequests int
}

// RestClient to interact with the ONTAP REST API
type RestClient struct {
	connectionProfile     ConnectionProfile
	ctx                   context.Context
	maxConcurrentRequests int
	httpClient            httpclient.HTTPClient
	requestSlots          chan int
	mode                  string
	responses             []MockResponse
	tag                   string
}

// CallCreateMethod returns response from POST results.  An error is reported if an error is received.
func (r *RestClient) CallCreateMethod(baseURL string, query *RestQuery, body map[string]interface{}) (int, RestResponse, error) {
	if query == nil {
		query = r.NewQuery()
	}
	// TODO: make this a connection paramter ?
	query.Set("return_timeout", "60")
	statusCode, response, err := r.callAPIMethod("POST", baseURL, query, body)
	if err != nil {
		tflog.Debug(r.ctx, fmt.Sprintf("CallCreateMethod request failed %#v", statusCode))
		return statusCode, RestResponse{}, err
	}

	// TODO: handle waitOnCompletion
	return statusCode, response, err
}

// CallDeleteMethod returns response from DELETE results.  An error is reported if an error is received.
func (r *RestClient) CallDeleteMethod(baseURL string, query *RestQuery, body map[string]interface{}) (int, RestResponse, error) {
	if query == nil {
		query = r.NewQuery()
	}
	// TODO: make this a connection paramter ?
	query.Set("return_timeout", "60")
	statusCode, response, err := r.callAPIMethod("DELETE", baseURL, query, body)
	if err != nil {
		tflog.Debug(r.ctx, fmt.Sprintf("CallDeleteMethod request failed %#v", statusCode))
		return statusCode, RestResponse{}, err
	}

	// TODO: handle waitOnCompletion
	return statusCode, response, err
}

// GetNilOrOneRecord returns nil if no record is found or a single record.  An error is reported if multiple records are received.
func (r *RestClient) GetNilOrOneRecord(baseURL string, query *RestQuery, body map[string]interface{}) (int, map[string]interface{}, error) {
	statusCode, response, err := r.callAPIMethod("GET", baseURL, query, body)
	if err != nil {
		return statusCode, nil, err
	}
	if response.NumRecords > 1 {
		msg := fmt.Sprintf("received 2 or more records when only one is expected - statusCode %d, err=%#v, response=%#v", statusCode, err, response)
		tflog.Error(r.ctx, msg)
		return statusCode, nil, errors.New(msg)
	}
	if response.NumRecords == 1 {
		return statusCode, response.Records[0], err
	}
	return statusCode, nil, err
}

// GetZeroOrMoreRecords returns a list of records.
func (r *RestClient) GetZeroOrMoreRecords(baseURL string, query *RestQuery, body map[string]interface{}) (int, []map[string]interface{}, error) {
	statusCode, response, err := r.callAPIMethod("GET", baseURL, query, body)
	if err != nil {
		return statusCode, nil, err
	}
	return statusCode, response.Records, err
}

// callAPIMethod can be used to make a request to any REST API method, receiving response as bytes
func (r *RestClient) callAPIMethod(method string, baseURL string, query *RestQuery, body map[string]interface{}) (int, RestResponse, error) {
	if r.mode == "mock" {
		return r.mockCallAPIMethod(method, baseURL, query, body)
	}
	r.waitForAvailableSlot()
	defer r.releaseSlot()

	values := url.Values{}
	if query != nil {
		values = query.Values
	}
	statusCode, response, httpClientErr := r.httpClient.Do(baseURL, &httpclient.Request{
		Method: method,
		Body:   body,
		Query:  values,
	})

	// TODO: error handling for HTTTP status code >=300
	// TODO: handle async calls (job in response)
	return r.unmarshalResponse(statusCode, response, httpClientErr)
}

// NewClient creates a new REST client and a supporting HTTP client
func NewClient(ctx context.Context, cxProfile ConnectionProfile, tag string) (*RestClient, error) {
	var httpProfile httpclient.HTTPProfile
	err := mapstructure.Decode(cxProfile, &httpProfile)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("decode error on ConnectionProfile %#v to HTTPProfile", cxProfile))
		return nil, fmt.Errorf("decode error on ConnectionProfile %#v to HTTPProfile", cxProfile)
	}
	httpProfile.APIRoot = "api"
	maxConcurrentRequests := cxProfile.MaxConcurrentRequests
	if maxConcurrentRequests == 0 {
		maxConcurrentRequests = 6
	}
	client := RestClient{
		connectionProfile:     cxProfile,
		ctx:                   ctx,
		httpClient:            httpclient.NewClient(ctx, httpProfile, tag),
		maxConcurrentRequests: maxConcurrentRequests,
		mode:                  "prod",
		requestSlots:          make(chan int, maxConcurrentRequests),
		tag:                   tag,
	}
	return &client, nil
}

func (r *RestClient) waitForAvailableSlot() {
	r.requestSlots <- 1
}

func (r *RestClient) releaseSlot() {
	<-r.requestSlots
}

// NewQuery is used to provide query parameters.  Set and Add functions are inherited from url.Values
func (r *RestClient) NewQuery() *RestQuery {
	query := new(RestQuery)
	query.Values = url.Values{}
	return query
}

// RestQuery is a wrapper around urlValues, and supports a Fields method in addition to Set, Add.
type RestQuery struct {
	url.Values
}

// Fields adds a list of fields to query
func (q *RestQuery) Fields(fields []string) {
	q.Set("fields", strings.Join(fields, ","))
}

// Equals is a test function for Unit Testing
func (r *RestClient) Equals(r2 *RestClient) (ok bool, firstDiff string) {
	if r.connectionProfile != r2.connectionProfile {
		return false, fmt.Sprintf("expected %#v, got %#v", r.connectionProfile, r2.connectionProfile)
	}
	if r.tag != r2.tag {
		return false, fmt.Sprintf("expected %#v, got %#v", r.tag, r2.tag)
	}
	return true, ""
}
