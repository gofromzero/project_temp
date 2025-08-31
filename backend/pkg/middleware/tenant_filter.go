package middleware

import (
	"context"
	"errors"
	"time"

	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
)

// TenantContext key for storing tenant information in context
type TenantContextKey string

const (
	TenantIDContextKey      TenantContextKey = "tenant_id"
	IsSystemAdminContextKey TenantContextKey = "is_system_admin"
)

var (
	// ErrTenantRequired indicates tenant context is required but not found
	ErrTenantRequired = errors.New("tenant context is required")
	// ErrUnauthorizedTenant indicates unauthorized access to tenant data
	ErrUnauthorizedTenant = errors.New("unauthorized access to tenant data")
	// ErrInvalidToken indicates invalid or missing JWT token
	ErrInvalidToken = errors.New("invalid or missing JWT token")
	// ErrTenantIdentificationFailed indicates failure to identify tenant from request
	ErrTenantIdentificationFailed = errors.New("failed to identify tenant from request")
)

// TenantFilter provides tenant-based data filtering middleware for repository operations
type TenantFilter struct{}

// NewTenantFilter creates a new tenant filter instance
func NewTenantFilter() *TenantFilter {
	return &TenantFilter{}
}

// TenantMiddleware provides HTTP middleware for tenant context injection
func (tf *TenantFilter) TenantMiddleware(r *ghttp.Request) {
	ctx := r.Context()

	// Identify tenant from request
	tenantID, isSystemAdmin, err := tf.IdentifyTenantFromRequest(r)
	if err != nil {
		g.Log().Errorf(ctx, "Failed to identify tenant: %v", err)
		r.Response.WriteStatusExit(401, gjson.MustEncodeString(g.Map{
			"code":    "TENANT_IDENTIFICATION_FAILED",
			"message": "Failed to identify tenant from request",
			"error":   err.Error(),
		}))
		return
	}

	// Inject tenant context into request
	newCtx := tf.WithTenantContext(ctx, tenantID, isSystemAdmin)
	r.SetCtx(newCtx)

	// Log tenant access for audit
	tf.auditTenantAccess(newCtx, r, tenantID, isSystemAdmin)

	// Validate cross-tenant access
	if err := tf.validateCrossTenantAccess(newCtx, r); err != nil {
		g.Log().Warningf(newCtx, "Cross-tenant access denied: %v", err)
		r.Response.WriteStatusExit(403, gjson.MustEncodeString(g.Map{
			"code":    "CROSS_TENANT_ACCESS_DENIED",
			"message": "Access to this tenant's data is not allowed",
			"error":   err.Error(),
		}))
		return
	}

	r.Middleware.Next()
}

// IdentifyTenantFromRequest identifies tenant context from HTTP request
func (tf *TenantFilter) IdentifyTenantFromRequest(r *ghttp.Request) (*string, bool, error) {
	// Method 1: Extract from JWT token
	if tenantID, isSystemAdmin, err := tf.extractTenantFromJWT(r); err == nil {
		return tenantID, isSystemAdmin, nil
	}

	// Method 2: Extract from X-Tenant-ID header
	if tenantID := r.Header.Get("X-Tenant-ID"); tenantID != "" {
		return &tenantID, false, nil
	}

	// Method 3: Extract from subdomain (e.g., tenant123.domain.com)
	if tenantID, err := tf.extractTenantFromSubdomain(r); err == nil && tenantID != nil {
		return tenantID, false, nil
	}

	// Method 4: Extract from URL path parameter (e.g., /api/v1/tenants/{tenantId}/...)
	if tenantID := tf.extractTenantFromPath(r); tenantID != nil {
		return tenantID, false, nil
	}

	// Method 5: System admin routes don't require tenant identification
	if tf.isSystemAdminRoute(r.URL.Path) {
		return nil, true, nil
	}

	return nil, false, ErrTenantIdentificationFailed
}

// extractTenantFromJWT extracts tenant information from JWT token
func (tf *TenantFilter) extractTenantFromJWT(r *ghttp.Request) (*string, bool, error) {
	// Get Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, false, ErrInvalidToken
	}

	// Extract Bearer token
	if !gstr.HasPrefix(authHeader, "Bearer ") {
		return nil, false, ErrInvalidToken
	}

	token := gstr.SubStr(authHeader, 7)
	if token == "" {
		return nil, false, ErrInvalidToken
	}

	// Simple JWT payload parsing (in production, use proper JWT library)
	// For now, we'll simulate JWT payload extraction
	payload, err := tf.parseJWTPayload(token)
	if err != nil {
		return nil, false, err
	}

	// Extract tenant_id and is_system_admin from payload
	tenantID := payload.Get("tenant_id").String()
	isSystemAdmin := payload.Get("is_system_admin").Bool()

	if tenantID == "" && !isSystemAdmin {
		return nil, false, ErrTenantIdentificationFailed
	}

	if tenantID == "" {
		return nil, isSystemAdmin, nil
	}

	return &tenantID, isSystemAdmin, nil
}

