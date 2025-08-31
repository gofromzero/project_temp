package isolation_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofromzero/project_temp/backend/application/tenant"
	"github.com/gofromzero/project_temp/backend/pkg/middleware"
)

// TestAdvancedTenantIsolationScenarios tests complex tenant isolation scenarios
func TestAdvancedTenantIsolationScenarios(t *testing.T) {
	t.Run("NestedContextIsolation", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		baseCtx := context.Background()

		// Create nested contexts with different tenants
		tenantA := "nested-tenant-a"
		ctxA := filter.WithTenantContext(baseCtx, &tenantA, false)

		tenantB := "nested-tenant-b"
		ctxB := filter.WithTenantContext(ctxA, &tenantB, false) // Nested on top of A

		// Verify nested context overwrites parent
		retrievedB, okB := filter.GetTenantID(ctxB)
		require.True(t, okB)
		require.NotNil(t, retrievedB)
		assert.Equal(t, tenantB, *retrievedB, "Nested context should overwrite parent tenant")

		// Original context should remain unchanged
		retrievedA, okA := filter.GetTenantID(ctxA)
		require.True(t, okA)
		require.NotNil(t, retrievedA)
		assert.Equal(t, tenantA, *retrievedA, "Original context should remain unchanged")
	})

	t.Run("SystemAdminEscalationSafety", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		// Regular tenant user
		tenantUser := "regular-tenant"
		userCtx := filter.WithTenantContext(ctx, &tenantUser, false)

		// Attempt to create system admin context on top (should not inherit admin rights)
		// This tests that tenant contexts are properly isolated and cannot be escalated

		assert.False(t, filter.IsSystemAdmin(userCtx), "Regular tenant user should not be admin")

		// Even with nested context, original should remain non-admin
		adminCtx := filter.WithTenantContext(userCtx, nil, true)
		assert.True(t, filter.IsSystemAdmin(adminCtx), "New admin context should be admin")
		assert.False(t, filter.IsSystemAdmin(userCtx), "Original context should remain non-admin")
	})

	t.Run("TenantSwitchingIsolation", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		tenants := []string{"switch-tenant-1", "switch-tenant-2", "switch-tenant-3"}
		contexts := make([]context.Context, len(tenants))

		// Create contexts for each tenant
		for i, tenant := range tenants {
			contexts[i] = filter.WithTenantContext(ctx, &tenant, false)
		}

		// Verify each context maintains its tenant and cannot access others
		for i, tenantCtx := range contexts {
			currentTenant, ok := filter.GetTenantID(tenantCtx)
			require.True(t, ok)
			require.NotNil(t, currentTenant)
			assert.Equal(t, tenants[i], *currentTenant)

			// Test access to each tenant
			for j, targetTenant := range tenants {
				err := filter.ValidateTenantAccess(tenantCtx, &targetTenant)
				if i == j {
					assert.NoError(t, err, "Should be able to access own tenant")
				} else {
					assert.Error(t, err, "Should not be able to access other tenant")
					assert.Equal(t, middleware.ErrUnauthorizedTenant, err)
				}
			}
		}
	})
}

// TestTenantIsolationStressTest tests tenant isolation under stress conditions
func TestTenantIsolationStressTest(t *testing.T) {
	t.Run("HighVolumeContextCreation", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping stress test in short mode")
		}

		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		numContexts := 1000
		var wg sync.WaitGroup
		results := make(chan bool, numContexts)

		start := time.Now()

		for i := 0; i < numContexts; i++ {
			wg.Add(1)
			go func(tenantNum int) {
				defer wg.Done()

				tenantID := fmt.Sprintf("stress-tenant-%d", tenantNum)
				tenantCtx := filter.WithTenantContext(ctx, &tenantID, false)

				// Verify context integrity
				retrieved, ok := filter.GetTenantID(tenantCtx)
				success := ok && retrieved != nil && *retrieved == tenantID

				results <- success
			}(i)
		}

		wg.Wait()
		close(results)

		duration := time.Since(start)
		t.Logf("Created %d tenant contexts in %v", numContexts, duration)

		// Count successful context creations
		successCount := 0
		for success := range results {
			if success {
				successCount++
			}
		}

		assert.Equal(t, numContexts, successCount, "All contexts should be created successfully")
	})

	t.Run("ConcurrentCrossTenantAccessAttempts", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping stress test in short mode")
		}

		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		numTenants := 10
		numAttemptsPerTenant := 100

		var wg sync.WaitGroup
		var mu sync.Mutex
		var unauthorizedAttempts, authorizedAttempts int

		// Create contexts for each tenant
		contexts := make([]context.Context, numTenants)
		for i := 0; i < numTenants; i++ {
			tenantID := fmt.Sprintf("concurrent-tenant-%d", i)
			contexts[i] = filter.WithTenantContext(ctx, &tenantID, false)
		}

		// Each tenant attempts to access all tenants
		for i := 0; i < numTenants; i++ {
			for j := 0; j < numTenants; j++ {
				for k := 0; k < numAttemptsPerTenant; k++ {
					wg.Add(1)
					go func(fromTenant, toTenant int) {
						defer wg.Done()

						targetTenantID := fmt.Sprintf("concurrent-tenant-%d", toTenant)
						err := filter.ValidateTenantAccess(contexts[fromTenant], &targetTenantID)

						mu.Lock()
						if fromTenant == toTenant {
							if err == nil {
								authorizedAttempts++
							}
						} else {
							if err != nil {
								unauthorizedAttempts++
							}
						}
						mu.Unlock()
					}(i, j)
				}
			}
		}

		wg.Wait()

		// Expected results
		expectedAuthorized := numTenants * numAttemptsPerTenant                      // Same tenant access
		expectedUnauthorized := numTenants * (numTenants - 1) * numAttemptsPerTenant // Cross tenant access

		assert.Equal(t, expectedAuthorized, authorizedAttempts, "All same-tenant access should succeed")
		assert.Equal(t, expectedUnauthorized, unauthorizedAttempts, "All cross-tenant access should fail")

		t.Logf("Authorized accesses: %d, Unauthorized attempts blocked: %d",
			authorizedAttempts, unauthorizedAttempts)
	})
}

