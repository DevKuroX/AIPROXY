// Package services provides business logic services for ai_proxy.
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/DevKuroX/AIPROXY/internal/auth/crypto"
	"github.com/DevKuroX/AIPROXY/internal/auth/oauth"
	"github.com/DevKuroX/AIPROXY/internal/config"
	"github.com/DevKuroX/AIPROXY/internal/storage"
)

// TokenExpiryBuffer is the default time before expiry to trigger refresh.
// ref: open-sse/services/tokenRefresh.js:6
const TokenExpiryBuffer = 5 * time.Minute

// Provider-specific refresh lead times.
// ref: open-sse/config/appConstants.js:149-157
var RefreshLeadTimes = map[string]time.Duration{
	"codex":       5 * 24 * time.Hour, // 5 days
	"claude":      4 * time.Hour,      // 4 hours
	"iflow":       24 * time.Hour,     // 24 hours
	"qwen":        20 * time.Minute,   // 20 minutes
	"kimi-coding": 5 * time.Minute,    // 5 minutes
	"antigravity": 5 * time.Minute,    // 5 minutes
}

// OAuth endpoints for token refresh.
// ref: open-sse/config/appConstants.js:160-186
var OAuthEndpoints = map[string]struct {
	TokenURL string
}{
	"google":    {TokenURL: "https://oauth2.googleapis.com/token"},
	"openai":    {TokenURL: "https://auth.openai.com/oauth/token"},
	"anthr0pic": {TokenURL: "https://api.anthr0pic.com/v1/oauth/token"},
	"qwen":      {TokenURL: "https://chat.qwen.ai/api/v1/oauth2/token"},
	"iflow":     {TokenURL: "https://iflow.cn/oauth/token"},
	"github":    {TokenURL: "https://github.com/login/oauth/access_token"},
}