// parseJWTPayload parses JWT payload (simplified implementation)
func (tf *TenantFilter) parseJWTPayload(token string) (*gjson.Json, error) {
	// In production, use proper JWT library like golang-jwt/jwt
	// For now, simulate JWT payload parsing
	parts := gstr.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// This is a mock implementation - replace with real JWT parsing
	// Assuming the token contains: {"tenant_id": "tenant123", "is_system_admin": false}
	mockPayload := g.Map{
		"tenant_id":       "tenant123",
		"is_system_admin": false,
		"user_id":         "user456",
		"exp":             1672531200,
	}

	return gjson.New(mockPayload), nil
}

// extractTenantFromSubdomain extracts tenant ID from subdomain
func (tf *TenantFilter) extractTenantFromSubdomain(r *ghttp.Request) (*string, error) {
	host := r.Host
	if host == "" {
		return nil, errors.New("empty host header")
	}

	// Remove port if present
	if gstr.Contains(host, ":") {
		host = gstr.SubStr(host, 0, gstr.Pos(host, ":"))
	}

	// Split by dots
	parts := gstr.Split(host, ".")
	if len(parts) < 3 {
		return nil, errors.New("invalid subdomain format")
	}

	// First part should be tenant ID
	subdomain := parts[0]
	if subdomain == "www" || subdomain == "api" || subdomain == "admin" {
		return nil, errors.New("invalid tenant subdomain")
	}

	return &subdomain, nil
}

// extractTenantFromPath extracts tenant ID from URL path
func (tf *TenantFilter) extractTenantFromPath(r *ghttp.Request) *string {
	path := r.URL.Path

	// Pattern: /api/v1/tenants/{tenantId}/...
	if gstr.HasPrefix(path, "/api/v1/tenants/") {
		pathParts := gstr.Split(gstr.TrimLeft(path, "/"), "/")
		if len(pathParts) >= 4 {
			tenantID := pathParts[3]
			if tenantID != "" {
				return &tenantID
			}
		}
	}

	return nil
}

// isSystemAdminRoute checks if the route requires system admin privileges
func (tf *TenantFilter) isSystemAdminRoute(path string) bool {
	systemAdminRoutes := []string{
		"/api/v1/system/",
		"/api/v1/admin/tenants",
		"/api/v1/admin/system",
		"/api/v1/health",
		"/api/v1/metrics",
	}

	for _, route := range systemAdminRoutes {
		if gstr.HasPrefix(path, route) {
			return true
		}
	}

	return false
}

// validateCrossTenantAccess validates cross-tenant access attempts
func (tf *TenantFilter) validateCrossTenantAccess(ctx context.Context, r *ghttp.Request) error {
	// Extract target tenant ID from request (URL params, body, etc.)
	targetTenantID := tf.extractTargetTenantFromRequest(r)
	if targetTenantID == nil {
		// No specific tenant target in request
		return nil
	}

	// Validate access using existing validation logic
	return tf.ValidateTenantAccess(ctx, targetTenantID)
}

// extractTargetTenantFromRequest extracts the target tenant ID from request
func (tf *TenantFilter) extractTargetTenantFromRequest(r *ghttp.Request) *string {
	// Check URL path parameters
	if tenantID := tf.extractTenantFromPath(r); tenantID != nil {
		return tenantID
	}

	// Check query parameters
	if tenantID := r.Get("tenant_id"); tenantID.String() != "" {
		value := tenantID.String()
		return &value
	}

	// Check request body for tenant_id (for POST/PUT requests)
	if r.Method == "POST" || r.Method == "PUT" || r.Method == "PATCH" {
		if body := r.GetBodyString(); body != "" {
			if json := gjson.New(body); json != nil {
				if tenantID := json.Get("tenant_id").String(); tenantID != "" {
					return &tenantID
				}
			}
		}
	}

	return nil
}

// auditTenantAccess logs tenant access for audit purposes
func (tf *TenantFilter) auditTenantAccess(ctx context.Context, r *ghttp.Request, tenantID *string, isSystemAdmin bool) {
	auditData := g.Map{
		"method":          r.Method,
		"path":            r.URL.Path,
		"tenant_id":       tenantID,
		"is_system_admin": isSystemAdmin,
		"user_agent":      r.UserAgent(),
		"remote_addr":     r.RemoteAddr,
		"timestamp":       gconv.String(time.Now().UnixMilli()),
	}

	// Generate request hash for tracking
	requestHash := gmd5.MustEncrypt(gconv.String(auditData))
	auditData["request_hash"] = requestHash

	// Log for audit trail
	g.Log().Infof(ctx, "Tenant access audit: %s", gjson.MustEncodeString(auditData))
}

