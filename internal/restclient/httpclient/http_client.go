package httpclient

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// HTTPClient represents a client for interaction with a ONTAP REST API
type HTTPClient struct {
	cxProfile  HTTPProfile
	ctx        context.Context
	httpClient http.Client
	tag        string
}

// HTTPProfile defines the connection attributes to build the base URL and authentication header
type HTTPProfile struct {
	APIRoot       string
	Hostname      string
	Username      string
	Password      string
	ValidateCerts bool
}

// Do sends the API Request, parses the response as JSON, and returns the HTTP status code as int, the "result" value as byte
// possible errors:
//
//	no response body:
//		failed to build HTTP request - statusCode forced to -1
//		failed to send HTTP request - statusCode forced to -1 unless it is present in the response
//		failed to read HTTP response body - statusCode from response if present, otherwise -1
//		empty response body (check with POST/PATCH/DELETE if this is really a problem)  - statusCode from response if present, otherwise -1
func (c *HTTPClient) Do(baseURL string, req *Request) (int, []byte, error) {
	httpReq, err := req.BuildHTTPReq(c, baseURL)
	statusCode := -1
	if err != nil {
		return statusCode, nil, err
	}
	tflog.Debug(c.ctx, fmt.Sprintf("sending: %s %s", httpReq.Method, httpReq.URL.String()), map[string]any{"body": req.Body})
	httpRes, err := c.httpClient.Do(httpReq)
	if httpRes != nil {
		statusCode = httpRes.StatusCode
	}
	if err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("HTTP request failed: %s, statusCode: %d", err, statusCode))
		return statusCode, nil, err
	}

	defer httpRes.Body.Close()

	body, err := io.ReadAll(httpRes.Body)
	if err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("HTTP response read failed: %s, statusCode: %d", err, statusCode))
		return statusCode, nil, err
	}

	if body == nil {
		return httpRes.StatusCode, nil, fmt.Errorf("no result returned in REST response.  statusCode %d", statusCode)
	}

	tflog.Debug(c.ctx, fmt.Sprintf("received: %s %s %d", req.Method, httpReq.URL.String(), statusCode), map[string]any{"res": string(body)})

	return httpRes.StatusCode, body, nil
}

// NewClient creates a new HTTP client
func NewClient(ctx context.Context, cxProfile HTTPProfile, tag string) HTTPClient {
	client := HTTPClient{
		cxProfile: cxProfile,
		ctx:       ctx,
		tag:       tag,
	}
	client.httpClient = client.create()
	return client
}

// create configures and creates the http client
func (c HTTPClient) create() http.Client {
	if !c.cxProfile.ValidateCerts {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	return http.Client{Timeout: 120 * time.Second}
}
