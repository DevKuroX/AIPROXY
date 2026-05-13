// ref: open-sse/services/accountFallback.js
package router

import (
	"context"
	"strings"
	"sync"
	"time"
)

type AccountState struct {
	AccountID        string
	UnavailableUntil time.Time
	BackoffLevel     int
	LastError        string
	LastErrorTime    time.Time
}

type AccountHealthState struct {
	mu              sync.RWMutex
	accounts        map[string]*AccountState
	cleanupInterval time.Duration
	entryTTL        time.Duration
}

func NewAccountHealthState() *AccountHealthState {
	return &AccountHealthState{
		accounts:        make(map[string]*AccountState),
		cleanupInterval: 5 * time.Minute,
		entryTTL:        30 * time.Minute,
	}
}

func (t *AccountHealthState) getOrCreateState(accountID string) *AccountState {
	state, exists := t.accounts[accountID]
	if !exists {
		state = &AccountState{
			AccountID:    accountID,
			BackoffLevel: 0,
		}
		t.accounts[accountID] = state
	}
	return state
}

func (t *AccountHealthState) IsAvailable(accountID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state, exists := t.accounts[accountID]
	if !exists {
		return true
	}

	return state.UnavailableUntil.IsZero() || state.UnavailableUntil.Before(time.Now())
}

func (t *AccountHealthState) MarkRateLimited(accountID string, duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.getOrCreateState(accountID)
	state.UnavailableUntil = time.Now().Add(duration)
}

func (t *AccountHealthState) MarkUnavailable(accountID string, duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.getOrCreateState(accountID)
	state.UnavailableUntil = time.Now().Add(duration)
}

func (t *AccountHealthState) MarkAvailable(accountID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.getOrCreateState(accountID)
	state.UnavailableUntil = time.Time{}
}

func (t *AccountHealthState) GetBackoffLevel(accountID string) int {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state, exists := t.accounts[accountID]
	if !exists {
		return 0
	}
	return state.BackoffLevel
}

func (t *AccountHealthState) IncrementBackoff(accountID string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.getOrCreateState(accountID)
	state.BackoffLevel++
	if state.BackoffLevel > BackoffMaxLevel {
		state.BackoffLevel = BackoffMaxLevel
	}
	return state.BackoffLevel
}

func (t *AccountHealthState) ResetBackoff(accountID string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.getOrCreateState(accountID)
	state.BackoffLevel = 0
}

func (t *AccountHealthState) SetLastError(accountID string, errMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	state := t.getOrCreateState(accountID)
	state.LastError = errMsg
	state.LastErrorTime = time.Now()
}

func (t *AccountHealthState) GetLastError(accountID string) (string, time.Time) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state, exists := t.accounts[accountID]
	if !exists {
		return "", time.Time{}
	}
	return state.LastError, state.LastErrorTime
}

func (t *AccountHealthState) GetUnavailableUntil(accountID string) time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state, exists := t.accounts[accountID]
	if !exists {
		return time.Time{}
	}
	return state.UnavailableUntil
}

func (t *AccountHealthState) StartCleanupTask(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(t.cleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				t.cleanup()
			}
		}
	}()
}

func (t *AccountHealthState) cleanup() {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	for accountID, state := range t.accounts {
		cooldownExpired := state.UnavailableUntil.IsZero() || state.UnavailableUntil.Before(now)
		lastErrorStale := state.LastErrorTime.IsZero() || state.LastErrorTime.Add(t.entryTTL).Before(now)

		if cooldownExpired && lastErrorStale {
			delete(t.accounts, accountID)
		}
	}
}

func (t *AccountHealthState) GetAllUnavailable() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var unavailable []string
	now := time.Now()
	for accountID, state := range t.accounts {
		if !state.UnavailableUntil.IsZero() && state.UnavailableUntil.After(now) {
			unavailable = append(unavailable, accountID)
		}
	}
	return unavailable
}

func ContainsBackoffText(errorText string) bool {
	if errorText == "" {
		return false
	}
	lowerText := strings.ToLower(errorText)
	backoffPatterns := []string{
		"rate limit",
		"too many requests",
		"quota exceeded",
		"capacity",
		"overloaded",
	}
	for _, pattern := range backoffPatterns {
		if strings.Contains(lowerText, pattern) {
			return true
		}
	}
	return false
}
