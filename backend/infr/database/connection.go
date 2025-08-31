package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofromzero/project_temp/backend/pkg/middleware"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

var (
	// ErrTenantDataIntegrityViolation indicates a violation of tenant data integrity
	ErrTenantDataIntegrityViolation = errors.New("tenant data integrity violation detected")
	// ErrTenantIsolationValidationFailed indicates tenant isolation validation failed
	ErrTenantIsolationValidationFailed = errors.New("tenant isolation validation failed")
)

// Connection provides database connection with tenant context
type Connection struct {
	db           gdb.DB
	tenantFilter *middleware.TenantFilter
	auditEnabled bool
}

// TenantDataAccessAudit represents an audit record for tenant data access
type TenantDataAccessAudit struct {
	ID            string    `json:"id"`
	TenantID      *string   `json:"tenant_id"`
	Operation     string    `json:"operation"`
	TableName     string    `json:"table_name"`
	RecordsCount  int       `json:"records_count"`
	IsSystemAdmin bool      `json:"is_system_admin"`
	Timestamp     time.Time `json:"timestamp"`
	Context       string    `json:"context"`
}

// NewConnection creates a new database connection with tenant filtering
func NewConnection() *Connection {
	return &Connection{
		db:           g.DB(),
		tenantFilter: middleware.NewTenantFilter(),
		auditEnabled: true, // Enable audit by default
	}
}

