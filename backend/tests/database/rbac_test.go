package database

import (
	"testing"
	"time"

	"github.com/gofromzero/project_temp/backend/domain/auth"
	"github.com/stretchr/testify/assert"
)

func TestRoleEntity(t *testing.T) {
	t.Run("创建角色实体", func(t *testing.T) {
		tenantID := "tenant-123"
		description := "Test role description"

		role := &auth.Role{
			ID:          "role-456",
			TenantID:    &tenantID,
			Name:        "Test Role",
			Code:        "test-role",
			Description: &description,
			IsSystem:    false,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		assert.Equal(t, "role-456", role.ID)
		assert.Equal(t, "tenant-123", *role.TenantID)
		assert.Equal(t, "Test Role", role.Name)
		assert.Equal(t, "test-role", role.Code)
		assert.Equal(t, "Test role description", *role.Description)
		assert.False(t, role.IsSystem)
		assert.True(t, role.IsTenantRole())
		assert.False(t, role.IsSystemRole())
		assert.Equal(t, "roles", role.TableName())
	})

	t.Run("系统内置角色", func(t *testing.T) {
		systemRole := &auth.Role{
			ID:       "system-admin-role",
			TenantID: nil, // 系统角色没有租户ID
			Name:     "System Administrator",
			Code:     "system-admin",
			IsSystem: true,
		}

		assert.True(t, systemRole.IsSystemRole())
		assert.False(t, systemRole.IsTenantRole())
		assert.Nil(t, systemRole.TenantID)
	})
}

func TestPermissionEntity(t *testing.T) {
	t.Run("创建权限实体", func(t *testing.T) {
		description := "Create tenant permission"

		permission := &auth.Permission{
			ID:          "perm-123",
			Name:        "Create Tenant",
			Code:        "tenant.create",
			Resource:    "tenant",
			Action:      "create",
			Scope:       auth.ScopeSystem,
			Description: &description,
			IsSystem:    true,
		}

		assert.Equal(t, "perm-123", permission.ID)
		assert.Equal(t, "Create Tenant", permission.Name)
		assert.Equal(t, "tenant.create", permission.Code)
		assert.Equal(t, "tenant", permission.Resource)
		assert.Equal(t, "create", permission.Action)
		assert.Equal(t, auth.ScopeSystem, permission.Scope)
		assert.True(t, permission.IsSystemPermission())
		assert.Equal(t, "tenant.create", permission.GetFullCode())
		assert.Equal(t, "permissions", permission.TableName())
	})

	t.Run("权限范围验证", func(t *testing.T) {
		systemPerm := &auth.Permission{Scope: auth.ScopeSystem}
		tenantPerm := &auth.Permission{Scope: auth.ScopeTenant}
		selfPerm := &auth.Permission{Scope: auth.ScopeSelf}

		assert.Equal(t, auth.ScopeSystem, systemPerm.Scope)
		assert.Equal(t, auth.ScopeTenant, tenantPerm.Scope)
		assert.Equal(t, auth.ScopeSelf, selfPerm.Scope)
	})

	t.Run("权限代码格式验证", func(t *testing.T) {
		userReadPerm := &auth.Permission{
			Resource: "user",
			Action:   "read",
		}

		tenantUpdatePerm := &auth.Permission{
			Resource: "tenant",
			Action:   "update",
		}

		assert.Equal(t, "user.read", userReadPerm.GetFullCode())
		assert.Equal(t, "tenant.update", tenantUpdatePerm.GetFullCode())
	})
}

func TestUserRoleEntity(t *testing.T) {
	t.Run("创建用户角色关联", func(t *testing.T) {
		userRole := &auth.UserRole{
			ID:        "user-role-123",
			UserID:    "user-456",
			RoleID:    "role-789",
			CreatedAt: time.Now(),
		}

		assert.Equal(t, "user-role-123", userRole.ID)
		assert.Equal(t, "user-456", userRole.UserID)
		assert.Equal(t, "role-789", userRole.RoleID)
		assert.Equal(t, "user_roles", userRole.TableName())
	})
}

func TestRolePermissionEntity(t *testing.T) {
	t.Run("创建角色权限关联", func(t *testing.T) {
		rolePermission := &auth.RolePermission{
			ID:           "role-perm-123",
			RoleID:       "role-456",
			PermissionID: "perm-789",
			CreatedAt:    time.Now(),
		}

		assert.Equal(t, "role-perm-123", rolePermission.ID)
		assert.Equal(t, "role-456", rolePermission.RoleID)
		assert.Equal(t, "perm-789", rolePermission.PermissionID)
		assert.Equal(t, "role_permissions", rolePermission.TableName())
	})
}

func TestPermissionScopes(t *testing.T) {
	t.Run("权限范围常量", func(t *testing.T) {
		assert.Equal(t, auth.PermissionScope("system"), auth.ScopeSystem)
		assert.Equal(t, auth.PermissionScope("tenant"), auth.ScopeTenant)
		assert.Equal(t, auth.PermissionScope("self"), auth.ScopeSelf)
	})
}

func TestRBACScenarios(t *testing.T) {
	t.Run("完整RBAC场景测试", func(t *testing.T) {
		// 1. 创建系统管理员角色
		systemAdminRole := &auth.Role{
			ID:       "system-admin",
			TenantID: nil,
			Name:     "System Administrator",
			Code:     "system-admin",
			IsSystem: true,
		}

		// 2. 创建租户管理员角色
		tenantID := "tenant-123"
		tenantAdminRole := &auth.Role{
			ID:       "tenant-admin",
			TenantID: &tenantID,
			Name:     "Tenant Administrator",
			Code:     "tenant-admin",
			IsSystem: false,
		}

		// 3. 创建权限
		createTenantPerm := &auth.Permission{
			ID:       "create-tenant",
			Name:     "Create Tenant",
			Code:     "tenant.create",
			Resource: "tenant",
			Action:   "create",
			Scope:    auth.ScopeSystem,
			IsSystem: true,
		}

		manageUsersPerm := &auth.Permission{
			ID:       "manage-users",
			Name:     "Manage Users",
			Code:     "user.manage",
			Resource: "user",
			Action:   "manage",
			Scope:    auth.ScopeTenant,
			IsSystem: false,
		}

		// 4. 验证角色属性
		assert.True(t, systemAdminRole.IsSystemRole())
		assert.False(t, systemAdminRole.IsTenantRole())
		assert.False(t, tenantAdminRole.IsSystemRole())
		assert.True(t, tenantAdminRole.IsTenantRole())

		// 5. 验证权限属性
		assert.True(t, createTenantPerm.IsSystemPermission())
		assert.Equal(t, auth.ScopeSystem, createTenantPerm.Scope)
		assert.False(t, manageUsersPerm.IsSystemPermission())
		assert.Equal(t, auth.ScopeTenant, manageUsersPerm.Scope)

		// 6. 创建关联关系
		systemAdminUserRole := &auth.UserRole{
			ID:     "sys-admin-user-role",
			UserID: "system-user-1",
			RoleID: systemAdminRole.ID,
		}

		systemAdminRolePerm := &auth.RolePermission{
			ID:           "sys-admin-role-perm",
			RoleID:       systemAdminRole.ID,
			PermissionID: createTenantPerm.ID,
		}

		// 验证关联关系
		assert.Equal(t, "system-user-1", systemAdminUserRole.UserID)
		assert.Equal(t, "system-admin", systemAdminUserRole.RoleID)
		assert.Equal(t, "system-admin", systemAdminRolePerm.RoleID)
		assert.Equal(t, "create-tenant", systemAdminRolePerm.PermissionID)
	})

	t.Run("多租户权限隔离验证", func(t *testing.T) {
		tenant1ID := "tenant-1"
		tenant2ID := "tenant-2"

		// 租户1的管理员角色
		tenant1AdminRole := &auth.Role{
			ID:       "tenant1-admin",
			TenantID: &tenant1ID,
			Name:     "Tenant 1 Admin",
			Code:     "admin",
			IsSystem: false,
		}

		// 租户2的管理员角色（同样的code但不同租户）
		tenant2AdminRole := &auth.Role{
			ID:       "tenant2-admin",
			TenantID: &tenant2ID,
			Name:     "Tenant 2 Admin",
			Code:     "admin", // 相同的代码，但在不同租户下
			IsSystem: false,
		}

		// 验证不同租户可以有相同的角色代码
		assert.Equal(t, tenant1AdminRole.Code, tenant2AdminRole.Code)
		assert.NotEqual(t, *tenant1AdminRole.TenantID, *tenant2AdminRole.TenantID)
		assert.True(t, tenant1AdminRole.IsTenantRole())
		assert.True(t, tenant2AdminRole.IsTenantRole())
	})
}
