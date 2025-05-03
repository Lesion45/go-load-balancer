package strategies

import (
	"sync"
	"time"
)

// TokenBucket is a rate limiting algorithm that allows a burst of requests
type TokenBucket struct {
	tokens         int
	maxTokens      int
	refillInterval time.Duration
	lastRefill     time.Time
	mu             sync.Mutex
}

// NewTokenBucket creates a new TokenBucket with the specified rate and burst size
// rate - tokens per second (must be > 0)
// burst - maximum tokens available (must be > 0)
// Returns a pointer to the TokenBucket instance or panics if invalid parameters are provided
func NewTokenBucket(rate int, burst int) *TokenBucket {
	if rate <= 0 {
		panic("rate must be greater than 0")
	}

	if burst <= 0 {
		panic("burst must be greater than 0")

	}

	return &TokenBucket{
		tokens:         burst,
		maxTokens:      burst,
		refillInterval: time.Second / time.Duration(rate),
		lastRefill:     time.Now(),
	}
}

// Allow checks if a request can be allowed based on the current token count
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)

	newTokens := int(elapsed / tb.refillInterval)
	if newTokens > 0 {
		tb.tokens = min(tb.maxTokens, tb.tokens+newTokens)
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}
