package strategies

import (
	"errors"
	"sync/atomic"

	"github.com/Lesion45/go-load-balancer/internal/backend"
	"go.uber.org/zap"
)

// RoundRobin implements a round-robin load balancing strategy.
type RoundRobin struct {
	Log          *zap.Logger
	CurrentIndex *atomic.Int64
	Backends     []*backend.SimpleBackend
}

// NewRoundRobin creates a new RoundRobin instance.
// It takes a logger and a slice of backends as arguments.
// It panics if log or backends are nil.
func NewRoundRobin(log *zap.Logger, backends []*backend.SimpleBackend) *RoundRobin {
	if log == nil {
		panic("log cannot be nil")
	}

	if backends == nil {
		panic("backends cannot be nil")
	}

	return &RoundRobin{
		Log:          log,
		CurrentIndex: &atomic.Int64{},
		Backends:     backends,
	}
}

// NextBackend selects the next backend in a round-robin fashion.
// It returns the URL of the selected backend and an error if no backends are available.
func (s *RoundRobin) NextBackend() (*backend.SimpleBackend, error) {
	const op = "RoundRobin.getNextBackend"

	s.Log.Info("trying to get next backend", zap.String("operation", op))

	if len(s.Backends) == 0 {
		s.Log.Error("no Backends available", zap.String("operation", op))
		return nil, nil
	}

	index := s.CurrentIndex.Add(1) % int64(len(s.Backends))
	attempts := 0

	for attempts < len(s.Backends) {
		url := s.Backends[index].GetUrl()
		if s.Backends[index].GetStatus() {
			s.Log.Info("selected backend",
				zap.String("operation", op),
				zap.String("backend", url),
			)
			return s.Backends[index], nil
		}
		attempts++
		index = s.CurrentIndex.Add(1) % int64(len(s.Backends))
	}

	s.Log.Warn("all Backends are down",
		zap.String("operation", op),
	)

	return nil, errors.New("all backends are down")
}
