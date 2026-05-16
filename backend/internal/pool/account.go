package pool

import "time"

type AccountState string

const (
	StateActive      AccountState = "active"
	StateRateLimited AccountState = "rate_limited"
	StateExhausted   AccountState = "exhausted"
	StateError       AccountState = "error"
	StateBanned      AccountState = "banned"
)

// RateLimitThreshold: when UsagePercent >= 80%, auto-mark as rate_limited and rotate
const RateLimitThreshold = 0.80

// MaxErrorsBeforeCooldown: consecutive provider errors before marking error state
const MaxErrorsBeforeCooldown = 3

type Account struct {
	ID               string
	Provider         string
	Email            string
	AccessToken      string
	RefreshToken     string
	ExpiresAt        time.Time
	UnavailableUntil time.Time
	IsActive         bool

	State        AccountState
	CreditLimit  float64
	CreditUsed   float64
	BackoffLevel int
	ErrorCount   int
}

func (a *Account) RemainingCredit() float64 {
	return a.CreditLimit - a.CreditUsed
}

func (a *Account) UsagePercent() float64 {
	if a.CreditLimit <= 0 {
		return 0
	}
	return a.CreditUsed / a.CreditLimit
}

func (a *Account) IsAvailable() bool {
	if !a.IsActive {
		return false
	}
	switch a.State {
	case StateExhausted, StateBanned:
		return false
	case StateError, StateRateLimited:
		return time.Now().After(a.UnavailableUntil)
	}
	return true
}

func (a *Account) IsTokenExpired() bool {
	return time.Now().After(a.ExpiresAt)
}