// Provider OAuth configurations.
// ref: open-sse/config/providers.js
var ProviderConfigs = map[string]ProviderOAuthConfig{
	"claude": {
		ClientID:  "9d1c250a-e61b-44d9-88ed-5944d1962f5e",
		TokenURL:  "https://api.anthr0pic.com/v1/oauth/token",
	},
	"gemini": {
		ClientID:     "681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com",
		ClientSecret: config.GetConfig().GeminiClientSecret,
		TokenURL:     "https://oauth2.googleapis.com/token",
	},
	"gemini-cli": {
		ClientID:     "681255809395-oo8ft2oprdrnp9e3aqf6av3hmdib135j.apps.googleusercontent.com",
		ClientSecret: config.GetConfig().GeminiClientSecret,
		TokenURL:     "https://oauth2.googleapis.com/token",
	},
	"antigravity": {
		ClientID:     "1071006060591-tmhssin2h21lcre235vtolojh4g403ep.apps.googleusercontent.com",
		ClientSecret: config.GetConfig().AntigravityClientSecret,
		TokenURL:     "https://oauth2.googleapis.com/token",
	},
	"codex": {
		ClientID: "app_EMoamEEZ73f0CkXaXp7hrann",
		TokenURL: "https://auth.openai.com/oauth/token",
	},
	"qwen": {
		ClientID: "f0304373b74a44d2b584a3fb70ca9e56",
		TokenURL: "https://chat.qwen.ai/api/v1/oauth2/token",
	},
	"github": {
		ClientID: "Iv1.b507a08c87ecfe98",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
	"iflow": {
		ClientID:     "10009311001",
		ClientSecret: config.GetConfig().IflowClientSecret,
		TokenURL:     "https://iflow.cn/oauth/token",
	},
	"kiro": {
		TokenURL: "https://prod.us-east-1.auth.desktop.kiro.dev/refreshToken",
	},
}

// ProviderOAuthConfig holds OAuth configuration for a provider.
// ref: open-sse/config/providers.js
type ProviderOAuthConfig struct {
	ClientID     string
	ClientSecret string
	TokenURL     string
}

// RefreshResult represents the result of a token refresh operation.
// ref: open-sse/services/tokenRefresh.js:79-83
type RefreshResult struct {
	AccessToken          string         `json:"accessToken"`
	RefreshToken         string         `json:"refreshToken,omitempty"`
	ExpiresIn            int            `json:"expiresIn"`
	Error                string         `json:"error,omitempty"`
	Code                 string         `json:"code,omitempty"`
	ProviderSpecificData map[string]any `json:"providerSpecificData,omitempty"`
}

// pendingRefresh tracks in-flight refresh operations for deduplication.
// ref: open-sse/services/tokenRefresh.js:9-12
type pendingRefresh struct {
	wg      sync.WaitGroup
	result  *RefreshResult
	err     error
	started time.Time
}

// TokenRefreshService handles OAuth token refresh for provider accounts.
// ref: open-sse/services/tokenRefresh.js
type TokenRefreshService struct {
	storage    *storage.DB
	httpClient *http.Client
	logger     Logger

	// pending tracks in-flight refresh operations to prevent race conditions
	// that cause refresh_token_reused errors.
	// ref: open-sse/services/tokenRefresh.js:9
	mu      sync.RWMutex
	pending map[string]*pendingRefresh

	// encryptionKey is used to decrypt stored tokens
	encryptionKey []byte
}



// NewTokenRefreshService creates a new token refresh service.
func NewTokenRefreshService(store *storage.DB, encryptionKey []byte, logger Logger) *TokenRefreshService {
	return &TokenRefreshService{
		storage:       store,
		encryptionKey: encryptionKey,
		logger:        logger,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		pending: make(map[string]*pendingRefresh),
	}
}

// RefreshIfNeeded refreshes the token if it's expired or about to expire.
// Returns the current valid access token, refreshing if necessary.
// ref: open-sse/services/tokenRefresh.js:515-534
func (s *TokenRefreshService) RefreshIfNeeded(ctx context.Context, providerID, accountID string) (*RefreshResult, error) {
	// Check if refresh is needed
	needsRefresh, err := s.NeedsRefresh(ctx, providerID, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to check refresh status: %w", err)
	}

	if !needsRefresh {
		// Token is still valid, return current token
		token, err := s.getStoredToken(ctx, providerID, accountID)
		if err != nil {
			return nil, fmt.Errorf("failed to get stored token: %w", err)
		}
		return &RefreshResult{
			AccessToken:  token.AccessToken,
			RefreshToken: token.RefreshToken,
			ExpiresIn:    int(time.Until(token.ExpiresAt).Seconds()),
		}, nil
	}

	// Token needs refresh
	return s.refreshWithDedup(ctx, providerID, accountID)
}

// ForceRefresh forces a token refresh regardless of expiry.
// ref: open-sse/services/tokenRefresh.js:34-90
func (s *TokenRefreshService) ForceRefresh(ctx context.Context, providerID, accountID string) (*RefreshResult, error) {
	return s.refreshWithDedup(ctx, providerID, accountID)
}

// NeedsRefresh checks if the token needs to be refreshed.
// Returns true if the token is expired or will expire within the buffer period.
// ref: open-sse/services/tokenRefresh.js:515-534
func (s *TokenRefreshService) NeedsRefresh(ctx context.Context, providerID, accountID string) (bool, error) {
	token, err := s.getStoredToken(ctx, providerID, accountID)
	if err != nil {
		// Token not found or error - needs refresh
		return true, err
	}

	if token == nil {
		return true, nil
	}

	// Get provider-specific lead time
	providerName := s.getProviderName(providerID)
	leadTime := RefreshLeadTimes[providerName]
	if leadTime == 0 {
		leadTime = TokenExpiryBuffer
	}

	// Check if token expires within the lead time
	return time.Until(token.ExpiresAt) < leadTime, nil
}

// refreshWithDedup performs token refresh with deduplication to prevent
// race conditions that cause refresh_token_reused errors.
// ref: open-sse/services/tokenRefresh.js:515-534
func (s *TokenRefreshService) refreshWithDedup(ctx context.Context, providerID, accountID string) (*RefreshResult, error) {
	// Create cache key for dedup
	cacheKey := fmt.Sprintf("%s:%s", providerID, accountID)

	// Check for existing in-flight refresh
	s.mu.RLock()
	if pending, exists := s.pending[cacheKey]; exists {
		s.mu.RUnlock()
		s.logger.Info("TOKEN_REFRESH", "reusing in-flight refresh", "provider", providerID, "account", accountID)
		pending.wg.Wait()
		return pending.result, pending.err
	}
	s.mu.RUnlock()

	// Create new pending entry
	pending := &pendingRefresh{
		started: time.Now(),
	}
	pending.wg.Add(1)

	s.mu.Lock()
	// Double-check after acquiring write lock
	if existing, exists := s.pending[cacheKey]; exists {
		s.mu.Unlock()
		existing.wg.Wait()
		return existing.result, existing.err
	}
	s.pending[cacheKey] = pending
	s.mu.Unlock()

	// Ensure cleanup
	defer func() {
		s.mu.Lock()
		delete(s.pending, cacheKey)
		s.mu.Unlock()
		pending.wg.Done()
	}()

	// Perform the actual refresh
	result, err := s.doRefresh(ctx, providerID, accountID)
	pending.result = result
	pending.err = err

	return result, err
}

// IsUnrecoverableRefreshError checks if a refresh result indicates an
// unrecoverable error that requires re-authentication.
// ref: open-sse/services/tokenRefresh.js:15-24
func IsUnrecoverableRefreshError(result *RefreshResult) bool {
	if result == nil {
		return false
	}
	switch result.Error {
	case "unrecoverable_refresh_error",
		"refresh_token_reused",
		"invalid_request",
		"invalid_grant":
		return true
	default:
		return false
	}
}

// doRefresh performs the actual token refresh for a provider.
// ref: open-sse/services/tokenRefresh.js:536-581
func (s *TokenRefreshService) doRefresh(ctx context.Context, providerID, accountID string) (*RefreshResult, error) {
	// Get stored token
	token, err := s.getStoredToken(ctx, providerID, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stored token: %w", err)
	}

	if token == nil || token.RefreshToken == "" {
		s.logger.Warn("TOKEN_REFRESH", "no refresh token available", "provider", providerID)
		return nil, fmt.Errorf("no refresh token available")
	}

	// Get provider name and config
	providerName := s.getProviderName(providerID)
	config, ok := ProviderConfigs[providerName]
	if !ok {
		s.logger.Warn("TOKEN_REFRESH", "unsupported provider for token refresh", "provider", providerName)
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Refresh based on provider type
	var result *RefreshResult
	switch providerName {
	case "gemini", "gemini-cli", "antigravity":
		result, err = s.refreshGoogleToken(ctx, token.RefreshToken, config)
	case "claude":
		result, err = s.refreshCL4udeToken(ctx, token.RefreshToken, config)
	case "codex":
		result, err = s.refreshCodexToken(ctx, token.RefreshToken, config)
	case "qwen":
		result, err = s.refreshQwenToken(ctx, token.RefreshToken, config)
	case "github":
		result, err = s.refreshGitHubToken(ctx, token.RefreshToken, config)
	case "kiro":
		result, err = s.refreshKiroToken(ctx, token.RefreshToken, token.ProviderSpecificData)
	case "iflow":
		result, err = s.refreshIflowToken(ctx, token.RefreshToken, config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}

	if err != nil {
		return nil, err
	}

	// Save the new token
	if result != nil && result.Error == "" {
		newToken := &oauth.OAuthToken{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			ExpiresAt:    time.Now().Add(time.Duration(result.ExpiresIn) * time.Second),
			TokenType:    "Bearer",
		}
		if err := s.saveToken(ctx, providerID, accountID, newToken); err != nil {
			s.logger.Error("TOKEN_REFRESH", "failed to save refreshed token", "error", err)
			return nil, fmt.Errorf("failed to save token: %w", err)
		}
	}

	return result, nil
}

// refreshGoogleToken refreshes a Google OAuth token (Gemini, Antigravity).
// ref: open-sse/services/tokenRefresh.js:128-157
func (s *TokenRefreshService) refreshGoogleToken(ctx context.Context, refreshToken string, config ProviderOAuthConfig) (*RefreshResult, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {config.ClientID},
		"client_secret": {config.ClientSecret},
	}

	return s.doFormPost(ctx, config.TokenURL, data, nil)
}

// refreshCL4udeToken refreshes a CL4ude OAuth token.
// ref: open-sse/services/tokenRefresh.js:95-123
func (s *TokenRefreshService) refreshCL4udeToken(ctx context.Context, refreshToken string, config ProviderOAuthConfig) (*RefreshResult, error) {
	body := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"client_id":     config.ClientID,
	}

	return s.doJSONPost(ctx, config.TokenURL, body, nil)
}

// refreshCodexToken refreshes a Codex (OpenAI) OAuth token.
// OpenAI uses rotating (one-time-use) refresh tokens.
// ref: open-sse/services/tokenRefresh.js:219-282
func (s *TokenRefreshService) refreshCodexToken(ctx context.Context, refreshToken string, config ProviderOAuthConfig) (*RefreshResult, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {config.ClientID},
		"scope":         {"openid profile email offline_access"},
	}

	result, err := s.doFormPost(ctx, config.TokenURL, data, nil)
	if err != nil {
		return nil, err
	}

	// Check for unrecoverable errors (token reused/expired)
	// Auth0 revokes whole family on retry
	if result.Error != "" {
		switch result.Error {
		case "refresh_token_reused", "invalid_grant", "token_expired", "invalid_token":
			s.logger.Error("TOKEN_REFRESH", "Codex refresh token already used or invalid. Re-auth required.",
				"code", result.Code)
			return &RefreshResult{
				Error: "unrecoverable_refresh_error",
				Code:  result.Error,
			}, nil
		}
	}

	return result, nil
}

