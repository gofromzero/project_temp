package handlers

import (
	"context"
	"net/http"

	"github.com/gofromzero/project_temp/backend/application/auth"
	"github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/gofromzero/project_temp/backend/pkg/utils"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AuthHandler handles authentication related HTTP requests
type AuthHandler struct {
	authService *auth.AuthService
}

// NewAuthHandler creates a new auth handler instance
func NewAuthHandler(userRepo user.UserRepository, tenantRepo tenant.TenantRepository) (*AuthHandler, error) {
	authService, err := auth.NewAuthService(userRepo, tenantRepo)
	if err != nil {
		return nil, err
	}

	return &AuthHandler{
		authService: authService,
	}, nil
}

// LoginRequest represents the login request structure
type LoginRequest struct {
	Email      string  `json:"email" v:"required|email"`
	Password   string  `json:"password" v:"required|min:6"`
	TenantCode *string `json:"tenantCode,omitempty"`
}

// RefreshTokenRequest represents the refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" v:"required"`
}

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	Email      string  `json:"email" v:"required|email"`
	Password   string  `json:"password" v:"required|min:6"`
	Username   string  `json:"username" v:"required|min:3|max:50"`
	FirstName  string  `json:"firstName" v:"required|min:1|max:100"`
	LastName   string  `json:"lastName" v:"required|min:1|max:100"`
	Phone      *string `json:"phone,omitempty" v:"phone"`
	TenantCode *string `json:"tenantCode,omitempty"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(r *ghttp.Request) {
	ctx := r.Context()

	// Parse request body
	var req LoginRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Validate input
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   "Validation failed: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Convert to service request
	serviceReq := &auth.LoginRequest{
		Email:      req.Email,
		Password:   req.Password,
		TenantCode: req.TenantCode,
	}

	// Attempt login
	response, err := h.authService.Login(ctx, serviceReq)
	if err != nil {
		// Map service errors to HTTP status codes
		status := http.StatusInternalServerError
		message := "Authentication failed"

		switch err {
		case auth.ErrInvalidCredentials:
			status = http.StatusUnauthorized
			message = "Invalid email or password"
		case auth.ErrUserLocked:
			status = http.StatusForbidden
			message = "Account is locked"
		case auth.ErrUserInactive:
			status = http.StatusForbidden
			message = "Account is inactive"
		case auth.ErrTenantNotFound:
			status = http.StatusNotFound
			message = "Tenant not found"
		case auth.ErrTenantSuspended:
			status = http.StatusForbidden
			message = "Tenant account is suspended"
		case auth.ErrTooManyAttempts:
			status = http.StatusTooManyRequests
			message = "Too many login attempts. Please try again later"
		}

		g.Log().Error(ctx, "Login failed:", err)
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   message,
		})
		r.Response.Status = status
		return
	}

	// Success response
	r.Response.WriteJson(APIResponse{
		Success: true,
		Message: "Login successful",
		Data:    response,
	})
	r.Response.Status = http.StatusOK
}

// RefreshToken handles POST /auth/refresh
func (h *AuthHandler) RefreshToken(r *ghttp.Request) {
	ctx := r.Context()

	// Parse request body
	var req RefreshTokenRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Validate input
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   "Validation failed: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Refresh token
	response, err := h.authService.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		status := http.StatusUnauthorized
		message := "Invalid refresh token"

		if err == utils.ErrExpiredToken {
			message = "Refresh token expired"
		} else if err == utils.ErrBlacklistedToken {
			message = "Refresh token has been revoked"
		} else if err == auth.ErrUserInactive {
			status = http.StatusForbidden
			message = "User account is inactive"
		}

		g.Log().Warning(ctx, "Token refresh failed:", err)
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   message,
		})
		r.Response.Status = status
		return
	}

	// Success response
	r.Response.WriteJson(APIResponse{
		Success: true,
		Message: "Token refreshed successfully",
		Data:    response,
	})
	r.Response.Status = http.StatusOK
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(r *ghttp.Request) {
	ctx := r.Context()

	// Extract tokens from request
	authHeader := r.Header.Get("Authorization")
	var accessToken string
	if authHeader != "" {
		token, err := utils.ExtractBearerToken(authHeader)
		if err == nil {
			accessToken = token
		}
	}

	// Get refresh token from body (optional)
	var refreshToken string
	if r.GetBody() != nil {
		var body map[string]string
		if err := r.Parse(&body); err == nil {
			refreshToken = body["refreshToken"]
		}
	}

	// Logout (blacklist tokens)
	err := h.authService.Logout(ctx, accessToken, refreshToken)
	if err != nil {
		g.Log().Error(ctx, "Logout failed:", err)
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   "Logout failed",
		})
		r.Response.Status = http.StatusInternalServerError
		return
	}

	// Success response
	r.Response.WriteJson(APIResponse{
		Success: true,
		Message: "Logged out successfully",
	})
	r.Response.Status = http.StatusOK
}

// Register handles POST /auth/register (admin-only)
func (h *AuthHandler) Register(r *ghttp.Request) {
	ctx := r.Context()

	// Parse request body
	var req RegisterRequest
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   "Invalid request format: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Validate input
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   "Validation failed: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Get admin user from context (set by authentication middleware)
	// For now, we'll extract from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   "Authorization required",
		})
		r.Response.Status = http.StatusUnauthorized
		return
	}

	token, err := utils.ExtractBearerToken(authHeader)
	if err != nil {
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   "Invalid authorization header",
		})
		r.Response.Status = http.StatusUnauthorized
		return
	}

	// Validate token and get admin user
	adminUser, err := h.getAuthenticatedUser(ctx, token)
	if err != nil {
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   "Authentication failed",
		})
		r.Response.Status = http.StatusUnauthorized
		return
	}

	// Convert to service request
	serviceReq := &auth.RegisterRequest{
		Email:      req.Email,
		Password:   req.Password,
		Username:   req.Username,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Phone:      req.Phone,
		TenantCode: req.TenantCode,
	}

	// Attempt registration
	response, err := h.authService.Register(ctx, serviceReq, adminUser)
	if err != nil {
		// Map service errors to HTTP status codes
		status := http.StatusInternalServerError
		message := "Registration failed"

		switch err {
		case auth.ErrUnauthorized:
			status = http.StatusForbidden
			message = "Insufficient permissions to create users"
		case auth.ErrTenantNotFound:
			status = http.StatusNotFound
			message = "Tenant not found"
		case auth.ErrEmailAlreadyExists:
			status = http.StatusConflict
			message = "Email already exists"
		case auth.ErrUsernameAlreadyExists:
			status = http.StatusConflict
			message = "Username already exists"
		}

		g.Log().Error(ctx, "Registration failed:", err)
		r.Response.WriteJson(APIResponse{
			Success: false,
			Error:   message,
		})
		r.Response.Status = status
		return
	}

	// Success response
	r.Response.WriteJson(APIResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    response,
	})
	r.Response.Status = http.StatusCreated
}

// Protected endpoint for testing authentication
func (h *AuthHandler) Profile(r *ghttp.Request) {
	// This endpoint requires authentication middleware
	// The middleware should set user info in context

	// For now, return a simple response
	r.Response.WriteJson(APIResponse{
		Success: true,
		Message: "Profile access granted",
		Data: map[string]string{
			"endpoint": "auth/profile",
			"status":   "authenticated",
		},
	})
	r.Response.Status = http.StatusOK
}

// getAuthenticatedUser validates JWT token and returns the authenticated user
func (h *AuthHandler) getAuthenticatedUser(ctx context.Context, token string) (*user.User, error) {
	// Create a temporary JWT manager to validate token
	jwtManager, err := utils.NewJWTManager()
	if err != nil {
		return nil, err
	}

	// Validate token
	payload, err := jwtManager.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	// Get user from database via auth service
	foundUser, _, err := h.authService.GetUserByID(ctx, payload.UserID)
	if err != nil {
		return nil, err
	}

	return foundUser, nil
}
