// Package oauth provides OAuth2 flow abstractions and types.
package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// KiroFlow implements OAuth2 device code flow for Kiro (AWS Builder ID/IDC).
// ref: 9router/src/lib/oauth/services/kiro.js:50-205
type KiroFlow struct {
	ClientID     string
	ClientSecret string
	Region       string
	Client       *http.Client
}

// NewKiroFlow creates a new Kiro OAuth flow instance.
func NewKiroFlow(clientID, clientSecret string, region string) *KiroFlow {
	if region == "" {
		region = "us-east-1"
	}
	return &KiroFlow{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Region:       region,
		Client:       &http.Client{Timeout: 30 * time.Second},
	}
}

// kiroDeviceAuthURL returns the device authorization endpoint for the configured region.
func (f *KiroFlow) kiroDeviceAuthURL() string {
	return fmt.Sprintf("https://oidc.%s.amazonaws.com/device_authorization", f.Region)
}

// kiroTokenURL returns the token endpoint for the configured region.
func (f *KiroFlow) kiroTokenURL() string {
	return fmt.Sprintf("https://oidc.%s.amazonaws.com/token", f.Region)
}

// Start initiates the device code flow by requesting a device code.
// ref: 9router/src/lib/oauth/services/kiro.js:52-81
func (f *KiroFlow) Start(ctx context.Context) (*DeviceCodeResponse, error) {
	reqBody := map[string]string{
		"clientId":     f.ClientID,
		"clientSecret": f.ClientSecret,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.kiroDeviceAuthURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("device authorize failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result DeviceCodeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// Poll checks if the user has completed the OAuth flow and returns tokens.
// ref: 9router/src/lib/oauth/services/kiro.js:86-123
func (f *KiroFlow) Poll(ctx context.Context, deviceCode string) (*TokenPair, error) {
	reqBody := map[string]string{
		"clientId":     f.ClientID,
		"clientSecret": f.ClientSecret,
		"deviceCode":   deviceCode,
		"grantType":    "urn:ietf:params:oauth:grant-type:device_code",
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.kiroTokenURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("poll token failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result TokenPair
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// Refresh obtains a new token pair using a refresh token.
// ref: 9router/src/lib/oauth/services/kiro.js:174-205
func (f *KiroFlow) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	reqBody := map[string]string{
		"clientId":     f.ClientID,
		"clientSecret": f.ClientSecret,
		"refreshToken": refreshToken,
		"grantType":    "refresh_token",
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.kiroTokenURL(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result TokenPair
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// Revoke invalidates the given token.
// ref: 9router/src/lib/oauth/services/kiro.js:125-145
func (f *KiroFlow) Revoke(ctx context.Context, token string) error {
	// AWS OIDC does not provide a standard revoke endpoint
	// Token revocation is handled server-side by token expiration
	// This is a no-op for Kiro
	return nil
}
