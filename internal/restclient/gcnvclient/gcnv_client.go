package gcnvclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// GCNVClient defines the client for GCNV ONTAP API
type GCNVClient struct {
	httpClient *http.Client
	ctx        context.Context
	profile    GCNVProfile
}

// GCNVProfile defines the GCNV connection profile
type GCNVProfile struct {
	ProjectID     string `mapstructure:"project_id"`
	Location      string `mapstructure:"location"`
	StoragePool   string `mapstructure:"storage_pool"`
	AuthToken     string `mapstructure:"auth_token"`
	CustomBaseUrl string `mapstructure:"custom_base_url"`
}

// NewClient creates a new GCNV client.
// Token acquisition and endpoint reachability are validated lazily on the first
// API call (Invoke), so no network round-trip is made here.
func NewClient(ctx context.Context, profile GCNVProfile) (*GCNVClient, error) {
	client := &GCNVClient{
		httpClient: &http.Client{},
		ctx:        ctx,
		profile:    profile,
	}
	return client, nil
}

// ValidateAPIEndpoint checks if the GCNV API endpoint is accessible
func (c *GCNVClient) ValidateAPIEndpoint() error {
	// Fetch token if not provided
	if c.profile.AuthToken == "" {
		tflog.Debug(c.ctx, "Validating API endpoint: fetching access token")
		token, err := GetAccessToken(c.ctx)
		if err != nil {
			return fmt.Errorf("failed to get access token: %w", err)
		}
		c.profile.AuthToken = token
	}

	// Use custom base URL directly (includes version)
	customBaseUrl := c.profile.CustomBaseUrl
	if customBaseUrl == "" {
		customBaseUrl = "https://netapp.googleapis.com/v1"
	}

	// Construct validation URL
	validationURL := fmt.Sprintf("%s/projects/%s/locations/%s/storagePools/%s/ontap/api/cluster",
		customBaseUrl,
		c.profile.ProjectID,
		c.profile.Location,
		c.profile.StoragePool,
	)

	tflog.Info(c.ctx, fmt.Sprintf("Validating GCNV API endpoint: %s (custom_base_url=%s)",
		validationURL, customBaseUrl))

	// Make a simple GET request to /cluster endpoint
	req, err := http.NewRequest("GET", validationURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.profile.AuthToken))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to connect to GCNV API endpoint: %w. Check network connectivity and firewall settings", err)
	}
	defer resp.Body.Close()

	// Read response
	responseBody, _ := io.ReadAll(resp.Body)

	// Check if we got HTML instead of JSON (indicates wrong endpoint)
	if resp.StatusCode == 404 || (resp.StatusCode >= 400 && len(responseBody) > 0 && responseBody[0] == '<') {
		return fmt.Errorf("API endpoint not found (HTTP %d). Please verify your GCNV configuration:\n"+
			"  - custom_base_url: '%s' (should be like 'https://netapp.googleapis.com/v1')\n"+
			"  - project_id: '%s'\n"+
			"  - location: '%s'\n"+
			"  - storage_pool: '%s'\n"+
			"Attempted URL: %s",
			resp.StatusCode, customBaseUrl, c.profile.ProjectID, c.profile.Location, c.profile.StoragePool, validationURL)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API endpoint returned error (HTTP %d). Check your credentials and permissions. Response: %s",
			resp.StatusCode, string(responseBody))
	}

	tflog.Info(c.ctx, "GCNV API endpoint validation successful")
	return nil
}

