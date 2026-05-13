package models

import "time"

// ProviderNode represents a custom OpenAI-compatible endpoint.
type ProviderNode struct {
	ID                string    `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	BaseURL           string    `json:"base_url" db:"base_url"`
	APIKey            string    `json:"-" db:"api_key"` // never expose in JSON
	CompatibleFormat  string    `json:"compatible_format" db:"compatible_format"` // "openai", "anthropic", etc.
	Enabled           bool      `json:"enabled" db:"enabled"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// ModelAlias represents a model alias mapping.
type ModelAlias struct {
	ID          string    `json:"id" db:"id"`
	NodeID      string    `json:"node_id" db:"node_id"`
	Alias       string    `json:"alias" db:"alias"`        // user-facing model name
	TargetModel string    `json:"target_model" db:"target_model"` // actual model at provider
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
