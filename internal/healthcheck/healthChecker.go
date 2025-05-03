package healthcheck

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Lesion45/go-load-balancer/internal/backend"
	"go.uber.org/zap"
)

// HealthChecker defines the interface for health checkers.
type HealthChecker interface {
	Start(ctx context.Context, backends []*backend.SimpleBackend)
	runHealthCheckWorker(ctx context.Context, backend *backend.SimpleBackend)
	getCheckInterval(backend *backend.SimpleBackend) time.Duration
	Check(ctx context.Context, backend *backend.SimpleBackend) (bool, error)
}

// HTTPHealthChecker implements the HealthChecker interface for HTTP health checks.
type HTTPHealthChecker struct {
	Log          *zap.Logger
	Client       *http.Client
	Endpoint     string
	BaseInterval time.Duration
	MaxInterval  time.Duration
}

// NewHTTPHealthChecker creates a new HTTPHealthChecker instance.
// It takes a logger, an HTTP client, an endpoint, a base interval, and a maximum interval as arguments.
// It panics if the logger or client are nil, or if the endpoint is empty.
// It also panics if the base interval or maximum interval are less than or equal to zero.
func NewHTTPHealthChecker(log *zap.Logger, client *http.Client, endpoint string, baseInterval, maxInterval time.Duration) *HTTPHealthChecker {
	if log == nil {
		panic("log cannot be nil")
	}

	if client == nil {
		panic("client cannot be nil")
	}

	if endpoint == "" {
		panic("endpoint cannot be empty")
	}

	if baseInterval <= 0 {
		panic("baseInterval must be greater than zero")
	}

	if maxInterval <= 0 {
		panic("maxInterval must be greater than zero")
	}

	return &HTTPHealthChecker{
		Log:          log,
		Client:       client,
		Endpoint:     endpoint,
		BaseInterval: baseInterval,
		MaxInterval:  maxInterval,
	}
}

// Start starts the health check workers for the provided backends.
// It logs the start of the workers and starts a goroutine for each backend.
func (c *HTTPHealthChecker) Start(ctx context.Context, backends []*backend.SimpleBackend) {
	const op = "HTTPHealthChecker.Start"

	c.Log.Info("starting workers for backends", zap.String("op", op))

	for _, backend := range backends {
		go c.runHealthCheckWorker(ctx, backend)
		c.Log.Info("worker has started", zap.String("op", op), zap.String("client", backend.GetUrl()))
	}

	c.Log.Info("all workers have started", zap.String("op", op), zap.Int("count", len(backends)))
}

// startWorker starts a health check worker for a specific backend.
// It enters a loop to perform health checks at intervals.
func (c *HTTPHealthChecker) runHealthCheckWorker(ctx context.Context, backend *backend.SimpleBackend) {
	const op = "HTTPHealthChecker.runHealthCheckWorker"

	for {
		select {
		case <-ctx.Done():
			c.Log.Info("shutting down health checker worker",
				zap.String("op", op),
				zap.String("client", backend.GetUrl()),
			)

			return
		default:
			interval := c.getCheckInterval(backend)
			url := backend.GetUrl()
			time.Sleep(interval)

			start := time.Now()
			c.Log.Info("healthcheck has started",
				zap.String("op", op),
				zap.String("client", url),
				zap.Bool("alive", backend.GetStatus()),
			)

			alive, err := c.Check(ctx, backend)
			if err != nil {
				c.Log.Warn("error during healthcheck",
					zap.String("op", op),
					zap.String("client", url),
					zap.Error(err),
				)
			}

			backend.SetStatus(alive)

			if alive {
				c.Log.Info("client is healthy",
					zap.String("op", op),
					zap.String("client", url),
				)
				backend.ResetFailCount()
			} else {
				c.Log.Warn("client is unhealthy",
					zap.String("op", op),
					zap.String("client", url),
					zap.Int("fail count", backend.GetFailCount()),
				)
				backend.IncrementFailCount()
			}

			c.Log.Info("healthcheck has ended",
				zap.String("op", op),
				zap.String("client", url),
				zap.Bool("alive", backend.GetStatus()),
				zap.Duration("took", time.Since(start)),
			)
		}
	}
}

// getCheckInterval calculates the interval for the next health check.
func (c *HTTPHealthChecker) getCheckInterval(backend *backend.SimpleBackend) time.Duration {
	failCount := backend.GetFailCount()
	if failCount == 0 {
		return c.BaseInterval
	}

	interval := c.BaseInterval * time.Duration(1<<failCount)
	if interval > c.MaxInterval {
		return c.MaxInterval
	}
	return interval
}

// Check performs the health check for a specific backend.
// It sends an HTTP GET request to the backend's URL and checks the response status code.
// If the status code is 200, it returns true; otherwise, it returns false.
func (c *HTTPHealthChecker) Check(ctx context.Context, backend *backend.SimpleBackend) (bool, error) {
	const op = "HTTPHealthChecker.Check"

	url := backend.GetUrl() + c.Endpoint
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.Log.Error("failed to create request",
			zap.String("op", op),
			zap.Error(err),
		)

		return false, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := c.Client.Do(req)
	if err != nil {
		c.Log.Error("failed to send request",
			zap.String("op", op),
			zap.String("client", url),
			zap.Error(err),
		)

		return false, fmt.Errorf("%s: %w", op, err)
	}

	if resp.StatusCode != 200 {
		c.Log.Warn("unexpected status code",
			zap.String("op", op),
			zap.String("client", url),
			zap.Int("code", resp.StatusCode),
		)

		return false, nil
	}

	return true, nil
}