// Invoke sends the API Request to the GCNV ONTAP API endpoint
func (c *GCNVClient) Invoke(baseURL string, method string, body map[string]interface{}, queryValues url.Values) (int, []byte, error) {
	statusCode := -1

	// If no token is provided, fetch it automatically using Application Default Credentials (ADC)
	if c.profile.AuthToken == "" {
		tflog.Debug(c.ctx, "No auth token provided, fetching using Application Default Credentials (ADC)")

		token, err := GetAccessToken(c.ctx)
		if err != nil {
			return statusCode, nil, fmt.Errorf("failed to get access token: %w", err)
		}
		c.profile.AuthToken = token
		tflog.Debug(c.ctx, fmt.Sprintf("Successfully fetched access token, length=%d", len(token)))
	}

	// Log GCNV profile for debugging
	tflog.Debug(c.ctx, fmt.Sprintf("GCNV profile: ProjectID=%s, Location=%s, StoragePool=%s, AuthToken length=%d",
		c.profile.ProjectID,
		c.profile.Location,
		c.profile.StoragePool,
		len(c.profile.AuthToken)))

	// Use custom base URL directly (includes version)
	customBaseUrlPrefix := c.profile.CustomBaseUrl
	if customBaseUrlPrefix == "" {
		customBaseUrlPrefix = "https://netapp.googleapis.com/v1"
	}

	// Construct GCNV base URL
	gcnvBaseURL := fmt.Sprintf("%s/projects/%s/locations/%s/storagePools/%s/ontap/api",
		customBaseUrlPrefix,
		c.profile.ProjectID,
		c.profile.Location,
		c.profile.StoragePool,
	)

	tflog.Debug(c.ctx, fmt.Sprintf("GCNV base URL: %s (custom_base_url=%s)",
		gcnvBaseURL, customBaseUrlPrefix))

	// Build full URL — strip any leading/trailing slashes from baseURL to avoid malformed paths
	fullURL := fmt.Sprintf("%s/%s", gcnvBaseURL, strings.Trim(baseURL, "/"))

	// Replace 'fields' with 'ontap_fields' for GCNV
	if len(queryValues) > 0 {
		gcnvQueryValues := url.Values{}
		for key, values := range queryValues {
			if key == "fields" {
				gcnvQueryValues["ontap_fields"] = values
			} else {
				gcnvQueryValues[key] = values
			}
		}
		fullURL = fmt.Sprintf("%s?%s", fullURL, gcnvQueryValues.Encode())
	}

	// Wrap body in "body" object for POST/PATCH requests
	var requestBody io.Reader
	if len(body) > 0 && (method == "POST" || method == "PATCH") {
		// Transform body for GCNV-specific requirements
		transformedBody := c.transformRequestBody(body, method, baseURL)

		wrappedBody := map[string]interface{}{
			"body": transformedBody,
		}
		bodyJSON, err := json.Marshal(wrappedBody)
		if err != nil {
			tflog.Error(c.ctx, fmt.Sprintf("Error marshalling request body: %v", err))
			return statusCode, nil, err
		}
		tflog.Debug(c.ctx, "GCNV request JSON", map[string]any{
			"json": string(bodyJSON),
		})
		requestBody = bytes.NewReader(bodyJSON)
	}

	// Create HTTP request
	req, err := http.NewRequest(method, fullURL, requestBody)
	if err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("Error creating HTTP request: %v", err))
		return statusCode, nil, err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.profile.AuthToken))

	tflog.Debug(c.ctx, "GCNV request", map[string]any{
		"method": method,
		"url":    fullURL,
	})

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("Error sending HTTP request: %v", err))
		return statusCode, nil, err
	}
	defer resp.Body.Close()
	statusCode = resp.StatusCode

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		tflog.Error(c.ctx, fmt.Sprintf("Error reading response body: %v", err))
		return statusCode, nil, err
	}

	tflog.Debug(c.ctx, "GCNV response", map[string]any{
		"statusCode": statusCode,
		"body":       string(responseBody),
	})

	return statusCode, responseBody, nil
}

