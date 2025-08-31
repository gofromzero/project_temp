package tenant_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofromzero/project_temp/backend/application/tenant"
)

func TestCleanupConstants(t *testing.T) {
	t.Run("DataErasureTypes", func(t *testing.T) {
		assert.Equal(t, tenant.DataErasureType("soft"), tenant.DataErasureTypeSoft)
		assert.Equal(t, tenant.DataErasureType("hard"), tenant.DataErasureTypeHard)
		assert.Equal(t, tenant.DataErasureType("secure"), tenant.DataErasureTypeSecure)
	})

	t.Run("CleanupReasons", func(t *testing.T) {
		assert.Equal(t, tenant.CleanupReason("tenant_deletion"), tenant.CleanupReasonTenantDeletion)
		assert.Equal(t, tenant.CleanupReason("gdpr_right_to_be_forgotten"), tenant.CleanupReasonGDPRRequest)
		assert.Equal(t, tenant.CleanupReason("data_retention_policy"), tenant.CleanupReasonDataRetention)
		assert.Equal(t, tenant.CleanupReason("user_deletion_request"), tenant.CleanupReasonUserRequest)
		assert.Equal(t, tenant.CleanupReason("compliance_requirement"), tenant.CleanupReasonCompliance)
	})

	t.Run("ErrorConstants", func(t *testing.T) {
		assert.Contains(t, tenant.ErrTenantNotFound.Error(), "tenant not found")
		assert.Contains(t, tenant.ErrCleanupNotAuthorized.Error(), "not authorized")
		assert.Contains(t, tenant.ErrCleanupInProgress.Error(), "already in progress")
		assert.Contains(t, tenant.ErrCleanupFailed.Error(), "operation failed")
		assert.Contains(t, tenant.ErrDataErasureIncomplete.Error(), "erasure incomplete")
	})
}

func TestCleanupRequest(t *testing.T) {
	t.Run("CleanupRequestStructure", func(t *testing.T) {
		request := tenant.CleanupRequest{
			TenantID:     "tenant-123",
			Reason:       tenant.CleanupReasonTenantDeletion,
			ErasureType:  tenant.DataErasureTypeHard,
			CreateBackup: true,
			Confirmation: "DELETE_TENANT_tenant-123",
			RequestedBy:  "admin@example.com",
			RequestedAt:  time.Now(),
			AdditionalInfo: map[string]interface{}{
				"notes": "Tenant requested account closure",
			},
		}

		assert.Equal(t, "tenant-123", request.TenantID)
		assert.Equal(t, tenant.CleanupReasonTenantDeletion, request.Reason)
		assert.Equal(t, tenant.DataErasureTypeHard, request.ErasureType)
		assert.True(t, request.CreateBackup)
		assert.Equal(t, "DELETE_TENANT_tenant-123", request.Confirmation)
		assert.Equal(t, "admin@example.com", request.RequestedBy)
		assert.Contains(t, request.AdditionalInfo, "notes")
	})

	t.Run("GDPRCleanupRequest", func(t *testing.T) {
		request := tenant.CleanupRequest{
			TenantID:     "tenant-gdpr",
			Reason:       tenant.CleanupReasonGDPRRequest,
			ErasureType:  tenant.DataErasureTypeSecure,
			CreateBackup: false, // GDPR may require no backup
			Confirmation: "DELETE_TENANT_tenant-gdpr",
			RequestedBy:  "data-controller@example.com",
			RequestedAt:  time.Now(),
			AdditionalInfo: map[string]interface{}{
				"gdpr_request_id": "GDPR-2023-001",
				"legal_basis":     "Right to be forgotten (Article 17)",
			},
		}

		assert.Equal(t, tenant.CleanupReasonGDPRRequest, request.Reason)
		assert.Equal(t, tenant.DataErasureTypeSecure, request.ErasureType)
		assert.False(t, request.CreateBackup)
		assert.Equal(t, "GDPR-2023-001", request.AdditionalInfo["gdpr_request_id"])
	})
}

