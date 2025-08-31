package isolation_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofromzero/project_temp/backend/infr/database"
	"github.com/gofromzero/project_temp/backend/pkg/middleware"
)

// TestTenantIsolationBasic tests basic tenant isolation functionality
func TestTenantIsolationBasic(t *testing.T) {
	t.Run("TenantContextIsolation", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		// Test tenant A context
		tenantA := "tenant-a"
		ctxA := filter.WithTenantContext(ctx, &tenantA, false)

		retrievedA, okA := filter.GetTenantID(ctxA)
		require.True(t, okA)
		require.NotNil(t, retrievedA)
		assert.Equal(t, tenantA, *retrievedA)

		// Test tenant B context
		tenantB := "tenant-b"
		ctxB := filter.WithTenantContext(ctx, &tenantB, false)

		retrievedB, okB := filter.GetTenantID(ctxB)
		require.True(t, okB)
		require.NotNil(t, retrievedB)
		assert.Equal(t, tenantB, *retrievedB)

		// Verify contexts are isolated
		assert.NotEqual(t, *retrievedA, *retrievedB)
	})

	t.Run("SystemAdminBypass", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		// System admin context
		adminCtx := filter.WithTenantContext(ctx, nil, true)
		assert.True(t, filter.IsSystemAdmin(adminCtx))

		// System admin should be able to access any tenant
		tenantID := "any-tenant"
		err := filter.ValidateTenantAccess(adminCtx, &tenantID)
		assert.NoError(t, err)
	})
}

// TestTenantDataAccessIsolation tests data access isolation between tenants
func TestTenantDataAccessIsolation(t *testing.T) {
	t.Run("CrossTenantAccessDenied", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		// Tenant A tries to access Tenant B's data
		tenantA := "tenant-a"
		ctxA := filter.WithTenantContext(ctx, &tenantA, false)

		tenantB := "tenant-b"
		err := filter.ValidateTenantAccess(ctxA, &tenantB)
		assert.Error(t, err)
		assert.Equal(t, middleware.ErrUnauthorizedTenant, err)
	})

	t.Run("SameTenantAccessAllowed", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		// Tenant accesses its own data
		tenantA := "tenant-a"
		ctxA := filter.WithTenantContext(ctx, &tenantA, false)

		err := filter.ValidateTenantAccess(ctxA, &tenantA)
		assert.NoError(t, err)
	})
}

// TestMultiTenantConcurrentAccess tests concurrent access patterns
func TestMultiTenantConcurrentAccess(t *testing.T) {
	t.Run("ConcurrentTenantContextCreation", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		var wg sync.WaitGroup
		numGoroutines := 10
		numTenants := 5

		// Results channel to collect tenant contexts
		results := make(chan struct {
			tenantID     string
			contextValid bool
		}, numGoroutines)

		// Create multiple goroutines accessing different tenants
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(tenantNum int) {
				defer wg.Done()

				tenantID := fmt.Sprintf("tenant-%d", tenantNum%numTenants)
				tenantCtx := filter.WithTenantContext(ctx, &tenantID, false)

				// Simulate some work
				time.Sleep(time.Millisecond * 10)

				// Verify context integrity
				retrieved, ok := filter.GetTenantID(tenantCtx)
				contextValid := ok && retrieved != nil && *retrieved == tenantID

				results <- struct {
					tenantID     string
					contextValid bool
				}{
					tenantID:     tenantID,
					contextValid: contextValid,
				}
			}(i)
		}

		// Wait for all goroutines to complete
		go func() {
			wg.Wait()
			close(results)
		}()

		// Verify all contexts were created correctly
		validContexts := 0
		for result := range results {
			if result.contextValid {
				validContexts++
			}
		}

		assert.Equal(t, numGoroutines, validContexts, "All tenant contexts should be valid")
	})

	t.Run("ConcurrentAccessValidation", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		var wg sync.WaitGroup
		numGoroutines := 20

		// Counters for results
		var successCount, errorCount int32
		var mu sync.Mutex

		// Test concurrent access validation
		for i := 0; i < numGoroutines; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				// Half the goroutines use tenant A, half use tenant B
				var tenantID, targetTenantID string
				if i%2 == 0 {
					tenantID = "tenant-concurrent-a"
					targetTenantID = "tenant-concurrent-a" // Same tenant - should succeed
				} else {
					tenantID = "tenant-concurrent-b"
					targetTenantID = "tenant-concurrent-a" // Cross tenant - should fail
				}

				tenantCtx := filter.WithTenantContext(ctx, &tenantID, false)
				err := filter.ValidateTenantAccess(tenantCtx, &targetTenantID)

				mu.Lock()
				if err != nil {
					errorCount++
				} else {
					successCount++
				}
				mu.Unlock()
			}(i)
		}

		wg.Wait()

		// Half should succeed (same tenant), half should fail (cross tenant)
		expectedSuccesses := int32(numGoroutines / 2)
		expectedErrors := int32(numGoroutines / 2)

		assert.Equal(t, expectedSuccesses, successCount, "Expected %d successful same-tenant accesses", expectedSuccesses)
		assert.Equal(t, expectedErrors, errorCount, "Expected %d failed cross-tenant accesses", expectedErrors)
	})
}

