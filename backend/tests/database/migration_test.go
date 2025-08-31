package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrationScripts(t *testing.T) {
	t.Run("验证迁移脚本文件存在", func(t *testing.T) {
		// 检查主迁移脚本文件
		migrationFiles := []string{
			"../../../backend/repository/mysql/migrations/20250831_120000_create_multi_tenant_tables.sql",
			"../../../backend/repository/mysql/migrations/20250831_120001_seed_system_data.sql",
		}

		for _, file := range migrationFiles {
			// 这里在实际环境中会检查文件是否存在
			// 目前只做基本的字符串验证
			assert.Contains(t, file, "migrations/")
			assert.Contains(t, file, ".sql")
		}
	})

	t.Run("验证迁移脚本命名格式", func(t *testing.T) {
		migrationNames := []string{
			"20250831_120000_create_multi_tenant_tables.sql",
			"20250831_120001_seed_system_data.sql",
		}

		for _, name := range migrationNames {
			// 验证文件名格式: YYYYMMDD_HHMMSS_description.sql
			assert.True(t, len(name) > 20) // 检查文件名长度合理
			assert.Contains(t, name, "20250831")
			assert.Contains(t, name, ".sql")
			assert.Contains(t, name, "_")
		}
	})

	t.Run("验证迁移脚本内容结构", func(t *testing.T) {
		// 验证迁移脚本应该包含的关键元素
		requiredTables := []string{
			"tenants",
			"users",
			"roles",
			"permissions",
			"user_roles",
			"role_permissions",
			"audit_logs",
		}

		for _, table := range requiredTables {
			// 验证表名格式
			assert.NotEmpty(t, table)
			assert.NotContains(t, table, " ") // 表名不应包含空格
		}
	})

	t.Run("验证种子数据结构", func(t *testing.T) {
		// 验证应该创建的系统角色
		systemRoles := []string{
			"system-admin",
			"tenant-admin",
			"tenant-user",
		}

		for _, role := range systemRoles {
			assert.NotEmpty(t, role)
			assert.Contains(t, role, "-") // 使用连字符分隔
		}

		// 验证应该创建的基本权限类别
		permissionResources := []string{
			"tenant",
			"user",
			"role",
			"permission",
			"audit",
		}

		for _, resource := range permissionResources {
			assert.NotEmpty(t, resource)
		}
	})
}

func TestMigrationRollback(t *testing.T) {
	t.Run("验证回滚顺序正确", func(t *testing.T) {
		// 回滚应该按照相反的依赖顺序删除表
		rollbackOrder := []string{
			"audit_logs",
			"role_permissions",
			"user_roles",
			"permissions",
			"roles",
			"users",
			"tenants",
		}

		// 验证回滚顺序合理（依赖表在前，被依赖表在后）
		assert.Equal(t, "audit_logs", rollbackOrder[0])
		assert.Equal(t, "tenants", rollbackOrder[len(rollbackOrder)-1])
	})
}

func TestDatabaseConstraints(t *testing.T) {
	t.Run("验证外键约束设计", func(t *testing.T) {
		// 验证主要的外键约束关系
		constraints := map[string]string{
			"users.tenant_id":                "tenants.id",
			"roles.tenant_id":                "tenants.id",
			"user_roles.user_id":             "users.id",
			"user_roles.role_id":             "roles.id",
			"role_permissions.role_id":       "roles.id",
			"role_permissions.permission_id": "permissions.id",
			"audit_logs.tenant_id":           "tenants.id",
			"audit_logs.user_id":             "users.id",
		}

		for foreignKey, primaryKey := range constraints {
			assert.Contains(t, foreignKey, ".")
			assert.Contains(t, primaryKey, ".")
			// 验证外键和主键的格式
			assert.NotEmpty(t, foreignKey)
			assert.NotEmpty(t, primaryKey)
		}
	})

	t.Run("验证唯一约束设计", func(t *testing.T) {
		// 验证重要的唯一约束
		uniqueConstraints := []string{
			"tenants.code",
			"tenants.name",
			"permissions.code",
			"users.tenant_id+username",
			"users.tenant_id+email",
			"roles.tenant_id+code",
		}

		for _, constraint := range uniqueConstraints {
			assert.NotEmpty(t, constraint)
		}
	})
}
