package middleware

import (
	"net/http"
	"sync"
	"time"
)

type bucket struct {
	tokens    float64
	lastRefill time.Time
	capacity  float64
	refillRate float64 // tokens per second
}

// RateLimiter implements token bucket rate limiting per key (IP or API key).
// Thread-safe via sync.RWMutex.
type RateLimiter struct {
	mu      sync.RWMutex
	buckets map[string]*bucket
	ipRate  float64 // per-IP: 100/min = 100/60 per sec
	keyRate float64 // per-key: 1000/min = 1000/60 per sec
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*bucket),
		ipRate:  100.0 / 60.0,  // 100 req/min
		keyRate: 1000.0 / 60.0, // 1000 req/min
	}
}

func (rl *RateLimiter) getBucket(key string, rate float64, capacity float64) *bucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.buckets[key]
	if !exists {
		b = &bucket{
			tokens:     capacity,
			lastRefill: time.Now(),
			capacity:   capacity,
			refillRate: rate,
		}
		rl.buckets[key] = b
		return b
	}

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefill = now

	return b
}

func (rl *RateLimiter) Allow(key string, rate float64, capacity float64) bool {
	b := rl.getBucket(key, rate, capacity)

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// RateLimitByIP middleware limits requests per IP address.
// Returns 429 Too Many Requests when limit exceeded.
func (rl *RateLimiter) RateLimitByIP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		if !rl.Allow(ip, rl.ipRate, 100) {
			http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RateLimitByKey middleware limits requests per API key ID.
// Returns 429 Too Many Requests when limit exceeded.
func (rl *RateLimiter) RateLimitByKey(keyID string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if keyID != "" {
			if !rl.Allow("key:"+keyID, rl.keyRate, 1000) {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