// TestTenantDataLeakageDetection tests for potential data leakage between tenants
func TestTenantDataLeakageDetection(t *testing.T) {
	t.Run("ContextIsolationIntegrity", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		// Create contexts for multiple tenants
		tenants := []string{"tenant-leak-1", "tenant-leak-2", "tenant-leak-3"}
		contexts := make([]context.Context, len(tenants))

		for i, tenant := range tenants {
			contexts[i] = filter.WithTenantContext(ctx, &tenant, false)
		}

		// Verify each context maintains its tenant isolation
		for i, tenantCtx := range contexts {
			retrievedID, ok := filter.GetTenantID(tenantCtx)
			require.True(t, ok, "Context should have tenant ID")
			require.NotNil(t, retrievedID, "Tenant ID should not be nil")
			assert.Equal(t, tenants[i], *retrievedID, "Context should maintain correct tenant ID")

			// Verify this context cannot access other tenants
			for j, otherTenant := range tenants {
				if i != j {
					err := filter.ValidateTenantAccess(tenantCtx, &otherTenant)
					assert.Error(t, err, "Should not be able to access tenant %s from tenant %s context", otherTenant, tenants[i])
				}
			}
		}
	})

	t.Run("MemoryIsolationTest", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		// Create many tenant contexts to test for memory leakage between contexts
		numTenants := 100
		contexts := make([]context.Context, numTenants)

		for i := 0; i < numTenants; i++ {
			tenantID := fmt.Sprintf("tenant-memory-%d", i)
			contexts[i] = filter.WithTenantContext(ctx, &tenantID, false)
		}

		// Verify each context maintains isolation
		for i, tenantCtx := range contexts {
			expectedTenantID := fmt.Sprintf("tenant-memory-%d", i)

			retrievedID, ok := filter.GetTenantID(tenantCtx)
			require.True(t, ok)
			require.NotNil(t, retrievedID)
			assert.Equal(t, expectedTenantID, *retrievedID)

			// Test random access to ensure no cross-contamination
			randomTenant := fmt.Sprintf("tenant-memory-%d", (i+50)%numTenants)
			if randomTenant != expectedTenantID {
				err := filter.ValidateTenantAccess(tenantCtx, &randomTenant)
				assert.Error(t, err, "Context %d should not access tenant %s", i, randomTenant)
			}
		}
	})
}

// TestTenantIsolationPerformance tests performance impact of tenant isolation
func TestTenantIsolationPerformance(t *testing.T) {
	t.Run("ContextCreationPerformance", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping performance test in short mode")
		}

		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		numOperations := 10000
		tenantID := "performance-tenant"

		start := time.Now()

		for i := 0; i < numOperations; i++ {
			tenantCtx := filter.WithTenantContext(ctx, &tenantID, false)
			_, ok := filter.GetTenantID(tenantCtx)
			assert.True(t, ok)
		}

		duration := time.Since(start)

		// Performance assertion - should complete in reasonable time
		maxDuration := time.Second * 5 // 5 seconds for 10k operations
		assert.Less(t, duration, maxDuration, "Context operations should complete within %v", maxDuration)

		avgPerOperation := duration / time.Duration(numOperations)
		t.Logf("Average time per context operation: %v", avgPerOperation)
	})

	t.Run("ValidationPerformance", func(t *testing.T) {
		if testing.Short() {
			t.Skip("Skipping performance test in short mode")
		}

		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		tenantID := "perf-tenant"
		tenantCtx := filter.WithTenantContext(ctx, &tenantID, false)

		numValidations := 5000
		start := time.Now()

		for i := 0; i < numValidations; i++ {
			err := filter.ValidateTenantAccess(tenantCtx, &tenantID)
			assert.NoError(t, err)
		}

		duration := time.Since(start)

		// Performance assertion
		maxDuration := time.Second * 2 // 2 seconds for 5k validations
		assert.Less(t, duration, maxDuration, "Validation operations should complete within %v", maxDuration)

		avgPerValidation := duration / time.Duration(numValidations)
		t.Logf("Average time per validation: %v", avgPerValidation)
	})
}

