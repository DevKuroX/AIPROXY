// Package executor provides provider-specific request/response handling.
// ref: _ref/9router/open-sse/executors/gemini-cli.js
package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strings"
)

// Gemini CLI constants.
// ref: open-sse/config/appConstants.js:4-5
const (
	GeminiCLIVersion   = "0.31.0"
	GeminiCLIClientAPI = "google-genai-sdk/1.41.0 gl-node/v22.19.0"
)

// geminiCLIOAuthTokenEndpoint is the Google OAuth token endpoint.
// ref: open-sse/config/appConstants.js:161-163
const geminiCLIOAuthTokenEndpoint = "https://oauth2.googleapis.com/token"

// GeminiCLIExecutor implements the Executor interface for Gemini CLI API.
// ref: open-sse/executors/gemini-cli.js:5-8
type GeminiCLIExecutor struct {
	BaseExecutor
	config        *GeminiCLIConfig
	currentModel  string
	clientID      string
	clientSecret  string
}

// GeminiCLIConfig holds Gemini CLI-specific configuration.
// ref: open-sse/config/providers.js:61-66
type GeminiCLIConfig struct {
	BaseURL     string `json:"baseUrl"`
	ClientID    string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

// GeminiCLICredentials holds Gemini CLI authentication data.
// ref: open-sse/executors/gemini-cli.js:54-59
type GeminiCLICredentials struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
	ProjectID    string `json:"projectId"`
}

// GeminiCLITokenResponse represents the OAuth token response from Google.
type GeminiCLITokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// NewGeminiCLIExecutor creates a new Gemini CLI executor.
// ref: open-sse/executors/gemini-cli.js:6-8
func NewGeminiCLIExecutor(config *GeminiCLIConfig) *GeminiCLIExecutor {
	if config == nil {
		config = &GeminiCLIConfig{
			BaseURL:      "https://cloudcode-pa.googleapis.com/v1internal",
			ClientID:     "GOOGLE_OAUTH_CLIENT_ID",
			ClientSecret: "GOOGLE_OAUTH_CLIENT_SECRET",
		}
	}
	return &GeminiCLIExecutor{
		BaseExecutor:  NewBaseExecutor("gemini-cli"),
		config:        config,
		clientID:      config.ClientID,
		clientSecret:  config.ClientSecret,
	}
}

// geminiCLIUserAgent builds the User-Agent header for Gemini CLI.
// ref: open-sse/config/appConstants.js:7-10
func geminiCLIUserAgent(model string) string {
	os := runtime.GOOS
	if os == "windows" {
		os = "windows"
	}
	arch := runtime.GOARCH
	if model == "" {
		model = "unknown"
	}
	return fmt.Sprintf("GeminiCLI/%s/%s (%s; %s)", GeminiCLIVersion, model, os, arch)
}

// PrepareRequest modifies the outgoing request for Gemini CLI.
// ref: open-sse/executors/gemini-cli.js:10-31
func (e *GeminiCLIExecutor) PrepareRequest(ctx context.Context, req *http.Request, body []byte) error {
	// Parse model from path or body for User-Agent
	// ref: open-sse/executors/gemini-cli.js:25-27
	model := e.extractModel(req, body)
	e.currentModel = model

	// Build URL with action suffix
	// ref: open-sse/executors/gemini-cli.js:10-13
	stream := strings.Contains(req.Header.Get("Accept"), "text/event-stream")
	e.buildURL(req, stream)

	// Extract credentials from context or request headers
	// ref: open-sse/executors/gemini-cli.js:15-22
	creds := e.extractCredentials(req)
	if creds != nil && creds.AccessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", creds.AccessToken))
	}

	// Set Gemini CLI specific headers
	// ref: open-sse/executors/gemini-cli.js:17-21
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", geminiCLIUserAgent(model))
	req.Header.Set("X-Goog-Api-Client", GeminiCLIClientAPI)

	if stream {
		req.Header.Set("Accept", "text/event-stream")
	} else {
		req.Header.Set("Accept", "application/json")
	}

	// Transform request body to add project if missing
	// ref: open-sse/executors/gemini-cli.js:28-30
	if len(body) > 0 && creds != nil && creds.ProjectID != "" {
		transformed := e.transformRequestBody(body, creds.ProjectID)
		if transformed != nil {
			// Update request body would require creating a new request
			// For now, this is handled at the proxy level
		}
	}

	return nil
}

