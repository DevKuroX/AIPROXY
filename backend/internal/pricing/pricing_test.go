package pricing

import (
	"testing"
)

func TestNewCalculator(t *testing.T) {
	c := NewCalculator()
	if c == nil {
		t.Fatal("NewCalculator returned nil")
	}
}

func TestLookupPricing(t *testing.T) {
	p, err := LookupPricing("gpt-4")
	if err != nil {
		t.Logf("LookupPricing returned: %v", err)
	}
	if p != nil {
		t.Logf("Pricing: %+v", *p)
	}
}

func TestCalculateCost(t *testing.T) {
	p := &Pricing{
		PromptPrice:  10.0,
		CompletionPrice: 30.0,
		Model: "gpt-4",
		
	}
	cost := CalculateCost(1000, 500, p)
	if cost > 0 {
		t.Logf("Cost: %f", cost)
	}
}
