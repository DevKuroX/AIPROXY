package router

import (
	"errors"
	"sync"

	"github.com/DevKuroX/AIPROXY/internal/models"
)

var ErrComboNotFound = errors.New("combo not found")

// rotationState tracks round-robin state per combo
// ref: open-sse/services/combo.js:12
type rotationState struct {
	index               int
	consecutiveUseCount int
}

// ComboResolver handles model combo resolution with rotation
type ComboResolver struct {
	mu         sync.RWMutex
	rotation   map[string]*rotationState
	comboStore ComboStore
}

// ComboStore interface for combo data access
type ComboStore interface {
	GetComboByName(name string) (*models.Combo, error)
	ListCombos() ([]models.Combo, error)
}

// NewComboResolver creates a new combo resolver
func NewComboResolver(store ComboStore) *ComboResolver {
	return &ComboResolver{
		rotation:   make(map[string]*rotationState),
		comboStore: store,
	}
}

// normalizeStickyLimit ensures sticky limit is valid
// ref: open-sse/services/combo.js:14-17
func normalizeStickyLimit(stickyLimit int) int {
	if stickyLimit > 0 {
		return stickyLimit
	}
	return 1
}

// rotateModelsFromIndex rotates model list starting from given index
// ref: open-sse/services/combo.js:19-26
func rotateModelsFromIndex(models []string, currentIndex int) []string {
	if currentIndex <= 0 || currentIndex >= len(models) {
		return models
	}
	rotated := make([]string, len(models))
	copy(rotated, models[currentIndex:])
	copy(rotated[len(models)-currentIndex:], models[:currentIndex])
	return rotated
}

// GetRotatedModels returns rotated model list based on strategy
// ref: open-sse/services/combo.js:36-65
func (cr *ComboResolver) GetRotatedModels(models []string, comboName string, strategy string, stickyLimit int) []string {
	if len(models) <= 1 || strategy != "round-robin" {
		return models
	}

	rotationKey := comboName
	if rotationKey == "" {
		rotationKey = "__default__"
	}

	normalizedStickyLimit := normalizeStickyLimit(stickyLimit)

	cr.mu.Lock()
	defer cr.mu.Unlock()

	state, exists := cr.rotation[rotationKey]
	if !exists {
		state = &rotationState{index: 0, consecutiveUseCount: 0}
		cr.rotation[rotationKey] = state
	}

	currentIndex := state.index % len(models)
	rotatedModels := rotateModelsFromIndex(models, currentIndex)
	nextUseCount := state.consecutiveUseCount + 1

	if nextUseCount >= normalizedStickyLimit {
		cr.rotation[rotationKey] = &rotationState{
			index:               (currentIndex + 1) % len(models),
			consecutiveUseCount: 0,
		}
	} else {
		cr.rotation[rotationKey] = &rotationState{
			index:               currentIndex,
			consecutiveUseCount: nextUseCount,
		}
	}

	return rotatedModels
}

// ResolveCombo resolves a combo name to a list of models
// ref: open-sse/services/combo.js:82-94
func (cr *ComboResolver) ResolveCombo(comboName string) ([]string, error) {
	combo, err := cr.comboStore.GetComboByName(comboName)
	if err != nil {
		return nil, ErrComboNotFound
	}

	if len(combo.Models) == 0 {
		return nil, errors.New("combo has no models")
	}

	return cr.GetRotatedModels(combo.Models, comboName, combo.Strategy, combo.StickyLimit), nil
}

// ResetComboRotation resets rotation state for a specific combo or all combos
// ref: open-sse/services/combo.js:71-74
func (cr *ComboResolver) ResetComboRotation(comboName string) {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if comboName != "" {
		delete(cr.rotation, comboName)
	} else {
		cr.rotation = make(map[string]*rotationState)
	}
}

// GetCombo retrieves combo by name
func (cr *ComboResolver) GetCombo(comboName string) (*models.Combo, error) {
	return cr.comboStore.GetComboByName(comboName)
}

// ListCombos retrieves all combos
func (cr *ComboResolver) ListCombos() ([]models.Combo, error) {
	return cr.comboStore.ListCombos()
}
