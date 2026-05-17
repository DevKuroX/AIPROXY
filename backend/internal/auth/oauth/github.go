package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GitHubFlow struct {
	ClientID     string
	ClientSecret string
	Scope        string
	Client       *http.Client
}

func NewGitHubFlow(clientID, clientSecret, scope string) *GitHubFlow {
	if scope == "" {
		scope = "repo,user"
	}
	return &GitHubFlow{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scope:        scope,
		Client:       &http.Client{Timeout: 30 * time.Second},
	}
}

const (
	githubDeviceAuthURL = "https://github.com/login/device/code"
	githubTokenURL      = "https://github.com/login/oauth/access_token"
)

func (f *GitHubFlow) Start(ctx context.Context) (*DeviceCodeResponse, error) {
	data := url.Values{
		"client_id": {f.ClientID},
		"scope":     {f.Scope},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubDeviceAuthURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
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

type gitHubTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
	Error       string `json:"error,omitempty"`
	ErrorDesc   string `json:"error_description,omitempty"`
}

func (f *GitHubFlow) Poll(ctx context.Context, deviceCode string) (*TokenPair, error) {
	data := url.Values{
		"client_id":   {f.ClientID},
		"device_code": {deviceCode},
		"grant_type":  {"urn:ietf:params:oauth:grant-type:device_code"},
	}

	if f.ClientSecret != "" {
		data.Set("client_secret", f.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token poll failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp gitHubTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if tokenResp.Error != "" {
		// authorization_pending, slow_down, etc. — caller should retry
		if tokenResp.Error == "authorization_pending" {
			return nil, fmt.Errorf("authorization_pending")
		}
		if tokenResp.Error == "slow_down" {
			return nil, fmt.Errorf("slow_down")
		}
		return nil, fmt.Errorf("oauth error: %s: %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response")
	}

	return &TokenPair{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: "", // GitHub device flow does not issue refresh tokens
		ExpiresIn:    28800, // GitHub tokens expire after 8 hours
	}, nil
}

func (f *GitHubFlow) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	data := url.Values{
		"client_id":     {f.ClientID},
		"refresh_token": {refreshToken},
		"grant_type":    {"refresh_token"},
	}

	if f.ClientSecret != "" {
		data.Set("client_secret", f.ClientSecret)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, githubTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var tokenResp gitHubTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("refresh error: %s: %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return &TokenPair{
		AccessToken:  tokenResp.AccessToken,
		ExpiresIn:    28800,
	}, nil
}

func (f *GitHubFlow) Revoke(ctx context.Context, token string) error {
	data := url.Values{
		"access_token": {token},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		fmt.Sprintf("https://api.github.com/applications/%s/token", f.ClientID),
		bytes.NewReader([]byte(data.Encode())),
	)
	if err != nil {
		return fmt.Errorf("create revoke request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := f.Client.Do(req)
	if err != nil {
		return fmt.Errorf("revoke request failed: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
