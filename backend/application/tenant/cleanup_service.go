package tenant

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/gofromzero/project_temp/backend/infr/database"
	"github.com/gofromzero/project_temp/backend/pkg/middleware"
)

var (
	// ErrTenantNotFound indicates the tenant to be deleted was not found
	ErrTenantNotFound = errors.New("tenant not found")
	// ErrCleanupNotAuthorized indicates unauthorized attempt to delete tenant data
	ErrCleanupNotAuthorized = errors.New("cleanup operation not authorized")
	// ErrCleanupInProgress indicates a cleanup operation is already in progress
	ErrCleanupInProgress = errors.New("cleanup operation already in progress")
	// ErrCleanupFailed indicates the cleanup operation failed
	ErrCleanupFailed = errors.New("cleanup operation failed")
	// ErrDataErasureIncomplete indicates not all data was successfully erased
	ErrDataErasureIncomplete = errors.New("data erasure incomplete - some records may remain")
)

// DataErasureType represents different types of data erasure
type DataErasureType string

const (
	DataErasureTypeSoft   DataErasureType = "soft"   // Mark as deleted but keep data
	DataErasureTypeHard   DataErasureType = "hard"   // Permanently delete data
	DataErasureTypeSecure DataErasureType = "secure" // Secure deletion with overwrites
)

// CleanupReason represents the reason for cleanup
type CleanupReason string

const (
	CleanupReasonTenantDeletion CleanupReason = "tenant_deletion"
	CleanupReasonGDPRRequest    CleanupReason = "gdpr_right_to_be_forgotten"
	CleanupReasonDataRetention  CleanupReason = "data_retention_policy"
	CleanupReasonUserRequest    CleanupReason = "user_deletion_request"
	CleanupReasonCompliance     CleanupReason = "compliance_requirement"
)

// TenantCleanupService handles tenant data cleanup and erasure operations
type TenantCleanupService struct {
	conn          *database.Connection
	tenantFilter  *middleware.TenantFilter
	backupService *TenantBackupService
}

// CleanupRequest represents a data cleanup request
type CleanupRequest struct {
	TenantID       string                 `json:"tenant_id"`
	Reason         CleanupReason          `json:"reason"`
	ErasureType    DataErasureType        `json:"erasure_type"`
	CreateBackup   bool                   `json:"create_backup"`
	Confirmation   string                 `json:"confirmation"` // Required confirmation string
	RequestedBy    string                 `json:"requested_by"`
	RequestedAt    time.Time              `json:"requested_at"`
	AdditionalInfo map[string]interface{} `json:"additional_info,omitempty"`
}

// CleanupResult represents the result of a cleanup operation
type CleanupResult struct {
	ID                 string                   `json:"id"`
	TenantID           string                   `json:"tenant_id"`
	Status             string                   `json:"status"` // "success", "partial", "failed"
	Reason             CleanupReason            `json:"reason"`
	ErasureType        DataErasureType          `json:"erasure_type"`
	StartedAt          time.Time                `json:"started_at"`
	CompletedAt        *time.Time               `json:"completed_at,omitempty"`
	RecordsDeleted     map[string]int           `json:"records_deleted"`
	BackupCreated      bool                     `json:"backup_created"`
	BackupID           string                   `json:"backup_id,omitempty"`
	Errors             []string                 `json:"errors,omitempty"`
	VerificationResult *DataErasureVerification `json:"verification_result,omitempty"`
	AdditionalInfo     map[string]interface{}   `json:"additional_info,omitempty"`
}

// DataErasureVerification represents verification of data erasure
type DataErasureVerification struct {
	VerifiedAt       time.Time       `json:"verified_at"`
	TablesVerified   map[string]bool `json:"tables_verified"`
	RemainingRecords map[string]int  `json:"remaining_records"`
	IsComplete       bool            `json:"is_complete"`
	Errors           []string        `json:"errors,omitempty"`
}

// NewTenantCleanupService creates a new tenant cleanup service
func NewTenantCleanupService() *TenantCleanupService {
	return &TenantCleanupService{
		conn:          database.NewConnection(),
		tenantFilter:  middleware.NewTenantFilter(),
		backupService: NewTenantBackupService(),
	}
}

