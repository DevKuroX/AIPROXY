package helpers

import (
	"encoding/json"
)

// ConvertResponsesToChat converts Responses API request to Chat Completions format
// ref: open-sse/translator/helpers/responsesApiHelper.js - convertResponsesApiFormat
func ConvertResponsesToChat(body []byte) ([]byte, error) {
	var req map[string]any
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, err
	}

	// Check if this is a Responses API request (has "input" field)
	input, hasInput := req["input"]
	if !hasInput {
		return body, nil // Not a Responses API request, return as-is
	}

	result := make(map[string]any)
	for k, v := range req {
		result[k] = v
	}

	result["messages"] = []any{}

	// Convert instructions to system message
	// ref: open-sse/translator/helpers/responsesApiHelper.js:36-38
	if instructions, ok := req["instructions"].(string); ok && instructions != "" {
		result["messages"] = append(result["messages"].([]any), map[string]any{
			"role":    "system",
			"content": instructions,
		})
	}

	// Normalize and convert input items
	// ref: open-sse/translator/helpers/responsesApiHelper.js:45-118
	inputItems := normalizeResponsesInput(input)
	if inputItems == nil {
		return body, nil
	}

	// Process input items
	var currentAssistantMsg map[string]any
	var pendingToolResults []map[string]any

	for _, item := range inputItems {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		// Determine item type
		// ref: open-sse/translator/helpers/responsesApiHelper.js:50-51
		itemType, _ := itemMap["type"].(string)
		if itemType == "" {
			if _, hasRole := itemMap["role"]; hasRole {
				itemType = "message"
			}
		}

		switch itemType {
		case "message":
			// Flush pending assistant message with tool calls
			if currentAssistantMsg != nil {
				result["messages"] = append(result["messages"].([]any), currentAssistantMsg)
				currentAssistantMsg = nil
			}
			// Flush pending tool results
			if len(pendingToolResults) > 0 {
				for _, tr := range pendingToolResults {
					result["messages"] = append(result["messages"].([]any), tr)
				}
				pendingToolResults = nil
			}

			// Convert content
			content := convertContent(itemMap["content"])
			role, _ := itemMap["role"].(string)
			result["messages"] = append(result["messages"].([]any), map[string]any{
				"role":    role,
				"content": content,
			})

		case "function_call":
			// Start or append to assistant message with tool_calls
			// ref: open-sse/translator/helpers/responsesApiHelper.js:81-99
			if currentAssistantMsg == nil {
				currentAssistantMsg = map[string]any{
					"role":       "assistant",
					"content":    nil,
					"tool_calls": []any{},
				}
			}

			name, _ := itemMap["name"].(string)
			if name == "" {
				continue // Skip items with empty name
			}

			callId, _ := itemMap["call_id"].(string)
			arguments, _ := itemMap["arguments"].(string)

			toolCalls := currentAssistantMsg["tool_calls"].([]any)
			toolCalls = append(toolCalls, map[string]any{
				"id":   callId,
				"type": "function",
				"function": map[string]any{
					"name":      name,
					"arguments": arguments,
				},
			})
			currentAssistantMsg["tool_calls"] = toolCalls

		case "function_call_output":
			// Flush assistant message first if exists
			if currentAssistantMsg != nil {
				result["messages"] = append(result["messages"].([]any), currentAssistantMsg)
				currentAssistantMsg = nil
			}

			// Add tool result
			callId, _ := itemMap["call_id"].(string)
			output := itemMap["output"]
			outputStr, ok := output.(string)
			if !ok {
				outputBytes, _ := json.Marshal(output)
				outputStr = string(outputBytes)
			}

			pendingToolResults = append(pendingToolResults, map[string]any{
				"role":         "tool",
				"tool_call_id": callId,
				"content":      outputStr,
			})

		case "reasoning":
			// Skip reasoning items - they are for display only
			continue
		}
	}

	// Flush remaining
	if currentAssistantMsg != nil {
		result["messages"] = append(result["messages"].([]any), currentAssistantMsg)
	}
	if len(pendingToolResults) > 0 {
		for _, tr := range pendingToolResults {
			result["messages"] = append(result["messages"].([]any), tr)
		}
	}

	// Cleanup Responses API specific fields
	// ref: open-sse/translator/helpers/responsesApiHelper.js:130-137
	delete(result, "input")
	delete(result, "instructions")
	delete(result, "include")
	delete(result, "prompt_cache_key")
	delete(result, "store")
	delete(result, "reasoning")

	return json.Marshal(result)
}

// normalizeResponsesInput normalizes Responses API input to array format
// ref: open-sse/translator/helpers/responsesApiHelper.js:9-22
func normalizeResponsesInput(input any) []any {
	switch v := input.(type) {
	case string:
		text := v
		if text == "" {
			text = "..."
		}
		return []any{
			map[string]any{
				"type": "message",
				"role": "user",
				"content": []any{
					map[string]any{
						"type": "input_text",
						"text": text,
					},
				},
			},
		}
	case []any:
		if len(v) == 0 {
			return []any{
				map[string]any{
					"type": "message",
					"role": "user",
					"content": []any{
						map[string]any{
							"type": "input_text",
							"text": "...",
						},
					},
				},
			}
		}
		return v
	default:
		return nil
	}
}

// convertContent converts Responses API content to Chat Completions format
// ref: open-sse/translator/helpers/responsesApiHelper.js:68-78
func convertContent(content any) any {
	contentArray, ok := content.([]any)
	if !ok {
		return content
	}

	result := make([]any, 0, len(contentArray))
	for _, c := range contentArray {
		cp, ok := c.(map[string]any)
		if !ok {
			result = append(result, c)
			continue
		}

		cType, _ := cp["type"].(string)
		switch cType {
		case "input_text":
			text, _ := cp["text"].(string)
			result = append(result, map[string]any{
				"type": "text",
				"text": text,
			})
		case "output_text":
			text, _ := cp["text"].(string)
			result = append(result, map[string]any{
				"type": "text",
				"text": text,
			})
		case "input_image":
			imageURL, _ := cp["image_url"].(string)
			if imageURL == "" {
				imageURL, _ = cp["file_id"].(string)
			}
			detail, _ := cp["detail"].(string)
			if detail == "" {
				detail = "auto"
			}
			result = append(result, map[string]any{
				"type": "image_url",
				"image_url": map[string]any{
					"url":    imageURL,
					"detail": detail,
				},
			})
		default:
			result = append(result, c)
		}
	}

	return result
}