// buildURL constructs the Gemini CLI API URL.
// ref: open-sse/executors/gemini-cli.js:10-13
func (e *GeminiCLIExecutor) buildURL(req *http.Request, stream bool) {
	var action string
	if stream {
		action = "streamGenerateContent?alt=sse"
	} else {
		action = "generateContent"
	}
	
	// Update URL path to include action
	// Original: https://cloudcode-pa.googleapis.com/v1internal
	// With action: https://cloudcode-pa.googleapis.com/v1internal:generateContent
	if !strings.HasSuffix(req.URL.Path, action) && !strings.Contains(req.URL.Path, ":") {
		req.URL.Path = req.URL.Path + ":" + action
	}
}

// extractModel extracts the model name from request path or body.
func (e *GeminiCLIExecutor) extractModel(req *http.Request, body []byte) string {
	// Try to extract from path first
	path := req.URL.Path
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		modelPart := path[idx+1:]
		// Remove action suffix if present
		if colonIdx := strings.Index(modelPart, ":"); colonIdx >= 0 {
			return modelPart[:colonIdx]
		}
		return modelPart
	}

	// Try to extract from body
	if len(body) > 0 {
		var bodyMap map[string]interface{}
		if err := json.Unmarshal(body, &bodyMap); err == nil {
			if model, ok := bodyMap["model"].(string); ok {
				return model
			}
		}
	}

	return "unknown"
}

// extractCredentials extracts Gemini CLI credentials from request headers.
func (e *GeminiCLIExecutor) extractCredentials(req *http.Request) *GeminiCLICredentials {
	// For now, extract from Authorization header
	// Full implementation would need context-based credential management
	auth := req.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return &GeminiCLICredentials{
			AccessToken: strings.TrimPrefix(auth, "Bearer "),
		}
	}
	return nil
}

// transformRequestBody adds project ID to the request body if missing.
// ref: open-sse/executors/gemini-cli.js:28-30
func (e *GeminiCLIExecutor) transformRequestBody(body []byte, projectID string) []byte {
	var bodyMap map[string]interface{}
	if err := json.Unmarshal(body, &bodyMap); err != nil {
		return body
	}

	if _, ok := bodyMap["project"]; !ok {
		bodyMap["project"] = projectID
		transformed, err := json.Marshal(bodyMap)
		if err != nil {
			return body
		}
		return transformed
	}
	return body
}

// TransformResponse reads and returns the response body unchanged.
func (e *GeminiCLIExecutor) TransformResponse(ctx context.Context, resp *http.Response) ([]byte, error) {
	return io.ReadAll(resp.Body)
}

// HandleError returns the error unchanged.
func (e *GeminiCLIExecutor) HandleError(ctx context.Context, err error) error {
	return err
}

// RefreshCredentials refreshes the Gemini CLI OAuth token.
// ref: open-sse/executors/gemini-cli.js:34-64
func (e *GeminiCLIExecutor) RefreshCredentials(ctx context.Context, creds *GeminiCLICredentials) (*GeminiCLICredentials, error) {
	if creds == nil || creds.RefreshToken == "" {
		return nil, fmt.Errorf("no refresh token available")
	}

	// Build form data
	// ref: open-sse/executors/gemini-cli.js:41-46
	formData := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {creds.RefreshToken},
		"client_id":     {e.clientID},
		"client_secret": {e.clientSecret},
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", geminiCLIOAuthTokenEndpoint, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token refresh request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status: %d", resp.StatusCode)
	}

	// Parse response
	// ref: open-sse/executors/gemini-cli.js:51-59
	var tokenResp GeminiCLITokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	newCreds := &GeminiCLICredentials{
		AccessToken:  tokenResp.AccessToken,
		ExpiresIn:    tokenResp.ExpiresIn,
		ProjectID:    creds.ProjectID,
	}

	// Use new refresh token if provided, otherwise keep the old one
	// ref: open-sse/executors/gemini-cli.js:56
	if tokenResp.RefreshToken != "" {
		newCreds.RefreshToken = tokenResp.RefreshToken
	} else {
		newCreds.RefreshToken = creds.RefreshToken
	}

	return newCreds, nil
}
