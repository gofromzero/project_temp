package auth

import (
	"os"
	"testing"
	"time"

	"github.com/gofromzero/project_temp/backend/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Set environment variable for JWT secret
	os.Setenv("JWT_SECRET", "test-secret-key-for-jwt-testing")

	// Run tests
	code := m.Run()

	// Cleanup
	os.Unsetenv("JWT_SECRET")
	os.Exit(code)
}

// Create a manual JWT manager for testing without GoFrame config dependency
func createTestJWTManager() *utils.JWTManager {
	config := &utils.JWTConfig{
		SecretKey:          "test-secret-key-for-jwt-testing",
		AccessTokenExpiry:  1 * time.Hour,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		BlacklistKeyPrefix: "jwt_blacklist:",
	}
	return &utils.JWTManager{Config: config}
}

func TestJWTManager(t *testing.T) {
	// Create test JWT manager directly
	jwtManager := createTestJWTManager()
	require.NotNil(t, jwtManager)

	t.Run("Generate and validate access token", func(t *testing.T) {
		userID := "user-123"
		username := "testuser"
		email := "test@example.com"
		tenantID := "tenant-456"
		roles := []string{"admin", "user"}
		isAdmin := false

		// Generate access token
		token, err := jwtManager.GenerateAccessToken(userID, username, email, &tenantID, roles, isAdmin)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate token (skip Redis blacklist check for unit test)
		payload, err := jwtManager.ValidateTokenWithoutBlacklist(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, payload.UserID)
		assert.Equal(t, username, payload.Username)
		assert.Equal(t, email, payload.Email)
		assert.Equal(t, tenantID, *payload.TenantID)
		assert.Equal(t, roles, payload.Roles)
		assert.Equal(t, isAdmin, payload.IsAdmin)
		assert.Equal(t, "access", payload.TokenType)
	})

	t.Run("Generate and validate refresh token", func(t *testing.T) {
		userID := "user-123"
		tenantID := "tenant-456"

		// Generate refresh token
		token, err := jwtManager.GenerateRefreshToken(userID, &tenantID)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate token
		payload, err := jwtManager.ValidateTokenWithoutBlacklist(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, payload.UserID)
		assert.Equal(t, tenantID, *payload.TenantID)
		assert.Equal(t, "refresh", payload.TokenType)
	})

	t.Run("System admin token (no tenant)", func(t *testing.T) {
		userID := "admin-123"
		username := "sysadmin"
		email := "admin@system.com"
		roles := []string{"system_admin"}
		isAdmin := true

		// Generate token for system admin
		token, err := jwtManager.GenerateAccessToken(userID, username, email, nil, roles, isAdmin)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		// Validate token
		payload, err := jwtManager.ValidateTokenWithoutBlacklist(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, payload.UserID)
		assert.Nil(t, payload.TenantID)
		assert.True(t, payload.IsAdmin)
		assert.Equal(t, roles, payload.Roles)
	})

	t.Run("Invalid token validation", func(t *testing.T) {
		// Test with invalid token
		_, err := jwtManager.ValidateTokenWithoutBlacklist("invalid.token.here")
		assert.Error(t, err)
		assert.Equal(t, utils.ErrInvalidToken, err)

		// Test with empty token
		_, err = jwtManager.ValidateTokenWithoutBlacklist("")
		assert.Error(t, err)
		assert.Equal(t, utils.ErrInvalidToken, err)
	})
}

func TestExtractBearerToken(t *testing.T) {
	t.Run("Valid Bearer token", func(t *testing.T) {
		authHeader := "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"
		token, err := utils.ExtractBearerToken(authHeader)
		assert.NoError(t, err)
		assert.Equal(t, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", token)
	})

	t.Run("Empty authorization header", func(t *testing.T) {
		_, err := utils.ExtractBearerToken("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "authorization header is empty")
	})

	t.Run("Invalid authorization header format", func(t *testing.T) {
		_, err := utils.ExtractBearerToken("Basic dXNlcjpwYXNz")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid authorization header format")
	})

	t.Run("Bearer without token", func(t *testing.T) {
		_, err := utils.ExtractBearerToken("Bearer")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid authorization header format")
	})
}

func TestTokenPayloadStructure(t *testing.T) {
	jwtManager := createTestJWTManager()

	t.Run("Complete payload validation", func(t *testing.T) {
		userID := "user-789"
		username := "fulltest"
		email := "fulltest@example.com"
		tenantID := "tenant-999"
		roles := []string{"manager", "viewer"}
		isAdmin := false

		token, err := jwtManager.GenerateAccessToken(userID, username, email, &tenantID, roles, isAdmin)
		assert.NoError(t, err)

		payload, err := jwtManager.ValidateTokenWithoutBlacklist(token)
		assert.NoError(t, err)

		// Verify all payload fields
		assert.Equal(t, userID, payload.UserID)
		assert.Equal(t, username, payload.Username)
		assert.Equal(t, email, payload.Email)
		assert.NotNil(t, payload.TenantID)
		assert.Equal(t, tenantID, *payload.TenantID)
		assert.Equal(t, roles, payload.Roles)
		assert.Equal(t, isAdmin, payload.IsAdmin)
		assert.Equal(t, "access", payload.TokenType)

		// Verify JWT standard claims
		assert.Equal(t, "multi-tenant-admin", payload.Issuer)
		assert.Equal(t, userID, payload.Subject)
		assert.NotNil(t, payload.IssuedAt)
		assert.NotNil(t, payload.ExpiresAt)
		assert.NotNil(t, payload.NotBefore)
	})

	t.Run("Token expiry durations", func(t *testing.T) {
		accessExpiry := jwtManager.GetExpiryDuration("access")
		refreshExpiry := jwtManager.GetExpiryDuration("refresh")
		unknownExpiry := jwtManager.GetExpiryDuration("unknown")

		assert.Equal(t, time.Hour, accessExpiry)
		assert.Equal(t, 7*24*time.Hour, refreshExpiry)
		assert.Equal(t, time.Hour, unknownExpiry) // defaults to access token expiry
	})
}
