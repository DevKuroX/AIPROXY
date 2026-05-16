package stream

import (
	"context"
	"net/http/httptest"
	"testing"
)

func TestSSEWriter_WriteChunk(t *testing.T) {
	tests := []struct {
		name     string
		chunk    *StreamChunk
		expected string
	}{
		{
			name: "simple data chunk",
			chunk: &StreamChunk{
				Data: []byte("hello"),
			},
			expected: "data: hello\n\n",
		},
		{
			name: "chunk with event",
			chunk: &StreamChunk{
				Event: "message",
				Data:  []byte("hello"),
			},
			expected: "event: message\ndata: hello\n\n",
		},
		{
			name: "chunk with id",
			chunk: &StreamChunk{
				ID:   "123",
				Data: []byte("hello"),
			},
			expected: "id: 123\ndata: hello\n\n",
		},
		{
			name: "chunk with retry",
			chunk: &StreamChunk{
				Retry: 3000,
				Data:  []byte("hello"),
			},
			expected: "retry: 3000\ndata: hello\n\n",
		},
		{
			name: "multi-line data",
			chunk: &StreamChunk{
				Data: []byte("line1\nline2\nline3"),
			},
			expected: "data: line1\ndata: line2\ndata: line3\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			writer := NewSSEWriter(recorder)

			ctx := context.Background()
			err := writer.WriteChunk(ctx, tt.chunk)
			if err != nil {
				t.Fatalf("WriteChunk failed: %v", err)
			}

			// Check headers
			if recorder.Header().Get("Content-Type") != "text/event-stream" {
				t.Errorf("Content-Type header not set correctly")
			}
			if recorder.Header().Get("Cache-Control") != "no-cache" {
				t.Errorf("Cache-Control header not set correctly")
			}

			// Check body
			body := recorder.Body.String()
			if body != tt.expected {
				t.Errorf("Expected body:\n%q\nGot:\n%q", tt.expected, body)
			}
		})
	}
}

func TestSSEWriter_WriteError(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewSSEWriter(recorder)

	ctx := context.Background()
	err := writer.WriteError(ctx, NewStreamError("test error", "api_error", ""))

	if err != nil {
		t.Fatalf("WriteError failed: %v", err)
	}

	expected := `event: error
data: {"error":{"message":"test error","type":"api_error"}}

`
	body := recorder.Body.String()
	if body != expected {
		t.Errorf("Expected:\n%q\nGot:\n%q", expected, body)
	}
}

func TestSSEWriter_Close(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewSSEWriter(recorder)

	err := writer.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Close just flushes, [DONE] is written by handler separately
	body := recorder.Body.String()
	if body == "" {
		t.Log("Close flushed (body empty as expected)")
	}

	// Second close should be no-op
	err = writer.Close()
	if err != nil {
		t.Fatalf("Second Close failed: %v", err)
	}
}

func TestSSEWriter_ContextCancellation(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewSSEWriter(recorder)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	chunk := &StreamChunk{Data: []byte("test")}
	err := writer.WriteChunk(ctx, chunk)
	if err == nil {
		t.Error("Expected error on cancelled context, got nil")
	}
}

func TestSSEWriter_WriteRaw(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewSSEWriter(recorder)

	ctx := context.Background()
	data := []byte("data: raw\n\n")
	err := writer.WriteRaw(ctx, data)
	if err != nil {
		t.Fatalf("WriteRaw failed: %v", err)
	}

	body := recorder.Body.String()
	if body != string(data) {
		t.Errorf("Expected raw data, got: %q", body)
	}
}

func TestSSEWriter_Flush(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := NewSSEWriter(recorder)

	err := writer.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	// Test flush on closed writer
	writer.Close()
	err = writer.Flush()
	if err == nil {
		t.Error("Expected error when flushing closed writer")
	}
}