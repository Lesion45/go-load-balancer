package strategies

import (
	"sync/atomic"
	"testing"

	"github.com/Lesion45/go-load-balancer/internal/backend"
	"github.com/Lesion45/go-load-balancer/internal/utils"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestRoundRobin_NextBackend(t *testing.T) {
	logger := zaptest.NewLogger(t)

	t.Run("empty backends list returns empty string", func(t *testing.T) {
		rr := &RoundRobin{
			Log:          logger,
			CurrentIndex: &atomic.Int64{},
			Backends:     []*backend.SimpleBackend{},
		}

		url, err := rr.NextBackend()
		assert.NoError(t, err)
		assert.Empty(t, url, "Should return empty string when no backends available")
	})

	t.Run("single healthy backend is always selected", func(t *testing.T) {
		b1 := utils.CreateTestBackend("http://backend1", true, 0)
		rr := &RoundRobin{
			Log:          logger,
			CurrentIndex: &atomic.Int64{},
			Backends:     []*backend.SimpleBackend{b1},
		}

		for i := 0; i < 3; i++ {
			url, err := rr.NextBackend()
			assert.NoError(t, err)
			assert.Equal(t, b1, url, "Should always return the only healthy backend")
		}
	})

	t.Run("all unhealthy backends return empty string", func(t *testing.T) {
		b1 := utils.CreateTestBackend("http://backend1", false, 0)
		b2 := utils.CreateTestBackend("http://backend2", false, 0)
		rr := &RoundRobin{
			Log:          logger,
			CurrentIndex: &atomic.Int64{},
			Backends:     []*backend.SimpleBackend{b1, b2},
		}

		url, err := rr.NextBackend()
		assert.Error(t, err)
		assert.Empty(t, url, "Should return empty string when all backends are unhealthy")
	})

	t.Run("skips unhealthy backends in rotation", func(t *testing.T) {
		b1 := utils.CreateTestBackend("http://backend1", false, 0) // unhealthy
		b2 := utils.CreateTestBackend("http://backend2", true, 0)
		b3 := utils.CreateTestBackend("http://backend3", false, 0) // unhealthy
		b4 := utils.CreateTestBackend("http://backend4", true, 0)
		rr := &RoundRobin{
			Log:          logger,
			CurrentIndex: &atomic.Int64{},
			Backends:     []*backend.SimpleBackend{b1, b2, b3, b4},
		}

		url, err := rr.NextBackend()
		assert.NoError(t, err)
		assert.Equal(t, b2, url)

		url, err = rr.NextBackend()
		assert.NoError(t, err)
		assert.Equal(t, b4, url)

		url, err = rr.NextBackend()
		assert.NoError(t, err)
		assert.Equal(t, b2, url)
	})

	t.Run("returns empty when all backends become unhealthy", func(t *testing.T) {
		b1 := utils.CreateTestBackend("http://backend1", true, 0)
		b2 := utils.CreateTestBackend("http://backend2", true, 0)
		rr := &RoundRobin{
			Log:          logger,
			CurrentIndex: &atomic.Int64{},
			Backends:     []*backend.SimpleBackend{b1, b2},
		}

		b1.SetStatus(false)
		b2.SetStatus(false)

		url, err := rr.NextBackend()
		assert.Error(t, err)
		assert.Empty(t, url, "Should return empty when all backends are unhealthy")
	})

	t.Run("handles index overflow correctly", func(t *testing.T) {
		b1 := utils.CreateTestBackend("http://backend1", true, 0)
		b2 := utils.CreateTestBackend("http://backend2", true, 0)
		rr := &RoundRobin{
			Log:          logger,
			CurrentIndex: &atomic.Int64{},
			Backends:     []*backend.SimpleBackend{b1, b2},
		}

		rr.CurrentIndex.Store(int64(^uint64(0) >> 1))

		url, err := rr.NextBackend()
		assert.NoError(t, err)
		assert.NotEmpty(t, url, "Should handle index overflow correctly")
	})
}
