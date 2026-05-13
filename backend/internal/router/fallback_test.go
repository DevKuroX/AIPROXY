package router

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCheckFallbackError_TextRules(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		errorText      string
		backoffLevel   int
		wantBackoff    bool
		wantCooldown   time.Duration
		wantNewLevel   int
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
			wantCooldown: CooldownLong,
		},
		{
			name:         "request not allowed text triggers short cooldown",
			status:       403,
			errorText:    "request not allowed",
			backoffLevel: 0,
			wantCooldown: CooldownShort,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckFallbackError(tt.status, tt.errorText, tt.backoffLevel)

			if !result.ShouldFallback {
				t.Error("expected ShouldFallback to be true")
			}

			if tt.wantBackoff {
				if result.NewBackoffLevel != tt.wantNewLevel {
					t.Errorf("expected NewBackoffLevel %d, got %d", tt.wantNewLevel, result.NewBackoffLevel)
				}
				if result.Cooldown == 0 {
					t.Error("expected non-zero cooldown for backoff")
				}
			}

			if tt.wantCooldown != 0 && result.Cooldown != tt.wantCooldown {
				t.Errorf("expected Cooldown %v, got %v", tt.wantCooldown, result.Cooldown)
			}
		})
	}
}

func TestCheckFallbackError_StatusRules(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		errorText    string
		wantCooldown time.Duration
		wantRefresh  bool
	}{
		{
			name:        "401 status triggers long cooldown and refresh",
			status:      401,
			errorText:   "",
			wantCooldown: CooldownLong,
			wantRefresh: true,
		},
		{
			name:        "403 status triggers long cooldown and refresh",
			status:      403,
			errorText:   "",
			wantCooldown: CooldownLong,
			wantRefresh: true,
		},
		{
			name:        "402 status triggers long cooldown",
			status:      402,
			errorText:   "",
			wantCooldown: CooldownLong,
		},
		{
			name:        "404 status triggers long cooldown",
			status:      404,
			errorText:   "",
			wantCooldown: CooldownLong,
		},
		{
			name:        "429 status triggers backoff",
			status:      429,
			errorText:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckFallbackError(tt.status, tt.errorText, 0)

			if !result.ShouldFallback {
				t.Error("expected ShouldFallback to be true")
			}

			if tt.wantCooldown != 0 && result.Cooldown != tt.wantCooldown {
				t.Errorf("expected Cooldown %v, got %v", tt.wantCooldown, result.Cooldown)
			}

			if tt.wantRefresh && !result.NeedsRefresh {
				t.Error("expected NeedsRefresh to be true")
			}
		})
	}
}

func TestCheckFallbackError_DefaultTransient(t *testing.T) {
	result := CheckFallbackError(500, "unknown error", 0)

	if !result.ShouldFallback {
		t.Error("expected ShouldFallback to be true for unmatched errors")
	}
	if result.Cooldown != TransientCooldown {
		t.Errorf("expected TransientCooldown, got %v", result.Cooldown)
	}
}

func TestGetQuotaCooldown(t *testing.T) {
	tests := []struct {
		level    int
		expected time.Duration
	}{
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
	}

	for _, tt := range tests {
		cooldown := getQuotaCooldown(tt.level)
		if cooldown != tt.expected {
			t.Errorf("level %d: expected %v, got %v", tt.level, tt.expected, cooldown)
		}
	}

	maxCooldown := getQuotaCooldown(100)
	if maxCooldown > BackoffMax {
		t.Errorf("cooldown should not exceed BackoffMax, got %v", maxCooldown)
	}
}

func TestFallbackChain_TryNext(t *testing.T) {
	accounts := []*ProviderAccount{
		{Name: "account1", APIKey: "key1", BaseURL: "http://provider1.com"},
		{Name: "account2", APIKey: "key2", BaseURL: "http://provider2.com"},
		{Name: "account3", APIKey: "key3", BaseURL: "http://provider3.com"},
	}

	fc := NewFallbackChain(accounts)

	current := fc.GetCurrent()
	if current == nil {
		t.Fatal("expected non-nil current account")
	}
	if current.Name != "account1" {
		t.Errorf("expected account1, got %s", current.Name)
	}

	next, err := fc.TryNext(context.Background(), current, errors.New("test error"), 429, "rate limit")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next.Name != "account2" {
		t.Errorf("expected account2, got %s", next.Name)
	}

	if current.RateLimitedUntil.IsZero() {
		t.Error("expected RateLimitedUntil to be set")
	}

	next2, err := fc.TryNext(context.Background(), next, errors.New("another error"), 500, "server error")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if next2.Name != "account3" {
		t.Errorf("expected account3, got %s", next2.Name)
	}

	_, err = fc.TryNext(context.Background(), next2, errors.New("last error"), 500, "server error")
	if err == nil {
		t.Error("expected error when all accounts exhausted")
	}
}

func TestFallbackChain_ExhaustedScenario(t *testing.T) {
	accounts := []*ProviderAccount{
		{Name: "account1", APIKey: "key1", BaseURL: "http://provider1.com"},
	}

	fc := NewFallbackChain(accounts)

	current := fc.GetCurrent()
	if current == nil {
		t.Fatal("expected non-nil current account")
	}

	_, err := fc.TryNext(context.Background(), current, errors.New("error"), 500, "server error")
	if err == nil {
		t.Error("expected error when no more accounts")
	}

	var exhaustedErr *AllAccountsExhaustedError
	if !errors.As(err, &exhaustedErr) {
		t.Errorf("expected AllAccountsExhaustedError, got %T", err)
	}
}