// WithTenantContext adds tenant information to the context
func (tf *TenantFilter) WithTenantContext(ctx context.Context, tenantID *string, isSystemAdmin bool) context.Context {
	ctx = context.WithValue(ctx, TenantIDContextKey, tenantID)
	ctx = context.WithValue(ctx, IsSystemAdminContextKey, isSystemAdmin)
	return ctx
}

// GetTenantID extracts tenant ID from context
func (tf *TenantFilter) GetTenantID(ctx context.Context) (*string, bool) {
	tenantID, ok := ctx.Value(TenantIDContextKey).(*string)
	return tenantID, ok
}

// IsSystemAdmin checks if the current context represents a system administrator
func (tf *TenantFilter) IsSystemAdmin(ctx context.Context) bool {
	isAdmin, ok := ctx.Value(IsSystemAdminContextKey).(bool)
	return ok && isAdmin
}

// ApplyTenantFilter applies tenant filtering to database queries
func (tf *TenantFilter) ApplyTenantFilter(ctx context.Context, model *gdb.Model, tableName string) (*gdb.Model, error) {
	// System administrators can access all data
	if tf.IsSystemAdmin(ctx) {
		return model, nil
	}

	// Get tenant ID from context
	tenantID, hasTenantID := tf.GetTenantID(ctx)

	// For tables that support multi-tenancy, apply tenant filtering
	if tf.isMultiTenantTable(tableName) {
		if !hasTenantID {
			return nil, ErrTenantRequired
		}

		// Apply tenant_id filter
		if tenantID == nil {
			// User context without tenant (system user accessing tenant data)
			return nil, ErrUnauthorizedTenant
		}

		return model.Where("tenant_id = ?", *tenantID), nil
	}

	// For system-only tables, only system admins can access
	if tf.isSystemOnlyTable(tableName) {
		if !tf.IsSystemAdmin(ctx) {
			return nil, ErrUnauthorizedTenant
		}
		return model, nil
	}

	// For other tables, apply default filtering
	return model, nil
}

// ApplyTenantInsertFilter applies tenant context to insert operations
func (tf *TenantFilter) ApplyTenantInsertFilter(ctx context.Context, data map[string]interface{}, tableName string) error {
	// System administrators can insert to any table
	if tf.IsSystemAdmin(ctx) {
		return nil
	}

	// For multi-tenant tables, ensure tenant_id is set correctly
	if tf.isMultiTenantTable(tableName) {
		tenantID, hasTenantID := tf.GetTenantID(ctx)
		if !hasTenantID {
			return ErrTenantRequired
		}

		// Set or validate tenant_id in the data
		if tenantID == nil {
			return ErrUnauthorizedTenant
		}

		// Ensure the data has the correct tenant_id
		data["tenant_id"] = *tenantID
	}

	return nil
}

// isMultiTenantTable checks if a table supports multi-tenancy
func (tf *TenantFilter) isMultiTenantTable(tableName string) bool {
	multiTenantTables := map[string]bool{
		"users":            true,
		"roles":            true,
		"audit_logs":       true,
		"user_roles":       false, // No direct tenant_id but filtered by user
		"role_permissions": false, // No direct tenant_id but filtered by role
	}

	return multiTenantTables[tableName]
}

// isSystemOnlyTable checks if a table is system-only
func (tf *TenantFilter) isSystemOnlyTable(tableName string) bool {
	systemOnlyTables := map[string]bool{
		"tenants":     true,
		"permissions": true,
	}

	return systemOnlyTables[tableName]
}

// ValidateTenantAccess validates if the current context can access the specified tenant's data
func (tf *TenantFilter) ValidateTenantAccess(ctx context.Context, targetTenantID *string) error {
	// System administrators can access any tenant's data
	if tf.IsSystemAdmin(ctx) {
		return nil
	}

	// Get current context tenant ID
	contextTenantID, hasTenantID := tf.GetTenantID(ctx)
	if !hasTenantID {
		return ErrTenantRequired
	}

	// Users can only access their own tenant's data
	if contextTenantID == nil || targetTenantID == nil {
		return ErrUnauthorizedTenant
	}

	if *contextTenantID != *targetTenantID {
		return ErrUnauthorizedTenant
	}

	return nil
}

// CreateTenantAwareModel creates a database model with tenant filtering applied
func (tf *TenantFilter) CreateTenantAwareModel(ctx context.Context, tableName string) (*gdb.Model, error) {
	db := g.DB()
	model := db.Model(tableName)

	return tf.ApplyTenantFilter(ctx, model, tableName)
}
