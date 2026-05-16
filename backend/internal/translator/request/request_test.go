package request

import (
	"testing"
)

func TestOpenAIRequestTypes(t *testing.T) {
	msg := OpenAIMessage{
		Role:    "user",
		Content: "hello",
	}
	if msg.Role != "user" {
		t.Fatalf("expected user, got %s", msg.Role)
	}
}

func TestCLaudeRequestTypes(t *testing.T) {
	req := CLaudeRequest{
		Model: "claude-sonnet-4",
		Messages: []CLaudeMessage{
			{Role: "user", Content: []CLaudeContent{{Type: "text", Text: "hello"}}},
		},
	}
	if len(req.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(req.Messages))
	}
}

func TestGeminiRequestTypes(t *testing.T) {
	req := GeminiRequest{
		Contents: []GeminiContent{
			{Role: "user", Parts: []GeminiPart{{Text: "hello"}}},
		},
	}
	if len(req.Contents) != 1 {
		t.Fatalf("expected 1 content, got %d", len(req.Contents))
	}
}

func TestTranslateRequest(t *testing.T) {
	temp := 0.7
	req := &OpenAIRequest{
		Model: "claude-sonnet-4",
		Messages: []OpenAIMessage{
			{Role: "user", Content: "hello"},
		},
		Temperature: &temp,
	}

	claudeReq, err := TranslateRequest("claude-sonnet-4", req, false)
	if err != nil {
		t.Fatalf("TranslateRequest failed: %v", err)
	}
	if claudeReq.Model != "claude-sonnet-4" {
		t.Fatalf("expected claude-sonnet-4, got %s", claudeReq.Model)
	}
}

func TestTranslateOpenAIToGeminiRequest(t *testing.T) {
	temp := 0.5
	req := &OpenAIRequest{
		Model: "gemini-3-flash",
		Messages: []OpenAIMessage{
			{Role: "user", Content: "hello"},
		},
		Temperature: &temp,
	}

	geminiReq := TranslateOpenAIToGeminiRequest("gemini-3-flash", req, false)
	if geminiReq == nil {
		t.Fatal("TranslateOpenAIToGeminiRequest returned nil")
	}
	if geminiReq.Model != "gemini-3-flash" {
		t.Fatalf("expected gemini-3-flash, got %s", geminiReq.Model)
	}
}

func TestTranslateImageRequest(t *testing.T) {
	req := &ImageGenerationRequest{
		Model:  "dall-e-3",
		Prompt: "a cat",
		N:      1,
		Size:   "1024x1024",
	}

	imgReq := TranslateImageRequest(req)
	if imgReq == nil {
		t.Fatal("TranslateImageRequest returned nil")
	}
	if imgReq.Prompt != "a cat" {
		t.Fatalf("expected 'a cat', got '%s'", imgReq.Prompt)
	}
}