// refreshQwenToken refreshes a Qwen OAuth token.
// ref: open-sse/services/tokenRefresh.js:162-211
func (s *TokenRefreshService) refreshQwenToken(ctx context.Context, refreshToken string, config ProviderOAuthConfig) (*RefreshResult, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {config.ClientID},
	}

	result, err := s.doFormPost(ctx, config.TokenURL, data, nil)
	if err != nil {
		return nil, err
	}

	// Qwen may return resource_url in response
	if result.ProviderSpecificData == nil && result.Error == "" {
		result.ProviderSpecificData = make(map[string]any)
	}

	return result, nil
}

// refreshGitHubToken refreshes a GitHub OAuth token.
// ref: open-sse/services/tokenRefresh.js:423-464
func (s *TokenRefreshService) refreshGitHubToken(ctx context.Context, refreshToken string, config ProviderOAuthConfig) (*RefreshResult, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {config.ClientID},
	}

	// Add client_secret if available
	if config.ClientSecret != "" {
		data.Set("client_secret", config.ClientSecret)
	}

	return s.doFormPost(ctx, config.TokenURL, data, map[string]string{
		"Accept": "application/json",
	})
}

// refreshKiroToken refreshes a Kiro (AWS CodeWhisperer) token.
// Supports both AWS SSO OIDC (Builder ID/IDC) and Social Auth (Google/GitHub).
// ref: open-sse/services/tokenRefresh.js:288-373
func (s *TokenRefreshService) refreshKiroToken(ctx context.Context, refreshToken string, providerData map[string]any) (*RefreshResult, error) {
	// Check for AWS SSO OIDC (Builder ID/IDC)
	clientID, _ := providerData["clientId"].(string)
	clientSecret, _ := providerData["clientSecret"].(string)
	authMethod, _ := providerData["authMethod"].(string)
	region, _ := providerData["region"].(string)

	if clientID != "" && clientSecret != "" {
		// AWS SSO OIDC path
		isIDC := authMethod == "idc"
		endpoint := "https://oidc.us-east-1.amazonaws.com/token"
		if isIDC && region != "" {
			endpoint = fmt.Sprintf("https://oidc.%s.amazonaws.com/token", region)
		}

		body := map[string]string{
			"clientId":     clientID,
			"clientSecret": clientSecret,
			"refreshToken": refreshToken,
			"grantType":    "refresh_token",
		}

		return s.doJSONPost(ctx, endpoint, body, nil)
	}

	// Social Auth path (Google/GitHub) - use Kiro's refresh endpoint
	config := ProviderConfigs["kiro"]
	body := map[string]string{
		"refreshToken": refreshToken,
	}

	headers := map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
		"User-Agent":   "kiro-cli/1.0.0",
	}

	return s.doJSONPostWithHeaders(ctx, config.TokenURL, body, headers)
}

