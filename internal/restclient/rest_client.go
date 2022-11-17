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
}

// CallCreateMethod returns response from POST results.  An error is reported if an error is received.
func (c *RestClient) CallCreateMethod(baseURL string, query *RestQuery, body map[string]interface{}) (int, RestResponse, error) {
	statusCode, response, err := c.callAPIMethod("POST", baseURL, query, body)
	if err != nil {
		tflog.Debug(c.ctx, fmt.Sprintf("CallCreateMethod request failed %#v", statusCode))
		return statusCode, RestResponse{}, err
	}

	// TODO: handle waitOnCompletion
	return statusCode, response, err
}

// CallDeleteMethod returns response from DELETE results.  An error is reported if an error is received.
func (c *RestClient) CallDeleteMethod(baseURL string, query *RestQuery, body map[string]interface{}) (int, RestResponse, error) {
	statusCode, response, err := c.callAPIMethod("DELETE", baseURL, query, body)
	if err != nil {
		tflog.Debug(c.ctx, fmt.Sprintf("CallDeleteMethod request failed %#v", statusCode))
		return statusCode, RestResponse{}, err
	}

	// TODO: handle waitOnCompletion
	return statusCode, response, err
}

// GetNilOrOneRecord returns nil if no record is found or a single record.  An error is reported if multiple records are received.
func (c *RestClient) GetNilOrOneRecord(baseURL string, query *RestQuery, body map[string]interface{}) (int, map[string]interface{}, error) {
	statusCode, response, err := c.callAPIMethod("GET", baseURL, query, body)
	if err != nil {
		return statusCode, nil, err
	}
	if response.NumRecords > 1 {
		msg := fmt.Sprintf("received 2 or more records when only one is expected - statusCode %d, err=%#v, response=%#v", statusCode, err, response)
		tflog.Error(c.ctx, msg)
		return statusCode, nil, errors.New(msg)
	}
	if response.NumRecords == 1 {
		return statusCode, response.Records[0], err
	}
	return statusCode, nil, err
}

// GetZeroOrMoreRecords returns a list of records.
func (c *RestClient) GetZeroOrMoreRecords(baseURL string, query *RestQuery, body map[string]interface{}) (int, []map[string]interface{}, error) {
	statusCode, response, err := c.callAPIMethod("GET", baseURL, query, body)
	if err != nil {
		return statusCode, nil, err
	}
	return statusCode, response.Records, err
}

// callAPIMethod can be used to make a request to any REST API method, receiving response as bytes
func (c *RestClient) callAPIMethod(method string, baseURL string, query *RestQuery, body map[string]interface{}) (int, RestResponse, error) {
	if c.mode == "mock" {
		return c.mockCallAPIMethod(method, baseURL, query, body)
	}
	c.waitForAvailableSlot()
	defer c.releaseSlot()

	values := url.Values{}
	if query != nil {
		values = query.Values
	}
	statusCode, response, httpClientErr := c.httpClient.Do(baseURL, &httpclient.Request{
		Method: method,
		Body:   body,
		Query:  values,
	})
	// TODO: error handling for HTTTP status code >=300
	// TODO: handle async calls (job in response)
	return c.unmarshalResponse(statusCode, response, httpClientErr)
}

// NewClient creates a new REST client and a supporting HTTP client
func NewClient(ctx context.Context, cxProfile ConnectionProfile) (*RestClient, error) {
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
		httpClient:            httpclient.NewClient(ctx, httpProfile),
		maxConcurrentRequests: maxConcurrentRequests,
		mode:                  "prod",
		requestSlots:          make(chan int, maxConcurrentRequests),
	}
	return &client, nil
}

func (c *RestClient) waitForAvailableSlot() {
	c.requestSlots <- 1
}

func (c *RestClient) releaseSlot() {
	<-c.requestSlots
}

// NewQuery is used to provide query parameters.  Set and Add functions are inherited from url.Values
func (c *RestClient) NewQuery() *RestQuery {
	query := new(RestQuery)
	query.Values = url.Values{}
	return query
}

// RestQuery is a wrapper around urlValues, and supports a Fields method ina ddition to Set, Add.
type RestQuery struct {
	url.Values
}

// Fields adds a list of fields to query
func (q *RestQuery) Fields(fields []string) {
	q.Set("fields", strings.Join(fields, ","))
}
