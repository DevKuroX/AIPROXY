package services

import (
	"testing"
)

func TestResolveProviderAlias(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult string
	}{
		{"short alias cc", "cc", "claude"},
		{"short alias cx", "cx", "codex"},
		{"short alias gc", "gc", "gemini-cli"},
		{"short alias qw", "qw", "qwen"},
		{"short alias ag", "ag", "antigravity"},
		{"short alias gh", "gh", "github"},
		{"short alias ds", "ds", "deepseek"},
		{"full name openai", "openai", "openai"},
		{"full name deepseek", "deepseek", "deepseek"},
		{"unknown alias returns as-is", "unknown-provider", "unknown-provider"},
		{"empty string", "", ""},
		{"vertex alias vx", "vx", "vertex"},
		{"groq", "groq", "groq"},
		{"xai", "xai", "xai"},
		{"perplexity short pplx", "pplx", "perplexity"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewModelService(nil)
			got := svc.ResolveProviderAlias(tt.input)

			if got != tt.wantResult {
				t.Errorf("ResolveProviderAlias(%q) = %q, want %q", tt.input, got, tt.wantResult)
			}
		})
	}
}

func TestParseModel(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantProvider    string
		wantModel       string
		wantIsAlias     bool
		wantProviderAlias string
	}{
		{
			name:            "standard provider/model format",
			input:           "openai/gpt-4",
			wantProvider:    "openai",
			wantModel:       "gpt-4",
			wantIsAlias:     false,
			wantProviderAlias: "openai",
		},
		{
			name:            "alias/model format",
			input:           "cc/claude-3-opus",
			wantProvider:    "claude",
			wantModel:       "claude-3-opus",
			wantIsAlias:     false,
			wantProviderAlias: "cc",
		},
		{
			name:            "model only (alias format)",
			input:           "gpt-4",
			wantProvider:    "",
			wantModel:       "gpt-4",
			wantIsAlias:     true,
			wantProviderAlias: "",
		},
		{
			name:            "empty string",
			input:           "",
			wantProvider:    "",
			wantModel:       "",
			wantIsAlias:     false,
			wantProviderAlias: "",
		},
		{
			name:            "deep model path (only first slash used)",
			input:           "vertex/projects/my-project/models/gemini-pro",
			wantProvider:    "vertex",
			wantModel:       "projects/my-project/models/gemini-pro",
			wantIsAlias:     false,
			wantProviderAlias: "vertex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewModelService(nil)
			got := svc.ParseModel(tt.input)

			if got.Provider != tt.wantProvider {
				t.Errorf("Provider = %q, want %q", got.Provider, tt.wantProvider)
			}

			if got.Model != tt.wantModel {
				t.Errorf("Model = %q, want %q", got.Model, tt.wantModel)
			}

			if got.IsAlias != tt.wantIsAlias {
				t.Errorf("IsAlias = %v, want %v", got.IsAlias, tt.wantIsAlias)
			}

			if got.ProviderAlias != tt.wantProviderAlias {
				t.Errorf("ProviderAlias = %q, want %q", got.ProviderAlias, tt.wantProviderAlias)
			}
		})
	}
}

func TestInferProviderFromModelName(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantResult string
	}{
		{"claude prefix", "claude-3-opus", "anthr0pic"},
		{"claude lowercase", "CLAUDE-3-OPUS", "anthr0pic"},
		{"gemini prefix", "gemini-1.5-pro", "gemini"},
		{"gpt prefix", "gpt-4o", "openai"},
		{"o1 prefix", "o1-preview", "openai"},
		{"o3 prefix", "o3-mini", "openai"},
		{"o4 prefix", "o4-mini", "openai"},
		{"deepseek prefix", "deepseek-chat", "openrouter"},
		{"unknown defaults to openai", "unknown-model", "openai"},
		{"empty defaults to openai", "", "openai"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferProviderFromModelName(tt.input)

			if got != tt.wantResult {
				t.Errorf("inferProviderFromModelName(%q) = %q, want %q", tt.input, got, tt.wantResult)
			}
		})
	}
}