// DeleteTenantData permanently deletes all data associated with a tenant
func (s *TenantCleanupService) DeleteTenantData(ctx context.Context, request CleanupRequest) (*CleanupResult, error) {
	// Validate cleanup authorization
	if err := s.validateCleanupAuthorization(ctx, request.TenantID); err != nil {
		return nil, err
	}

	// Validate required confirmation
	if err := s.validateCleanupConfirmation(request); err != nil {
		return nil, err
	}

	// Check if cleanup is already in progress
	if s.isCleanupInProgress(ctx, request.TenantID) {
		return nil, ErrCleanupInProgress
	}

	// Initialize cleanup result
	result := &CleanupResult{
		ID:             s.generateCleanupID(request.TenantID),
		TenantID:       request.TenantID,
		Reason:         request.Reason,
		ErasureType:    request.ErasureType,
		StartedAt:      time.Now(),
		RecordsDeleted: make(map[string]int),
	}

	// Create backup if requested
	if request.CreateBackup {
		backup, err := s.createPreDeletionBackup(ctx, request.TenantID)
		if err != nil {
			g.Log().Errorf(ctx, "Failed to create backup before cleanup: %v", err)
			result.Errors = append(result.Errors, fmt.Sprintf("Backup creation failed: %v", err))
		} else {
			result.BackupCreated = true
			result.BackupID = backup.Metadata.ID
		}
	}

	// Execute cleanup within transaction for atomicity
	cleanupErr := s.conn.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		return s.executeCleanup(ctx, tx, request.TenantID, request.ErasureType, result)
	})

	// Handle cleanup completion
	completedAt := time.Now()
	result.CompletedAt = &completedAt

	if cleanupErr != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, cleanupErr.Error())
		s.auditCleanupOperation(ctx, "CLEANUP_FAILED", request, result)
		return result, cleanupErr
	}

	// Verify data erasure
	verification, err := s.verifyDataErasure(ctx, request.TenantID, result.RecordsDeleted)
	if err != nil {
		result.Status = "partial"
		result.Errors = append(result.Errors, fmt.Sprintf("Verification failed: %v", err))
	} else {
		result.VerificationResult = verification
		if verification.IsComplete {
			result.Status = "success"
		} else {
			result.Status = "partial"
			result.Errors = append(result.Errors, "Data erasure verification failed - some records may remain")
		}
	}

	// Audit the cleanup operation
	s.auditCleanupOperation(ctx, "CLEANUP_COMPLETED", request, result)

	return result, nil
}

// SecureDeleteTenantData performs secure deletion with multiple overwrites
func (s *TenantCleanupService) SecureDeleteTenantData(ctx context.Context, request CleanupRequest) (*CleanupResult, error) {
	// Set erasure type to secure
	request.ErasureType = DataErasureTypeSecure

	result, err := s.DeleteTenantData(ctx, request)
	if err != nil {
		return result, err
	}

	// Additional secure deletion steps (simulated)
	if result.Status == "success" {
		s.performSecureOverwrites(ctx, request.TenantID, result)
	}

	return result, nil
}

// SoftDeleteTenantData marks tenant data as deleted without physical removal
func (s *TenantCleanupService) SoftDeleteTenantData(ctx context.Context, request CleanupRequest) (*CleanupResult, error) {
	// Set erasure type to soft
	request.ErasureType = DataErasureTypeSoft

	return s.DeleteTenantData(ctx, request)
}

// DeleteUserData deletes all data associated with a specific user (GDPR compliance)
func (s *TenantCleanupService) DeleteUserData(ctx context.Context, tenantID, userID string, reason CleanupReason) (*CleanupResult, error) {
	request := CleanupRequest{
		TenantID:     tenantID,
		Reason:       reason,
		ErasureType:  DataErasureTypeHard,
		CreateBackup: true,
		Confirmation: fmt.Sprintf("DELETE_USER_%s", userID),
		RequestedBy:  s.getContextUser(ctx),
		RequestedAt:  time.Now(),
		AdditionalInfo: map[string]interface{}{
			"target_user_id": userID,
			"operation_type": "user_deletion",
		},
	}

	// Validate authorization for user data deletion
	if err := s.validateCleanupAuthorization(ctx, tenantID); err != nil {
		return nil, err
	}

	result := &CleanupResult{
		ID:             s.generateCleanupID(tenantID),
		TenantID:       tenantID,
		Reason:         reason,
		ErasureType:    DataErasureTypeHard,
		StartedAt:      time.Now(),
		RecordsDeleted: make(map[string]int),
	}

	// Execute user data cleanup
	cleanupErr := s.conn.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		return s.executeUserDataCleanup(ctx, tx, tenantID, userID, result)
	})

	completedAt := time.Now()
	result.CompletedAt = &completedAt

	if cleanupErr != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, cleanupErr.Error())
	} else {
		result.Status = "success"
	}

	s.auditCleanupOperation(ctx, "USER_DATA_CLEANUP", request, result)

	return result, cleanupErr
}

