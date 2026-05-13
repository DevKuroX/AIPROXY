// Package oauth provides OAuth2 flow abstractions and types.
package oauth

import (
	"time"
)

// OAuthToken represents an OAuth2 token with metadata.
// ref: open-sse/services/tokenRefresh.js:71-83
type OAuthToken struct {
	AccessToken           string                 `json:"access_token"`
	RefreshToken          string                 `json:"refresh_token,omitempty"`
	ExpiresAt             time.Time              `json:"expires_at"`
	TokenType             string                 `json:"token_type,omitempty"`
	Scope                 string                 `json:"scope,omitempty"`
	ProviderSpecificData  map[string]interface{} `json:"provider_specific_data,omitempty"`
	CreatedAt             time.Time              `json:"created_at"`
	UpdatedAt             time.Time              `json:"updated_at"`
}

// DeviceCodeResponse represents the response from a device authorization request.
// Used for OAuth2 device flow (e.g., GitHub, Google device auth).
type DeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval,omitempty"`
}

// TokenPair represents an access/refresh token pair returned from OAuth flows.
// ref: open-sse/services/tokenRefresh.js:79-83
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn    int    `json:"expires_in"`
}

// OAuthAccount represents an OAuth-linked account with encrypted token storage.
type OAuthAccount struct {
	ProviderID string `json:"provider_id"`
	AccountID  string `json:"account_id"`
	Email      string `json:"email,omitempty"`
	Token      []byte `json:"token"` // Encrypted token data
}

// RefreshError represents an error from a token refresh operation.
// ref: open-sse/services/tokenRefresh.js:15-24
type RefreshError struct {
	ErrorType string `json:"error"`
	Message   string `json:"error_description,omitempty"`
}

// IsUnrecoverable checks if the refresh error is unrecoverable.
// ref: open-sse/services/tokenRefresh.js:15-24
func (e *RefreshError) IsUnrecoverable() bool {
	if e == nil {
		return false
	}
	switch e.ErrorType {
	case "unrecoverable_refresh_error",
		"refresh_token_reused",
		"invalid_request",
		"invalid_grant":
		return true
	default:
		return false
	}
}

// Error implements the error interface.
func (e *RefreshError) Error() string {
	if e.Message != "" {
		return e.ErrorType + ": " + e.Message
	}
	return e.ErrorType
}


