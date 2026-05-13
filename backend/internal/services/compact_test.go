package services

import (
	"testing"
)

type mockLogger struct{}

func (m *mockLogger) Debug(msg string, args ...any) {}
func (m *mockLogger) Info(msg string, args ...any)  {}
func (m *mockLogger) Warn(msg string, args ...any)  {}
func (m *mockLogger) Error(msg string, args ...any) {}

func TestNewCompactService(t *testing.T) {
	logger := &mockLogger{}
	svc := NewCompactService(logger)

	if svc == nil {
		t.Error("expected non-nil service")
	}
}

func TestCompactResponse_Nil(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	result := svc.CompactResponse(nil)

	if result != nil {
		t.Errorf("expected nil for nil input, got %v", result)
	}
}

func TestCompactResponse_Empty(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	input := map[string]any{}
	result := svc.CompactResponse(input)

	if result == nil {
		t.Error("expected non-nil result for empty input")
	}

	if len(result) != 0 {
		t.Errorf("expected empty result, got %v", result)
	}
}

func TestCompactResponse_ChoicesWithMessage(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	input := map[string]any{
		"id": "test-id",
		"choices": []any{
			map[string]any{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": "Hello world",
				},
			},
		},
	}

	result := svc.CompactResponse(input)

	if result["id"] != "test-id" {
		t.Errorf("expected id to be preserved, got %v", result["id"])
	}

	choices, ok := result["choices"].([]any)
	if !ok {
		t.Fatal("expected choices to be []any")
	}

	if len(choices) != 1 {
		t.Errorf("expected 1 choice, got %d", len(choices))
	}
}

func TestCompactResponse_ThinkingBlocksRemoved(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	input := map[string]any{
		"choices": []any{
			map[string]any{
				"message": map[string]any{
					"content": "<|thinking|>internal thoughts<|/thinking|>Hello world",
				},
			},
		},
	}

	result := svc.CompactResponse(input)

	choices := result["choices"].([]any)
	choice := choices[0].(map[string]any)
	message := choice["message"].(map[string]any)
	content := message["content"].(string)

	if content != "Hello world" {
		t.Errorf("expected thinking block removed, got %q", content)
	}
}

func TestCompactResponse_StreamingDelta(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	input := map[string]any{
		"choices": []any{
			map[string]any{
				"delta": map[string]any{
					"content": "streaming content",
				},
			},
		},
	}

	result := svc.CompactResponse(input)

	choices := result["choices"].([]any)
	choice := choices[0].(map[string]any)
	delta := choice["delta"].(map[string]any)

	if delta["content"] != "streaming content" {
		t.Errorf("expected content preserved, got %v", delta["content"])
	}
}

func TestCompactResponse_UsageCompacted(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	input := map[string]any{
		"usage": map[string]any{
			"prompt_tokens":     100,
			"completion_tokens": 50,
			"total_tokens":      150,
			"extra_field":       "should be removed",
		},
	}

	result := svc.CompactResponse(input)

	usage, ok := result["usage"].(map[string]any)
	if !ok {
		t.Fatal("expected usage to be map[string]any")
	}

	if usage["prompt_tokens"] != 100 {
		t.Errorf("expected prompt_tokens preserved, got %v", usage["prompt_tokens"])
	}

	if usage["completion_tokens"] != 50 {
		t.Errorf("expected completion_tokens preserved, got %v", usage["completion_tokens"])
	}

	if _, exists := usage["extra_field"]; exists {
		t.Error("expected extra_field to be removed")
	}
}

func TestCompactResponse_ResponsesAPI(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	input := map[string]any{
		"output": []any{
			map[string]any{
				"type": "message",
				"content": []any{
					map[string]any{
						"type": "text",
						"text": "response text",
					},
				},
			},
		},
	}

	result := svc.CompactResponse(input)

	output, ok := result["output"].([]any)
	if !ok {
		t.Fatal("expected output to be []any")
	}

	if len(output) != 1 {
		t.Errorf("expected 1 output item, got %d", len(output))
	}
}

