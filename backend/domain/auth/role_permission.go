package auth

import "time"

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	ID           string    `json:"id" gorm:"type:varchar(36);primaryKey"`
	RoleID       string    `json:"roleId" gorm:"type:varchar(36);not null;index"`
	PermissionID string    `json:"permissionId" gorm:"type:varchar(36);not null;index"`
	CreatedAt    time.Time `json:"createdAt" gorm:"autoCreateTime"`
}

// TableName returns the table name for GORM
func (RolePermission) TableName() string {
	return "role_permissions"
}

// RolePermissionRepository defines the repository interface for role-permission associations
type RolePermissionRepository interface {
	Create(rolePermission *RolePermission) error
	GetByID(id string) (*RolePermission, error)
	GetPermissionsByRoleID(roleID string) ([]*Permission, error)
	GetRolesByPermissionID(permissionID string) ([]*Role, error)
	Delete(id string) error
	DeleteByRoleIDAndPermissionID(roleID, permissionID string) error
	ExistsByRoleIDAndPermissionID(roleID, permissionID string) (bool, error)
	GetRolePermissionAssociations(roleID string) ([]*RolePermission, error)
}