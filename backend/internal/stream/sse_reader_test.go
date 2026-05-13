package stream

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestSSEReader_ReadChunk(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *StreamChunk
	}{
		{
			name:  "simple data",
			input: "data: hello\n\n",
			expected: &StreamChunk{
				Data: []byte("hello"),
			},
		},
		{
			name:  "data with event",
			input: "event: message\ndata: hello\n\n",
			expected: &StreamChunk{
				Event: "message",
				Data:  []byte("hello"),
			},
		},
		{
			name:  "data with id and retry",
			input: "id: 123\nretry: 1000\ndata: hello\n\n",
			expected: &StreamChunk{
				ID:    "123",
				Retry: 1000,
				Data:  []byte("hello"),
			},
		},
		{
			name:  "multi-line data",
			input: "data: line1\ndata: line2\n\n",
			expected: &StreamChunk{
				Data: []byte("line1\nline2"),
			},
		},
		{
			name:  "[DONE] marker",
			input: "data: [DONE]\n\n",
			expected: &StreamChunk{
				Data: []byte("[DONE]"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := NewSSEReader(strings.NewReader(tt.input))
			chunk, err := reader.ReadChunk(context.Background())
			if err != nil {
				t.Fatalf("ReadChunk() error = %v", err)
			}

			if chunk.Event != tt.expected.Event {
				t.Errorf("Event = %v, want %v", chunk.Event, tt.expected.Event)
			}
			if chunk.ID != tt.expected.ID {
				t.Errorf("ID = %v, want %v", chunk.ID, tt.expected.ID)
			}
			if chunk.Retry != tt.expected.Retry {
				t.Errorf("Retry = %v, want %v", chunk.Retry, tt.expected.Retry)
			}
			if !bytes.Equal(chunk.Data, tt.expected.Data) {
				t.Errorf("Data = %s, want %s", chunk.Data, tt.expected.Data)
			}
		})
	}
}

func TestIsDoneChunk(t *testing.T) {
	tests := []struct {
		name     string
		chunk    *StreamChunk
		expected bool
	}{
		{
			name: "[DONE] marker",
			chunk: &StreamChunk{
				Data: []byte("[DONE]"),
			},
			expected: true,
		},
		{
			name: "not done",
			chunk: &StreamChunk{
				Data: []byte("hello"),
			},
			expected: false,
		},
		{
			name:     "nil chunk",
			chunk:    nil,
			expected: false,
		},
		{
			name: "empty data",
			chunk: &StreamChunk{
				Data: []byte{},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsDoneChunk(tt.chunk)
			if result != tt.expected {
				t.Errorf("IsDoneChunk() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestParseSSELine(t *testing.T) {
	tests := []struct {
		name           string
		line           string
		expectedField  string
		expectedValue  string
		expectedOk     bool
	}{
		{
			name:          "data field",
			line:          "data: hello",
			expectedField: "data",
			expectedValue: "hello",
			expectedOk:    true,
		},
		{
			name:          "data with space",
			line:          "data:  hello",
			expectedField: "data",
			expectedValue: "hello",
			expectedOk:    true,
		},
		{
			name:          "comment",
			line:          ": comment",
			expectedField: "",
			expectedValue: "",
			expectedOk:    false,
		},
		{
			name:          "empty line",
			line:          "",
			expectedField: "",
			expectedValue: "",
			expectedOk:    false,
		},
		{
			name:          "field without value",
			line:          "event",
			expectedField: "event",
			expectedValue: "",
			expectedOk:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, value, ok := ParseSSELine(tt.line)
			if field != tt.expectedField {
				t.Errorf("field = %v, want %v", field, tt.expectedField)
			}
			if value != tt.expectedValue {
				t.Errorf("value = %v, want %v", value, tt.expectedValue)
			}
			if ok != tt.expectedOk {
				t.Errorf("ok = %v, want %v", ok, tt.expectedOk)
			}
		})
	}
}
