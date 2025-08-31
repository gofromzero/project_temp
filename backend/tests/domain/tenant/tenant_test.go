package tenant_test

import (
	"testing"
	"time"

	"github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantStatus_IsValid(t *testing.T) {
	testCases := []struct {
		name     string
		status   tenant.TenantStatus
		expected bool
	}{
		{"Active status", tenant.StatusActive, true},
		{"Suspended status", tenant.StatusSuspended, true},
		{"Disabled status", tenant.StatusDisabled, true},
		{"Invalid status", tenant.TenantStatus("invalid"), false},
		{"Empty status", tenant.TenantStatus(""), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.status.IsValid())
		})
	}
}

func TestNewTenant(t *testing.T) {
	name := "Test Tenant"
	code := "test-tenant"

	tn := tenant.NewTenant(name, code)

	assert.NotEmpty(t, tn.ID)
	assert.Equal(t, name, tn.Name)
	assert.Equal(t, code, tn.Code)
	assert.Equal(t, tenant.StatusActive, tn.Status)
	assert.Equal(t, 100, tn.Config.MaxUsers)
	assert.NotNil(t, tn.Config.Features)
	assert.WithinDuration(t, time.Now(), tn.CreatedAt, time.Second)
	assert.WithinDuration(t, time.Now(), tn.UpdatedAt, time.Second)
}

func TestTenant_Validate(t *testing.T) {
	testCases := []struct {
		name      string
		setupTenant func() *tenant.Tenant
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid tenant",
			setupTenant: func() *tenant.Tenant {
				return tenant.NewTenant("Test", "test")
			},
			expectError: false,
		},
		{
			name: "Empty name",
			setupTenant: func() *tenant.Tenant {
				return tenant.NewTenant("", "test")
			},
			expectError: true,
			errorMsg:    "tenant name is required",
		},
		{
			name: "Empty code",
			setupTenant: func() *tenant.Tenant {
				return tenant.NewTenant("Test", "")
			},
			expectError: true,
			errorMsg:    "tenant code is required",
		},
		{
			name: "Invalid status",
			setupTenant: func() *tenant.Tenant {
				tn := tenant.NewTenant("Test", "test")
				tn.Status = tenant.TenantStatus("invalid")
				return tn
			},
			expectError: true,
			errorMsg:    "invalid tenant status",
		},
		{
			name: "Invalid max users",
			setupTenant: func() *tenant.Tenant {
				tn := tenant.NewTenant("Test", "test")
				tn.Config.MaxUsers = 0
				return tn
			},
			expectError: true,
			errorMsg:    "max users must be greater than 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tn := tc.setupTenant()
			err := tn.Validate()

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTenant_StatusMethods(t *testing.T) {
	tn := tenant.NewTenant("Test", "test")

	// Test initial active status
	assert.True(t, tn.IsActive())
	assert.False(t, tn.IsSuspended())
	assert.False(t, tn.IsDisabled())

	// Test suspended status
	err := tn.UpdateStatus(tenant.StatusSuspended)
	require.NoError(t, err)
	assert.False(t, tn.IsActive())
	assert.True(t, tn.IsSuspended())
	assert.False(t, tn.IsDisabled())

	// Test disabled status
	err = tn.UpdateStatus(tenant.StatusDisabled)
	require.NoError(t, err)
	assert.False(t, tn.IsActive())
	assert.False(t, tn.IsSuspended())
	assert.True(t, tn.IsDisabled())

	// Test invalid status
	err = tn.UpdateStatus(tenant.TenantStatus("invalid"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tenant status")
}

func TestTenant_UpdateName(t *testing.T) {
	tn := tenant.NewTenant("Test", "test")
	originalUpdatedAt := tn.UpdatedAt
	
	// Add small delay to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	// Valid name update
	newName := "Updated Tenant"
	err := tn.UpdateName(newName)
	require.NoError(t, err)
	assert.Equal(t, newName, tn.Name)
	assert.True(t, tn.UpdatedAt.After(originalUpdatedAt))

	// Empty name should fail
	err = tn.UpdateName("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tenant name is required")
}

func TestTenant_SetConfig(t *testing.T) {
	tn := tenant.NewTenant("Test", "test")
	originalUpdatedAt := tn.UpdatedAt
	
	// Add small delay to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	// Valid config update
	theme := "dark"
	newConfig := tenant.TenantConfig{
		MaxUsers: 200,
		Features: []string{"feature1", "feature2"},
		Theme:    &theme,
	}

	err := tn.SetConfig(newConfig)
	require.NoError(t, err)
	assert.Equal(t, newConfig, tn.Config)
	assert.True(t, tn.UpdatedAt.After(originalUpdatedAt))

	// Invalid config (zero max users)
	invalidConfig := tenant.TenantConfig{
		MaxUsers: 0,
	}
	err = tn.SetConfig(invalidConfig)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max users must be greater than 0")
}

func TestTenant_CanCreateUsers(t *testing.T) {
	tn := tenant.NewTenant("Test", "test")
	tn.Config.MaxUsers = 5

	testCases := []struct {
		name             string
		status           tenant.TenantStatus
		currentUserCount int
		expected         bool
	}{
		{"Active with room for users", tenant.StatusActive, 3, true},
		{"Active at max users", tenant.StatusActive, 5, false},
		{"Active over max users", tenant.StatusActive, 6, false},
		{"Suspended with room", tenant.StatusSuspended, 3, false},
		{"Disabled with room", tenant.StatusDisabled, 3, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tn.Status = tc.status
			result := tn.CanCreateUsers(tc.currentUserCount)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestTenant_GetConfigJSON(t *testing.T) {
	tn := tenant.NewTenant("Test", "test")
	theme := "light"
	tn.Config = tenant.TenantConfig{
		MaxUsers: 50,
		Features: []string{"feature1"},
		Theme:    &theme,
	}

	jsonData, err := tn.GetConfigJSON()
	require.NoError(t, err)
	assert.Contains(t, string(jsonData), `"maxUsers":50`)
	assert.Contains(t, string(jsonData), `"features":["feature1"]`)
	assert.Contains(t, string(jsonData), `"theme":"light"`)
}