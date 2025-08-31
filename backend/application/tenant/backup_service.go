package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"

	"github.com/gofromzero/project_temp/backend/infr/database"
	"github.com/gofromzero/project_temp/backend/pkg/middleware"
)

var (
	// ErrInvalidBackupFormat indicates an unsupported backup format
	ErrInvalidBackupFormat = errors.New("invalid backup format")
	// ErrBackupNotFound indicates backup file was not found
	ErrBackupNotFound = errors.New("backup not found")
	// ErrInvalidTenantForRestore indicates tenant mismatch during restore
	ErrInvalidTenantForRestore = errors.New("invalid tenant for restore operation")
	// ErrBackupIntegrityCheck indicates backup integrity validation failed
	ErrBackupIntegrityCheck = errors.New("backup integrity check failed")
)

// BackupFormat represents supported backup formats
type BackupFormat string

const (
	BackupFormatJSON BackupFormat = "json"
	BackupFormatSQL  BackupFormat = "sql"
)

// BackupType represents backup types
type BackupType string

const (
	BackupTypeFull        BackupType = "full"
	BackupTypeIncremental BackupType = "incremental"
)

// TenantBackupService handles tenant data backup and restore operations
type TenantBackupService struct {
	conn         *database.Connection
	tenantFilter *middleware.TenantFilter
}

// TenantBackupMetadata represents backup metadata
type TenantBackupMetadata struct {
	ID                string                 `json:"id"`
	TenantID          string                 `json:"tenant_id"`
	BackupType        BackupType             `json:"backup_type"`
	Format            BackupFormat           `json:"format"`
	CreatedAt         time.Time              `json:"created_at"`
	Size              int64                  `json:"size"`
	RecordCounts      map[string]int         `json:"record_counts"`
	Checksum          string                 `json:"checksum"`
	Version           string                 `json:"version"`
	AdditionalInfo    map[string]interface{} `json:"additional_info"`
	LastIncrementalAt *time.Time             `json:"last_incremental_at,omitempty"`
}

// TenantBackupData represents the complete backup data structure
type TenantBackupData struct {
	Metadata TenantBackupMetadata   `json:"metadata"`
	Data     map[string]interface{} `json:"data"`
}

// TenantRestoreOptions represents options for restore operations
type TenantRestoreOptions struct {
	TargetTenantID     string   `json:"target_tenant_id"`
	OverwriteExisting  bool     `json:"overwrite_existing"`
	ValidateIntegrity  bool     `json:"validate_integrity"`
	ConflictResolution string   `json:"conflict_resolution"` // "skip", "overwrite", "merge"
	TableFilters       []string `json:"table_filters"`       // Specific tables to restore
}

// NewTenantBackupService creates a new tenant backup service
func NewTenantBackupService() *TenantBackupService {
	return &TenantBackupService{
		conn:         database.NewConnection(),
		tenantFilter: middleware.NewTenantFilter(),
	}
}

// CreateFullBackup creates a full backup of all tenant data
func (s *TenantBackupService) CreateFullBackup(ctx context.Context, tenantID string, format BackupFormat) (*TenantBackupData, error) {
	// Validate backup format
	if !s.isValidFormat(format) {
		return nil, ErrInvalidBackupFormat
	}

	// Validate tenant access - only system admin or tenant owner can backup
	if err := s.validateBackupAccess(ctx, tenantID); err != nil {
		return nil, err
	}

	// Get all tenant-related tables
	tables := s.getTenantTables()

	// Create backup metadata
	metadata := TenantBackupMetadata{
		ID:           s.generateBackupID(tenantID),
		TenantID:     tenantID,
		BackupType:   BackupTypeFull,
		Format:       format,
		CreatedAt:    time.Now(),
		RecordCounts: make(map[string]int),
		Version:      "1.0",
		AdditionalInfo: map[string]interface{}{
			"created_by": s.getContextUser(ctx),
			"source":     "tenant_backup_service",
		},
	}

	// Collect data from all tables
	backupData := make(map[string]interface{})
	totalSize := int64(0)

	for _, table := range tables {
		tableData, recordCount, size, err := s.exportTableData(ctx, tenantID, table, format)
		if err != nil {
			g.Log().Errorf(ctx, "Failed to export table %s for tenant %s: %v", table, tenantID, err)
			continue // Skip tables with errors but continue backup
		}

		backupData[table] = tableData
		metadata.RecordCounts[table] = recordCount
		totalSize += size
	}

	metadata.Size = totalSize
	metadata.Checksum = s.calculateChecksum(backupData)

	backup := &TenantBackupData{
		Metadata: metadata,
		Data:     backupData,
	}

	// Log backup creation
	s.auditBackupOperation(ctx, "CREATE_FULL_BACKUP", tenantID, &metadata)

	return backup, nil
}