// refreshIflowToken refreshes an iFlow OAuth token.
// ref: open-sse/services/tokenRefresh.js:378-418
func (s *TokenRefreshService) refreshIflowToken(ctx context.Context, refreshToken string, config ProviderOAuthConfig) (*RefreshResult, error) {
	data := url.Values{
		"grant_type":    {"refresh_token"},
		"refresh_token": {refreshToken},
		"client_id":     {config.ClientID},
		"client_secret": {config.ClientSecret},
	}

	return s.doFormPost(ctx, config.TokenURL, data, nil)
}

// doFormPost performs a form-encoded POST request for token refresh.
func (s *TokenRefreshService) doFormPost(ctx context.Context, endpoint string, data url.Values, extraHeaders map[string]string) (*RefreshResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	return s.executeRefresh(req, endpoint)
}

// doJSONPost performs a JSON POST request for token refresh.
func (s *TokenRefreshService) doJSONPost(ctx context.Context, endpoint string, body map[string]string, extraHeaders map[string]string) (*RefreshResult, error) {
	return s.doJSONPostWithHeaders(ctx, endpoint, body, extraHeaders)
}

// doJSONPostWithHeaders performs a JSON POST request with custom headers.
func (s *TokenRefreshService) doJSONPostWithHeaders(ctx context.Context, endpoint string, body map[string]string, extraHeaders map[string]string) (*RefreshResult, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(jsonBody)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for k, v := range extraHeaders {
		req.Header.Set(k, v)
	}

	return s.executeRefresh(req, endpoint)
}