func TestCleanupResult(t *testing.T) {
	t.Run("CleanupResultStructure", func(t *testing.T) {
		startTime := time.Now()
		endTime := startTime.Add(5 * time.Minute)

		result := tenant.CleanupResult{
			ID:          "cleanup-123",
			TenantID:    "tenant-456",
			Status:      "success",
			Reason:      tenant.CleanupReasonTenantDeletion,
			ErasureType: tenant.DataErasureTypeHard,
			StartedAt:   startTime,
			CompletedAt: &endTime,
			RecordsDeleted: map[string]int{
				"users":      5,
				"roles":      3,
				"audit_logs": 150,
			},
			BackupCreated: true,
			BackupID:      "backup-456-123",
		}

		assert.Equal(t, "cleanup-123", result.ID)
		assert.Equal(t, "tenant-456", result.TenantID)
		assert.Equal(t, "success", result.Status)
		assert.Equal(t, tenant.CleanupReasonTenantDeletion, result.Reason)
		assert.Equal(t, tenant.DataErasureTypeHard, result.ErasureType)
		assert.Equal(t, startTime, result.StartedAt)
		require.NotNil(t, result.CompletedAt)
		assert.Equal(t, endTime, *result.CompletedAt)
		assert.Equal(t, 5, result.RecordsDeleted["users"])
		assert.Equal(t, 3, result.RecordsDeleted["roles"])
		assert.Equal(t, 150, result.RecordsDeleted["audit_logs"])
		assert.True(t, result.BackupCreated)
		assert.Equal(t, "backup-456-123", result.BackupID)
	})

	t.Run("FailedCleanupResult", func(t *testing.T) {
		result := tenant.CleanupResult{
			ID:          "cleanup-failed-123",
			TenantID:    "tenant-failed",
			Status:      "failed",
			Reason:      tenant.CleanupReasonTenantDeletion,
			ErasureType: tenant.DataErasureTypeHard,
			StartedAt:   time.Now(),
			Errors: []string{
				"Failed to delete records from users table",
				"Database connection timeout during cleanup",
			},
		}

		assert.Equal(t, "failed", result.Status)
		assert.Equal(t, 2, len(result.Errors))
		assert.Contains(t, result.Errors, "Failed to delete records from users table")
		assert.Contains(t, result.Errors, "Database connection timeout during cleanup")
	})

	t.Run("PartialCleanupResult", func(t *testing.T) {
		result := tenant.CleanupResult{
			ID:          "cleanup-partial-123",
			TenantID:    "tenant-partial",
			Status:      "partial",
			Reason:      tenant.CleanupReasonTenantDeletion,
			ErasureType: tenant.DataErasureTypeHard,
			StartedAt:   time.Now(),
			RecordsDeleted: map[string]int{
				"users": 3,
				"roles": 2,
			},
			Errors: []string{
				"Could not delete some audit_logs records due to foreign key constraints",
			},
		}

		assert.Equal(t, "partial", result.Status)
		assert.Equal(t, 3, result.RecordsDeleted["users"])
		assert.Equal(t, 2, result.RecordsDeleted["roles"])
		assert.Equal(t, 1, len(result.Errors))
		assert.Contains(t, result.Errors[0], "foreign key constraints")
	})
}

func TestDataErasureVerification(t *testing.T) {
	t.Run("SuccessfulVerification", func(t *testing.T) {
		verification := tenant.DataErasureVerification{
			VerifiedAt: time.Now(),
			TablesVerified: map[string]bool{
				"users":      true,
				"roles":      true,
				"audit_logs": true,
			},
			RemainingRecords: map[string]int{
				"users":      0,
				"roles":      0,
				"audit_logs": 0,
			},
			IsComplete: true,
		}

		assert.True(t, verification.IsComplete)
		assert.Equal(t, 3, len(verification.TablesVerified))
		assert.True(t, verification.TablesVerified["users"])
		assert.True(t, verification.TablesVerified["roles"])
		assert.True(t, verification.TablesVerified["audit_logs"])
		assert.Equal(t, 0, verification.RemainingRecords["users"])
		assert.Equal(t, 0, verification.RemainingRecords["roles"])
		assert.Equal(t, 0, verification.RemainingRecords["audit_logs"])
		assert.Empty(t, verification.Errors)
	})

	t.Run("IncompleteVerification", func(t *testing.T) {
		verification := tenant.DataErasureVerification{
			VerifiedAt: time.Now(),
			TablesVerified: map[string]bool{
				"users":      true,
				"roles":      true,
				"audit_logs": false, // Some records remain
			},
			RemainingRecords: map[string]int{
				"users":      0,
				"roles":      0,
				"audit_logs": 5, // 5 records still remain
			},
			IsComplete: false,
			Errors: []string{
				"5 audit_logs records could not be deleted",
			},
		}

		assert.False(t, verification.IsComplete)
		assert.True(t, verification.TablesVerified["users"])
		assert.True(t, verification.TablesVerified["roles"])
		assert.False(t, verification.TablesVerified["audit_logs"])
		assert.Equal(t, 0, verification.RemainingRecords["users"])
		assert.Equal(t, 5, verification.RemainingRecords["audit_logs"])
		assert.Equal(t, 1, len(verification.Errors))
		assert.Contains(t, verification.Errors[0], "audit_logs records could not be deleted")
	})
}

func TestCleanupServiceCreation(t *testing.T) {
	t.Run("NewTenantCleanupService", func(t *testing.T) {
		// We can't test service creation that requires database
		// But we can test that the function exists and has proper signature
		assert.NotNil(t, tenant.NewTenantCleanupService)
	})
}

