package pool

import (
	"errors"
	"math"
	"sync"
	"time"
)

const (
	BackoffBase     = 1 * time.Second
	BackoffMax      = 4 * time.Minute
	BackoffMaxLevel = 8
)

type Pool struct {
	accounts map[string][]*Account
	mu       sync.RWMutex
}

func New() *Pool {
	return &Pool{accounts: make(map[string][]*Account)}
}

func (p *Pool) AddAccount(account *Account) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if account == nil {
		return
	}
	if account.State == "" {
		account.State = StateActive
	}
	p.accounts[account.Provider] = append(p.accounts[account.Provider], account)
}

func (p *Pool) GetAccount(provider string) (*Account, error) {
	return p.selectAccount(provider)
}

// selectAccount implements STICKY selection (not load balance):
//  1. Pick the active account with most remaining credit
//  2. Keep using it until it hits rate_limited (80%) or exhausted
//  3. When no active left, fallback to rate_limited accounts whose cooldown has expired
//  4. If all are in cooldown, return retryable error with the soonest available
func (p *Pool) selectAccount(provider string) (*Account, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	accounts := p.accounts[provider]
	if len(accounts) == 0 {
		return nil, errors.New("no accounts for provider: " + provider)
	}

	var bestActive *Account
	var bestCooldown *Account
	var soonestTime time.Time

	for _, acc := range accounts {
		if !acc.IsActive {
			continue
		}

		// Credit-based transition: Active → RateLimited at 80%
		if acc.State == StateActive && acc.UsagePercent() >= RateLimitThreshold {
			acc.State = StateRateLimited
			acc.UnavailableUntil = time.Now().Add(calculateBackoff(acc.BackoffLevel))
		}

		// Expired rate_limited cooldown → back to active
		if acc.State == StateRateLimited && time.Now().After(acc.UnavailableUntil) {
			acc.BackoffLevel = 0
			if acc.RemainingCredit() <= 0 {
				acc.State = StateExhausted
			} else {
				acc.State = StateActive
			}
		}

		// Expired error cooldown → back to active
		if acc.State == StateError && time.Now().After(acc.UnavailableUntil) {
			acc.BackoffLevel = 0
			acc.ErrorCount = 0
			acc.State = StateActive
		}

		switch acc.State {
		case StateActive:
			if bestActive == nil || acc.RemainingCredit() > bestActive.RemainingCredit() {
				bestActive = acc
			}

		case StateRateLimited, StateError:
			if bestCooldown == nil || acc.UnavailableUntil.Before(soonestTime) {
				bestCooldown = acc
				soonestTime = acc.UnavailableUntil
			}

		case StateExhausted, StateBanned:
			continue
		}
	}

	// 1. Active account available (sticky: same one gets picked each time)
	if bestActive != nil {
		return bestActive, nil
	}

	// 2. No active → fallback: pick the cooldown that expires soonest
	if bestCooldown != nil {
		return nil, &PoolRetryableError{
			Message:    "all accounts unavailable for " + provider,
			RetryAfter: time.Until(soonestTime),
		}
	}

	return nil, errors.New("no available accounts for provider: " + provider)
}

type PoolRetryableError struct {
	Message    string
	RetryAfter time.Duration
}

func (e *PoolRetryableError) Error() string {
	return e.Message
}

// MarkUsed tracks credit consumption and transitions to rate_limited/exhausted.
func (p *Pool) MarkUsed(accountID string, creditUsed float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, accounts := range p.accounts {
		for _, acc := range accounts {
			if acc.ID == accountID {
				acc.CreditUsed += creditUsed
				if acc.RemainingCredit() <= 0 {
					acc.State = StateExhausted
				} else if acc.UsagePercent() >= RateLimitThreshold && acc.State == StateActive {
					acc.State = StateRateLimited
					acc.UnavailableUntil = time.Now().Add(calculateBackoff(acc.BackoffLevel))
				}
				return nil
			}
		}
	}
	return errors.New("account not found: " + accountID)
}

// MarkRateLimited transitions account to rate_limited with exponential backoff cooldown.
func (p *Pool) MarkRateLimited(accountID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, accounts := range p.accounts {
		for _, acc := range accounts {
			if acc.ID == accountID {
				acc.State = StateRateLimited
				acc.BackoffLevel++
				if acc.BackoffLevel > BackoffMaxLevel {
					acc.BackoffLevel = BackoffMaxLevel
				}
				acc.UnavailableUntil = time.Now().Add(calculateBackoff(acc.BackoffLevel))
				return nil
			}
		}
	}
	return errors.New("account not found: " + accountID)
}

