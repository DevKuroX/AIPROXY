package services

import (
	"testing"
	"time"
)

func TestTokenExpiryBuffer(t *testing.T) {
	if TokenExpiryBuffer != 5*time.Minute {
		t.Errorf("TokenExpiryBuffer = %v, want 5 minutes", TokenExpiryBuffer)
	}
}

func TestRefreshLeadTimes(t *testing.T) {
	tests := []struct {
		provider string
		want     time.Duration
	}{
		{"codex", 5 * 24 * time.Hour},
		{"claude", 4 * time.Hour},
		{"iflow", 24 * time.Hour},
		{"qwen", 20 * time.Minute},
		{"kimi-coding", 5 * time.Minute},
		{"antigravity", 5 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := RefreshLeadTimes[tt.provider]
			if got != tt.want {
				t.Errorf("RefreshLeadTimes[%q] = %v, want %v", tt.provider, got, tt.want)
			}
		})
	}
}

func TestOAuthEndpoints(t *testing.T) {
	tests := []struct {
		provider  string
		wantToken string
	}{
		{"google", "https://oauth2.googleapis.com/token"},
		{"openai", "https://auth.openai.com/oauth/token"},
		{"anthr0pic", "https://api.anthr0pic.com/v1/oauth/token"},
		{"qwen", "https://chat.qwen.ai/api/v1/oauth2/token"},
		{"iflow", "https://iflow.cn/oauth/token"},
		{"github", "https://github.com/login/oauth/access_token"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got, exists := OAuthEndpoints[tt.provider]
			if !exists {
				t.Fatalf("OAuthEndpoints[%q] does not exist", tt.provider)
			}

			if got.TokenURL != tt.wantToken {
				t.Errorf("TokenURL = %q, want %q", got.TokenURL, tt.wantToken)
			}
		})
	}
}

func TestProviderConfigs(t *testing.T) {
	tests := []struct {
		provider         string
		wantClientID     bool
		wantTokenURL     string
		wantClientSecret bool
	}{
		{"claude", true, "https://api.anthr0pic.com/v1/oauth/token", false},
		{"gemini", true, "https://oauth2.googleapis.com/token", true},
		{"gemini-cli", true, "https://oauth2.googleapis.com/token", true},
		{"antigravity", true, "https://oauth2.googleapis.com/token", true},
		{"codex", true, "https://auth.openai.com/oauth/token", false},
		{"qwen", true, "https://chat.qwen.ai/api/v1/oauth2/token", false},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got, exists := ProviderConfigs[tt.provider]
			if !exists {
				t.Fatalf("ProviderConfigs[%q] does not exist", tt.provider)
			}

			if tt.wantClientID && got.ClientID == "" {
				t.Error("expected ClientID to be set")
			}

			if got.TokenURL != tt.wantTokenURL {
				t.Errorf("TokenURL = %q, want %q", got.TokenURL, tt.wantTokenURL)
			}

			if tt.wantClientSecret && got.ClientSecret == "" {
				t.Error("expected ClientSecret to be set")
			}
		})
	}
}

func TestIsUnrecoverableRefreshError(t *testing.T) {
	tests := []struct {
		name   string
		result *RefreshResult
		want   bool
	}{
		{
			name:   "nil result",
			result: nil,
			want:   false,
		},
		{
			name:   "unrecoverable_refresh_error",
			result: &RefreshResult{Error: "unrecoverable_refresh_error"},
			want:   true,
		},
		{
			name:   "refresh_token_reused",
			result: &RefreshResult{Error: "refresh_token_reused"},
			want:   true,
		},
		{
			name:   "invalid_request",
			result: &RefreshResult{Error: "invalid_request"},
			want:   true,
		},
		{
			name:   "invalid_grant",
			result: &RefreshResult{Error: "invalid_grant"},
			want:   true,
		},
		{
			name:   "unknown error",
			result: &RefreshResult{Error: "some_other_error"},
			want:   false,
		},
		{
			name:   "empty error",
			result: &RefreshResult{Error: ""},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsUnrecoverableRefreshError(tt.result)
			if got != tt.want {
				t.Errorf("IsUnrecoverableRefreshError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRefreshResult(t *testing.T) {
	result := &RefreshResult{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
		ExpiresIn:    3600,
	}

	if result.AccessToken != "new-access-token" {
		t.Errorf("AccessToken mismatch")
	}

	if result.RefreshToken != "new-refresh-token" {
		t.Errorf("RefreshToken mismatch")
	}

	if result.ExpiresIn != 3600 {
		t.Errorf("ExpiresIn mismatch")
	}
}

func TestProviderOAuthConfig(t *testing.T) {
	config := ProviderOAuthConfig{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		TokenURL:     "https://example.com/token",
	}

	if config.ClientID != "test-client-id" {
		t.Errorf("ClientID mismatch")
	}

	if config.ClientSecret != "test-client-secret" {
		t.Errorf("ClientSecret mismatch")
	}

	if config.TokenURL != "https://example.com/token" {
		t.Errorf("TokenURL mismatch")
	}
}
