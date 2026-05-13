package models

import "time"

// Provider represents an LLM provider configuration.
type Provider struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"` // "openai", "claude", "gemini", etc.
	BaseURL   string    `json:"base_url" db:"base_url"`
	APIKey    string    `json:"-" db:"api_key"` // never expose in JSON
	Enabled   bool      `json:"enabled" db:"enabled"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ProviderAccount represents an API key account for a provider.
type ProviderAccount struct {
	ID         string    `json:"id" db:"id"`
	ProviderID string    `json:"provider_id" db:"provider_id"`
	Name       string    `json:"name" db:"name"`
	APIKey     string    `json:"-" db:"api_key"` // never expose in JSON
	IsActive   bool      `json:"is_active" db:"is_active"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
