package router

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
)

// ref: open-sse/config/errorConfig.js:32-36
const (
	BackoffBase           = 2 * time.Second
	BackoffMax            = 5 * time.Minute
	BackoffMaxLevel       = 15
	TransientCooldown     = 30 * time.Second
	MaxRateLimitCooldown  = 30 * time.Minute
)

// ref: open-sse/config/errorConfig.js:45-48
const (
	CooldownLong  = 2 * time.Minute
	CooldownShort = 5 * time.Second
)

// ErrAllAccountsExhausted is returned when all accounts in the fallback chain have been tried.
var ErrAllAccountsExhausted = errors.New("all accounts exhausted")

type AllAccountsExhaustedError struct {
	LastError string
}

func NewAllAccountsExhaustedError(lastError string) *AllAccountsExhaustedError {
	return &AllAccountsExhaustedError{LastError: lastError}
}

func (e *AllAccountsExhaustedError) Error() string {
	return "all accounts exhausted: " + e.LastError
}

func (e *AllAccountsExhaustedError) Unwrap() error {
	return ErrAllAccountsExhausted
}

// ErrorRule represents a single error classification rule.
// ref: open-sse/config/errorConfig.js:50-58
type ErrorRule struct {
	Text       string        // substring match (case-insensitive) on error message
	Status     int           // HTTP status code match (0 = not status-based)
	Cooldown   time.Duration // fixed cooldown duration (0 = use backoff)
	UseBackoff bool          // true = use exponential backoff (rate limit)
}

// ERROR_RULES is the unified error classification rules.
// Checked top-to-bottom: text rules first (by order), then status rules.
// ref: open-sse/config/errorConfig.js:59-76
var ERROR_RULES = []ErrorRule{
	// --- Text-based rules (checked first, order = priority) ---
	{Text: "no credentials", Cooldown: CooldownLong},
	{Text: "request not allowed", Cooldown: CooldownShort},
	{Text: "improperly formed request", Cooldown: CooldownLong},
	{Text: "rate limit", UseBackoff: true},
	{Text: "too many requests", UseBackoff: true},
	{Text: "quota exceeded", UseBackoff: true},
	{Text: "capacity", UseBackoff: true},
	{Text: "overloaded", UseBackoff: true},

	// --- Status-based rules (fallback when text doesn't match) ---
	{Status: 401, Cooldown: CooldownLong},
	{Status: 402, Cooldown: CooldownLong},
	{Status: 403, Cooldown: CooldownLong},
	{Status: 404, Cooldown: CooldownLong},
	{Status: 429, UseBackoff: true},
}

// ProviderAccount represents an account with its availability state.
// ref: open-sse/services/accountFallback.js:55-58
type ProviderAccount struct {
	Name              string
	APIKey            string
	BaseURL           string
	RateLimitedUntil  time.Time // when rate limit cooldown expires
	UnavailableUntil  time.Time // when account becomes available again
	BackoffLevel      int       // current exponential backoff level
	NeedsRefresh      bool      // true if OAuth token needs refresh (401/403)
	AttemptCount      int       // number of fallback attempts made
}

// FallbackResult contains the result of checking whether fallback should occur.
// ref: open-sse/services/accountFallback.js:21-22
type FallbackResult struct {
	ShouldFallback   bool
	Cooldown         time.Duration
	NewBackoffLevel  int
	NeedsRefresh     bool // true for 401/403 errors
}

// FallbackChain manages fallback attempts across accounts.
// ref: open-sse/services/accountFallback.js
type FallbackChain struct {
	mu             sync.RWMutex
	accounts       []*ProviderAccount
	currentIndex   int
	maxAttempts    int
}

// NewFallbackChain creates a new fallback chain from accounts.
func NewFallbackChain(accounts []*ProviderAccount) *FallbackChain {
	return &FallbackChain{
		accounts:     accounts,
		currentIndex: 0,
		maxAttempts:  len(accounts),
	}
}

// GetCurrent returns the current account in the chain.
func (fc *FallbackChain) GetCurrent() *ProviderAccount {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	if fc.currentIndex >= len(fc.accounts) {
		return nil
	}
	return fc.accounts[fc.currentIndex]
}

// getQuotaCooldown calculates exponential backoff cooldown for rate limits (429).
// Level 1: 2s, Level 2: 4s, Level 3: 8s... → max 5min.
// ref: open-sse/services/accountFallback.js:9-13
func getQuotaCooldown(backoffLevel int) time.Duration {
	level := backoffLevel - 1
	if level < 0 {
		level = 0
	}
	cooldown := BackoffBase * time.Duration(1<<uint(level))
	if cooldown > BackoffMax {
		cooldown = BackoffMax
	}
	return cooldown
}

