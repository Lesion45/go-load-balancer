package strategies

import (
	"testing"

	"github.com/Lesion45/go-load-balancer/internal/backend"
	"github.com/Lesion45/go-load-balancer/internal/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestLeastConnections_NextBackend(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("empty backends list returns empty string", func(t *testing.T) {
		lb := &LeastConnections{
			Log:      logger,
			Backends: []*backend.SimpleBackend{},
		}

		url, err := lb.NextBackend()
		assert.NoError(t, err)
		assert.Empty(t, url, "Should return empty string when no backends available")
	})

	t.Run("all unhealthy backends return empty string", func(t *testing.T) {
		b1 := utils.CreateTestBackend("http://backend1", false, 0)
		b2 := utils.CreateTestBackend("http://backend2", false, 0)

		lb := &LeastConnections{
			Log:      logger,
			Backends: []*backend.SimpleBackend{b1, b2},
		}

		url, err := lb.NextBackend()
		assert.NoError(t, err)
		assert.Empty(t, url, "Should return empty string when all backends are unhealthy")
	})

	t.Run("single healthy backend is always selected", func(t *testing.T) {
		b1 := utils.CreateTestBackend("http://backend1", true, 0)

		lb := &LeastConnections{
			Log:      logger,
			Backends: []*backend.SimpleBackend{b1},
		}

		url, err := lb.NextBackend()
		assert.NoError(t, err)
		assert.Equal(t, b1, url, "Should select the only healthy backend")
		assert.Equal(t, int64(1), b1.GetActiveConnections(), "Should increment connection count")
	})

	t.Run("select backend with minimum connections", func(t *testing.T) {
		b1 := utils.CreateTestBackend("http://backend1", true, 3)
		b2 := utils.CreateTestBackend("http://backend2", true, 1)
		b3 := utils.CreateTestBackend("http://backend3", true, 2)

		lb := &LeastConnections{
			Log:      logger,
			Backends: []*backend.SimpleBackend{b1, b2, b3},
		}

		url, err := lb.NextBackend()
		assert.NoError(t, err)
		assert.Equal(t, b2, url, "Should select backend with least connections")
		assert.Equal(t, int64(2), b2.GetActiveConnections(), "Should increment connection count")
	})

	t.Run("equal connections selects first healthy backend", func(t *testing.T) {
		b1 := utils.CreateTestBackend("http://backend1", true, 1)
		b2 := utils.CreateTestBackend("http://backend2", true, 1)

		lb := &LeastConnections{
			Log:      logger,
			Backends: []*backend.SimpleBackend{b1, b2},
		}

		url, err := lb.NextBackend()
		assert.NoError(t, err)
		assert.Equal(t, b1, url, "Should select first backend when connections are equal")
	})

	t.Run("only consider healthy backends", func(t *testing.T) {
		b1 := utils.CreateTestBackend("http://backend1", false, 0) // unhealthy
		b2 := utils.CreateTestBackend("http://backend2", true, 2)  // healthy
		b3 := utils.CreateTestBackend("http://backend3", false, 0) // unhealthy
		b4 := utils.CreateTestBackend("http://backend4", true, 1)  // healthy

		lb := &LeastConnections{
			Log:      logger,
			Backends: []*backend.SimpleBackend{b1, b2, b3, b4},
		}

		url, err := lb.NextBackend()
		assert.NoError(t, err)
		assert.Equal(t, b4, url, "Should select healthy backend with least connections")
	})

	t.Run("connection counter increments on selection", func(t *testing.T) {
		b1 := utils.CreateTestBackend("http://backend1", true, 0)
		initialConnections := b1.GetActiveConnections()

		lb := &LeastConnections{
			Log:      logger,
			Backends: []*backend.SimpleBackend{b1},
		}

		_, err := lb.NextBackend()
		assert.NoError(t, err)
		assert.Equal(t, initialConnections+1, b1.GetActiveConnections(), "Should increment connection count after selection")
	})
}
