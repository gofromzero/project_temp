package middleware

import (
	"regexp"
	"strings"

	"github.com/gofromzero/project_temp/backend/pkg/response"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
)

// ValidationMiddleware provides request validation functionality
type ValidationMiddleware struct {
	// Future: could add custom validation rules here
}

// NewValidationMiddleware creates a new validation middleware
func NewValidationMiddleware() *ValidationMiddleware {
	return &ValidationMiddleware{}
}

// ValidateRequest performs comprehensive request validation
func (v *ValidationMiddleware) ValidateRequest() ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		// Add request ID if not present
		if r.Header.Get("X-Request-ID") == "" {
			r.Header.Set("X-Request-ID", generateRequestID())
		}

		// Validate common request parameters based on path
		if err := v.validatePathParameters(r); err != nil {
			response.WriteValidationError(r, err.Error(), "path", nil)
			return
		}

		// Validate request body size
		if err := v.validateRequestSize(r); err != nil {
			response.WriteValidationError(r, err.Error(), "body", nil)
			return
		}

		// Continue to next middleware/handler
		r.Middleware.Next()
	}
}

// ValidateTenantID validates tenant ID parameter
func (v *ValidationMiddleware) ValidateTenantID(r *ghttp.Request) error {
	tenantID := r.Get("id").String()

	if tenantID == "" {
		return &ValidationError{
			Field:   "id",
			Message: "Tenant ID is required",
		}
	}

	// Validate UUID format (basic check)
	if !v.isValidTenantID(tenantID) {
		return &ValidationError{
			Field:   "id",
			Message: "Invalid tenant ID format",
		}
	}

	return nil
}

// ValidateCreateTenantRequest validates tenant creation request
func (v *ValidationMiddleware) ValidateCreateTenantRequest(data interface{}) []*ValidationError {
	var errors []*ValidationError

	// Use GoFrame's built-in validation
	if err := g.Validator().Data(data).Run(nil); err != nil {
		// Parse GoFrame validation errors - simplified approach
		for field, fieldError := range err.Map() {
			errors = append(errors, &ValidationError{
				Field:   field,
				Message: fieldError.Error(),
			})
		}
	}

	return errors
}

// ValidateUpdateTenantRequest validates tenant update request
func (v *ValidationMiddleware) ValidateUpdateTenantRequest(data interface{}) []*ValidationError {
	var errors []*ValidationError

	// Use GoFrame's built-in validation
	if err := g.Validator().Data(data).Run(nil); err != nil {
		for field, fieldError := range err.Map() {
			errors = append(errors, &ValidationError{
				Field:   field,
				Message: fieldError.Error(),
			})
		}
	}

	return errors
}

// ValidateDeleteTenantRequest validates tenant deletion request
func (v *ValidationMiddleware) ValidateDeleteTenantRequest(data interface{}, tenantID string) []*ValidationError {
	var errors []*ValidationError

	// Use GoFrame's built-in validation for basic rules
	if err := g.Validator().Data(data).Run(nil); err != nil {
		for field, fieldError := range err.Map() {
			errors = append(errors, &ValidationError{
				Field:   field,
				Message: fieldError.Error(),
			})
		}
	}

	// Additional business logic validation for deletion
	// This would be called from the handler with the parsed request

	return errors
}

// ValidatePaginationParams validates pagination parameters
func (v *ValidationMiddleware) ValidatePaginationParams(r *ghttp.Request) []*ValidationError {
	var errors []*ValidationError

	page := r.Get("page").Int()
	limit := r.Get("limit").Int()

	if page < 0 {
		errors = append(errors, &ValidationError{
			Field:   "page",
			Message: "Page must be greater than or equal to 1",
		})
	}

	if limit < 0 {
		errors = append(errors, &ValidationError{
			Field:   "limit",
			Message: "Limit must be greater than 0",
		})
	}

	if limit > 100 {
		errors = append(errors, &ValidationError{
			Field:   "limit",
			Message: "Limit cannot exceed 100",
		})
	}

	return errors
}

// Business logic validation

// ValidateTenantCodeUniqueness validates that tenant code is unique
func (v *ValidationMiddleware) ValidateTenantCodeUniqueness(code string) *ValidationError {
	// This would integrate with the repository to check uniqueness
	// For now, just basic format validation

	if len(code) < 2 {
		return &ValidationError{
			Field:   "code",
			Message: "Tenant code must be at least 2 characters",
		}
	}

	if len(code) > 100 {
		return &ValidationError{
			Field:   "code",
			Message: "Tenant code cannot exceed 100 characters",
		}
	}

	// Check alphanumeric
	if !v.isAlphanumeric(code) {
		return &ValidationError{
			Field:   "code",
			Message: "Tenant code must contain only alphanumeric characters",
		}
	}

	return nil
}

// Helper methods

func (v *ValidationMiddleware) validatePathParameters(r *ghttp.Request) error {
	// Validate tenant ID if present in path
	if strings.Contains(r.URL.Path, "/tenants/") && r.Get("id").String() != "" {
		return v.ValidateTenantID(r)
	}

	return nil
}

func (v *ValidationMiddleware) validateRequestSize(r *ghttp.Request) error {
	// Limit request body size (e.g., 1MB)
	const maxBodySize = 1024 * 1024 // 1MB

	if r.ContentLength > maxBodySize {
		return &ValidationError{
			Field:   "body",
			Message: "Request body too large (max 1MB)",
		}
	}

	return nil
}

func (v *ValidationMiddleware) isValidTenantID(id string) bool {
	// Basic validation for UUID-like format or simple alphanumeric
	// More sophisticated validation could be added
	if len(id) == 0 || len(id) > 100 {
		return false
	}

	// Allow alphanumeric and hyphens (UUID format)
	matched, _ := regexp.MatchString("^[a-zA-Z0-9-_]+$", id)
	return matched
}

func (v *ValidationMiddleware) isAlphanumeric(str string) bool {
	matched, _ := regexp.MatchString("^[a-zA-Z0-9]+$", str)
	return matched
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}

// Helper function
func generateRequestID() string {
	// Use GoFrame's GUID generation for proper uniqueness
	return "req_" + guid.S()
}
