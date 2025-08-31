package auth

import "time"

// Role represents the role domain entity in RBAC model
type Role struct {
	ID          string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID    *string   `json:"tenantId,omitempty" gorm:"type:varchar(36);index;comment:null for system roles"`
	Name        string    `json:"name" gorm:"type:varchar(100);not null"`
	Code        string    `json:"code" gorm:"type:varchar(100);not null"`
	Description *string   `json:"description,omitempty" gorm:"type:text"`
	IsSystem    bool      `json:"isSystem" gorm:"default:false;comment:system built-in role"`
	CreatedAt   time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (Role) TableName() string {
	return "roles"
}

// IsSystemRole checks if the role is a system built-in role
func (r *Role) IsSystemRole() bool {
	return r.IsSystem
}

// IsTenantRole checks if the role belongs to a specific tenant
func (r *Role) IsTenantRole() bool {
	return r.TenantID != nil
}

// RoleRepository defines the repository interface for role operations
type RoleRepository interface {
	Create(role *Role) error
	GetByID(id string) (*Role, error)
	GetByCode(tenantID *string, code string) (*Role, error)
	GetByTenantID(tenantID string, offset, limit int) ([]*Role, error)
	GetSystemRoles(offset, limit int) ([]*Role, error)
	Update(role *Role) error
	Delete(id string) error
	Count(tenantID *string) (int64, error)
	ExistsByCode(tenantID *string, code string) (bool, error)
}