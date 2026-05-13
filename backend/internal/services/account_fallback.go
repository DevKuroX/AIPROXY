package services

import (
	"context"
	"log/slog"
	"strings"
	"time"
)

// ref: open-sse/config/errorConfig.js:32-36
// Exponential backoff config for rate limits
var backoffConfig = struct {
	base     time.Duration
	max      time.Duration
	maxLevel int
}{
	base:     2000 * time.Millisecond,
	max:      5 * time.Minute,
	maxLevel: 15,
}

// ref: open-sse/config/errorConfig.js:39
// Default cooldown for transient/unknown errors
const transientCooldownMs = 30 * time.Second

// ref: open-sse/config/errorConfig.js:45-48
// Cooldown durations (ms)
var cooldown = struct {
	long  time.Duration
	short time.Duration
}{
	long:  2 * time.Minute,
	short: 5 * time.Second,
}

// ref: open-sse/config/errorConfig.js:59-76
// ErrorRule represents a single error classification rule
type ErrorRule struct {
	Text       string        // substring match (case-insensitive) on error message
	Status     int           // HTTP status code match
	CooldownMs time.Duration // fixed cooldown duration
	Backoff    bool          // true = use exponential backoff (rate limit)
}

// ref: open-sse/config/errorConfig.js:59-76
// Unified error classification rules.
// Checked top-to-bottom: text rules first (by order), then status rules.
var errorRules = []ErrorRule{
	// --- Text-based rules (checked first, order = priority) ---
	{Text: "no credentials", CooldownMs: cooldown.long},
	{Text: "request not allowed", CooldownMs: cooldown.short},
	{Text: "improperly formed request", CooldownMs: cooldown.long},
	{Text: "rate limit", Backoff: true},
	{Text: "too many requests", Backoff: true},
	{Text: "quota exceeded", Backoff: true},
	{Text: "capacity", Backoff: true},
	{Text: "overloaded", Backoff: true},

	// --- Status-based rules (fallback when text doesn't match) ---
	{Status: 401, CooldownMs: cooldown.long},
	{Status: 402, CooldownMs: cooldown.long},
	{Status: 403, CooldownMs: cooldown.long},
	{Status: 404, CooldownMs: cooldown.long},
	{Status: 429, Backoff: true},
}

// FallbackResult represents the result of checking a fallback error
// ref: open-sse/services/accountFallback.js:21-22
type FallbackResult struct {
	ShouldFallback  bool          `json:"should_fallback"`
	CooldownMs      time.Duration `json:"cooldown_ms"`
	NewBackoffLevel *int          `json:"new_backoff_level,omitempty"` // nil if no backoff
}

// AccountFallbackService handles account fallback decisions based on errors
type AccountFallbackService struct {
	logger *slog.Logger
}

// NewAccountFallbackService creates a new AccountFallbackService
func NewAccountFallbackService(logger *slog.Logger) *AccountFallbackService {
	if logger == nil {
		logger = slog.Default()
	}
	return &AccountFallbackService{logger: logger}
}

// GetQuotaCooldown calculates exponential backoff cooldown for rate limits (429)
// Level 1: 1s, Level 2: 2s, Level 3: 4s... → max 5 min
// ref: open-sse/services/accountFallback.js:9-13
func (s *AccountFallbackService) GetQuotaCooldown(backoffLevel int) time.Duration {
	level := max(0, backoffLevel-1)
	cooldown := time.Duration(float64(backoffConfig.base) * float64(uint(1)<<level))
	return min(cooldown, backoffConfig.max)
}

