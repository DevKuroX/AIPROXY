package services

import (
	"encoding/json"
	"regexp"
	"strings"
)

// CompactService handles response compaction for bandwidth optimization.
// Removes thinking blocks and reduces verbosity in API responses.
// ref: 9router/open-sse/executors/codex.js:155-156 (_compact flag handling)
// ref: 9router/src/app/api/v1/responses/compact/route.js (compact endpoint)
// ref: 9router/open-sse/translator/helpers/openaiHelper.js:8-31 (thinking block removal)
type CompactService struct {
	logger Logger
}

// NewCompactService creates a new CompactService instance.
func NewCompactService(logger Logger) *CompactService {
	return &CompactService{
		logger: logger,
	}
}

// CompactResponse compacts a full API response by removing thinking blocks
// and reducing verbosity. Works with both streaming and non-streaming formats.
// ref: 9router/open-sse/translator/helpers/openaiHelper.js:8-57
// Returns a compacted copy of the response without modifying the original.
func (s *CompactService) CompactResponse(response map[string]any) map[string]any {
	if response == nil {
		return nil
	}

	// Create a copy to avoid modifying original
	result := make(map[string]any)
	for k, v := range response {
		result[k] = v
	}

	// Handle non-streaming response format
	// ref: 9router/open-sse/translator/helpers/openaiHelper.js:26-54
	if choices, ok := result["choices"].([]any); ok {
		result["choices"] = s.compactChoices(choices)
	}

	// Handle streaming chunk format (has delta instead of message)
	// ref: 9router/open-sse/handlers/chatCore/streamingHandler.js
	if choices, ok := result["choices"].([]any); ok {
		for i, choice := range choices {
			if choiceMap, ok := choice.(map[string]any); ok {
				if delta, ok := choiceMap["delta"].(map[string]any); ok {
					choiceMap["delta"] = s.compactDelta(delta)
					choices[i] = choiceMap
				}
			}
		}
	}

	// Compact usage data
	// ref: 9router/open-sse/services/usage.js
	if usage, ok := result["usage"].(map[string]any); ok {
		result["usage"] = s.CompactUsage(usage)
	}

	// Handle Responses API format (input/output instead of messages/choices)
	// ref: 9router/open-sse/executors/codex.js
	if output, ok := result["output"].([]any); ok {
		result["output"] = s.compactOutput(output)
	}

	return result
}

// compactChoices compacts the choices array in a non-streaming response.
// ref: 9router/open-sse/translator/helpers/openaiHelper.js:26-54
func (s *CompactService) compactChoices(choices []any) []any {
	if choices == nil {
		return nil
	}

	result := make([]any, len(choices))
	for i, choice := range choices {
		if choiceMap, ok := choice.(map[string]any); ok {
			compactedChoice := make(map[string]any)
			for k, v := range choiceMap {
				compactedChoice[k] = v
			}

			// Compact message content
			if message, ok := compactedChoice["message"].(map[string]any); ok {
				compactedChoice["message"] = s.compactMessage(message)
			}

			result[i] = compactedChoice
		} else {
			result[i] = choice
		}
	}
	return result
}

// compactMessage compacts a message object by removing thinking blocks.
// ref: 9router/open-sse/translator/helpers/openaiHelper.js:26-54
func (s *CompactService) compactMessage(message map[string]any) map[string]any {
	if message == nil {
		return nil
	}

	result := make(map[string]any)
	for k, v := range message {
		result[k] = v
	}

	// Handle string content - remove thinking blocks from text
	// ref: 9router/open-sse/translator/helpers/openaiHelper.js:23
	if content, ok := result["content"].(string); ok {
		result["content"] = s.RemoveThinking(content)
	}

	// Handle array content - filter out thinking blocks
	// ref: 9router/open-sse/translator/helpers/openaiHelper.js:26-54
	if content, ok := result["content"].([]any); ok {
		result["content"] = s.compactContentArray(content)
	}

	return result
}

// compactDelta compacts a streaming delta object.
// ref: 9router/open-sse/handlers/chatCore/streamingHandler.js
func (s *CompactService) compactDelta(delta map[string]any) map[string]any {
	if delta == nil {
		return nil
	}

	result := make(map[string]any)
	for k, v := range delta {
		result[k] = v
	}

	// Handle string content
	if content, ok := result["content"].(string); ok {
		result["content"] = s.RemoveThinking(content)
	}

	// Handle array content
	if content, ok := result["content"].([]any); ok {
		result["content"] = s.compactContentArray(content)
	}

	return result
}

// compactContentArray filters out thinking blocks from content array.
// ref: 9router/open-sse/translator/helpers/openaiHelper.js:26-46
func (s *CompactService) compactContentArray(content []any) []any {
	if content == nil {
		return nil
	}

	var result []any
	for _, block := range content {
		if blockMap, ok := block.(map[string]any); ok {
			blockType, _ := blockMap["type"].(string)

			// Skip thinking blocks
			// ref: 9router/open-sse/translator/helpers/openaiHelper.js:30-31
			if blockType == "thinking" || blockType == "redacted_thinking" {
				continue
			}

			// Clean the block by removing signature and other non-essential fields
			// ref: 9router/open-sse/translator/helpers/openaiHelper.js:35-37
			cleanBlock := s.compactContentBlock(blockMap)
			result = append(result, cleanBlock)
		} else {
			result = append(result, block)
		}
	}

	// If all content was filtered, add empty text
	// ref: 9router/open-sse/translator/helpers/openaiHelper.js:49-51
	if len(result) == 0 {
		result = append(result, map[string]any{
			"type": "text",
			"text": "",
		})
	}

	return result
}

