package services

import (
	"context"
	"testing"
	"time"
)

func TestNewAccountFallbackService(t *testing.T) {
	t.Run("with nil logger uses default", func(t *testing.T) {
		svc := NewAccountFallbackService(nil)
		if svc == nil {
			t.Error("expected non-nil service")
		}
	})
}

func TestGetQuotaCooldown(t *testing.T) {
	tests := []struct {
		name          string
		backoffLevel  int
		wantMin       time.Duration
		wantMax       time.Duration
	}{
		{
			name:         "level 0 returns base",
			backoffLevel: 0,
			wantMin:      0,
			wantMax:      backoffConfig.base,
		},
		{
			name:         "level 1 returns base",
			backoffLevel: 1,
			wantMin:      backoffConfig.base,
			wantMax:      backoffConfig.base,
		},
		{
			name:         "level 2 returns 2x base",
			backoffLevel: 2,
			wantMin:      backoffConfig.base * 2,
			wantMax:      backoffConfig.base * 2,
		},
		{
			name:         "level 3 returns 4x base",
			backoffLevel: 3,
			wantMin:      backoffConfig.base * 4,
			wantMax:      backoffConfig.base * 4,
		},
		{
			name:         "high level caps at max",
			backoffLevel: 20,
			wantMin:      backoffConfig.max,
			wantMax:      backoffConfig.max,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAccountFallbackService(nil)
			got := svc.GetQuotaCooldown(tt.backoffLevel)

			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("GetQuotaCooldown(%d) = %v, want between %v and %v", tt.backoffLevel, got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCheckFallbackError_TextRules(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		errorText    string
		backoffLevel int
		wantBackoff  bool
		wantCooldown time.Duration
		wantNewLevel int
	}{
		{
			name:         "rate limit text triggers backoff",
			status:       500,
			errorText:    "rate limit exceeded",
			backoffLevel: 0,
			wantBackoff:  true,
			wantNewLevel: 1,
		},
		{
			name:         "too many requests text triggers backoff",
			status:       500,
			errorText:    "too many requests",
			backoffLevel: 2,
			wantBackoff:  true,
			wantNewLevel: 3,
		},
		{
			name:         "quota exceeded text triggers backoff",
			status:       402,
			errorText:    "quota exceeded for account",
			backoffLevel: 0,
			wantBackoff:  true,
			wantNewLevel: 1,
		},
		{
			name:         "capacity text triggers backoff",
			status:       503,
			errorText:    "insufficient capacity",
			backoffLevel: 0,
			wantBackoff:  true,
			wantNewLevel: 1,
		},
		{
			name:         "overloaded text triggers backoff",
			status:       503,
			errorText:    "service overloaded",
			backoffLevel: 0,
			wantBackoff:  true,
			wantNewLevel: 1,
		},
		{
			name:         "no credentials text triggers long cooldown",
			status:       401,
			errorText:    "no credentials provided",
			backoffLevel: 0,
			wantCooldown: cooldown.long,
		},
		{
			name:         "request not allowed text triggers short cooldown",
			status:       403,
			errorText:    "request not allowed",
			backoffLevel: 0,
			wantCooldown: cooldown.short,
		},
		{
			name:         "improperly formed request triggers long cooldown",
			status:       400,
			errorText:    "improperly formed request",
			backoffLevel: 0,
			wantCooldown: cooldown.long,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAccountFallbackService(nil)
			result := svc.CheckFallbackError(context.Background(), tt.status, tt.errorText, tt.backoffLevel)

			if !result.ShouldFallback {
				t.Error("expected ShouldFallback to be true")
			}

			if tt.wantBackoff {
				if result.NewBackoffLevel == nil {
					t.Error("expected NewBackoffLevel to be non-nil for backoff case")
				} else if *result.NewBackoffLevel != tt.wantNewLevel {
					t.Errorf("expected NewBackoffLevel %d, got %d", tt.wantNewLevel, *result.NewBackoffLevel)
				}
				if result.CooldownMs == 0 {
					t.Error("expected non-zero cooldown for backoff")
				}
			}

			if tt.wantCooldown != 0 && result.CooldownMs != tt.wantCooldown {
				t.Errorf("expected CooldownMs %v, got %v", tt.wantCooldown, result.CooldownMs)
			}
		})
	}
}

func TestCheckFallbackError_StatusRules(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		errorText    string
		backoffLevel int
		wantBackoff  bool
		wantCooldown time.Duration
	}{
		{
			name:         "status 401 triggers long cooldown",
			status:       401,
			errorText:    "",
			backoffLevel: 0,
			wantCooldown: cooldown.long,
		},
		{
			name:         "status 402 triggers long cooldown",
			status:       402,
			errorText:    "",
			backoffLevel: 0,
			wantCooldown: cooldown.long,
		},
		{
			name:         "status 403 triggers long cooldown",
			status:       403,
			errorText:    "",
			backoffLevel: 0,
			wantCooldown: cooldown.long,
		},
		{
			name:         "status 404 triggers long cooldown",
			status:       404,
			errorText:    "",
			backoffLevel: 0,
			wantCooldown: cooldown.long,
		},
		{
			name:         "status 429 triggers backoff",
			status:       429,
			errorText:    "",
			backoffLevel: 0,
			wantBackoff:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAccountFallbackService(nil)
			result := svc.CheckFallbackError(context.Background(), tt.status, tt.errorText, tt.backoffLevel)

			if !result.ShouldFallback {
				t.Error("expected ShouldFallback to be true")
			}

			if tt.wantBackoff {
				if result.NewBackoffLevel == nil {
					t.Error("expected NewBackoffLevel for backoff case")
				}
			}

			if tt.wantCooldown != 0 && result.CooldownMs != tt.wantCooldown {
				t.Errorf("expected CooldownMs %v, got %v", tt.wantCooldown, result.CooldownMs)
			}
		})
	}
}

