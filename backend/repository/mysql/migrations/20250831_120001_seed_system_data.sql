-- +goose Up
-- +goose StatementBegin

-- Insert system permissions
INSERT IGNORE INTO `permissions` (`id`, `name`, `code`, `resource`, `action`, `scope`, `description`, `is_system`) VALUES
-- Tenant management permissions
('perm-tenant-create', 'Create Tenant', 'tenant.create', 'tenant', 'create', 'system', 'Create new tenant', 1),
('perm-tenant-read', 'Read Tenant', 'tenant.read', 'tenant', 'read', 'system', 'Read tenant information', 1),
('perm-tenant-update', 'Update Tenant', 'tenant.update', 'tenant', 'update', 'system', 'Update tenant information', 1),
('perm-tenant-delete', 'Delete Tenant', 'tenant.delete', 'tenant', 'delete', 'system', 'Delete tenant', 1),
('perm-tenant-list', 'List Tenants', 'tenant.list', 'tenant', 'list', 'system', 'List all tenants', 1),

-- User management permissions
('perm-user-create', 'Create User', 'user.create', 'user', 'create', 'tenant', 'Create new user in tenant', 1),
('perm-user-read', 'Read User', 'user.read', 'user', 'read', 'tenant', 'Read user information', 1),
('perm-user-update', 'Update User', 'user.update', 'user', 'update', 'tenant', 'Update user information', 1),
('perm-user-delete', 'Delete User', 'user.delete', 'user', 'delete', 'tenant', 'Delete user', 1),
('perm-user-list', 'List Users', 'user.list', 'user', 'list', 'tenant', 'List users in tenant', 1),
('perm-user-self-update', 'Update Own Profile', 'user.self.update', 'user', 'update', 'self', 'Update own profile', 1),

-- Role management permissions
('perm-role-create', 'Create Role', 'role.create', 'role', 'create', 'tenant', 'Create new role in tenant', 1),
('perm-role-read', 'Read Role', 'role.read', 'role', 'read', 'tenant', 'Read role information', 1),
('perm-role-update', 'Update Role', 'role.update', 'role', 'update', 'tenant', 'Update role information', 1),
('perm-role-delete', 'Delete Role', 'role.delete', 'role', 'delete', 'tenant', 'Delete role', 1),
('perm-role-list', 'List Roles', 'role.list', 'role', 'list', 'tenant', 'List roles in tenant', 1),
('perm-role-assign', 'Assign Role', 'role.assign', 'role', 'assign', 'tenant', 'Assign role to user', 1),

-- Permission management permissions
('perm-permission-read', 'Read Permission', 'permission.read', 'permission', 'read', 'system', 'Read permission information', 1),
('perm-permission-list', 'List Permissions', 'permission.list', 'permission', 'list', 'system', 'List all permissions', 1),

-- Audit log permissions
('perm-audit-read', 'Read Audit Log', 'audit.read', 'audit', 'read', 'tenant', 'Read audit logs', 1),
('perm-audit-list', 'List Audit Logs', 'audit.list', 'audit', 'list', 'tenant', 'List audit logs', 1);

-- Insert system roles
INSERT IGNORE INTO `roles` (`id`, `tenant_id`, `name`, `code`, `description`, `is_system`) VALUES
-- System administrator role (no tenant_id)
('role-system-admin', NULL, 'System Administrator', 'system-admin', 'System administrator with full access', 1),

-- Default tenant roles
('role-tenant-admin', NULL, 'Tenant Administrator', 'tenant-admin', 'Tenant administrator with full tenant access', 1),
('role-tenant-user', NULL, 'Tenant User', 'tenant-user', 'Regular tenant user with basic access', 1);

-- Assign all system permissions to system admin role
INSERT IGNORE INTO `role_permissions` (`id`, `role_id`, `permission_id`) 
SELECT 
    CONCAT('rp-sysadmin-', p.id) as id,
    'role-system-admin' as role_id,
    p.id as permission_id
FROM `permissions` p 
WHERE p.is_system = 1;

-- Assign tenant-level permissions to tenant admin role
INSERT IGNORE INTO `role_permissions` (`id`, `role_id`, `permission_id`) VALUES
-- Tenant admin can manage users, roles, and view audit logs
('rp-tenantadmin-user-create', 'role-tenant-admin', 'perm-user-create'),
('rp-tenantadmin-user-read', 'role-tenant-admin', 'perm-user-read'),
('rp-tenantadmin-user-update', 'role-tenant-admin', 'perm-user-update'),
('rp-tenantadmin-user-delete', 'role-tenant-admin', 'perm-user-delete'),
('rp-tenantadmin-user-list', 'role-tenant-admin', 'perm-user-list'),
('rp-tenantadmin-role-create', 'role-tenant-admin', 'perm-role-create'),
('rp-tenantadmin-role-read', 'role-tenant-admin', 'perm-role-read'),
('rp-tenantadmin-role-update', 'role-tenant-admin', 'perm-role-update'),
('rp-tenantadmin-role-delete', 'role-tenant-admin', 'perm-role-delete'),
('rp-tenantadmin-role-list', 'role-tenant-admin', 'perm-role-list'),
('rp-tenantadmin-role-assign', 'role-tenant-admin', 'perm-role-assign'),
('rp-tenantadmin-audit-read', 'role-tenant-admin', 'perm-audit-read'),
('rp-tenantadmin-audit-list', 'role-tenant-admin', 'perm-audit-list');

-- Assign basic permissions to tenant user role
INSERT IGNORE INTO `role_permissions` (`id`, `role_id`, `permission_id`) VALUES
('rp-tenantuser-user-self-update', 'role-tenant-user', 'perm-user-self-update'),
('rp-tenantuser-user-read', 'role-tenant-user', 'perm-user-read'),
('rp-tenantuser-role-read', 'role-tenant-user', 'perm-role-read');

-- Create initial system administrator user
INSERT IGNORE INTO `users` (`id`, `tenant_id`, `username`, `email`, `hashed_password`, `first_name`, `last_name`, `status`) VALUES
('user-system-admin', NULL, 'admin', 'admin@system.local', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'System', 'Administrator', 'active');

-- Assign system admin role to the initial system administrator
INSERT IGNORE INTO `user_roles` (`id`, `user_id`, `role_id`) VALUES
('ur-system-admin', 'user-system-admin', 'role-system-admin');

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove seed data
DELETE FROM `user_roles` WHERE `id` = 'ur-system-admin';
DELETE FROM `users` WHERE `id` = 'user-system-admin';
DELETE FROM `role_permissions` WHERE `role_id` IN ('role-system-admin', 'role-tenant-admin', 'role-tenant-user');
DELETE FROM `roles` WHERE `id` IN ('role-system-admin', 'role-tenant-admin', 'role-tenant-user');
DELETE FROM `permissions` WHERE `is_system` = 1;

-- +goose StatementEnd