// transformRequestBody transforms the ONTAP API request body to GCNV format
func (c *GCNVClient) transformRequestBody(body map[string]interface{}, method, baseURL string) map[string]interface{} {
	// Only transform for volume creation/updates
	if !strings.Contains(baseURL, "storage/volumes") {
		tflog.Debug(c.ctx, "Skipping transformation - not a volume request", map[string]any{
			"baseURL": baseURL,
		})
		return body
	}

	tflog.Debug(c.ctx, "Starting body transformation for GCNV", map[string]any{
		"baseURL": baseURL,
	})

	transformedBody := make(map[string]interface{})
	for k, v := range body {
		transformedBody[k] = v
	}

	// Extract size from space.size and convert to string format
	if space, ok := body["space"].(map[string]interface{}); ok {
		// Copy the space map so we can mutate it (drop space.size, set logical_space
		// fields) without modifying the caller's body.
		spaceCopy := make(map[string]interface{}, len(space))
		for k, v := range space {
			spaceCopy[k] = v
		}
		transformedBody["space"] = spaceCopy

		tflog.Debug(c.ctx, "Found space object", map[string]any{
			"space": spaceCopy,
		})

		var sizeStr string
		sizeConverted := false
		switch sizeBytes := space["size"].(type) {
		case float64:
			sizeStr = formatSize(int64(sizeBytes))
			sizeConverted = true
			tflog.Debug(c.ctx, "Transformed size for GCNV", map[string]any{
				"originalBytes": sizeBytes,
				"gcnvSize":      sizeStr,
			})
		case int64:
			sizeStr = formatSize(sizeBytes)
			sizeConverted = true
			tflog.Debug(c.ctx, "Transformed size for GCNV (int64)", map[string]any{
				"originalBytes": sizeBytes,
				"gcnvSize":      sizeStr,
			})
		case int:
			sizeStr = formatSize(int64(sizeBytes))
			sizeConverted = true
			tflog.Debug(c.ctx, "Transformed size for GCNV (int)", map[string]any{
				"originalBytes": sizeBytes,
				"gcnvSize":      sizeStr,
			})
		default:
			if _, present := space["size"]; present {
				tflog.Debug(c.ctx, "Size field type not recognized", map[string]any{
					"sizeValue": space["size"],
					"sizeType":  fmt.Sprintf("%T", space["size"]),
				})
			}
		}

		if sizeConverted {
			transformedBody["size"] = sizeStr
			// GCNV rejects requests that contain both top-level `size` and
			// `space.size` ("cannot specify both 'size' and 'space.size'").
			delete(spaceCopy, "size")
		}

		// GCNV requires logical_space.enforcement and reporting to be true
		if logicalSpace, ok := spaceCopy["logical_space"].(map[string]interface{}); ok {
			logicalSpaceCopy := make(map[string]interface{}, len(logicalSpace))
			for k, v := range logicalSpace {
				logicalSpaceCopy[k] = v
			}
			logicalSpaceCopy["enforcement"] = true
			logicalSpaceCopy["reporting"] = true
			spaceCopy["logical_space"] = logicalSpaceCopy
			tflog.Debug(c.ctx, "Set logical_space.enforcement and reporting to true for GCNV")
		}
	} else {
		tflog.Debug(c.ctx, "No space object found in body")
	}

	// GCNV only supports autosize.mode = "off". Remove the autosize block
	// entirely when a non-"off" mode is present to avoid a 400 INVALID_ARGUMENT
	// error ("field 'autosize.mode' must be one of [off]").
	if autosize, ok := transformedBody["autosize"].(map[string]interface{}); ok {
		if mode, modePresent := autosize["mode"]; modePresent {
			if modeStr, _ := mode.(string); modeStr != "off" && modeStr != "" {
				delete(transformedBody, "autosize")
				tflog.Debug(c.ctx, "Removed autosize block for GCNV (mode not supported)", map[string]any{
					"mode": modeStr,
				})
			}
		}
	}

	return transformedBody
}

// formatSize converts bytes to a human-readable string with appropriate unit
func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)

	switch {
	case bytes >= TB && bytes%TB == 0:
		return fmt.Sprintf("%dTB", bytes/TB)
	case bytes >= GB && bytes%GB == 0:
		return fmt.Sprintf("%dGB", bytes/GB)
	case bytes >= MB && bytes%MB == 0:
		return fmt.Sprintf("%dMB", bytes/MB)
	case bytes >= KB && bytes%KB == 0:
		return fmt.Sprintf("%dKB", bytes/KB)
	default:
		// Default to MB for partial values
		return fmt.Sprintf("%dMB", bytes/MB)
	}
}