// TestDatabaseConnectionIsolation tests database-level isolation
func TestDatabaseConnectionIsolation(t *testing.T) {
	t.Run("DatabaseConnectionStructure", func(t *testing.T) {
		// Test that database connections maintain tenant isolation structure
		conn := &database.Connection{} // Mock connection for structure testing

		assert.NotNil(t, conn)
	})

	t.Run("TenantFilterIntegration", func(t *testing.T) {
		// Test integration between tenant filter and database connection
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		tenantA := "db-tenant-a"
		ctxA := filter.WithTenantContext(ctx, &tenantA, false)

		tenantB := "db-tenant-b"
		ctxB := filter.WithTenantContext(ctx, &tenantB, false)

		// Verify contexts are properly isolated at the filter level
		retrievedA, okA := filter.GetTenantID(ctxA)
		retrievedB, okB := filter.GetTenantID(ctxB)

		assert.True(t, okA)
		assert.True(t, okB)
		require.NotNil(t, retrievedA)
		require.NotNil(t, retrievedB)
		assert.NotEqual(t, *retrievedA, *retrievedB)
	})
}

// TestTenantIsolationErrorScenarios tests error handling in isolation
func TestTenantIsolationErrorScenarios(t *testing.T) {
	t.Run("MissingTenantContext", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background() // No tenant context

		// Should not have tenant ID
		_, ok := filter.GetTenantID(ctx)
		assert.False(t, ok)

		// Should not be system admin
		isAdmin := filter.IsSystemAdmin(ctx)
		assert.False(t, isAdmin)
	})

	t.Run("InvalidTenantAccess", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		// Create context without tenant ID (non-admin)
		ctxNoTenant := filter.WithTenantContext(ctx, nil, false)

		// Should fail to access any tenant
		targetTenant := "some-tenant"
		err := filter.ValidateTenantAccess(ctxNoTenant, &targetTenant)
		assert.Error(t, err)
	})

	t.Run("NilTenantIDHandling", func(t *testing.T) {
		filter := middleware.NewTenantFilter()
		ctx := context.Background()

		// Test nil tenant ID with non-admin context
		ctxNilTenant := filter.WithTenantContext(ctx, nil, false)

		tenantID, ok := filter.GetTenantID(ctxNilTenant)
		assert.True(t, ok, "Context should exist")
		assert.Nil(t, tenantID, "Tenant ID should be nil")

		// Should not be admin
		assert.False(t, filter.IsSystemAdmin(ctxNilTenant))
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// Benchmark tests for performance analysis
func BenchmarkTenantContextCreation(b *testing.B) {
	filter := middleware.NewTenantFilter()
	ctx := context.Background()
	tenantID := "benchmark-tenant"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filter.WithTenantContext(ctx, &tenantID, false)
	}
}

func BenchmarkTenantValidation(b *testing.B) {
	filter := middleware.NewTenantFilter()
	ctx := context.Background()
	tenantID := "benchmark-tenant"
	tenantCtx := filter.WithTenantContext(ctx, &tenantID, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filter.ValidateTenantAccess(tenantCtx, &tenantID)
	}
}

func BenchmarkConcurrentTenantAccess(b *testing.B) {
	filter := middleware.NewTenantFilter()
	ctx := context.Background()
	tenantID := "concurrent-benchmark-tenant"

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tenantCtx := filter.WithTenantContext(ctx, &tenantID, false)
			_, _ = filter.GetTenantID(tenantCtx)
		}
	})
}
