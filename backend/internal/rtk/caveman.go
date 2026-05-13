// ref: open-sse/rtk/caveman.js
package rtk

import "strings"

const sep = "\n\n"

// InjectCaveman appends a caveman-style instruction to the system prompt.
// ref: open-sse/rtk/caveman.js:10-29
func InjectCaveman(systemPrompt, level string) string {
	prompt := GetCavemanPrompt(level)
	if prompt == "" || systemPrompt == "" {
		return systemPrompt
	}
	// ref: open-sse/rtk/caveman.js:35-37 (append pattern)
	return systemPrompt + sep + prompt
}

// InjectCavemanIntoMessages injects caveman prompt into OpenAI-style messages.
// Returns the modified messages or creates a new system message if none exists.
// ref: open-sse/rtk/caveman.js:32-52
func InjectCavemanIntoMessages(messages []map[string]any, level string) []map[string]any {
	if len(messages) == 0 {
		return messages
	}

	prompt := GetCavemanPrompt(level)
	if prompt == "" {
		return messages
	}

	// Find system or developer message
	// ref: open-sse/rtk/caveman.js:46-51
	for i, msg := range messages {
		role, _ := msg["role"].(string)
		if role == "system" || role == "developer" {
			messages[i] = appendToOpenAIMessage(msg, prompt)
			return messages
		}
	}

	// No system message found, prepend one
	// ref: open-sse/rtk/caveman.js:50
	newMsg := map[string]any{
		"role":    "system",
		"content": prompt,
	}
	return append([]map[string]any{newMsg}, messages...)
}

// appendToOpenAIMessage appends prompt to an OpenAI message.
// ref: open-sse/rtk/caveman.js:54-63
func appendToOpenAIMessage(msg map[string]any, prompt string) map[string]any {
	result := make(map[string]any)
	for k, v := range msg {
		result[k] = v
	}

	switch content := msg["content"].(type) {
	case string:
		// ref: open-sse/rtk/caveman.js:55-56
		result["content"] = content + sep + prompt
	case []any:
		// ref: open-sse/rtk/caveman.js:57-59 (Responses-style array)
		newContent := make([]any, len(content))
		copy(newContent, content)
		newContent = append(newContent, map[string]any{
			"type": "input_text",
			"text": prompt,
		})
		result["content"] = newContent
	default:
		result["content"] = prompt
	}

	return result
}

// InjectCavemanIntoClaudeSystem injects caveman into Claude-style system.
// Claude system is either a string or an array of blocks.
// ref: open-sse/rtk/caveman.js:67-86
func InjectCavemanIntoClaudeSystem(system any, level string) any {
	if system == nil {
		return GetCavemanPrompt(level)
	}

	prompt := GetCavemanPrompt(level)
	if prompt == "" {
		return system
	}

	switch s := system.(type) {
	case string:
		if s == "" {
			return prompt
		}
		// ref: open-sse/rtk/caveman.js:68-70
		return s + sep + prompt
	case []any:
		// ref: open-sse/rtk/caveman.js:72-83
		block := map[string]any{
			"type": "text",
			"text": prompt,
		}

		// Find last cache block with remove
		// ref: open-sse/rtk/caveman.js:74-77
		lastCacheIdx := -1
		for i := len(s) - 1; i >= 0; i-- {
			if block, ok := s[i].(map[string]any); ok {
				if _, hasRemove := block["remove"]; hasRemove {
					lastCacheIdx = i
					break
				}
			}
		}

		// Insert before last cache block or append
		// ref: open-sse/rtk/caveman.js:78-82
		result := make([]any, len(s))
		copy(result, s)
		if lastCacheIdx >= 0 {
			result = append(result[:lastCacheIdx], append([]any{block}, result[lastCacheIdx:]...)...)
		} else {
			result = append(result, block)
		}
		return result
	default:
		return prompt
	}
}

// InjectCavemanIntoGeminiSystem injects caveman into Gemini-style system instruction.
// ref: open-sse/rtk/caveman.js:90-100
func InjectCavemanIntoGeminiSystem(body map[string]any, level string) {
	prompt := GetCavemanPrompt(level)
	if prompt == "" {
		return
	}

	// Determine target (body or body.request for antigravity)
	// ref: open-sse/rtk/caveman.js:91
	target := body
	if req, ok := body["request"].(map[string]any); ok {
		target = req
	}

	// Determine key (snake_case or camelCase)
	// ref: open-sse/rtk/caveman.js:92-93
	key := "systemInstruction"
	if _, hasSnake := target["system_instruction"]; hasSnake {
		key = "system_instruction"
	}

	// Inject into parts
	// ref: open-sse/rtk/caveman.js:94-99
	if sys, ok := target[key].(map[string]any); ok {
		if parts, ok := sys["parts"].([]any); ok {
			sys["parts"] = append(parts, map[string]any{"text": prompt})
			return
		}
	}

	target[key] = map[string]any{
		"parts": []any{map[string]any{"text": prompt}},
	}
}

// InjectCavemanIntoInstructions injects caveman into OpenAI Responses instructions field.
// ref: open-sse/rtk/caveman.js:34-38
func InjectCavemanIntoInstructions(instructions, level string) string {
	if instructions == "" {
		return GetCavemanPrompt(level)
	}
	prompt := GetCavemanPrompt(level)
	if prompt == "" {
		return instructions
	}
	return instructions + sep + prompt
}

// HasSystemMessage checks if messages contain a system or developer message.
func HasSystemMessage(messages []map[string]any) bool {
	for _, msg := range messages {
		role, _ := msg["role"].(string)
		if role == "system" || role == "developer" {
			return true
		}
	}
	return false
}

// IsArrayContent checks if a message content is an array.
func IsArrayContent(content any) bool {
	_, ok := content.([]any)
	return ok
}

// IsStringContent checks if a message content is a string.
func IsStringContent(content any) bool {
	_, ok := content.(string)
	return ok
}

// stripArticles removes leading articles from text (helper for ultra level).
// Not used in injection but available for future use.
func stripArticles(text string) string {
	articles := []string{"a ", "an ", "the ", "A ", "An ", "The "}
	for _, article := range articles {
		if strings.HasPrefix(text, article) {
			return text[len(article):]
		}
	}
	return text
}