func TestIsAccountUnavailable(t *testing.T) {
	tests := []struct {
		name     string
		account  *ProviderAccount
		expected bool
	}{
		{
			name:     "nil account is unavailable",
			account:  nil,
			expected: true,
		},
		{
			name:     "account with no cooldown is available",
			account:  &ProviderAccount{Name: "test"},
			expected: false,
		},
		{
			name: "account with expired cooldown is available",
			account: &ProviderAccount{
				Name:             "test",
				RateLimitedUntil: time.Now().Add(-1 * time.Hour),
			},
			expected: false,
		},
		{
			name: "account with active rate limit is unavailable",
			account: &ProviderAccount{
				Name:             "test",
				RateLimitedUntil: time.Now().Add(1 * time.Hour),
			},
			expected: true,
		},
		{
			name: "account with active unavailable until is unavailable",
			account: &ProviderAccount{
				Name:             "test",
				UnavailableUntil: time.Now().Add(1 * time.Hour),
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsAccountUnavailable(tt.account)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestFormatRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		until    time.Time
		expected string
	}{
		{
			name:     "zero time returns empty",
			until:    time.Time{},
			expected: "",
		},
		{
			name:     "past time returns 0s",
			until:    time.Now().Add(-1 * time.Hour),
			expected: "reset after 0s",
		},
		{
			name:     "30 seconds",
			until:    time.Now().Add(30 * time.Second),
			expected: "reset after 30s",
		},
		{
			name:     "2 minutes 30 seconds",
			until:    time.Now().Add(150 * time.Second),
			expected: "reset after 2m 30s",
		},
		{
			name:     "1 hour 30 minutes",
			until:    time.Now().Add(90 * time.Minute),
			expected: "reset after 1h 30m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRetryAfter(tt.until)
			if !stringsContains(result, tt.expected) {
				t.Errorf("expected result to contain %q, got %q", tt.expected, result)
			}
		})
	}
}

func stringsContains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && stringsContains(s[1:], substr)
}

func TestShouldRetry(t *testing.T) {
	tests := []struct {
		status    int
		errorText string
		expected  bool
	}{
		{429, "rate limit", true},
		{500, "server error", true},
		{401, "unauthorized", true},
		{200, "", false},
	}

	for _, tt := range tests {
		result := ShouldRetry(tt.status, tt.errorText)
		if result != tt.expected {
			t.Errorf("ShouldRetry(%d, %q) = %v, expected %v", tt.status, tt.errorText, result, tt.expected)
		}
	}
}

func TestFallbackChain_RateLimitCooldown(t *testing.T) {
	accounts := []*ProviderAccount{
		{Name: "account1", APIKey: "key1", BaseURL: "http://provider1.com"},
		{Name: "account2", APIKey: "key2", BaseURL: "http://provider2.com"},
	}

	fc := NewFallbackChain(accounts)

	current := fc.GetCurrent()
	_, _ = fc.TryNext(context.Background(), current, errors.New("rate limit"), 429, "rate limit exceeded")

	if current.RateLimitedUntil.IsZero() {
		t.Error("expected RateLimitedUntil to be set after 429")
	}

	if current.BackoffLevel != 1 {
		t.Errorf("expected BackoffLevel 1, got %d", current.BackoffLevel)
	}
}

func TestFallbackChain_NeedsRefresh(t *testing.T) {
	accounts := []*ProviderAccount{
		{Name: "account1", APIKey: "key1", BaseURL: "http://provider1.com"},
	}

	fc := NewFallbackChain(accounts)

	current := fc.GetCurrent()
	_, _ = fc.TryNext(context.Background(), current, errors.New("unauthorized"), 401, "invalid token")

	if !current.NeedsRefresh {
		t.Error("expected NeedsRefresh to be true for 401 error")
	}
}

func TestHandleChatCompletions_Fallback(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": {"message": "rate limit exceeded"}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "success"}}]}`))
	}))
	defer server.Close()

	accounts := []*ProviderAccount{
		{Name: "account1", APIKey: "key1", BaseURL: server.URL},
		{Name: "account2", APIKey: "key2", BaseURL: server.URL},
	}

	fc := NewFallbackChain(accounts)

	current := fc.GetCurrent()
	if current == nil {
		t.Fatal("expected current account")
	}

	_, err := fc.TryNext(context.Background(), current, errors.New("rate limit"), 429, "rate limit exceeded")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	next := fc.GetCurrent()
	if next == nil || next.Name != "account2" {
		t.Errorf("expected to move to account2, got %v", next)
	}
}

func TestGetEarliestRateLimitedUntil(t *testing.T) {
	now := time.Now()
	accounts := []*ProviderAccount{
		{Name: "account1", RateLimitedUntil: now.Add(1 * time.Hour)},
		{Name: "account2", RateLimitedUntil: now.Add(30 * time.Minute)},
		{Name: "account3", RateLimitedUntil: now.Add(2 * time.Hour)},
	}

	earliest := GetEarliestRateLimitedUntil(accounts)
	if earliest.IsZero() {
		t.Error("expected non-zero earliest time")
	}

	expected := now.Add(30 * time.Minute)
	diff := earliest.Sub(expected)
	if diff < -time.Second || diff > time.Second {
		t.Errorf("expected earliest around %v, got %v", expected, earliest)
	}
}

func TestAllAccountsExhaustedError(t *testing.T) {
	err := NewAllAccountsExhaustedError("last error message")

	if err.Error() != "all accounts exhausted: last error message" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	if !errors.Is(err, ErrAllAccountsExhausted) {
		t.Error("expected to match ErrAllAccountsExhausted")
	}
}
