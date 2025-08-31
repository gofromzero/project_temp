package middleware

import (
	"context"
	"errors"
	
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// TenantContext key for storing tenant information in context
type TenantContextKey string

const (
	TenantIDContextKey     TenantContextKey = "tenant_id"
	IsSystemAdminContextKey TenantContextKey = "is_system_admin"
)

var (
	// ErrTenantRequired indicates tenant context is required but not found
	ErrTenantRequired = errors.New("tenant context is required")
	// ErrUnauthorizedTenant indicates unauthorized access to tenant data
	ErrUnauthorizedTenant = errors.New("unauthorized access to tenant data")
)

// TenantFilter provides tenant-based data filtering middleware for repository operations
type TenantFilter struct{}

// NewTenantFilter creates a new tenant filter instance
func NewTenantFilter() *TenantFilter {
	return &TenantFilter{}
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