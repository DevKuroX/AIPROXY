package response

import (
	"encoding/json"
	"fmt"
	"time"
)

// ref: open-sse/translator/request/gemini-to-openai.js

type GeminiResponse struct {
	Candidates    []GeminiCandidate `json:"candidates"`
	UsageMetadata *GeminiUsageMeta  `json:"usageMetadata,omitempty"`
}

type GeminiCandidate struct {
	Content       GeminiContent `json:"content"`
	FinishReason  string        `json:"finishReason,omitempty"`
	SafetyRatings []interface{} `json:"safetyRatings,omitempty"`
}

type GeminiUsageMeta struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type GeminiContent struct {
	Role  string       `json:"role,omitempty"`
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text             string                 `json:"text,omitempty"`
	FunctionCall     *GeminiFunctionCall    `json:"functionCall,omitempty"`
	FunctionResponse *GeminiFunctionResponse `json:"functionResponse,omitempty"`
}

type GeminiFunctionCall struct {
	ID   string                 `json:"id,omitempty"`
	Name string                 `json:"name"`
	Args map[string]interface{} `json:"args,omitempty"`
}

type GeminiFunctionResponse struct {
	ID       string                 `json:"id,omitempty"`
	Name     string                 `json:"name,omitempty"`
	Response map[string]interface{} `json:"response,omitempty"`
}

// ref: open-sse/translator/request/gemini-to-openai.js:6-69
func TranslateGeminiToOpenAIResponse(body *GeminiResponse, model string) (*OpenAIResponse, error) {
	if body == nil || len(body.Candidates) == 0 {
		return nil, nil
	}

	result := &OpenAIResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []OpenAIChoice{},
	}

	candidate := body.Candidates[0]

	var content string
	var toolCalls []OpenAIToolCall
	toolCallIndex := 0

	for _, part := range candidate.Content.Parts {
		if part.Text != "" {
			content += part.Text
		}

		if part.FunctionCall != nil {
			args := "{}"
			if len(part.FunctionCall.Args) > 0 {
				argsBytes, _ := json.Marshal(part.FunctionCall.Args)
				args = string(argsBytes)
			}

			toolCalls = append(toolCalls, OpenAIToolCall{
				Index: toolCallIndex,
				ID:    part.FunctionCall.ID,
				Type:  "function",
				Function: OpenAIFunctionCall{
					Name:      part.FunctionCall.Name,
					Arguments: args,
				},
			})
			toolCallIndex++
		}
	}

	finishReason := convertGeminiFinishReason(candidate.FinishReason)

	choice := OpenAIChoice{
		Index:        0,
		Message: OpenAIMessage{
			Role: "assistant",
		},
		FinishReason: finishReason,
	}

	if len(toolCalls) > 0 {
		choice.Message.ToolCalls = toolCalls
	}
	choice.Message.Content = content

	result.Choices = append(result.Choices, choice)

	if body.UsageMetadata != nil {
		result.Usage = OpenAIUsage{
			PromptTokens:     body.UsageMetadata.PromptTokenCount,
			CompletionTokens: body.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      body.UsageMetadata.TotalTokenCount,
		}
	}

	return result, nil
}

// ref: open-sse/translator/response/claude-to-openai.js:194-202
func convertGeminiFinishReason(reason string) string {
	switch reason {
	case "STOP":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	case "SAFETY":
		return "content_filter"
	case "RECITATION":
		return "content_filter"
	case "TOOL_CODE":
		return "tool_calls"
	default:
		return "stop"
	}
}