// CreateIncrementalBackup creates an incremental backup since the last backup
func (s *TenantBackupService) CreateIncrementalBackup(ctx context.Context, tenantID string, format BackupFormat, since time.Time) (*TenantBackupData, error) {
	// Validate backup format
	if !s.isValidFormat(format) {
		return nil, ErrInvalidBackupFormat
	}

	// Validate tenant access
	if err := s.validateBackupAccess(ctx, tenantID); err != nil {
		return nil, err
	}

	// Get all tenant-related tables
	tables := s.getTenantTables()

	// Create backup metadata
	metadata := TenantBackupMetadata{
		ID:                s.generateBackupID(tenantID),
		TenantID:          tenantID,
		BackupType:        BackupTypeIncremental,
		Format:            format,
		CreatedAt:         time.Now(),
		RecordCounts:      make(map[string]int),
		Version:           "1.0",
		LastIncrementalAt: &since,
		AdditionalInfo: map[string]interface{}{
			"created_by":        s.getContextUser(ctx),
			"source":            "tenant_backup_service",
			"incremental_since": since.Format(time.RFC3339),
		},
	}

	// Collect incremental data from all tables
	backupData := make(map[string]interface{})
	totalSize := int64(0)

	for _, table := range tables {
		tableData, recordCount, size, err := s.exportIncrementalTableData(ctx, tenantID, table, format, since)
		if err != nil {
			g.Log().Errorf(ctx, "Failed to export incremental data for table %s: %v", table, err)
			continue
		}

		// Only include tables with changes
		if recordCount > 0 {
			backupData[table] = tableData
			metadata.RecordCounts[table] = recordCount
			totalSize += size
		}
	}

	metadata.Size = totalSize
	metadata.Checksum = s.calculateChecksum(backupData)

	backup := &TenantBackupData{
		Metadata: metadata,
		Data:     backupData,
	}

	// Log backup creation
	s.auditBackupOperation(ctx, "CREATE_INCREMENTAL_BACKUP", tenantID, &metadata)

	return backup, nil
}

// RestoreTenantData restores tenant data from backup
func (s *TenantBackupService) RestoreTenantData(ctx context.Context, backup *TenantBackupData, options TenantRestoreOptions) error {
	// Validate restore access
	if err := s.validateRestoreAccess(ctx, options.TargetTenantID); err != nil {
		return err
	}

	// Validate backup integrity if requested
	if options.ValidateIntegrity {
		if err := s.validateBackupIntegrity(backup); err != nil {
			return fmt.Errorf("backup integrity validation failed: %w", err)
		}
	}

	// Ensure target tenant exists or create it
	if err := s.ensureTargetTenantExists(ctx, options.TargetTenantID); err != nil {
		return fmt.Errorf("failed to prepare target tenant: %w", err)
	}

	// Start transaction for atomicity
	err := s.conn.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// Restore data for each table
		for tableName, tableData := range backup.Data {
			// Apply table filters if specified
			if len(options.TableFilters) > 0 && !s.isTableInFilter(tableName, options.TableFilters) {
				continue
			}

			if err := s.restoreTableData(ctx, tx, options.TargetTenantID, tableName, tableData, options); err != nil {
				return fmt.Errorf("failed to restore table %s: %w", tableName, err)
			}
		}

		return nil
	})

	if err != nil {
		s.auditBackupOperation(ctx, "RESTORE_FAILED", options.TargetTenantID, &backup.Metadata)
		return err
	}

	// Log successful restore
	s.auditBackupOperation(ctx, "RESTORE_SUCCESS", options.TargetTenantID, &backup.Metadata)

	return nil
}