// CheckFallbackError checks if error should trigger account fallback (switch to next account)
// Config-driven: matches ERROR_RULES top-to-bottom (text rules first, then status)
// ref: open-sse/services/accountFallback.js:23-50
func (s *AccountFallbackService) CheckFallbackError(ctx context.Context, status int, errorText string, backoffLevel int) FallbackResult {
	lowerError := ""
	if errorText != "" {
		lowerError = strings.ToLower(errorText)
	}

	for _, rule := range errorRules {
		// Text-based rule: match substring in error message
		// ref: open-sse/services/accountFallback.js:29-36
		if rule.Text != "" && lowerError != "" && strings.Contains(lowerError, rule.Text) {
			if rule.Backoff {
				newLevel := min(backoffLevel+1, backoffConfig.maxLevel)
				return FallbackResult{
					ShouldFallback:  true,
					CooldownMs:      s.GetQuotaCooldown(newLevel),
					NewBackoffLevel: &newLevel,
				}
			}
			return FallbackResult{
				ShouldFallback: true,
				CooldownMs:     rule.CooldownMs,
			}
		}

		// Status-based rule: match HTTP status code
		// ref: open-sse/services/accountFallback.js:38-45
		if rule.Status != 0 && rule.Status == status {
			if rule.Backoff {
				newLevel := min(backoffLevel+1, backoffConfig.maxLevel)
				return FallbackResult{
					ShouldFallback:  true,
					CooldownMs:      s.GetQuotaCooldown(newLevel),
					NewBackoffLevel: &newLevel,
				}
			}
			return FallbackResult{
				ShouldFallback: true,
				CooldownMs:     rule.CooldownMs,
			}
		}
	}

	// Default: transient cooldown for any unmatched error
	// ref: open-sse/services/accountFallback.js:48-49
	return FallbackResult{
		ShouldFallback: true,
		CooldownMs:     transientCooldownMs,
	}
}

// IsAccountUnavailable checks if account is currently unavailable (cooldown not expired)
// ref: open-sse/services/accountFallback.js:55-58
func (s *AccountFallbackService) IsAccountUnavailable(unavailableUntil string) bool {
	if unavailableUntil == "" {
		return false
	}

	t, err := time.Parse(time.RFC3339, unavailableUntil)
	if err != nil {
		s.logger.Warn("failed to parse unavailableUntil timestamp", "value", unavailableUntil, "error", err)
		return false
	}

	return t.After(time.Now())
}

// GetUnavailableUntil calculates unavailable until timestamp as ISO string
// ref: open-sse/services/accountFallback.js:63-65
func (s *AccountFallbackService) GetUnavailableUntil(cooldownMs time.Duration) string {
	return time.Now().Add(cooldownMs).Format(time.RFC3339)
}

// GetQuotaCooldown standalone function for convenience
func GetQuotaCooldown(backoffLevel int) time.Duration {
	level := max(0, backoffLevel-1)
	cooldown := time.Duration(float64(backoffConfig.base) * float64(uint(1)<<level))
	return min(cooldown, backoffConfig.max)
}

// CheckFallbackError standalone function for convenience
func CheckFallbackError(status int, errorText string, backoffLevel int) FallbackResult {
	lowerError := ""
	if errorText != "" {
		lowerError = strings.ToLower(errorText)
	}

	for _, rule := range errorRules {
		if rule.Text != "" && lowerError != "" && strings.Contains(lowerError, rule.Text) {
			if rule.Backoff {
				newLevel := min(backoffLevel+1, backoffConfig.maxLevel)
				return FallbackResult{
					ShouldFallback:  true,
					CooldownMs:      GetQuotaCooldown(newLevel),
					NewBackoffLevel: &newLevel,
				}
			}
			return FallbackResult{
				ShouldFallback: true,
				CooldownMs:     rule.CooldownMs,
			}
		}

		if rule.Status != 0 && rule.Status == status {
			if rule.Backoff {
				newLevel := min(backoffLevel+1, backoffConfig.maxLevel)
				return FallbackResult{
					ShouldFallback:  true,
					CooldownMs:      GetQuotaCooldown(newLevel),
					NewBackoffLevel: &newLevel,
				}
			}
			return FallbackResult{
				ShouldFallback: true,
				CooldownMs:     rule.CooldownMs,
			}
		}
	}

	return FallbackResult{
		ShouldFallback: true,
		CooldownMs:     transientCooldownMs,
	}
}

// IsAccountUnavailable standalone function for convenience
func IsAccountUnavailable(unavailableUntil string) bool {
	if unavailableUntil == "" {
		return false
	}

	t, err := time.Parse(time.RFC3339, unavailableUntil)
	if err != nil {
		return false
	}

	return t.After(time.Now())
}

// GetUnavailableUntil standalone function for convenience
func GetUnavailableUntil(cooldownMs time.Duration) string {
	return time.Now().Add(cooldownMs).Format(time.RFC3339)
}
