package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gofromzero/project_temp/backend/application/auth"
	"github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/gofromzero/project_temp/backend/pkg/utils"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	jwtManager  *utils.JWTManager
	authService *auth.AuthService
	publicPaths []string
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authService *auth.AuthService, publicPaths []string) (*AuthMiddleware, error) {
	jwtManager, err := utils.NewJWTManager()
	if err != nil {
		return nil, err
	}

	return &AuthMiddleware{
		jwtManager:  jwtManager,
		authService: authService,
		publicPaths: publicPaths,
	}, nil
}

// NewMinimalAuthMiddleware creates a minimal auth middleware for development/testing
func NewMinimalAuthMiddleware(publicPaths []string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager:  nil, // Will be set later when needed
		authService: nil, // Will be set later when needed
		publicPaths: publicPaths,
	}
}

// AuthContextKey is the context key for storing authenticated user
type AuthContextKey string

const (
	UserContextKey   AuthContextKey = "auth_user"
	TokenContextKey  AuthContextKey = "auth_token"
	TenantContextKey AuthContextKey = "auth_tenant"
)

// Authenticate is the main authentication middleware function
func (m *AuthMiddleware) Authenticate(r *ghttp.Request) {
	ctx := r.Context()

	// Check if this is a public path
	if m.isPublicPath(r.URL.Path) {
		r.Middleware.Next()
		return
	}

	// Extract token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		m.sendUnauthorizedResponse(r, "Authorization header missing")
		return
	}

	token, err := utils.ExtractBearerToken(authHeader)
	if err != nil {
		m.sendUnauthorizedResponse(r, "Invalid authorization header format")
		return
	}

	// Validate token
	payload, err := m.jwtManager.ValidateToken(token)
	if err != nil {
		var message string
		switch err {
		case utils.ErrExpiredToken:
			message = "Token has expired"
		case utils.ErrBlacklistedToken:
			message = "Token has been revoked"
		case utils.ErrInvalidToken:
			message = "Invalid token"
		default:
			message = "Authentication failed"
		}
		m.sendUnauthorizedResponse(r, message)
		return
	}

	// Get user from database to ensure user still exists and is active
	authenticatedUser, authenticatedTenant, err := m.authService.GetUserByID(ctx, payload.UserID)
	if err != nil {
		g.Log().Warning(ctx, "Failed to get authenticated user:", err)
		m.sendUnauthorizedResponse(r, "User not found")
		return
	}

	// Check user is still active
	if authenticatedUser.Status != user.StatusActive {
		m.sendUnauthorizedResponse(r, "User account is not active")
		return
	}

	// Check tenant is still active (if applicable)
	if authenticatedTenant != nil && authenticatedTenant.Status != tenant.StatusActive {
		m.sendUnauthorizedResponse(r, "Tenant account is not active")
		return
	}

	// Store authentication info in context
	ctx = context.WithValue(ctx, UserContextKey, authenticatedUser)
	ctx = context.WithValue(ctx, TokenContextKey, payload)
	if authenticatedTenant != nil {
		ctx = context.WithValue(ctx, TenantContextKey, authenticatedTenant)
	}

	// Update request context and continue
	r.SetCtx(ctx)
	r.Middleware.Next()
}

// RequireRole checks if the authenticated user has the required role
func (m *AuthMiddleware) RequireRole(requiredRole string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		ctx := r.Context()

		// Get user from context (should be set by Authenticate middleware)
		user, ok := ctx.Value(UserContextKey).(*user.User)
		if !ok {
			m.sendForbiddenResponse(r, "Authentication required")
			return
		}

		payload, ok := ctx.Value(TokenContextKey).(*utils.TokenPayload)
		if !ok {
			m.sendForbiddenResponse(r, "Invalid authentication context")
			return
		}

		// Check if user has the required role
		hasRole := false
		for _, role := range payload.Roles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		// System admins bypass role checks
		if payload.IsAdmin || user.IsSystemAdmin() {
			hasRole = true
		}

		if !hasRole {
			m.sendForbiddenResponse(r, "Insufficient permissions")
			return
		}

		r.Middleware.Next()
	}
}

// RequireSystemAdmin checks if the authenticated user is a system administrator
func (m *AuthMiddleware) RequireSystemAdmin(r *ghttp.Request) {
	ctx := r.Context()

	user, ok := ctx.Value(UserContextKey).(*user.User)
	if !ok {
		m.sendForbiddenResponse(r, "Authentication required")
		return
	}

	if !user.IsSystemAdmin() {
		m.sendForbiddenResponse(r, "System administrator access required")
		return
	}

	r.Middleware.Next()
}

// GetAuthenticatedUser retrieves the authenticated user from context
func GetAuthenticatedUser(ctx context.Context) (*user.User, bool) {
	user, ok := ctx.Value(UserContextKey).(*user.User)
	return user, ok
}

// GetAuthenticatedTenant retrieves the authenticated tenant from context
func GetAuthenticatedTenant(ctx context.Context) (*tenant.Tenant, bool) {
	tenant, ok := ctx.Value(TenantContextKey).(*tenant.Tenant)
	return tenant, ok
}

// GetTokenPayload retrieves the token payload from context
func GetTokenPayload(ctx context.Context) (*utils.TokenPayload, bool) {
	payload, ok := ctx.Value(TokenContextKey).(*utils.TokenPayload)
	return payload, ok
}

// isPublicPath checks if the request path should skip authentication
func (m *AuthMiddleware) isPublicPath(path string) bool {
	for _, publicPath := range m.publicPaths {
		if strings.HasPrefix(path, publicPath) {
			return true
		}
	}
	return false
}

// sendUnauthorizedResponse sends a 401 Unauthorized response
func (m *AuthMiddleware) sendUnauthorizedResponse(r *ghttp.Request, message string) {
	r.Response.WriteJson(g.Map{
		"success": false,
		"error":   message,
	})
	r.Response.Status = http.StatusUnauthorized
	r.ExitAll()
}

// sendForbiddenResponse sends a 403 Forbidden response
func (m *AuthMiddleware) sendForbiddenResponse(r *ghttp.Request, message string) {
	r.Response.WriteJson(g.Map{
		"success": false,
		"error":   message,
	})
	r.Response.Status = http.StatusForbidden
	r.ExitAll()
}
