package middleware

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
)

// LoggingMiddleware provides HTTP request logging functionality
type LoggingMiddleware struct{}

// NewLoggingMiddleware creates a new logging middleware instance
func NewLoggingMiddleware() *LoggingMiddleware {
	return &LoggingMiddleware{}
}

// RequestLogger logs HTTP requests with structured format
func (m *LoggingMiddleware) RequestLogger(r *ghttp.Request) {
	// Generate unique request ID
	requestID := guid.S()

	// Store request ID in context for potential use by other handlers
	r.SetCtx(context.WithValue(r.Context(), "request_id", requestID))

	// Record start time
	startTime := time.Now()

	// Log incoming request
	g.Log().Info(r.Context(), "HTTP Request Started", g.Map{
		"request_id":  requestID,
		"method":      r.Method,
		"path":        r.URL.Path,
		"query":       r.URL.RawQuery,
		"remote_addr": r.RemoteAddr,
		"user_agent":  r.Header.Get("User-Agent"),
		"timestamp":   startTime.Format(time.RFC3339),
	})

	// Continue with request processing
	r.Middleware.Next()

	// Calculate processing time
	duration := time.Since(startTime)
	statusCode := r.Response.Status

	// Determine log level based on status code
	logLevel := "Info"
	if statusCode >= 400 && statusCode < 500 {
		logLevel = "Warning"
	} else if statusCode >= 500 {
		logLevel = "Error"
	}

	// Log completed request
	logData := g.Map{
		"request_id":    requestID,
		"method":        r.Method,
		"path":          r.URL.Path,
		"status_code":   statusCode,
		"duration_ms":   float64(duration.Nanoseconds()) / 1e6,
		"response_size": r.Response.BufferLength(),
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	switch logLevel {
	case "Error":
		g.Log().Error(r.Context(), "HTTP Request Completed with Error", logData)
	case "Warning":
		g.Log().Warning(r.Context(), "HTTP Request Completed with Warning", logData)
	default:
		g.Log().Info(r.Context(), "HTTP Request Completed", logData)
	}
}

// ErrorLogger logs application errors and exceptions
func (m *LoggingMiddleware) ErrorLogger(r *ghttp.Request) {
	// Catch any panics and log them
	defer func() {
		if err := recover(); err != nil {
			requestID := r.Context().Value("request_id")
			if requestID == nil {
				requestID = "unknown"
			}

			g.Log().Error(r.Context(), "HTTP Request Panic", g.Map{
				"request_id": requestID,
				"method":     r.Method,
				"path":       r.URL.Path,
				"error":      err,
				"timestamp":  time.Now().Format(time.RFC3339),
			})

			// Return 500 error
			r.Response.WriteStatusExit(500, "Internal Server Error")
		}
	}()

	r.Middleware.Next()
}
