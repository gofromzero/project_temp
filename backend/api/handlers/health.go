package handlers

import (
	"time"

	"github.com/gofromzero/project_temp/backend/pkg/utils"
	"github.com/gogf/gf/v2/net/ghttp"
)

// HealthHandler handles health check endpoints
type HealthHandler struct{}

// NewHealthHandler creates a new health handler instance
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// ComponentStatus represents the status of a system component
type ComponentStatus struct {
	Status       string  `json:"status"`       // "healthy", "degraded", "unhealthy"
	ResponseTime float64 `json:"responseTime"` // Response time in milliseconds
	Message      string  `json:"message,omitempty"`
}

// HealthResponse represents the complete health check response
type HealthResponse struct {
	Status     string                      `json:"status"`    // Overall system status
	Components map[string]ComponentStatus  `json:"components"` // Individual component statuses
	Timestamp  string                     `json:"timestamp"`  // ISO 8601 timestamp
}

// Health handles GET /health requests
func (h *HealthHandler) Health(r *ghttp.Request) {
	response := &HealthResponse{
		Status:     "healthy",
		Components: make(map[string]ComponentStatus),
		Timestamp:  time.Now().Format(time.RFC3339),
	}

	// Test database connection
	dbStart := time.Now()
	if err := utils.TestDatabaseConnection(); err != nil {
		response.Components["database"] = ComponentStatus{
			Status:       "unhealthy",
			ResponseTime: float64(time.Since(dbStart).Nanoseconds()) / 1e6,
			Message:      err.Error(),
		}
		response.Status = "degraded"
	} else {
		response.Components["database"] = ComponentStatus{
			Status:       "healthy",
			ResponseTime: float64(time.Since(dbStart).Nanoseconds()) / 1e6,
		}
	}

	// Test Redis connection
	redisStart := time.Now()
	if err := utils.TestRedisConnection(); err != nil {
		response.Components["redis"] = ComponentStatus{
			Status:       "unhealthy",
			ResponseTime: float64(time.Since(redisStart).Nanoseconds()) / 1e6,
			Message:      err.Error(),
		}
		// Only set to degraded if not already unhealthy
		if response.Status == "healthy" {
			response.Status = "degraded"
		}
	} else {
		response.Components["redis"] = ComponentStatus{
			Status:       "healthy",
			ResponseTime: float64(time.Since(redisStart).Nanoseconds()) / 1e6,
		}
	}

	// Determine overall status
	unhealthyCount := 0
	for _, component := range response.Components {
		if component.Status == "unhealthy" {
			unhealthyCount++
		}
	}

	// If more than half components are unhealthy, system is unhealthy
	totalComponents := len(response.Components)
	if totalComponents > 0 && unhealthyCount > totalComponents/2 {
		response.Status = "unhealthy"
	}

	// Set appropriate HTTP status code
	statusCode := 200
	if response.Status == "unhealthy" {
		statusCode = 503 // Service Unavailable
	} else if response.Status == "degraded" {
		statusCode = 200 // Still OK, but with warnings
	}

	r.Response.Status = statusCode
	r.Response.WriteJson(response)
}