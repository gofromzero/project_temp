package interfaces

import "github.com/gofromzero/project_temp/backend/domain/auth"

// RoleRepository defines the repository interface for role operations
type RoleRepository interface {
	auth.RoleRepository
}

// PermissionRepository defines the repository interface for permission operations
type PermissionRepository interface {
	auth.PermissionRepository
}

// UserRoleRepository defines the repository interface for user-role associations
type UserRoleRepository interface {
	auth.UserRoleRepository
}

// RolePermissionRepository defines the repository interface for role-permission associations
type RolePermissionRepository interface {
	auth.RolePermissionRepository
}
