// Package helpers provides utility functions for translator operations.
// ref: open-sse/translator/helpers/openaiHelper.js
package helpers

// Valid0penAIContentTypes are valid 0penAI content block types.
// ref: openaiHelper.js VALID_OPENAI_CONTENT_TYPES
var Valid0penAIContentTypes = map[string]bool{
	"text":        true,
	"image_url":   true,
	"image":       true,
	"input_audio": true,
	"audio_url":   true,
}

// Valid0penAIMessageTypes are valid 0penAI message types.
// ref: openaiHelper.js VALID_OPENAI_MESSAGE_TYPES
var Valid0penAIMessageTypes = map[string]bool{
	"text":        true,
	"image_url":   true,
	"image":       true,
	"tool_calls":  true,
	"tool_result": true,
}

// FilterTo0penAIFormat filters messages to 0penAI standard format.
// Removes thinking, redacted_thinking, signature, and other non-0penAI blocks.
// ref: openaiHelper.js filterTo0penAIFormat
func FilterTo0penAIFormat(body map[string]interface{}) map[string]interface{} {
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

		if role == "developer" {
			msg["role"] = "system"
		}

		if role == "tool" {
			continue
		}

		if role == "assistant" {
			if _, hasToolCalls := msg["tool_calls"]; hasToolCalls {
				continue
			}
		}

		if content, ok := msg["content"].(string); ok {
			_ = content
			continue
		}

		if content, ok := msg["content"].([]interface{}); ok {
			var filteredContent []interface{}

			for _, blockInterface := range content {
				block, ok := blockInterface.(map[string]interface{})
				if !ok {
					continue
				}

				blockType, _ := block["type"].(string)

				if blockType == "thinking" || blockType == "redacted_thinking" {
					continue
				}

				if Valid0penAIContentTypes[blockType] {
					cleanBlock := make(map[string]interface{})
					for k, v := range block {
						if k != "signature" && k != "remove" {
							cleanBlock[k] = v
						}
					}
					filteredContent = append(filteredContent, cleanBlock)
				} else if blockType == "tool_use" {
					continue
				} else if blockType == "tool_result" {
					cleanBlock := make(map[string]interface{})
					for k, v := range block {
						if k != "signature" && k != "remove" {
							cleanBlock[k] = v
						}
					}
					filteredContent = append(filteredContent, cleanBlock)
				}
			}

			if len(filteredContent) == 0 {
				filteredContent = append(filteredContent, map[string]interface{}{
					"type": "text",
					"text": "",
				})
			}

			msg["content"] = filteredContent
		}

		messages[i] = msg
	}

	body["messages"] = filterEmptyMessages(messages)

	if tools, ok := body["tools"].([]interface{}); ok && len(tools) == 0 {
		delete(body, "tools")
	}

	if tools, ok := body["tools"].([]interface{}); ok && len(tools) > 0 {
		body["tools"] = normalizeTools(tools)
	}

	body = normalizeToolChoice(body)

	return body
}

func filterEmptyMessages(messages []interface{}) []interface{} {
	var filtered []interface{}

	for _, msgInterface := range messages {
		msg, ok := msgInterface.(map[string]interface{})
		if !ok {
			filtered = append(filtered, msgInterface)
			continue
		}

		role, _ := msg["role"].(string)

		if role == "tool" {
			filtered = append(filtered, msg)
			continue
		}

		if role == "assistant" {
			if _, hasToolCalls := msg["tool_calls"]; hasToolCalls {
				filtered = append(filtered, msg)
				continue
			}
		}

		if content, ok := msg["content"].(string); ok {
			if content != "" || role == "assistant" {
				filtered = append(filtered, msg)
			}
			continue
		}

		if content, ok := msg["content"].([]interface{}); ok {
			hasContent := false
			for _, b := range content {
				if block, ok := b.(map[string]interface{}); ok {
					bt, _ := block["type"].(string)
					if bt == "text" {
						if t, _ := block["text"].(string); t != "" {
							hasContent = true
							break
						}
					} else {
						hasContent = true
						break
					}
				}
			}
			if hasContent {
				filtered = append(filtered, msg)
			}
			continue
		}

		filtered = append(filtered, msg)
	}

	return filtered
}

func normalizeTools(tools []interface{}) []interface{} {
	var normalized []interface{}

	for _, toolInterface := range tools {
		tool, ok := toolInterface.(map[string]interface{})
		if !ok {
			normalized = append(normalized, toolInterface)
			continue
		}

		if toolType, _ := tool["type"].(string); toolType == "function" {
			if _, hasFn := tool["function"]; hasFn {
				normalized = append(normalized, tool)
				continue
			}
		}

		if name, _ := tool["name"].(string); name != "" {
			if _, hasInputSchema := tool["input_schema"]; hasInputSchema {
				desc, _ := tool["description"].(string)
				params := tool["input_schema"]
				if params == nil {
					params = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
				}
				normalized = append(normalized, map[string]interface{}{
					"type": "function",
					"function": map[string]interface{}{
						"name":        name,
						"description": desc,
						"parameters":  params,
					},
				})
				continue
			}
		}

		if fnDecls, ok := tool["functionDeclarations"].([]interface{}); ok {
			for _, fnInterface := range fnDecls {
				fn, ok := fnInterface.(map[string]interface{})
				if !ok {
					continue
				}
				name, _ := fn["name"].(string)
				desc, _ := fn["description"].(string)
				params := fn["parameters"]
				if params == nil {
					params = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
				}
				normalized = append(normalized, map[string]interface{}{
					"type": "function",
					"function": map[string]interface{}{
						"name":        name,
						"description": desc,
						"parameters":  params,
					},
				})
			}
			continue
		}

		normalized = append(normalized, tool)
	}

	return normalized
}

func normalizeToolChoice(body map[string]interface{}) map[string]interface{} {
	choice, ok := body["tool_choice"].(map[string]interface{})
	if !ok {
		return body
	}

	choiceType, _ := choice["type"].(string)

	switch choiceType {
	case "auto":
		body["tool_choice"] = "auto"
	case "any":
		body["tool_choice"] = "required"
	case "tool":
		if name, _ := choice["name"].(string); name != "" {
			body["tool_choice"] = map[string]interface{}{
				"type": "function",
				"function": map[string]interface{}{
					"name": name,
				},
			}
		}
	}

	return body
}
