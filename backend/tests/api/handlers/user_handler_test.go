package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofromzero/project_temp/backend/api/handlers"
	"github.com/gofromzero/project_temp/backend/api/middleware"
	"github.com/gofromzero/project_temp/backend/application/user"
	domainTenant "github.com/gofromzero/project_temp/backend/domain/tenant"
	domainUser "github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserService is a mock implementation of UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req user.CreateUserRequest) (*user.UserResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.UserResponse), args.Error(1)
}

func (m *MockUserService) GetUsers(ctx context.Context, req user.ListUsersRequest) (*user.ListUsersResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.ListUsersResponse), args.Error(1)
}

func (m *MockUserService) GetUserByID(ctx context.Context, userID string, requesterTenantID *string) (*user.UserResponse, error) {
	args := m.Called(ctx, userID, requesterTenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.UserResponse), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, userID string, req user.UpdateUserRequest, requesterTenantID *string) (*user.UserResponse, error) {
	args := m.Called(ctx, userID, req, requesterTenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.UserResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID string, requesterTenantID *string) error {
	args := m.Called(ctx, userID, requesterTenantID)
	return args.Error(0)
}

// Test helper function to create HTTP request with authenticated context
func createAuthenticatedRequest(method, url string, body interface{}, user *domainUser.User, tenant *domainTenant.Tenant) (*ghttp.Request, error) {
	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	// Create HTTP request
	httpReq := httptest.NewRequest(method, url, bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")

	// Create GF request wrapper (simplified for testing)
	gfReq := &ghttp.Request{}
	gfReq.Request = httpReq

	// Add authentication context
	ctx := httpReq.Context()
	if user != nil {
		ctx = context.WithValue(ctx, middleware.UserContextKey, user)
	}
	if tenant != nil {
		ctx = context.WithValue(ctx, middleware.TenantContextKey, tenant)
	}

	gfReq.SetCtx(ctx)
	return gfReq, nil
}

// Create a mock user handler with injected service
func createTestUserHandler(mockService *MockUserService) *handlers.UserHandler {
	// This is a simplified approach - in real implementation, 
	// you'd need proper dependency injection
	return &handlers.UserHandler{
		UserService: mockService, // This would require exposing the field or constructor
	}
}

func TestUserHandler_CreateUser(t *testing.T) {
	mockService := &MockUserService{}
	handler := createTestUserHandler(mockService)

	t.Run("Successful user creation", func(t *testing.T) {
		// Setup authenticated user (tenant admin)
		tenantID := "tenant-1"
		authUser := &domainUser.User{
			ID:       "admin-1",
			TenantID: &tenantID,
			Status:   domainUser.StatusActive,
		}
		authTenant := &domainTenant.Tenant{
			ID:     tenantID,
			Status: domainTenant.StatusActive,
		}

		// Request payload
		reqBody := map[string]interface{}{
			"username":  "testuser",
			"email":     "test@example.com",
			"password":  "password123",
			"firstName": "Test",
			"lastName":  "User",
		}

		// Expected response
		expectedResponse := &user.UserResponse{
			ID:        "user-1",
			TenantID:  &tenantID,
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			Status:    "active",
		}

		// Setup mock expectations
		mockService.On("CreateUser", mock.Anything, mock.MatchedBy(func(req user.CreateUserRequest) bool {
			return req.Username == "testuser" && req.TenantID != nil && *req.TenantID == tenantID
		})).Return(expectedResponse, nil)

		// Create request
		req, err := createAuthenticatedRequest("POST", "/v1/users", reqBody, authUser, authTenant)
		require.NoError(t, err)

		// Create response recorder
		recorder := httptest.NewRecorder()
		req.Response = &ghttp.Response{ResponseWriter: recorder}

		// Execute handler
		handler.CreateUser(req)

		// Assertions
		assert.Equal(t, http.StatusCreated, recorder.Code)
		
		var response map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "User created successfully", response["message"])
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})

	t.Run("Validation error", func(t *testing.T) {
		// Setup authenticated user
		tenantID := "tenant-1"
		authUser := &domainUser.User{
			ID:       "admin-1",
			TenantID: &tenantID,
			Status:   domainUser.StatusActive,
		}

		// Invalid request payload (missing required fields)
		reqBody := map[string]interface{}{
			"username": "", // Invalid empty username
		}

		// Create request
		req, err := createAuthenticatedRequest("POST", "/v1/users", reqBody, authUser, nil)
		require.NoError(t, err)

		// Create response recorder
		recorder := httptest.NewRecorder()
		req.Response = &ghttp.Response{ResponseWriter: recorder}

		// Execute handler
		handler.CreateUser(req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, recorder.Code)

		var response map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
		assert.Contains(t, response["error"].(string), "Validation failed")
	})
}

func TestUserHandler_GetUser(t *testing.T) {
	mockService := &MockUserService{}
	handler := createTestUserHandler(mockService)

	t.Run("Get user successfully", func(t *testing.T) {
		// Setup authenticated user (same tenant)
		tenantID := "tenant-1"
		authUser := &domainUser.User{
			ID:       "admin-1",
			TenantID: &tenantID,
			Status:   domainUser.StatusActive,
		}

		userID := "user-1"
		expectedResponse := &user.UserResponse{
			ID:        userID,
			TenantID:  &tenantID,
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			Status:    "active",
		}

		// Setup mock expectations
		mockService.On("GetUserByID", mock.Anything, userID, &tenantID).Return(expectedResponse, nil)

		// Create request
		req, err := createAuthenticatedRequest("GET", "/v1/users/"+userID, nil, authUser, nil)
		require.NoError(t, err)
		req.SetParam("id", userID) // Simulate path parameter

		// Create response recorder
		recorder := httptest.NewRecorder()
		req.Response = &ghttp.Response{ResponseWriter: recorder}

		// Execute handler
		handler.GetUser(req)

		// Assertions
		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})

	t.Run("User not found", func(t *testing.T) {
		// Setup authenticated user
		tenantID := "tenant-1"
		authUser := &domainUser.User{
			ID:       "admin-1",
			TenantID: &tenantID,
			Status:   domainUser.StatusActive,
		}

		userID := "nonexistent-user"

		// Setup mock expectations
		mockService.On("GetUserByID", mock.Anything, userID, &tenantID).Return(nil, &domainUser.ValidationError{
			Field:   "user",
			Message: "用户不存在",
			Code:    "NOT_FOUND",
		})

		// Create request
		req, err := createAuthenticatedRequest("GET", "/v1/users/"+userID, nil, authUser, nil)
		require.NoError(t, err)
		req.SetParam("id", userID)

		// Create response recorder
		recorder := httptest.NewRecorder()
		req.Response = &ghttp.Response{ResponseWriter: recorder}

		// Execute handler
		handler.GetUser(req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, recorder.Code) // Domain validation errors return 400

		var response map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
		assert.Equal(t, "用户不存在", response["error"])

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_UpdateUser(t *testing.T) {
	mockService := &MockUserService{}
	handler := createTestUserHandler(mockService)

	t.Run("Update user successfully", func(t *testing.T) {
		// Setup authenticated user
		tenantID := "tenant-1"
		authUser := &domainUser.User{
			ID:       "admin-1",
			TenantID: &tenantID,
			Status:   domainUser.StatusActive,
		}

		userID := "user-1"
		reqBody := map[string]interface{}{
			"firstName": "UpdatedFirst",
			"lastName":  "UpdatedLast",
		}

		expectedResponse := &user.UserResponse{
			ID:        userID,
			TenantID:  &tenantID,
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "UpdatedFirst",
			LastName:  "UpdatedLast",
			Status:    "active",
		}

		// Setup mock expectations
		mockService.On("UpdateUser", mock.Anything, userID, mock.AnythingOfType("user.UpdateUserRequest"), &tenantID).Return(expectedResponse, nil)

		// Create request
		req, err := createAuthenticatedRequest("PUT", "/v1/users/"+userID, reqBody, authUser, nil)
		require.NoError(t, err)
		req.SetParam("id", userID)

		// Create response recorder
		recorder := httptest.NewRecorder()
		req.Response = &ghttp.Response{ResponseWriter: recorder}

		// Execute handler
		handler.UpdateUser(req)

		// Assertions
		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "User updated successfully", response["message"])

		mockService.AssertExpectations(t)
	})
}

func TestUserHandler_DeleteUser(t *testing.T) {
	mockService := &MockUserService{}
	handler := createTestUserHandler(mockService)

	t.Run("Delete user successfully", func(t *testing.T) {
		// Setup authenticated user
		tenantID := "tenant-1"
		authUser := &domainUser.User{
			ID:       "admin-1",
			TenantID: &tenantID,
			Status:   domainUser.StatusActive,
		}

		userID := "user-1"

		// Setup mock expectations
		mockService.On("DeleteUser", mock.Anything, userID, &tenantID).Return(nil)

		// Create request
		req, err := createAuthenticatedRequest("DELETE", "/v1/users/"+userID, nil, authUser, nil)
		require.NoError(t, err)
		req.SetParam("id", userID)

		// Create response recorder
		recorder := httptest.NewRecorder()
		req.Response = &ghttp.Response{ResponseWriter: recorder}

		// Execute handler
		handler.DeleteUser(req)

		// Assertions
		assert.Equal(t, http.StatusOK, recorder.Code)

		var response map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.True(t, response["success"].(bool))
		assert.Equal(t, "User deleted successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("Cannot delete system admin", func(t *testing.T) {
		// Setup authenticated system admin
		authUser := &domainUser.User{
			ID:       "admin-1",
			TenantID: nil, // System admin
			Status:   domainUser.StatusActive,
		}

		userID := "system-admin-1"

		// Setup mock expectations
		mockService.On("DeleteUser", mock.Anything, userID, (*string)(nil)).Return(&domainUser.ValidationError{
			Field:   "user",
			Message: "系统管理员用户不能被删除",
			Code:    "SYSTEM_ADMIN_UNDELETABLE",
		})

		// Create request
		req, err := createAuthenticatedRequest("DELETE", "/v1/users/"+userID, nil, authUser, nil)
		require.NoError(t, err)
		req.SetParam("id", userID)

		// Create response recorder
		recorder := httptest.NewRecorder()
		req.Response = &ghttp.Response{ResponseWriter: recorder}

		// Execute handler
		handler.DeleteUser(req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, recorder.Code) // Domain validation error

		var response map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.False(t, response["success"].(bool))
		assert.Equal(t, "系统管理员用户不能被删除", response["error"])

		mockService.AssertExpectations(t)
	})
}