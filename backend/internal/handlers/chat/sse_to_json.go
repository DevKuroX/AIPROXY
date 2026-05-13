package chat

import (
	"context"
	"encoding/json"
	"io"

	"github.com/DevKuroX/AIPROXY/internal/stream"
)

// SSEToJSONConverter converts SSE streams to JSON responses
// ref: open-sse/handlers/chatCore/sseToJsonHandler.js
type SSEToJSONConverter struct{}

// Convert converts an SSE stream to a JSON response
func (c *SSEToJSONConverter) Convert(
	ctx context.Context,
	reader io.Reader,
	format stream.StreamFormat,
) ([]byte, *stream.Usage, error) {
	// Create SSE reader
	sseReader := stream.NewSSEReader(reader)
	defer sseReader.Close()

	// Buffer all chunks
	var chunks []*stream.StreamChunk
	for {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
		}

		chunk, err := sseReader.ReadChunk(ctx)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, nil, err
		}

		// Skip [DONE] marker
		if stream.IsDoneChunk(chunk) {
			break
		}

		chunks = append(chunks, chunk)
	}

	// Extract usage
	usage, _ := stream.ExtractUsageFromStream(chunks)

	// Build final response based on format
	response, err := c.buildResponse(chunks, format)
	if err != nil {
		return nil, nil, err
	}

	// Add usage to response
	if usage != nil {
		var respMap map[string]interface{}
		if err := json.Unmarshal(response, &respMap); err == nil {
			respMap["usage"] = usage
			response, err = json.Marshal(respMap)
			if err != nil {
				return nil, nil, err
			}
		}
	}

	return response, usage, nil
}

// buildResponse builds final JSON response from chunks
func (c *SSEToJSONConverter) buildResponse(
	chunks []*stream.StreamChunk,
	format stream.StreamFormat,
) ([]byte, error) {
	switch format {
	case stream.FormatOpenAI:
		return c.buildOpenAIResponse(chunks)
	case stream.FormatCL4ude:
		return c.buildCL4udeResponse(chunks)
	case stream.FormatGemini:
		return c.buildGeminiResponse(chunks)
	default:
		return c.buildOpenAIResponse(chunks)
	}
}

// buildOpenAIResponse builds OpenAI format response
func (c *SSEToJSONConverter) buildOpenAIResponse(chunks []*stream.StreamChunk) ([]byte, error) {
	if len(chunks) == 0 {
		return json.Marshal(map[string]interface{}{
			"object": "chat.completion",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "",
					},
					"finish_reason": "stop",
				},
			},
		})
	}

	// Merge content from all chunks
	var content string
	var finishReason string

	for _, chunk := range chunks {
		if len(chunk.Data) == 0 {
			continue
		}

		var data map[string]interface{}
		if err := json.Unmarshal(chunk.Data, &data); err != nil {
			continue
		}

		// Extract content
		if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if chunkContent, ok := delta["content"].(string); ok {
						content += chunkContent
					}
				}
				if reason, ok := choice["finish_reason"].(string); ok && reason != "" {
					finishReason = reason
				}
			}
		}
	}

	response := map[string]interface{}{
		"object": "chat.completion",
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": content,
				},
				"finish_reason": finishReason,
			},
		},
	}

	return json.Marshal(response)
}

// buildCL4udeResponse builds CL4ude format response
func (c *SSEToJSONConverter) buildCL4udeResponse(chunks []*stream.StreamChunk) ([]byte, error) {
	// Similar to OpenAI but with CL4ude format
	return c.buildOpenAIResponse(chunks)
}

// buildGeminiResponse builds Gemini format response
func (c *SSEToJSONConverter) buildGeminiResponse(chunks []*stream.StreamChunk) ([]byte, error) {
	// Similar to OpenAI but with Gemini format
	return c.buildOpenAIResponse(chunks)
}

// ConvertReader converts any reader to JSON (auto-detects format)
func (c *SSEToJSONConverter) ConvertReader(
	ctx context.Context,
	reader io.Reader,
	contentType string,
) ([]byte, *stream.Usage, error) {
	// Detect format
	format, err := stream.DetectStreamFormatFromReader(reader)
	if err != nil {
		return nil, nil, err
	}

	// Reset reader (would need a TeeReader or re-reading)
	// For now, assume format detection doesn't consume data
	return c.Convert(ctx, reader, format)
}

// IsSSEResponse checks if response is SSE
func IsSSEResponse(contentType string) bool {
	return contentType == "text/event-stream"
}

// ShouldConvertSSEToJSON checks if SSE should be converted to JSON
func ShouldConvertSSEToJSON(streamRequested bool, forceJSON bool) bool {
	return !streamRequested || forceJSON
}
