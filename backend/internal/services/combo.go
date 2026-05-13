package services

import (
	"strings"
	"sync"

	"github.com/DevKuroX/AIPROXY/internal/models"
)

// rotationState tracks rotation state per combo (for round-robin strategy).
// ref: open-sse/services/combo.js:8-12
type rotationState struct {
	index              int
	consecutiveUseCount int
}

// ComboService handles combo (model combo) operations with fallback support.
// ref: open-sse/services/combo.js
type ComboService struct {
	// comboRotationState tracks rotation state per combo (for round-robin strategy)
	// ref: open-sse/services/combo.js:12
	rotationMutex sync.RWMutex
	rotationMap   map[string]*rotationState
}

// NewComboService creates a new ComboService.
func NewComboService() *ComboService {
	return &ComboService{
		rotationMap: make(map[string]*rotationState),
	}
}

// normalizeStickyLimit parses and validates sticky limit.
// ref: open-sse/services/combo.js:14-17
func normalizeStickyLimit(stickyLimit int) int {
	if stickyLimit > 0 {
		return stickyLimit
	}
	return 1
}

// rotateModelsFromIndex rotates the models array from the given index.
// ref: open-sse/services/combo.js:19-26
func rotateModelsFromIndex(models []string, currentIndex int) []string {
	if currentIndex <= 0 || currentIndex >= len(models) {
		return models
	}
	
	rotatedModels := make([]string, len(models))
	copy(rotatedModels, models[currentIndex:])
	copy(rotatedModels[len(models)-currentIndex:], models[:currentIndex])
	return rotatedModels
}

// GetRotatedModels returns rotated model list based on strategy.
// ref: open-sse/services/combo.js:36-65
func (s *ComboService) GetRotatedModels(models []string, comboName string, strategy string, stickyLimit int) []string {
	// Early return: no rotation needed
	// ref: open-sse/services/combo.js:37-39
	if len(models) <= 1 || strategy != "round-robin" {
		return models
	}

	rotationKey := comboName
	if rotationKey == "" {
		rotationKey = "__default__"
	}
	normalizedStickyLimit := normalizeStickyLimit(stickyLimit)

	// Get or create rotation state
	// ref: open-sse/services/combo.js:43-46
	s.rotationMutex.Lock()
	defer s.rotationMutex.Unlock()

	state, exists := s.rotationMap[rotationKey]
	if !exists {
		state = &rotationState{index: 0, consecutiveUseCount: 0}
		s.rotationMap[rotationKey] = state
	}

	currentIndex := state.index % len(models)
	rotatedModels := rotateModelsFromIndex(models, currentIndex)
	nextUseCount := state.consecutiveUseCount + 1

	// Update rotation state based on sticky limit
	// ref: open-sse/services/combo.js:52-62
	if nextUseCount >= normalizedStickyLimit {
		s.rotationMap[rotationKey] = &rotationState{
			index:              (currentIndex + 1) % len(models),
			consecutiveUseCount: 0,
		}
	} else {
		s.rotationMap[rotationKey] = &rotationState{
			index:              currentIndex,
			consecutiveUseCount: nextUseCount,
		}
	}

	return rotatedModels
}

// ResetComboRotation resets in-memory rotation state when combo/settings change.
// ref: open-sse/services/combo.js:71-74
func (s *ComboService) ResetComboRotation(comboName string) {
	s.rotationMutex.Lock()
	defer s.rotationMutex.Unlock()

	if comboName != "" {
		delete(s.rotationMap, comboName)
	} else {
		s.rotationMap = make(map[string]*rotationState)
	}
}

// GetComboModelsFromData gets combo models from combos data.
// ref: open-sse/services/combo.js:82-94
func GetComboModelsFromData(modelStr string, combosData []models.Combo) []string {
	// Don't check if it's in provider/model format
	// ref: open-sse/services/combo.js:84
	if strings.Contains(modelStr, "/") {
		return nil
	}

	// Find the combo by name
	// ref: open-sse/services/combo.js:89-93
	for i := range combosData {
		if combosData[i].Name == modelStr && len(combosData[i].Models) > 0 {
			return combosData[i].Models
		}
	}
	return nil
}
