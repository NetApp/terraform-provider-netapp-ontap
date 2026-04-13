package gcnvclient

import (
	"context"
	"fmt"

	"golang.org/x/oauth2/google"
)

// GetAccessToken retrieves a Google Cloud access token using Application Default Credentials (ADC)
// ADC checks credentials in the following order:
// 1. GOOGLE_APPLICATION_CREDENTIALS environment variable pointing to a service account key file
// 2. gcloud auth application-default login credentials
// 3. Workload Identity (when running on GCE, GKE, or Cloud Run)
func GetAccessToken(ctx context.Context) (string, error) {
	// Use Application Default Credentials with Cloud Platform scope
	// This automatically discovers credentials from multiple sources
	creds, err := google.FindDefaultCredentials(ctx, "https://www.googleapis.com/auth/cloud-platform")
	if err != nil {
		return "", fmt.Errorf("failed to find default credentials: %w\nPlease set GOOGLE_APPLICATION_CREDENTIALS or run 'gcloud auth application-default login'", err)
	}

	// Get the access token
	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	return token.AccessToken, nil
}

// GetAccessTokenWithScopes retrieves a Google Cloud access token with custom scopes using ADC
func GetAccessTokenWithScopes(ctx context.Context, scopes ...string) (string, error) {
	if len(scopes) == 0 || (len(scopes) == 1 && scopes[0] == "https://www.googleapis.com/auth/cloud-platform") {
		return GetAccessToken(ctx)
	}

	// Use Application Default Credentials with custom scopes
	creds, err := google.FindDefaultCredentials(ctx, scopes...)
	if err != nil {
		return "", fmt.Errorf("failed to find default credentials: %w", err)
	}

	// Get the access token
	token, err := creds.TokenSource.Token()
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	return token.AccessToken, nil
}
