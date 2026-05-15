package pool

import (
	"errors"
	"sync"
	"time"
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

func (p *Pool) selectAccount(provider string) (*Account, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	accounts := p.accounts[provider]
	if len(accounts) == 0 {
		return nil, errors.New("no accounts for provider: " + provider)
	}

	active := make([]*Account, 0)
	depleting := make([]*Account, 0)
	rateLimited := make([]*Account, 0)

	for _, acc := range accounts {
		if !acc.IsActive {
			continue
		}

		switch acc.State {
		case StateActive:
			if acc.UsagePercent() >= DepletingThreshold {
				acc.State = StateDepleting
				depleting = append(depleting, acc)
			} else {
				active = append(active, acc)
			}

		case StateDepleting:
			if acc.RemainingCredit() <= 0 {
				acc.State = StateExhausted
			} else {
				depleting = append(depleting, acc)
			}

		case StateRateLimited:
			if !time.Now().Before(acc.UnavailableUntil) {
				if acc.RemainingCredit() <= 0 {
					acc.State = StateExhausted
				} else if acc.UsagePercent() >= DepletingThreshold {
					acc.State = StateDepleting
					depleting = append(depleting, acc)
				} else {
					acc.State = StateActive
					active = append(active, acc)
				}
			} else {
				rateLimited = append(rateLimited, acc)
			}

		case StateExhausted:
			continue
		}
	}

	if len(active) > 0 {
		return active[0], nil
	}
	if len(depleting) > 0 {
		return depleting[0], nil
	}
	if len(rateLimited) > 0 {
		soonest := rateLimited[0]
		for _, rl := range rateLimited[1:] {
			if rl.UnavailableUntil.Before(soonest.UnavailableUntil) {
				soonest = rl
			}
		}
		return nil, &PoolRetryableError{
			Message:   "all accounts rate limited",
			RetryAfter: time.Until(soonest.UnavailableUntil),
		}
	}

	return nil, errors.New("no available accounts for provider: " + provider)
}

type PoolRetryableError struct {
	Message   string
	RetryAfter time.Duration
}

func (e *PoolRetryableError) Error() string {
	return e.Message
}

func (p *Pool) MarkUsed(accountID string, creditUsed float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, accounts := range p.accounts {
		for _, acc := range accounts {
			if acc.ID == accountID {
				acc.CreditUsed += creditUsed
				if acc.UsagePercent() >= DepletingThreshold && acc.State == StateActive {
					acc.State = StateDepleting
				}
				if acc.RemainingCredit() <= 0 {
					acc.State = StateExhausted
				}
				return nil
			}
		}
	}
	return errors.New("account not found: " + accountID)
}

func (p *Pool) MarkRateLimited(accountID string, cooldown time.Duration) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, accounts := range p.accounts {
		for _, acc := range accounts {
			if acc.ID == accountID {
				acc.State = StateRateLimited
				acc.UnavailableUntil = time.Now().Add(cooldown)
				acc.BackoffLevel++
				return nil
			}
		}
	}
	return errors.New("account not found: " + accountID)
}

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

func (p *Pool) Count(provider string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	accounts := p.accounts[provider]
	return len(accounts)
}
