package pool

import (
	"testing"
	"time"
)

func TestNewPool(t *testing.T) {
	p := New()
	if p == nil {
		t.Fatal("New() returned nil")
	}
	if p.Count("test") != 0 {
		t.Fatal("New pool should have 0 accounts")
	}
}

func TestAddAndGetAccount(t *testing.T) {
	p := New()

	acc := &Account{
		ID:          "acc1",
		Provider:    "test-provider",
		Email:       "test@test.com",
		AccessToken: "token1",
		IsActive:    true,
	}
	p.AddAccount(acc)

	if p.Count("test-provider") != 1 {
		t.Fatalf("expected 1 account, got %d", p.Count("test-provider"))
	}

	got, err := p.GetAccount("test-provider")
	if err != nil {
		t.Fatalf("GetAccount failed: %v", err)
	}
	if got.ID != "acc1" {
		t.Fatalf("expected acc1, got %s", got.ID)
	}
}

func TestGetAccountNoAccounts(t *testing.T) {
	p := New()
	_, err := p.GetAccount("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}
}

func TestMarkRateLimited(t *testing.T) {
	p := New()
	acc := &Account{ID: "acc1", Provider: "test", IsActive: true, CreditLimit: 100}
	p.AddAccount(acc)

	if err := p.MarkRateLimited("acc1"); err != nil {
		t.Fatalf("MarkRateLimited failed: %v", err)
	}

	// Account should be in backoff
	if acc.State != StateRateLimited {
		t.Fatalf("expected StateRateLimited, got %s", acc.State)
	}
	if acc.BackoffLevel != 1 {
		t.Fatalf("expected backoff level 1, got %d", acc.BackoffLevel)
	}
}

func TestMarkError(t *testing.T) {
	p := New()
	acc := &Account{ID: "acc1", Provider: "test", IsActive: true}
	p.AddAccount(acc)

	// First 2 errors should not trigger cooldown
	for i := 0; i < MaxErrorsBeforeCooldown-1; i++ {
		if err := p.MarkError("acc1"); err != nil {
			t.Fatalf("MarkError attempt %d failed: %v", i, err)
		}
	}

	if acc.State == StateError {
		t.Fatal("State should not be StateError before reaching limit")
	}

	// 3rd error should trigger cooldown
	if err := p.MarkError("acc1"); err != nil {
		t.Fatalf("MarkError attempt 3 failed: %v", err)
	}
	if acc.State != StateError {
		t.Fatalf("expected StateError, got %s", acc.State)
	}
}

func TestMarkSuccess(t *testing.T) {
	p := New()
	acc := &Account{ID: "acc1", Provider: "test", IsActive: true}
	p.AddAccount(acc)

	p.MarkError("acc1")
	p.MarkError("acc1")
	if acc.ErrorCount != 2 {
		t.Fatalf("expected error count 2, got %d", acc.ErrorCount)
	}

	p.MarkSuccess("acc1")
	if acc.ErrorCount != 0 {
		t.Fatalf("expected error count 0 after success, got %d", acc.ErrorCount)
	}
}

func TestMarkExhausted(t *testing.T) {
	p := New()
	acc := &Account{ID: "acc1", Provider: "test", IsActive: true, State: StateActive}
	p.AddAccount(acc)

	p.MarkExhausted("acc1")
	if acc.State != StateExhausted {
		t.Fatalf("expected StateExhausted, got %s", acc.State)
	}
}

func TestMarkBanned(t *testing.T) {
	p := New()
	acc := &Account{ID: "acc1", Provider: "test", IsActive: true}
	p.AddAccount(acc)

	p.MarkBanned("acc1")
	if acc.State != StateBanned {
		t.Fatalf("expected StateBanned, got %s", acc.State)
	}
}

func TestAccountStateTransitions(t *testing.T) {
	p := New()
	acc := &Account{ID: "acc1", Provider: "test", IsActive: true, CreditLimit: 100, CreditUsed: 0}
	p.AddAccount(acc)

	// Initial state
	if acc.State != StateActive {
		t.Fatalf("expected default StateActive, got %s", acc.State)
	}

	// Use 80% credit → should auto-transition to rate_limited
	p.MarkUsed("acc1", 85)
	if acc.State != StateRateLimited {
		t.Fatalf("expected StateRateLimited at high usage, got %s", acc.State)
	}

	// After backoff expires with no credit → exhausted
	acc.UnavailableUntil = time.Now().Add(-1 * time.Second)
	acc.CreditUsed = 100
	// Next selectAccount should transition to exhausted
	p.GetAccount("test")
	if acc.State != StateExhausted {
		t.Fatalf("expected StateExhausted after backoff with 0 credit, got %s", acc.State)
	}
}

func TestSelectAccountSticky(t *testing.T) {
	p := New()
	acc1 := &Account{ID: "acc1", Provider: "test", IsActive: true, CreditLimit: 100, CreditUsed: 10}
	acc2 := &Account{ID: "acc2", Provider: "test", IsActive: true, CreditLimit: 100, CreditUsed: 50}
	p.AddAccount(acc1)
	p.AddAccount(acc2)

	// Should pick acc1 (more remaining credit)
	got, _ := p.GetAccount("test")
	if got.ID != "acc1" {
		t.Fatalf("expected acc1 (more credit), got %s", got.ID)
	}
}

func TestSelectAccountFallbackToCooldown(t *testing.T) {
	p := New()
	acc1 := &Account{ID: "acc1", Provider: "test", IsActive: true, CreditLimit: 100, CreditUsed: 95}
	acc1.State = StateRateLimited
	acc1.UnavailableUntil = time.Now().Add(10 * time.Minute)

	acc2 := &Account{ID: "acc2", Provider: "test", IsActive: true, CreditLimit: 100, CreditUsed: 50}
	acc2.State = StateRateLimited
	acc2.UnavailableUntil = time.Now().Add(1 * time.Minute)

	p.AddAccount(acc1)
	p.AddAccount(acc2)

	_, err := p.GetAccount("test")
	if err == nil {
		t.Fatal("expected retryable error when all accounts rate limited")
	}
}

func TestAccountIsAvailable(t *testing.T) {
	acc := &Account{ID: "acc1", IsActive: true, State: StateActive}
	if !acc.IsAvailable() {
		t.Fatal("active account should be available")
	}

	acc.State = StateExhausted
	if acc.IsAvailable() {
		t.Fatal("exhausted account should not be available")
	}

	acc.State = StateBanned
	if acc.IsAvailable() {
		t.Fatal("banned account should not be available")
	}

	acc.State = StateRateLimited
	acc.UnavailableUntil = time.Now().Add(-1 * time.Second)
	if !acc.IsAvailable() {
		t.Fatal("rate_limited account with expired cooldown should be available")
	}
}

func TestRemainingCredit(t *testing.T) {
	acc := &Account{CreditLimit: 100, CreditUsed: 30}
	if acc.RemainingCredit() != 70 {
		t.Fatalf("expected 70 remaining, got %f", acc.RemainingCredit())
	}
}

func TestUsagePercent(t *testing.T) {
	acc := &Account{CreditLimit: 100, CreditUsed: 25}
	if acc.UsagePercent() != 0.25 {
		t.Fatalf("expected 0.25 usage, got %f", acc.UsagePercent())
	}

	acc.CreditLimit = 0
	if acc.UsagePercent() != 0 {
		t.Fatalf("expected 0 usage when no limit, got %f", acc.UsagePercent())
	}
}

func TestCalculateBackoff(t *testing.T) {
	if calculateBackoff(0) != 1*time.Second {
		t.Fatalf("expected 1s backoff, got %v", calculateBackoff(0))
	}
	if calculateBackoff(3) != 8*time.Second {
		t.Fatalf("expected 8s backoff, got %v", calculateBackoff(3))
	}
	if calculateBackoff(10) > BackoffMax {
		t.Fatalf("backoff should not exceed max")
	}
	if calculateBackoff(-1) != 1*time.Second {
		t.Fatalf("negative level should default to 1s")
	}
}

func TestSyncAccounts(t *testing.T) {
	p := New()

	// Add existing account with state and future cooldown
	existing := &Account{
		ID: "acc1", Provider: "test", IsActive: true,
		State: StateRateLimited, BackoffLevel: 3, CreditUsed: 50,
		UnavailableUntil: time.Now().Add(10 * time.Minute),
	}
	p.AddAccount(existing)

	// Sync with updated DB accounts (include active acc2 so it's not removed)
	active := &Account{ID: "acc2", Provider: "test", IsActive: true, CreditLimit: 100}
	dbAcc := &Account{
		ID: "acc1", Provider: "test", IsActive: true,
		CreditLimit: 100, AccessToken: "new-token",
	}
	p.SyncAccounts("test", []*Account{dbAcc, active})

	// acc1 (rate_limited) should NOT be returned by GetAccount (it picks active first)
	got, err := p.GetAccount("test")
	if err != nil {
		t.Fatalf("GetAccount failed: %v", err)
	}
	if got.ID != "acc2" {
		t.Fatalf("expected acc2 (active) to be selected, got %s", got.ID)
	}

	// Directly verify acc1 preserved its state
	accounts := p.ListAccounts("test")
	var found *Account
	for _, a := range accounts {
		if a.ID == "acc1" {
			found = a
			break
		}
	}
	if found == nil {
		t.Fatal("acc1 not found after sync")
	}
	if found.State != StateRateLimited {
		t.Fatalf("Sync should preserve StateRateLimited, got %s", found.State)
	}
	if found.BackoffLevel != 3 {
		t.Fatalf("Sync should preserve BackoffLevel 3, got %d", found.BackoffLevel)
	}
	if found.AccessToken != "new-token" {
		t.Fatalf("Sync should update AccessToken, got %s", found.AccessToken)
	}
}
