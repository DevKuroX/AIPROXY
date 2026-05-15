package router

import "strings"

// ParseModel parses a model string in "provider/model" format.
// Returns (provider, model) tuple.
// Examples:
//   - "kiro/claude-haiku-4.5" → ("kiro", "claude-haiku-4.5")
//   - "claude-haiku-4.5" → ("", "claude-haiku-4.5")
//   - "" → ("", "")
func ParseModel(modelStr string) (provider, model string) {
	modelStr = strings.TrimSpace(modelStr)
	if modelStr == "" {
		return "", ""
	}

	if strings.Contains(modelStr, "/") {
		parts := strings.SplitN(modelStr, "/", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}

	return "", modelStr
}
