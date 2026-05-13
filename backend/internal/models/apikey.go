package models

import "time"

// APIKey represents a client-facing API key for authenticating requests to the proxy.
type APIKey struct {
	ID         string     `json:"id" db:"id"`
	Key        string     `json:"key" db:"key"`
	KeyHash    string     `json:"-" db:"key_hash"`
	Name       string     `json:"name" db:"name"`
	IsActive   bool       `json:"is_active" db:"is_active"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
}
