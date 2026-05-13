package request

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const CLaudeSystemPrompt = "You are Code Assistant, a highly skilled software development AI with expertise in all programming languages, frameworks, and software engineering practices."

func TranslateRequest(model string, body *OpenAIRequest, stream bool) (*CLaudeRequest, error) {
	result := &CLaudeRequest{
		Model:   model,
		Stream:  stream,
		MaxTokens: adjustMaxTokens(body),
	}

	if body.Temperature != nil {
		result.Temperature = body.Temperature
	}

	if body.TopP != nil {
		result.TopP = body.TopP
	}

	if len(body.Stop) > 0 {
		result.StopSequences = body.Stop
	}

	systemParts := []string{}

	if len(body.Messages) > 0 {
		for _, msg := range body.Messages {
			if msg.Role == "system" {
				text := extractTextContent(msg.Content)
				if text != "" {
					systemParts = append(systemParts, text)
				}
			}
		}

		var nonSystemMessages []OpenAIMessage
		for _, m := range body.Messages {
			if m.Role != "system" {
				nonSystemMessages = append(nonSystemMessages, m)
			}
		}

		result.Messages = []CLaudeMessage{}
		var currentRole string
		var currentParts []CLaudeContent

		flushCurrentMessage := func() {
			if currentRole != "" && len(currentParts) > 0 {
				result.Messages = append(result.Messages, CLaudeMessage{
					Role:    currentRole,
					Content: currentParts,
				})
				currentParts = nil
			}
		}

		for _, msg := range nonSystemMessages {
			newRole := msg.Role
			if newRole == "tool" {
				newRole = "user"
			}

			blocks := getContentBlocksFromMessage(msg)

			hasToolResult := false
			for _, b := range blocks {
				if b.Type == "tool_result" {
					hasToolResult = true
					break
				}
			}

			if hasToolResult {
				var toolResultBlocks []CLaudeContent
				var otherBlocks []CLaudeContent

				for _, b := range blocks {
					if b.Type == "tool_result" {
						toolResultBlocks = append(toolResultBlocks, b)
					} else {
						otherBlocks = append(otherBlocks, b)
					}
				}

				flushCurrentMessage()

				if len(toolResultBlocks) > 0 {
					result.Messages = append(result.Messages, CLaudeMessage{
						Role:    "user",
						Content: toolResultBlocks,
					})
				}

				if len(otherBlocks) > 0 {
					currentRole = newRole
					currentParts = otherBlocks
				}
				continue
			}

			if currentRole != newRole {
				flushCurrentMessage()
				currentRole = newRole
			}

			currentParts = append(currentParts, blocks...)

			hasToolUse := false
			for _, b := range blocks {
				if b.Type == "tool_use" {
					hasToolUse = true
					break
				}
			}
			if hasToolUse {
				flushCurrentMessage()
			}
		}

		flushCurrentMessage()

		for i := len(result.Messages) - 1; i >= 0; i-- {
			msg := result.Messages[i]
			if msg.Role == "assistant" && len(msg.Content) > 0 {
				for j := len(msg.Content) - 1; j >= 0; j-- {
					block := msg.Content[j]
					if block.Type == "text" || block.Type == "tool_use" || block.Type == "image" {
						msg.Content[j].Remove = &CLaudeRemove{Type: "ephemeral"}
						break
					}
				}
				break
			}
		}
	}

	if body.ResponseFormat != nil {
		if body.ResponseFormat.Type == "json_schema" && body.ResponseFormat.JSONSchema != nil {
			schemaJSON, _ := json.MarshalIndent(body.ResponseFormat.JSONSchema.Schema, "", "  ")
			systemParts = append(systemParts, fmt.Sprintf("You must respond with valid JSON that strictly follows this JSON schema:\n```json\n%s\n```\nRespond ONLY with the JSON object, no other text.", string(schemaJSON)))
		} else if body.ResponseFormat.Type == "json_object" {
			systemParts = append(systemParts, "You must respond with valid JSON. Respond ONLY with a JSON object, no other text.")
		}
	}

	result.System = []CLaudeSystemBlock{
		{Type: "text", Text: CLaudeSystemPrompt},
	}

	if len(systemParts) > 0 {
		result.System = append(result.System, CLaudeSystemBlock{
			Type:   "text",
			Text:   strings.Join(systemParts, "\n"),
			Remove: &CLaudeRemove{Type: "ephemeral", TTL: "1h"},
		})
	}

	if len(body.Tools) > 0 {
		result.Tools = []CLaudeTool{}

		for _, tool := range body.Tools {
			toolType := tool.Type

			if toolType != "" && toolType != "function" {
				var inputSchema map[string]interface{}
				if tool.Function != nil && tool.Function.Parameters != nil {
					inputSchema = tool.Function.Parameters
				} else {
					inputSchema = map[string]interface{}{
						"type":       "object",
						"properties": map[string]interface{}{},
						"required":   []string{},
					}
				}
				result.Tools = append(result.Tools, CLaudeTool{
					Name:        toolType,
					InputSchema: inputSchema,
				})
				continue
			}

			toolData := tool.Function
			if toolData == nil {
				continue
			}

			originalName := toolData.Name
			inputSchema := toolData.Parameters
			if inputSchema == nil {
				inputSchema = map[string]interface{}{
					"type":       "object",
					"properties": map[string]interface{}{},
					"required":   []string{},
				}
			}

			result.Tools = append(result.Tools, CLaudeTool{
				Name:        originalName,
				Description: toolData.Description,
				InputSchema: inputSchema,
			})
		}

		if len(result.Tools) > 0 {
			result.Tools[len(result.Tools)-1].Remove = &CLaudeRemove{Type: "ephemeral", TTL: "1h"}
		}
	}

	if body.ToolChoice != nil {
		result.ToolChoice = convertOpenAIToolChoice(body.ToolChoice)
	}

	if body.Thinking != nil {
		result.Thinking = &CLaudeThinking{
			Type: body.Thinking.Type,
		}
		if body.Thinking.Type == "" {
			result.Thinking.Type = "enabled"
		}
		if body.Thinking.BudgetTokens > 0 {
			result.Thinking.BudgetTokens = body.Thinking.BudgetTokens
		}
		if body.Thinking.MaxTokens > 0 {
			result.Thinking.MaxTokens = body.Thinking.MaxTokens
		}
	}

	if body.ReasoningEffort != "" && result.Thinking == nil {
		effortToBudget := map[string]int{
			"none":   0,
			"low":    4096,
			"medium": 8192,
			"high":   16384,
			"xhigh":  32768,
		}

		budget := effortToBudget[strings.ToLower(body.ReasoningEffort)]
		if budget > 0 {
			result.Thinking = &CLaudeThinking{
				Type:         "enabled",
				BudgetTokens: budget,
			}
		}
	}

	return result, nil
}