// CheckFallbackError checks if error should trigger account fallback.
// Config-driven: matches ERROR_RULES top-to-bottom (text rules first, then status).
// ref: open-sse/services/accountFallback.js:23-50
func CheckFallbackError(status int, errorText string, backoffLevel int) FallbackResult {
	lowerError := ""
	if errorText != "" {
		lowerError = strings.ToLower(errorText)
	}

	for _, rule := range ERROR_RULES {
		// Text-based rule: match substring in error message
		if rule.Text != "" && lowerError != "" && strings.Contains(lowerError, rule.Text) {
			if rule.UseBackoff {
				newLevel := backoffLevel + 1
				if newLevel > BackoffMaxLevel {
					newLevel = BackoffMaxLevel
				}
				return FallbackResult{
					ShouldFallback:  true,
					Cooldown:        getQuotaCooldown(newLevel),
					NewBackoffLevel: newLevel,
				}
			}
			return FallbackResult{
				ShouldFallback: true,
				Cooldown:       rule.Cooldown,
			}
		}

		// Status-based rule: match HTTP status code
		if rule.Status != 0 && rule.Status == status {
			if rule.UseBackoff {
				newLevel := backoffLevel + 1
				if newLevel > BackoffMaxLevel {
					newLevel = BackoffMaxLevel
				}
				return FallbackResult{
					ShouldFallback:  true,
					Cooldown:        getQuotaCooldown(newLevel),
					NewBackoffLevel: newLevel,
				}
			}
			return FallbackResult{
				ShouldFallback: true,
				Cooldown:       rule.Cooldown,
				NeedsRefresh:   (status == 401 || status == 403), // Mark for OAuth refresh
			}
		}
	}

	// Default: transient cooldown for any unmatched error
	return FallbackResult{
		ShouldFallback: true,
		Cooldown:       TransientCooldown,
	}
}

// ShouldRetry determines if fallback should occur based on status and error text.
// ref: open-sse/services/accountFallback.js:23-50
func ShouldRetry(status int, errorText string) bool {
	result := CheckFallbackError(status, errorText, 0)
	return result.ShouldFallback
}

// IsAccountUnavailable checks if account is currently unavailable (cooldown not expired).
// ref: open-sse/services/accountFallback.js:55-58
func IsAccountUnavailable(account *ProviderAccount) bool {
	if account == nil {
		return true
	}
	now := time.Now()
	if !account.RateLimitedUntil.IsZero() && account.RateLimitedUntil.After(now) {
		return true
	}
	if !account.UnavailableUntil.IsZero() && account.UnavailableUntil.After(now) {
		return true
	}
	return false
}

// GetUnavailableUntil calculates unavailable until timestamp.
// ref: open-sse/services/accountFallback.js:63-65
func GetUnavailableUntil(cooldown time.Duration) time.Time {
	return time.Now().Add(cooldown)
}

// TryNext attempts to get the next available account in the chain.
// It marks the current account as unavailable based on the error and moves to the next.
// Returns nil when all accounts are exhausted.
// ref: open-sse/services/accountFallback.js
func (fc *FallbackChain) TryNext(ctx context.Context, currentAccount *ProviderAccount, err error, status int, errorText string) (*ProviderAccount, error) {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	// Mark current account as unavailable if provided
	if currentAccount != nil && err != nil {
		result := CheckFallbackError(status, errorText, currentAccount.BackoffLevel)
		
		// Update backoff level if applicable
		if result.NewBackoffLevel > 0 {
			currentAccount.BackoffLevel = result.NewBackoffLevel
		}
		
		// Set cooldown
		if result.Cooldown > 0 {
			currentAccount.RateLimitedUntil = GetUnavailableUntil(result.Cooldown)
		}
		
		// Mark for OAuth refresh if needed
		if result.NeedsRefresh {
			currentAccount.NeedsRefresh = true
		}
		
		currentAccount.AttemptCount++
	}

	// Find next available account
	for i := fc.currentIndex + 1; i < len(fc.accounts); i++ {
		account := fc.accounts[i]
		if !IsAccountUnavailable(account) {
			fc.currentIndex = i
			return account, nil
		}
	}

	return nil, NewAllAccountsExhaustedError(errorText)
}

// HasMoreAttempts checks if there are more accounts to try.
func (fc *FallbackChain) HasMoreAttempts() bool {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.currentIndex < len(fc.accounts)-1
}

// Reset resets the chain to start from the first account.
func (fc *FallbackChain) Reset() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.currentIndex = 0
}

// GetEarliestRateLimitedUntil returns the earliest rateLimitedUntil from a list of accounts.
// ref: open-sse/services/accountFallback.js:72-83
func GetEarliestRateLimitedUntil(accounts []*ProviderAccount) time.Time {
	var earliest time.Time
	now := time.Now()
	for _, acc := range accounts {
		if acc.RateLimitedUntil.IsZero() {
			continue
		}
		if acc.RateLimitedUntil.Before(now) {
			continue
		}
		if earliest.IsZero() || acc.RateLimitedUntil.Before(earliest) {
			earliest = acc.RateLimitedUntil
		}
	}
	return earliest
}

// FormatRetryAfter formats rateLimitedUntil to human-readable string.
// ref: open-sse/services/accountFallback.js:90-103
func FormatRetryAfter(rateLimitedUntil time.Time) string {
	if rateLimitedUntil.IsZero() {
		return ""
	}
	diff := time.Until(rateLimitedUntil)
	if diff <= 0 {
		return "reset after 0s"
	}
	totalSec := int(diff.Seconds()) + 1
	h := totalSec / 3600
	m := (totalSec % 3600) / 60
	s := totalSec % 60
	parts := make([]string, 0, 3)
	if h > 0 {
		parts = append(parts, formatDuration(h, "h"))
	}
	if m > 0 {
		parts = append(parts, formatDuration(m, "m"))
	}
	if s > 0 || len(parts) == 0 {
		parts = append(parts, formatDuration(s, "s"))
	}
	return "reset after " + strings.Join(parts, " ")
}

func formatDuration(val int, unit string) string {
	return strings.TrimSuffix(strings.TrimSuffix(
		strings.Replace(strings.Replace(string(rune('0'+val/10))+string(rune('0'+val%10)), "0", "", 1), "0", "", 1),
		" "), unit) + unit
}
