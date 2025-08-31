package tenant_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofromzero/project_temp/backend/application/tenant"
)

func TestBackupConstants(t *testing.T) {
	t.Run("BackupFormats", func(t *testing.T) {
		assert.Equal(t, tenant.BackupFormat("json"), tenant.BackupFormatJSON)
		assert.Equal(t, tenant.BackupFormat("sql"), tenant.BackupFormatSQL)
	})

	t.Run("BackupTypes", func(t *testing.T) {
		assert.Equal(t, tenant.BackupType("full"), tenant.BackupTypeFull)
		assert.Equal(t, tenant.BackupType("incremental"), tenant.BackupTypeIncremental)
	})

	t.Run("ErrorConstants", func(t *testing.T) {
		assert.Contains(t, tenant.ErrInvalidBackupFormat.Error(), "invalid backup format")
		assert.Contains(t, tenant.ErrBackupNotFound.Error(), "backup not found")
		assert.Contains(t, tenant.ErrInvalidTenantForRestore.Error(), "invalid tenant")
		assert.Contains(t, tenant.ErrBackupIntegrityCheck.Error(), "integrity check failed")
	})
}

func TestTenantBackupMetadata(t *testing.T) {
	t.Run("MetadataStructure", func(t *testing.T) {
		metadata := tenant.TenantBackupMetadata{
			ID:           "backup-123",
			TenantID:     "tenant-456",
			BackupType:   tenant.BackupTypeFull,
			Format:       tenant.BackupFormatJSON,
			CreatedAt:    time.Now(),
			Size:         1024,
			RecordCounts: map[string]int{"users": 5, "roles": 3},
			Checksum:     "checksum-abc",
			Version:      "1.0",
			AdditionalInfo: map[string]interface{}{
				"created_by": "system",
				"source":     "test",
			},
		}

		assert.Equal(t, "backup-123", metadata.ID)
		assert.Equal(t, "tenant-456", metadata.TenantID)
		assert.Equal(t, tenant.BackupTypeFull, metadata.BackupType)
		assert.Equal(t, tenant.BackupFormatJSON, metadata.Format)
		assert.Equal(t, int64(1024), metadata.Size)
		assert.Equal(t, 5, metadata.RecordCounts["users"])
		assert.Equal(t, 3, metadata.RecordCounts["roles"])
		assert.Equal(t, "checksum-abc", metadata.Checksum)
		assert.Equal(t, "1.0", metadata.Version)
		assert.Equal(t, "system", metadata.AdditionalInfo["created_by"])
	})

	t.Run("IncrementalMetadata", func(t *testing.T) {
		since := time.Now().Add(-24 * time.Hour)
		metadata := tenant.TenantBackupMetadata{
			ID:                "inc-backup-123",
			TenantID:          "tenant-456",
			BackupType:        tenant.BackupTypeIncremental,
			LastIncrementalAt: &since,
		}

		assert.Equal(t, tenant.BackupTypeIncremental, metadata.BackupType)
		require.NotNil(t, metadata.LastIncrementalAt)
		assert.True(t, metadata.LastIncrementalAt.Before(time.Now()))
	})
}

func TestTenantBackupData(t *testing.T) {
	t.Run("BackupDataStructure", func(t *testing.T) {
		metadata := tenant.TenantBackupMetadata{
			ID:           "backup-123",
			TenantID:     "tenant-456",
			BackupType:   tenant.BackupTypeFull,
			Format:       tenant.BackupFormatJSON,
			CreatedAt:    time.Now(),
			RecordCounts: map[string]int{"users": 2},
		}

		data := map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": "1", "name": "User 1", "tenant_id": "tenant-456"},
				{"id": "2", "name": "User 2", "tenant_id": "tenant-456"},
			},
		}

		backup := tenant.TenantBackupData{
			Metadata: metadata,
			Data:     data,
		}

		assert.Equal(t, "backup-123", backup.Metadata.ID)
		assert.Equal(t, "tenant-456", backup.Metadata.TenantID)
		assert.Contains(t, backup.Data, "users")

		users := backup.Data["users"].([]map[string]interface{})
		assert.Equal(t, 2, len(users))
		assert.Equal(t, "User 1", users[0]["name"])
	})

	t.Run("JSONSerialization", func(t *testing.T) {
		metadata := tenant.TenantBackupMetadata{
			ID:         "backup-test",
			TenantID:   "tenant-test",
			BackupType: tenant.BackupTypeFull,
			Format:     tenant.BackupFormatJSON,
			CreatedAt:  time.Now(),
		}

		backup := tenant.TenantBackupData{
			Metadata: metadata,
			Data: map[string]interface{}{
				"test_table": []map[string]interface{}{
					{"id": "1", "name": "Test"},
				},
			},
		}

		// Test JSON marshaling
		jsonBytes, err := json.Marshal(backup)
		assert.NoError(t, err)
		assert.Contains(t, string(jsonBytes), "backup-test")

		// Test JSON unmarshaling
		var restored tenant.TenantBackupData
		err = json.Unmarshal(jsonBytes, &restored)
		assert.NoError(t, err)
		assert.Equal(t, "backup-test", restored.Metadata.ID)
		assert.Equal(t, "tenant-test", restored.Metadata.TenantID)
	})
}

