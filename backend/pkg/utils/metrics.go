package utils

import (
	"fmt"
	"sync"
	"time"
)

// Metrics holds various performance metrics
type Metrics struct {
	mu sync.RWMutex

	// HTTP Metrics
	HttpRequestsTotal   map[string]int64           // method:path -> count
	HttpRequestDuration map[string][]time.Duration // method:path -> durations
	HttpResponseStatus  map[int]int64              // status_code -> count

	// Database Metrics
	DatabaseConnectionsActive int64
	DatabaseQueryTotal        int64
	DatabaseQueryDuration     []time.Duration

	// Redis Metrics
	RedisConnectionsActive int64
	RedisCacheHits         int64
	RedisCacheMisses       int64
}

// Global metrics instance
var globalMetrics = &Metrics{
	HttpRequestsTotal:     make(map[string]int64),
	HttpRequestDuration:   make(map[string][]time.Duration),
	HttpResponseStatus:    make(map[int]int64),
	DatabaseQueryDuration: make([]time.Duration, 0),
}

// GetMetrics returns the global metrics instance
func GetMetrics() *Metrics {
	return globalMetrics
}

// RecordHttpRequest records HTTP request metrics
func (m *Metrics) RecordHttpRequest(method, path string, duration time.Duration, statusCode int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Record request count
	key := fmt.Sprintf("%s:%s", method, path)
	m.HttpRequestsTotal[key]++

	// Record duration
	if m.HttpRequestDuration[key] == nil {
		m.HttpRequestDuration[key] = make([]time.Duration, 0)
	}
	m.HttpRequestDuration[key] = append(m.HttpRequestDuration[key], duration)

	// Keep only last 1000 durations per endpoint to prevent memory issues
	if len(m.HttpRequestDuration[key]) > 1000 {
		m.HttpRequestDuration[key] = m.HttpRequestDuration[key][len(m.HttpRequestDuration[key])-1000:]
	}

	// Record status code
	m.HttpResponseStatus[statusCode]++
}

// RecordDatabaseQuery records database query metrics
func (m *Metrics) RecordDatabaseQuery(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DatabaseQueryTotal++
	m.DatabaseQueryDuration = append(m.DatabaseQueryDuration, duration)

	// Keep only last 1000 durations to prevent memory issues
	if len(m.DatabaseQueryDuration) > 1000 {
		m.DatabaseQueryDuration = m.DatabaseQueryDuration[len(m.DatabaseQueryDuration)-1000:]
	}
}

// SetDatabaseConnections sets the current active database connections
func (m *Metrics) SetDatabaseConnections(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.DatabaseConnectionsActive = count
}

// RecordCacheHit records Redis cache hit
func (m *Metrics) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RedisCacheHits++
}

// RecordCacheMiss records Redis cache miss
func (m *Metrics) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RedisCacheMisses++
}

// SetRedisConnections sets the current active Redis connections
func (m *Metrics) SetRedisConnections(count int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RedisConnectionsActive = count
}

// GetAverageHttpDuration calculates average HTTP request duration for an endpoint
func (m *Metrics) GetAverageHttpDuration(method, path string) time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	key := fmt.Sprintf("%s:%s", method, path)
	durations := m.HttpRequestDuration[key]

	if len(durations) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range durations {
		total += d
	}

	return total / time.Duration(len(durations))
}

// GetAverageDatabaseQueryDuration calculates average database query duration
func (m *Metrics) GetAverageDatabaseQueryDuration() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if len(m.DatabaseQueryDuration) == 0 {
		return 0
	}

	var total time.Duration
	for _, d := range m.DatabaseQueryDuration {
		total += d
	}

	return total / time.Duration(len(m.DatabaseQueryDuration))
}

// GetCacheHitRatio calculates Redis cache hit ratio
func (m *Metrics) GetCacheHitRatio() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	total := m.RedisCacheHits + m.RedisCacheMisses
	if total == 0 {
		return 0
	}

	return float64(m.RedisCacheHits) / float64(total)
}

// Reset clears all metrics (useful for testing)
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.HttpRequestsTotal = make(map[string]int64)
	m.HttpRequestDuration = make(map[string][]time.Duration)
	m.HttpResponseStatus = make(map[int]int64)
	m.DatabaseConnectionsActive = 0
	m.DatabaseQueryTotal = 0
	m.DatabaseQueryDuration = make([]time.Duration, 0)
	m.RedisConnectionsActive = 0
	m.RedisCacheHits = 0
	m.RedisCacheMisses = 0
}
