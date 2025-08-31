package handlers

import (
	"fmt"
	"strings"

	"github.com/gofromzero/project_temp/backend/pkg/utils"
	"github.com/gogf/gf/v2/net/ghttp"
)

// MetricsHandler handles metrics endpoint for Prometheus format
type MetricsHandler struct{}

// NewMetricsHandler creates a new metrics handler instance
func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{}
}

// Metrics returns system performance metrics in Prometheus format
func (h *MetricsHandler) Metrics(r *ghttp.Request) {
	metrics := utils.GetMetrics()
	
	var sb strings.Builder
	
	// HTTP Request Total
	sb.WriteString("# HELP http_requests_total Total number of HTTP requests\n")
	sb.WriteString("# TYPE http_requests_total counter\n")
	for endpoint, count := range metrics.HttpRequestsTotal {
		parts := strings.Split(endpoint, ":")
		if len(parts) == 2 {
			method, path := parts[0], parts[1]
			sb.WriteString(fmt.Sprintf("http_requests_total{method=\"%s\",path=\"%s\"} %d\n", method, path, count))
		}
	}
	
	// HTTP Response Status
	sb.WriteString("# HELP http_response_status_total Total number of HTTP responses by status code\n")
	sb.WriteString("# TYPE http_response_status_total counter\n")
	for status, count := range metrics.HttpResponseStatus {
		sb.WriteString(fmt.Sprintf("http_response_status_total{status=\"%d\"} %d\n", status, count))
	}
	
	// HTTP Request Duration Average
	sb.WriteString("# HELP http_request_duration_average Average HTTP request duration in milliseconds\n")
	sb.WriteString("# TYPE http_request_duration_average gauge\n")
	for endpoint := range metrics.HttpRequestsTotal {
		parts := strings.Split(endpoint, ":")
		if len(parts) == 2 {
			method, path := parts[0], parts[1]
			avgDuration := metrics.GetAverageHttpDuration(method, path)
			durationMs := float64(avgDuration.Nanoseconds()) / 1e6
			sb.WriteString(fmt.Sprintf("http_request_duration_average{method=\"%s\",path=\"%s\"} %.3f\n", method, path, durationMs))
		}
	}
	
	// Database Metrics
	sb.WriteString("# HELP database_connections_active Current number of active database connections\n")
	sb.WriteString("# TYPE database_connections_active gauge\n")
	sb.WriteString(fmt.Sprintf("database_connections_active %d\n", metrics.DatabaseConnectionsActive))
	
	sb.WriteString("# HELP database_queries_total Total number of database queries\n")
	sb.WriteString("# TYPE database_queries_total counter\n")
	sb.WriteString(fmt.Sprintf("database_queries_total %d\n", metrics.DatabaseQueryTotal))
	
	avgDbDuration := metrics.GetAverageDatabaseQueryDuration()
	if avgDbDuration > 0 {
		sb.WriteString("# HELP database_query_duration_average Average database query duration in milliseconds\n")
		sb.WriteString("# TYPE database_query_duration_average gauge\n")
		durationMs := float64(avgDbDuration.Nanoseconds()) / 1e6
		sb.WriteString(fmt.Sprintf("database_query_duration_average %.3f\n", durationMs))
	}
	
	// Redis Metrics
	sb.WriteString("# HELP redis_connections_active Current number of active Redis connections\n")
	sb.WriteString("# TYPE redis_connections_active gauge\n")
	sb.WriteString(fmt.Sprintf("redis_connections_active %d\n", metrics.RedisConnectionsActive))
	
	sb.WriteString("# HELP redis_cache_hits_total Total number of Redis cache hits\n")
	sb.WriteString("# TYPE redis_cache_hits_total counter\n")
	sb.WriteString(fmt.Sprintf("redis_cache_hits_total %d\n", metrics.RedisCacheHits))
	
	sb.WriteString("# HELP redis_cache_misses_total Total number of Redis cache misses\n")
	sb.WriteString("# TYPE redis_cache_misses_total counter\n")
	sb.WriteString(fmt.Sprintf("redis_cache_misses_total %d\n", metrics.RedisCacheMisses))
	
	hitRatio := metrics.GetCacheHitRatio()
	sb.WriteString("# HELP redis_cache_hit_ratio Redis cache hit ratio (0.0 to 1.0)\n")
	sb.WriteString("# TYPE redis_cache_hit_ratio gauge\n")
	sb.WriteString(fmt.Sprintf("redis_cache_hit_ratio %.3f\n", hitRatio))
	
	
	r.Response.Header().Set("Content-Type", "text/plain; charset=utf-8")
	r.Response.Write(sb.String())
}