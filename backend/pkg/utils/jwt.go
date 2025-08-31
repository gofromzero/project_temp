package utils

import (
	"context"
	"errors"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken     = errors.New("invalid token")
	ErrExpiredToken     = errors.New("token has expired")
	ErrBlacklistedToken = errors.New("token has been blacklisted")
)

// TokenPayload represents the JWT token payload structure
type TokenPayload struct {
	UserID    string   `json:"user_id"`
	TenantID  *string  `json:"tenant_id,omitempty"` // null for system admins
	Username  string   `json:"username"`
	Email     string   `json:"email"`
	Roles     []string `json:"roles"`
	IsAdmin   bool     `json:"is_admin"`   // system admin flag
	TokenType string   `json:"token_type"` // access or refresh
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	BlacklistKeyPrefix string
}

// JWTManager handles JWT operations
type JWTManager struct {
	Config *JWTConfig // Export for testing
}

// NewJWTManager creates a new JWT manager instance
func NewJWTManager() (*JWTManager, error) {
	secretKey := g.Cfg().MustGet(context.Background(), "jwt.secret_key").String()
	if secretKey == "" {
		return nil, errors.New("JWT secret key not configured")
	}

	config := &JWTConfig{
		SecretKey:          secretKey,
		AccessTokenExpiry:  1 * time.Hour,      // 1 hour for access tokens
		RefreshTokenExpiry: 7 * 24 * time.Hour, // 7 days for refresh tokens
		BlacklistKeyPrefix: "jwt_blacklist:",
	}

	return &JWTManager{Config: config}, nil
}

// GenerateAccessToken creates a new access token
func (j *JWTManager) GenerateAccessToken(userID, username, email string, tenantID *string, roles []string, isAdmin bool) (string, error) {
	now := time.Now()
	payload := &TokenPayload{
		UserID:    userID,
		TenantID:  tenantID,
		Username:  username,
		Email:     email,
		Roles:     roles,
		IsAdmin:   isAdmin,
		TokenType: "access",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "multi-tenant-admin",
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.Config.AccessTokenExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(j.Config.SecretKey))
}

// GenerateRefreshToken creates a new refresh token
func (j *JWTManager) GenerateRefreshToken(userID string, tenantID *string) (string, error) {
	now := time.Now()
	payload := &TokenPayload{
		UserID:    userID,
		TenantID:  tenantID,
		TokenType: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "multi-tenant-admin",
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.Config.RefreshTokenExpiry)),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	return token.SignedString([]byte(j.Config.SecretKey))
}

// ValidateToken validates and parses a JWT token
func (j *JWTManager) ValidateToken(tokenString string) (*TokenPayload, error) {
	// Check if token is blacklisted first
	ctx := context.Background()
	isBlacklisted, err := j.isTokenBlacklisted(ctx, tokenString)
	if err != nil {
		return nil, err
	}
	if isBlacklisted {
		return nil, ErrBlacklistedToken
	}

	token, err := jwt.ParseWithClaims(tokenString, &TokenPayload{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.Config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	payload, ok := token.Claims.(*TokenPayload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}

// ValidateTokenWithoutBlacklist validates JWT token without checking blacklist (for unit tests)
func (j *JWTManager) ValidateTokenWithoutBlacklist(tokenString string) (*TokenPayload, error) {
	token, err := jwt.ParseWithClaims(tokenString, &TokenPayload{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(j.Config.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	payload, ok := token.Claims.(*TokenPayload)
	if !ok {
		return nil, ErrInvalidToken
	}

	return payload, nil
}

// BlacklistToken adds a token to the blacklist
func (j *JWTManager) BlacklistToken(ctx context.Context, tokenString string) error {
	// Parse token to get expiration time
	payload, err := j.ValidateToken(tokenString)
	if err != nil && !errors.Is(err, ErrBlacklistedToken) {
		// Even if token is invalid/expired, we still blacklist it
		payload = &TokenPayload{}
	}

	key := j.Config.BlacklistKeyPrefix + tokenString
	expiry := j.Config.AccessTokenExpiry
	if payload.TokenType == "refresh" {
		expiry = j.Config.RefreshTokenExpiry
	}

	// Set token in Redis with expiry
	err = g.Redis().SetEX(ctx, key, "1", int64(expiry.Seconds()))
	return err
}

// isTokenBlacklisted checks if token is in blacklist
func (j *JWTManager) isTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	key := j.Config.BlacklistKeyPrefix + tokenString
	result, err := g.Redis().Get(ctx, key)
	if err != nil {
		// If Redis is down, allow the request (fail open)
		g.Log().Warning(ctx, "Redis blacklist check failed:", err)
		return false, nil
	}
	return result.String() == "1", nil
}

// ExtractBearerToken extracts JWT token from Authorization header
func ExtractBearerToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		return "", errors.New("invalid authorization header format")
	}

	return authHeader[len(bearerPrefix):], nil
}

// GetExpiryDuration returns the appropriate expiry duration for token type
func (j *JWTManager) GetExpiryDuration(tokenType string) time.Duration {
	if tokenType == "refresh" {
		return j.Config.RefreshTokenExpiry
	}
	return j.Config.AccessTokenExpiry
}
