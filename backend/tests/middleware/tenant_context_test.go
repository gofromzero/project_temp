package middleware_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofromzero/project_temp/backend/pkg/middleware"
)

func TestTenantFilter_ContextOperations(t *testing.T) {
	filter := middleware.NewTenantFilter()

	t.Run("WithTenantContext and GetTenantID", func(t *testing.T) {
		ctx := context.Background()
		tenantID := "tenant123"

		// Add tenant context
		newCtx := filter.WithTenantContext(ctx, &tenantID, false)

		// Retrieve tenant ID
		retrievedID, ok := filter.GetTenantID(newCtx)
		assert.True(t, ok)
		require.NotNil(t, retrievedID)
		assert.Equal(t, tenantID, *retrievedID)
	})

	t.Run("WithTenantContext and IsSystemAdmin", func(t *testing.T) {
		ctx := context.Background()

		// Add system admin context
		newCtx := filter.WithTenantContext(ctx, nil, true)

		// Check system admin status
		isAdmin := filter.IsSystemAdmin(newCtx)
		assert.True(t, isAdmin)
	})

	t.Run("Empty context", func(t *testing.T) {
		ctx := context.Background()

		// Check empty context
		tenantID, ok := filter.GetTenantID(ctx)
		assert.False(t, ok)
		assert.Nil(t, tenantID)

		isAdmin := filter.IsSystemAdmin(ctx)
		assert.False(t, isAdmin)
	})
}

func TestTenantFilter_ValidateTenantAccess(t *testing.T) {
	filter := middleware.NewTenantFilter()

	testCases := []struct {
		name                 string
		contextTenantID      *string
		contextIsSystemAdmin bool
		targetTenantID       *string
		expectError          bool
	}{
		{
			name:                 "System admin can access any tenant",
			contextTenantID:      nil,
			contextIsSystemAdmin: true,
			targetTenantID:       stringPtr("any-tenant"),
			expectError:          false,
		},
		{
			name:                 "User can access own tenant",
			contextTenantID:      stringPtr("tenant123"),
			contextIsSystemAdmin: false,
			targetTenantID:       stringPtr("tenant123"),
			expectError:          false,
		},
		{
			name:                 "User cannot access different tenant",
			contextTenantID:      stringPtr("tenant123"),
			contextIsSystemAdmin: false,
			targetTenantID:       stringPtr("tenant456"),
			expectError:          true,
		},
		{
			name:                 "Missing context tenant ID",
			contextTenantID:      nil,
			contextIsSystemAdmin: false,
			targetTenantID:       stringPtr("tenant123"),
			expectError:          true,
		},
		{
			name:                 "Nil target tenant ID with non-admin context",
			contextTenantID:      stringPtr("tenant123"),
			contextIsSystemAdmin: false,
			targetTenantID:       nil,
			expectError:          true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = filter.WithTenantContext(ctx, tc.contextTenantID, tc.contextIsSystemAdmin)

			err := filter.ValidateTenantAccess(ctx, tc.targetTenantID)

			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTenantFilter_TableClassification(t *testing.T) {
	// We'll need to create helper methods or export the private methods for testing
	// For now, let's test the functionality through public methods

	t.Run("Multi-tenant table filtering", func(t *testing.T) {
		// This would test isMultiTenantTable indirectly through CreateTenantAwareModel
		// But we need database connection for that, so we'll skip for now
		t.Skip("Requires database connection")
	})

	t.Run("System-only table access", func(t *testing.T) {
		// This would test isSystemOnlyTable indirectly
		t.Skip("Requires database connection")
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