func TestCheckFallbackError_TextRulesTakePriority(t *testing.T) {
	// Text rules should take priority over status rules
	// Status 429 with "no credentials" text should use text rule (long cooldown)
	// not status rule (backoff)
	svc := NewAccountFallbackService(nil)
	result := svc.CheckFallbackError(context.Background(), 429, "no credentials", 0)

	if !result.ShouldFallback {
		t.Error("expected ShouldFallback to be true")
	}

	// Text rule for "no credentials" has fixed cooldown, not backoff
	if result.NewBackoffLevel != nil {
		t.Error("expected no backoff for 'no credentials' text rule even with 429 status")
	}

	if result.CooldownMs != cooldown.long {
		t.Errorf("expected long cooldown for 'no credentials', got %v", result.CooldownMs)
	}
}

func TestCheckFallbackError_UnknownError(t *testing.T) {
	svc := NewAccountFallbackService(nil)
	result := svc.CheckFallbackError(context.Background(), 500, "unknown error", 0)

	if !result.ShouldFallback {
		t.Error("expected ShouldFallback to be true for unknown errors")
	}

	if result.CooldownMs != transientCooldownMs {
		t.Errorf("expected transient cooldown %v, got %v", transientCooldownMs, result.CooldownMs)
	}
}

func TestCheckFallbackError_CaseInsensitive(t *testing.T) {
	svc := NewAccountFallbackService(nil)

	tests := []struct {
		name      string
		errorText string
		wantBackoff bool
	}{
		{"lowercase rate limit", "rate limit exceeded", true},
		{"uppercase RATE LIMIT", "RATE LIMIT EXCEEDED", true},
		{"mixed case Rate Limit", "Rate Limit Exceeded", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.CheckFallbackError(context.Background(), 500, tt.errorText, 0)

			if !result.ShouldFallback {
				t.Error("expected ShouldFallback to be true")
			}

			if tt.wantBackoff && result.NewBackoffLevel == nil {
				t.Error("expected backoff for rate limit text")
			}
		})
	}
}

