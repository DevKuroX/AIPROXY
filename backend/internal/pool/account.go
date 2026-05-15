package pool

import "time"

type AccountState string

const (
	StateActive      AccountState = "active"
	StateDepleting   AccountState = "depleting"
	StateRateLimited AccountState = "rate_limited"
	StateExhausted   AccountState = "exhausted"
)

const DepletingThreshold = 0.80

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
	if time.Now().Before(a.UnavailableUntil) {
		return false
	}
	return true
}

func (a *Account) IsTokenExpired() bool {
	return time.Now().After(a.ExpiresAt)
}
