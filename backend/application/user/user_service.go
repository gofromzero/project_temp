package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainUser "github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/gofromzero/project_temp/backend/infr/repository/mysql"
	"github.com/gogf/gf/v2/util/guid"
)

// UserService provides high-level user operations
type UserService struct {
	UserRepository domainUser.UserRepository // Exported for testing
}

// NewUserService creates a new user application service
func NewUserService() *UserService {
	return &UserService{
		UserRepository: mysql.NewUserRepository(),
	}
}

// NewUserServiceWithRepository creates a new user service with custom repository (for testing)
func NewUserServiceWithRepository(repo domainUser.UserRepository) *UserService {
	return &UserService{
		UserRepository: repo,
	}
}

// CreateUserRequest represents the request to create a new user
type CreateUserRequest struct {
	TenantID  *string `json:"tenantId,omitempty"`   // 系统管理员可指定，租户管理员自动设置
	Username  string  `json:"username" validate:"required,min=3,max=50"`
	Email     string  `json:"email" validate:"required,email"`
	Password  string  `json:"password" validate:"required,min=8"`
	FirstName string  `json:"firstName" validate:"required,max=100"`
	LastName  string  `json:"lastName" validate:"required,max=100"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,max=20"`
	Avatar    *string `json:"avatar,omitempty" validate:"omitempty,url,max=500"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	FirstName *string `json:"firstName,omitempty" validate:"omitempty,max=100"`
	LastName  *string `json:"lastName,omitempty" validate:"omitempty,max=100"`
	Phone     *string `json:"phone,omitempty" validate:"omitempty,max=20"`
	Avatar    *string `json:"avatar,omitempty" validate:"omitempty,url,max=500"`
	Status    *string `json:"status,omitempty" validate:"omitempty,oneof=active inactive locked"`
}

// UserResponse represents user information in API responses
type UserResponse struct {
	ID          string     `json:"id"`
	TenantID    *string    `json:"tenantId,omitempty"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	FirstName   string     `json:"firstName"`
	LastName    string     `json:"lastName"`
	Avatar      *string    `json:"avatar,omitempty"`
	Phone       *string    `json:"phone,omitempty"`
	Status      string     `json:"status"`
	LastLoginAt *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// Pagination represents pagination information
type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
	Pages int `json:"pages"`
}

