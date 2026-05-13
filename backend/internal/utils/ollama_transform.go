// ref: _ref/9router/open-sse/utils/ollamaTransform.js
package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

type OllamaResponse struct {
	Model    string         `json:"model"`
	Message  OllamaMessage  `json:"message"`
	Done     bool           `json:"done"`
}

type OllamaMessage struct {
	Role      string               `json:"role"`
	Content   string               `json:"content"`
	ToolCalls []OllamaToolCall     `json:"tool_calls,omitempty"`
}

type OllamaToolCall struct {
	Function OllamaFunctionCall `json:"function"`
}

type OllamaFunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type OpenAIChunk struct {
	ID      string        `json:"id,omitempty"`
	Object  string        `json:"object,omitempty"`
	Created int64         `json:"created,omitempty"`
	Model   string        `json:"model,omitempty"`
	Choices []OpenAIChoice `json:"choices"`
}

type OpenAIChoice struct {
	Index        int         `json:"index"`
	Delta        OpenAIDelta `json:"delta"`
	FinishReason string      `json:"finish_reason,omitempty"`
}

type OpenAIDelta struct {
	Role             string          `json:"role,omitempty"`
	Content          string          `json:"content,omitempty"`
	ReasoningContent string          `json:"reasoning_content,omitempty"`
	ToolCalls        []ToolCallDelta `json:"tool_calls,omitempty"`
}

type ToolCallDelta struct {
	Index    int                   `json:"index"`
	ID       string                `json:"id,omitempty"`
	Function ToolCallDeltaFunction `json:"function,omitempty"`
}

type ToolCallDeltaFunction struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type PendingOllamaToolCall struct {
	Name      string
	Arguments string
}

func TransformToOllama(reader io.Reader, model string, writer io.Writer) error {
	decoder := json.NewDecoder(reader)
	encoder := json.NewEncoder(writer)

	pendingToolCalls := make(map[int]*PendingOllamaToolCall)

	for {
		var line string
		if err := decoder.Decode(&line); err != nil {
			if err == io.EOF {
				break
			}
			continue
		}

		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}

		data := strings.TrimPrefix(line, "data:")
		data = strings.TrimSpace(data)

		if data == "[DONE]" {
			endResp := OllamaResponse{
				Model:   model,
				Message: OllamaMessage{Role: "assistant", Content: ""},
				Done:    true,
			}
			encoder.Encode(endResp)
			return nil
		}

		var chunk OpenAIChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		choice := chunk.Choices[0]
		delta := choice.Delta

		for _, tc := range delta.ToolCalls {
			idx := tc.Index
			if pendingToolCalls[idx] == nil {
				pendingToolCalls[idx] = &PendingOllamaToolCall{
					Name:      "",
					Arguments: "",
				}
			}
			pendingToolCalls[idx].Name += tc.Function.Name
			pendingToolCalls[idx].Arguments += tc.Function.Arguments
		}

		if delta.Content != "" {
			resp := OllamaResponse{
				Model:   model,
				Message: OllamaMessage{Role: "assistant", Content: delta.Content},
				Done:    false,
			}
			encoder.Encode(resp)
		}

		finishReason := choice.FinishReason
		if finishReason == "tool_calls" || finishReason == "stop" {
			if len(pendingToolCalls) > 0 {
				var toolCallsArr []OllamaToolCall
				for _, tc := range pendingToolCalls {
					var args map[string]interface{}
					json.Unmarshal([]byte(tc.Arguments), &args)
					if args == nil {
						args = make(map[string]interface{})
					}
					toolCallsArr = append(toolCallsArr, OllamaToolCall{
						Function: OllamaFunctionCall{
							Name:      tc.Name,
							Arguments: args,
						},
					})
				}

				resp := OllamaResponse{
					Model: model,
					Message: OllamaMessage{
						Role:      "assistant",
						Content:   "",
						ToolCalls: toolCallsArr,
					},
					Done: true,
				}
				encoder.Encode(resp)
				pendingToolCalls = make(map[int]*PendingOllamaToolCall)
			} else if finishReason == "stop" {
				endResp := OllamaResponse{
					Model:   model,
					Message: OllamaMessage{Role: "assistant", Content: ""},
					Done:    true,
				}
				encoder.Encode(endResp)
			}
		}
	}

	return nil
}

func FormatOllamaJSON(model string, content string, done bool) string {
	resp := OllamaResponse{
		Model:   model,
		Message: OllamaMessage{Role: "assistant", Content: content},
		Done:    done,
	}
	data, _ := json.Marshal(resp)
	return string(data)
}

func ParseOllamaRequest(body []byte) (map[string]interface{}, error) {
	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		return nil, fmt.Errorf("failed to parse Ollama request: %w", err)
	}
	return req, nil
}
