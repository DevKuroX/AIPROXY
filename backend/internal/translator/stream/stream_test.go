package stream

import (
	"testing"
)

func TestCLaudeEventType(t *testing.T) {
	event := CLaudeEvent{
		Type: "content_block_delta",
		Delta: &CLaudeDelta{
			Type: "text_delta",
			Text: "Hello",
		},
	}
	if event.Type != "content_block_delta" {
		t.Fatalf("expected content_block_delta, got %s", event.Type)
	}
}

func TestCLaudeEventContentBlock(t *testing.T) {
	content := &CLaudeContent{Type: "text", Text: "Hello"}
	event := CLaudeEvent{
		Type:         "content_block_start",
		Index:        0,
		ContentBlock: content,
	}
	if event.ContentBlock.Text != "Hello" {
		t.Fatalf("expected Hello, got %s", event.ContentBlock.Text)
	}
}

func TestCLaudeEventMessageStart(t *testing.T) {
	event := CLaudeEvent{
		Type:    "message_start",
		Message: &CLaudeMessage{ID: "msg_1", Role: "assistant"},
	}
	if event.Message.ID != "msg_1" {
		t.Fatalf("expected msg_1, got %s", event.Message.ID)
	}
}


func TestCLaudeUsageType(t *testing.T) {
	u := CLaudeUsage{InputTokens: 100, OutputTokens: 50}
	if u.InputTokens != 100 {
		t.Fatalf("expected 100, got %d", u.InputTokens)
	}
}
