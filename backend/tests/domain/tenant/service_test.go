package tenant_test

import (
	"errors"
	"testing"

	"github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockTenantRepository is a mock implementation of TenantRepository
type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) Create(tenant *tenant.Tenant) error {
	args := m.Called(tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) GetByID(id string) (*tenant.Tenant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tenant.Tenant), args.Error(1)
}

func (m *MockTenantRepository) GetByCode(code string) (*tenant.Tenant, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tenant.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Update(tenant *tenant.Tenant) error {
	args := m.Called(tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTenantRepository) List(offset, limit int) ([]*tenant.Tenant, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]*tenant.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func TestService_CreateTenant(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := tenant.NewService(mockRepo)

	t.Run("Successful creation", func(t *testing.T) {
		// Setup: code doesn't exist
		mockRepo.On("GetByCode", "test-code").Return(nil, errors.New("not found")).Once()
		mockRepo.On("Create", mock.AnythingOfType("*tenant.Tenant")).Return(nil).Once()

		result, err := service.CreateTenant("Test Tenant", "test-code", nil)

		require.NoError(t, err)
		assert.Equal(t, "Test Tenant", result.Name)
		assert.Equal(t, "test-code", result.Code)
		assert.Equal(t, tenant.StatusActive, result.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Code already exists", func(t *testing.T) {
		existingTenant := tenant.NewTenant("Existing", "existing-code")
		mockRepo.On("GetByCode", "existing-code").Return(existingTenant, nil).Once()

		result, err := service.CreateTenant("Test Tenant", "existing-code", nil)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("With custom config", func(t *testing.T) {
		config := &tenant.TenantConfig{
			MaxUsers: 50,
			Features: []string{"feature1"},
		}

		mockRepo.On("GetByCode", "test-code").Return(nil, errors.New("not found")).Once()
		mockRepo.On("Create", mock.MatchedBy(func(t *tenant.Tenant) bool {
			return t.Config.MaxUsers == 50 && len(t.Config.Features) == 1
		})).Return(nil).Once()

		result, err := service.CreateTenant("Test Tenant", "test-code", config)

		require.NoError(t, err)
		assert.Equal(t, 50, result.Config.MaxUsers)
		assert.Equal(t, []string{"feature1"}, result.Config.Features)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid config", func(t *testing.T) {
		config := &tenant.TenantConfig{
			MaxUsers: 0, // Invalid
		}

		mockRepo.On("GetByCode", "test-code").Return(nil, errors.New("not found")).Once()

		result, err := service.CreateTenant("Test Tenant", "test-code", config)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid tenant config")
		mockRepo.AssertExpectations(t)
	})
}

func TestService_UpdateTenant(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := tenant.NewService(mockRepo)

	t.Run("Successful update", func(t *testing.T) {
		existingTenant := tenant.NewTenant("Old Name", "test-code")
		mockRepo.On("GetByID", "tenant-id").Return(existingTenant, nil).Once()
		mockRepo.On("Update", mock.AnythingOfType("*tenant.Tenant")).Return(nil).Once()

		newName := "New Name"
		updates := tenant.TenantUpdates{
			Name: &newName,
		}

		result, err := service.UpdateTenant("tenant-id", updates)

		require.NoError(t, err)
		assert.Equal(t, "New Name", result.Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Tenant not found", func(t *testing.T) {
		mockRepo.On("GetByID", "nonexistent-id").Return(nil, errors.New("not found")).Once()

		updates := tenant.TenantUpdates{}
		result, err := service.UpdateTenant("nonexistent-id", updates)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "tenant not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("Update status", func(t *testing.T) {
		existingTenant := tenant.NewTenant("Test", "test-code")
		mockRepo.On("GetByID", "tenant-id").Return(existingTenant, nil).Once()
		mockRepo.On("Update", mock.MatchedBy(func(t *tenant.Tenant) bool {
			return t.Status == tenant.StatusSuspended
		})).Return(nil).Once()

		status := tenant.StatusSuspended
		updates := tenant.TenantUpdates{
			Status: &status,
		}

		result, err := service.UpdateTenant("tenant-id", updates)

		require.NoError(t, err)
		assert.Equal(t, tenant.StatusSuspended, result.Status)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_GetTenant(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := tenant.NewService(mockRepo)

	t.Run("Successful retrieval", func(t *testing.T) {
		expectedTenant := tenant.NewTenant("Test", "test-code")
		mockRepo.On("GetByID", "tenant-id").Return(expectedTenant, nil).Once()

		result, err := service.GetTenant("tenant-id")

		require.NoError(t, err)
		assert.Equal(t, expectedTenant, result)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Tenant not found", func(t *testing.T) {
		mockRepo.On("GetByID", "nonexistent-id").Return(nil, errors.New("not found")).Once()

		result, err := service.GetTenant("nonexistent-id")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "tenant not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestService_StatusManagement(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := tenant.NewService(mockRepo)

	testCases := []struct {
		name           string
		method         func(string) (*tenant.Tenant, error)
		expectedStatus tenant.TenantStatus
	}{
		{"Activate", service.ActivateTenant, tenant.StatusActive},
		{"Suspend", service.SuspendTenant, tenant.StatusSuspended},
		{"Disable", service.DisableTenant, tenant.StatusDisabled},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			existingTenant := tenant.NewTenant("Test", "test-code")
			mockRepo.On("GetByID", "tenant-id").Return(existingTenant, nil).Once()
			mockRepo.On("Update", mock.MatchedBy(func(t *tenant.Tenant) bool {
				return t.Status == tc.expectedStatus
			})).Return(nil).Once()

			result, err := tc.method("tenant-id")

			require.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, result.Status)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_ValidateConfig(t *testing.T) {
	service := tenant.NewService(nil) // No repo needed for validation

	testCases := []struct {
		name        string
		config      tenant.TenantConfig
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid config",
			config: tenant.TenantConfig{
				MaxUsers: 100,
				Features: []string{"feature1"},
			},
			expectError: false,
		},
		{
			name: "Invalid max users",
			config: tenant.TenantConfig{
				MaxUsers: 0,
			},
			expectError: true,
			errorMsg:    "max users must be greater than 0",
		},
		{
			name: "Invalid domain too short",
			config: tenant.TenantConfig{
				MaxUsers: 100,
				Domain:   &[]string{"ab"}[0],
			},
			expectError: true,
			errorMsg:    "domain must be between 3 and 253 characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := service.ValidateConfig(tc.config)

			if tc.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_CanTenantCreateUsers(t *testing.T) {
	mockRepo := new(MockTenantRepository)
	service := tenant.NewService(mockRepo)

	t.Run("Can create users", func(t *testing.T) {
		tenant := tenant.NewTenant("Test", "test")
		tenant.Config.MaxUsers = 10
		mockRepo.On("GetByID", "tenant-id").Return(tenant, nil).Once()

		canCreate, err := service.CanTenantCreateUsers("tenant-id", 5)

		require.NoError(t, err)
		assert.True(t, canCreate)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Cannot create users - at limit", func(t *testing.T) {
		tenant := tenant.NewTenant("Test", "test")
		tenant.Config.MaxUsers = 10
		mockRepo.On("GetByID", "tenant-id").Return(tenant, nil).Once()

		canCreate, err := service.CanTenantCreateUsers("tenant-id", 10)

		require.NoError(t, err)
		assert.False(t, canCreate)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Tenant not found", func(t *testing.T) {
		mockRepo.On("GetByID", "nonexistent-id").Return(nil, errors.New("not found")).Once()

		canCreate, err := service.CanTenantCreateUsers("nonexistent-id", 5)

		require.Error(t, err)
		assert.False(t, canCreate)
		mockRepo.AssertExpectations(t)
	})
}