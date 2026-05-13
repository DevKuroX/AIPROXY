// ref: _ref/9router/open-sse/config/models.js
package config

type ModelInfo struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Type          []string `json:"type"`
	ContextWindow int      `json:"context_window"`
	Provider      string   `json:"provider"`
	InputPrice    float64  `json:"input_price"`
	OutputPrice   float64  `json:"output_price"`
}

var DefaultModelInfo = ModelInfo{
	Type:          []string{"chat"},
	ContextWindow: 200000,
}

var Models = map[string]ModelInfo{
	"claude-sonnet-4-20250514": {
		ID:            "claude-sonnet-4-20250514",
		Name:          "Claude Sonnet 4",
		Type:          []string{"chat"},
		ContextWindow: 200000,
		Provider:      "claude",
		InputPrice:    3.0,
		OutputPrice:   15.0,
	},
	"claude-opus-4-20250514": {
		ID:            "claude-opus-4-20250514",
		Name:          "Claude Opus 4",
		Type:          []string{"chat"},
		ContextWindow: 200000,
		Provider:      "claude",
		InputPrice:    15.0,
		OutputPrice:   75.0,
	},
	"claude-3-5-sonnet-20241022": {
		ID:            "claude-3-5-sonnet-20241022",
		Name:          "Claude 3.5 Sonnet",
		Type:          []string{"chat"},
		ContextWindow: 200000,
		Provider:      "claude",
		InputPrice:    3.0,
		OutputPrice:   15.0,
	},
	"claude-3-5-haiku-20241022": {
		ID:            "claude-3-5-haiku-20241022",
		Name:          "Claude 3.5 Haiku",
		Type:          []string{"chat"},
		ContextWindow: 200000,
		Provider:      "claude",
		InputPrice:    0.8,
		OutputPrice:   4.0,
	},
	"gpt-4o": {
		ID:            "gpt-4o",
		Name:          "GPT-4o",
		Type:          []string{"chat"},
		ContextWindow: 128000,
		Provider:      "openai",
		InputPrice:    2.5,
		OutputPrice:   10.0,
	},
	"gpt-4o-mini": {
		ID:            "gpt-4o-mini",
		Name:          "GPT-4o Mini",
		Type:          []string{"chat"},
		ContextWindow: 128000,
		Provider:      "openai",
		InputPrice:    0.15,
		OutputPrice:   0.6,
	},
	"gpt-4-turbo": {
		ID:            "gpt-4-turbo",
		Name:          "GPT-4 Turbo",
		Type:          []string{"chat"},
		ContextWindow: 128000,
		Provider:      "openai",
		InputPrice:    10.0,
		OutputPrice:   30.0,
	},
	"o1": {
		ID:            "o1",
		Name:          "O1",
		Type:          []string{"chat"},
		ContextWindow: 200000,
		Provider:      "openai",
		InputPrice:    15.0,
		OutputPrice:   60.0,
	},
	"o1-mini": {
		ID:            "o1-mini",
		Name:          "O1 Mini",
		Type:          []string{"chat"},
		ContextWindow: 128000,
		Provider:      "openai",
		InputPrice:    1.5,
		OutputPrice:   6.0,
	},
	"o1-preview": {
		ID:            "o1-preview",
		Name:          "O1 Preview",
		Type:          []string{"chat"},
		ContextWindow: 128000,
		Provider:      "openai",
		InputPrice:    15.0,
		OutputPrice:   60.0,
	},
	"gemini-2.5-pro": {
		ID:            "gemini-2.5-pro",
		Name:          "Gemini 2.5 Pro",
		Type:          []string{"chat"},
		ContextWindow: 1000000,
		Provider:      "gemini",
		InputPrice:    1.25,
		OutputPrice:   10.0,
	},
	"gemini-2.5-flash": {
		ID:            "gemini-2.5-flash",
		Name:          "Gemini 2.5 Flash",
		Type:          []string{"chat"},
		ContextWindow: 1000000,
		Provider:      "gemini",
		InputPrice:    0.075,
		OutputPrice:   0.3,
	},
	"gemini-2.0-flash": {
		ID:            "gemini-2.0-flash",
		Name:          "Gemini 2.0 Flash",
		Type:          []string{"chat"},
		ContextWindow: 1000000,
		Provider:      "gemini",
		InputPrice:    0.1,
		OutputPrice:   0.4,
	},
	"deepseek-chat": {
		ID:            "deepseek-chat",
		Name:          "DeepSeek Chat",
		Type:          []string{"chat"},
		ContextWindow: 64000,
		Provider:      "deepseek",
		InputPrice:    0.14,
		OutputPrice:   0.28,
	},
	"deepseek-reasoner": {
		ID:            "deepseek-reasoner",
		Name:          "DeepSeek Reasoner",
		Type:          []string{"chat"},
		ContextWindow: 64000,
		Provider:      "deepseek",
		InputPrice:    0.55,
		OutputPrice:   2.19,
	},
	"codex-mini": {
		ID:            "codex-mini",
		Name:          "Codex Mini",
		Type:          []string{"chat"},
		ContextWindow: 128000,
		Provider:      "codex",
	},
}

func GetModelInfo(modelID string) ModelInfo {
	if m, ok := Models[modelID]; ok {
		return m
	}
	return DefaultModelInfo
}

func GetModelProvider(modelID string) string {
	if m, ok := Models[modelID]; ok {
		return m.Provider
	}
	return ""
}

func GetModelContextWindow(modelID string) int {
	if m, ok := Models[modelID]; ok {
		return m.ContextWindow
	}
	return DefaultModelInfo.ContextWindow
}

func GetModelsByProvider(provider string) []string {
	var result []string
	for id, m := range Models {
		if m.Provider == provider {
			result = append(result, id)
		}
	}
	return result
}

func GetAllModels() []string {
	var result []string
	for id := range Models {
		result = append(result, id)
	}
	return result
}

func ModelExists(modelID string) bool {
	_, ok := Models[modelID]
	return ok
}
