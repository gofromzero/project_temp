package main

import (
	"context"
	"os"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/glog"
)

// runMigrations executes database migrations
func runMigrations(ctx context.Context) error {
	db := g.DB()

	// Execute the main schema migration
	schemaMigration := `
-- Create tenants table
CREATE TABLE IF NOT EXISTS ` + "`tenants`" + ` (
  ` + "`id`" + ` varchar(36) NOT NULL COMMENT '租户唯一标识符',
  ` + "`name`" + ` varchar(255) NOT NULL COMMENT '租户名称',
  ` + "`code`" + ` varchar(100) NOT NULL COMMENT '租户代码，用于子域名等',
  ` + "`status`" + ` enum('active','suspended','disabled') NOT NULL DEFAULT 'active' COMMENT '租户状态',
  ` + "`config`" + ` json NOT NULL COMMENT '租户配置信息',
  ` + "`admin_user_id`" + ` varchar(36) DEFAULT NULL COMMENT '管理员用户ID',
  ` + "`created_at`" + ` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  ` + "`updated_at`" + ` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (` + "`id`" + `),
  UNIQUE KEY ` + "`uk_tenants_code`" + ` (` + "`code`" + `),
  UNIQUE KEY ` + "`uk_tenants_name`" + ` (` + "`name`" + `),
  KEY ` + "`idx_tenants_status`" + ` (` + "`status`" + `),
  KEY ` + "`idx_tenants_admin_user_id`" + ` (` + "`admin_user_id`" + `),
  KEY ` + "`idx_tenants_created_at`" + ` (` + "`created_at`" + `)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户表';

-- Create users table with multi-tenant support
CREATE TABLE IF NOT EXISTS ` + "`users`" + ` (
  ` + "`id`" + ` varchar(36) NOT NULL COMMENT '用户唯一标识符',
  ` + "`tenant_id`" + ` varchar(36) DEFAULT NULL COMMENT '所属租户ID，为空表示系统管理员',
  ` + "`username`" + ` varchar(100) NOT NULL COMMENT '用户名',
  ` + "`email`" + ` varchar(255) NOT NULL COMMENT '邮箱地址',
  ` + "`hashed_password`" + ` varchar(255) NOT NULL COMMENT '加密后的密码',
  ` + "`first_name`" + ` varchar(100) NOT NULL COMMENT '名',
  ` + "`last_name`" + ` varchar(100) NOT NULL COMMENT '姓',
  ` + "`avatar`" + ` varchar(500) DEFAULT NULL COMMENT '头像URL',
  ` + "`phone`" + ` varchar(20) DEFAULT NULL COMMENT '电话号码',
  ` + "`status`" + ` enum('active','inactive','locked') NOT NULL DEFAULT 'active' COMMENT '用户状态',
  ` + "`last_login_at`" + ` datetime DEFAULT NULL COMMENT '最后登录时间',
  ` + "`created_at`" + ` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  ` + "`updated_at`" + ` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (` + "`id`" + `),
  KEY ` + "`idx_users_tenant_id`" + ` (` + "`tenant_id`" + `),
  KEY ` + "`idx_users_status`" + ` (` + "`status`" + `),
  KEY ` + "`idx_users_last_login_at`" + ` (` + "`last_login_at`" + `),
  KEY ` + "`idx_users_created_at`" + ` (` + "`created_at`" + `),
  UNIQUE KEY ` + "`uk_users_tenant_username`" + ` (` + "`tenant_id`" + `, ` + "`username`" + `),
  UNIQUE KEY ` + "`uk_users_tenant_email`" + ` (` + "`tenant_id`" + `, ` + "`email`" + `)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- Create other tables...
-- (Additional table creation statements would go here)
`

	_, err := db.Exec(ctx, schemaMigration)
	if err != nil {
		return err
	}

	glog.Info(ctx, "Database migration completed successfully")
	return nil
}

// Migration command handler
func handleMigrate(ctx context.Context) {
	if len(os.Args) < 3 {
		glog.Info(ctx, "Usage: go run cmd/main.go migrate [up|down]")
		return
	}

	direction := os.Args[2]

	switch direction {
	case "up":
		err := runMigrations(ctx)
		if err != nil {
			glog.Fatal(ctx, "Migration failed:", err)
		}
		glog.Info(ctx, "Migration up completed")

	case "down":
		glog.Info(ctx, "Migration down - dropping all tables")
		err := rollbackMigrations(ctx)
		if err != nil {
			glog.Fatal(ctx, "Rollback failed:", err)
		}
		glog.Info(ctx, "Migration down completed")

	default:
		glog.Error(ctx, "Unknown migration direction:", direction)
	}
}

// rollbackMigrations drops all tables in reverse dependency order
func rollbackMigrations(ctx context.Context) error {
	db := g.DB()

	tables := []string{
		"audit_logs",
		"role_permissions",
		"user_roles",
		"permissions",
		"roles",
		"users",
		"tenants",
	}

	for _, table := range tables {
		_, err := db.Exec(ctx, "DROP TABLE IF EXISTS `"+table+"`")
		if err != nil {
			return err
		}
		glog.Info(ctx, "Dropped table:", table)
	}

	return nil
}
