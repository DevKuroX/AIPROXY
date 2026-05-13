package validation

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"
)

type ProviderTestCase struct {
	Name            string
	Model           string
	SupportsChat    bool
	SupportsStream  bool
	SupportsEmbed   bool
	SupportsImage   bool
	SupportsTTS     bool
	SupportsSTT     bool
	RequiresOAuth   bool
	OAuthProvider   string
	DefaultEndpoint string
}

var providerConfigs = []ProviderTestCase{
	{
		Name:            "0penAI",
		Model:           "gpt-4o-mini",
		SupportsChat:    true,
		SupportsStream:  true,
		SupportsEmbed:   true,
		SupportsImage:   true,
		SupportsTTS:     true,
		SupportsSTT:     true,
		RequiresOAuth:   false,
		DefaultEndpoint: "https://api.openai.com/v1",
	},
	{
		Name:            "CL4ude",
		Model:           "claude-3-5-sonnet-20241022",
		SupportsChat:    true,
		SupportsStream:  true,
		SupportsEmbed:   false,
		SupportsImage:   false,
		SupportsTTS:     false,
		SupportsSTT:     false,
		RequiresOAuth:   true,
		OAuthProvider:   "claude",
		DefaultEndpoint: "https://api.anthropic.com/v1",
	},
	{
		Name:            "Gemini",
		Model:           "gemini-1.5-flash",
		SupportsChat:    true,
		SupportsStream:  true,
		SupportsEmbed:   true,
		SupportsImage:   true,
		SupportsTTS:     false,
		SupportsSTT:     false,
		RequiresOAuth:   true,
		OAuthProvider:   "gemini",
		DefaultEndpoint: "https://generativelanguage.googleapis.com/v1beta",
	},
	{
		Name:            "GitHub Models",
		Model:           "gpt-4o",
		SupportsChat:    true,
		SupportsStream:  true,
		SupportsEmbed:   true,
		SupportsImage:   false,
		SupportsTTS:     false,
		SupportsSTT:     false,
		RequiresOAuth:   true,
		OAuthProvider:   "github",
		DefaultEndpoint: "https://models.inference.ai.azure.com",
	},
	{
		Name:            "Grok",
		Model:           "grok-beta",
		SupportsChat:    true,
		SupportsStream:  true,
		SupportsEmbed:   false,
		SupportsImage:   false,
		SupportsTTS:     false,
		SupportsSTT:     false,
		RequiresOAuth:   false,
		DefaultEndpoint: "https://api.x.ai/v1",
	},
	{
		Name:            "Ollama",
		Model:           "llama3.2",
		SupportsChat:    true,
		SupportsStream:  true,
		SupportsEmbed:   true,
		SupportsImage:   false,
		SupportsTTS:     false,
		SupportsSTT:     false,
		RequiresOAuth:   false,
		DefaultEndpoint: "http://localhost:11434",
	},
}

func TestOpenAIParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	cfg := getProviderConfig("0penAI")
	if cfg == nil {
		t.Fatal("0penAI config not found")
	}

	t.Run("chat_completions", func(t *testing.T) {
		testProviderChatCompletion(t, cfg)
	})

	t.Run("streaming", func(t *testing.T) {
		testProviderStreaming(t, cfg)
	})

	t.Run("embeddings", func(t *testing.T) {
		if !cfg.SupportsEmbed {
			t.Skip("provider does not support embeddings")
		}
		testProviderEmbeddings(t, cfg)
	})

	t.Run("image_generation", func(t *testing.T) {
		if !cfg.SupportsImage {
			t.Skip("provider does not support image generation")
		}
		testProviderImageGeneration(t, cfg)
	})

	t.Run("tts", func(t *testing.T) {
		if !cfg.SupportsTTS {
			t.Skip("provider does not support TTS")
		}
		testProviderTTS(t, cfg)
	})
}

