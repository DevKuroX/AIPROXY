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

// CL4udeFlow implements OAuth2 device code flow for Claude.
// ref: open-sse/services/tokenRefresh.js:95-123
type CL4udeFlow struct {
	ClientID string
	Client   *http.Client
}

// NewCL4udeFlow creates a new Claude OAuth flow instance.
func NewCL4udeFlow(clientID string) *CL4udeFlow {
	return &CL4udeFlow{
		ClientID: clientID,
		Client:   &http.Client{Timeout: 30 * time.Second},
	}
}

const (
	claudeDeviceAuthorizeURL = "https://claude.ai/oauth/device/authorize"
	claudeTokenURL           = "https://claude.ai/oauth/token"
)

// Start initiates the device code flow by requesting a device code.
// ref: open-sse/services/tokenRefresh.js:95-108
func (f *CL4udeFlow) Start(ctx context.Context) (*DeviceCodeResponse, error) {
	reqBody := map[string]string{
		"client_id": f.ClientID,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claudeDeviceAuthorizeURL, bytes.NewReader(body))
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
// ref: open-sse/services/tokenRefresh.js:95-123
func (f *CL4udeFlow) Poll(ctx context.Context, deviceCode string) (*TokenPair, error) {
	reqBody := map[string]string{
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
		"device_code": deviceCode,
		"client_id":   f.ClientID,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claudeTokenURL, bytes.NewReader(body))
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
		return nil, fmt.Errorf("token poll failed: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	var result TokenPair
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

// Refresh obtains a new token pair using a refresh token.
// ref: open-sse/services/tokenRefresh.js:95-123
func (f *CL4udeFlow) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	reqBody := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     f.ClientID,
	}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claudeTokenURL, bytes.NewReader(body))
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
// Claude does not currently support token revocation endpoint.
func (f *CL4udeFlow) Revoke(ctx context.Context, token string) error {
	return fmt.Errorf("token revocation not supported by Claude OAuth")
}
