package router

import (
	"encoding/json"
	"net/http"
)

// HandleDCP applies Context Deduplication/Pruning to request messages.
//  1. Removes duplicate tool calls (keeps last occurrence by tool_call_id)
//  2. Removes tool results marked as errors (is_error=true)
//  3. Compresses remaining tool output via RTK
func HandleDCP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Messages []json.RawMessage `json:"messages"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
		return
	}
	r.Body.Close()

	if len(req.Messages) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{"messages": []interface{}{}})
		return
	}

	var messages []map[string]interface{}
	for _, raw := range req.Messages {
		var msg map[string]interface{}
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}
		messages = append(messages, msg)
	}

	// Step 1: Remove duplicate tool calls (by tool_call_id)
	messages = dedupToolCalls(messages)
	beforePrune := len(messages)

	// Step 2: Remove errored tool results
	messages = pruneErrorToolResults(messages)
	prunedCount := beforePrune - len(messages)

	// Step 3: Compress remaining tool output via RTK
	if isRTKEnabled() {
		compressToolMessages(messages)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"messages":     messages,
		"pruned":       prunedCount,
	})
}

// dedupToolCalls removes duplicate tool_call entries within assistant messages.
// Keeps the LAST occurrence of each tool_call_id (assistant may retry on error).
func dedupToolCalls(messages []map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}

	for _, msg := range messages {
		role, _ := msg["role"].(string)
		if role != "assistant" {
			result = append(result, msg)
			continue
		}

		toolCalls, ok := msg["tool_calls"].([]interface{})
		if !ok || len(toolCalls) == 0 {
			result = append(result, msg)
			continue
		}

		// Track tool_call_ids in this message
		var kept []interface{}
		localSeen := make(map[string]bool)

		for _, tc := range toolCalls {
			tcMap, ok := tc.(map[string]interface{})
			if !ok {
				kept = append(kept, tc)
				continue
			}

			// Get the id from nested function structure or direct
			tcID := getToolCallID(tcMap)
			if tcID == "" {
				kept = append(kept, tc)
				continue
			}

			// If already seen in this message, it's a duplicate
			if localSeen[tcID] {
				continue
			}
			localSeen[tcID] = true
			kept = append(kept, tc)
		}

		msg["tool_calls"] = kept
		result = append(result, msg)
	}

	return result
}

// pruneErrorToolResults removes tool_result blocks where is_error is true.
func pruneErrorToolResults(messages []map[string]interface{}) []map[string]interface{} {
	var result []map[string]interface{}

	for _, msg := range messages {
		role, _ := msg["role"].(string)

		if role == "tool" {
			if isError, ok := msg["is_error"].(bool); ok && isError {
				continue
			}
			result = append(result, msg)
			continue
		}

		if role == "user" || role == "assistant" {
			content, ok := msg["content"].([]interface{})
			if !ok {
				result = append(result, msg)
				continue
			}

			var kept []interface{}
			for _, block := range content {
				b, ok := block.(map[string]interface{})
				if !ok {
					kept = append(kept, block)
					continue
				}
				blockType, _ := b["type"].(string)
				if blockType == "tool_result" {
					if isError, ok := b["is_error"].(bool); ok && isError {
						continue
					}
				}
				kept = append(kept, block)
			}
			msg["content"] = kept
		}

		result = append(result, msg)
	}

	return result
}

// compressToolMessages applies RTK compression to tool message content.
func compressToolMessages(messages []map[string]interface{}) {
	rtkEnabled := isRTKEnabled()
	if !rtkEnabled {
		return
	}

	for _, msg := range messages {
		role, _ := msg["role"].(string)

		// OpenAI tool message
		if role == "tool" {
			if content, ok := msg["content"].(string); ok && len(content) > 500 {
				// Apply basic compression: truncate to 2000 chars max
				if len(content) > 2000 {
					msg["content"] = content[:2000] + "\n...[RTK truncated]"
				}
			}
		}

		// Claude tool_result blocks
		if content, ok := msg["content"].([]interface{}); ok {
			for _, block := range content {
				b, ok := block.(map[string]interface{})
				if !ok {
					continue
				}
				blockType, _ := b["type"].(string)
				if blockType == "tool_result" {
					if text, ok := b["content"].(string); ok && len(text) > 500 {
						if len(text) > 2000 {
							b["content"] = text[:2000] + "\n...[RTK truncated]"
						}
					}
				}
			}
		}
	}
}

// getToolCallID extracts tool_call_id from various formats.
func getToolCallID(tc map[string]interface{}) string {
	// Direct id field
	if id, ok := tc["id"].(string); ok && id != "" {
		return id
	}
	// Nested function.id
	if fn, ok := tc["function"].(map[string]interface{}); ok {
		if id, ok := fn["id"].(string); ok && id != "" {
			return id
		}
	}
	return ""
}
