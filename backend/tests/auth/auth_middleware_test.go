package auth

import (
	"context"
	"testing"

	"github.com/gofromzero/project_temp/backend/api/middleware"
	"github.com/gofromzero/project_temp/backend/application/auth"
	"github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/gofromzero/project_temp/backend/pkg/utils"
	"github.com/stretchr/testify/suite"
)

// AuthMiddlewareTestSuite is the test suite for authentication middleware
type AuthMiddlewareTestSuite struct {
	suite.Suite
	userRepo     *MockUserRepository
	tenantRepo   *MockTenantRepository
	authService  *auth.AuthService
	middleware   *middleware.AuthMiddleware
	jwtManager   *utils.JWTManager
}

// SetupSuite sets up the test suite
func (suite *AuthMiddlewareTestSuite) SetupSuite() {
	suite.userRepo = new(MockUserRepository)
	suite.tenantRepo = new(MockTenantRepository)
	
	var err error
	suite.authService, err = auth.NewAuthService(suite.userRepo, suite.tenantRepo)
	suite.Require().NoError(err)

	// Create test JWT manager
	suite.jwtManager = createTestJWTManager()

	// Create middleware with public paths
	publicPaths := []string{"/auth/login", "/auth/register", "/health"}
	suite.middleware, err = middleware.NewAuthMiddleware(suite.authService, publicPaths)
	suite.Require().NoError(err)
}

// SetupTest sets up each test
func (suite *AuthMiddlewareTestSuite) SetupTest() {
	// Reset mocks for each test
	suite.userRepo.ExpectedCalls = nil
	suite.userRepo.Calls = nil
	suite.tenantRepo.ExpectedCalls = nil
	suite.tenantRepo.Calls = nil
}

// createTestToken creates a test JWT token
func (suite *AuthMiddlewareTestSuite) createTestToken(userID string, tenantID *string, isAdmin bool) string {
	token, err := suite.jwtManager.GenerateAccessToken(
		userID, "testuser", "test@example.com", tenantID, []string{"user"}, isAdmin,
	)
	suite.Require().NoError(err)
	return token
}

// TestMiddlewareCreation tests that middleware can be created successfully
func (suite *AuthMiddlewareTestSuite) TestMiddlewareCreation() {
	// Test that middleware is created with auth service and JWT manager
	suite.NotNil(suite.middleware, "Middleware should be created successfully")
	
	// Test that new middleware can be created with different public paths
	publicPaths := []string{"/public", "/health"}
	newMiddleware, err := middleware.NewAuthMiddleware(suite.authService, publicPaths)
	suite.NoError(err, "Should create new middleware without error")
	suite.NotNil(newMiddleware, "New middleware should not be nil")
}

// TestJWTTokenValidation tests JWT token validation logic
func (suite *AuthMiddlewareTestSuite) TestJWTTokenValidation() {
	// Test valid token creation and validation
	token := suite.createTestToken("user-123", nil, false)
	suite.NotEmpty(token, "Token should be created successfully")

	// Test token payload extraction
	payload, err := suite.jwtManager.ValidateTokenWithoutBlacklist(token)
	suite.NoError(err, "Valid token should be validated successfully")
	suite.Equal("user-123", payload.UserID, "Token payload should contain correct user ID")
}

// TestAuthServiceIntegration tests authentication service integration
func (suite *AuthMiddlewareTestSuite) TestAuthServiceIntegration() {
	// Test that auth service is accessible and functional
	suite.NotNil(suite.authService, "Auth service should be available")
	
	// Test that JWT manager is properly configured
	suite.NotNil(suite.jwtManager, "JWT manager should be available")
	suite.NotEmpty(suite.jwtManager.Config.SecretKey, "JWT manager should have secret key configured")
}

// TestInvalidTokenHandling tests invalid token validation
func (suite *AuthMiddlewareTestSuite) TestInvalidTokenHandling() {
	// Test invalid token validation
	_, err := suite.jwtManager.ValidateTokenWithoutBlacklist("invalid-token")
	suite.Error(err, "Invalid token should fail validation")
	suite.Equal(utils.ErrInvalidToken, err, "Should return specific invalid token error")

	// Test empty token validation
	_, err = suite.jwtManager.ValidateTokenWithoutBlacklist("")
	suite.Error(err, "Empty token should fail validation")
}

// TestBearerTokenExtraction tests Bearer token extraction utility
func (suite *AuthMiddlewareTestSuite) TestBearerTokenExtraction() {
	// Test valid Bearer token extraction
	token, err := utils.ExtractBearerToken("Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
	suite.NoError(err, "Valid Bearer token should be extracted")
	suite.Equal("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", token, "Extracted token should match")

	// Test invalid header format
	_, err = utils.ExtractBearerToken("Basic dXNlcjpwYXNz")
	suite.Error(err, "Invalid header format should fail")

	// Test empty header
	_, err = utils.ExtractBearerToken("")
	suite.Error(err, "Empty header should fail")
}

// TestContextUtilityFunctions tests context helper functions
func (suite *AuthMiddlewareTestSuite) TestContextUtilityFunctions() {
	// Create test context with authentication values
	ctx := context.Background()
	testUser := &user.User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Status:   user.StatusActive,
	}
	testPayload := &utils.TokenPayload{
		UserID:   "user-123",
		Username: "testuser",
		Email:    "test@example.com",
	}

	// Test context value setting and getting
	ctx = context.WithValue(ctx, middleware.UserContextKey, testUser)
	ctx = context.WithValue(ctx, middleware.TokenContextKey, testPayload)

	// Test context getters
	user, userOk := middleware.GetAuthenticatedUser(ctx)
	suite.True(userOk, "Should retrieve user from context")
	suite.Equal(testUser.ID, user.ID, "Retrieved user should match")

	payload, payloadOk := middleware.GetTokenPayload(ctx)
	suite.True(payloadOk, "Should retrieve token payload from context")
	suite.Equal(testPayload.UserID, payload.UserID, "Retrieved payload should match")
}

// TestAuthMiddlewareTestSuite runs the test suite
func TestAuthMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}