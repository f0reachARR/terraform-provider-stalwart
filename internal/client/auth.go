package client

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// AuthenticatedClient wraps the generated client with authentication
type AuthenticatedClient struct {
	*ClientWithResponses
	baseURL  string
	username string
	password string
	token    string
	tokenExp time.Time
}

// NewAuthenticatedClient creates a new authenticated Stalwart client
func NewAuthenticatedClient(baseURL, username, password string) (*AuthenticatedClient, error) {
	// Create the base client with authentication interceptor
	authClient := &AuthenticatedClient{
		baseURL:  baseURL,
		username: username,
		password: password,
	}

	// Create the generated client with request editor for auth
	client, err := NewClientWithResponses(
		baseURL,
		WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
		WithRequestEditorFn(authClient.addAuthHeader),
	)
	if err != nil {
		return nil, err
	}

	authClient.ClientWithResponses = client

	return authClient, nil
}

// authenticate obtains an OAuth token if needed
func (c *AuthenticatedClient) authenticate(ctx context.Context) error {
	// Check if token is still valid
	if c.token != "" && time.Now().Before(c.tokenExp) {
		return nil
	}

	// Create OAuth request using basic auth
	// Note: We don't actually need to create a separate request here,
	// we can just use the generated client directly

	// Prepare OAuth request body
	body := PostOauthJSONRequestBody{
		Type:        ptrString("code"),
		ClientId:    ptrString("webadmin"),
		RedirectUri: ptrString("stalwart://auth"),
		Nonce:       ptrString(fmt.Sprintf("%d", time.Now().Unix())),
	}

	// Make the OAuth request
	resp, err := c.PostOauthWithResponse(ctx, body, func(ctx context.Context, req *http.Request) error {
		// Use basic auth for OAuth request
		req.SetBasicAuth(c.username, c.password)
		return nil
	})
	if err != nil {
		return fmt.Errorf("authentication request failed: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("authentication failed with status %d", resp.StatusCode())
	}

	// Parse the response
	if resp.JSON200 != nil && resp.JSON200.Data != nil && resp.JSON200.Data.Code != nil {
		c.token = *resp.JSON200.Data.Code
		// Token expires in 1 hour (conservative estimate)
		c.tokenExp = time.Now().Add(50 * time.Minute)
		return nil
	}

	return fmt.Errorf("failed to extract auth token from response")
}

// addAuthHeader is a request editor that adds the auth token to requests
func (c *AuthenticatedClient) addAuthHeader(ctx context.Context, req *http.Request) error {
	// Skip auth for the OAuth endpoint itself
	if req.URL.Path == "/oauth" || req.URL.Path == "oauth" {
		return nil
	}

	// Ensure we have a valid token
	if err := c.authenticate(ctx); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Add bearer token
	req.Header.Set("Authorization", "Bearer "+c.token)
	return nil
}

// Helper function to create string pointers
func ptrString(s string) *string {
	return &s
}

// Helper function to create float32 pointers
func ptrFloat32(f float32) *float32 {
	return &f
}

// Helper function to create int pointers
func ptrInt(i int) *int {
	return &i
}
