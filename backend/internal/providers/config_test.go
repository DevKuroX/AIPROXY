package providers

import (
	"testing"
)

func TestDefaultContextWindow(t *testing.T) {
	if DefaultContextWindow != 128000 {
		t.Fatalf("expected 128000, got %d", DefaultContextWindow)
	}
}

func TestGetContextWindowDefault(t *testing.T) {
	p := ProviderConfig{Name: "test", Type: TypeOpenAI}
	if p.GetContextWindow() != DefaultContextWindow {
		t.Fatalf("expected default %d, got %d", DefaultContextWindow, p.GetContextWindow())
	}
}

func TestGetContextWindowCustom(t *testing.T) {
	p := ProviderConfig{Name: "test", Type: TypeClaude, ContextWindow: 200000}
	if p.GetContextWindow() != 200000 {
		t.Fatalf("expected 200000, got %d", p.GetContextWindow())
	}
}

func TestGetProviderConfigExists(t *testing.T) {
	cfg, ok := GetProviderConfig("openai")
	if !ok {
		t.Fatal("openai provider should exist")
	}
	if cfg.Name != "OpenAI" {
		t.Fatalf("expected OpenAI, got %s", cfg.Name)
	}
	if cfg.Type != TypeOpenAI {
		t.Fatalf("expected TypeOpenAI, got %s", cfg.Type)
	}
}

func TestGetProviderConfigNotExists(t *testing.T) {
	_, ok := GetProviderConfig("nonexistent-provider-xyz")
	if ok {
		t.Fatal("nonexistent provider should not exist")
	}
}

func TestGetProviderConfigKiro(t *testing.T) {
	cfg, ok := GetProviderConfig("kiro")
	if !ok {
		t.Fatal("kiro provider should exist")
	}
	if cfg.Type != TypeKiro {
		t.Fatalf("expected TypeKiro, got %s", cfg.Type)
	}
	if cfg.AuthType != AuthTypeOAuth {
		t.Fatalf("expected AuthTypeOAuth, got %s", cfg.AuthType)
	}
}

func TestGetProviderConfigGeminiWeb(t *testing.T) {
	cfg, ok := GetProviderConfig("gemini-web")
	if !ok {
		t.Fatal("gemini-web provider should exist")
	}
	if cfg.AuthType != AuthTypeCookie {
		t.Fatalf("expected AuthTypeCookie, got %s", cfg.AuthType)
	}
	if cfg.Format != FormatGeminiWeb {
		t.Fatalf("expected FormatGeminiWeb, got %s", cfg.Format)
	}
}

func TestAuthTypeConstants(t *testing.T) {
	tests := []struct {
		authType string
		expected string
	}{
		{AuthTypeOAuth, "oauth"},
		{AuthTypeBearer, "bearer"},
		{AuthTypeNone, "none"},
		{AuthTypeCookie, "cookie"},
		{AuthTypeAPIKey, "apikey"},
	}
	for _, tt := range tests {
		if tt.authType != tt.expected {
			t.Fatalf("expected %s, got %s", tt.expected, tt.authType)
		}
	}
}

func TestTypeConstants(t *testing.T) {
	tests := []struct {
		typeName string
		expected string
	}{
		{TypeKiro, "kiro"},
		{TypeOpenAI, "openai"},
		{TypeClaude, "claude"},
		{TypeGemini, "gemini"},
	}
	for _, tt := range tests {
		if tt.typeName != tt.expected {
			t.Fatalf("expected %s, got %s", tt.expected, tt.typeName)
		}
	}
}

func TestProviderExists(t *testing.T) {
	// Verify key providers exist
	keyProviders := []string{
		"openai", "anthropic", "deepseek", "groq",
		"kiro", "gemini-web", "claude", "opencode",
		"deepgram", "ollama",
	}
	for _, name := range keyProviders {
		_, ok := GetProviderConfig(name)
		if !ok {
			t.Fatalf("key provider %s should exist", name)
		}
	}
}

func TestProviderConfigConsistency(t *testing.T) {
	for name, cfg := range PROVIDERS {
		if cfg.Name == "" {
			t.Fatalf("provider %s has empty Name", name)
		}
		if cfg.Type == "" {
			t.Fatalf("provider %s has empty Type", name)
		}
		if cfg.BaseURL == "" && name != "azure" {
			t.Fatalf("provider %s has empty BaseURL", name)
		}
		if cfg.AuthType == "" {
			t.Fatalf("provider %s has empty AuthType", name)
		}
	}
}

func TestClaudeAPIHeaders(t *testing.T) {
	if len(CLAUDE_API_HEADERS) == 0 {
		t.Fatal("CLAUDE_API_HEADERS should not be empty")
	}
	if _, ok := CLAUDE_API_HEADERS["Anthropic-Version"]; !ok {
		t.Fatal("CLAUDE_API_HEADERS missing Anthropic-Version")
	}
}
