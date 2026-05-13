package request

import (
	"encoding/json"
	"regexp"
	"strings"
)

// ref: open-sse/translator/request/openai-to-gemini.js

var defaultSafetySettings = []GeminiSafety{
	{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_NONE"},
	{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_NONE"},
	{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_NONE"},
	{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_NONE"},
}

// ref: open-sse/translator/request/openai-to-gemini.js:26-36
func sanitizeGeminiFunctionName(name string) string {
	if name == "" {
		return "_unknown"
	}
	re := regexp.MustCompile(`[^a-zA-Z0-9_.:\-]`)
	sanitized := re.ReplaceAllString(name, "_")
	if !regexp.MustCompile(`^[a-zA-Z_]`).MatchString(sanitized) {
		sanitized = "_" + sanitized
	}
	if len(sanitized) > 64 {
		sanitized = sanitized[:64]
	}
	return sanitized
}

// ref: open-sse/translator/request/openai-to-gemini.js:39-221
func TranslateOpenAIToGeminiRequest(model string, body *OpenAIRequest, stream bool) *GeminiRequest {
	result := &GeminiRequest{
		Model:          model,
		Contents:       []GeminiContent{},
		GenerationConfig: &GeminiGenConfig{},
		SafetySettings: defaultSafetySettings,
	}

	if body.Temperature != nil {
		result.GenerationConfig.Temperature = body.Temperature
	}
	if body.TopP != nil {
		result.GenerationConfig.TopP = body.TopP
	}
	if body.MaxTokens > 0 {
		result.GenerationConfig.MaxOutputTokens = body.MaxTokens
	}

	tcID2Name := make(map[string]string)
	if len(body.Messages) > 0 {
		for _, msg := range body.Messages {
			if msg.Role == "assistant" && len(msg.ToolCalls) > 0 {
				for _, tc := range msg.ToolCalls {
				if tc.Type == "function" && tc.ID != "" && tc.Function.Name != "" {
					tcID2Name[tc.ID] = tc.Function.Name
				}
				}
			}
		}
	}

	toolResponses := make(map[string]interface{})
	if len(body.Messages) > 0 {
		for _, msg := range body.Messages {
			if msg.Role == "tool" && msg.ToolCallID != "" {
				toolResponses[msg.ToolCallID] = msg.Content
			}
		}
	}

	if len(body.Messages) > 0 {
		for i, msg := range body.Messages {
			role := msg.Role
			content := msg.Content

			if role == "system" && len(body.Messages) > 1 {
				text := extractTextContent(content)
				if text != "" {
					result.SystemInstruction = &GeminiContent{
						Role:  "user",
						Parts: []GeminiPart{{Text: text}},
					}
				}
			} else if role == "user" || (role == "system" && len(body.Messages) == 1) {
				parts := convertOpenAIContentToParts(content)
				if len(parts) > 0 {
					result.Contents = append(result.Contents, GeminiContent{Role: "user", Parts: parts})
				}
			} else if role == "assistant" {
				parts := []GeminiPart{}

				if content != nil {
					text := extractTextContent(content)
					if text != "" {
						parts = append(parts, GeminiPart{Text: text})
					}
				}

				if len(msg.ToolCalls) > 0 {
					toolCallIDs := []string{}
					for _, tc := range msg.ToolCalls {
					if tc.Type != "function" || tc.Function.Name == "" {
						continue
					}

						var args map[string]interface{}
						if tc.Function.Arguments != "" {
							json.Unmarshal([]byte(tc.Function.Arguments), &args)
						}
						if args == nil {
							args = make(map[string]interface{})
						}

						parts = append(parts, GeminiPart{
							FunctionCall: &GeminiFunctionCall{
								ID:   tc.ID,
								Name: sanitizeGeminiFunctionName(tc.Function.Name),
								Args: args,
							},
						})
						toolCallIDs = append(toolCallIDs, tc.ID)
					}

					if len(parts) > 0 {
						result.Contents = append(result.Contents, GeminiContent{Role: "model", Parts: parts})
					}

					hasActualResponses := false
					for _, fid := range toolCallIDs {
						if _, ok := toolResponses[fid]; ok {
							hasActualResponses = true
							break
						}
					}

					if hasActualResponses {
						toolParts := []GeminiPart{}
						for _, fid := range toolCallIDs {
							resp, ok := toolResponses[fid]
							if !ok {
								continue
							}

							name := tcID2Name[fid]
							if name == "" {
								idParts := strings.Split(fid, "-")
								if len(idParts) > 2 {
									name = strings.Join(idParts[:len(idParts)-2], "-")
								} else {
									name = fid
								}
							}

							var parsedResp map[string]interface{}
							switch v := resp.(type) {
							case string:
								json.Unmarshal([]byte(v), &parsedResp)
								if parsedResp == nil {
									parsedResp = map[string]interface{}{"result": v}
								}
							case map[string]interface{}:
								parsedResp = v
							default:
								parsedResp = map[string]interface{}{"result": resp}
							}

							toolParts = append(toolParts, GeminiPart{
								FunctionResponse: &GeminiFunctionResponse{
									ID:   fid,
									Name: sanitizeGeminiFunctionName(name),
									Response: map[string]interface{}{
										"result": parsedResp,
									},
								},
							})
						}
						if len(toolParts) > 0 {
							result.Contents = append(result.Contents, GeminiContent{Role: "user", Parts: toolParts})
						}
					}
				} else if len(parts) > 0 {
					result.Contents = append(result.Contents, GeminiContent{Role: "model", Parts: parts})
				}
			}
			_ = i
		}
	}

	if len(body.Tools) > 0 {
		functionDeclarations := []GeminiFunctionDecl{}
		for _, t := range body.Tools {
			if t.Type == "function" && t.Function != nil {
				params := t.Function.Parameters
				if params == nil {
					params = map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
				}
				functionDeclarations = append(functionDeclarations, GeminiFunctionDecl{
					Name:        sanitizeGeminiFunctionName(t.Function.Name),
					Description: t.Function.Description,
					Parameters:  params,
				})
			}
		}
		if len(functionDeclarations) > 0 {
			result.Tools = []GeminiTool{{FunctionDeclarations: functionDeclarations}}
		}
	}

	return result
}

// ref: open-sse/translator/helpers/geminiHelper.js:convertOpenAIContentToParts
func convertOpenAIContentToParts(content interface{}) []GeminiPart {
	parts := []GeminiPart{}

	if content == nil {
		return parts
	}

	switch v := content.(type) {
	case string:
		if v != "" {
			parts = append(parts, GeminiPart{Text: v})
		}
	case []interface{}:
		for _, item := range v {
			if m, ok := item.(map[string]interface{}); ok {
				partType, _ := m["type"].(string)
				switch partType {
				case "text":
					if text, ok := m["text"].(string); ok && text != "" {
						parts = append(parts, GeminiPart{Text: text})
					}
				case "image_url":
					if imgURL, ok := m["image_url"].(map[string]interface{}); ok {
						if url, ok := imgURL["url"].(string); ok && strings.HasPrefix(url, "data:") {
							parts = parseDataURL(url, parts)
						}
					}
				}
			}
		}
	case map[string]interface{}:
		if text, ok := v["text"].(string); ok && text != "" {
			parts = append(parts, GeminiPart{Text: text})
		}
	}

	return parts
}

func parseDataURL(url string, parts []GeminiPart) []GeminiPart {
	if !strings.HasPrefix(url, "data:") {
		return parts
	}

	rest := url[5:]
	semicolonIdx := strings.Index(rest, ";")
	if semicolonIdx == -1 {
		return parts
	}
	mimeType := rest[:semicolonIdx]

	commaIdx := strings.Index(rest, ",")
	if commaIdx == -1 {
		return parts
	}
	data := rest[commaIdx+1:]

	if data != "" {
		parts = append(parts, GeminiPart{
			InlineData: &GeminiInlineData{
				MimeType: mimeType,
				Data:     data,
			},
		})
	}

	return parts
}
