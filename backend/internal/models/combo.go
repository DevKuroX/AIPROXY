package models

import (
	"context"
	"time"
)

// Combo represents a model combo with fallback/round-robin strategies.
// ref: open-sse/services/combo.js:8-12
type Combo struct {
	Name        string    `json:"name" db:"name"`
	Models      []string  `json:"models" db:"models"`       // ordered list of model identifiers
	Strategy    string    `json:"strategy" db:"strategy"`   // "fallback" or "round-robin"
	StickyLimit int       `json:"sticky_limit" db:"sticky_limit"` // requests per model before switching (round-robin)
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// ComboStore defines the interface for combo CRUD operations.
type ComboStore interface {
	CreateCombo(ctx context.Context, combo *Combo) error
	GetComboByName(ctx context.Context, name string) (*Combo, error)
	ListCombos(ctx context.Context) ([]Combo, error)
	UpdateCombo(ctx context.Context, combo *Combo) error
	DeleteCombo(ctx context.Context, name string) error
}
