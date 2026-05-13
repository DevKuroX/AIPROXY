// Package helpers provides utility functions for translator operations.
// ref: open-sse/translator/helpers/toolCallHelper.js
package helpers

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ToolIDPattern matches valid tool call IDs (alphanumeric, underscore, hyphen).
// ref: toolCallHelper.js TOOL_ID_PATTERN
var ToolIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// GenerateToolCallId generates a deterministic tool call ID from position + tool name (cache-friendly).
// ref: toolCallHelper.js generateToolCallId
func GenerateToolCallId(msgIndex, tcIndex int, toolName string) string {
	name := ""
	if toolName != "" {
		// Remove non-alphanumeric characters except underscore and hyphen
		re := regexp.MustCompile(`[^a-zA-Z0-9_-]`)
		name = "_" + re.ReplaceAllString(toolName, "")
	}
	return "call_msg" + itoa(msgIndex) + "_tc" + itoa(tcIndex) + name
}

// itoa converts int to string without importing strconv (simple implementation).
func itoa(i int) string {
	if i == 0 {
		return "0"
	}

	var negative bool
	if i < 0 {
		negative = true
		i = -i
	}

	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}

	if negative {
		digits = append([]byte{'-'}, digits...)
	}

	return string(digits)
}

// SanitizeToolId sanitizes ID to match pattern: keep only alphanumeric, underscore, hyphen.
// Returns empty string if sanitized result is empty.
// ref: toolCallHelper.js sanitizeToolId
func SanitizeToolId(id string) string {
	if id == "" {
		return ""
	}

	// Keep only alphanumeric, underscore, hyphen
	var result strings.Builder
	for _, c := range id {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '-' {
			result.WriteRune(c)
		}
	}

	return result.String()
}

