-- +goose Up
-- +goose StatementBegin

-- Create tenants table
CREATE TABLE `tenants` (
  `id` varchar(36) NOT NULL COMMENT '租户唯一标识符',
  `name` varchar(255) NOT NULL COMMENT '租户名称',
  `code` varchar(100) NOT NULL COMMENT '租户代码，用于子域名等',
  `status` enum('active','suspended','disabled') NOT NULL DEFAULT 'active' COMMENT '租户状态',
  `config` json NOT NULL COMMENT '租户配置信息',
  `admin_user_id` varchar(36) DEFAULT NULL COMMENT '管理员用户ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_tenants_code` (`code`),
  UNIQUE KEY `uk_tenants_name` (`name`),
  KEY `idx_tenants_status` (`status`),
  KEY `idx_tenants_admin_user_id` (`admin_user_id`),
  KEY `idx_tenants_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='租户表';

-- Create users table with multi-tenant support
CREATE TABLE `users` (
  `id` varchar(36) NOT NULL COMMENT '用户唯一标识符',
  `tenant_id` varchar(36) DEFAULT NULL COMMENT '所属租户ID，为空表示系统管理员',
  `username` varchar(100) NOT NULL COMMENT '用户名',
  `email` varchar(255) NOT NULL COMMENT '邮箱地址',
  `hashed_password` varchar(255) NOT NULL COMMENT '加密后的密码',
  `first_name` varchar(100) NOT NULL COMMENT '名',
  `last_name` varchar(100) NOT NULL COMMENT '姓',
  `avatar` varchar(500) DEFAULT NULL COMMENT '头像URL',
  `phone` varchar(20) DEFAULT NULL COMMENT '电话号码',
  `status` enum('active','inactive','locked') NOT NULL DEFAULT 'active' COMMENT '用户状态',
  `last_login_at` datetime DEFAULT NULL COMMENT '最后登录时间',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_users_tenant_id` (`tenant_id`),
  KEY `idx_users_status` (`status`),
  KEY `idx_users_last_login_at` (`last_login_at`),
  KEY `idx_users_created_at` (`created_at`),
  UNIQUE KEY `uk_users_tenant_username` (`tenant_id`, `username`),
  UNIQUE KEY `uk_users_tenant_email` (`tenant_id`, `email`),
  CONSTRAINT `fk_users_tenant_id` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- Create roles table for RBAC
CREATE TABLE `roles` (
  `id` varchar(36) NOT NULL COMMENT '角色唯一标识符',
  `tenant_id` varchar(36) DEFAULT NULL COMMENT '所属租户ID，为空表示系统角色',
  `name` varchar(100) NOT NULL COMMENT '角色名称',
  `code` varchar(100) NOT NULL COMMENT '角色代码',
  `description` text DEFAULT NULL COMMENT '角色描述',
  `is_system` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为系统内置角色',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_roles_tenant_id` (`tenant_id`),
  KEY `idx_roles_is_system` (`is_system`),
  KEY `idx_roles_created_at` (`created_at`),
  UNIQUE KEY `uk_roles_tenant_code` (`tenant_id`, `code`),
  CONSTRAINT `fk_roles_tenant_id` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色表';

-- Create permissions table for RBAC
CREATE TABLE `permissions` (
  `id` varchar(36) NOT NULL COMMENT '权限唯一标识符',
  `name` varchar(100) NOT NULL COMMENT '权限名称',
  `code` varchar(100) NOT NULL COMMENT '权限代码，格式: resource.action',
  `resource` varchar(50) NOT NULL COMMENT '资源名称',
  `action` varchar(50) NOT NULL COMMENT '操作类型',
  `scope` enum('system','tenant','self') NOT NULL COMMENT '权限范围',
  `description` text DEFAULT NULL COMMENT '权限描述',
  `is_system` tinyint(1) NOT NULL DEFAULT 0 COMMENT '是否为系统内置权限',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_permissions_code` (`code`),
  KEY `idx_permissions_resource` (`resource`),
  KEY `idx_permissions_action` (`action`),
  KEY `idx_permissions_scope` (`scope`),
  KEY `idx_permissions_is_system` (`is_system`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';

-- Create user_roles junction table
CREATE TABLE `user_roles` (
  `id` varchar(36) NOT NULL COMMENT '关联唯一标识符',
  `user_id` varchar(36) NOT NULL COMMENT '用户ID',
  `role_id` varchar(36) NOT NULL COMMENT '角色ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_roles_user_id` (`user_id`),
  KEY `idx_user_roles_role_id` (`role_id`),
  KEY `idx_user_roles_created_at` (`created_at`),
  UNIQUE KEY `uk_user_roles_user_role` (`user_id`, `role_id`),
  CONSTRAINT `fk_user_roles_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_user_roles_role_id` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户角色关联表';

-- Create role_permissions junction table
CREATE TABLE `role_permissions` (
  `id` varchar(36) NOT NULL COMMENT '关联唯一标识符',
  `role_id` varchar(36) NOT NULL COMMENT '角色ID',
  `permission_id` varchar(36) NOT NULL COMMENT '权限ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_role_permissions_role_id` (`role_id`),
  KEY `idx_role_permissions_permission_id` (`permission_id`),
  KEY `idx_role_permissions_created_at` (`created_at`),
  UNIQUE KEY `uk_role_permissions_role_permission` (`role_id`, `permission_id`),
  CONSTRAINT `fk_role_permissions_role_id` FOREIGN KEY (`role_id`) REFERENCES `roles` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_role_permissions_permission_id` FOREIGN KEY (`permission_id`) REFERENCES `permissions` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='角色权限关联表';

-- Create audit_logs table
CREATE TABLE `audit_logs` (
  `id` varchar(36) NOT NULL COMMENT '日志唯一标识符',
  `tenant_id` varchar(36) DEFAULT NULL COMMENT '租户ID',
  `user_id` varchar(36) NOT NULL COMMENT '操作用户ID',
  `action` varchar(50) NOT NULL COMMENT '操作类型',
  `resource` varchar(50) NOT NULL COMMENT '操作资源',
  `resource_id` varchar(36) DEFAULT NULL COMMENT '资源ID',
  `details` json NOT NULL COMMENT '操作详情',
  `ip_address` varchar(45) DEFAULT NULL COMMENT 'IP地址',
  `user_agent` varchar(500) DEFAULT NULL COMMENT '用户代理',
  `timestamp` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '操作时间',
  PRIMARY KEY (`id`),
  KEY `idx_audit_logs_tenant_id` (`tenant_id`),
  KEY `idx_audit_logs_user_id` (`user_id`),
  KEY `idx_audit_logs_action` (`action`),
  KEY `idx_audit_logs_resource` (`resource`),
  KEY `idx_audit_logs_resource_id` (`resource_id`),
  KEY `idx_audit_logs_timestamp` (`timestamp`),
  KEY `idx_audit_logs_tenant_timestamp` (`tenant_id`, `timestamp`),
  CONSTRAINT `fk_audit_logs_tenant_id` FOREIGN KEY (`tenant_id`) REFERENCES `tenants` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `fk_audit_logs_user_id` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='审计日志表';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS `audit_logs`;
DROP TABLE IF EXISTS `role_permissions`;
DROP TABLE IF EXISTS `user_roles`;
DROP TABLE IF EXISTS `permissions`;
DROP TABLE IF EXISTS `roles`;
DROP TABLE IF EXISTS `users`;
DROP TABLE IF EXISTS `tenants`;

-- +goose StatementEnd