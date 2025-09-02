package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/gofromzero/project_temp/backend/api/handlers"
	"github.com/gofromzero/project_temp/backend/application/user"
	domainUser "github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository implements the UserRepository interface for testing
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(user *domainUser.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockRepository) GetByID(id string) (*domainUser.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockRepository) GetByUsername(tenantID *string, username string) (*domainUser.User, error) {
	args := m.Called(tenantID, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockRepository) GetByEmail(tenantID *string, email string) (*domainUser.User, error) {
	args := m.Called(tenantID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockRepository) GetByTenantID(tenantID string, offset, limit int) ([]*domainUser.User, error) {
	args := m.Called(tenantID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainUser.User), args.Error(1)
}

func (m *MockRepository) GetSystemAdmins(offset, limit int) ([]*domainUser.User, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainUser.User), args.Error(1)
}

func (m *MockRepository) Update(user *domainUser.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockRepository) Count(tenantID *string) (int64, error) {
	args := m.Called(tenantID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) ExistsByUsername(tenantID *string, username string) (bool, error) {
	args := m.Called(tenantID, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockRepository) ExistsByEmail(tenantID *string, email string) (bool, error) {
	args := m.Called(tenantID, email)
	return args.Bool(0), args.Error(1)
}

// TestUserHandler_ServiceIntegration tests the handler with actual user service
func TestUserHandler_ServiceIntegration(t *testing.T) {
	// Create mock repository
	mockRepo := &MockRepository{}
	
	// Create user service with mock repository
	userService := user.NewUserServiceWithRepository(mockRepo)
	
	// Create handler with user service
	handler := handlers.NewUserHandlerWithService(userService)
	
	// Verify handler was created successfully
	require.NotNil(t, handler)
	require.NotNil(t, handler.UserService)
}

// TestUserService_CreateUserWithMockRepository tests the full user creation flow
func TestUserService_CreateUserWithMockRepository(t *testing.T) {
	mockRepo := &MockRepository{}
	userService := user.NewUserServiceWithRepository(mockRepo)
	
	ctx := context.Background()
	tenantID := "tenant-123"
	
	// Setup mock expectations
	mockRepo.On("ExistsByUsername", &tenantID, "testuser").Return(false, nil)
	mockRepo.On("ExistsByEmail", &tenantID, "test@example.com").Return(false, nil)
	mockRepo.On("Create", mock.AnythingOfType("*user.User")).Return(nil)
	
	// Test user creation
	req := user.CreateUserRequest{
		TenantID:  &tenantID,
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
	}
	
	response, err := userService.CreateUser(ctx, req)
	
	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "testuser", response.Username)
	assert.Equal(t, "test@example.com", response.Email)
	assert.Equal(t, &tenantID, response.TenantID)
	assert.Equal(t, "active", response.Status)
	
	// Verify all mock expectations were met
	mockRepo.AssertExpectations(t)
}

// TestUserService_GetUsers tests user listing with pagination
func TestUserService_GetUsers(t *testing.T) {
	mockRepo := &MockRepository{}
	userService := user.NewUserServiceWithRepository(mockRepo)
	
	ctx := context.Background()
	tenantID := "tenant-123"
	
	// Mock users data
	mockUsers := []*domainUser.User{
		{
			ID:        "user-1",
			TenantID:  &tenantID,
			Username:  "user1",
			Email:     "user1@example.com",
			FirstName: "User",
			LastName:  "One",
			Status:    domainUser.StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "user-2",
			TenantID:  &tenantID,
			Username:  "user2",
			Email:     "user2@example.com",
			FirstName: "User",
			LastName:  "Two",
			Status:    domainUser.StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	
	// Setup mock expectations
	mockRepo.On("GetByTenantID", tenantID, 0, 10).Return(mockUsers, nil)
	mockRepo.On("Count", &tenantID).Return(int64(2), nil)
	
	// Test user listing
	req := user.ListUsersRequest{
		TenantID: &tenantID,
		Page:     1,
		Limit:    10,
	}
	
	response, err := userService.GetUsers(ctx, req)
	
	// Assertions
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Len(t, response.Users, 2)
	assert.Equal(t, 1, response.Pagination.Page)
	assert.Equal(t, 10, response.Pagination.Limit)
	assert.Equal(t, 2, response.Pagination.Total)
	assert.Equal(t, 1, response.Pagination.Pages)
	
	// Verify user data
	assert.Equal(t, "user1", response.Users[0].Username)
	assert.Equal(t, "user2", response.Users[1].Username)
	
	mockRepo.AssertExpectations(t)
}

// TestUserService_ErrorHandling tests error scenarios
func TestUserService_ErrorHandling(t *testing.T) {
	mockRepo := &MockRepository{}
	userService := user.NewUserServiceWithRepository(mockRepo)
	
	ctx := context.Background()
	tenantID := "tenant-123"
	
	t.Run("Duplicate username error", func(t *testing.T) {
		// Setup mock to return duplicate username
		mockRepo.On("ExistsByUsername", &tenantID, "duplicate").Return(true, nil)
		
		req := user.CreateUserRequest{
			TenantID:  &tenantID,
			Username:  "duplicate",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}
		
		_, err := userService.CreateUser(ctx, req)
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "username already exists")
		
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Domain password validation", func(t *testing.T) {
		// Test domain-level password validation directly
		err := domainUser.ValidatePassword("weak")
		
		require.Error(t, err)
		assert.Contains(t, err.Error(), "password")
	})
}