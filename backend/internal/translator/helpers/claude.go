// Package helpers provides utility functions for translator operations.
// ref: open-sse/translator/helpers/claudeHelper.js
package helpers

var ClaudeFormatProvidersWithoutOutputConfig = map[string]bool{
	"minimax":    true,
	"minimax-cn": true,
}

func HasValidContent(msg map[string]interface{}) bool {
	if msg == nil {
		return false
	}

	content := msg["content"]

	if str, ok := content.(string); ok {
		return len(str) > 0 && str != ""
	}

	if arr, ok := content.([]interface{}); ok {
		for _, item := range arr {
			if block, ok := item.(map[string]interface{}); ok {
				blockType, _ := block["type"].(string)
				if blockType == "text" {
					if text, _ := block["text"].(string); text != "" {
						return true
					}
				}
				if blockType == "tool_use" || blockType == "tool_result" {
					return true
				}
			}
		}
	}

	return false
}

func FixToolUseOrdering(messages []interface{}) []interface{} {
	if len(messages) <= 1 {
		return messages
	}

	for _, msgInterface := range messages {
		msg, ok := msgInterface.(map[string]interface{})
		if !ok {
			continue
		}

		role, _ := msg["role"].(string)
		if role != "assistant" {
			continue
		}

		content, ok := msg["content"].([]interface{})
		if !ok {
			continue
		}

		hasToolUse := false
		for _, b := range content {
			if block, ok := b.(map[string]interface{}); ok {
				if blockType, _ := block["type"].(string); blockType == "tool_use" {
					hasToolUse = true
					break
				}
			}
		}

		if !hasToolUse {
			continue
		}

		var newContent []interface{}
		foundToolUse := false

		for _, b := range content {
			block, ok := b.(map[string]interface{})
			if !ok {
				continue
			}

			blockType, _ := block["type"].(string)

			if blockType == "tool_use" {
				foundToolUse = true
				newContent = append(newContent, block)
			} else if blockType == "thinking" || blockType == "redacted_thinking" {
				newContent = append(newContent, block)
			} else if !foundToolUse {
				newContent = append(newContent, block)
			}
		}

		msg["content"] = newContent
	}

	var merged []interface{}

	for _, msgInterface := range messages {
		msg, ok := msgInterface.(map[string]interface{})
		if !ok {
			merged = append(merged, msgInterface)
			continue
		}

		if len(merged) == 0 {
			merged = append(merged, msg)
			continue
		}

		last := merged[len(merged)-1]
		lastMsg, ok := last.(map[string]interface{})
		if !ok {
			merged = append(merged, msg)
			continue
		}

		lastRole, _ := lastMsg["role"].(string)
		msgRole, _ := msg["role"].(string)

		if lastRole != msgRole {
			merged = append(merged, msg)
			continue
		}

		lastContent := normalizeContentToArray(lastMsg["content"])
		msgContent := normalizeContentToArray(msg["content"])

		var toolResults []interface{}
		var otherContent []interface{}

		for _, b := range lastContent {
			if block, ok := b.(map[string]interface{}); ok {
				if blockType, _ := block["type"].(string); blockType == "tool_result" {
					toolResults = append(toolResults, block)
				} else {
					otherContent = append(otherContent, block)
				}
			}
		}
		for _, b := range msgContent {
			if block, ok := b.(map[string]interface{}); ok {
				if blockType, _ := block["type"].(string); blockType == "tool_result" {
					toolResults = append(toolResults, block)
				} else {
					otherContent = append(otherContent, block)
				}
			}
		}

		mergedContent := append(toolResults, otherContent...)
		lastMsg["content"] = mergedContent
	}

	return merged
}

func normalizeContentToArray(content interface{}) []interface{} {
	if content == nil {
		return nil
	}

	if arr, ok := content.([]interface{}); ok {
		return arr
	}

	if str, ok := content.(string); ok {
		return []interface{}{
			map[string]interface{}{
				"type": "text",
				"text": str,
			},
		}
	}

	return nil
}

func PrepareCL4udeRequest(body map[string]interface{}, provider string) map[string]interface{} {
	if body == nil {
		return body
	}

	if ClaudeFormatProvidersWithoutOutputConfig[provider] {
		delete(body, "output_config")
	}

	if system, ok := body["system"].([]interface{}); ok {
		for i, blockInterface := range system {
			if block, ok := blockInterface.(map[string]interface{}); ok {
				delete(block, "remove")
				if i == len(system)-1 {
					block["remove"] = map[string]interface{}{
						"type": "ephemeral",
						"ttl":  "1h",
					}
				}
				system[i] = block
			}
		}
		body["system"] = system
	}

	if messages, ok := body["messages"].([]interface{}); ok {
		var filtered []interface{}
		msgLen := len(messages)

		for i, msgInterface := range messages {
			msg, ok := msgInterface.(map[string]interface{})
			if !ok {
				continue
			}

			if content, ok := msg["content"].([]interface{}); ok {
				for _, blockInterface := range content {
					if block, ok := blockInterface.(map[string]interface{}); ok {
						delete(block, "remove")
					}
				}
			}

			role, _ := msg["role"].(string)
			isFinalAssistant := i == msgLen-1 && role == "assistant"

			if isFinalAssistant || HasValidContent(msg) {
				filtered = append(filtered, msg)
			}
		}

		filtered = FixToolUseOrdering(filtered)
		body["messages"] = filtered
	}

	if tools, ok := body["tools"].([]interface{}); ok {
		if provider != "claude" {
			var filteredTools []interface{}
			for _, toolInterface := range tools {
				if tool, ok := toolInterface.(map[string]interface{}); ok {
					toolType, _ := tool["type"].(string)
					if toolType == "" || toolType == "function" {
						filteredTools = append(filteredTools, tool)
					}
				}
			}
			tools = filteredTools
		}

		for i, toolInterface := range tools {
			if tool, ok := toolInterface.(map[string]interface{}); ok {
				delete(tool, "remove")
				if i == len(tools)-1 {
					tool["remove"] = map[string]interface{}{
						"type": "ephemeral",
						"ttl":  "1h",
					}
				}
				tools[i] = tool
			}
		}

		if len(tools) == 0 {
			delete(body, "tools")
			delete(body, "tool_choice")
		} else {
			body["tools"] = tools
		}
	}

	return body
}