func TestTenantRestoreOptions(t *testing.T) {
	t.Run("RestoreOptionsStructure", func(t *testing.T) {
		options := tenant.TenantRestoreOptions{
			TargetTenantID:     "target-tenant",
			OverwriteExisting:  true,
			ValidateIntegrity:  true,
			ConflictResolution: "overwrite",
			TableFilters:       []string{"users", "roles"},
		}

		assert.Equal(t, "target-tenant", options.TargetTenantID)
		assert.True(t, options.OverwriteExisting)
		assert.True(t, options.ValidateIntegrity)
		assert.Equal(t, "overwrite", options.ConflictResolution)
		assert.Equal(t, []string{"users", "roles"}, options.TableFilters)
	})

	t.Run("DefaultRestoreOptions", func(t *testing.T) {
		options := tenant.TenantRestoreOptions{
			TargetTenantID: "target-tenant",
		}

		assert.Equal(t, "target-tenant", options.TargetTenantID)
		assert.False(t, options.OverwriteExisting)      // Default false
		assert.False(t, options.ValidateIntegrity)      // Default false
		assert.Equal(t, "", options.ConflictResolution) // Default empty
		assert.Nil(t, options.TableFilters)             // Default nil
	})
}

func TestBackupFormats(t *testing.T) {
	t.Run("JSONFormatHandling", func(t *testing.T) {
		// Test JSON format constants and validation
		format := tenant.BackupFormatJSON
		assert.Equal(t, "json", string(format))

		// Test that we can work with JSON data structures
		testData := []map[string]interface{}{
			{"id": "1", "name": "Test User", "tenant_id": "tenant-123"},
			{"id": "2", "name": "Test User 2", "tenant_id": "tenant-123"},
		}

		jsonBytes, err := json.Marshal(testData)
		assert.NoError(t, err)
		assert.Contains(t, string(jsonBytes), "Test User")

		var restored []map[string]interface{}
		err = json.Unmarshal(jsonBytes, &restored)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(restored))
		assert.Equal(t, "Test User", restored[0]["name"])
	})

	t.Run("SQLFormatHandling", func(t *testing.T) {
		// Test SQL format constants
		format := tenant.BackupFormatSQL
		assert.Equal(t, "sql", string(format))

		// Test SQL statement generation patterns
		tableName := "users"
		columns := []string{"id", "name", "tenant_id"}
		values := []string{"'1'", "'Test User'", "'tenant-123'"}

		expectedSQL := "INSERT INTO users (id, name, tenant_id) VALUES ('1', 'Test User', 'tenant-123')"
		actualSQL := "INSERT INTO " + tableName + " (" +
			strings.Join(columns, ", ") + ") VALUES (" +
			strings.Join(values, ", ") + ")"

		assert.Equal(t, expectedSQL, actualSQL)
	})
}

