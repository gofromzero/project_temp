package auth

import (
	"context"
	"errors"
	"time"

	"github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/gofromzero/project_temp/backend/pkg/utils"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/guid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserLocked        = errors.New("user account is locked")
	ErrUserInactive      = errors.New("user account is inactive")
	ErrTenantNotFound    = errors.New("tenant not found")
	ErrTenantSuspended   = errors.New("tenant is suspended or disabled")
	ErrTooManyAttempts   = errors.New("too many login attempts, account locked")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrUsernameAlreadyExists = errors.New("username already exists")
	ErrUnauthorized      = errors.New("unauthorized access")
)

// LoginRequest represents login request data
type LoginRequest struct {
	Email      string  `json:"email" v:"required|email"`
	Password   string  `json:"password" v:"required|min:6"`
	TenantCode *string `json:"tenantCode,omitempty"`
}

// LoginResponse represents login response data
type LoginResponse struct {
	Token        string            `json:"token"`
	RefreshToken string            `json:"refreshToken"`
	User         *UserInfo         `json:"user"`
}

// RegisterRequest represents user registration request data
type RegisterRequest struct {
	Email      string  `json:"email" v:"required|email"`
	Password   string  `json:"password" v:"required|min:6"`
	Username   string  `json:"username" v:"required|min:3|max:50"`
	FirstName  string  `json:"firstName" v:"required|min:1|max:100"`
	LastName   string  `json:"lastName" v:"required|min:1|max:100"`
	Phone      *string `json:"phone,omitempty" v:"phone"`
	TenantCode *string `json:"tenantCode,omitempty"`
}

// RegisterResponse represents user registration response data
type RegisterResponse struct {
	User *UserInfo `json:"user"`
}

// UserInfo represents user information in response
type UserInfo struct {
	ID       string              `json:"id"`
	TenantID *string             `json:"tenantId,omitempty"`
	Username string              `json:"username"`
	Email    string              `json:"email"`
	Profile  user.UserProfile    `json:"profile"`
	Status   user.UserStatus     `json:"status"`
	Roles    []string            `json:"roles"`
	IsAdmin  bool                `json:"isAdmin"`
}

// AuthService handles authentication operations
type AuthService struct {
	jwtManager     *utils.JWTManager
	userRepo       user.UserRepository
	tenantRepo     tenant.TenantRepository
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo user.UserRepository, tenantRepo tenant.TenantRepository) (*AuthService, error) {
	jwtManager, err := utils.NewJWTManager()
	if err != nil {
		return nil, err
	}

	return &AuthService{
		jwtManager: jwtManager,
		userRepo:   userRepo,
		tenantRepo: tenantRepo,
	}, nil
}

// Login authenticates a user and returns JWT tokens
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*LoginResponse, error) {
	// Get user by email and tenant
	foundUser, foundTenant, err := s.getUserByEmailAndTenant(ctx, req.Email, req.TenantCode)
	if err != nil {
		return nil, err
	}

	// Check login attempt limits
	if err := s.checkLoginAttempts(ctx, foundUser.ID); err != nil {
		return nil, err
	}

	// Verify password
	if !foundUser.CheckPassword(req.Password) {
		// Record failed attempt
		s.recordFailedAttempt(ctx, foundUser.ID)
		return nil, ErrInvalidCredentials
	}

	// Check user status
	if foundUser.Status == user.StatusLocked {
		return nil, ErrUserLocked
	}
	if foundUser.Status == user.StatusInactive {
		return nil, ErrUserInactive
	}

	// Check tenant status if applicable
	if foundTenant != nil && foundTenant.Status != tenant.StatusActive {
		return nil, ErrTenantSuspended
	}

	// Get user roles (simulated for now)
	roles := s.getUserRoles(ctx, foundUser.ID)
	isAdmin := foundUser.IsSystemAdmin() || s.isUserAdmin(ctx, foundUser.ID, foundUser.TenantID)

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(
		foundUser.ID, foundUser.Username, foundUser.Email, foundUser.TenantID, roles, isAdmin,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(foundUser.ID, foundUser.TenantID)
	if err != nil {
		return nil, err
	}

	// Clear failed attempts and update last login
	s.clearFailedAttempts(ctx, foundUser.ID)
	foundUser.UpdateLastLogin()
	s.updateUserLastLogin(ctx, foundUser)

	return &LoginResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: &UserInfo{
			ID:       foundUser.ID,
			TenantID: foundUser.TenantID,
			Username: foundUser.Username,
			Email:    foundUser.Email,
			Profile:  foundUser.GetProfile(),
			Status:   foundUser.Status,
			Roles:    roles,
			IsAdmin:  isAdmin,
		},
	}, nil
}

