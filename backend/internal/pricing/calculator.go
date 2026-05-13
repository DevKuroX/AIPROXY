package pricing

import (
	"strings"
	"sync"
)

// Pricing represents the pricing for a model.
// ref: open-sse/services/usage.js
type Pricing struct {
	Model            string  `json:"model"`
	PromptPrice      float64 `json:"prompt_price"`       // Per 1M tokens
	CompletionPrice  float64 `json:"completion_price"`   // Per 1M tokens
	CacheReadPrice   float64 `json:"cache_read_price"`   // Per 1M tokens (optional)
	CacheWritePrice  float64 `json:"cache_write_price"`  // Per 1M tokens (optional)
}

// Calculator handles cost calculations for token usage.
// ref: open-sse/services/usage.js
type Calculator struct {
	mu       sync.RWMutex
	pricings map[string]*Pricing
}

// NewCalculator creates a new pricing calculator with default pricings.
func NewCalculator() *Calculator {
	c := &Calculator{
		pricings: make(map[string]*Pricing),
	}
	for _, p := range DefaultPricings {
		c.pricings[p.Model] = p
	}
	return c
}

// CalculateCost calculates the cost for given token usage.
// Formula: (promptTokens/1M * promptPrice) + (completionTokens/1M * completionPrice)
// ref: open-sse/services/usage.js
func CalculateCost(promptTokens, completionTokens int, pricing *Pricing) float64 {
	if pricing == nil {
		return 0
	}

	const tokensPerMillion = 1_000_000

	promptCost := float64(promptTokens) / tokensPerMillion * pricing.PromptPrice
	completionCost := float64(completionTokens) / tokensPerMillion * pricing.CompletionPrice

	return promptCost + completionCost
}

// LookupPricing finds pricing for a model. Exact match first, then wildcard patterns.
// Wildcard patterns: "claude-*", "gpt-4*", etc.
// ref: open-sse/services/usage.js
func (c *Calculator) LookupPricing(model string) (*Pricing, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Exact match
	if p, ok := c.pricings[model]; ok {
		return p, nil
	}

	// Wildcard matching
	modelLower := strings.ToLower(model)
	for pattern, p := range c.pricings {
		if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(modelLower, strings.ToLower(prefix)) {
				return p, nil
			}
		}
		if strings.HasPrefix(pattern, "*") {
			suffix := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(modelLower, strings.ToLower(suffix)) {
				return p, nil
			}
		}
	}

	return nil, nil
}

// AddPricing adds or updates pricing for a model.
func (c *Calculator) AddPricing(p *Pricing) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pricings[p.Model] = p
}

// RemovePricing removes pricing for a model.
func (c *Calculator) RemovePricing(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.pricings, model)
}

// GetAllPricings returns all registered pricings.
func (c *Calculator) GetAllPricings() []*Pricing {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*Pricing, 0, len(c.pricings))
	for _, p := range c.pricings {
		result = append(result, p)
	}
	return result
}

// Global default calculator instance
var defaultCalculator = NewCalculator()

// LookupPricing is a convenience function using the default calculator.
func LookupPricing(model string) (*Pricing, error) {
	return defaultCalculator.LookupPricing(model)
}
