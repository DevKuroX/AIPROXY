package models

import "time"

// UsageLog represents a single API usage record.
type UsageLog struct {
	ID           string    `json:"id" db:"id"`
	APIKeyID     string    `json:"api_key_id" db:"api_key_id"`
	Provider     string    `json:"provider" db:"provider"`
	Model        string    `json:"model" db:"model"`
	InputTokens  int64     `json:"input_tokens" db:"input_tokens"`
	OutputTokens int64     `json:"output_tokens" db:"output_tokens"`
	TotalTokens  int64     `json:"total_tokens" db:"total_tokens"`
	Cost         float64   `json:"cost" db:"cost"`
	RequestID    string    `json:"request_id" db:"request_id"`
	Status       int       `json:"status" db:"status"`
	LatencyMs    int64     `json:"latency_ms" db:"latency_ms"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// PricingRule represents a pricing rule for a provider/model combination.
type PricingRule struct {
	ID           string    `json:"id" db:"id"`
	Provider     string    `json:"provider" db:"provider"`
	Model        string    `json:"model" db:"model"`
	InputPrice   float64   `json:"input_price" db:"input_price"`   // Price per 1M input tokens
	OutputPrice  float64   `json:"output_price" db:"output_price"` // Price per 1M output tokens
	IsActive     bool      `json:"is_active" db:"is_active"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