// executeRefresh executes the HTTP request and parses the response.
func (s *TokenRefreshService) executeRefresh(req *http.Request, endpoint string) (*RefreshResult, error) {
	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("TOKEN_REFRESH", "network error during refresh", "error", err, "endpoint", endpoint)
		return nil, fmt.Errorf("network error: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errorResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
			Code             string `json:"code"`
		}
		if err := json.Unmarshal(body, &errorResp); err == nil && errorResp.Error != "" {
			s.logger.Error("TOKEN_REFRESH", "token refresh failed",
				"status", resp.StatusCode, "error", errorResp.Error, "endpoint", endpoint)
			return &RefreshResult{
				Error: errorResp.Error,
				Code:  errorResp.Code,
			}, nil
		}

		// Try nested error format
		var nestedErrorResp struct {
			Error any `json:"error"`
		}
		if err := json.Unmarshal(body, &nestedErrorResp); err == nil {
			switch e := nestedErrorResp.Error.(type) {
			case string:
				return &RefreshResult{Error: e}, nil
			case map[string]any:
				if code, ok := e["code"].(string); ok {
					return &RefreshResult{Error: code}, nil
				}
			}
		}

		s.logger.Error("TOKEN_REFRESH", "token refresh failed",
			"status", resp.StatusCode, "body", string(body), "endpoint", endpoint)
		return nil, fmt.Errorf("refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse successful response
	var tokenResp struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiresIn    int    `json:"expires_in"`
		TokenType    string `json:"token_type"`
		ResourceURL  string `json:"resource_url"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	s.logger.Info("TOKEN_REFRESH", "successfully refreshed token",
		"hasAccessToken", tokenResp.AccessToken != "",
		"hasRefreshToken", tokenResp.RefreshToken != "",
		"expiresIn", tokenResp.ExpiresIn)

	result := &RefreshResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		ExpiresIn:    tokenResp.ExpiresIn,
	}

	if tokenResp.ResourceURL != "" {
		result.ProviderSpecificData = map[string]any{
			"resourceUrl": tokenResp.ResourceURL,
		}
	}

	return result, nil
}

// getStoredToken retrieves a stored OAuth token from the database.
func (s *TokenRefreshService) getStoredToken(ctx context.Context, providerID, accountID string) (*oauth.OAuthToken, error) {
	var (
		encryptedAccessToken  string
		encryptedRefreshToken string
		expiresAt             time.Time
	)

	err := s.storage.Pool().QueryRow(ctx,
		"SELECT encrypted_access_token, encrypted_refresh_token, expires_at FROM oauth_tokens WHERE provider_id = $1 AND account_id = $2",
		providerID, accountID,
	).Scan(&encryptedAccessToken, &encryptedRefreshToken, &expiresAt)

	if err != nil {
		return nil, err
	}

	// Decrypt tokens (using encryption key)
	accessToken, err := s.decryptToken(encryptedAccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt access token: %w", err)
	}

	refreshToken := ""
	if encryptedRefreshToken != "" {
		refreshToken, err = s.decryptToken(encryptedRefreshToken)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt refresh token: %w", err)
		}
	}

	return &oauth.OAuthToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

// saveToken saves an OAuth token to the database.
func (s *TokenRefreshService) saveToken(ctx context.Context, providerID, accountID string, token *oauth.OAuthToken) error {
	// Encrypt tokens
	encryptedAccessToken, err := s.encryptToken(token.AccessToken)
	if err != nil {
		return fmt.Errorf("failed to encrypt access token: %w", err)
	}

	var encryptedRefreshToken string
	if token.RefreshToken != "" {
		encryptedRefreshToken, err = s.encryptToken(token.RefreshToken)
		if err != nil {
			return fmt.Errorf("failed to encrypt refresh token: %w", err)
		}
	}

	// Upsert token
	_, err = s.storage.Pool().Exec(ctx,
		`INSERT INTO oauth_tokens (provider_id, account_id, encrypted_access_token, encrypted_refresh_token, expires_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, NOW())
		 ON CONFLICT (provider_id, account_id)
		 DO UPDATE SET encrypted_access_token = $3, encrypted_refresh_token = $4, expires_at = $5, updated_at = NOW()`,
		providerID, accountID, encryptedAccessToken, encryptedRefreshToken, token.ExpiresAt,
	)

	return err
}

// encryptToken encrypts a token using AES-256-GCM via the crypto package.
func (s *TokenRefreshService) encryptToken(token string) (string, error) {
	encBytes, err := crypto.Encrypt([]byte(token), s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("token encryption failed: %w", err)
	}
	return string(encBytes), nil
}

// decryptToken decrypts a token using AES-256-GCM via the crypto package.
func (s *TokenRefreshService) decryptToken(encrypted string) (string, error) {
	decBytes, err := crypto.Decrypt([]byte(encrypted), s.encryptionKey)
	if err != nil {
		return "", fmt.Errorf("token decryption failed: %w", err)
	}
	return string(decBytes), nil
}

// getProviderName maps provider ID to provider name for configuration lookup.
func (s *TokenRefreshService) getProviderName(providerID string) string {
	// This should query the providers table to get the provider name
	// For now, return a default mapping
	return providerID
}

// GetRefreshLeadTime returns the provider-specific refresh lead time.
// ref: open-sse/services/tokenRefresh.js:27-29
func GetRefreshLeadTime(provider string) time.Duration {
	if lead, ok := RefreshLeadTimes[provider]; ok {
		return lead
	}
	return TokenExpiryBuffer
}
