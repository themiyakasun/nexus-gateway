package ratelimit

import (
	"math"
	"sync"
	"time"
)

type TokenBucket struct {
	capacity float64
	rate float64
	tokens float64
	lastRefill time.Time
	mux sync.Mutex
}

func NewTokenBucket(capacity float64, rate float64) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		rate: rate,
		tokens: capacity,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mux.Lock()
	defer tb.mux.Unlock()

	now := time.Now()

	elapsed := now.Sub(tb.lastRefill).Seconds()

	tb.tokens = math.Min(tb.capacity, tb.tokens+(elapsed*tb.rate))
	tb.lastRefill = now

	if tb.tokens >= 1.0 {
		tb.tokens -= 1.0
		return true
	}

	return false
}

type IPRateLimiter struct {
	limiters map[string]*TokenBucket
	mux sync.RWMutex
	rate float64
	capacity float64
}

func NewIPRateLimiter(capacity float64, rate float64) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*TokenBucket),
		rate: rate,
		capacity: capacity,
	}
}

func (i *IPRateLimiter) GetBucket(ip string) *TokenBucket {
	i.mux.Lock()
	defer i.mux.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = NewTokenBucket(i.capacity, i.rate)
		i.limiters[ip] = limiter
	}

	return limiter
}