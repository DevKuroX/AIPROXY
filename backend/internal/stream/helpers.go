package stream

import (
	"context"
	"encoding/json"
	"io"
	"strings"
)

// DetectStreamFormatFromReader detects stream format from reader content
// ref: open-sse/utils/streamHelpers.js - detectFormat
func DetectStreamFormatFromReader(reader io.Reader) (StreamFormat, error) {
	// Read first few bytes to detect format
	buf := make([]byte, 1024)
	n, err := reader.Read(buf)
	if err != nil && err != io.EOF {
		return FormatOpenAI, err
	}

	data := string(buf[:n])
	
	// Check for format indicators
	if strings.Contains(data, `"claude"`) || strings.Contains(data, `"type":"completion"`) {
		return FormatCL4ude, nil
	} else if strings.Contains(data, `"gemini"`) || strings.Contains(data, `"candidates"`) {
		return FormatGemini, nil
	} else if strings.Contains(data, `"antigravity"`) {
		return FormatAntigravity, nil
	} else if strings.Contains(data, `"object":"chat.completion.chunk"`) {
		return FormatOpenAI, nil
	} else if strings.Contains(data, `"object":"response"`) {
		return FormatOpenAIResponses, nil
	}
	
	return FormatOpenAI, nil
}

// NormalizeSSEChunk normalizes SSE chunk based on target format
// ref: open-sse/utils/streamHelpers.js - normalizeChunk
func NormalizeSSEChunk(chunk *StreamChunk, format StreamFormat) (*StreamChunk, error) {
	if chunk == nil {
		return nil, nil
	}

	// Clone the chunk
	normalized := &StreamChunk{
		Event: chunk.Event,
		Data:  make([]byte, len(chunk.Data)),
		ID:    chunk.ID,
		Retry: chunk.Retry,
	}
	copy(normalized.Data, chunk.Data)

	// Apply format-specific normalization
	switch format {
	case FormatOpenAI:
		// Ensure OpenAI format
		normalized.Event = "message"
	case FormatCL4ude:
		// Ensure CL4ude format
		normalized.Event = "completion"
	case FormatGemini:
		// Ensure Gemini format
		normalized.Event = "candidate"
	}

	return normalized, nil
}

// ExtractUsageFromStream extracts usage from stream chunks
// ref: open-sse/utils/stream.js - extractUsageFromStream
func ExtractUsageFromStream(chunks []*StreamChunk) (*Usage, error) {
	usage := &Usage{}
	
	for _, chunk := range chunks {
		if len(chunk.Data) == 0 {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal(chunk.Data, &data); err != nil {
			continue
		}

		// Extract usage from different formats
		if usageData, ok := data["usage"].(map[string]interface{}); ok {
			if prompt, ok := usageData["prompt_tokens"].(float64); ok {
				usage.PromptTokens = int(prompt)
			}
			if completion, ok := usageData["completion_tokens"].(float64); ok {
				usage.CompletionTokens = int(completion)
			}
			if total, ok := usageData["total_tokens"].(float64); ok {
				usage.TotalTokens = int(total)
			}
		}
		
		// Check for usage at root level (some formats)
		if prompt, ok := data["prompt_tokens"].(float64); ok {
			usage.PromptTokens = int(prompt)
		}
		if completion, ok := data["completion_tokens"].(float64); ok {
			usage.CompletionTokens = int(completion)
		}
		if total, ok := data["total_tokens"].(float64); ok {
			usage.TotalTokens = int(total)
		}
	}

	return usage, nil
}

// IsStreamComplete checks if stream is complete based on chunk
// ref: open-sse/utils/streamHelpers.js - hasValuableContent
func IsStreamComplete(chunk *StreamChunk) bool {
	if chunk == nil {
		return false
	}
	
	// Check for [DONE] marker
	if IsDoneChunk(chunk) {
		return true
	}

	// Check for completion indicators in data
	if len(chunk.Data) > 0 {
		var data map[string]interface{}
		if err := json.Unmarshal(chunk.Data, &data); err == nil {
			// OpenAI: finish_reason != null
			if finishReason, ok := data["finish_reason"].(string); ok && finishReason != "" && finishReason != "null" {
				return true
			}
			// CL4ude: stop_reason != null
			if stopReason, ok := data["stop_reason"].(string); ok && stopReason != "" && stopReason != "null" {
				return true
			}
		}
	}

	return false
}

// CreateErrorChunk creates an SSE error chunk
// ref: open-sse/utils/error.js - createErrorResult
func CreateErrorChunk(message, errorType string, code int) *StreamChunk {
	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    errorType,
			"code":    code,
		},
	}

	data, _ := json.Marshal(errorResp)
	return &StreamChunk{
		Event: "error",
		Data:  data,
	}
}

// CreateDoneChunk creates a [DONE] chunk
func CreateDoneChunk() *StreamChunk {
	return &StreamChunk{
		Data: []byte("[DONE]"),
	}
}

// BufferStream buffers a stream into chunks (for non-streaming responses)
// ref: open-sse/handlers/chatCore/sseToJsonHandler.js
func BufferStream(reader StreamReader) ([]*StreamChunk, error) {
	var chunks []*StreamChunk
	ctx := context.Background()

	for {
		chunk, err := reader.ReadChunk(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return chunks, err
		}

		chunks = append(chunks, chunk)
		
		if IsDoneChunk(chunk) {
			break
		}
	}

	return chunks, nil
}
