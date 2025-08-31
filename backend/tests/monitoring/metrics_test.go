package monitoring

import (
	"testing"
	"time"

	"github.com/gofromzero/project_temp/backend/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_RecordHttpRequest(t *testing.T) {
	metrics := utils.GetMetrics()
	metrics.Reset() // Start with clean state
	
	// Test recording HTTP request
	method := "GET"
	path := "/test"
	duration := 100 * time.Millisecond
	statusCode := 200
	
	metrics.RecordHttpRequest(method, path, duration, statusCode)
	
	// Verify request count
	key := "GET:/test"
	assert.Equal(t, int64(1), metrics.HttpRequestsTotal[key], "Request count should be 1")
	
	// Verify status code count
	assert.Equal(t, int64(1), metrics.HttpResponseStatus[200], "Status code 200 count should be 1")
	
	// Verify duration recording
	avgDuration := metrics.GetAverageHttpDuration(method, path)
	assert.Equal(t, duration, avgDuration, "Average duration should match recorded duration")
}

func TestMetrics_RecordDatabaseQuery(t *testing.T) {
	metrics := utils.GetMetrics()
	metrics.Reset() // Start with clean state
	
	duration := 50 * time.Millisecond
	
	metrics.RecordDatabaseQuery(duration)
	
	assert.Equal(t, int64(1), metrics.DatabaseQueryTotal, "Database query total should be 1")
	
	avgDuration := metrics.GetAverageDatabaseQueryDuration()
	assert.Equal(t, duration, avgDuration, "Average database query duration should match recorded duration")
}

func TestMetrics_CacheMetrics(t *testing.T) {
	metrics := utils.GetMetrics()
	metrics.Reset() // Start with clean state
	
	// Record cache hits and misses
	metrics.RecordCacheHit()
	metrics.RecordCacheHit()
	metrics.RecordCacheMiss()
	
	assert.Equal(t, int64(2), metrics.RedisCacheHits, "Cache hits should be 2")
	assert.Equal(t, int64(1), metrics.RedisCacheMisses, "Cache misses should be 1")
	
	hitRatio := metrics.GetCacheHitRatio()
	expected := 2.0 / 3.0 // 2 hits out of 3 total
	assert.InDelta(t, expected, hitRatio, 0.01, "Cache hit ratio should be approximately 0.67")
}

func TestMetrics_ConnectionCounts(t *testing.T) {
	metrics := utils.GetMetrics()
	metrics.Reset() // Start with clean state
	
	// Set connection counts
	metrics.SetDatabaseConnections(5)
	metrics.SetRedisConnections(3)
	
	assert.Equal(t, int64(5), metrics.DatabaseConnectionsActive, "Database connections should be 5")
	assert.Equal(t, int64(3), metrics.RedisConnectionsActive, "Redis connections should be 3")
}

func TestMetrics_MultipleRequests(t *testing.T) {
	metrics := utils.GetMetrics()
	metrics.Reset() // Start with clean state
	
	// Record multiple requests to same endpoint
	method := "POST"
	path := "/api/users"
	
	durations := []time.Duration{
		100 * time.Millisecond,
		200 * time.Millisecond,
		150 * time.Millisecond,
	}
	
	for _, duration := range durations {
		metrics.RecordHttpRequest(method, path, duration, 201)
	}
	
	key := "POST:/api/users"
	assert.Equal(t, int64(3), metrics.HttpRequestsTotal[key], "Request count should be 3")
	assert.Equal(t, int64(3), metrics.HttpResponseStatus[201], "Status code 201 count should be 3")
	
	// Calculate expected average duration
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	expectedAvg := total / time.Duration(len(durations))
	
	avgDuration := metrics.GetAverageHttpDuration(method, path)
	assert.Equal(t, expectedAvg, avgDuration, "Average duration should match expected")
}

func TestMetrics_Reset(t *testing.T) {
	metrics := utils.GetMetrics()
	
	// Add some data
	metrics.RecordHttpRequest("GET", "/test", 100*time.Millisecond, 200)
	metrics.RecordDatabaseQuery(50 * time.Millisecond)
	metrics.RecordCacheHit()
	metrics.SetDatabaseConnections(5)
	
	// Reset metrics
	metrics.Reset()
	
	// Verify all metrics are cleared
	assert.Empty(t, metrics.HttpRequestsTotal, "HTTP requests should be empty after reset")
	assert.Empty(t, metrics.HttpResponseStatus, "HTTP response status should be empty after reset")
	assert.Equal(t, int64(0), metrics.DatabaseQueryTotal, "Database query total should be 0 after reset")
	assert.Equal(t, int64(0), metrics.RedisCacheHits, "Cache hits should be 0 after reset")
	assert.Equal(t, int64(0), metrics.DatabaseConnectionsActive, "Database connections should be 0 after reset")
}

func TestMetrics_EdgeCases(t *testing.T) {
	metrics := utils.GetMetrics()
	metrics.Reset()
	
	// Test zero duration case
	avgDuration := metrics.GetAverageHttpDuration("GET", "/nonexistent")
	assert.Equal(t, time.Duration(0), avgDuration, "Average duration for non-existent endpoint should be 0")
	
	avgDbDuration := metrics.GetAverageDatabaseQueryDuration()
	assert.Equal(t, time.Duration(0), avgDbDuration, "Average database duration with no queries should be 0")
	
	hitRatio := metrics.GetCacheHitRatio()
	assert.Equal(t, 0.0, hitRatio, "Cache hit ratio with no cache operations should be 0")
}