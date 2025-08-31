package auth

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofromzero/project_temp/backend/application/auth"
	"github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockUserRepository is a mock implementation of user.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *user.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id string) (*user.User, error) {
	args := m.Called(id)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(tenantID *string, username string) (*user.User, error) {
	args := m.Called(tenantID, username)
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(tenantID *string, email string) (*user.User, error) {
	args := m.Called(tenantID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByTenantID(tenantID string, offset, limit int) ([]*user.User, error) {
	args := m.Called(tenantID, offset, limit)
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *MockUserRepository) GetSystemAdmins(offset, limit int) ([]*user.User, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *user.User) error {
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

// MockTenantRepository is a mock implementation of tenant.TenantRepository
type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) Create(tenant *tenant.Tenant) error {
	args := m.Called(tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) GetByID(id string) (*tenant.Tenant, error) {
	args := m.Called(id)
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

// AuthServiceIntegrationTestSuite is the test suite for auth service integration tests
type AuthServiceIntegrationTestSuite struct {
	suite.Suite
	userRepo    *MockUserRepository
	tenantRepo  *MockTenantRepository
	authService *auth.AuthService
}

// SetupSuite sets up the test suite
func (suite *AuthServiceIntegrationTestSuite) SetupSuite() {
	suite.userRepo = new(MockUserRepository)
	suite.tenantRepo = new(MockTenantRepository)
	
	var err error
	suite.authService, err = auth.NewAuthService(suite.userRepo, suite.tenantRepo)
	suite.Require().NoError(err)
}

// SetupTest sets up each test
func (suite *AuthServiceIntegrationTestSuite) SetupTest() {
	// Reset mocks for each test
	suite.userRepo.ExpectedCalls = nil
	suite.userRepo.Calls = nil
	suite.tenantRepo.ExpectedCalls = nil
	suite.tenantRepo.Calls = nil
}

// createTestUser creates a test user with hashed password
func (suite *AuthServiceIntegrationTestSuite) createTestUser(email, password string, tenantID *string) *user.User {
	testUser := &user.User{
		ID:        "test-user-id",
		TenantID:  tenantID,
		Username:  "testuser",
		Email:     email,
		FirstName: "Test",
		LastName:  "User",
		Status:    user.StatusActive,
	}
	
	// Set password using bcrypt
	err := testUser.SetPassword(password)
	suite.Require().NoError(err)
	
	return testUser
}

// createTestTenant creates a test tenant
func (suite *AuthServiceIntegrationTestSuite) createTestTenant() *tenant.Tenant {
	return &tenant.Tenant{
		ID:     "test-tenant-id",
		Name:   "Test Tenant",
		Code:   "TEST",
		Status: tenant.StatusActive,
	}
}

// TestSuccessfulTenantLogin tests successful tenant user login
func (suite *AuthServiceIntegrationTestSuite) TestSuccessfulTenantLogin() {
	// Arrange
	ctx := context.Background()
	testTenant := suite.createTestTenant()
	testUser := suite.createTestUser("test@example.com", "password123", &testTenant.ID)
	
	suite.tenantRepo.On("GetByCode", "TEST").Return(testTenant, nil)
	suite.userRepo.On("GetByEmail", &testTenant.ID, "test@example.com").Return(testUser, nil)
	suite.userRepo.On("Update", mock.AnythingOfType("*user.User")).Return(nil)

	loginReq := &auth.LoginRequest{
		Email:      "test@example.com",
		Password:   "password123",
		TenantCode: &testTenant.Code,
	}

	// Act
	response, err := suite.authService.Login(ctx, loginReq)

	// Assert
	suite.NoError(err)
	suite.NotNil(response)
	suite.NotEmpty(response.Token)
	suite.NotEmpty(response.RefreshToken)
	suite.NotNil(response.User)
	suite.Equal("test-user-id", response.User.ID)
	suite.Equal(&testTenant.ID, response.User.TenantID)
	suite.Equal("test@example.com", response.User.Email)
	suite.Equal(user.StatusActive, response.User.Status)

	// Verify mock expectations
	suite.tenantRepo.AssertExpectations(suite.T())
	suite.userRepo.AssertExpectations(suite.T())
}

// TestSuccessfulSystemAdminLogin tests successful system admin login
func (suite *AuthServiceIntegrationTestSuite) TestSuccessfulSystemAdminLogin() {
	// Arrange
	ctx := context.Background()
	testUser := suite.createTestUser("admin@system.com", "adminpass123", nil) // System admin has nil tenant
	
	suite.userRepo.On("GetByEmail", (*string)(nil), "admin@system.com").Return(testUser, nil)
	suite.userRepo.On("Update", mock.AnythingOfType("*user.User")).Return(nil)

	loginReq := &auth.LoginRequest{
		Email:    "admin@system.com",
		Password: "adminpass123",
		// No tenant code for system admin
	}

	// Act
	response, err := suite.authService.Login(ctx, loginReq)

	// Assert
	suite.NoError(err)
	suite.NotNil(response)
	suite.NotEmpty(response.Token)
	suite.NotEmpty(response.RefreshToken)
	suite.NotNil(response.User)
	suite.Equal("test-user-id", response.User.ID)
	suite.Nil(response.User.TenantID) // System admin has no tenant
	suite.Equal("admin@system.com", response.User.Email)
	suite.Equal(user.StatusActive, response.User.Status)

	// Verify mock expectations
	suite.userRepo.AssertExpectations(suite.T())
}

// TestInvalidCredentials tests login with invalid password
func (suite *AuthServiceIntegrationTestSuite) TestInvalidCredentials() {
	// Arrange
	ctx := context.Background()
	testTenant := suite.createTestTenant()
	testUser := suite.createTestUser("test@example.com", "correctpassword", &testTenant.ID)
	
	suite.tenantRepo.On("GetByCode", "TEST").Return(testTenant, nil)
	suite.userRepo.On("GetByEmail", &testTenant.ID, "test@example.com").Return(testUser, nil)

	loginReq := &auth.LoginRequest{
		Email:      "test@example.com",
		Password:   "wrongpassword", // Wrong password
		TenantCode: &testTenant.Code,
	}

	// Act
	response, err := suite.authService.Login(ctx, loginReq)

	// Assert
	suite.Error(err)
	suite.Equal(auth.ErrInvalidCredentials, err)
	suite.Nil(response)

	// Verify mock expectations
	suite.tenantRepo.AssertExpectations(suite.T())
	suite.userRepo.AssertExpectations(suite.T())
}

// TestTenantNotFound tests login with invalid tenant code
func (suite *AuthServiceIntegrationTestSuite) TestTenantNotFound() {
	// Arrange
	ctx := context.Background()
	suite.tenantRepo.On("GetByCode", "INVALID").Return(nil, fmt.Errorf("tenant not found"))

	loginReq := &auth.LoginRequest{
		Email:      "test@example.com",
		Password:   "password123",
		TenantCode: stringPtr("INVALID"),
	}

	// Act
	response, err := suite.authService.Login(ctx, loginReq)

	// Assert
	suite.Error(err)
	suite.Equal(auth.ErrTenantNotFound, err)
	suite.Nil(response)

	// Verify mock expectations
	suite.tenantRepo.AssertExpectations(suite.T())
}

// TestUserNotFound tests login with invalid user email
func (suite *AuthServiceIntegrationTestSuite) TestUserNotFound() {
	// Arrange
	ctx := context.Background()
	testTenant := suite.createTestTenant()
	
	suite.tenantRepo.On("GetByCode", "TEST").Return(testTenant, nil)
	suite.userRepo.On("GetByEmail", &testTenant.ID, "nonexistent@example.com").Return(nil, fmt.Errorf("user not found"))

	loginReq := &auth.LoginRequest{
		Email:      "nonexistent@example.com",
		Password:   "password123",
		TenantCode: &testTenant.Code,
	}

	// Act
	response, err := suite.authService.Login(ctx, loginReq)

	// Assert
	suite.Error(err)
	suite.Equal(auth.ErrInvalidCredentials, err)
	suite.Nil(response)

	// Verify mock expectations
	suite.tenantRepo.AssertExpectations(suite.T())
	suite.userRepo.AssertExpectations(suite.T())
}

// TestInactiveUser tests login with inactive user
func (suite *AuthServiceIntegrationTestSuite) TestInactiveUser() {
	// Arrange
	ctx := context.Background()
	testTenant := suite.createTestTenant()
	testUser := suite.createTestUser("test@example.com", "password123", &testTenant.ID)
	testUser.Status = user.StatusInactive // Set user as inactive
	
	suite.tenantRepo.On("GetByCode", "TEST").Return(testTenant, nil)
	suite.userRepo.On("GetByEmail", &testTenant.ID, "test@example.com").Return(testUser, nil)

	loginReq := &auth.LoginRequest{
		Email:      "test@example.com",
		Password:   "password123",
		TenantCode: &testTenant.Code,
	}

	// Act
	response, err := suite.authService.Login(ctx, loginReq)

	// Assert
	suite.Error(err)
	suite.Equal(auth.ErrUserInactive, err)
	suite.Nil(response)

	// Verify mock expectations
	suite.tenantRepo.AssertExpectations(suite.T())
	suite.userRepo.AssertExpectations(suite.T())
}

// TestSuspendedTenant tests login with suspended tenant
func (suite *AuthServiceIntegrationTestSuite) TestSuspendedTenant() {
	// Arrange
	ctx := context.Background()
	testTenant := suite.createTestTenant()
	testTenant.Status = tenant.StatusSuspended // Set tenant as suspended
	testUser := suite.createTestUser("test@example.com", "password123", &testTenant.ID)
	
	suite.tenantRepo.On("GetByCode", "TEST").Return(testTenant, nil)
	suite.userRepo.On("GetByEmail", &testTenant.ID, "test@example.com").Return(testUser, nil)

	loginReq := &auth.LoginRequest{
		Email:      "test@example.com",
		Password:   "password123",
		TenantCode: &testTenant.Code,
	}

	// Act
	response, err := suite.authService.Login(ctx, loginReq)

	// Assert
	suite.Error(err)
	suite.Equal(auth.ErrTenantSuspended, err)
	suite.Nil(response)

	// Verify mock expectations
	suite.tenantRepo.AssertExpectations(suite.T())
	suite.userRepo.AssertExpectations(suite.T())
}

// TestBcryptPasswordVerification tests that bcrypt password verification works correctly
func (suite *AuthServiceIntegrationTestSuite) TestBcryptPasswordVerification() {
	// Arrange
	ctx := context.Background()
	testTenant := suite.createTestTenant()
	
	// Create a user with a specific password
	plainPassword := "MySecurePassword123!"
	testUser := suite.createTestUser("test@example.com", plainPassword, &testTenant.ID)
	
	suite.tenantRepo.On("GetByCode", "TEST").Return(testTenant, nil)
	suite.userRepo.On("GetByEmail", &testTenant.ID, "test@example.com").Return(testUser, nil)
	suite.userRepo.On("Update", mock.AnythingOfType("*user.User")).Return(nil)

	loginReq := &auth.LoginRequest{
		Email:      "test@example.com",
		Password:   plainPassword, // Use the same password
		TenantCode: &testTenant.Code,
	}

	// Act
	response, err := suite.authService.Login(ctx, loginReq)

	// Assert
	suite.NoError(err)
	suite.NotNil(response)
	
	// Verify the password was properly hashed and verified
	suite.True(testUser.CheckPassword(plainPassword)) // Direct check
	suite.False(testUser.CheckPassword("wrongpassword")) // Should fail with wrong password

	// Verify mock expectations
	suite.tenantRepo.AssertExpectations(suite.T())
	suite.userRepo.AssertExpectations(suite.T())
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// TestSuccessfulRegistrationBySystemAdmin tests system admin registering a new user
func (suite *AuthServiceIntegrationTestSuite) TestSuccessfulRegistrationBySystemAdmin() {
	// Arrange
	ctx := context.Background()
	testTenant := suite.createTestTenant()
	adminUser := suite.createTestUser("admin@system.com", "adminpass", nil) // System admin
	
	suite.tenantRepo.On("GetByCode", "TEST").Return(testTenant, nil)
	suite.userRepo.On("ExistsByEmail", &testTenant.ID, "newuser@example.com").Return(false, nil)
	suite.userRepo.On("ExistsByUsername", &testTenant.ID, "newuser").Return(false, nil)
	suite.userRepo.On("Create", mock.AnythingOfType("*user.User")).Return(nil)

	registerReq := &auth.RegisterRequest{
		Email:      "newuser@example.com",
		Password:   "password123",
		Username:   "newuser",
		FirstName:  "New",
		LastName:   "User",
		TenantCode: &testTenant.Code,
	}

	// Act
	response, err := suite.authService.Register(ctx, registerReq, adminUser)

	// Assert
	suite.NoError(err)
	suite.NotNil(response)
	suite.NotNil(response.User)
	suite.Equal("newuser@example.com", response.User.Email)
	suite.Equal("newuser", response.User.Username)
	suite.Equal(&testTenant.ID, response.User.TenantID)
	suite.Equal(user.StatusActive, response.User.Status)

	// Verify mock expectations
	suite.tenantRepo.AssertExpectations(suite.T())
	suite.userRepo.AssertExpectations(suite.T())
}

// TestRegistrationUnauthorized tests registration with non-admin user
func (suite *AuthServiceIntegrationTestSuite) TestRegistrationUnauthorized() {
	// Arrange
	ctx := context.Background()
	testTenant := suite.createTestTenant()
	regularUser := suite.createTestUser("user@tenant.com", "userpass", &testTenant.ID) // Regular tenant user

	registerReq := &auth.RegisterRequest{
		Email:      "newuser@example.com",
		Password:   "password123",
		Username:   "newuser",
		FirstName:  "New",
		LastName:   "User",
		TenantCode: &testTenant.Code,
	}

	// Act
	response, err := suite.authService.Register(ctx, registerReq, regularUser)

	// Assert
	suite.Error(err)
	suite.Equal(auth.ErrUnauthorized, err)
	suite.Nil(response)
}

// TestRegistrationEmailAlreadyExists tests registration with existing email
func (suite *AuthServiceIntegrationTestSuite) TestRegistrationEmailAlreadyExists() {
	// Arrange
	ctx := context.Background()
	testTenant := suite.createTestTenant()
	adminUser := suite.createTestUser("admin@system.com", "adminpass", nil)
	
	suite.tenantRepo.On("GetByCode", "TEST").Return(testTenant, nil)
	suite.userRepo.On("ExistsByEmail", &testTenant.ID, "existing@example.com").Return(true, nil)

	registerReq := &auth.RegisterRequest{
		Email:      "existing@example.com", // Email already exists
		Password:   "password123",
		Username:   "newuser",
		FirstName:  "New",
		LastName:   "User",
		TenantCode: &testTenant.Code,
	}

	// Act
	response, err := suite.authService.Register(ctx, registerReq, adminUser)

	// Assert
	suite.Error(err)
	suite.Equal(auth.ErrEmailAlreadyExists, err)
	suite.Nil(response)

	// Verify mock expectations
	suite.tenantRepo.AssertExpectations(suite.T())
	suite.userRepo.AssertExpectations(suite.T())
}

// TestSuccessfulTokenRefresh tests token refresh flow
func (suite *AuthServiceIntegrationTestSuite) TestSuccessfulTokenRefresh() {
	// Arrange
	ctx := context.Background()
	testUser := suite.createTestUser("test@example.com", "password123", nil) // System admin
	
	// Mock JWT manager for testing without config dependency
	jwtManager := createTestJWTManager()
	refreshToken, err := jwtManager.GenerateRefreshToken(testUser.ID, testUser.TenantID)
	suite.Require().NoError(err)

	suite.userRepo.On("GetByID", testUser.ID).Return(testUser, nil)

	// Act
	response, err := suite.authService.RefreshToken(ctx, refreshToken)

	// Assert
	suite.NoError(err)
	suite.NotNil(response)
	suite.NotEmpty(response.Token)
	suite.NotEmpty(response.RefreshToken)
	suite.NotEqual(refreshToken, response.RefreshToken) // Should be a new refresh token
	suite.NotNil(response.User)
	suite.Equal(testUser.ID, response.User.ID)

	// Verify mock expectations
	suite.userRepo.AssertExpectations(suite.T())
}

// TestRefreshWithInactiveUser tests refresh token with inactive user
func (suite *AuthServiceIntegrationTestSuite) TestRefreshWithInactiveUser() {
	// Arrange
	ctx := context.Background()
	testUser := suite.createTestUser("test@example.com", "password123", nil)
	testUser.Status = user.StatusInactive // Make user inactive
	
	// Mock JWT manager for testing without config dependency
	jwtManager := createTestJWTManager()
	refreshToken, err := jwtManager.GenerateRefreshToken(testUser.ID, testUser.TenantID)
	suite.Require().NoError(err)

	suite.userRepo.On("GetByID", testUser.ID).Return(testUser, nil)

	// Act
	response, err := suite.authService.RefreshToken(ctx, refreshToken)

	// Assert
	suite.Error(err)
	suite.Equal(auth.ErrUserInactive, err)
	suite.Nil(response)

	// Verify mock expectations
	suite.userRepo.AssertExpectations(suite.T())
}

// Helper function defined in jwt_test.go to avoid duplication

// TestAuthServiceIntegrationSuite runs the test suite
func TestAuthServiceIntegrationSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceIntegrationTestSuite))
}