func adjustMaxTokens(body *OpenAIRequest) int {
	if body.MaxTokens > 0 {
		return body.MaxTokens
	}
	return 4096
}

func extractTextContent(content interface{}) string {
	if content == nil {
		return ""
	}

	switch v := content.(type) {
	case string:
		return v
	case []interface{}:
		var texts []string
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				if t, ok := m["type"].(string); ok && t == "text" {
					if text, ok := m["text"].(string); ok {
						texts = append(texts, text)
					}
				}
			}
		}
		return strings.Join(texts, "\n")
	}
	return ""
}

func getContentBlocksFromMessage(msg OpenAIMessage) []CLaudeContent {
	var blocks []CLaudeContent

	if msg.Role == "tool" {
		blocks = append(blocks, CLaudeContent{
			Type:      "tool_result",
			ToolUseID: msg.ToolCallID,
			Content:   msg.Content,
		})
		return blocks
	}

	if msg.Role == "user" {
		switch c := msg.Content.(type) {
		case string:
			if c != "" {
				blocks = append(blocks, CLaudeContent{
					Type: "text",
					Text: c,
				})
			}
		case []interface{}:
			for _, part := range c {
				p, ok := part.(map[string]interface{})
				if !ok {
					continue
				}

				partType, _ := p["type"].(string)

				switch partType {
				case "text":
					if text, ok := p["text"].(string); ok && text != "" {
						blocks = append(blocks, CLaudeContent{
							Type: "text",
							Text: text,
						})
					}
				case "tool_result":
					tr := CLaudeContent{
						Type:      "tool_result",
						ToolUseID: getString(p, "tool_use_id"),
						Content:   p["content"],
					}
					if isErr, ok := p["is_error"].(bool); ok {
						tr.IsError = isErr
					}
					blocks = append(blocks, tr)
				case "image_url":
					if imgURL, ok := p["image_url"].(map[string]interface{}); ok {
						url, _ := imgURL["url"].(string)
						blocks = append(blocks, parseImageURL(url)...)
					}
				case "image":
					if source, ok := p["source"].(map[string]interface{}); ok {
						src := &CLaudeImageSource{}
						if t, ok := source["type"].(string); ok {
							src.Type = t
						}
						if mt, ok := source["media_type"].(string); ok {
							src.MediaType = mt
						}
						if d, ok := source["data"].(string); ok {
							src.Data = d
						}
						if u, ok := source["url"].(string); ok {
							src.URL = u
						}
						blocks = append(blocks, CLaudeContent{
							Type:   "image",
							Source: src,
						})
					}
				}
			}
		}
	}

	if msg.Role == "assistant" {
		switch c := msg.Content.(type) {
		case []interface{}:
			for _, part := range c {
				p, ok := part.(map[string]interface{})
				if !ok {
					continue
				}

				partType, _ := p["type"].(string)

				switch partType {
				case "text":
					if text, ok := p["text"].(string); ok && text != "" {
						blocks = append(blocks, CLaudeContent{
							Type: "text",
							Text: text,
						})
					}
				case "tool_use":
					blocks = append(blocks, CLaudeContent{
						Type:  "tool_use",
						ID:    getString(p, "id"),
						Name:  getString(p, "name"),
						Input: p["input"],
					})
				case "thinking":
					blocks = append(blocks, CLaudeContent{
						Type: "thinking",
						Text: getString(p, "text"),
					})
				}
			}
		case string:
			if c != "" {
				blocks = append(blocks, CLaudeContent{
					Type: "text",
					Text: c,
				})
			}
		}

		for _, tc := range msg.ToolCalls {
			if tc.Type == "function" {
				var input interface{}
				if err := json.Unmarshal([]byte(tc.Function.Arguments), &input); err != nil {
					input = tc.Function.Arguments
				}

				blocks = append(blocks, CLaudeContent{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Function.Name,
					Input: input,
				})
			}
		}
	}

	return blocks
}