// MarkError increments error counter. On 3rd consecutive error, transitions to StateError.
func (p *Pool) MarkError(accountID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, accounts := range p.accounts {
		for _, acc := range accounts {
			if acc.ID == accountID {
				acc.ErrorCount++
				if acc.ErrorCount >= MaxErrorsBeforeCooldown {
					acc.State = StateError
					acc.BackoffLevel++
					if acc.BackoffLevel > BackoffMaxLevel {
						acc.BackoffLevel = BackoffMaxLevel
					}
					acc.UnavailableUntil = time.Now().Add(calculateBackoff(acc.BackoffLevel))
				}
				return nil
			}
		}
	}
	return errors.New("account not found: " + accountID)
}

// MarkSuccess resets the error counter (call on successful provider response).
func (p *Pool) MarkSuccess(accountID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, accounts := range p.accounts {
		for _, acc := range accounts {
			if acc.ID == accountID {
				acc.ErrorCount = 0
				return nil
			}
		}
	}
	return errors.New("account not found: " + accountID)
}

// MarkExhausted manually marks an account as exhausted.
func (p *Pool) MarkExhausted(accountID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, accounts := range p.accounts {
		for _, acc := range accounts {
			if acc.ID == accountID {
				acc.State = StateExhausted
				return nil
			}
		}
	}
	return errors.New("account not found: " + accountID)
}

// MarkBanned permanently disables an account (manual action only).
func (p *Pool) MarkBanned(accountID string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, accounts := range p.accounts {
		for _, acc := range accounts {
			if acc.ID == accountID {
				acc.State = StateBanned
				return nil
			}
		}
	}
	return errors.New("account not found: " + accountID)
}

func (p *Pool) Count(provider string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	accounts := p.accounts[provider]
	return len(accounts)
}

// SyncAccounts merges DB accounts into pool, preserving in-memory state for existing ones.
// New accounts are added; removed accounts stay but are marked inactive.
func (p *Pool) SyncAccounts(provider string, dbAccounts []*Account) {
	p.mu.Lock()
	defer p.mu.Unlock()

	existing := p.accounts[provider]
	existingByID := make(map[string]*Account, len(existing))
	for _, acc := range existing {
		existingByID[acc.ID] = acc
	}

	seen := make(map[string]bool, len(dbAccounts))
	merged := make([]*Account, 0, len(dbAccounts))

	for _, dbAcc := range dbAccounts {
		if dbAcc == nil {
			continue
		}
		seen[dbAcc.ID] = true

		if cur, ok := existingByID[dbAcc.ID]; ok {
			// Preserve in-memory state, update DB fields
			cur.Email = dbAcc.Email
			cur.IsActive = dbAcc.IsActive
			cur.CreditLimit = dbAcc.CreditLimit
			cur.CreditUsed = dbAcc.CreditUsed
			cur.ExpiresAt = dbAcc.ExpiresAt
			if dbAcc.AccessToken != "" && dbAcc.AccessToken != "virtual" {
				cur.AccessToken = dbAcc.AccessToken
			}
			if dbAcc.RefreshToken != "" {
				cur.RefreshToken = dbAcc.RefreshToken
			}
			merged = append(merged, cur)
		} else {
			// New account from DB
			if dbAcc.State == "" {
				dbAcc.State = StateActive
			}
			merged = append(merged, dbAcc)
		}
	}

	// Carry over accounts not in DB but still in pool (mark as inactive so they drain)
	for _, acc := range existing {
		if !seen[acc.ID] {
			acc.IsActive = false
			merged = append(merged, acc)
		}
	}

	p.accounts[provider] = merged
}

// ListAccounts returns all accounts for a provider (for sync/debug).
func (p *Pool) ListAccounts(provider string) []*Account {
	p.mu.RLock()
	defer p.mu.RUnlock()
	accounts := p.accounts[provider]
	result := make([]*Account, len(accounts))
	copy(result, accounts)
	return result
}

func calculateBackoff(level int) time.Duration {
	if level < 0 {
		level = 0
	}
	if level > BackoffMaxLevel {
		level = BackoffMaxLevel
	}
	d := time.Duration(math.Pow(2, float64(level))) * BackoffBase
	if d > BackoffMax {
		return BackoffMax
	}
	return d
}