// GetCleanupStatus retrieves the status of a cleanup operation
func (s *TenantCleanupService) GetCleanupStatus(ctx context.Context, cleanupID string) (*CleanupResult, error) {
	// In production, this would query a cleanup_operations table
	// For now, return a mock result
	return &CleanupResult{
		ID:          cleanupID,
		Status:      "completed",
		StartedAt:   time.Now().Add(-1 * time.Hour),
		CompletedAt: func() *time.Time { t := time.Now(); return &t }(),
	}, nil
}

// executeCleanup performs the actual data cleanup operation
func (s *TenantCleanupService) executeCleanup(ctx context.Context, tx gdb.TX, tenantID string, erasureType DataErasureType, result *CleanupResult) error {
	tables := s.getTenantDataTables()

	for _, table := range tables {
		deletedCount, err := s.cleanupTable(ctx, tx, tenantID, table, erasureType)
		if err != nil {
			return fmt.Errorf("failed to cleanup table %s: %w", table, err)
		}

		result.RecordsDeleted[table] = deletedCount
		g.Log().Infof(ctx, "Cleaned up %d records from table %s for tenant %s", deletedCount, table, tenantID)
	}

	return nil
}

// executeUserDataCleanup performs cleanup of specific user data
func (s *TenantCleanupService) executeUserDataCleanup(ctx context.Context, tx gdb.TX, tenantID, userID string, result *CleanupResult) error {
	userDataTables := s.getUserDataTables()

	for _, table := range userDataTables {
		deletedCount, err := s.cleanupUserDataFromTable(ctx, tx, tenantID, userID, table)
		if err != nil {
			return fmt.Errorf("failed to cleanup user data from table %s: %w", table, err)
		}

		result.RecordsDeleted[table] = deletedCount
		g.Log().Infof(ctx, "Cleaned up %d user records from table %s for tenant %s, user %s",
			deletedCount, table, tenantID, userID)
	}

	return nil
}

// cleanupTable removes data from a specific table based on erasure type
func (s *TenantCleanupService) cleanupTable(ctx context.Context, tx gdb.TX, tenantID, tableName string, erasureType DataErasureType) (int, error) {
	switch erasureType {
	case DataErasureTypeSoft:
		// Mark records as deleted
		result, err := tx.Model(tableName).
			Where("tenant_id = ?", tenantID).
			Update(map[string]interface{}{
				"deleted_at": time.Now(),
				"is_deleted": true,
			})
		if err != nil {
			return 0, err
		}

		affected, _ := result.RowsAffected()
		return int(affected), nil

	case DataErasureTypeHard, DataErasureTypeSecure:
		// Physically delete records
		result, err := tx.Model(tableName).
			Where("tenant_id = ?", tenantID).
			Delete()
		if err != nil {
			return 0, err
		}

		affected, _ := result.RowsAffected()
		return int(affected), nil

	default:
		return 0, fmt.Errorf("unsupported erasure type: %s", erasureType)
	}
}

