package loadbalancer

import (
	"github.com/Lesion45/go-load-balancer/internal/backend"
)

// Settings holds the configuration for the load balancer strategy.
type Settings struct {
	StrategyName string
	Backends     []*backend.SimpleBackend
}
