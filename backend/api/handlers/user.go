package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/gofromzero/project_temp/backend/api/middleware"
	"github.com/gofromzero/project_temp/backend/application/user"
	domainUser "github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	UserService *user.UserService // Exported for testing
}

// NewUserHandler creates a new user handler
func NewUserHandler() *UserHandler {
	return &UserHandler{
		UserService: user.NewUserService(),
	}
}

// NewUserHandlerWithService creates a new user handler with custom service (for testing)
func NewUserHandlerWithService(service *user.UserService) *UserHandler {
	return &UserHandler{
		UserService: service,
	}
}

// CreateUser handles POST /users requests
func (h *UserHandler) CreateUser(r *ghttp.Request) {
	ctx := r.Context()

	var req user.CreateUserRequest

	// Parse request body
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Get tenant context from middleware
	tenantID := h.getTenantIDFromContext(ctx)
	
	// For tenant admins, force their tenant ID; for system admins, allow specifying tenant
	if !h.isSystemAdmin(ctx) {
		req.TenantID = tenantID
	}

	// Validate request
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Validation failed: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Create user
	response, err := h.UserService.CreateUser(ctx, req)
	if err != nil {
		status, message := h.handleError(err)
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   message,
		})
		r.Response.Status = status
		return
	}

	// Return success response
	r.Response.WriteJson(g.Map{
		"success": true,
		"message": "User created successfully",
		"data":    response,
	})
	r.Response.Status = http.StatusCreated
}

// ListUsers handles GET /users requests
func (h *UserHandler) ListUsers(r *ghttp.Request) {
	ctx := r.Context()

	var req user.ListUsersRequest

	// Parse query parameters
	if page := r.Get("page").Int(); page > 0 {
		req.Page = page
	} else {
		req.Page = 1
	}

	if limit := r.Get("limit").Int(); limit > 0 {
		req.Limit = limit
	} else {
		req.Limit = 10
	}

	req.Status = r.Get("status").String()

	// Get tenant context
	tenantID := h.getTenantIDFromContext(ctx)
	
	// For tenant admins, force their tenant ID; for system admins, allow querying all or specific tenant
	if !h.isSystemAdmin(ctx) {
		req.TenantID = tenantID
	} else {
		// System admin can specify tenant ID or query all
		if specifiedTenant := r.Get("tenantId").String(); specifiedTenant != "" {
			req.TenantID = &specifiedTenant
		}
	}

	// Validate request
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Validation failed: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Get users
	response, err := h.UserService.GetUsers(ctx, req)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to retrieve users: " + err.Error(),
		})
		r.Response.Status = http.StatusInternalServerError
		return
	}

	// Return success response
	r.Response.WriteJson(g.Map{
		"success":    true,
		"data":       response.Users,
		"pagination": response.Pagination,
	})
}

// GetUser handles GET /users/{id} requests
func (h *UserHandler) GetUser(r *ghttp.Request) {
	ctx := r.Context()

	userID := r.Get("id").String()
	if userID == "" {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "User ID is required",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Get tenant context for boundary validation
	tenantID := h.getTenantIDFromContext(ctx)
	if !h.isSystemAdmin(ctx) {
		// Tenant admin - use their tenant ID for boundary checking
	} else {
		// System admin - no tenant boundary restrictions
		tenantID = nil
	}

	// Get user
	userResponse, err := h.UserService.GetUserByID(ctx, userID, tenantID)
	if err != nil {
		status, message := h.handleError(err)
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   message,
		})
		r.Response.Status = status
		return
	}

	r.Response.WriteJson(g.Map{
		"success": true,
		"data":    userResponse,
	})
}

// UpdateUser handles PUT /users/{id} requests
func (h *UserHandler) UpdateUser(r *ghttp.Request) {
	ctx := r.Context()

	userID := r.Get("id").String()
	if userID == "" {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "User ID is required",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	var req user.UpdateUserRequest

	// Parse request body
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Validate request
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Validation failed: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Get tenant context for boundary validation
	tenantID := h.getTenantIDFromContext(ctx)
	if !h.isSystemAdmin(ctx) {
		// Tenant admin - use their tenant ID for boundary checking
	} else {
		// System admin - no tenant boundary restrictions
		tenantID = nil
	}

	// Update user
	userResponse, err := h.UserService.UpdateUser(ctx, userID, req, tenantID)
	if err != nil {
		status, message := h.handleError(err)
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   message,
		})
		r.Response.Status = status
		return
	}

	r.Response.WriteJson(g.Map{
		"success": true,
		"message": "User updated successfully",
		"data":    userResponse,
	})
}

// DeleteUser handles DELETE /users/{id} requests
func (h *UserHandler) DeleteUser(r *ghttp.Request) {
	ctx := r.Context()

	userID := r.Get("id").String()
	if userID == "" {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "User ID is required",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Get tenant context for boundary validation
	tenantID := h.getTenantIDFromContext(ctx)
	if !h.isSystemAdmin(ctx) {
		// Tenant admin - use their tenant ID for boundary checking
	} else {
		// System admin - no tenant boundary restrictions
		tenantID = nil
	}

	// Delete user
	err := h.UserService.DeleteUser(ctx, userID, tenantID)
	if err != nil {
		status, message := h.handleError(err)
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   message,
		})
		r.Response.Status = status
		return
	}

	r.Response.WriteJson(g.Map{
		"success": true,
		"message": "User deleted successfully",
	})
	r.Response.Status = http.StatusOK
}

// Helper methods

// getTenantIDFromContext extracts tenant ID from authenticated user context
func (h *UserHandler) getTenantIDFromContext(ctx context.Context) *string {
	// Get authenticated tenant from context
	if tenant, ok := middleware.GetAuthenticatedTenant(ctx); ok {
		return &tenant.ID
	}

	// Get authenticated user from context (system admin case)
	if user, ok := middleware.GetAuthenticatedUser(ctx); ok {
		return user.TenantID
	}

	return nil
}

// isSystemAdmin checks if the current user is a system administrator
func (h *UserHandler) isSystemAdmin(ctx context.Context) bool {
	// Get authenticated user from context
	if user, ok := middleware.GetAuthenticatedUser(ctx); ok {
		return user.IsSystemAdmin()
	}
	return false
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && findSubstring(s, substr)))
}

// findSubstring simple substring search
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// handleError processes various error types and returns appropriate HTTP status and message
func (h *UserHandler) handleError(err error) (int, string) {
	// Check if it's a domain validation error
	var validationErr *domainUser.ValidationError
	if errors.As(err, &validationErr) {
		return http.StatusBadRequest, validationErr.Message
	}

	// Handle specific error messages
	errMsg := err.Error()
	switch {
	case errMsg == "username already exists in this tenant" ||
		 errMsg == "email already exists in this tenant":
		return http.StatusConflict, errMsg
	case contains(errMsg, "failed to get user"):
		return http.StatusNotFound, "用户不存在"
	case errMsg == "access denied: user not in your tenant":
		return http.StatusForbidden, "访问被拒绝：用户不在您的租户范围内"
	case contains(errMsg, "validation") || contains(errMsg, "invalid"):
		return http.StatusBadRequest, errMsg
	default:
		return http.StatusInternalServerError, "服务器内部错误"
	}
}