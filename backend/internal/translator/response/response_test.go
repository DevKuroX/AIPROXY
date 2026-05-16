package response

import (
	"testing"
)

func TestClaudeResponseType(t *testing.T) {
	resp := CLaudeResponse{
		ID:      "msg_123",
		Type:    "message",
		Content: []CLaudeContentBlock{{Type: "text", Text: "hello"}},
		Role:    "assistant",
		Model:   "claude-sonnet-4",
	}
	if resp.ID != "msg_123" {
		t.Fatalf("expected msg_123, got %s", resp.ID)
	}
}

func TestTranslateClaudeToOpenAIResponse(t *testing.T) {
	resp := &CLaudeResponse{
		ID:      "msg_1",
		Type:    "message",
		Role:    "assistant",
		Model:   "claude-sonnet-4",
		Content: []CLaudeContentBlock{{Type: "text", Text: "Hello there!"}},
		StopReason: "end_turn",
		Usage:  CLaudeUsage{InputTokens: 10, OutputTokens: 5},
	}

	openaiResp, err := TranslateResponse(resp)
	if err != nil {
		t.Fatalf("TranslateResponse failed: %v", err)
	}
	if openaiResp == nil {
		t.Fatal("TranslateResponse returned nil")
	}
	if len(openaiResp.Choices) == 0 {
		t.Fatal("expected at least 1 choice")
	}
}

func TestTranslateGeminiToOpenAIResponse(t *testing.T) {
	resp := &GeminiResponse{
		Candidates: []GeminiCandidate{
			{
				Content: GeminiContent{
					Role:  "model",
					Parts: []GeminiPart{{Text: "Hi there"}},
				},
			},
		},
	}

	openaiResp, err := TranslateGeminiToOpenAIResponse(resp, "gemini-3-flash")
	if err != nil {
		t.Fatalf("TranslateGeminiToOpenAIResponse failed: %v", err)
	}
	if openaiResp == nil {
		t.Fatal("TranslateGeminiToOpenAIResponse returned nil")
	}
}
