package user

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

// UserStatus represents the possible status values for a user
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusLocked   UserStatus = "locked"
)

// ValidationError represents domain validation errors
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code"`
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Business rule constants
const (
	MinUsernameLength = 3
	MaxUsernameLength = 50
	MinPasswordLength = 8
	MaxPasswordLength = 128
	MaxFirstNameLength = 100
	MaxLastNameLength = 100
	MaxPhoneLength = 20
	MaxAvatarURLLength = 500
)

// Email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Phone validation regex (international format)
var phoneRegex = regexp.MustCompile(`^(\+\d{1,3}[- ]?)?\d{4,14}$`)

// Username validation regex (alphanumeric and underscore only)
var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// UserProfile represents the user profile information
type UserProfile struct {
	FirstName string  `json:"firstName"`
	LastName  string  `json:"lastName"`
	Avatar    *string `json:"avatar,omitempty"`
	Phone     *string `json:"phone,omitempty"`
}

// User represents the user domain entity
type User struct {
	ID             string     `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID       *string    `json:"tenantId,omitempty" gorm:"type:varchar(36);index;comment:null for system administrators"`
	Username       string     `json:"username" gorm:"type:varchar(100);not null"`
	Email          string     `json:"email" gorm:"type:varchar(255);not null"`
	HashedPassword string     `json:"-" gorm:"type:varchar(255);not null"`
	FirstName      string     `json:"firstName" gorm:"type:varchar(100);not null"`
	LastName       string     `json:"lastName" gorm:"type:varchar(100);not null"`
	Avatar         *string    `json:"avatar,omitempty" gorm:"type:varchar(500)"`
	Phone          *string    `json:"phone,omitempty" gorm:"type:varchar(20)"`
	Status         UserStatus `json:"status" gorm:"type:enum('active','inactive','locked');default:'active'"`
	LastLoginAt    *time.Time `json:"lastLoginAt,omitempty" gorm:"index"`
	CreatedAt      time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt      time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (User) TableName() string {
	return "users"
}

// IsActive checks if the user is active
func (u *User) IsActive() bool {
	return u.Status == StatusActive
}

// IsSystemAdmin checks if the user is a system administrator (no tenant ID)
func (u *User) IsSystemAdmin() bool {
	return u.TenantID == nil
}

// GetProfile returns the user profile information
func (u *User) GetProfile() UserProfile {
	return UserProfile{
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Avatar:    u.Avatar,
		Phone:     u.Phone,
	}
}

// SetPassword hashes and sets the user password
func (u *User) SetPassword(password string) error {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.HashedPassword = string(hashedBytes)
	return nil
}

// CheckPassword verifies the provided password against the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.HashedPassword), []byte(password))
	return err == nil
}

// GetFullName returns the user's full name
func (u *User) GetFullName() string {
	return u.FirstName + " " + u.LastName
}

// UpdateLastLogin sets the last login timestamp to current time
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
}

// ValidateUsername validates username according to business rules
func (u *User) ValidateUsername() error {
	if u.Username == "" {
		return &ValidationError{
			Field:   "username",
			Message: "用户名不能为空",
			Code:    "REQUIRED",
		}
	}

	if len(u.Username) < MinUsernameLength {
		return &ValidationError{
			Field:   "username",
			Message: fmt.Sprintf("用户名长度不能少于%d个字符", MinUsernameLength),
			Code:    "MIN_LENGTH",
		}
	}

	if len(u.Username) > MaxUsernameLength {
		return &ValidationError{
			Field:   "username",
			Message: fmt.Sprintf("用户名长度不能超过%d个字符", MaxUsernameLength),
			Code:    "MAX_LENGTH",
		}
	}

	if !usernameRegex.MatchString(u.Username) {
		return &ValidationError{
			Field:   "username",
			Message: "用户名只能包含字母、数字和下划线",
			Code:    "INVALID_FORMAT",
		}
	}

	// Check for reserved usernames
	reservedNames := []string{"admin", "root", "system", "administrator", "test", "demo"}
	lowerUsername := strings.ToLower(u.Username)
	for _, reserved := range reservedNames {
		if lowerUsername == reserved {
			return &ValidationError{
				Field:   "username",
				Message: "该用户名为系统保留用户名",
				Code:    "RESERVED_NAME",
			}
		}
	}

	return nil
}

// ValidateEmail validates email format and business rules
func (u *User) ValidateEmail() error {
	if u.Email == "" {
		return &ValidationError{
			Field:   "email",
			Message: "邮箱地址不能为空",
			Code:    "REQUIRED",
		}
	}

	if !emailRegex.MatchString(u.Email) {
		return &ValidationError{
			Field:   "email",
			Message: "邮箱地址格式不正确",
			Code:    "INVALID_FORMAT",
		}
	}

	// Normalize email to lowercase
	u.Email = strings.ToLower(strings.TrimSpace(u.Email))

	return nil
}

// ValidatePassword validates password strength according to business rules
func ValidatePassword(password string) error {
	if password == "" {
		return &ValidationError{
			Field:   "password",
			Message: "密码不能为空",
			Code:    "REQUIRED",
		}
	}

	if len(password) < MinPasswordLength {
		return &ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("密码长度不能少于%d个字符", MinPasswordLength),
			Code:    "MIN_LENGTH",
		}
	}

	if len(password) > MaxPasswordLength {
		return &ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("密码长度不能超过%d个字符", MaxPasswordLength),
			Code:    "MAX_LENGTH",
		}
	}

	// Check password complexity
	var hasLower, hasNumber bool
	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}

	if !hasLower || !hasNumber {
		return &ValidationError{
			Field:   "password",
			Message: "密码必须包含至少一个小写字母和一个数字",
			Code:    "WEAK_PASSWORD",
		}
	}

	// Check for common weak passwords (before complexity check)
	weakPasswords := []string{"password", "123456", "admin123", "qwerty"}
	lowerPassword := strings.ToLower(password)
	for _, weak := range weakPasswords {
		if lowerPassword == weak {
			return &ValidationError{
				Field:   "password",
				Message: "密码过于简单，请使用更复杂的密码",
				Code:    "COMMON_PASSWORD",
			}
		}
	}

	return nil
}

// ValidateProfile validates user profile information
func (u *User) ValidateProfile() error {
	if u.FirstName == "" {
		return &ValidationError{
			Field:   "firstName",
			Message: "姓不能为空",
			Code:    "REQUIRED",
		}
	}

	if len(u.FirstName) > MaxFirstNameLength {
		return &ValidationError{
			Field:   "firstName",
			Message: fmt.Sprintf("姓的长度不能超过%d个字符", MaxFirstNameLength),
			Code:    "MAX_LENGTH",
		}
	}

	if u.LastName == "" {
		return &ValidationError{
			Field:   "lastName",
			Message: "名不能为空",
			Code:    "REQUIRED",
		}
	}

	if len(u.LastName) > MaxLastNameLength {
		return &ValidationError{
			Field:   "lastName",
			Message: fmt.Sprintf("名的长度不能超过%d个字符", MaxLastNameLength),
			Code:    "MAX_LENGTH",
		}
	}

	// Validate phone if provided
	if u.Phone != nil && *u.Phone != "" {
		if len(*u.Phone) > MaxPhoneLength {
			return &ValidationError{
				Field:   "phone",
				Message: fmt.Sprintf("电话号码长度不能超过%d个字符", MaxPhoneLength),
				Code:    "MAX_LENGTH",
			}
		}

		if !phoneRegex.MatchString(*u.Phone) {
			return &ValidationError{
				Field:   "phone",
				Message: "电话号码格式不正确",
				Code:    "INVALID_FORMAT",
			}
		}
	}

	// Validate avatar URL if provided
	if u.Avatar != nil && *u.Avatar != "" {
		if len(*u.Avatar) > MaxAvatarURLLength {
			return &ValidationError{
				Field:   "avatar",
				Message: fmt.Sprintf("头像URL长度不能超过%d个字符", MaxAvatarURLLength),
				Code:    "MAX_LENGTH",
			}
		}

		// Basic URL format validation
		avatar := strings.TrimSpace(*u.Avatar)
		if !strings.HasPrefix(avatar, "http://") && !strings.HasPrefix(avatar, "https://") {
			return &ValidationError{
				Field:   "avatar",
				Message: "头像URL格式不正确，必须以http://或https://开头",
				Code:    "INVALID_FORMAT",
			}
		}
		u.Avatar = &avatar
	}

	return nil
}

// ValidateStatusTransition validates if status transition is allowed
func (u *User) ValidateStatusTransition(newStatus UserStatus) error {
	if u.Status == newStatus {
		return nil // No change
	}

	// Define allowed status transitions
	allowedTransitions := map[UserStatus][]UserStatus{
		StatusActive:   {StatusInactive, StatusLocked},
		StatusInactive: {StatusActive, StatusLocked},
		StatusLocked:   {StatusActive, StatusInactive},
	}

	allowed, exists := allowedTransitions[u.Status]
	if !exists {
		return &ValidationError{
			Field:   "status",
			Message: fmt.Sprintf("无效的当前状态: %s", u.Status),
			Code:    "INVALID_CURRENT_STATUS",
		}
	}

	for _, validStatus := range allowed {
		if newStatus == validStatus {
			return nil
		}
	}

	return &ValidationError{
		Field:   "status",
		Message: fmt.Sprintf("不允许从%s状态转换到%s状态", u.Status, newStatus),
		Code:    "INVALID_TRANSITION",
	}
}

// CanBeDeleted checks if user can be safely deleted
func (u *User) CanBeDeleted() error {
	// System admins cannot be deleted
	if u.IsSystemAdmin() {
		return &ValidationError{
			Field:   "user",
			Message: "系统管理员用户不能被删除",
			Code:    "SYSTEM_ADMIN_UNDELETABLE",
		}
	}

	// Cannot delete locked users (they should be unlocked first)
	if u.Status == StatusLocked {
		return &ValidationError{
			Field:   "status",
			Message: "已锁定的用户不能直接删除，请先解锁",
			Code:    "LOCKED_USER_UNDELETABLE",
		}
	}

	return nil
}

// Validate performs comprehensive validation of the user entity
func (u *User) Validate() error {
	// Validate username
	if err := u.ValidateUsername(); err != nil {
		return err
	}

	// Validate email
	if err := u.ValidateEmail(); err != nil {
		return err
	}

	// Validate profile
	if err := u.ValidateProfile(); err != nil {
		return err
	}

	// Validate status
	if u.Status != StatusActive && u.Status != StatusInactive && u.Status != StatusLocked {
		return &ValidationError{
			Field:   "status",
			Message: "用户状态必须是active、inactive或locked之一",
			Code:    "INVALID_STATUS",
		}
	}

	return nil
}

// SetPasswordWithValidation hashes and sets the user password with validation
func (u *User) SetPasswordWithValidation(password string) error {
	// Validate password first
	if err := ValidatePassword(password); err != nil {
		return err
	}

	// Hash and set password
	return u.SetPassword(password)
}

// UpdateStatus changes user status with validation
func (u *User) UpdateStatus(newStatus UserStatus) error {
	// Validate transition
	if err := u.ValidateStatusTransition(newStatus); err != nil {
		return err
	}

	u.Status = newStatus
	u.UpdatedAt = time.Now()
	return nil
}

// Lock locks the user account
func (u *User) Lock() error {
	return u.UpdateStatus(StatusLocked)
}

// Unlock unlocks the user account
func (u *User) Unlock() error {
	if u.Status == StatusLocked {
		return u.UpdateStatus(StatusActive)
	}
	return nil
}

// Deactivate deactivates the user account
func (u *User) Deactivate() error {
	return u.UpdateStatus(StatusInactive)
}

// Activate activates the user account
func (u *User) Activate() error {
	return u.UpdateStatus(StatusActive)
}

// UserRepository defines the repository interface for user operations
type UserRepository interface {
	Create(user *User) error
	GetByID(id string) (*User, error)
	GetByUsername(tenantID *string, username string) (*User, error)
	GetByEmail(tenantID *string, email string) (*User, error)
	GetByTenantID(tenantID string, offset, limit int) ([]*User, error)
	GetSystemAdmins(offset, limit int) ([]*User, error)
	Update(user *User) error
	Delete(id string) error
	Count(tenantID *string) (int64, error)
	ExistsByUsername(tenantID *string, username string) (bool, error)
	ExistsByEmail(tenantID *string, email string) (bool, error)
}
