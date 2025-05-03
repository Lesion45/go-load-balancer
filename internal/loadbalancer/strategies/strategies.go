package strategies

import (
	"fmt"

	"github.com/Lesion45/go-load-balancer/internal/backend"
	"github.com/Lesion45/go-load-balancer/internal/loadbalancer"
	"go.uber.org/zap"
)

// BalancerStrategy defines the interface for load balancing strategies.
type BalancerStrategy interface {
	NextBackend() (*backend.SimpleBackend, error)
}

// NewBalancerStrategy creates a new load balancer strategy based on the provided settings.
func NewBalancerStrategy(log *zap.Logger, settings *loadbalancer.Settings) BalancerStrategy {
	switch settings.StrategyName {
	case "round_robin":
		return NewRoundRobin(log, settings.Backends)
	case "least_connections":
		return NewLeastConnections(log, settings.Backends)
	default:
		panic(fmt.Sprintf("unknown strategy: %s", settings.StrategyName))
	}
}