// ExportTenantDataJSON exports tenant data in JSON format
func (s *TenantBackupService) ExportTenantDataJSON(ctx context.Context, tenantID string) (string, error) {
	backup, err := s.CreateFullBackup(ctx, tenantID, BackupFormatJSON)
	if err != nil {
		return "", err
	}

	jsonData, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal backup to JSON: %w", err)
	}

	return string(jsonData), nil
}

// ExportTenantDataSQL exports tenant data in SQL format
func (s *TenantBackupService) ExportTenantDataSQL(ctx context.Context, tenantID string) (string, error) {
	backup, err := s.CreateFullBackup(ctx, tenantID, BackupFormatSQL)
	if err != nil {
		return "", err
	}

	var sqlBuilder strings.Builder

	// Add header
	sqlBuilder.WriteString(fmt.Sprintf("-- Tenant Data Export for Tenant: %s\n", tenantID))
	sqlBuilder.WriteString(fmt.Sprintf("-- Generated at: %s\n", time.Now().Format(time.RFC3339)))
	sqlBuilder.WriteString(fmt.Sprintf("-- Backup ID: %s\n\n", backup.Metadata.ID))

	// Convert data to SQL statements
	for tableName, tableData := range backup.Data {
		if sqlStatements, ok := tableData.([]string); ok {
			sqlBuilder.WriteString(fmt.Sprintf("-- Table: %s\n", tableName))
			for _, statement := range sqlStatements {
				sqlBuilder.WriteString(statement)
				sqlBuilder.WriteString(";\n")
			}
			sqlBuilder.WriteString("\n")
		}
	}

	return sqlBuilder.String(), nil
}

// ImportTenantDataFromJSON imports tenant data from JSON backup
func (s *TenantBackupService) ImportTenantDataFromJSON(ctx context.Context, jsonData string, options TenantRestoreOptions) error {
	var backup TenantBackupData
	if err := json.Unmarshal([]byte(jsonData), &backup); err != nil {
		return fmt.Errorf("failed to parse JSON backup: %w", err)
	}

	return s.RestoreTenantData(ctx, &backup, options)
}

// exportTableData exports data from a specific table
func (s *TenantBackupService) exportTableData(ctx context.Context, tenantID, tableName string, format BackupFormat) (interface{}, int, int64, error) {
	// Create tenant context for data access
	tenantCtx := s.tenantFilter.WithTenantContext(ctx, &tenantID, false)

	// Get tenant-aware model
	model, err := s.conn.GetTenantAwareModel(tenantCtx, tableName)
	if err != nil {
		return nil, 0, 0, err
	}

	// Query all records
	result, err := model.All()
	if err != nil {
		return nil, 0, 0, err
	}

	recordCount := len(result)

	switch format {
	case BackupFormatJSON:
		// Convert to JSON-serializable format
		data := make([]map[string]interface{}, recordCount)
		for i, record := range result {
			data[i] = record.Map()
		}

		// Estimate size
		jsonBytes, _ := json.Marshal(data)
		size := int64(len(jsonBytes))

		return data, recordCount, size, nil

	case BackupFormatSQL:
		// Generate SQL INSERT statements
		var sqlStatements []string

		if recordCount > 0 {
			// Get column names from first record
			firstRecord := result[0].Map()
			var columns []string
			for col := range firstRecord {
				columns = append(columns, col)
			}

			// Generate INSERT statements
			for _, record := range result {
				values := make([]string, len(columns))
				recordMap := record.Map()

				for i, col := range columns {
					val := recordMap[col]
					if val == nil {
						values[i] = "NULL"
					} else {
						values[i] = fmt.Sprintf("'%s'", gconv.String(val))
					}
				}

				sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
					tableName,
					strings.Join(columns, ", "),
					strings.Join(values, ", "))
				sqlStatements = append(sqlStatements, sql)
			}
		}

		// Estimate size
		totalSize := int64(0)
		for _, stmt := range sqlStatements {
			totalSize += int64(len(stmt))
		}

		return sqlStatements, recordCount, totalSize, nil
	}

	return nil, 0, 0, ErrInvalidBackupFormat
}