// RefreshToken generates new access token from refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error) {
	// Validate refresh token
	payload, err := s.jwtManager.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Ensure this is a refresh token
	if payload.TokenType != "refresh" {
		return nil, utils.ErrInvalidToken
	}

	// Get user info
	foundUser, _, err := s.getUserByID(ctx, payload.UserID)
	if err != nil {
		return nil, err
	}

	// Check user is still active
	if foundUser.Status != user.StatusActive {
		return nil, ErrUserInactive
	}

	// Get current roles and admin status
	roles := s.getUserRoles(ctx, foundUser.ID)
	isAdmin := foundUser.IsSystemAdmin() || s.isUserAdmin(ctx, foundUser.ID, foundUser.TenantID)

	// Generate new tokens
	newAccessToken, err := s.jwtManager.GenerateAccessToken(
		foundUser.ID, foundUser.Username, foundUser.Email, foundUser.TenantID, roles, isAdmin,
	)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := s.jwtManager.GenerateRefreshToken(foundUser.ID, foundUser.TenantID)
	if err != nil {
		return nil, err
	}

	// Blacklist the old refresh token
	s.jwtManager.BlacklistToken(ctx, refreshToken)

	return &LoginResponse{
		Token:        newAccessToken,
		RefreshToken: newRefreshToken,
		User: &UserInfo{
			ID:       foundUser.ID,
			TenantID: foundUser.TenantID,
			Username: foundUser.Username,
			Email:    foundUser.Email,
			Profile:  foundUser.GetProfile(),
			Status:   foundUser.Status,
			Roles:    roles,
			IsAdmin:  isAdmin,
		},
	}, nil
}

// Logout invalidates the user's tokens
func (s *AuthService) Logout(ctx context.Context, accessToken, refreshToken string) error {
	// Blacklist both tokens
	if accessToken != "" {
		s.jwtManager.BlacklistToken(ctx, accessToken)
	}
	if refreshToken != "" {
		s.jwtManager.BlacklistToken(ctx, refreshToken)
	}
	return nil
}

// Register creates a new user account (admin-only operation)
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest, adminUser *user.User) (*RegisterResponse, error) {
	// Verify admin permissions
	if adminUser == nil {
		return nil, ErrUnauthorized
	}

	// Determine target tenant
	var targetTenantID *string

	if req.TenantCode != nil && *req.TenantCode != "" {
		// Creating user for specific tenant - admin must be system admin or tenant admin
		foundTenant, err := s.tenantRepo.GetByCode(*req.TenantCode)
		if err != nil {
			return nil, ErrTenantNotFound
		}

		// Check admin permissions for this tenant
		if !adminUser.IsSystemAdmin() && (adminUser.TenantID == nil || *adminUser.TenantID != foundTenant.ID) {
			return nil, ErrUnauthorized
		}

		targetTenantID = &foundTenant.ID
	} else {
		// Creating system admin - only system admins can do this
		if !adminUser.IsSystemAdmin() {
			return nil, ErrUnauthorized
		}
	}

	// Check if email already exists
	exists, err := s.userRepo.ExistsByEmail(targetTenantID, req.Email)
	if err != nil {
		g.Log().Error(ctx, "Failed to check email existence:", err)
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyExists
	}

	// Check if username already exists in the target tenant
	exists, err = s.userRepo.ExistsByUsername(targetTenantID, req.Username)
	if err != nil {
		g.Log().Error(ctx, "Failed to check username existence:", err)
		return nil, err
	}
	if exists {
		return nil, ErrUsernameAlreadyExists
	}

	// Create new user
	newUser := &user.User{
		ID:        s.generateUserID(),
		TenantID:  targetTenantID,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Status:    user.StatusActive,
	}

	// Set password
	if err := newUser.SetPassword(req.Password); err != nil {
		g.Log().Error(ctx, "Failed to hash password:", err)
		return nil, err
	}

	// Save user to database
	if err := s.userRepo.Create(newUser); err != nil {
		g.Log().Error(ctx, "Failed to create user:", err)
		return nil, err
	}

	// Get user roles for response (simulated for now)
	roles := s.getUserRoles(ctx, newUser.ID)
	isAdmin := newUser.IsSystemAdmin() || s.isUserAdmin(ctx, newUser.ID, newUser.TenantID)

	return &RegisterResponse{
		User: &UserInfo{
			ID:       newUser.ID,
			TenantID: newUser.TenantID,
			Username: newUser.Username,
			Email:    newUser.Email,
			Profile:  newUser.GetProfile(),
			Status:   newUser.Status,
			Roles:    roles,
			IsAdmin:  isAdmin,
		},
	}, nil
}

