// Package helpers provides utility functions for translator operations.
// ref: open-sse/translator/helpers/maxTokensHelper.js
package helpers

const (
	DefaultMaxTokens = 4096
	DefaultMinTokens = 1024
)

func AdjustMaxTokens(body map[string]interface{}) int {
	if body == nil {
		return DefaultMaxTokens
	}

	maxTokens := DefaultMaxTokens
	if mt, ok := body["max_tokens"].(float64); ok {
		maxTokens = int(mt)
	} else if mt, ok := body["max_tokens"].(int); ok {
		maxTokens = mt
	}

	if tools, ok := body["tools"].([]interface{}); ok && len(tools) > 0 {
		if maxTokens < DefaultMinTokens {
			maxTokens = DefaultMinTokens
		}
	}

	if thinking, ok := body["thinking"].(map[string]interface{}); ok {
		if budgetTokens, ok := thinking["budget_tokens"].(float64); ok {
			if maxTokens <= int(budgetTokens) {
				maxTokens = int(budgetTokens) + 1024
			}
		} else if budgetTokens, ok := thinking["budget_tokens"].(int); ok {
			if maxTokens <= budgetTokens {
				maxTokens = budgetTokens + 1024
			}
		}
	}

	return maxTokens
}
