// ref: _ref/9router/open-sse/utils/reasoningContentInjector.js
package utils

import (
	"encoding/json"
	"regexp"
	"strings"
)

type ReasoningBlock struct {
	Type    string `json:"type"`
	Thinking string `json:"thinking,omitempty"`
	Text    string `json:"text,omitempty"`
}

var thinkTagRegex = regexp.MustCompile(`(?s)<think>(.*?)</think>`)

func ExtractReasoningContent(content string) (reasoning, cleanContent string) {
	matches := thinkTagRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return "", content
	}

	var reasoningParts []string
	for _, match := range matches {
		if len(match) > 1 {
			reasoningParts = append(reasoningParts, strings.TrimSpace(match[1]))
		}
	}

	cleanContent = thinkTagRegex.ReplaceAllString(content, "")
	cleanContent = strings.TrimSpace(cleanContent)
	reasoning = strings.Join(reasoningParts, "\n")

	return reasoning, cleanContent
}

func InjectReasoningContent(blocks []interface{}, reasoning string) []interface{} {
	if reasoning == "" {
		return blocks
	}

	reasoningBlock := map[string]interface{}{
		"type":     "thinking",
		"thinking": reasoning,
	}

	result := []interface{}{reasoningBlock}
	result = append(result, blocks...)

	return result
}

func ConvertReasoningToOpenAI(content string, reasoning string) map[string]interface{} {
	return map[string]interface{}{
		"role":              "assistant",
		"content":           content,
		"reasoning_content": reasoning,
	}
}

func ConvertReasoningToCL4ude(content string, reasoning string) []interface{} {
	var blocks []interface{}

	if reasoning != "" {
		blocks = append(blocks, map[string]interface{}{
			"type":     "thinking",
			"thinking": reasoning,
		})
	}

	if content != "" {
		blocks = append(blocks, map[string]interface{}{
			"type": "text",
			"text": content,
		})
	}

	return blocks
}

func ExtractReasoningFromResponse(response map[string]interface{}) (reasoning string, content string) {
	if output, ok := response["output"].([]interface{}); ok {
		for _, block := range output {
			if b, ok := block.(map[string]interface{}); ok {
				switch b["type"] {
				case "thinking":
					if t, ok := b["thinking"].(string); ok {
						reasoning += t
					}
				case "message":
					if c, ok := b["content"].([]interface{}); ok {
						for _, cb := range c {
							if cbMap, ok := cb.(map[string]interface{}); ok {
								if cbMap["type"] == "output_text" {
									if text, ok := cbMap["text"].(string); ok {
										content += text
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return reasoning, content
}

func FormatReasoningForStream(text string, isThinking bool) string {
	if isThinking {
		return text
	}
	return text
}

type ContentExtractor struct {
	buffer string
}

func NewContentExtractor() *ContentExtractor {
	return &ContentExtractor{}
}

func (e *ContentExtractor) ExtractFromCL4ude(chunk map[string]interface{}) (content string, reasoning string, isDone bool) {
	chunkType, _ := chunk["type"].(string)

	switch chunkType {
	case "content_block_start":
		return "", "", false

	case "content_block_delta":
		if delta, ok := chunk["delta"].(map[string]interface{}); ok {
			if text, ok := delta["text"].(string); ok {
				content = text
			}
			if thinking, ok := delta["thinking"].(string); ok {
				reasoning = thinking
			}
		}
		return content, reasoning, false

	case "content_block_stop":
		return "", "", false

	case "message_stop":
		return "", "", true

	default:
		return "", "", false
	}
}

func (e *ContentExtractor) ExtractFromOpenAI(chunk map[string]interface{}) (content string, reasoning string, isDone bool) {
	if choices, ok := chunk["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if delta, ok := choice["delta"].(map[string]interface{}); ok {
				if c, ok := delta["content"].(string); ok {
					content = c
				}
				if r, ok := delta["reasoning_content"].(string); ok {
					reasoning = r
				}
			}
			if finishReason, ok := choice["finish_reason"].(string); ok && finishReason != "" {
				isDone = true
			}
		}
	}
	return content, reasoning, isDone
}

func ParseSSELine(line string) map[string]interface{} {
	line = strings.TrimSpace(line)
	if !strings.HasPrefix(line, "data:") {
		return nil
	}

	data := strings.TrimPrefix(line, "data:")
	data = strings.TrimSpace(data)

	if data == "[DONE]" {
		return map[string]interface{}{"done": true}
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil
	}

	return result
}

func FormatSSE(data interface{}) string {
	jsonData, _ := json.Marshal(data)
	return "data: " + string(jsonData) + "\n\n"
}
