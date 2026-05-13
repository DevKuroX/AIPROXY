package services

import (
	"testing"

	"github.com/DevKuroX/AIPROXY/internal/models"
)

func TestDetectFormat(t *testing.T) {
	svc := NewProviderService(nil, &mockLogger{})

	tests := []struct {
		name string
		body map[string]any
		want RequestFormat
	}{
		{
			name: "openai with messages",
			body: map[string]any{
				"messages": []any{
					map[string]any{"role": "user", "content": "hello"},
				},
			},
			want: FormatOpenAI,
		},
		{
			name: "openai responses with array input",
			body: map[string]any{
				"input": []any{"hello", "world"},
			},
			want: FormatOpenAIResponses,
		},
		{
			name: "openai responses with string input",
			body: map[string]any{
				"input": "hello world",
			},
			want: FormatOpenAIResponses,
		},
		{
			name: "claude format",
			body: map[string]any{
				"messages": []any{
					map[string]any{"role": "user", "content": "hello"},
				},
				"max_tokens": 100,
			},
			want: FormatOpenAI,
		},
		{
			name: "gemini format with contents",
			body: map[string]any{
				"contents": []any{
					map[string]any{"role": "user", "parts": []any{}},
				},
			},
			want: FormatGemini,
		},
		{
			name: "antigravity format",
			body: map[string]any{
				"request": map[string]any{
					"contents": []any{},
				},
				"userAgent": "antigravity",
			},
			want: FormatAntigravity,
		},
		{
			name: "empty body defaults to openai",
			body: map[string]any{},
			want: FormatOpenAI,
		},
		{
			name: "nil body defaults to openai",
			body: nil,
			want: FormatOpenAI,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.DetectFormat(tt.body)
			if got != tt.want {
				t.Errorf("DetectFormat() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBodyHasMessages(t *testing.T) {
	tests := []struct {
		name string
		body map[string]any
		want bool
	}{
		{
			name: "has messages array",
			body: map[string]any{
				"messages": []any{map[string]any{"role": "user"}},
			},
			want: true,
		},
		{
			name: "no messages key",
			body: map[string]any{
				"input": "hello",
			},
			want: false,
		},
		{
			name: "empty messages array still returns true (key exists)",
			body: map[string]any{
				"messages": []any{},
			},
			want: true,
		},
		{
			name: "nil body",
			body: nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bodyHasMessages(tt.body)
			if got != tt.want {
				t.Errorf("bodyHasMessages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildUpstreamURL(t *testing.T) {
	svc := NewProviderService(nil, &mockLogger{})

	tests := []struct {
		name     string
		provider *models.Provider
		baseURL  string
		apiType  string
		stream   bool
		model    string
		want     string
	}{
		{
			name:     "claude provider",
			provider: &models.Provider{Type: "claude"},
			baseURL:  "https://api.anthropic.com/v1/messages",
			apiType:  "chat",
			stream:   false,
			model:    "claude-3-opus",
			want:     "https://api.anthropic.com/v1/messages?beta=true",
		},
		{
			name:     "gemini non-streaming",
			provider: &models.Provider{Type: "gemini"},
			baseURL:  "https://generativelanguage.googleapis.com/v1beta/models",
			apiType:  "chat",
			stream:   false,
			model:    "gemini-1.5-pro",
			want:     "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-pro:generateContent",
		},
		{
			name:     "gemini streaming",
			provider: &models.Provider{Type: "gemini"},
			baseURL:  "https://generativelanguage.googleapis.com/v1beta/models",
			apiType:  "chat",
			stream:   true,
			model:    "gemini-1.5-pro",
			want:     "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-pro:streamGenerateContent?alt=sse",
		},
		{
			name:     "openai chat",
			provider: &models.Provider{Type: "openai"},
			baseURL:  "https://api.openai.com/v1",
			apiType:  "chat",
			stream:   false,
			model:    "gpt-4",
			want:     "https://api.openai.com/v1/chat/completions",
		},
		{
			name:     "openai responses api",
			provider: &models.Provider{Type: "openai"},
			baseURL:  "https://api.openai.com/v1",
			apiType:  "responses",
			stream:   false,
			model:    "gpt-4o",
			want:     "https://api.openai.com/v1/responses",
		},
		{
			name:     "qwen provider",
			provider: &models.Provider{Type: "qwen"},
			baseURL:  "https://chat.qwen.ai/api/v1",
			apiType:  "chat",
			stream:   false,
			model:    "qwen-max",
			want:     "https://chat.qwen.ai/api/v1/chat/completions",
		},
		{
			name:     "antigravity streaming",
			provider: &models.Provider{Type: "antigravity"},
			baseURL:  "https://cloudcode-pa.googleapis.com",
			apiType:  "chat",
			stream:   true,
			model:    "gemini-pro",
			want:     "https://cloudcode-pa.googleapis.com/v1internal:streamGenerateContent?alt=sse",
		},
		{
			name:     "antigravity non-streaming",
			provider: &models.Provider{Type: "antigravity"},
			baseURL:  "https://cloudcode-pa.googleapis.com",
			apiType:  "chat",
			stream:   false,
			model:    "gemini-pro",
			want:     "https://cloudcode-pa.googleapis.com/v1internal:generateContent",
		},
		{
			name:     "glm claude-compatible",
			provider: &models.Provider{Type: "glm"},
			baseURL:  "https://open.bigmodel.cn/api/paas/v4/chat/completions",
			apiType:  "chat",
			stream:   false,
			model:    "glm-4",
			want:     "https://open.bigmodel.cn/api/paas/v4/chat/completions?beta=true",
		},
		{
			name:     "default provider appends chat/completions",
			provider: &models.Provider{Type: "unknown"},
			baseURL:  "https://api.example.com/v1",
			apiType:  "chat",
			stream:   false,
			model:    "model-x",
			want:     "https://api.example.com/v1/chat/completions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.BuildUpstreamURL(tt.provider, tt.baseURL, tt.apiType, tt.stream, tt.model)
			if got != tt.want {
				t.Errorf("BuildUpstreamURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildHeaders(t *testing.T) {
	svc := NewProviderService(nil, &mockLogger{})

	tests := []struct {
		name       string
		provider   *models.Provider
		creds      *ProviderCredentials
		stream     bool
		wantHeader map[string]string
	}{
		{
			name:     "claude with api key",
			provider: &models.Provider{Type: "claude"},
			creds:    &ProviderCredentials{APIKey: "sk-test", AccessToken: ""},
			stream:   false,
			wantHeader: map[string]string{
				"Content-Type":      "application/json",
				"x-api-key":         "sk-test",
				"anthropic-version": "2023-06-01",
			},
		},
		{
			name:     "claude with access token",
			provider: &models.Provider{Type: "claude"},
			creds:    &ProviderCredentials{APIKey: "", AccessToken: "token-abc"},
			stream:   false,
			wantHeader: map[string]string{
				"Content-Type":      "application/json",
				"Authorization":     "Bearer token-abc",
				"anthropic-version": "2023-06-01",
			},
		},
		{
			name:     "openai with api key",
			provider: &models.Provider{Type: "openai"},
			creds:    &ProviderCredentials{APIKey: "sk-openai"},
			stream:   false,
			wantHeader: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer sk-openai",
			},
		},
		{
			name:     "gemini with api key",
			provider: &models.Provider{Type: "gemini"},
			creds:    &ProviderCredentials{APIKey: "gemini-key"},
			stream:   false,
			wantHeader: map[string]string{
				"Content-Type":    "application/json",
				"x-goog-api-key": "gemini-key",
			},
		},
		{
			name:     "github with token",
			provider: &models.Provider{Type: "github"},
			creds:    &ProviderCredentials{CopilotToken: "ghu-xyz"},
			stream:   false,
			wantHeader: map[string]string{
				"Content-Type":      "application/json",
				"Authorization":     "Bearer ghu-xyz",
				"copilot-integration-id": "vscode-chat",
			},
		},
		{
			name:     "no credentials",
			provider: &models.Provider{Type: "openai"},
			creds:    &ProviderCredentials{},
			stream:   false,
			wantHeader: map[string]string{
				"Content-Type": "application/json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.BuildHeaders(tt.provider, tt.creds, tt.stream)

			for key, wantValue := range tt.wantHeader {
				gotValue, exists := got[key]
				if !exists {
					t.Errorf("missing header %q", key)
					continue
				}
				if gotValue != wantValue {
					t.Errorf("header %q = %q, want %q", key, gotValue, wantValue)
				}
			}
		})
	}
}

func TestRequestFormatConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant RequestFormat
		want     string
	}{
		{"FormatOpenAI", FormatOpenAI, "openai"},
		{"FormatOpenAIResponses", FormatOpenAIResponses, "openai-responses"},
		{"FormatClaude", FormatClaude, "claude"},
		{"FormatGemini", FormatGemini, "gemini"},
		{"FormatAntigravity", FormatAntigravity, "antigravity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.constant) != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.constant, tt.want)
			}
		})
	}
}

func TestProviderCredentials(t *testing.T) {
	creds := ProviderCredentials{
		APIKey:       "test-key",
		AccessToken:  "test-token",
		CopilotToken: "test-copilot",
	}

	if creds.APIKey != "test-key" {
		t.Errorf("APIKey = %q, want test-key", creds.APIKey)
	}

	if creds.AccessToken != "test-token" {
		t.Errorf("AccessToken = %q, want test-token", creds.AccessToken)
	}

	if creds.CopilotToken != "test-copilot" {
		t.Errorf("CopilotToken = %q, want test-copilot", creds.CopilotToken)
	}
}

func TestProviderConfig(t *testing.T) {
	config := ProviderConfig{
		BaseURL:      "https://api.example.com",
		Format:       "openai",
		Headers:      map[string]string{"X-Custom": "value"},
		ClientID:     "client-123",
		ClientSecret: "secret-456",
		TokenURL:     "https://auth.example.com/token",
		AuthURL:      "https://auth.example.com/authorize",
	}

	if config.BaseURL != "https://api.example.com" {
		t.Errorf("BaseURL mismatch")
	}

	if config.Format != "openai" {
		t.Errorf("Format mismatch")
	}

	if config.Headers["X-Custom"] != "value" {
		t.Errorf("Headers mismatch")
	}
}

func TestNewProviderService(t *testing.T) {
	logger := &mockLogger{}
	svc := NewProviderService(nil, logger)

	if svc == nil {
		t.Fatal("expected non-nil service")
	}

	if svc.logger == nil {
		t.Error("expected logger to be set")
	}

	if svc.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
}
