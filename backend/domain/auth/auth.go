package auth

// AuthService defines the authentication domain service interface
type AuthService interface {
	Authenticate(username, password string) (string, error)
	ValidateToken(token string) (*Claims, error)
	RefreshToken(token string) (string, error)
}

// Claims represents JWT token claims
type Claims struct {
	UserID   uint64 `json:"user_id"`
	TenantID uint64 `json:"tenant_id"`
	Username string `json:"username"`
}