func TestResolveModelAliasFromMap(t *testing.T) {
	svc := NewModelService(nil)

	aliases := map[string]string{
		"fast": "openai/gpt-4o-mini",
		"smart": "claude/claude-3-opus",
	}

	tests := []struct {
		name         string
		alias        string
		aliases      map[string]string
		wantProvider string
		wantModel    string
		wantNil      bool
	}{
		{
			name:         "found alias with provider",
			alias:        "fast",
			aliases:      aliases,
			wantProvider: "openai",
			wantModel:    "gpt-4o-mini",
		},
		{
			name:         "found alias with resolved provider alias",
			alias:        "smart",
			aliases:      aliases,
			wantProvider: "claude",
			wantModel:    "claude-3-opus",
		},
		{
			name:    "not found",
			alias:   "unknown",
			aliases: aliases,
			wantNil: true,
		},
		{
			name:    "nil aliases map",
			alias:   "any",
			aliases: nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.ResolveModelAliasFromMap(tt.alias, tt.aliases)

			if tt.wantNil {
				if got != nil {
					t.Errorf("expected nil, got %+v", got)
				}
				return
			}

			if got == nil {
				t.Fatal("expected non-nil result")
			}

			if got.Provider != tt.wantProvider {
				t.Errorf("Provider = %q, want %q", got.Provider, tt.wantProvider)
			}

			if got.Model != tt.wantModel {
				t.Errorf("Model = %q, want %q", got.Model, tt.wantModel)
			}
		})
	}
}

func TestGetModelInfoWithAliases(t *testing.T) {
	svc := NewModelService(nil)

	aliases := map[string]string{
		"fast":  "openai/gpt-4o-mini",
		"smart": "anthr0pic/claude-3-opus",
	}

	tests := []struct {
		name         string
		modelStr     string
		aliases      map[string]string
		wantProvider string
		wantModel    string
	}{
		{
			name:         "full provider/model path",
			modelStr:     "openai/gpt-4",
			aliases:      aliases,
			wantProvider: "openai",
			wantModel:    "gpt-4",
		},
		{
			name:         "alias resolved",
			modelStr:     "fast",
			aliases:      aliases,
			wantProvider: "openai",
			wantModel:    "gpt-4o-mini",
		},
		{
			name:         "unknown alias infers openai as default",
			modelStr:     "unknown-model",
			aliases:      aliases,
			wantProvider: "openai",
			wantModel:    "unknown-model",
		},
		{
			name:         "gpt model infers openai",
			modelStr:     "gpt-4-turbo",
			aliases:      aliases,
			wantProvider: "openai",
			wantModel:    "gpt-4-turbo",
		},
		{
			name:         "claude prefix infers anthr0pic",
			modelStr:     "claude-3-opus",
			aliases:      aliases,
			wantProvider: "anthr0pic",
			wantModel:    "claude-3-opus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.GetModelInfoWithAliases(tt.modelStr, tt.aliases)

			if got == nil {
				t.Fatal("expected non-nil result")
			}

			if got.Provider != tt.wantProvider {
				t.Errorf("Provider = %q, want %q", got.Provider, tt.wantProvider)
			}

			if got.Model != tt.wantModel {
				t.Errorf("Model = %q, want %q", got.Model, tt.wantModel)
			}
		})
	}
}

func TestParseModel_PreservesOriginalInput(t *testing.T) {
	svc := NewModelService(nil)

	result := svc.ParseModel("openai/gpt-4")

	if result.Provider != "openai" {
		t.Errorf("expected provider openai, got %q", result.Provider)
	}

	if result.Model != "gpt-4" {
		t.Errorf("expected model gpt-4, got %q", result.Model)
	}
}

func TestResolveProviderAlias_AllShortAliases(t *testing.T) {
	svc := NewModelService(nil)

	shortAliases := map[string]string{
		"cc":  "claude",
		"cx":  "codex",
		"gc":  "gemini-cli",
		"qw":  "qwen",
		"if":  "iflow",
		"ag":  "antigravity",
		"gh":  "github",
		"kr":  "kiro",
		"cu":  "cursor",
		"kc":  "kilocode",
		"kmc": "kimi-coding",
		"cl":  "cline",
		"oc":  "opencode",
		"ocg": "opencode-go",
		"el":  "elevenlabs",
		"ds":  "deepseek",
		"cmc": "commandcode",
	}

	for alias, expected := range shortAliases {
		t.Run("alias_"+alias, func(t *testing.T) {
			got := svc.ResolveProviderAlias(alias)
			if got != expected {
				t.Errorf("ResolveProviderAlias(%q) = %q, want %q", alias, got, expected)
			}
		})
	}
}
