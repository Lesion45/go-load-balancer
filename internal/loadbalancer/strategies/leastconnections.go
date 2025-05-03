package strategies

import (
	"github.com/Lesion45/go-load-balancer/internal/backend"
	"go.uber.org/zap"
)

// LeastConnections implements a load balancing strategy that selects the backend with the least number of active connections.
type LeastConnections struct {
	Log      *zap.Logger
	Backends []*backend.SimpleBackend
}

// NewLeastConnections creates a new LeastConnections load balancing strategy.
// It takes a logger and a slice of backends as arguments.
// It panics if the logger or backends are nil.
func NewLeastConnections(log *zap.Logger, backends []*backend.SimpleBackend) *LeastConnections {
	if log == nil {
		panic("log cannot be nil")
	}

	if backends == nil {
		panic("backends cannot be nil")
	}

	return &LeastConnections{
		Log:      log,
		Backends: backends,
	}
}

// NextBackend selects the backend with the least number of active connections.
// It returns the URL of the selected backend and an error if no backends are available.
func (l *LeastConnections) NextBackend() (*backend.SimpleBackend, error) {
	const op = "LeastConnections.getNextBackend"

	l.Log.Info("trying to get next backend", zap.String("operation", op))

	var selectedBackend *backend.SimpleBackend
	minConn := int64(1<<63 - 1)

	for _, b := range l.Backends {
		if b.GetStatus() && (selectedBackend == nil || b.GetActiveConnections() < minConn) {
			selectedBackend = b
			minConn = b.GetActiveConnections()
		}
	}

	if selectedBackend == nil {
		l.Log.Warn("no Backends available", zap.String("operation", op))
		return nil, nil
	}

	l.Log.Info("selected backend",
		zap.String("operation", op),
		zap.String("backend", selectedBackend.GetUrl()),
	)

	selectedBackend.AddActiveConnection()

	return selectedBackend, nil
}
