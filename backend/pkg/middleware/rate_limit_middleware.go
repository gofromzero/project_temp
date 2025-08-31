package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gofromzero/project_temp/backend/pkg/response"
	"github.com/gogf/gf/v2/net/ghttp"
)

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int           `json:"requestsPerMinute"`
	BurstSize         int           `json:"burstSize"`
	WindowSize        time.Duration `json:"windowSize"`
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 60, // 60 requests per minute
		BurstSize:         10, // Allow burst of 10 requests
		WindowSize:        time.Minute,
	}
}

// ClientInfo tracks rate limiting information for a client
type ClientInfo struct {
	Requests  []time.Time  `json:"requests"`
	LastReset time.Time    `json:"lastReset"`
	Mutex     sync.RWMutex `json:"-"`
}

// RateLimitMiddleware provides rate limiting functionality
type RateLimitMiddleware struct {
	config  RateLimitConfig
	clients map[string]*ClientInfo
	mutex   sync.RWMutex
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(config ...RateLimitConfig) *RateLimitMiddleware {
	cfg := DefaultRateLimitConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	rl := &RateLimitMiddleware{
		config:  cfg,
		clients: make(map[string]*ClientInfo),
	}

	// Start cleanup goroutine to remove old client entries
	go rl.cleanupRoutine()

	return rl
}

// RateLimit middleware function
func (rl *RateLimitMiddleware) RateLimit() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		clientID := rl.getClientID(r)

		if !rl.allowRequest(clientID) {
			response.WriteRateLimitExceeded(r, "Rate limit exceeded. Please try again later.")
			return
		}

		// Add rate limit headers
		rl.addRateLimitHeaders(r, clientID)

		// Continue to next middleware/handler
		r.Middleware.Next()
	}
}

// allowRequest checks if a request should be allowed
func (rl *RateLimitMiddleware) allowRequest(clientID string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	client, exists := rl.clients[clientID]

	if !exists {
		client = &ClientInfo{
			Requests:  []time.Time{now},
			LastReset: now,
		}
		rl.clients[clientID] = client
		return true
	}

	client.Mutex.Lock()
	defer client.Mutex.Unlock()

	// Remove expired requests
	windowStart := now.Add(-rl.config.WindowSize)
	validRequests := make([]time.Time, 0)

	for _, reqTime := range client.Requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	client.Requests = validRequests

	// Check if request should be allowed
	if len(client.Requests) >= rl.config.RequestsPerMinute {
		return false
	}

	// Add current request
	client.Requests = append(client.Requests, now)

	return true
}

// addRateLimitHeaders adds rate limiting headers to response
func (rl *RateLimitMiddleware) addRateLimitHeaders(r *ghttp.Request, clientID string) {
	rl.mutex.RLock()
	client, exists := rl.clients[clientID]
	rl.mutex.RUnlock()

	if !exists {
		return
	}

	client.Mutex.RLock()
	currentRequests := len(client.Requests)
	client.Mutex.RUnlock()

	remaining := rl.config.RequestsPerMinute - currentRequests
	if remaining < 0 {
		remaining = 0
	}

	// Add standard rate limit headers
	r.Response.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RequestsPerMinute))
	r.Response.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
	r.Response.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(rl.config.WindowSize).Unix()))
}

// getClientID extracts client identifier from request
func (rl *RateLimitMiddleware) getClientID(r *ghttp.Request) string {
	// Priority order for client identification:
	// 1. API Key (if present)
	// 2. User ID from JWT (if authenticated)
	// 3. IP Address (fallback)

	// Check for API key
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		return "api_" + apiKey
	}

	// Check for user authentication (simplified)
	if userID := r.Header.Get("X-User-ID"); userID != "" {
		return "user_" + userID
	}

	// Fallback to IP address
	return "ip_" + r.GetClientIp()
}

// cleanupRoutine removes old client entries periodically
func (rl *RateLimitMiddleware) cleanupRoutine() {
	ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes old client entries that haven't been used recently
func (rl *RateLimitMiddleware) cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	cutoff := time.Now().Add(-2 * rl.config.WindowSize) // Keep entries for 2 windows

	for clientID, client := range rl.clients {
		client.Mutex.RLock()
		shouldDelete := len(client.Requests) == 0 ||
			(len(client.Requests) > 0 && client.Requests[len(client.Requests)-1].Before(cutoff))
		client.Mutex.RUnlock()

		if shouldDelete {
			delete(rl.clients, clientID)
		}
	}
}

// GetClientStats returns rate limiting statistics for a client
func (rl *RateLimitMiddleware) GetClientStats(clientID string) map[string]interface{} {
	rl.mutex.RLock()
	client, exists := rl.clients[clientID]
	rl.mutex.RUnlock()

	if !exists {
		return map[string]interface{}{
			"requests":  0,
			"remaining": rl.config.RequestsPerMinute,
			"reset_at":  time.Now().Add(rl.config.WindowSize),
		}
	}

	client.Mutex.RLock()
	defer client.Mutex.RUnlock()

	now := time.Now()
	windowStart := now.Add(-rl.config.WindowSize)

	// Count valid requests in current window
	validRequests := 0
	for _, reqTime := range client.Requests {
		if reqTime.After(windowStart) {
			validRequests++
		}
	}

	remaining := rl.config.RequestsPerMinute - validRequests
	if remaining < 0 {
		remaining = 0
	}

	return map[string]interface{}{
		"requests":     validRequests,
		"remaining":    remaining,
		"reset_at":     now.Add(rl.config.WindowSize),
		"window_start": windowStart,
	}
}