// compactContentBlock removes non-essential fields from a content block.
// ref: 9router/open-sse/translator/helpers/openaiHelper.js:35-37
func (s *CompactService) compactContentBlock(block map[string]any) map[string]any {
	if block == nil {
		return nil
	}

	result := make(map[string]any)
	for k, v := range block {
		// Remove signature field (internal use only)
		// ref: 9router/open-sse/translator/helpers/openaiHelper.js:35-36
		if k == "signature" {
			continue
		}
		result[k] = v
	}
	return result
}

// compactOutput compacts the output array in Responses API format.
// ref: 9router/open-sse/executors/codex.js
func (s *CompactService) compactOutput(output []any) []any {
	if output == nil {
		return nil
	}

	var result []any
	for _, item := range output {
		if itemMap, ok := item.(map[string]any); ok {
			itemType, _ := itemMap["type"].(string)

			// Skip thinking items
			if itemType == "thinking" || itemType == "redacted_thinking" {
				continue
			}

			// Compact the item
			compactedItem := make(map[string]any)
			for k, v := range itemMap {
				compactedItem[k] = v
			}

			// Handle content array in output items
			if content, ok := compactedItem["content"].([]any); ok {
				compactedItem["content"] = s.compactContentArray(content)
			}

			result = append(result, compactedItem)
		} else {
			result = append(result, item)
		}
	}

	return result
}

// RemoveThinking removes thinking blocks from text content.
// Thinking blocks are typically enclosed in special markers like:
// - <|thinking|>...</|thinking|>
// - <thinking>...</thinking>
// - <|start|>thinking<|message|>...<|end|>
// ref: 9router/open-sse/translator/helpers/openaiHelper.js:30-31
func (s *CompactService) RemoveThinking(content string) string {
	if content == "" {
		return content
	}

	// Remove <|thinking|>...</|thinking|> blocks
	// ref: Common thinking block format
	thinkingPattern1 := regexp.MustCompile(`<\|thinking\|>[\s\S]*?<\|/thinking\|>`)
	content = thinkingPattern1.ReplaceAllString(content, "")

	// Remove <thinking>...</thinking> blocks (XML-style)
	thinkingPattern2 := regexp.MustCompile(`<thinking>[\s\S]*?</thinking>`)
	content = thinkingPattern2.ReplaceAllString(content, "")

	// Remove <|start|>thinking<|message|>...<|end|> blocks (Codex style)
	// ref: 9router/open-sse/executors/codex.js
	thinkingPattern3 := regexp.MustCompile(`<\|start\|>thinking<\|message\|>[\s\S]*?<\|end\|>`)
	content = thinkingPattern3.ReplaceAllString(content, "")

	// Remove redacted thinking blocks
	redactedPattern := regexp.MustCompile(`<\|redacted_thinking\|>[\s\S]*?<\|/redacted_thinking\|>`)
	content = redactedPattern.ReplaceAllString(content, "")

	// Clean up extra whitespace
	content = strings.TrimSpace(content)

	return content
}

// CompactUsage reduces usage data to essential fields only.
// ref: 9router/open-sse/services/usage.js
func (s *CompactService) CompactUsage(usage map[string]any) map[string]any {
	if usage == nil {
		return nil
	}

	// Keep only essential usage fields
	result := make(map[string]any)

	// Essential token counts
	if v, ok := usage["prompt_tokens"]; ok {
		result["prompt_tokens"] = v
	}
	if v, ok := usage["completion_tokens"]; ok {
		result["completion_tokens"] = v
	}
	if v, ok := usage["total_tokens"]; ok {
		result["total_tokens"] = v
	}

	// Keep input/output tokens (Responses API format)
	if v, ok := usage["input_tokens"]; ok {
		result["input_tokens"] = v
	}
	if v, ok := usage["output_tokens"]; ok {
		result["output_tokens"] = v
	}

	return result
}

// CompactStreamingLine compacts a single streaming SSE line.
// Handles "data: {...}" format for streaming responses.
// ref: 9router/open-sse/handlers/chatCore/streamingHandler.js
func (s *CompactService) CompactStreamingLine(line string) string {
	if line == "" {
		return line
	}

	// Handle SSE data lines
	if strings.HasPrefix(line, "data: ") {
		jsonPart := strings.TrimPrefix(line, "data: ")

		// Skip [DONE] markers
		if jsonPart == "[DONE]" {
			return line
		}

		var response map[string]any
		if err := json.Unmarshal([]byte(jsonPart), &response); err == nil {
			compacted := s.CompactResponse(response)
			if compactedJSON, err := json.Marshal(compacted); err == nil {
				return "data: " + string(compactedJSON)
			}
		}
	}

	return line
}

// CompactStreamingChunk compacts a streaming chunk (used in SSE-to-JSON handler).
// ref: 9router/open-sse/handlers/chatCore/sseToJsonHandler.js
func (s *CompactService) CompactStreamingChunk(chunk map[string]any) map[string]any {
	return s.CompactResponse(chunk)
}
