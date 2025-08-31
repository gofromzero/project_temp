package database

import (
	"context"
	
	"github.com/gofromzero/project_temp/backend/pkg/middleware"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// Connection provides database connection with tenant context
type Connection struct {
	db           gdb.DB
	tenantFilter *middleware.TenantFilter
}

// NewConnection creates a new database connection with tenant filtering
func NewConnection() *Connection {
	return &Connection{
		db:           g.DB(),
		tenantFilter: middleware.NewTenantFilter(),
	}
}

// GetTenantAwareModel returns a database model with tenant filtering applied
func (c *Connection) GetTenantAwareModel(ctx context.Context, tableName string) (*gdb.Model, error) {
	return c.tenantFilter.CreateTenantAwareModel(ctx, tableName)
}

// WithTenantContext creates a new context with tenant information
func (c *Connection) WithTenantContext(ctx context.Context, tenantID *string, isSystemAdmin bool) context.Context {
	return c.tenantFilter.WithTenantContext(ctx, tenantID, isSystemAdmin)
}

// ValidateTenantAccess validates tenant access permissions
func (c *Connection) ValidateTenantAccess(ctx context.Context, targetTenantID *string) error {
	return c.tenantFilter.ValidateTenantAccess(ctx, targetTenantID)
}

// InsertWithTenantFilter performs insert operation with tenant filtering
func (c *Connection) InsertWithTenantFilter(ctx context.Context, tableName string, data interface{}) (interface{}, error) {
	// Convert data to map for tenant filtering
	var dataMap map[string]interface{}
	
	switch d := data.(type) {
	case map[string]interface{}:
		dataMap = d
	default:
		// For struct data, let the database handle it directly if it's properly tagged
		model, err := c.GetTenantAwareModel(ctx, tableName)
		if err != nil {
			return nil, err
		}
		return model.Insert(data)
	}

	// Apply tenant insert filter
	err := c.tenantFilter.ApplyTenantInsertFilter(ctx, dataMap, tableName)
	if err != nil {
		return nil, err
	}

	// Get tenant-aware model and insert
	model, err := c.GetTenantAwareModel(ctx, tableName)
	if err != nil {
		return nil, err
	}

	return model.Insert(dataMap)
}

// Transaction executes a function within a database transaction with tenant context
func (c *Connection) Transaction(ctx context.Context, fn func(ctx context.Context, tx gdb.TX) error) error {
	return c.db.Transaction(ctx, fn)
}