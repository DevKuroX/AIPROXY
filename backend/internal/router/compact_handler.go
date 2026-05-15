package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/DevKuroX/AIPROXY/internal/pool"
	"github.com/DevKuroX/AIPROXY/internal/providers"
)

func HandleCompact(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "failed to read body"})
		return
	}
	r.Body.Close()

	var req map[string]interface{}
	if err := json.Unmarshal(body, &req); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
		return
	}

	modelStr, _ := req["model"].(string)
	if modelStr == "" {
		json.NewEncoder(w).Encode(map[string]string{"error": "model required"})
		return
	}

	input, _ := req["input"].([]interface{})
	messages, _ := req["messages"].([]interface{})
	history := input
	if history == nil {
		history = messages
	}

	providerID, modelName := ParseModel(modelStr)
	providerCfg, exists := providers.GetProviderConfig(providerID)
	if !exists {
		json.NewEncoder(w).Encode(map[string]string{"error": "unknown provider"})
		return
	}

	// If Codex, use native compact endpoint
	if providerID == "codex" {
		handleNativeCompact(w, r, &providerCfg, modelName, body)
		return
	}

	// For all other providers: self-summary via LLM call
	compactResult, err := compactViaLLM(providerCfg, modelName, history)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	// If streaming is requested
	if s, ok := req["stream"].(bool); ok && s {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		json.NewEncoder(w).Encode(map[string]interface{}{
			"type": "conversation.compact",
			"compacted": compactResult,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"type": "conversation.compact",
		"compacted": compactResult,
	})
}

func compactViaLLM(cfg providers.ProviderConfig, model string, history []interface{}) (map[string]interface{}, error) {
	if len(history) == 0 {
		return map[string]interface{}{"input": []interface{}{}}, nil
	}

	// Build the conversation text to summarize
	var sb strings.Builder
	sb.WriteString("=== PERCAKAPAN ===\n")
	for _, msg := range history {
		if m, ok := msg.(map[string]interface{}); ok {
			role, _ := m["role"].(string)
			content, _ := m["content"].(string)
			if content == "" {
				if parts, ok := m["content"].([]interface{}); ok {
					for _, p := range parts {
						if part, ok := p.(map[string]interface{}); ok {
							if t, ok := part["text"].(string); ok {
								content += t
							}
						}
					}
				}
			}
			if content != "" {
				sb.WriteString(fmt.Sprintf("[%s]: %s\n", role, content))
			}
		}
	}
	conversationText := sb.String()

	// Build summary prompt
	summaryMessages := []map[string]interface{}{
		{
			"role": "system",
			"content": "You are a conversation summarizer. Summarize the following coding conversation. " +
				"Keep ALL of these in the summary:\n" +
				"- Requirements and goals\n" +
				"- File paths and code changes made\n" +
				"- Architectural decisions and reasons\n" +
				"- Errors encountered and fixes\n" +
				"- Current task status (what's done, what's next)\n" +
				"Output format:\n" +
				"SUMMARY:\n" +
				"[concise summary in 3-5 paragraphs]\n" +
				"KEY_FILES:\n" +
				"- path/to/file.go: what changed\n" +
				"PENDING:\n" +
				"- what still needs to be done",
		},
		{
			"role": "user",
			"content": conversationText,
		},
	}

	summaryBody, _ := json.Marshal(map[string]interface{}{
		"model":    model,
		"messages": summaryMessages,
		"max_tokens": 1000,
		"stream":   false,
	})

	// Call provider with the summary request (skip auth - reuse current provider)
	account := &pool.Account{
		ID:          "compact",
		Provider:    cfg.Type,
		IsActive:    true,
		AccessToken: "public",
	}

	resp, err := callProviderAPI(context.Background(), &cfg, model, account, summaryBody)
	if err != nil {
		return nil, fmt.Errorf("compact call failed: %w", err)
	}

	summary := ""
	if choices, ok := resp["choices"].([]interface{}); ok && len(choices) > 0 {
		if first, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := first["message"].(map[string]interface{}); ok {
				if c, ok := msg["content"].(string); ok {
					summary = c
				}
			}
		}
	}

	if summary == "" {
		return nil, fmt.Errorf("compact returned empty summary")
	}

	return map[string]interface{}{
		"input": []interface{}{
			map[string]interface{}{
				"role":    "user",
				"content": summary,
			},
		},
	}, nil
}

func handleNativeCompact(w http.ResponseWriter, r *http.Request, cfg *providers.ProviderConfig, model string, body []byte) {
	var reqMap map[string]interface{}
	json.Unmarshal(body, &reqMap)
	reqMap["_compact"] = true
	reqMap["model"] = model
	modifiedBody, _ := json.Marshal(reqMap)

	account := &pool.Account{
		ID:          "compact",
		Provider:    cfg.Type,
		IsActive:    true,
		AccessToken: "public",
	}

	resp, err := callProviderAPI(r.Context(), cfg, model, account, modifiedBody)
	if err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(resp)
}
