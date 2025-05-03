package ratelimiting

import (
	"fmt"

	"github.com/Lesion45/go-load-balancer/internal/ratelimiting/strategies"
)

// RateLimiter provides an interface for rate limiting strategies
type RateLimiter interface {
	Allow() bool
}

// NewRateLimiter creates a new RateLimiter based on the specified strategy name.
func NewRateLimiter(strategyName string, burst, rate int) RateLimiter {
	switch strategyName {
	case "token_bucket":
		return strategies.NewTokenBucket(rate, burst)
	default:
		panic(fmt.Sprintf("unknown rate limiting strategy: %s", strategyName))
	}
}
