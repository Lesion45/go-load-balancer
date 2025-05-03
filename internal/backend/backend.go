package backend

import (
	"errors"
	"sync"
	"time"

	"github.com/Lesion45/go-load-balancer/internal/ratelimiting"
)

type SimpleBackend struct {
	URL   string
	alive bool
	ratelimiting.RateLimiter
	lastChecked       time.Time
	activeConnections int64
	failCount         int
	mu                sync.RWMutex
}

// NewSimpleBackend creates a new SimpleBackend instance.
// It returns a pointer to the SimpleBackend instance and an error if the URL is empty.
func NewSimpleBackend(url string, rateLimiterStrategy string, burst int, rate int) (*SimpleBackend, error) {
	if url == "" {
		return nil, errors.New("URL cannot be empty")
	}

	return &SimpleBackend{
		URL:   url,
		alive: true,
		RateLimiter: ratelimiting.NewRateLimiter(
			rateLimiterStrategy,
			burst,
			rate,
		),
		lastChecked:       time.Now(),
		activeConnections: 0,
		failCount:         0,
	}, nil
}

func (b *SimpleBackend) GetUrl() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.URL
}

func (b *SimpleBackend) GetStatus() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.alive
}

func (b *SimpleBackend) SetStatus(status bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.alive = status
}

func (b *SimpleBackend) GetActiveConnections() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.activeConnections
}

func (b *SimpleBackend) AddActiveConnection() int64 {
	b.mu.RLock()
	defer b.mu.RUnlock()
	b.activeConnections++
	return b.activeConnections
}

func (b *SimpleBackend) RemoveActiveConnection() {
	b.mu.RLock()
	defer b.mu.RUnlock()
	b.activeConnections--
	return
}

func (b *SimpleBackend) GetFailCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.failCount
}

func (b *SimpleBackend) IncrementFailCount() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failCount++
}

func (b *SimpleBackend) ResetFailCount() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.failCount = 0
}