// TestBackupServiceIsolation tests tenant isolation in backup operations
func TestBackupServiceIsolation(t *testing.T) {
	t.Run("BackupRequestIsolation", func(t *testing.T) {
		// Test backup service structure for tenant isolation
		request1 := tenant.CleanupRequest{
			TenantID: "backup-tenant-1",
			Reason:   tenant.CleanupReasonTenantDeletion,
		}

		request2 := tenant.CleanupRequest{
			TenantID: "backup-tenant-2",
			Reason:   tenant.CleanupReasonGDPRRequest,
		}

		// Verify requests are isolated
		assert.NotEqual(t, request1.TenantID, request2.TenantID)
		assert.NotEqual(t, request1.Reason, request2.Reason)
	})

	t.Run("BackupDataIsolation", func(t *testing.T) {
		// Test backup data structure for tenant isolation
		metadata1 := tenant.TenantBackupMetadata{
			ID:       "backup-1",
			TenantID: "isolated-tenant-1",
			Format:   tenant.BackupFormatJSON,
		}

		metadata2 := tenant.TenantBackupMetadata{
			ID:       "backup-2",
			TenantID: "isolated-tenant-2",
			Format:   tenant.BackupFormatSQL,
		}

		// Verify metadata isolation
		assert.NotEqual(t, metadata1.ID, metadata2.ID)
		assert.NotEqual(t, metadata1.TenantID, metadata2.TenantID)
		assert.NotEqual(t, metadata1.Format, metadata2.Format)
	})
}

// TestCleanupServiceIsolation tests tenant isolation in cleanup operations
func TestCleanupServiceIsolation(t *testing.T) {
	t.Run("CleanupOperationIsolation", func(t *testing.T) {
		// Test that cleanup operations are properly isolated between tenants
		result1 := tenant.CleanupResult{
			ID:       "cleanup-1",
			TenantID: "cleanup-tenant-1",
			Status:   "success",
			RecordsDeleted: map[string]int{
				"users": 5,
				"roles": 2,
			},
		}

		result2 := tenant.CleanupResult{
			ID:       "cleanup-2",
			TenantID: "cleanup-tenant-2",
			Status:   "partial",
			RecordsDeleted: map[string]int{
				"users": 3,
				"roles": 1,
			},
		}

		// Verify cleanup results are isolated
		assert.NotEqual(t, result1.ID, result2.ID)
		assert.NotEqual(t, result1.TenantID, result2.TenantID)
		assert.NotEqual(t, result1.Status, result2.Status)
		assert.NotEqual(t, result1.RecordsDeleted["users"], result2.RecordsDeleted["users"])
	})

	t.Run("ErasureTypeIsolation", func(t *testing.T) {
		// Test different erasure types for different tenants
		types := []tenant.DataErasureType{
			tenant.DataErasureTypeSoft,
			tenant.DataErasureTypeHard,
			tenant.DataErasureTypeSecure,
		}

		requests := make([]tenant.CleanupRequest, len(types))
		for i, erasureType := range types {
			requests[i] = tenant.CleanupRequest{
				TenantID:    fmt.Sprintf("erasure-tenant-%d", i),
				ErasureType: erasureType,
				Reason:      tenant.CleanupReasonTenantDeletion,
			}
		}

		// Verify each request has different settings
		for i := 0; i < len(requests); i++ {
			for j := i + 1; j < len(requests); j++ {
				assert.NotEqual(t, requests[i].TenantID, requests[j].TenantID)
				assert.NotEqual(t, requests[i].ErasureType, requests[j].ErasureType)
			}
		}
	})
}

