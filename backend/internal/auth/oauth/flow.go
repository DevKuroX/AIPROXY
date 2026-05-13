package oauth

import "context"

// ProviderConfig holds OAuth2 provider configuration.
// ref: open-sse/services/tokenRefresh.js:35-36
type ProviderConfig struct {
	ClientID      string
	ClientSecret  string
	AuthURL       string
	TokenURL      string
	DeviceAuthURL string
	RevokeURL     string
}

type Flow interface {
	Start(ctx context.Context) (*DeviceCodeResponse, error)

	Poll(ctx context.Context, deviceCode string) (*TokenPair, error)

	Refresh(ctx context.Context, refreshToken string) (*TokenPair, error)

	Revoke(ctx context.Context, token string) error
}
