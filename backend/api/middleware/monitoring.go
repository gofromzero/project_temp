package middleware

import (
	"time"

	"github.com/gofromzero/project_temp/backend/pkg/utils"
	"github.com/gogf/gf/v2/net/ghttp"
)

// MonitoringMiddleware provides performance metrics collection functionality
type MonitoringMiddleware struct{}

// NewMonitoringMiddleware creates a new monitoring middleware instance
func NewMonitoringMiddleware() *MonitoringMiddleware {
	return &MonitoringMiddleware{}
}

// MetricsCollector collects HTTP request performance metrics
func (m *MonitoringMiddleware) MetricsCollector(r *ghttp.Request) {
	// Record start time
	startTime := time.Now()

	// Continue with request processing
	r.Middleware.Next()

	// Calculate processing time
	duration := time.Since(startTime)

	// Get request details
	method := r.Method
	path := r.URL.Path
	statusCode := r.Response.Status

	// Record metrics
	metrics := utils.GetMetrics()
	metrics.RecordHttpRequest(method, path, duration, statusCode)
}