// Private helper methods (simulated data access)

func (s *AuthService) getUserByEmailAndTenant(ctx context.Context, email string, tenantCode *string) (*user.User, *tenant.Tenant, error) {
	var targetTenant *tenant.Tenant
	var targetTenantID *string
	
	// If tenant code is provided, find the tenant
	if tenantCode != nil && *tenantCode != "" {
		foundTenant, err := s.tenantRepo.GetByCode(*tenantCode)
		if err != nil {
			g.Log().Warning(ctx, "Tenant not found by code:", *tenantCode, "error:", err)
			return nil, nil, ErrTenantNotFound
		}
		targetTenant = foundTenant
		targetTenantID = &foundTenant.ID
	}
	// If no tenant code provided, user must be system admin (tenantID will be nil)
	
	// Find user by email and tenant ID
	foundUser, err := s.userRepo.GetByEmail(targetTenantID, email)
	if err != nil {
		g.Log().Warning(ctx, "User not found by email:", email, "tenantID:", targetTenantID, "error:", err)
		return nil, nil, ErrInvalidCredentials
	}
	
	return foundUser, targetTenant, nil
}

func (s *AuthService) getUserByID(ctx context.Context, userID string) (*user.User, *tenant.Tenant, error) {
	foundUser, err := s.userRepo.GetByID(userID)
	if err != nil {
		g.Log().Warning(ctx, "User not found by ID:", userID, "error:", err)
		return nil, nil, ErrInvalidCredentials
	}
	
	var foundTenant *tenant.Tenant
	if foundUser.TenantID != nil {
		foundTenant, err = s.tenantRepo.GetByID(*foundUser.TenantID)
		if err != nil {
			g.Log().Warning(ctx, "Tenant not found for user:", userID, "tenantID:", *foundUser.TenantID, "error:", err)
		}
	}
	
	return foundUser, foundTenant, nil
}

func (s *AuthService) checkLoginAttempts(ctx context.Context, userID string) error {
	// Check Redis for failed login attempts
	key := "login_attempts:" + userID
	attempts, err := g.Redis().Get(ctx, key)
	if err != nil {
		// If Redis error, allow login (fail open)
		return nil
	}
	
	if attempts.Int() >= 5 {
		return ErrTooManyAttempts
	}
	return nil
}

func (s *AuthService) recordFailedAttempt(ctx context.Context, userID string) {
	key := "login_attempts:" + userID
	g.Redis().Incr(ctx, key)
	g.Redis().Expire(ctx, key, int64(15*time.Minute/time.Second)) // Lock for 15 minutes
}

func (s *AuthService) clearFailedAttempts(ctx context.Context, userID string) {
	key := "login_attempts:" + userID
	g.Redis().Del(ctx, key)
}

func (s *AuthService) getUserRoles(ctx context.Context, userID string) []string {
	// Simulated - would query user_roles and roles tables
	return []string{"user"}
}

func (s *AuthService) isUserAdmin(ctx context.Context, userID string, tenantID *string) bool {
	// Simulated - would check if user has admin roles
	return false
}

func (s *AuthService) updateUserLastLogin(ctx context.Context, user *user.User) error {
	return s.userRepo.Update(user)
}

// GetUserByID retrieves user by ID (public method for handler use)
func (s *AuthService) GetUserByID(ctx context.Context, userID string) (*user.User, *tenant.Tenant, error) {
	return s.getUserByID(ctx, userID)
}

func (s *AuthService) generateUserID() string {
	// Generate UUID for user ID
	return guid.S()
}