func TestIsAccountUnavailable(t *testing.T) {
	tests := []struct {
		name              string
		unavailableUntil  string
		wantUnavailable   bool
	}{
		{
			name:             "empty string means available",
			unavailableUntil: "",
			wantUnavailable:  false,
		},
		{
			name:             "future time means unavailable",
			unavailableUntil: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			wantUnavailable:  true,
		},
		{
			name:             "past time means available",
			unavailableUntil: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
			wantUnavailable:  false,
		},
		{
			name:             "invalid format means available",
			unavailableUntil: "not-a-timestamp",
			wantUnavailable:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewAccountFallbackService(nil)
			got := svc.IsAccountUnavailable(tt.unavailableUntil)

			if got != tt.wantUnavailable {
				t.Errorf("IsAccountUnavailable(%q) = %v, want %v", tt.unavailableUntil, got, tt.wantUnavailable)
			}
		})
	}
}

func TestGetUnavailableUntil(t *testing.T) {
	svc := NewAccountFallbackService(nil)
	cooldownMs := 5 * time.Minute

	result := svc.GetUnavailableUntil(cooldownMs)

	if result == "" {
		t.Error("expected non-empty result")
	}

	// Parse the result and verify it's approximately 5 minutes from now
	parsed, err := time.Parse(time.RFC3339, result)
	if err != nil {
		t.Errorf("failed to parse result: %v", err)
	}

	expected := time.Now().Add(cooldownMs)
	diff := parsed.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("expected timestamp within 1s of %v, got %v (diff: %v)", expected, parsed, diff)
	}
}

func TestGetQuotaCooldownStandalone(t *testing.T) {
	tests := []struct {
		name         string
		backoffLevel int
	}{
		{"level 0", 0},
		{"level 1", 1},
		{"level 5", 5},
		{"level 15", 15},
		{"level 20 (over max)", 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetQuotaCooldown(tt.backoffLevel)

			if got <= 0 {
				t.Error("expected positive cooldown")
			}

			if got > backoffConfig.max {
				t.Errorf("cooldown %v exceeds max %v", got, backoffConfig.max)
			}
		})
	}
}

func TestCheckFallbackErrorStandalone(t *testing.T) {
	result := CheckFallbackError(429, "rate limit", 0)

	if !result.ShouldFallback {
		t.Error("expected ShouldFallback to be true")
	}

	if result.NewBackoffLevel == nil {
		t.Error("expected backoff for rate limit")
	}
}

func TestIsAccountUnavailableStandalone(t *testing.T) {
	if IsAccountUnavailable("") {
		t.Error("expected false for empty string")
	}

	future := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	if !IsAccountUnavailable(future) {
		t.Error("expected true for future timestamp")
	}
}

func TestGetUnavailableUntilStandalone(t *testing.T) {
	result := GetUnavailableUntil(1 * time.Minute)

	if result == "" {
		t.Error("expected non-empty result")
	}

	_, err := time.Parse(time.RFC3339, result)
	if err != nil {
		t.Errorf("failed to parse result: %v", err)
	}
}

func TestCheckFallbackError_BackoffLevelCap(t *testing.T) {
	svc := NewAccountFallbackService(nil)

	// Test that backoff level is capped at maxLevel
	result := svc.CheckFallbackError(context.Background(), 429, "rate limit", backoffConfig.maxLevel)

	if result.NewBackoffLevel == nil {
		t.Fatal("expected NewBackoffLevel to be set")
	}

	if *result.NewBackoffLevel > backoffConfig.maxLevel {
		t.Errorf("backoff level %d exceeds max %d", *result.NewBackoffLevel, backoffConfig.maxLevel)
	}

	// Even starting at maxLevel, should stay at max
	result2 := svc.CheckFallbackError(context.Background(), 429, "rate limit", backoffConfig.maxLevel+10)
	if result2.NewBackoffLevel == nil {
		t.Fatal("expected NewBackoffLevel to be set")
	}

	if *result2.NewBackoffLevel > backoffConfig.maxLevel {
		t.Errorf("backoff level %d exceeds max %d", *result2.NewBackoffLevel, backoffConfig.maxLevel)
	}
}
