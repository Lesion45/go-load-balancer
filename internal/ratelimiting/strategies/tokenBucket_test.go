package strategies

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenBucket_Allow(t *testing.T) {
	t.Run("allows burst requests initially", func(t *testing.T) {
		tb := NewTokenBucket(10, 5)

		for i := 0; i < 5; i++ {
			assert.True(t, tb.Allow(), "Should allow burst requests")
		}

		assert.False(t, tb.Allow(), "Should reject after burst exhausted")
	})

	t.Run("refills tokens after interval", func(t *testing.T) {
		tb := NewTokenBucket(10, 5)

		for i := 0; i < 5; i++ {
			tb.Allow()
		}

		time.Sleep(110 * time.Millisecond)
		assert.True(t, tb.Allow(), "Should allow after refill interval")
		assert.False(t, tb.Allow(), "Should reject until next refill")
	})

	t.Run("never exceeds max tokens", func(t *testing.T) {
		tb := NewTokenBucket(1, 5)

		time.Sleep(6 * time.Second)

		allowed := 0
		for i := 0; i < 10; i++ {
			if tb.Allow() {
				allowed++
			}
		}
		assert.Equal(t, 5, allowed, "Should not exceed max tokens")
	})

	t.Run("handles high rate correctly", func(t *testing.T) {
		tb := NewTokenBucket(1000, 1000)

		for i := 0; i < 1000; i++ {
			assert.True(t, tb.Allow(), "Should allow all burst requests")
		}
		assert.False(t, tb.Allow(), "Should reject after burst")
	})

	t.Run("thread-safe under concurrent access", func(t *testing.T) {
		tb := NewTokenBucket(1000, 1000)
		var wg sync.WaitGroup
		allowed := 0
		mu := sync.Mutex{}

		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if tb.Allow() {
					mu.Lock()
					allowed++
					mu.Unlock()
				}
			}()
		}

		wg.Wait()
		assert.True(t, allowed <= 1000, "Should not exceed max tokens concurrently")
	})

	t.Run("exact timing behavior", func(t *testing.T) {
		tb := NewTokenBucket(2, 2)

		assert.True(t, tb.Allow())
		assert.True(t, tb.Allow())
		assert.False(t, tb.Allow())

		start := time.Now()
		time.Sleep(500 * time.Millisecond)
		assert.True(t, tb.Allow(), "Should allow after exact interval")
		assert.InDelta(t, 500.0, time.Since(start).Milliseconds(), 20.0, "Should respect exact timing")
	})

	t.Run("works correctly with single token", func(t *testing.T) {
		tb := NewTokenBucket(1, 1)

		assert.True(t, tb.Allow(), "Should allow first request")
		assert.False(t, tb.Allow(), "Should reject second request")

		time.Sleep(1100 * time.Millisecond)
		assert.True(t, tb.Allow(), "Should allow after refill")
	})
}
