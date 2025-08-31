package response

import (
	"net/http"
	"time"

	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
)

// ErrorCode represents standardized error codes
type ErrorCode string

const (
	// Client errors (4xx)
	ErrCodeBadRequest        ErrorCode = "BAD_REQUEST"
	ErrCodeUnauthorized      ErrorCode = "UNAUTHORIZED"
	ErrCodeForbidden         ErrorCode = "FORBIDDEN"
	ErrCodeNotFound          ErrorCode = "NOT_FOUND"
	ErrCodeConflict          ErrorCode = "CONFLICT"
	ErrCodeValidationFailed  ErrorCode = "VALIDATION_FAILED"
	ErrCodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"

	// Server errors (5xx)
	ErrCodeInternalServer     ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeServiceUnavailable ErrorCode = "SERVICE_UNAVAILABLE"
	ErrCodeDatabaseError      ErrorCode = "DATABASE_ERROR"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Success   bool        `json:"success"`
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"requestId,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
	Path      string      `json:"path,omitempty"`
}

// ErrorDetail contains detailed error information
type ErrorDetail struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
	Field   string      `json:"field,omitempty"`
}

// SuccessResponse represents a standardized success response
type SuccessResponse struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"requestId,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// WriteError writes a standardized error response
func WriteError(r *ghttp.Request, statusCode int, errorCode ErrorCode, message string, details ...interface{}) {
	errorDetail := ErrorDetail{
		Code:    errorCode,
		Message: message,
	}

	if len(details) > 0 {
		errorDetail.Details = details[0]
	}

	response := ErrorResponse{
		Success:   false,
		Error:     errorDetail,
		RequestID: getRequestID(r),
		Timestamp: time.Now(),
		Path:      r.URL.Path,
	}

	r.Response.Status = statusCode
	r.Response.WriteJson(response)
}

// WriteSuccess writes a standardized success response
func WriteSuccess(r *ghttp.Request, message string, data interface{}) {
	response := SuccessResponse{
		Success:   true,
		Message:   message,
		Data:      data,
		RequestID: getRequestID(r),
		Timestamp: time.Now(),
	}

	r.Response.WriteJson(response)
}

// WriteValidationError writes a validation error response
func WriteValidationError(r *ghttp.Request, message string, field string, details interface{}) {
	errorDetail := ErrorDetail{
		Code:    ErrCodeValidationFailed,
		Message: message,
		Field:   field,
		Details: details,
	}

	response := ErrorResponse{
		Success:   false,
		Error:     errorDetail,
		RequestID: getRequestID(r),
		Timestamp: time.Now(),
		Path:      r.URL.Path,
	}

	r.Response.Status = http.StatusBadRequest
	r.Response.WriteJson(response)
}

// WriteTenantNotFound writes a tenant not found error
func WriteTenantNotFound(r *ghttp.Request, tenantID string) {
	WriteError(r, http.StatusNotFound, ErrCodeNotFound,
		"Tenant not found", map[string]interface{}{
			"tenantId": tenantID,
		})
}

// WriteTenantConflict writes a tenant conflict error (e.g., duplicate code)
func WriteTenantConflict(r *ghttp.Request, message string, details interface{}) {
	WriteError(r, http.StatusConflict, ErrCodeConflict, message, details)
}

// WriteUnauthorized writes an unauthorized error
func WriteUnauthorized(r *ghttp.Request, message string) {
	WriteError(r, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

// WriteForbidden writes a forbidden error
func WriteForbidden(r *ghttp.Request, message string) {
	WriteError(r, http.StatusForbidden, ErrCodeForbidden, message)
}

// WriteInternalServerError writes an internal server error
func WriteInternalServerError(r *ghttp.Request, message string, details ...interface{}) {
	WriteError(r, http.StatusInternalServerError, ErrCodeInternalServer, message, details...)
}

// WriteRateLimitExceeded writes a rate limit exceeded error
func WriteRateLimitExceeded(r *ghttp.Request, message string) {
	WriteError(r, http.StatusTooManyRequests, ErrCodeRateLimitExceeded, message)
}

// Helper functions

func getRequestID(r *ghttp.Request) string {
	// Try to get request ID from context or header
	if requestID := r.Header.Get("X-Request-ID"); requestID != "" {
		return requestID
	}

	// Generate a simple request ID if not found
	return generateSimpleID()
}

func generateSimpleID() string {
	// Use GoFrame's GUID generation for proper uniqueness
	return "resp_" + guid.S()
}