func TestRemoveThinking(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty string",
			input: "",
			want:  "",
		},
		{
			name:  "no thinking blocks",
			input: "Hello world",
			want:  "Hello world",
		},
		{
			name:  "thinking block removed",
			input: "<|thinking|>internal<|/thinking|>Hello",
			want:  "Hello",
		},
		{
			name:  "xml style thinking removed",
			input: "<thinking>internal</thinking>Hello",
			want:  "Hello",
		},
		{
			name:  "codex style thinking removed",
			input: "<|start|>thinking<|message|>internal<|end|>Hello",
			want:  "Hello",
		},
		{
			name:  "redacted thinking removed",
			input: "<|redacted_thinking|>hidden<|/redacted_thinking|>Hello",
			want:  "Hello",
		},
		{
			name:  "multiple thinking blocks",
			input: "<|thinking|>first<|/thinking|>Hello<|thinking|>second<|/thinking|>World",
			want:  "HelloWorld",
		},
		{
			name:  "thinking at end",
			input: "Hello<|thinking|>trailing<|/thinking|>",
			want:  "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.RemoveThinking(tt.input)

			if got != tt.want {
				t.Errorf("RemoveThinking(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCompactUsage(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	tests := []struct {
		name       string
		input      map[string]any
		wantFields []string
	}{
		{
			name:       "nil input",
			input:      nil,
			wantFields: []string{},
		},
		{
			name: "all essential fields preserved",
			input: map[string]any{
				"prompt_tokens":     100,
				"completion_tokens": 50,
				"total_tokens":      150,
			},
			wantFields: []string{"prompt_tokens", "completion_tokens", "total_tokens"},
		},
		{
			name: "responses api format",
			input: map[string]any{
				"input_tokens":  100,
				"output_tokens": 50,
			},
			wantFields: []string{"input_tokens", "output_tokens"},
		},
		{
			name: "non-essential fields removed",
			input: map[string]any{
				"prompt_tokens": 100,
				"cache_read":    50,
				"cache_write":   25,
			},
			wantFields: []string{"prompt_tokens"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.CompactUsage(tt.input)

			if tt.input == nil {
				if got != nil {
					t.Errorf("expected nil for nil input, got %v", got)
				}
				return
			}

			for _, field := range tt.wantFields {
				if _, exists := got[field]; !exists {
					t.Errorf("expected field %q to exist", field)
				}
			}
		})
	}
}

func TestCompactStreamingLine(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "empty line",
			input: "",
			want:  "",
		},
		{
			name:  "non-data line",
			input: "some random text",
			want:  "some random text",
		},
		{
			name:  "done marker",
			input: "data: [DONE]",
			want:  "data: [DONE]",
		},
		{
			name:  "valid json data line",
			input: `data: {"id":"test","choices":[{"message":{"content":"hello"}}]}`,
			want:  `data: {"choices":[{"message":{"content":"hello"}}],"id":"test"}`,
		},
		{
			name:  "invalid json data line preserved",
			input: "data: not valid json",
			want:  "data: not valid json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.CompactStreamingLine(tt.input)

			if got != tt.want {
				t.Errorf("CompactStreamingLine(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCompactContentArray(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	tests := []struct {
		name       string
		input      []any
		wantLen    int
		wantEmpty  bool
	}{
		{
			name:      "nil input",
			input:     nil,
			wantLen:   0,
			wantEmpty: true,
		},
		{
			name: "thinking blocks filtered",
			input: []any{
				map[string]any{"type": "thinking", "content": "thoughts"},
				map[string]any{"type": "text", "text": "hello"},
			},
			wantLen: 1,
		},
		{
			name: "redacted thinking filtered",
			input: []any{
				map[string]any{"type": "redacted_thinking"},
				map[string]any{"type": "text", "text": "hello"},
			},
			wantLen: 1,
		},
		{
			name: "all thinking returns empty text",
			input: []any{
				map[string]any{"type": "thinking", "content": "thoughts"},
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.compactContentArray(tt.input)

			if tt.wantEmpty && got != nil {
				t.Errorf("expected nil, got %v", got)
				return
			}

			if !tt.wantEmpty && len(got) != tt.wantLen {
				t.Errorf("expected %d items, got %d", tt.wantLen, len(got))
			}

			if tt.name == "all thinking returns empty text" {
				if len(got) > 0 {
					item := got[0].(map[string]any)
					if item["type"] != "text" {
						t.Errorf("expected text type, got %v", item["type"])
					}
					if item["text"] != "" {
						t.Errorf("expected empty text, got %v", item["text"])
					}
				}
			}
		})
	}
}

func TestCompactContentBlock(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	tests := []struct {
		name         string
		input        map[string]any
		wantFields   []string
		wantRemoved  []string
	}{
		{
			name:       "nil input",
			input:      nil,
			wantFields: []string{},
		},
		{
			name: "signature removed",
			input: map[string]any{
				"type":      "text",
				"text":      "hello",
				"signature": "secret-signature",
			},
			wantFields:  []string{"type", "text"},
			wantRemoved: []string{"signature"},
		},
		{
			name: "other fields preserved",
			input: map[string]any{
				"type": "image",
				"url":  "https://example.com/image.png",
			},
			wantFields: []string{"type", "url"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.compactContentBlock(tt.input)

			if tt.input == nil {
				if got != nil {
					t.Errorf("expected nil for nil input, got %v", got)
				}
				return
			}

			for _, field := range tt.wantFields {
				if _, exists := got[field]; !exists {
					t.Errorf("expected field %q to exist", field)
				}
			}

			for _, field := range tt.wantRemoved {
				if _, exists := got[field]; exists {
					t.Errorf("expected field %q to be removed", field)
				}
			}
		})
	}
}

func TestCompactOutput(t *testing.T) {
	svc := NewCompactService(&mockLogger{})

	input := []any{
		map[string]any{
			"type": "message",
			"content": []any{
				map[string]any{
					"type": "text",
					"text": "Hello",
				},
				map[string]any{
					"type": "thinking",
					"content": "internal",
				},
			},
		},
	}

	got := svc.compactOutput(input)

	if len(got) != 1 {
		t.Errorf("expected 1 output item, got %d", len(got))
		return
	}

	item := got[0].(map[string]any)
	content := item["content"].([]any)

	if len(content) != 1 {
		t.Errorf("expected 1 content item after filtering, got %d", len(content))
	}
}
