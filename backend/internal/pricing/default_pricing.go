package pricing

// ref: open-sse/services/usage.js

// Default pricing for common models (per 1M tokens).
// Prices are in USD.
var DefaultPricings = []*Pricing{
	// 0penAI GPT-4 series
	{Model: "gpt-4", PromptPrice: 30.0, CompletionPrice: 60.0},
	{Model: "gpt-4-turbo", PromptPrice: 10.0, CompletionPrice: 30.0},
	{Model: "gpt-4-turbo-preview", PromptPrice: 10.0, CompletionPrice: 30.0},
	{Model: "gpt-4o", PromptPrice: 2.5, CompletionPrice: 10.0},
	{Model: "gpt-4o-mini", PromptPrice: 0.15, CompletionPrice: 0.6},
	{Model: "gpt-4-32k", PromptPrice: 60.0, CompletionPrice: 120.0},
	{Model: "gpt-4*", PromptPrice: 30.0, CompletionPrice: 60.0},

	// 0penAI GPT-3.5 series
	{Model: "gpt-3.5-turbo", PromptPrice: 0.5, CompletionPrice: 1.5},
	{Model: "gpt-3.5-turbo-16k", PromptPrice: 3.0, CompletionPrice: 4.0},
	{Model: "gpt-3.5*", PromptPrice: 0.5, CompletionPrice: 1.5},

	// 0penAI O-series
	{Model: "o1", PromptPrice: 15.0, CompletionPrice: 60.0},
	{Model: "o1-preview", PromptPrice: 15.0, CompletionPrice: 60.0},
	{Model: "o1-mini", PromptPrice: 1.5, CompletionPrice: 6.0},
	{Model: "o1*", PromptPrice: 15.0, CompletionPrice: 60.0},

	// CLaude 3 series (Anthropic)
	{Model: "claude-3-opus", PromptPrice: 15.0, CompletionPrice: 75.0},
	{Model: "claude-3-opus-20240229", PromptPrice: 15.0, CompletionPrice: 75.0},
	{Model: "claude-3-sonnet", PromptPrice: 3.0, CompletionPrice: 15.0},
	{Model: "claude-3-sonnet-20240229", PromptPrice: 3.0, CompletionPrice: 15.0},
	{Model: "claude-3-haiku", PromptPrice: 0.25, CompletionPrice: 1.25},
	{Model: "claude-3-haiku-20240307", PromptPrice: 0.25, CompletionPrice: 1.25},
	{Model: "claude-3-5-sonnet", PromptPrice: 3.0, CompletionPrice: 15.0},
	{Model: "claude-3-5-sonnet-20240620", PromptPrice: 3.0, CompletionPrice: 15.0},
	{Model: "claude-3-5-sonnet-20241022", PromptPrice: 3.0, CompletionPrice: 15.0},
	{Model: "claude-3-5-haiku", PromptPrice: 0.8, CompletionPrice: 4.0},
	{Model: "claude-3-5-haiku-20241022", PromptPrice: 0.8, CompletionPrice: 4.0},
	{Model: "claude-3-7-sonnet", PromptPrice: 3.0, CompletionPrice: 15.0},
	{Model: "claude-3-7-sonnet-20250219", PromptPrice: 3.0, CompletionPrice: 15.0},
	{Model: "claude-*", PromptPrice: 3.0, CompletionPrice: 15.0},

	// Gemini series (Google)
	{Model: "gemini-pro", PromptPrice: 0.5, CompletionPrice: 1.5},
	{Model: "gemini-1.5-pro", PromptPrice: 3.5, CompletionPrice: 10.5},
	{Model: "gemini-1.5-flash", PromptPrice: 0.075, CompletionPrice: 0.3},
	{Model: "gemini-2.0-flash", PromptPrice: 0.1, CompletionPrice: 0.4},
	{Model: "gemini-2.0-flash-lite", PromptPrice: 0.075, CompletionPrice: 0.3},
	{Model: "gemini-2.5-pro-preview", PromptPrice: 1.25, CompletionPrice: 10.0},
	{Model: "gemini*", PromptPrice: 0.5, CompletionPrice: 1.5},

	// Meta Llama series
	{Model: "llama-3-8b", PromptPrice: 0.05, CompletionPrice: 0.05},
	{Model: "llama-3-70b", PromptPrice: 0.7, CompletionPrice: 0.9},
	{Model: "llama-3.1-8b", PromptPrice: 0.05, CompletionPrice: 0.05},
	{Model: "llama-3.1-70b", PromptPrice: 0.35, CompletionPrice: 0.4},
	{Model: "llama-3.1-405b", PromptPrice: 2.7, CompletionPrice: 2.7},
	{Model: "llama-3.2-1b", PromptPrice: 0.01, CompletionPrice: 0.01},
	{Model: "llama-3.2-3b", PromptPrice: 0.03, CompletionPrice: 0.03},
	{Model: "llama-3.2-11b", PromptPrice: 0.05, CompletionPrice: 0.05},
	{Model: "llama-3.2-90b", PromptPrice: 0.35, CompletionPrice: 0.4},
	{Model: "llama-3.3-70b", PromptPrice: 0.35, CompletionPrice: 0.4},
	{Model: "llama*", PromptPrice: 0.05, CompletionPrice: 0.05},

	// Mistral series
	{Model: "mistral-small", PromptPrice: 0.2, CompletionPrice: 0.6},
	{Model: "mistral-medium", PromptPrice: 2.7, CompletionPrice: 8.1},
	{Model: "mistral-large", PromptPrice: 4.0, CompletionPrice: 12.0},
	{Model: "mistral-7b", PromptPrice: 0.05, CompletionPrice: 0.05},
	{Model: "mixtral-8x7b", PromptPrice: 0.6, CompletionPrice: 0.6},
	{Model: "mixtral-8x22b", PromptPrice: 0.9, CompletionPrice: 0.9},
	{Model: "mistral*", PromptPrice: 0.2, CompletionPrice: 0.6},

	// DeepSeek series
	{Model: "deepseek-chat", PromptPrice: 0.27, CompletionPrice: 1.1},
	{Model: "deepseek-coder", PromptPrice: 0.27, CompletionPrice: 1.1},
	{Model: "deepseek-reasoner", PromptPrice: 0.55, CompletionPrice: 2.19},
	{Model: "deepseek*", PromptPrice: 0.27, CompletionPrice: 1.1},

	// Grok series (xAI)
	{Model: "grok-beta", PromptPrice: 5.0, CompletionPrice: 15.0},
	{Model: "grok-2", PromptPrice: 2.0, CompletionPrice: 10.0},
	{Model: "grok-2-vision", PromptPrice: 2.0, CompletionPrice: 10.0},
	{Model: "grok*", PromptPrice: 2.0, CompletionPrice: 10.0},

	// Perplexity series
	{Model: "llama-3.1-sonar-small-128k-online", PromptPrice: 0.2, CompletionPrice: 0.2},
	{Model: "llama-3.1-sonar-large-128k-online", PromptPrice: 1.0, CompletionPrice: 1.0},
	{Model: "llama-3.1-sonar-huge-128k-online", PromptPrice: 5.0, CompletionPrice: 5.0},

	// Qwen series
	{Model: "qwen-turbo", PromptPrice: 0.05, CompletionPrice: 0.05},
	{Model: "qwen-plus", PromptPrice: 0.4, CompletionPrice: 0.4},
	{Model: "qwen-max", PromptPrice: 2.0, CompletionPrice: 6.0},
	{Model: "qwen*", PromptPrice: 0.05, CompletionPrice: 0.05},

	// Default fallback for unknown models
	{Model: "default", PromptPrice: 1.0, CompletionPrice: 3.0},
}