// NewConnectionWithConfig creates a new database connection with configuration
func NewConnectionWithConfig(auditEnabled bool) *Connection {
	return &Connection{
		db:           g.DB(),
		tenantFilter: middleware.NewTenantFilter(),
		auditEnabled: auditEnabled,
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

// QueryWithTenantValidation executes a query with automatic tenant validation
func (c *Connection) QueryWithTenantValidation(ctx context.Context, tableName string, conditions ...interface{}) (gdb.Result, error) {
	// Get tenant-aware model
	model, err := c.GetTenantAwareModel(ctx, tableName)
	if err != nil {
		return nil, err
	}

	// Apply additional conditions if provided
	if len(conditions) > 0 {
		model = model.Where(conditions[0], conditions[1:]...)
	}

	// Execute query
	result, err := model.All()
	if err != nil {
		return nil, err
	}

	// Validate tenant isolation
	if err := c.validateResultTenantIsolation(ctx, result, tableName); err != nil {
		return nil, err
	}

	// Audit data access
	c.auditDataAccess(ctx, "SELECT", tableName, len(result))

	return result, nil
}

// UpdateWithTenantValidation executes an update with tenant validation
func (c *Connection) UpdateWithTenantValidation(ctx context.Context, tableName string, data interface{}, conditions ...interface{}) (int64, error) {
	// Validate tenant access first
	if err := c.validateUpdateAccess(ctx, tableName, conditions...); err != nil {
		return 0, err
	}

	// Get tenant-aware model
	model, err := c.GetTenantAwareModel(ctx, tableName)
	if err != nil {
		return 0, err
	}

	// Apply conditions
	if len(conditions) > 0 {
		model = model.Where(conditions[0], conditions[1:]...)
	}

	// Execute update
	result, err := model.Update(data)
	if err != nil {
		return 0, err
	}

	affectedRows, _ := result.RowsAffected()

	// Audit data access
	c.auditDataAccess(ctx, "UPDATE", tableName, int(affectedRows))

	return affectedRows, nil
}

// DeleteWithTenantValidation executes a delete with tenant validation
func (c *Connection) DeleteWithTenantValidation(ctx context.Context, tableName string, conditions ...interface{}) (int64, error) {
	// Validate tenant access first
	if err := c.validateDeleteAccess(ctx, tableName, conditions...); err != nil {
		return 0, err
	}

	// Get tenant-aware model
	model, err := c.GetTenantAwareModel(ctx, tableName)
	if err != nil {
		return 0, err
	}

	// Apply conditions
	if len(conditions) > 0 {
		model = model.Where(conditions[0], conditions[1:]...)
	}

	// Execute delete
	result, err := model.Delete()
	if err != nil {
		return 0, err
	}

	affectedRows, _ := result.RowsAffected()

	// Audit data access
	c.auditDataAccess(ctx, "DELETE", tableName, int(affectedRows))

	return affectedRows, nil
}

// BulkInsertWithTenantFilter performs bulk insert with tenant filtering
func (c *Connection) BulkInsertWithTenantFilter(ctx context.Context, tableName string, dataList []interface{}) (int64, error) {
	if len(dataList) == 0 {
		return 0, nil
	}

	// Convert to map list for tenant filtering
	var mapDataList []map[string]interface{}
	for _, data := range dataList {
		if mapData, ok := data.(map[string]interface{}); ok {
			// Apply tenant insert filter to each record
			if err := c.tenantFilter.ApplyTenantInsertFilter(ctx, mapData, tableName); err != nil {
				return 0, fmt.Errorf("tenant filter failed for bulk insert: %w", err)
			}
			mapDataList = append(mapDataList, mapData)
		} else {
			return 0, errors.New("bulk insert requires map[string]interface{} data format")
		}
	}

	// Get tenant-aware model
	model, err := c.GetTenantAwareModel(ctx, tableName)
	if err != nil {
		return 0, err
	}

	// Execute bulk insert
	result, err := model.Insert(mapDataList)
	if err != nil {
		return 0, err
	}

	affectedRows, _ := result.RowsAffected()

	// Audit data access
	c.auditDataAccess(ctx, "BULK_INSERT", tableName, int(affectedRows))

	return affectedRows, nil
}

// ValidateTenantDataIntegrity validates that all data in the result belongs to the correct tenant
func (c *Connection) ValidateTenantDataIntegrity(ctx context.Context, tableName string, limit ...int) error {
	// System admins can skip this validation
	if c.tenantFilter.IsSystemAdmin(ctx) {
		return nil
	}

	// Get current tenant ID
	tenantID, hasTenantID := c.tenantFilter.GetTenantID(ctx)
	if !hasTenantID {
		return middleware.ErrTenantRequired
	}

	// Check if table supports multi-tenancy (would need to export this method from middleware)
	// For now, we'll check common multi-tenant tables
	if !c.isMultiTenantTable(tableName) {
		return nil
	}

	// Set default limit
	queryLimit := 1000
	if len(limit) > 0 && limit[0] > 0 {
		queryLimit = limit[0]
	}

	// Query records without tenant filter to check integrity
	rawModel := c.db.Model(tableName).Limit(queryLimit)
	result, err := rawModel.All()
	if err != nil {
		return err
	}

	// Check each record for tenant_id consistency
	var violations []string
	for _, record := range result {
		recordTenantID := record["tenant_id"]
		if recordTenantID == nil {
			violations = append(violations, fmt.Sprintf("Record %v has null tenant_id", record["id"]))
		} else if tenantID != nil && gconv.String(recordTenantID) != *tenantID {
			violations = append(violations, fmt.Sprintf("Record %v belongs to tenant %v, expected %v",
				record["id"], recordTenantID, *tenantID))
		}
	}

	if len(violations) > 0 {
		g.Log().Errorf(ctx, "Tenant data integrity violations in table %s: %v", tableName, violations)
		return ErrTenantDataIntegrityViolation
	}

	return nil
}

// validateResultTenantIsolation validates that query results respect tenant isolation
func (c *Connection) validateResultTenantIsolation(ctx context.Context, result gdb.Result, tableName string) error {
	// System admins can access any data
	if c.tenantFilter.IsSystemAdmin(ctx) {
		return nil
	}

	// Get current tenant ID
	tenantID, hasTenantID := c.tenantFilter.GetTenantID(ctx)
	if !hasTenantID {
		return middleware.ErrTenantRequired
	}

	// Skip validation for non-multi-tenant tables
	if !c.isMultiTenantTable(tableName) {
		return nil
	}

	// Validate each record in the result
	for _, record := range result {
		if recordTenantID, exists := record["tenant_id"]; exists {
			if tenantID != nil && gconv.String(recordTenantID) != *tenantID {
				g.Log().Warningf(ctx, "Tenant isolation violation: record %v from tenant %v accessed by tenant %v",
					record["id"], recordTenantID, *tenantID)
				return ErrTenantIsolationValidationFailed
			}
		}
	}

	return nil
}

// validateUpdateAccess validates tenant access for update operations
func (c *Connection) validateUpdateAccess(ctx context.Context, tableName string, conditions ...interface{}) error {
	// System admins can update any data
	if c.tenantFilter.IsSystemAdmin(ctx) {
		return nil
	}

	// For multi-tenant tables, ensure update conditions include proper tenant filtering
	if c.isMultiTenantTable(tableName) {
		// Query the records that would be affected to ensure they belong to the current tenant
		model, err := c.GetTenantAwareModel(ctx, tableName)
		if err != nil {
			return err
		}

		// Apply the same conditions to check what would be affected
		if len(conditions) > 0 {
			model = model.Where(conditions[0], conditions[1:]...)
		}

		// Get records that would be updated
		result, err := model.All()
		if err != nil {
			return err
		}

		// Validate tenant isolation on the affected records
		return c.validateResultTenantIsolation(ctx, result, tableName)
	}

	return nil
}

// validateDeleteAccess validates tenant access for delete operations
func (c *Connection) validateDeleteAccess(ctx context.Context, tableName string, conditions ...interface{}) error {
	// System admins can delete any data
	if c.tenantFilter.IsSystemAdmin(ctx) {
		return nil
	}

	// For multi-tenant tables, ensure delete conditions include proper tenant filtering
	if c.isMultiTenantTable(tableName) {
		// Query the records that would be affected to ensure they belong to the current tenant
		model, err := c.GetTenantAwareModel(ctx, tableName)
		if err != nil {
			return err
		}

		// Apply the same conditions to check what would be affected
		if len(conditions) > 0 {
			model = model.Where(conditions[0], conditions[1:]...)
		}

		// Get records that would be deleted
		result, err := model.All()
		if err != nil {
			return err
		}

		// Validate tenant isolation on the affected records
		return c.validateResultTenantIsolation(ctx, result, tableName)
	}

	return nil
}

// auditDataAccess logs tenant data access for audit purposes
func (c *Connection) auditDataAccess(ctx context.Context, operation, tableName string, recordsCount int) {
	if !c.auditEnabled {
		return
	}

	tenantID, _ := c.tenantFilter.GetTenantID(ctx)
	isSystemAdmin := c.tenantFilter.IsSystemAdmin(ctx)

	audit := TenantDataAccessAudit{
		ID:            gconv.String(time.Now().UnixNano()),
		TenantID:      tenantID,
		Operation:     operation,
		TableName:     tableName,
		RecordsCount:  recordsCount,
		IsSystemAdmin: isSystemAdmin,
		Timestamp:     time.Now(),
		Context:       fmt.Sprintf("operation=%s,table=%s,records=%d", operation, tableName, recordsCount),
	}

	// Log the audit record
	auditJSON := gjson.MustEncodeString(audit)
	g.Log().Infof(ctx, "Tenant data access audit: %s", auditJSON)

	// TODO: In production, also store audit records in a dedicated audit table
}

// isMultiTenantTable checks if a table supports multi-tenancy
// This duplicates logic from middleware but is needed here for integrity checks
func (c *Connection) isMultiTenantTable(tableName string) bool {
	multiTenantTables := map[string]bool{
		"users":            true,
		"roles":            true,
		"audit_logs":       true,
		"user_roles":       false, // No direct tenant_id but filtered by user
		"role_permissions": false, // No direct tenant_id but filtered by role
	}

	return multiTenantTables[tableName]
}

// GetDB returns the underlying database connection (use with caution)
func (c *Connection) GetDB() gdb.DB {
	return c.db
}

// SetAuditEnabled enables or disables audit logging
func (c *Connection) SetAuditEnabled(enabled bool) {
	c.auditEnabled = enabled
}
