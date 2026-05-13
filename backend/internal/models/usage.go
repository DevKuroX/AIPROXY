package models

import "time"

type UsageLogEntry struct {
	ID               string     `json:"id" db:"id"`
	Timestamp        time.Time  `json:"timestamp" db:"timestamp"`
	Model            string     `json:"model" db:"model"`
	ProviderID       string     `json:"provider_id" db:"provider_id"`
	AccountID        *string    `json:"account_id,omitempty" db:"account_id"`
	PromptTokens     int        `json:"prompt_tokens" db:"prompt_tokens"`
	CompletionTokens int        `json:"completion_tokens" db:"completion_tokens"`
	TotalTokens      int        `json:"total_tokens" db:"total_tokens"`
	CostUSD          float64    `json:"cost_usd" db:"cost_usd"`
	RTKBytesSaved    int        `json:"rtk_bytes_saved" db:"rtk_bytes_saved"`
	CavemanActive    bool       `json:"caveman_active" db:"caveman_active"`
	APIKeyID         *string    `json:"api_key_id,omitempty" db:"api_key_id"`
	DurationMs       int        `json:"duration_ms" db:"duration_ms"`
	Status           string     `json:"status" db:"status"`
	ErrorMessage     *string    `json:"error_message,omitempty" db:"error_message"`
}

type ModelPricing struct {
	ModelPattern         string    `json:"model_pattern" db:"model_pattern"`
	PromptPricePer1M     float64   `json:"prompt_price_per_1m" db:"prompt_price_per_1m"`
	CompletionPricePer1M float64   `json:"completion_price_per_1m" db:"completion_price_per_1m"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`
}

type UsageLogFilter struct {
	StartTime  *time.Time
	EndTime    *time.Time
	Model      string
	ProviderID string
	APIKeyID   string
	Status     string
	Limit      int
	Offset     int
}

type UsageLogStats struct {
	TotalRequests      int
	SuccessfulRequests int
	FailedRequests     int
	TotalPromptTokens  int64
	TotalCompletionTok int64
	TotalTokens        int64
	TotalCostUSD       float64
	TotalRTKBytesSaved int64
	AvgDurationMs      float64
}