func convertOpenAIToolChoice(choice interface{}) *CLaudeToolChoice {
	if choice == nil {
		return &CLaudeToolChoice{Type: "auto"}
	}

	switch v := choice.(type) {
	case string:
		switch v {
		case "auto", "none":
			return &CLaudeToolChoice{Type: "auto"}
		case "required":
			return &CLaudeToolChoice{Type: "any"}
		}
	case map[string]interface{}:
		choiceType, _ := v["type"].(string)
		if choiceType == "tool" {
			if fn, ok := v["function"].(map[string]interface{}); ok {
				if name, ok := fn["name"].(string); ok {
					return &CLaudeToolChoice{
						Type: "tool",
						Name: name,
					}
				}
			}
		}
		return &CLaudeToolChoice{Type: choiceType}
	}

	return &CLaudeToolChoice{Type: "auto"}
}

var dataURLRegex = regexp.MustCompile(`^data:([^;]+);base64,(.+)$`)

func parseImageURL(url string) []CLaudeContent {
	matches := dataURLRegex.FindStringSubmatch(url)
	if len(matches) == 3 {
		return []CLaudeContent{{
			Type: "image",
			Source: &CLaudeImageSource{
				Type:      "base64",
				MediaType: matches[1],
				Data:      matches[2],
			},
		}}
	}

	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return []CLaudeContent{{
			Type: "image",
			Source: &CLaudeImageSource{
				Type: "url",
				URL:  url,
			},
		}}
	}

	return nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
