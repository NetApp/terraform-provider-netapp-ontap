package restclient

import (
	"context"
	"fmt"
)

// MockResponse is used in Unit Testing to mock expected REST responses.
// It validate sthat the request matches ExpectedMethod and ExpectedURL, to return the other elements.
type MockResponse struct {
	ExpectedMethod string
	ExpectedURL    string
	StatusCode     int
	Response       RestResponse
	Err            error
}

// NewMockedRestClient is used in Unit Testing to mock expected REST responses.
func NewMockedRestClient(responses []MockResponse) (*RestClient, error) {
	cxProfile := ConnectionProfile{
		Hostname: "",
		Username: "",
		Password: "",
	}
	restclient, err := NewClient(context.Background(), cxProfile, "resource/version", 600)
	if err != nil {
		panic(err)
	}
	restclient.mode = "mock"
	restclient.responses = responses
	return restclient, nil
}

func (c *RestClient) mockCallAPIMethod(method string, baseURL string, query *RestQuery, body map[string]interface{}) (int, RestResponse, error) {
	if len(c.responses) == 0 {
		panic(fmt.Sprintf("Unexpected request: %s %s", method, baseURL))
	}
	expectedResponse := c.responses[0]
	if expectedResponse.ExpectedMethod != method || expectedResponse.ExpectedURL != baseURL {
		if len(c.responses) == 0 {
			panic(fmt.Sprintf("Unexpected request: %s %s, expecting %s %s", method, baseURL, expectedResponse.ExpectedMethod, expectedResponse.ExpectedURL))
		}
	}
	// remove element now that we know it is consumed
	c.responses = c.responses[1:]
	return expectedResponse.StatusCode, expectedResponse.Response, expectedResponse.Err
}
