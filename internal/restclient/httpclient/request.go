package httpclient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Request represents a request to a REST API
type Request struct {
	Method string                 `json:"method"`
	Body   map[string]interface{} `json:"body"`
	Query  url.Values             `json:"query"`
	// uuid   string
}

// BuildHTTPReq builds an HTTP request to carry out the REST request
func (r *Request) BuildHTTPReq(c *HTTPClient, baseURL string) (*http.Request, error) {
	url, err := r.BuildURL(c, baseURL, "")
	if err != nil {
		return nil, err
	}
	var req *http.Request
	var body io.Reader
	if len(r.Body) != 0 {
		var bodyJSON []byte
		bodyJSON, err := json.Marshal(r.Body)
		if err != nil {
			return nil, err
		}
		tflog.Debug(c.ctx, "Request JSON payload", map[string]any{
			"json": string(bodyJSON),
		})
		body = bytes.NewReader(bodyJSON)
	}
	req, err = http.NewRequest(r.Method, url, body)

	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	authMethod, err := c.setAuthMethod()
	if err != nil {
		return nil, fmt.Errorf("authentication error: %v", err)
	}

	// Apply authentication based on the determined method
	switch authMethod {
	case AuthMethodBasic:
		// basic_auth - use username and password
		req.SetBasicAuth(c.cxProfile.Username, c.cxProfile.Password)
	case AuthMethodSingleCert:
		// single_cert - certificate authentication only (no Basic Auth headers)
		// Certificate authentication is handled at TLS level in http.Client
	case AuthMethodCertKey:
		// cert_key - certificate with key authentication (no Basic Auth headers)
		// Certificate authentication is handled at TLS level in http.Client
	}
	// telemetry header
	req.Header.Set("X-Dot-Client-App", c.tag)
	// TODO: low pty: add support for form data (require to create a file)

	return req, err
}

// BuildURL using Host, ApiRoot, baseURL, uuid, any query element
func (r *Request) BuildURL(c *HTTPClient, baseURL string, uuid string) (string, error) {
	var err error
	if c == nil {
		err = errors.New("error in BuildUrl, HTTPClient is nil")
	} else if r == nil {
		err = errors.New("error in BuildUrl, request is nil")
	}
	if err != nil {
		return "", err
	}

	// Standard ONTAP URL construction
	if c.cxProfile.Hostname == "" || c.cxProfile.APIRoot == "" {
		return "", errors.New("error in BuildUrl, Hostname and APIRoot are required")
	}
	u := &url.URL{
		Scheme: "https",
		Host:   c.cxProfile.Hostname,
		Path:   c.cxProfile.APIRoot,
	}
	u = u.JoinPath(baseURL, uuid)
	if len(r.Query) != 0 {
		u.RawQuery = r.Query.Encode()
	}
	return u.String(), nil
}