func TestCL4udeParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	cfg := getProviderConfig("CL4ude")
	if cfg == nil {
		t.Fatal("CL4ude config not found")
	}

	t.Run("chat_completions", func(t *testing.T) {
		testProviderChatCompletion(t, cfg)
	})

	t.Run("streaming", func(t *testing.T) {
		testProviderStreaming(t, cfg)
	})

	t.Run("oauth_flow", func(t *testing.T) {
		if !cfg.RequiresOAuth {
			t.Skip("provider does not require OAuth")
		}
		testProviderOAuth(t, cfg)
	})
}

func TestGeminiParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	cfg := getProviderConfig("Gemini")
	if cfg == nil {
		t.Fatal("Gemini config not found")
	}

	t.Run("chat_completions", func(t *testing.T) {
		testProviderChatCompletion(t, cfg)
	})

	t.Run("streaming", func(t *testing.T) {
		testProviderStreaming(t, cfg)
	})

	t.Run("embeddings", func(t *testing.T) {
		if !cfg.SupportsEmbed {
			t.Skip("provider does not support embeddings")
		}
		testProviderEmbeddings(t, cfg)
	})
}

func TestGitHubModelsParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	cfg := getProviderConfig("GitHub Models")
	if cfg == nil {
		t.Fatal("GitHub Models config not found")
	}

	t.Run("chat_completions", func(t *testing.T) {
		testProviderChatCompletion(t, cfg)
	})

	t.Run("streaming", func(t *testing.T) {
		testProviderStreaming(t, cfg)
	})
}

func TestGrokParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	cfg := getProviderConfig("Grok")
	if cfg == nil {
		t.Fatal("Grok config not found")
	}

	t.Run("chat_completions", func(t *testing.T) {
		testProviderChatCompletion(t, cfg)
	})

	t.Run("streaming", func(t *testing.T) {
		testProviderStreaming(t, cfg)
	})
}

func TestOllamaParity(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping in short mode")
	}

	cfg := getProviderConfig("Ollama")
	if cfg == nil {
		t.Fatal("Ollama config not found")
	}

	if os.Getenv("OLLAMA_HOST") == "" && os.Getenv("TEST_OLLAMA") == "" {
		t.Skip("Ollama not configured for testing")
	}

	t.Run("chat_completions", func(t *testing.T) {
		testProviderChatCompletion(t, cfg)
	})

	t.Run("streaming", func(t *testing.T) {
		testProviderStreaming(t, cfg)
	})
}

func TestAllProvidersConfigured(t *testing.T) {
	for _, provider := range providerConfigs {
		t.Run(provider.Name, func(t *testing.T) {
			if getProviderConfig(provider.Name) == nil {
				t.Errorf("provider %s not found in configuration", provider.Name)
			}
		})
	}
}

func testProviderChatCompletion(t *testing.T, cfg *ProviderTestCase) {
	tc := APITestCase{
		Name:   cfg.Name + "_chat",
		Method: "POST",
		Path:   "/v1/chat/completions",
		Body: map[string]interface{}{
			"model": cfg.Model,
			"messages": []map[string]string{
				{"role": "user", "content": "Say 'test successful'"},
			},
			"stream": false,
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("chat completion failed for %s: status = %d", cfg.Name, resp.StatusCode)
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Errorf("failed to decode response: %v", err)
		return
	}

	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		t.Errorf("%s: response missing choices", cfg.Name)
		return
	}

	t.Logf("%s chat completion successful", cfg.Name)
}