// ListUsersRequest represents request parameters for listing users
type ListUsersRequest struct {
	TenantID *string `json:"tenantId,omitempty"` // 系统管理员可指定租户，租户管理员自动设置
	Page     int     `json:"page,omitempty" validate:"omitempty,min=1"`
	Limit    int     `json:"limit,omitempty" validate:"omitempty,min=1,max=100"`
	Status   string  `json:"status,omitempty" validate:"omitempty,oneof=active inactive locked"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Users      []*UserResponse `json:"users"`
	Pagination Pagination      `json:"pagination"`
}

// CreateUser creates a new user with validation and tenant isolation
func (s *UserService) CreateUser(ctx context.Context, req CreateUserRequest) (*UserResponse, error) {

	// Check username uniqueness within tenant
	exists, err := s.UserRepository.ExistsByUsername(req.TenantID, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username uniqueness: %w", err)
	}
	if exists {
		return nil, errors.New("username already exists in this tenant")
	}

	// Check email uniqueness within tenant
	exists, err = s.UserRepository.ExistsByEmail(req.TenantID, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email uniqueness: %w", err)
	}
	if exists {
		return nil, errors.New("email already exists in this tenant")
	}

	// Create user entity
	user := &domainUser.User{
		ID:        guid.S(),
		TenantID:  req.TenantID,
		Username:  req.Username,
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Phone:     req.Phone,
		Avatar:    req.Avatar,
		Status:    domainUser.StatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set password with validation
	if err := user.SetPasswordWithValidation(req.Password); err != nil {
		return nil, err
	}

	// Validate the complete user entity
	if err := user.Validate(); err != nil {
		return nil, err
	}

	// Create user in repository
	if err := s.UserRepository.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// GetUsers retrieves users with pagination and tenant filtering
func (s *UserService) GetUsers(ctx context.Context, req ListUsersRequest) (*ListUsersResponse, error) {
	// Set default pagination
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Limit <= 0 {
		req.Limit = 10
	}

	// Calculate offset
	offset := (req.Page - 1) * req.Limit

	var users []*domainUser.User
	var err error

	// Get users based on tenant context
	if req.TenantID != nil {
		// Get tenant users
		users, err = s.UserRepository.GetByTenantID(*req.TenantID, offset, req.Limit)
	} else {
		// Get system admins
		users, err = s.UserRepository.GetSystemAdmins(offset, req.Limit)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to retrieve users: %w", err)
	}

	// Apply status filter if specified
	if req.Status != "" {
		users = s.filterUsersByStatus(users, req.Status)
	}

	// Get total count
	total, err := s.UserRepository.Count(req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user count: %w", err)
	}

	// Calculate pagination
	pages := int((total + int64(req.Limit) - 1) / int64(req.Limit))

	// Convert to response format
	userResponses := make([]*UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = s.toUserResponse(user)
	}

	return &ListUsersResponse{
		Users: userResponses,
		Pagination: Pagination{
			Page:  req.Page,
			Limit: req.Limit,
			Total: int(total),
			Pages: pages,
		},
	}, nil
}

// GetUserByID retrieves a user by ID with tenant boundary validation
func (s *UserService) GetUserByID(ctx context.Context, userID string, requesterTenantID *string) (*UserResponse, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	user, err := s.UserRepository.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Validate tenant boundary
	if err := s.validateTenantBoundary(user, requesterTenantID); err != nil {
		return nil, err
	}

	return s.toUserResponse(user), nil
}

// UpdateUser updates user information with validation
func (s *UserService) UpdateUser(ctx context.Context, userID string, req UpdateUserRequest, requesterTenantID *string) (*UserResponse, error) {
	if userID == "" {
		return nil, errors.New("user ID is required")
	}

	// Get existing user
	user, err := s.UserRepository.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Validate tenant boundary
	if err := s.validateTenantBoundary(user, requesterTenantID); err != nil {
		return nil, err
	}

	// Apply updates
	if err := s.applyUserUpdates(user, req); err != nil {
		return nil, err
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Save changes
	if err := s.UserRepository.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return s.toUserResponse(user), nil
}

// DeleteUser deletes a user with validation
func (s *UserService) DeleteUser(ctx context.Context, userID string, requesterTenantID *string) error {
	if userID == "" {
		return errors.New("user ID is required")
	}

	// Get existing user to validate tenant boundary
	user, err := s.UserRepository.GetByID(userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Validate tenant boundary
	if err := s.validateTenantBoundary(user, requesterTenantID); err != nil {
		return err
	}

	// Check if user can be deleted using domain business rules
	if err := user.CanBeDeleted(); err != nil {
		return err
	}

	// Delete user
	if err := s.UserRepository.Delete(userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}


// validateTenantBoundary ensures users can only access data within their tenant scope
func (s *UserService) validateTenantBoundary(user *domainUser.User, requesterTenantID *string) error {
	// System admin (no tenant ID) can access all users
	if requesterTenantID == nil {
		return nil
	}

	// Tenant admin can only access users in their tenant
	if user.TenantID == nil || *user.TenantID != *requesterTenantID {
		return errors.New("access denied: user not in your tenant")
	}

	return nil
}

// applyUserUpdates applies update request to user entity
func (s *UserService) applyUserUpdates(user *domainUser.User, req UpdateUserRequest) error {
	if req.Email != nil {
		// Check email uniqueness within tenant
		exists, err := s.UserRepository.ExistsByEmail(user.TenantID, *req.Email)
		if err != nil {
			return fmt.Errorf("failed to check email uniqueness: %w", err)
		}
		if exists && *req.Email != user.Email {
			return errors.New("email already exists in this tenant")
		}
		user.Email = *req.Email
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}

	if req.LastName != nil {
		user.LastName = *req.LastName
	}

	if req.Phone != nil {
		user.Phone = req.Phone
	}

	if req.Avatar != nil {
		user.Avatar = req.Avatar
	}

	if req.Status != nil {
		status := domainUser.UserStatus(*req.Status)
		if err := user.UpdateStatus(status); err != nil {
			return err
		}
	}

	return nil
}

// toUserResponse converts domain user to API response format
func (s *UserService) toUserResponse(user *domainUser.User) *UserResponse {
	return &UserResponse{
		ID:          user.ID,
		TenantID:    user.TenantID,
		Username:    user.Username,
		Email:       user.Email,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Avatar:      user.Avatar,
		Phone:       user.Phone,
		Status:      string(user.Status),
		LastLoginAt: user.LastLoginAt,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}

// filterUsersByStatus filters users by status
func (s *UserService) filterUsersByStatus(users []*domainUser.User, status string) []*domainUser.User {
	var filtered []*domainUser.User
	targetStatus := domainUser.UserStatus(status)

	for _, user := range users {
		if user.Status == targetStatus {
			filtered = append(filtered, user)
		}
	}

	return filtered
}