// exportIncrementalTableData exports incremental data from a specific table
func (s *TenantBackupService) exportIncrementalTableData(ctx context.Context, tenantID, tableName string, format BackupFormat, since time.Time) (interface{}, int, int64, error) {
	// Create tenant context for data access
	tenantCtx := s.tenantFilter.WithTenantContext(ctx, &tenantID, false)

	// Get tenant-aware model
	model, err := s.conn.GetTenantAwareModel(tenantCtx, tableName)
	if err != nil {
		return nil, 0, 0, err
	}

	// Query records modified since the specified time
	// Assume tables have 'updated_at' column for incremental backup
	result, err := model.Where("updated_at > ?", since).All()
	if err != nil {
		return nil, 0, 0, err
	}

	recordCount := len(result)
	if recordCount == 0 {
		return nil, 0, 0, nil
	}

	return s.formatTableData(result, tableName, format)
}

// formatTableData formats table data according to the specified format
func (s *TenantBackupService) formatTableData(result gdb.Result, tableName string, format BackupFormat) (interface{}, int, int64, error) {
	recordCount := len(result)

	switch format {
	case BackupFormatJSON:
		data := make([]map[string]interface{}, recordCount)
		for i, record := range result {
			data[i] = record.Map()
		}

		jsonBytes, _ := json.Marshal(data)
		size := int64(len(jsonBytes))

		return data, recordCount, size, nil

	case BackupFormatSQL:
		var sqlStatements []string

		if recordCount > 0 {
			firstRecord := result[0].Map()
			var columns []string
			for col := range firstRecord {
				columns = append(columns, col)
			}

			for _, record := range result {
				values := make([]string, len(columns))
				recordMap := record.Map()

				for i, col := range columns {
					val := recordMap[col]
					if val == nil {
						values[i] = "NULL"
					} else {
						values[i] = fmt.Sprintf("'%s'", gconv.String(val))
					}
				}

				sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
					tableName,
					strings.Join(columns, ", "),
					strings.Join(values, ", "))
				sqlStatements = append(sqlStatements, sql)
			}
		}

		totalSize := int64(0)
		for _, stmt := range sqlStatements {
			totalSize += int64(len(stmt))
		}

		return sqlStatements, recordCount, totalSize, nil
	}

	return nil, 0, 0, ErrInvalidBackupFormat
}

// restoreTableData restores data to a specific table
func (s *TenantBackupService) restoreTableData(ctx context.Context, tx gdb.TX, tenantID, tableName string, tableData interface{}, options TenantRestoreOptions) error {
	switch data := tableData.(type) {
	case []map[string]interface{}:
		// JSON format data
		return s.restoreJSONTableData(ctx, tx, tenantID, tableName, data, options)
	case []string:
		// SQL format data
		return s.restoreSQLTableData(ctx, tx, data, options)
	default:
		return fmt.Errorf("unsupported table data format for table %s", tableName)
	}
}

// restoreJSONTableData restores JSON format table data
func (s *TenantBackupService) restoreJSONTableData(ctx context.Context, tx gdb.TX, tenantID, tableName string, data []map[string]interface{}, options TenantRestoreOptions) error {
	if len(data) == 0 {
		return nil
	}

	// Prepare data with correct tenant ID
	for _, record := range data {
		record["tenant_id"] = options.TargetTenantID
	}

	// Handle conflict resolution
	if !options.OverwriteExisting {
		// TODO: Implement conflict detection and resolution
		// For now, we'll just insert all records
	}

	// Bulk insert the data
	_, err := tx.Model(tableName).Insert(data)
	return err
}