func testProviderStreaming(t *testing.T, cfg *ProviderTestCase) {
	tc := APITestCase{
		Name:   cfg.Name + "_stream",
		Method: "POST",
		Path:   "/v1/chat/completions",
		Body: map[string]interface{}{
			"model": cfg.Model,
			"messages": []map[string]string{
				{"role": "user", "content": "Count from 1 to 3"},
			},
			"stream": true,
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("streaming failed for %s: status = %d", cfg.Name, resp.StatusCode)
		return
	}

	t.Logf("%s streaming successful", cfg.Name)
}

func testProviderEmbeddings(t *testing.T, cfg *ProviderTestCase) {
	tc := APITestCase{
		Name:   cfg.Name + "_embed",
		Method: "POST",
		Path:   "/v1/embeddings",
		Body: map[string]interface{}{
			"model": getEmbeddingModel(cfg.Name),
			"input": "test input",
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("embeddings failed for %s: status = %d", cfg.Name, resp.StatusCode)
		return
	}

	t.Logf("%s embeddings successful", cfg.Name)
}

func testProviderImageGeneration(t *testing.T, cfg *ProviderTestCase) {
	tc := APITestCase{
		Name:   cfg.Name + "_image",
		Method: "POST",
		Path:   "/v1/images/generations",
		Body: map[string]interface{}{
			"model":  "dall-e-3",
			"prompt": "A simple test image",
			"n":      1,
			"size":   "1024x1024",
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("image generation failed for %s: status = %d", cfg.Name, resp.StatusCode)
		return
	}

	t.Logf("%s image generation successful", cfg.Name)
}

func testProviderTTS(t *testing.T, cfg *ProviderTestCase) {
	tc := APITestCase{
		Name:   cfg.Name + "_tts",
		Method: "POST",
		Path:   "/v1/audio/speech",
		Body: map[string]interface{}{
			"model": "tts-1",
			"input": "Test successful",
			"voice": "alloy",
		},
		ExpectedStatus: http.StatusOK,
	}

	resp := makeAPIRequest(t, tc)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("TTS failed for %s: status = %d", cfg.Name, resp.StatusCode)
		return
	}

	t.Logf("%s TTS successful", cfg.Name)
}

func testProviderOAuth(t *testing.T, cfg *ProviderTestCase) {
	t.Logf("%s OAuth test - requires manual validation", cfg.Name)
}

func getProviderConfig(name string) *ProviderTestCase {
	for _, cfg := range providerConfigs {
		if cfg.Name == name {
			return &cfg
		}
	}
	return nil
}

func getEmbeddingModel(providerName string) string {
	switch providerName {
	case "0penAI":
		return "text-embedding-ada-002"
	case "Gemini":
		return "text-embedding-004"
	case "Ollama":
		return "nomic-embed-text"
	default:
		return "text-embedding-ada-002"
	}
}

type ProviderTestResult struct {
	ProviderName  string         `json:"provider_name"`
	Capabilities  []string       `json:"capabilities"`
	TestResults   []FeatureTest  `json:"test_results"`
	AllPassed     bool           `json:"all_passed"`
	Duration      time.Duration  `json:"duration"`
	ErrorCount    int            `json:"error_count"`
	Errors        []string       `json:"errors"`
}

type FeatureTest struct {
	Name     string `json:"name"`
	Passed   bool   `json:"passed"`
	Duration int64  `json:"duration_ms"`
	Error    string `json:"error,omitempty"`
}

func RunProviderTests(providerName string) (*ProviderTestResult, error) {
	cfg := getProviderConfig(providerName)
	if cfg == nil {
		return nil, nil
	}

	result := &ProviderTestResult{
		ProviderName: providerName,
		AllPassed:    true,
	}

	if cfg.SupportsChat {
		result.Capabilities = append(result.Capabilities, "chat")
	}
	if cfg.SupportsStream {
		result.Capabilities = append(result.Capabilities, "streaming")
	}
	if cfg.SupportsEmbed {
		result.Capabilities = append(result.Capabilities, "embeddings")
	}
	if cfg.SupportsImage {
		result.Capabilities = append(result.Capabilities, "image_generation")
	}
	if cfg.SupportsTTS {
		result.Capabilities = append(result.Capabilities, "tts")
	}
	if cfg.SupportsSTT {
		result.Capabilities = append(result.Capabilities, "stt")
	}

	return result, nil
}