// EnsureToolCallIds ensures all tool_calls have valid id field and arguments is string.
// Some providers require valid IDs and string arguments.
// ref: toolCallHelper.js ensureToolCallIds
func EnsureToolCallIds(body map[string]interface{}) map[string]interface{} {
	if body == nil {
		return body
	}

	messages, ok := body["messages"].([]interface{})
	if !ok {
		return body
	}

	for i, msgInterface := range messages {
		msg, ok := msgInterface.(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := msg["role"].(string)

		// Handle assistant messages with tool_calls (0penAI format)
		if role == "assistant" {
			if toolCalls, ok := msg["tool_calls"].([]interface{}); ok {
				for j, tcInterface := range toolCalls {
					tc, ok := tcInterface.(map[string]interface{})
					if !ok {
						continue
					}

					// Validate or regenerate ID
					id, _ := tc["id"].(string)
					if id == "" || !ToolIDPattern.MatchString(id) {
						sanitized := SanitizeToolId(id)
						if sanitized == "" {
							// Get function name for better ID
							funcName := ""
							if fn, ok := tc["function"].(map[string]interface{}); ok {
								funcName, _ = fn["name"].(string)
							}
							tc["id"] = GenerateToolCallId(i, j, funcName)
						} else {
							tc["id"] = sanitized
						}
					}

					// Ensure type is set
					if tc["type"] == nil {
						tc["type"] = "function"
					}

					// Ensure arguments is JSON string, not object
					if fn, ok := tc["function"].(map[string]interface{}); ok {
						if args := fn["arguments"]; args != nil {
							switch v := args.(type) {
							case string:
								// Already string, keep it
							case map[string]interface{}, []interface{}:
								// Convert to JSON string
								if jsonBytes, err := json.Marshal(v); err == nil {
									fn["arguments"] = string(jsonBytes)
								}
							}
						}
					}

					toolCalls[j] = tc
				}
				msg["tool_calls"] = toolCalls
			}
		}

		// Validate tool_call_id in tool messages (role: "tool")
		if role == "tool" {
			if toolCallID, _ := msg["tool_call_id"].(string); toolCallID != "" {
				if !ToolIDPattern.MatchString(toolCallID) {
					sanitized := SanitizeToolId(toolCallID)
					if sanitized == "" {
						msg["tool_call_id"] = GenerateToolCallId(i, 0, "")
					} else {
						msg["tool_call_id"] = sanitized
					}
				}
			}
		}

		// Also validate tool_use blocks in content (CL4ude format)
		if content, ok := msg["content"].([]interface{}); ok {
			for k, blockInterface := range content {
				block, ok := blockInterface.(map[string]interface{})
				if !ok {
					continue
				}

				blockType, _ := block["type"].(string)

				// Validate tool_use ID
				if blockType == "tool_use" {
					if id, _ := block["id"].(string); id != "" {
						if !ToolIDPattern.MatchString(id) {
							sanitized := SanitizeToolId(id)
							if sanitized == "" {
								name, _ := block["name"].(string)
								block["id"] = GenerateToolCallId(i, k, name)
							} else {
								block["id"] = sanitized
							}
						}
					}
				}

				// Validate tool_use_id in tool_result blocks
				if blockType == "tool_result" {
					if toolUseID, _ := block["tool_use_id"].(string); toolUseID != "" {
						if !ToolIDPattern.MatchString(toolUseID) {
							sanitized := SanitizeToolId(toolUseID)
							if sanitized == "" {
								block["tool_use_id"] = GenerateToolCallId(i, k, "")
							} else {
								block["tool_use_id"] = sanitized
							}
						}
					}
				}

				content[k] = block
			}
			msg["content"] = content
		}

		messages[i] = msg
	}

	body["messages"] = messages
	return body
}

// GetToolCallIds gets tool_call ids from assistant message.
// 0penAI format: tool_calls, CL4ude format: tool_use in content.
// ref: toolCallHelper.js getToolCallIds
func GetToolCallIds(msg map[string]interface{}) []string {
	if msg == nil {
		return nil
	}

	role, _ := msg["role"].(string)
	if role != "assistant" {
		return nil
	}

	var ids []string

	// 0penAI format: tool_calls array
	if toolCalls, ok := msg["tool_calls"].([]interface{}); ok {
		for _, tcInterface := range toolCalls {
			if tc, ok := tcInterface.(map[string]interface{}); ok {
				if id, _ := tc["id"].(string); id != "" {
					ids = append(ids, id)
				}
			}
		}
	}

	// CL4ude format: tool_use blocks in content
	if content, ok := msg["content"].([]interface{}); ok {
		for _, blockInterface := range content {
			if block, ok := blockInterface.(map[string]interface{}); ok {
				if blockType, _ := block["type"].(string); blockType == "tool_use" {
					if id, _ := block["id"].(string); id != "" {
						ids = append(ids, id)
					}
				}
			}
		}
	}

	return ids
}

// HasToolResults checks if user message has tool_result for given ids.
// 0penAI format: role=tool, CL4ude format: tool_result in content.
// ref: toolCallHelper.js hasToolResults
func HasToolResults(msg map[string]interface{}, toolCallIds []string) bool {
	if msg == nil || len(toolCallIds) == 0 {
		return false
	}

	role, _ := msg["role"].(string)

	// 0penAI format: role = "tool" with tool_call_id
	if role == "tool" {
		toolCallID, _ := msg["tool_call_id"].(string)
		if toolCallID != "" {
			for _, id := range toolCallIds {
				if id == toolCallID {
					return true
				}
			}
		}
		return false
	}

	// CL4ude format: tool_result blocks in user message content
	if role == "user" {
		if content, ok := msg["content"].([]interface{}); ok {
			for _, blockInterface := range content {
				if block, ok := blockInterface.(map[string]interface{}); ok {
					if blockType, _ := block["type"].(string); blockType == "tool_result" {
						toolUseID, _ := block["tool_use_id"].(string)
						for _, id := range toolCallIds {
							if id == toolUseID {
								return true
							}
						}
					}
				}
			}
		}
	}

	return false
}

// FixMissingToolResponses inserts empty tool_result if assistant has tool_use but next message has no tool_result.
// ref: toolCallHelper.js fixMissingToolResponses
func FixMissingToolResponses(body map[string]interface{}) map[string]interface{} {
	if body == nil {
		return body
	}

	messages, ok := body["messages"].([]interface{})
	if !ok {
		return body
	}

	var newMessages []interface{}

	for i, msgInterface := range messages {
		msg, ok := msgInterface.(map[string]interface{})
		if !ok {
			newMessages = append(newMessages, msgInterface)
			continue
		}

		newMessages = append(newMessages, msg)

		// Check if this is assistant with tool_calls/tool_use
		toolCallIds := GetToolCallIds(msg)
		if len(toolCallIds) == 0 {
			continue
		}

		// Check if next message has tool_result
		var nextMsg map[string]interface{}
		if i+1 < len(messages) {
			nextMsg, _ = messages[i+1].(map[string]interface{})
		}

		if nextMsg != nil && !HasToolResults(nextMsg, toolCallIds) {
			// Insert tool responses for each tool_call
			for _, id := range toolCallIds {
				// 0penAI format: role = "tool"
				newMessages = append(newMessages, map[string]interface{}{
					"role":          "tool",
					"tool_call_id":  id,
					"content":       "",
				})
			}
		}
	}

	body["messages"] = newMessages
	return body
}