// restoreSQLTableData restores SQL format table data
func (s *TenantBackupService) restoreSQLTableData(ctx context.Context, tx gdb.TX, sqlStatements []string, options TenantRestoreOptions) error {
	for _, sql := range sqlStatements {
		// Update SQL to use target tenant ID
		updatedSQL := s.updateSQLForTargetTenant(sql, options.TargetTenantID)

		_, err := tx.Exec(updatedSQL)
		if err != nil {
			return fmt.Errorf("failed to execute SQL: %s, error: %w", updatedSQL, err)
		}
	}

	return nil
}

// Helper methods

func (s *TenantBackupService) isValidFormat(format BackupFormat) bool {
	return format == BackupFormatJSON || format == BackupFormatSQL
}

func (s *TenantBackupService) validateBackupAccess(ctx context.Context, tenantID string) error {
	// System admin can backup any tenant
	if s.tenantFilter.IsSystemAdmin(ctx) {
		return nil
	}

	// Tenant users can only backup their own data
	contextTenantID, hasTenantID := s.tenantFilter.GetTenantID(ctx)
	if !hasTenantID {
		return middleware.ErrTenantRequired
	}

	if contextTenantID == nil || *contextTenantID != tenantID {
		return middleware.ErrUnauthorizedTenant
	}

	return nil
}

func (s *TenantBackupService) validateRestoreAccess(ctx context.Context, tenantID string) error {
	// Same logic as backup access validation
	return s.validateBackupAccess(ctx, tenantID)
}

func (s *TenantBackupService) getTenantTables() []string {
	return []string{
		"users",
		"roles",
		"user_roles",
		"role_permissions",
		"audit_logs",
	}
}

func (s *TenantBackupService) generateBackupID(tenantID string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("backup_%s_%d", tenantID, timestamp)
}

func (s *TenantBackupService) calculateChecksum(data map[string]interface{}) string {
	// Simple checksum calculation - in production, use proper hash
	jsonData, _ := json.Marshal(data)
	return fmt.Sprintf("md5_%d", len(jsonData))
}

func (s *TenantBackupService) getContextUser(ctx context.Context) string {
	// Extract user information from context if available
	// For now, return a placeholder
	return "system"
}

func (s *TenantBackupService) validateBackupIntegrity(backup *TenantBackupData) error {
	// Verify checksum
	calculatedChecksum := s.calculateChecksum(backup.Data)
	if calculatedChecksum != backup.Metadata.Checksum {
		return ErrBackupIntegrityCheck
	}

	// Verify record counts
	for tableName, expectedCount := range backup.Metadata.RecordCounts {
		if tableData, exists := backup.Data[tableName]; exists {
			var actualCount int
			switch data := tableData.(type) {
			case []map[string]interface{}:
				actualCount = len(data)
			case []string:
				actualCount = len(data)
			}

			if actualCount != expectedCount {
				return fmt.Errorf("record count mismatch for table %s: expected %d, got %d",
					tableName, expectedCount, actualCount)
			}
		}
	}

	return nil
}

func (s *TenantBackupService) ensureTargetTenantExists(ctx context.Context, tenantID string) error {
	// Check if tenant exists - for now, assume it does
	// In production, verify or create the tenant
	return nil
}

func (s *TenantBackupService) isTableInFilter(tableName string, filters []string) bool {
	for _, filter := range filters {
		if filter == tableName {
			return true
		}
	}
	return false
}

func (s *TenantBackupService) updateSQLForTargetTenant(sql, targetTenantID string) string {
	// Replace tenant_id values in SQL statements
	// This is a simplified implementation
	return strings.ReplaceAll(sql, "tenant_id = '", fmt.Sprintf("tenant_id = '%s", targetTenantID))
}

func (s *TenantBackupService) auditBackupOperation(ctx context.Context, operation, tenantID string, metadata *TenantBackupMetadata) {
	auditData := map[string]interface{}{
		"operation":   operation,
		"tenant_id":   tenantID,
		"backup_id":   metadata.ID,
		"backup_type": metadata.BackupType,
		"format":      metadata.Format,
		"size":        metadata.Size,
		"timestamp":   time.Now().Format(time.RFC3339),
	}

	g.Log().Infof(ctx, "Tenant backup audit: %s", gjson.MustEncodeString(auditData))
}
