package gcnvclient

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
)

// GetAccessToken retrieves a Google Cloud access token using a service account key file
// serviceAccountKeyPath: path to the service account JSON key file (e.g., "gcp_creds.json")
func GetAccessToken(ctx context.Context, serviceAccountKeyPath string) (string, error) {
	// Read the service account key file
	data, err := os.ReadFile(serviceAccountKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read service account key file %s: %w", serviceAccountKeyPath, err)
	}

	// Create credentials from the JSON key file with Cloud Platform scope
	creds, err := google.CredentialsFromJSON(ctx, data, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("failed to parse service account key file: %w", err)
	}

	// Get the access token
	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	return token.AccessToken, nil
}

// GetAccessTokenWithScopes retrieves a Google Cloud access token with custom scopes using a service account key file
// serviceAccountKeyPath: path to the service account JSON key file (e.g., "gcp_creds.json")
func GetAccessTokenWithScopes(ctx context.Context, serviceAccountKeyPath string, scopes ...string) (string, error) {
	if len(scopes) == 0 {
		return GetAccessToken(ctx, serviceAccountKeyPath)
	}

	// Read the service account key file
	data, err := os.ReadFile(serviceAccountKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read service account key file %s: %w", serviceAccountKeyPath, err)
	}

	// Create credentials from the JSON key file with custom scopes
	creds, err := google.CredentialsFromJSON(ctx, data, scopes...)
	if err != nil {
		return "", fmt.Errorf("failed to parse service account key file: %w", err)
	}

	// Get the access token
	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	return token.AccessToken, nil
}
