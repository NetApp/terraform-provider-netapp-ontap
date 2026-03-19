package httpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
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
	ClientCertFile string
	ClientKeyFile  string
	CACertFile     string
}

// AuthMethod represents the authentication method to use
type AuthMethod int

const (
	AuthMethodNone AuthMethod = iota
	AuthMethodBasic
	AuthMethodCert
	AuthMethodCertWithBasic
)

// setAuthMethod determines the authentication method based on available credentials
func (c *HTTPClient) setAuthMethod() (AuthMethod, error) {
	// defaults to cert authentication if both basic and client certificate authentication parameters are given
	if c.cxProfile.ClientCertFile != "" {
		if c.cxProfile.ClientKeyFile == "" {
			return AuthMethodNone, errors.New("error: cannot have a cert file without a key file")
		}
		// If username is also provided with cert, use both
		if c.cxProfile.Username != "" {
			return AuthMethodCertWithBasic, nil
		}
		return AuthMethodCert, nil
	}
	
	if c.cxProfile.ClientKeyFile != "" {
		return AuthMethodNone, errors.New("error: cannot have a key file without a cert file")
	}
	
	if c.cxProfile.Password == "" && c.cxProfile.Username == "" {
		return AuthMethodNone, errors.New("error: ONTAP module requires username/password or SSL certificate file(s)")
	}
	
	if c.cxProfile.Password != "" && c.cxProfile.Username != "" {
		return AuthMethodBasic, nil
	}
	
	return AuthMethodNone, errors.New("error: username and password have to be provided together")
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
	tflog.Debug(c.ctx, fmt.Sprintf("sending: %s %s", httpReq.Method, httpReq.URL.String()), map[string]any{"body": req.Body, "headers": httpReq.Header})
	httpRes, err := c.httpClient.Do(httpReq)
	if httpRes != nil {
		statusCode = httpRes.StatusCode
	}
	if err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("HTTP request failed: %s, statusCode: %d, err raw:%#v", err, statusCode, err))
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

func (c HTTPClient) create() http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: !c.cxProfile.ValidateCerts,
	}
	
	// Determine and log authentication method
	authMethod, err := c.setAuthMethod()
	if err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("Authentication configuration error: %v", err))
	} else {
		var authDesc string
		switch authMethod {
		case AuthMethodBasic:
			authDesc = "basic_auth (username/password)"
		case AuthMethodCert:
			authDesc = "cert_key (certificate only)"
		case AuthMethodCertWithBasic:
			authDesc = "cert_key + basic_auth (certificate with username)"
		default:
			authDesc = "none"
		}
		tflog.Debug(c.ctx, fmt.Sprintf("Using authentication method: %s", authDesc))
	}
	
	// Load client cert/key if provided (for cert-based auth)
	if c.cxProfile.ClientCertFile != "" && c.cxProfile.ClientKeyFile != "" {
		cert, err := tls.LoadX509KeyPair(c.cxProfile.ClientCertFile, c.cxProfile.ClientKeyFile)
		if err != nil {
			tflog.Error(c.ctx, fmt.Sprintf("Failed to load client certificate: %v", err))
		} else {
			tlsConfig.Certificates = []tls.Certificate{cert}
			tflog.Debug(c.ctx, "Client certificate loaded successfully")
		}
	}
	
	// Load CA cert if provided
	if c.cxProfile.CACertFile != "" {
		caCertPool := x509.NewCertPool()
		caCert, err := os.ReadFile(c.cxProfile.CACertFile)
		if err != nil {
			tflog.Error(c.ctx, fmt.Sprintf("Failed to load CA certificate: %v", err))
		} else {
			if caCertPool.AppendCertsFromPEM(caCert) {
				tlsConfig.RootCAs = caCertPool
				tflog.Debug(c.ctx, "CA certificate loaded successfully")
			} else {
				tflog.Error(c.ctx, "Failed to parse CA certificate")
			}
		}
	}
	
	return http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 120 * time.Second,
	}
}