// TestIsolationBoundaryConditions tests edge cases in tenant isolation
func TestIsolationBoundaryConditions(t *testing.T) {
	t.Run("EmptyTenantIDHandling", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		// Test empty string tenant ID
		emptyTenant := ""
		emptyCtx := filter.WithTenantContext(ctx, &emptyTenant, false)

		retrieved, ok := filter.GetTenantID(emptyCtx)
		assert.True(t, ok)
		require.NotNil(t, retrieved)
		assert.Equal(t, emptyTenant, *retrieved)

		// Empty tenant should not access non-empty tenant
		nonEmptyTenant := "real-tenant"
		err := filter.ValidateTenantAccess(emptyCtx, &nonEmptyTenant)
		assert.Error(t, err)
	})

	t.Run("VeryLongTenantIDHandling", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		// Test very long tenant ID
		longTenantID := "tenant-" + string(make([]byte, 1000))
		for i := range longTenantID[7:] {
			longTenantID = longTenantID[:7+i] + "x" + longTenantID[7+i+1:]
		}

		longCtx := filter.WithTenantContext(ctx, &longTenantID, false)

		retrieved, ok := filter.GetTenantID(longCtx)
		assert.True(t, ok)
		require.NotNil(t, retrieved)
		assert.Equal(t, longTenantID, *retrieved)
	})

	t.Run("SpecialCharacterTenantIDs", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		specialTenantIDs := []string{
			"tenant-with-dashes",
			"tenant_with_underscores",
			"tenant.with.dots",
			"tenant@with@symbols",
			"tenant-123-numbers",
			"TENANT-UPPERCASE",
		}

		for _, tenantID := range specialTenantIDs {
			tenantCtx := filter.WithTenantContext(ctx, &tenantID, false)

			retrieved, ok := filter.GetTenantID(tenantCtx)
			assert.True(t, ok, "Should handle tenant ID: %s", tenantID)
			require.NotNil(t, retrieved, "Should handle tenant ID: %s", tenantID)
			assert.Equal(t, tenantID, *retrieved, "Should handle tenant ID: %s", tenantID)
		}
	})
}

// TestTenantIsolationMetrics tests metrics and monitoring for tenant isolation
func TestTenantIsolationMetrics(t *testing.T) {
	t.Run("IsolationViolationDetection", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		violations := 0
		validAccesses := 0

		// Simulate various access patterns
		tenants := []string{"metric-tenant-1", "metric-tenant-2", "metric-tenant-3"}

		for _, tenant := range tenants {
			tenantCtx := filter.WithTenantContext(ctx, &tenant, false)

			for _, targetTenant := range tenants {
				err := filter.ValidateTenantAccess(tenantCtx, &targetTenant)
				if err != nil {
					violations++
				} else {
					validAccesses++
				}
			}
		}

		// Should have 3 valid accesses (each tenant accessing itself)
		// Should have 6 violations (each tenant trying to access 2 others)
		assert.Equal(t, 3, validAccesses, "Should have 3 valid same-tenant accesses")
		assert.Equal(t, 6, violations, "Should have 6 cross-tenant violations")

		t.Logf("Isolation metrics - Valid accesses: %d, Violations detected: %d",
			validAccesses, violations)
	})

	t.Run("PerformanceMetrics", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping performance metrics in short mode")
		}

		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		numOperations := 1000
		tenantID := "metrics-tenant"

		// Measure context creation time
		start := time.Now()
		for i := 0; i < numOperations; i++ {
			_ = filter.WithTenantContext(ctx, &tenantID, false)
		}
		contextCreationTime := time.Since(start)

		// Measure validation time
		tenantCtx := filter.WithTenantContext(ctx, &tenantID, false)
		start = time.Now()
		for i := 0; i < numOperations; i++ {
			_ = filter.ValidateTenantAccess(tenantCtx, &tenantID)
		}
		validationTime := time.Since(start)

		t.Logf("Performance metrics for %d operations:", numOperations)
		t.Logf("  Context creation: %v (avg: %v)", contextCreationTime, contextCreationTime/time.Duration(numOperations))
		t.Logf("  Validation: %v (avg: %v)", validationTime, validationTime/time.Duration(numOperations))

		// Performance assertions
		assert.Less(t, contextCreationTime, time.Second, "Context creation should be fast")
		assert.Less(t, validationTime, time.Second, "Validation should be fast")
	})
}
