package router

import (
	"math"
	"time"
)

const (
	RateLimitBackoffBase     = 1 * time.Second
	RateLimitBackoffMax      = 4 * time.Minute
	RateLimitBackoffMaxLevel = 8
)

func CalculateBackoff(level int) time.Duration {
	if level < 0 {
		level = 0
	}
	if level > RateLimitBackoffMaxLevel {
		level = RateLimitBackoffMaxLevel
	}
	d := time.Duration(math.Pow(2, float64(level))) * RateLimitBackoffBase
	if d > RateLimitBackoffMax {
		return RateLimitBackoffMax
	}
	return d
}


