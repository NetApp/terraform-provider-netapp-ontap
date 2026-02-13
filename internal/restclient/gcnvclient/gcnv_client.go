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
	ProjectID             string `mapstructure:"project_id"`
	Location              string `mapstructure:"location"`
	StoragePool           string `mapstructure:"storage_pool"`
	AuthToken             string `mapstructure:"auth_token"`
	ServiceAccountKeyPath string `mapstructure:"service_account_key_path"`
}

// NewClient creates a new GCNV client
func NewClient(ctx context.Context, profile GCNVProfile) (*GCNVClient, error) {
	return &GCNVClient{
		httpClient: &http.Client{},
		ctx:        ctx,
		profile:    profile,
	}, nil
}

// Invoke sends the API Request to the GCNV ONTAP API endpoint
func (c *GCNVClient) Invoke(baseURL string, method string, body map[string]interface{}, queryValues url.Values) (int, []byte, error) {
	statusCode := -1

	// If no token is provided, fetch it automatically using service account key file
	if c.profile.AuthToken == "" {
		serviceAccountKeyPath := c.profile.ServiceAccountKeyPath
		if serviceAccountKeyPath == "" {
			return statusCode, nil, fmt.Errorf("no auth token or service account key path provided")
		}
		tflog.Debug(c.ctx, fmt.Sprintf("No auth token provided, fetching using service account key file: %s", serviceAccountKeyPath))
		token, err := GetAccessToken(c.ctx, serviceAccountKeyPath)
		if err != nil {
			return statusCode, nil, fmt.Errorf("failed to get access token from %s: %w", serviceAccountKeyPath, err)
		}
		c.profile.AuthToken = token
		tflog.Debug(c.ctx, fmt.Sprintf("Successfully fetched access token from %s, length=%d", serviceAccountKeyPath, len(token)))
	}

	// Log GCNV profile for debugging
	tflog.Debug(c.ctx, fmt.Sprintf("GCNV profile: ProjectID=%s, Location=%s, StoragePool=%s, AuthToken length=%d",
		c.profile.ProjectID,
		c.profile.Location,
		c.profile.StoragePool,
		len(c.profile.AuthToken)))

	// Construct GCNV base URL
	gcnvBaseURL := fmt.Sprintf("https://autopush-netapp.sandbox.googleapis.com/v1beta1/projects/%s/locations/%s/storagePools/%s/ontap/api",
		c.profile.ProjectID,
		c.profile.Location,
		c.profile.StoragePool,
	)

	// Build full URL
	fullURL := fmt.Sprintf("%s/%s", gcnvBaseURL, baseURL)

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
		tflog.Debug(c.ctx, "Found space object", map[string]any{
			"space": space,
		})
		if sizeBytes, ok := space["size"].(float64); ok {
			// Convert bytes to appropriate unit (MB, GB, TB)
			sizeStr := formatSize(int64(sizeBytes))
			transformedBody["size"] = sizeStr
			tflog.Debug(c.ctx, "Transformed size for GCNV", map[string]any{
				"originalBytes": sizeBytes,
				"gcnvSize":      sizeStr,
			})
		} else if sizeBytes, ok := space["size"].(int64); ok {
			sizeStr := formatSize(sizeBytes)
			transformedBody["size"] = sizeStr
			tflog.Debug(c.ctx, "Transformed size for GCNV (int64)", map[string]any{
				"originalBytes": sizeBytes,
				"gcnvSize":      sizeStr,
			})
		} else if sizeBytes, ok := space["size"].(int); ok {
			sizeStr := formatSize(int64(sizeBytes))
			transformedBody["size"] = sizeStr
			tflog.Debug(c.ctx, "Transformed size for GCNV (int)", map[string]any{
				"originalBytes": sizeBytes,
				"gcnvSize":      sizeStr,
			})
		} else {
			tflog.Debug(c.ctx, "Size field type not recognized", map[string]any{
				"sizeValue": space["size"],
				"sizeType":  fmt.Sprintf("%T", space["size"]),
			})
		}

		// GCNV requires logical_space.enforcement and reporting to be true
		if logicalSpace, ok := space["logical_space"].(map[string]interface{}); ok {
			logicalSpace["enforcement"] = true
			logicalSpace["reporting"] = true
			tflog.Debug(c.ctx, "Set logical_space.enforcement and reporting to true for GCNV")
		}
	} else {
		tflog.Debug(c.ctx, "No space object found in body")
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
