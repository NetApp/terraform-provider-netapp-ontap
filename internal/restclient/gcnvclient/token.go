package gcnvclient

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2/google"
)

// GetAccessToken retrieves a Google Cloud access token using Application Default Credentials (ADC)
// ADC checks credentials in the following order:
// 1. GOOGLE_APPLICATION_CREDENTIALS environment variable pointing to a service account key file
// 2. gcloud auth application-default login credentials
//
// If serviceAccountKeyPath is provided, it temporarily sets GOOGLE_APPLICATION_CREDENTIALS
// to that path for this specific token request.
func GetAccessToken(ctx context.Context, serviceAccountKeyPath string) (string, error) {
	// If a service account key path is provided, temporarily set it as env var
	var originalEnv string
	var hadOriginalEnv bool
	if serviceAccountKeyPath != "" {
		originalEnv, hadOriginalEnv = os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", serviceAccountKeyPath)
		// Restore original env var after function completes
		defer func() {
			if hadOriginalEnv {
				os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", originalEnv)
			} else {
				os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
			}
		}()
	}

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
// If serviceAccountKeyPath is provided, it temporarily sets GOOGLE_APPLICATION_CREDENTIALS
func GetAccessTokenWithScopes(ctx context.Context, serviceAccountKeyPath string, scopes ...string) (string, error) {
	if len(scopes) == 0 {
		return GetAccessToken(ctx, serviceAccountKeyPath)
	}

	// If a service account key path is provided, temporarily set it as env var
	var originalEnv string
	var hadOriginalEnv bool
	if serviceAccountKeyPath != "" {
		originalEnv, hadOriginalEnv = os.LookupEnv("GOOGLE_APPLICATION_CREDENTIALS")
		os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", serviceAccountKeyPath)
		defer func() {
			if hadOriginalEnv {
				os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", originalEnv)
			} else {
				os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
			}
		}()
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