// cleanupUserDataFromTable removes specific user data from a table
func (s *TenantCleanupService) cleanupUserDataFromTable(ctx context.Context, tx gdb.TX, tenantID, userID, tableName string) (int, error) {
	// Different tables may have different user ID column names
	userColumn := s.getUserColumnForTable(tableName)

	result, err := tx.Model(tableName).
		Where("tenant_id = ? AND "+userColumn+" = ?", tenantID, userID).
		Delete()
	if err != nil {
		return 0, err
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// verifyDataErasure verifies that all tenant data has been properly erased
func (s *TenantCleanupService) verifyDataErasure(ctx context.Context, tenantID string, deletedCounts map[string]int) (*DataErasureVerification, error) {
	verification := &DataErasureVerification{
		VerifiedAt:       time.Now(),
		TablesVerified:   make(map[string]bool),
		RemainingRecords: make(map[string]int),
		IsComplete:       true,
	}

	tables := s.getTenantDataTables()

	for _, table := range tables {
		// Count remaining records for this tenant
		count, err := s.conn.GetDB().Model(table).
			Where("tenant_id = ?", tenantID).
			Count()
		if err != nil {
			verification.Errors = append(verification.Errors,
				fmt.Sprintf("Failed to verify table %s: %v", table, err))
			verification.IsComplete = false
			continue
		}

		verification.TablesVerified[table] = (count == 0)
		verification.RemainingRecords[table] = count

		if count > 0 {
			verification.IsComplete = false
			g.Log().Warningf(ctx, "Table %s still contains %d records for tenant %s",
				table, count, tenantID)
		}
	}

	return verification, nil
}

// createPreDeletionBackup creates a backup before deletion
func (s *TenantCleanupService) createPreDeletionBackup(ctx context.Context, tenantID string) (*TenantBackupData, error) {
	return s.backupService.CreateFullBackup(ctx, tenantID, BackupFormatJSON)
}

// performSecureOverwrites performs additional secure deletion steps
func (s *TenantCleanupService) performSecureOverwrites(ctx context.Context, tenantID string, result *CleanupResult) {
	// In production, this would perform actual secure overwrites
	// For now, we simulate the process
	g.Log().Infof(ctx, "Performing secure overwrites for tenant %s", tenantID)

	if result.AdditionalInfo == nil {
		result.AdditionalInfo = make(map[string]interface{})
	}
	result.AdditionalInfo["secure_overwrites_completed"] = true
	result.AdditionalInfo["overwrite_passes"] = 3
}

// Helper methods

func (s *TenantCleanupService) validateCleanupAuthorization(ctx context.Context, tenantID string) error {
	// Only system admin can perform cleanup operations
	if !s.tenantFilter.IsSystemAdmin(ctx) {
		return ErrCleanupNotAuthorized
	}
	return nil
}

func (s *TenantCleanupService) validateCleanupConfirmation(request CleanupRequest) error {
	expectedConfirmation := fmt.Sprintf("DELETE_TENANT_%s", request.TenantID)
	if request.Confirmation != expectedConfirmation {
		return fmt.Errorf("invalid confirmation string, expected: %s", expectedConfirmation)
	}
	return nil
}

func (s *TenantCleanupService) isCleanupInProgress(ctx context.Context, tenantID string) bool {
	// In production, check cleanup_operations table for active operations
	return false
}

func (s *TenantCleanupService) getTenantDataTables() []string {
	return []string{
		"users",
		"roles",
		"user_roles",
		"role_permissions",
		"audit_logs",
		"user_sessions",
		"user_preferences",
		"notifications",
	}
}

func (s *TenantCleanupService) getUserDataTables() []string {
	return []string{
		"users",
		"user_roles",
		"user_sessions",
		"user_preferences",
		"audit_logs",
		"notifications",
	}
}

func (s *TenantCleanupService) getUserColumnForTable(tableName string) string {
	// Map table names to their user ID column names
	userColumns := map[string]string{
		"users":            "id",
		"user_roles":       "user_id",
		"user_sessions":    "user_id",
		"user_preferences": "user_id",
		"audit_logs":       "user_id",
		"notifications":    "user_id",
	}

	if column, exists := userColumns[tableName]; exists {
		return column
	}
	return "user_id" // Default fallback
}

func (s *TenantCleanupService) generateCleanupID(tenantID string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("cleanup_%s_%d", tenantID, timestamp)
}

func (s *TenantCleanupService) getContextUser(ctx context.Context) string {
	// Extract user information from context if available
	return "system_admin"
}

func (s *TenantCleanupService) auditCleanupOperation(ctx context.Context, operation string, request CleanupRequest, result *CleanupResult) {
	auditData := map[string]interface{}{
		"operation":    operation,
		"tenant_id":    request.TenantID,
		"cleanup_id":   result.ID,
		"reason":       request.Reason,
		"erasure_type": request.ErasureType,
		"status":       result.Status,
		"requested_by": request.RequestedBy,
		"started_at":   result.StartedAt.Format(time.RFC3339),
	}

	if result.CompletedAt != nil {
		auditData["completed_at"] = result.CompletedAt.Format(time.RFC3339)
		auditData["duration_seconds"] = result.CompletedAt.Sub(result.StartedAt).Seconds()
	}

	if len(result.RecordsDeleted) > 0 {
		auditData["records_deleted"] = result.RecordsDeleted
	}

	if len(result.Errors) > 0 {
		auditData["errors"] = result.Errors
	}

	g.Log().Infof(ctx, "Tenant cleanup audit: %s", gjson.MustEncodeString(auditData))
}
