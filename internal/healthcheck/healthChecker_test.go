package healthcheck

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Lesion45/go-load-balancer/internal/backend"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestHTTPHealthChecker(t *testing.T) {
	t.Run("constructor creates valid instance", func(t *testing.T) {
		logger := zaptest.NewLogger(t)
		client := &http.Client{}
		endpoint := "/health"
		baseInterval := 1 * time.Second
		maxInterval := 30 * time.Second

		checker := NewHTTPHealthChecker(logger, client, endpoint, baseInterval, maxInterval)

		assert.Equal(t, logger, checker.Log)
		assert.Equal(t, client, checker.Client)
		assert.Equal(t, endpoint, checker.Endpoint)
		assert.Equal(t, baseInterval, checker.BaseInterval)
		assert.Equal(t, maxInterval, checker.MaxInterval)
	})

	t.Run("getCheckInterval calculates correct intervals", func(t *testing.T) {
		testCases := []struct {
			name         string
			failCount    int
			expected     time.Duration
			baseInterval time.Duration
			maxInterval  time.Duration
		}{
			{
				name:         "no failures - use base interval",
				failCount:    0,
				baseInterval: 1 * time.Second,
				maxInterval:  30 * time.Second,
				expected:     1 * time.Second,
			},
			{
				name:         "one failure - double interval",
				failCount:    1,
				baseInterval: 1 * time.Second,
				maxInterval:  30 * time.Second,
				expected:     2 * time.Second,
			},
			{
				name:         "three failures - exponential backoff",
				failCount:    3,
				baseInterval: 1 * time.Second,
				maxInterval:  30 * time.Second,
				expected:     8 * time.Second,
			},
			{
				name:         "many failures - capped at max interval",
				failCount:    10,
				baseInterval: 1 * time.Second,
				maxInterval:  30 * time.Second,
				expected:     30 * time.Second,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				logger := zaptest.NewLogger(t)
				checker := &HTTPHealthChecker{
					Log:          logger,
					BaseInterval: tc.baseInterval,
					MaxInterval:  tc.maxInterval,
				}

				backend := &backend.SimpleBackend{}
				for i := 0; i < tc.failCount; i++ {
					backend.IncrementFailCount()
				}

				interval := checker.getCheckInterval(backend)
				assert.Equal(t, tc.expected, interval)
			})
		}
	})

	t.Run("Check method handles different responses", func(t *testing.T) {
		testCases := []struct {
			name          string
			setupServer   func() *httptest.Server
			expectAlive   bool
			expectError   bool
			expectedError error
		}{
			{
				name: "successful health check (200 OK)",
				setupServer: func() *httptest.Server {
					return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusOK)
					}))
				},
				expectAlive: true,
				expectError: false,
			},
			{
				name: "unhealthy backend (500 status)",
				setupServer: func() *httptest.Server {
					return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						w.WriteHeader(http.StatusInternalServerError)
					}))
				},
				expectAlive: false,
				expectError: false,
			},
			{
				name: "connection refused",
				setupServer: func() *httptest.Server {
					srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
					srv.Close()
					return srv
				},
				expectAlive: false,
				expectError: true,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				logger := zaptest.NewLogger(t)
				server := tc.setupServer()
				if server != nil {
					defer server.Close()
				}

				checker := &HTTPHealthChecker{
					Log:      logger,
					Client:   server.Client(),
					Endpoint: "/health",
				}

				backend := &backend.SimpleBackend{URL: server.URL}
				alive, err := checker.Check(context.Background(), backend)

				if tc.expectError {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.Equal(t, tc.expectAlive, alive)
			})
		}
	})
}