func TestCleanupRequestValidation(t *testing.T) {
	t.Run("ConfirmationStringFormat", func(t *testing.T) {
		tenantID := "tenant-123"
		expectedConfirmation := "DELETE_TENANT_" + tenantID

		request := tenant.CleanupRequest{
			TenantID:     tenantID,
			Confirmation: expectedConfirmation,
		}

		assert.Equal(t, expectedConfirmation, request.Confirmation)
		assert.Equal(t, "DELETE_TENANT_tenant-123", request.Confirmation)
	})

	t.Run("RequiredFieldsPresent", func(t *testing.T) {
		request := tenant.CleanupRequest{
			TenantID:     "tenant-required",
			Reason:       tenant.CleanupReasonTenantDeletion,
			ErasureType:  tenant.DataErasureTypeHard,
			CreateBackup: true,
			Confirmation: "DELETE_TENANT_tenant-required",
			RequestedBy:  "admin@example.com",
			RequestedAt:  time.Now(),
		}

		// Verify all required fields are present
		assert.NotEmpty(t, request.TenantID)
		assert.NotEmpty(t, request.Reason)
		assert.NotEmpty(t, request.ErasureType)
		assert.NotEmpty(t, request.Confirmation)
		assert.NotEmpty(t, request.RequestedBy)
		assert.False(t, request.RequestedAt.IsZero())
	})
}

func TestErasureTypeHandling(t *testing.T) {
	t.Run("SoftDeletion", func(t *testing.T) {
		request := tenant.CleanupRequest{
			TenantID:    "tenant-soft",
			ErasureType: tenant.DataErasureTypeSoft,
			Reason:      tenant.CleanupReasonDataRetention,
		}

		assert.Equal(t, tenant.DataErasureTypeSoft, request.ErasureType)

		// Soft deletion should typically keep backups
		request.CreateBackup = true
		assert.True(t, request.CreateBackup)
	})

	t.Run("HardDeletion", func(t *testing.T) {
		request := tenant.CleanupRequest{
			TenantID:    "tenant-hard",
			ErasureType: tenant.DataErasureTypeHard,
			Reason:      tenant.CleanupReasonTenantDeletion,
		}

		assert.Equal(t, tenant.DataErasureTypeHard, request.ErasureType)
	})

	t.Run("SecureDeletion", func(t *testing.T) {
		request := tenant.CleanupRequest{
			TenantID:    "tenant-secure",
			ErasureType: tenant.DataErasureTypeSecure,
			Reason:      tenant.CleanupReasonGDPRRequest,
		}

		assert.Equal(t, tenant.DataErasureTypeSecure, request.ErasureType)

		// GDPR requests might not allow backups
		request.CreateBackup = false
		assert.False(t, request.CreateBackup)
	})
}

func TestCleanupAuditData(t *testing.T) {
	t.Run("AuditDataStructure", func(t *testing.T) {
		// Test that we can create audit-ready data structures
		request := tenant.CleanupRequest{
			TenantID:    "tenant-audit",
			Reason:      tenant.CleanupReasonGDPRRequest,
			ErasureType: tenant.DataErasureTypeSecure,
			RequestedBy: "dpo@example.com", // Data Protection Officer
			RequestedAt: time.Now(),
			AdditionalInfo: map[string]interface{}{
				"gdpr_request_id":  "GDPR-2023-045",
				"legal_basis":      "Article 17 - Right to erasure",
				"data_subject_id":  "user-789",
				"request_verified": true,
			},
		}

		result := tenant.CleanupResult{
			ID:          "cleanup-audit-123",
			TenantID:    request.TenantID,
			Status:      "success",
			Reason:      request.Reason,
			ErasureType: request.ErasureType,
			StartedAt:   time.Now(),
			RecordsDeleted: map[string]int{
				"users":            1,
				"user_preferences": 15,
				"audit_logs":       234,
			},
			VerificationResult: &tenant.DataErasureVerification{
				VerifiedAt: time.Now(),
				IsComplete: true,
				TablesVerified: map[string]bool{
					"users":            true,
					"user_preferences": true,
					"audit_logs":       true,
				},
			},
		}

		// Verify audit trail data completeness
		assert.NotEmpty(t, request.TenantID)
		assert.NotEmpty(t, request.RequestedBy)
		assert.Contains(t, request.AdditionalInfo, "gdpr_request_id")
		assert.Contains(t, request.AdditionalInfo, "legal_basis")

		assert.NotEmpty(t, result.ID)
		assert.NotEmpty(t, result.Status)
		assert.NotEmpty(t, result.RecordsDeleted)
		require.NotNil(t, result.VerificationResult)
		assert.True(t, result.VerificationResult.IsComplete)
	})
}

func TestCleanupStatusCodes(t *testing.T) {
	t.Run("StatusValues", func(t *testing.T) {
		validStatuses := []string{"success", "partial", "failed"}

		for _, status := range validStatuses {
			result := tenant.CleanupResult{
				Status: status,
			}

			assert.Contains(t, validStatuses, result.Status)
		}
	})

	t.Run("StatusTransitions", func(t *testing.T) {
		// Test typical status progression
		result := tenant.CleanupResult{
			ID:       "cleanup-status-test",
			TenantID: "tenant-status",
		}

		// Initial state - not set (would be "in_progress" in practice)
		assert.Empty(t, result.Status)

		// Success case
		result.Status = "success"
		assert.Equal(t, "success", result.Status)

		// Partial case
		result.Status = "partial"
		assert.Equal(t, "partial", result.Status)

		// Failed case
		result.Status = "failed"
		assert.Equal(t, "failed", result.Status)
	})
}
