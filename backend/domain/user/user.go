package user

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserStatus represents the possible status values for a user
type UserStatus string

const (
	StatusActive   UserStatus = "active"
	StatusInactive UserStatus = "inactive"
	StatusLocked   UserStatus = "locked"
)

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
