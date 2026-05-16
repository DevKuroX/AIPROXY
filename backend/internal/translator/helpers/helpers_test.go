package helpers

import (
	"testing"
)

func TestClaudeFormatProviders(t *testing.T) {
	if !ClaudeFormatProvidersWithoutOutputConfig["minimax"] {
		t.Fatal("minimax should be in ClaudeFormatProvidersWithoutOutputConfig")
	}
	if !ClaudeFormatProvidersWithoutOutputConfig["minimax-cn"] {
		t.Fatal("minimax-cn should be in ClaudeFormatProvidersWithoutOutputConfig")
	}
	if len(ClaudeFormatProvidersWithoutOutputConfig) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(ClaudeFormatProvidersWithoutOutputConfig))
	}
}

func TestHasValidContent(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		expected bool
	}{
		{
			name: "empty content",
			input: map[string]interface{}{
				"content": "",
			},
			expected: false,
		},
		{
			name: "array content",
			input: map[string]interface{}{
				"content": []interface{}{map[string]interface{}{"type": "text", "text": "hello"}},
			},
			expected: true,
		},
		{
			name: "nil content",
			input: map[string]interface{}{
				"content": nil,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasValidContent(tt.input)
			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestExtractTextContent(t *testing.T) {
	result := ExtractTextContent("hello world")
	if result != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", result)
	}

	result2 := ExtractTextContent([]interface{}{
		map[string]interface{}{"type": "text", "text": "hello"},
		map[string]interface{}{"type": "image", "image_url": "http://img"}},
	)
	if result2 != "hello" {
		t.Fatalf("expected 'hello', got '%s'", result2)
	}

	result3 := ExtractTextContent(nil)
	if result3 != "" {
		t.Fatalf("expected empty, got '%s'", result3)
	}
}

func TestConvertOpenAIContentToGeminiParts(t *testing.T) {
	result := Convert0penAIContentToGeminiParts("simple text")
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
	if result[0]["text"] != "simple text" {
		t.Fatalf("expected 'simple text', got '%v'", result[0]["text"])
	}
}

func TestGeminiSafetySetting(t *testing.T) {
	s := GeminiSafetySetting{
		Category:  "HARM_CATEGORY_HARASSMENT",
		Threshold: "BLOCK_NONE",
	}
	if s.Category != "HARM_CATEGORY_HARASSMENT" {
		t.Fatalf("expected HARM_CATEGORY_HARASSMENT, got %s", s.Category)
	}
}
