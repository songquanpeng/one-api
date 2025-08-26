package model

import (
	"testing"
	"time"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/stretchr/testify/assert"
)

// TestCacheSync tests the cache synchronization functionality
func TestCacheSync(t *testing.T) {
	// Enable memory cache for testing
	config.MemoryCacheEnabled = true

	// Initialize database connection for testing
	// Note: This assumes test database is configured
	InitDB()

	// Test channel cache initialization
	t.Run("TestInitChannelCache", func(t *testing.T) {
		// Initialize cache
		InitChannelCache()

		// Verify cache is built
		channelSyncLock.RLock()
		cacheExists := group2model2channels != nil
		channelSyncLock.RUnlock()

		assert.True(t, cacheExists, "Channel cache should be initialized")
	})

	// Test cache invalidation
	t.Run("TestInvalidateChannelCache", func(t *testing.T) {
		// Initialize cache first
		InitChannelCache()

		// Test cache invalidation
		InvalidateChannelCache(1)

		// Cache should be rebuilt after invalidation
		channelSyncLock.RLock()
		cacheExists := group2model2channels != nil
		channelSyncLock.RUnlock()

		assert.True(t, cacheExists, "Cache should exist after invalidation")
	})

	// Test abilities-based cache building
	t.Run("TestBuildChannelCacheFromAbilities", func(t *testing.T) {
		// Create test channels
		channels := []*Channel{
			{
				Id:       1,
				Status:   ChannelStatusEnabled,
				Group:    "default",
				Models:   "gpt-3.5-turbo",
				Priority: &[]int64{0}[0],
			},
		}

		// Build cache from abilities
		cache := buildChannelCacheFromAbilities(channels)

		assert.NotNil(t, cache, "Cache should not be nil")
	})
}

// TestChannelStatusUpdate tests channel status update with cache sync
func TestChannelStatusUpdate(t *testing.T) {
	// Enable memory cache for testing
	config.MemoryCacheEnabled = true

	// Test channel status update
	t.Run("TestUpdateChannelStatusById", func(t *testing.T) {
		// This test requires a test database with actual data
		// For now, we'll just test that the function doesn't panic
		assert.NotPanics(t, func() {
			UpdateChannelStatusById(999, ChannelStatusEnabled)
		}, "UpdateChannelStatusById should not panic")
	})
}

// BenchmarkCacheInit benchmarks cache initialization performance
func BenchmarkCacheInit(b *testing.B) {
	config.MemoryCacheEnabled = true
	InitDB()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		InitChannelCache()
	}
}

// TestCacheConsistency verifies cache consistency after updates
func TestCacheConsistency(t *testing.T) {
	config.MemoryCacheEnabled = true

	t.Run("TestCacheConsistencyAfterUpdate", func(t *testing.T) {
		// Initialize cache
		InitChannelCache()

		// Simulate channel status change
		InvalidateChannelCache(1)

		// Wait a bit for cache to rebuild
		time.Sleep(100 * time.Millisecond)

		// Verify cache is still consistent
		channelSyncLock.RLock()
		cacheExists := group2model2channels != nil
		channelSyncLock.RUnlock()

		assert.True(t, cacheExists, "Cache should remain consistent after updates")
	})
}
