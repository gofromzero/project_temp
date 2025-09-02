package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/gofromzero/project_temp/backend/application/user"
	domainUser "github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserRepository is a mock implementation of UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domainUser.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id string) (*domainUser.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(tenantID *string, username string) (*domainUser.User, error) {
	args := m.Called(tenantID, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(tenantID *string, email string) (*domainUser.User, error) {
	args := m.Called(tenantID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainUser.User), args.Error(1)
}

func (m *MockUserRepository) GetByTenantID(tenantID string, offset, limit int) ([]*domainUser.User, error) {
	args := m.Called(tenantID, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainUser.User), args.Error(1)
}

func (m *MockUserRepository) GetSystemAdmins(offset, limit int) ([]*domainUser.User, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainUser.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *domainUser.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) Count(tenantID *string) (int64, error) {
	args := m.Called(tenantID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsername(tenantID *string, username string) (bool, error) {
	args := m.Called(tenantID, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(tenantID *string, email string) (bool, error) {
	args := m.Called(tenantID, email)
	return args.Bool(0), args.Error(1)
}

// Test helper function to create a user service with mock repository
func createTestUserService(mockRepo *MockUserRepository) *user.UserService {
	return user.NewUserServiceWithRepository(mockRepo)
}

func TestUserService_CreateUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := createTestUserService(mockRepo)

	t.Run("Successful user creation", func(t *testing.T) {
		req := user.CreateUserRequest{
			TenantID:  stringPtr("tenant-1"),
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}

		// Setup mocks
		mockRepo.On("ExistsByUsername", req.TenantID, req.Username).Return(false, nil)
		mockRepo.On("ExistsByEmail", req.TenantID, req.Email).Return(false, nil)
		mockRepo.On("Create", mock.AnythingOfType("*user.User")).Return(nil)

		// Execute
		response, err := service.CreateUser(ctx, req)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, req.Username, response.Username)
		assert.Equal(t, req.Email, response.Email)
		assert.Equal(t, req.FirstName, response.FirstName)
		assert.Equal(t, req.LastName, response.LastName)
		assert.Equal(t, "active", response.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Username already exists", func(t *testing.T) {
		req := user.CreateUserRequest{
			TenantID:  stringPtr("tenant-1"),
			Username:  "existinguser",
			Email:     "test@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}

		// Setup mocks
		mockRepo.On("ExistsByUsername", req.TenantID, req.Username).Return(true, nil)

		// Execute
		response, err := service.CreateUser(ctx, req)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "username already exists in this tenant", err.Error())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Email already exists", func(t *testing.T) {
		req := user.CreateUserRequest{
			TenantID:  stringPtr("tenant-1"),
			Username:  "testuser",
			Email:     "existing@example.com",
			Password:  "password123",
			FirstName: "Test",
			LastName:  "User",
		}

		// Setup mocks
		mockRepo.On("ExistsByUsername", req.TenantID, req.Username).Return(false, nil)
		mockRepo.On("ExistsByEmail", req.TenantID, req.Email).Return(true, nil)

		// Execute
		response, err := service.CreateUser(ctx, req)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "email already exists in this tenant", err.Error())

		mockRepo.AssertExpectations(t)
	})

	t.Run("Invalid password", func(t *testing.T) {
		req := user.CreateUserRequest{
			TenantID:  stringPtr("tenant-1"),
			Username:  "testuser",
			Email:     "test@example.com",
			Password:  "weak", // Invalid password
			FirstName: "Test",
			LastName:  "User",
		}

		// Setup mocks
		mockRepo.On("ExistsByUsername", req.TenantID, req.Username).Return(false, nil)
		mockRepo.On("ExistsByEmail", req.TenantID, req.Email).Return(false, nil)

		// Execute
		response, err := service.CreateUser(ctx, req)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		var validationErr *domainUser.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "MIN_LENGTH", validationErr.Code)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUsers(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := createTestUserService(mockRepo)

	t.Run("Get tenant users successfully", func(t *testing.T) {
		tenantID := "tenant-1"
		req := user.ListUsersRequest{
			TenantID: &tenantID,
			Page:     1,
			Limit:    10,
		}

		users := []*domainUser.User{
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
		}

		// Setup mocks
		mockRepo.On("GetByTenantID", tenantID, 0, 10).Return(users, nil)
		mockRepo.On("Count", &tenantID).Return(int64(1), nil)

		// Execute
		response, err := service.GetUsers(ctx, req)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Users, 1)
		assert.Equal(t, users[0].Username, response.Users[0].Username)
		assert.Equal(t, 1, response.Pagination.Page)
		assert.Equal(t, 10, response.Pagination.Limit)
		assert.Equal(t, 1, response.Pagination.Total)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Get system admins successfully", func(t *testing.T) {
		req := user.ListUsersRequest{
			TenantID: nil, // System admins
			Page:     1,
			Limit:    10,
		}

		users := []*domainUser.User{
			{
				ID:        "admin-1",
				TenantID:  nil,
				Username:  "admin",
				Email:     "admin@example.com",
				FirstName: "Admin",
				LastName:  "User",
				Status:    domainUser.StatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}

		// Setup mocks
		mockRepo.On("GetSystemAdmins", 0, 10).Return(users, nil)
		mockRepo.On("Count", (*string)(nil)).Return(int64(1), nil)

		// Execute
		response, err := service.GetUsers(ctx, req)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Len(t, response.Users, 1)
		assert.Equal(t, users[0].Username, response.Users[0].Username)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_GetUserByID(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := createTestUserService(mockRepo)

	t.Run("Get user successfully", func(t *testing.T) {
		userID := "user-1"
		tenantID := "tenant-1"
		requesterTenantID := &tenantID

		testUser := &domainUser.User{
			ID:        userID,
			TenantID:  &tenantID,
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			Status:    domainUser.StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Setup mocks
		mockRepo.On("GetByID", userID).Return(testUser, nil)

		// Execute
		response, err := service.GetUserByID(ctx, userID, requesterTenantID)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, testUser.Username, response.Username)

		mockRepo.AssertExpectations(t)
	})

	t.Run("User not found", func(t *testing.T) {
		userID := "nonexistent-user"
		tenantID := "tenant-1"
		requesterTenantID := &tenantID

		// Setup mocks
		mockRepo.On("GetByID", userID).Return(nil, errors.New("user not found"))

		// Execute
		response, err := service.GetUserByID(ctx, userID, requesterTenantID)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to get user")

		mockRepo.AssertExpectations(t)
	})

	t.Run("Access denied - different tenant", func(t *testing.T) {
		userID := "user-1"
		userTenantID := "tenant-1"
		requesterTenantID := stringPtr("tenant-2")

		testUser := &domainUser.User{
			ID:       userID,
			TenantID: &userTenantID,
		}

		// Setup mocks
		mockRepo.On("GetByID", userID).Return(testUser, nil)

		// Execute
		response, err := service.GetUserByID(ctx, userID, requesterTenantID)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, "access denied: user not in your tenant", err.Error())

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := createTestUserService(mockRepo)

	t.Run("Update user successfully", func(t *testing.T) {
		userID := "user-1"
		tenantID := "tenant-1"
		requesterTenantID := &tenantID

		testUser := &domainUser.User{
			ID:        userID,
			TenantID:  &tenantID,
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			Status:    domainUser.StatusActive,
		}

		req := user.UpdateUserRequest{
			FirstName: stringPtr("UpdatedFirst"),
			LastName:  stringPtr("UpdatedLast"),
		}

		// Setup mocks
		mockRepo.On("GetByID", userID).Return(testUser, nil)
		mockRepo.On("Update", mock.AnythingOfType("*user.User")).Return(nil)

		// Execute
		response, err := service.UpdateUser(ctx, userID, req, requesterTenantID)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, "UpdatedFirst", response.FirstName)
		assert.Equal(t, "UpdatedLast", response.LastName)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Update email with uniqueness check", func(t *testing.T) {
		userID := "user-1"
		tenantID := "tenant-1"
		requesterTenantID := &tenantID

		testUser := &domainUser.User{
			ID:       userID,
			TenantID: &tenantID,
			Email:    "old@example.com",
		}

		req := user.UpdateUserRequest{
			Email: stringPtr("new@example.com"),
		}

		// Setup mocks
		mockRepo.On("GetByID", userID).Return(testUser, nil)
		mockRepo.On("ExistsByEmail", &tenantID, "new@example.com").Return(false, nil)
		mockRepo.On("Update", mock.AnythingOfType("*user.User")).Return(nil)

		// Execute
		response, err := service.UpdateUser(ctx, userID, req, requesterTenantID)

		// Assertions
		require.NoError(t, err)
		assert.NotNil(t, response)

		mockRepo.AssertExpectations(t)
	})
}

func TestUserService_DeleteUser(t *testing.T) {
	ctx := context.Background()
	mockRepo := &MockUserRepository{}
	service := createTestUserService(mockRepo)

	t.Run("Delete user successfully", func(t *testing.T) {
		userID := "user-1"
		tenantID := "tenant-1"
		requesterTenantID := &tenantID

		testUser := &domainUser.User{
			ID:       userID,
			TenantID: &tenantID,
			Status:   domainUser.StatusActive,
		}

		// Setup mocks
		mockRepo.On("GetByID", userID).Return(testUser, nil)
		mockRepo.On("Delete", userID).Return(nil)

		// Execute
		err := service.DeleteUser(ctx, userID, requesterTenantID)

		// Assertions
		require.NoError(t, err)

		mockRepo.AssertExpectations(t)
	})

	t.Run("Cannot delete system admin", func(t *testing.T) {
		userID := "admin-1"
		requesterTenantID := (*string)(nil) // System admin requester

		testUser := &domainUser.User{
			ID:       userID,
			TenantID: nil, // System admin
			Status:   domainUser.StatusActive,
		}

		// Setup mocks
		mockRepo.On("GetByID", userID).Return(testUser, nil)

		// Execute
		err := service.DeleteUser(ctx, userID, requesterTenantID)

		// Assertions
		require.Error(t, err)
		var validationErr *domainUser.ValidationError
		require.ErrorAs(t, err, &validationErr)
		assert.Equal(t, "SYSTEM_ADMIN_UNDELETABLE", validationErr.Code)

		mockRepo.AssertExpectations(t)
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}