func TestBackupDataStructures(t *testing.T) {
	t.Run("CompleteBackupWorkflow", func(t *testing.T) {
		// Create mock backup data
		metadata := tenant.TenantBackupMetadata{
			ID:         "backup-workflow-test",
			TenantID:   "tenant-workflow",
			BackupType: tenant.BackupTypeFull,
			Format:     tenant.BackupFormatJSON,
			CreatedAt:  time.Now(),
			Size:       2048,
			RecordCounts: map[string]int{
				"users": 3,
				"roles": 2,
			},
			Checksum: "workflow-checksum",
			Version:  "1.0",
			AdditionalInfo: map[string]interface{}{
				"created_by": "workflow-test",
				"source":     "unit-test",
			},
		}

		data := map[string]interface{}{
			"users": []map[string]interface{}{
				{"id": "1", "name": "User 1", "tenant_id": "tenant-workflow"},
				{"id": "2", "name": "User 2", "tenant_id": "tenant-workflow"},
				{"id": "3", "name": "User 3", "tenant_id": "tenant-workflow"},
			},
			"roles": []map[string]interface{}{
				{"id": "1", "name": "Admin", "tenant_id": "tenant-workflow"},
				{"id": "2", "name": "User", "tenant_id": "tenant-workflow"},
			},
		}

		backup := tenant.TenantBackupData{
			Metadata: metadata,
			Data:     data,
		}

		// Verify backup structure
		assert.Equal(t, "backup-workflow-test", backup.Metadata.ID)
		assert.Equal(t, "tenant-workflow", backup.Metadata.TenantID)
		assert.Equal(t, tenant.BackupTypeFull, backup.Metadata.BackupType)
		assert.Equal(t, tenant.BackupFormatJSON, backup.Metadata.Format)
		assert.Equal(t, int64(2048), backup.Metadata.Size)
		assert.Equal(t, 3, backup.Metadata.RecordCounts["users"])
		assert.Equal(t, 2, backup.Metadata.RecordCounts["roles"])

		// Verify data integrity
		users := backup.Data["users"].([]map[string]interface{})
		roles := backup.Data["roles"].([]map[string]interface{})

		assert.Equal(t, 3, len(users))
		assert.Equal(t, 2, len(roles))

		// All users should belong to the same tenant
		for _, user := range users {
			assert.Equal(t, "tenant-workflow", user["tenant_id"])
		}

		for _, role := range roles {
			assert.Equal(t, "tenant-workflow", role["tenant_id"])
		}

		// Test JSON serialization/deserialization
		jsonBytes, err := json.Marshal(backup)
		assert.NoError(t, err)

		var restoredBackup tenant.TenantBackupData
		err = json.Unmarshal(jsonBytes, &restoredBackup)
		assert.NoError(t, err)

		assert.Equal(t, backup.Metadata.ID, restoredBackup.Metadata.ID)
		assert.Equal(t, backup.Metadata.TenantID, restoredBackup.Metadata.TenantID)

		restoredUsers := restoredBackup.Data["users"].([]interface{})
		assert.Equal(t, 3, len(restoredUsers))
	})

	t.Run("IncrementalBackupMetadata", func(t *testing.T) {
		since := time.Now().Add(-24 * time.Hour)

		metadata := tenant.TenantBackupMetadata{
			ID:                "inc-backup-test",
			TenantID:          "tenant-inc",
			BackupType:        tenant.BackupTypeIncremental,
			Format:            tenant.BackupFormatJSON,
			CreatedAt:         time.Now(),
			LastIncrementalAt: &since,
			RecordCounts:      map[string]int{"users": 1}, // Only 1 user changed
			AdditionalInfo: map[string]interface{}{
				"incremental_since": since.Format(time.RFC3339),
			},
		}

		assert.Equal(t, tenant.BackupTypeIncremental, metadata.BackupType)
		assert.NotNil(t, metadata.LastIncrementalAt)
		assert.True(t, metadata.LastIncrementalAt.Before(time.Now()))
		assert.Equal(t, 1, metadata.RecordCounts["users"])
		assert.Contains(t, metadata.AdditionalInfo, "incremental_since")
	})
}

func TestRestoreOptionsValidation(t *testing.T) {
	t.Run("ConflictResolutionOptions", func(t *testing.T) {
		validOptions := []string{"skip", "overwrite", "merge"}

		for _, option := range validOptions {
			restoreOptions := tenant.TenantRestoreOptions{
				TargetTenantID:     "target-tenant",
				ConflictResolution: option,
			}

			assert.Equal(t, option, restoreOptions.ConflictResolution)
		}
	})

	t.Run("TableFilters", func(t *testing.T) {
		options := tenant.TenantRestoreOptions{
			TargetTenantID: "target-tenant",
			TableFilters:   []string{"users", "roles"},
		}

		// Test that we can check if a table is in the filter
		assert.Contains(t, options.TableFilters, "users")
		assert.Contains(t, options.TableFilters, "roles")
		assert.NotContains(t, options.TableFilters, "audit_logs")
	})

	t.Run("IntegrityValidationFlag", func(t *testing.T) {
		options := tenant.TenantRestoreOptions{
			TargetTenantID:    "target-tenant",
			ValidateIntegrity: true,
		}

		assert.True(t, options.ValidateIntegrity)
	})
}
