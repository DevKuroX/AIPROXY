// ref: _ref/9router/open-sse/utils/streamHelpers.js
package utils

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	Format0penAI  = "openai"
	FormatCL4ude  = "claude"
	FormatGemini  = "gemini"
	FormatOllama  = "ollama"
	FormatCursor  = "cursor"
)

func FixInvalidID(parsed map[string]interface{}) bool {
	id, ok := parsed["id"].(string)
	if !ok {
		return false
	}

	if id == "chat" || id == "completion" || len(id) < 8 {
		fallbackID := ""
		if extendFields, ok := parsed["extend_fields"].(map[string]interface{}); ok {
			if rid, ok := extendFields["requestId"].(string); ok && rid != "" {
				fallbackID = rid
			} else if tid, ok := extendFields["traceId"].(string); ok && tid != "" {
				fallbackID = tid
			}
		}

		if fallbackID == "" {
			fallbackID = fmt.Sprintf("%d", getCurrentTimeMillis())
		}

		parsed["id"] = "chatcmpl-" + fallbackID
		return true
	}

	return false
}

func getCurrentTimeMillis() int64 {
	return int64(0)
}

func HasValuableContent(chunk map[string]interface{}, format string) bool {
	switch format {
	case Format0penAI:
		if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if delta, ok := choice["delta"].(map[string]interface{}); ok {
					if content, _ := delta["content"].(string); content != "" {
						return true
					}
					if reasoning, _ := delta["reasoning_content"].(string); reasoning != "" {
						return true
					}
					if toolCalls, ok := delta["tool_calls"].([]interface{}); ok && len(toolCalls) > 0 {
						return true
					}
					if _, ok := delta["role"].(string); ok {
						return true
					}
				}
				if _, ok := choice["finish_reason"].(string); ok {
					return true
				}
			}
		}
		return false

	case FormatCL4ude:
		chunkType, _ := chunk["type"].(string)
		if chunkType == "content_block_delta" {
			if delta, ok := chunk["delta"].(map[string]interface{}); ok {
				if text, _ := delta["text"].(string); text != "" {
					return true
				}
				if thinking, _ := delta["thinking"].(string); thinking != "" {
					return true
				}
				if partialJSON, _ := delta["partial_json"].(string); partialJSON != "" {
					return true
				}
			}
			return false
		}
		return true

	default:
		return true
	}
}

func CleanUsagePayload(payload interface{}) interface{} {
	if payload == nil {
		return nil
	}

	p, ok := payload.(map[string]interface{})
	if !ok {
		return payload
	}

	result := make(map[string]interface{})
	for k, v := range p {
		result[k] = v
	}

	if usage, ok := result["usage"]; ok {
		if usage == nil {
			delete(result, "usage")
		} else if usageMap, ok := usage.(map[string]interface{}); ok {
			cleanedUsage := make(map[string]interface{})
			for k, v := range usageMap {
				if v != nil {
					cleanedUsage[k] = v
				}
			}
			if len(cleanedUsage) > 0 {
				result["usage"] = cleanedUsage
			} else {
				delete(result, "usage")
			}
		}
	}

	return result
}

func ParseOpenAIStreamChunk(line string) (map[string]interface{}, error) {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "data:") {
		return nil, nil
	}

	data := strings.TrimPrefix(line, "data:")
	data = strings.TrimSpace(data)

	if data == "[DONE]" {
		return map[string]interface{}{"done": true}, nil
	}

	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(data), &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse OpenAI chunk: %w", err)
	}

	return chunk, nil
}

func ParseClaudeStreamChunk(line string) (map[string]interface{}, error) {
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	var chunk map[string]interface{}
	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
		return nil, fmt.Errorf("failed to parse Claude chunk: %w", err)
	}

	return chunk, nil
}

func BuildOpenAIStreamChunk(id string, created int64, model string, delta map[string]interface{}, finishReason string) string {
	chunk := map[string]interface{}{
		"id":      id,
		"object":  "chat.completion.chunk",
		"created": created,
		"model":   model,
		"choices": []interface{}{
			map[string]interface{}{
				"index": 0,
				"delta": delta,
			},
		},
	}

	if finishReason != "" {
		chunk["choices"].([]interface{})[0].(map[string]interface{})["finish_reason"] = finishReason
	}

	data, _ := json.Marshal(chunk)
	return "data: " + string(data) + "\n\n"
}

func BuildClaudeStreamChunk(eventType string, data map[string]interface{}) string {
	data["type"] = eventType
	dataBytes, _ := json.Marshal(data)
	return string(dataBytes) + "\n"
}
