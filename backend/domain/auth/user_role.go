package auth

import "time"

// UserRole represents the many-to-many relationship between users and roles
type UserRole struct {
	ID        string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	UserID    string    `json:"userId" gorm:"type:varchar(36);not null;index"`
	RoleID    string    `json:"roleId" gorm:"type:varchar(36);not null;index"`
	CreatedAt time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// TableName returns the table name for GORM
func (UserRole) TableName() string {
	return "user_roles"
}

// UserRoleRepository defines the repository interface for user-role associations
type UserRoleRepository interface {
	Create(userRole *UserRole) error
	GetByID(id string) (*UserRole, error)
	GetRolesByUserID(userID string) ([]*Role, error)
	GetUsersByRoleID(roleID string) ([]string, error)
	Delete(id string) error
	DeleteByUserIDAndRoleID(userID, roleID string) error
	ExistsByUserIDAndRoleID(userID, roleID string) (bool, error)
	GetUserRoleAssociations(userID string) ([]*UserRole, error)
}