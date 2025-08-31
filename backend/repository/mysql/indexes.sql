-- Database Index Optimization for Multi-Tenant System
-- This file documents the comprehensive indexing strategy for performance optimization

-- TENANTS TABLE INDEXES
-- Primary key: id (already indexed)
-- Unique indexes for business logic constraints
-- uk_tenants_code: code (unique) - for tenant lookup by code
-- uk_tenants_name: name (unique) - for tenant lookup by name
-- Performance indexes
-- idx_tenants_status: status - for filtering active/suspended/disabled tenants
-- idx_tenants_admin_user_id: admin_user_id - for finding tenant by admin user
-- idx_tenants_created_at: created_at - for temporal queries

-- USERS TABLE INDEXES  
-- Primary key: id (already indexed)
-- Multi-tenant isolation indexes (CRITICAL)
-- idx_users_tenant_id: tenant_id - for tenant-based data filtering
-- Unique composite indexes for business constraints
-- uk_users_tenant_username: (tenant_id, username) - unique username per tenant
-- uk_users_tenant_email: (tenant_id, email) - unique email per tenant
-- Performance indexes
-- idx_users_status: status - for filtering active users
-- idx_users_last_login_at: last_login_at - for login analytics
-- idx_users_created_at: created_at - for temporal queries

-- ROLES TABLE INDEXES
-- Primary key: id (already indexed) 
-- Multi-tenant isolation indexes
-- idx_roles_tenant_id: tenant_id - for tenant-based role filtering
-- Unique composite indexes
-- uk_roles_tenant_code: (tenant_id, code) - unique role code per tenant
-- Performance indexes  
-- idx_roles_is_system: is_system - for separating system vs tenant roles
-- idx_roles_created_at: created_at - for temporal queries

-- PERMISSIONS TABLE INDEXES
-- Primary key: id (already indexed)
-- Unique business constraint
-- uk_permissions_code: code - globally unique permission codes
-- Performance indexes for permission lookup
-- idx_permissions_resource: resource - for resource-based permission queries
-- idx_permissions_action: action - for action-based permission queries  
-- idx_permissions_scope: scope - for scope-based filtering (system/tenant/self)
-- idx_permissions_is_system: is_system - for system vs custom permissions

-- USER_ROLES TABLE INDEXES (Junction table)
-- Primary key: id (already indexed)
-- Foreign key performance indexes
-- idx_user_roles_user_id: user_id - for finding roles by user
-- idx_user_roles_role_id: role_id - for finding users by role
-- Unique constraint index
-- uk_user_roles_user_role: (user_id, role_id) - prevent duplicate assignments
-- Temporal index
-- idx_user_roles_created_at: created_at - for assignment history

-- ROLE_PERMISSIONS TABLE INDEXES (Junction table)
-- Primary key: id (already indexed)
-- Foreign key performance indexes
-- idx_role_permissions_role_id: role_id - for finding permissions by role
-- idx_role_permissions_permission_id: permission_id - for finding roles by permission
-- Unique constraint index
-- uk_role_permissions_role_permission: (role_id, permission_id) - prevent duplicates
-- Temporal index
-- idx_role_permissions_created_at: created_at - for assignment history

-- AUDIT_LOGS TABLE INDEXES
-- Primary key: id (already indexed)
-- Multi-tenant isolation (CRITICAL)
-- idx_audit_logs_tenant_id: tenant_id - for tenant-based log filtering
-- Performance indexes for audit queries
-- idx_audit_logs_user_id: user_id - for finding logs by user
-- idx_audit_logs_action: action - for filtering by action type
-- idx_audit_logs_resource: resource - for filtering by resource type
-- idx_audit_logs_resource_id: resource_id - for finding logs for specific resource
-- Temporal performance (CRITICAL for log archival)
-- idx_audit_logs_timestamp: timestamp - for time-based queries
-- Composite index for most common audit query pattern
-- idx_audit_logs_tenant_timestamp: (tenant_id, timestamp) - tenant logs by time

-- PERFORMANCE ANALYSIS AND RECOMMENDATIONS

-- 1. MULTI-TENANT QUERY PATTERNS
-- All tenant-scoped queries MUST include tenant_id in WHERE clause to utilize:
-- - idx_users_tenant_id
-- - idx_roles_tenant_id  
-- - idx_audit_logs_tenant_id
-- - idx_audit_logs_tenant_timestamp (composite)

-- 2. USER AUTHENTICATION PATTERNS
-- Login queries will use composite unique indexes:
-- - uk_users_tenant_username for username-based login
-- - uk_users_tenant_email for email-based login

-- 3. RBAC PERMISSION CHECK PATTERNS
-- Permission resolution queries will efficiently use:
-- - idx_user_roles_user_id -> idx_role_permissions_role_id -> idx_permissions_code
-- - This creates an efficient join path for permission checking

-- 4. AUDIT LOG PERFORMANCE
-- Time-based audit queries will use:
-- - idx_audit_logs_tenant_timestamp for tenant-scoped temporal queries
-- - idx_audit_logs_timestamp for system-wide temporal queries
-- Log archival jobs can efficiently identify old records using timestamp index

-- 5. FOREIGN KEY PERFORMANCE  
-- All foreign key relationships have corresponding indexes to prevent table scans:
-- - tenant_id columns indexed in users, roles, audit_logs
-- - user_id/role_id indexed in junction tables
-- - role_id/permission_id indexed in junction tables

-- 6. COVERING INDEX OPPORTUNITIES
-- Consider adding covering indexes for frequent queries:
-- Users basic info: (tenant_id, status) INCLUDE (id, username, email, first_name, last_name)
-- Role permissions: (role_id) INCLUDE (permission_id, code, resource, action)

-- QUERY PERFORMANCE VALIDATION QUERIES
-- Use these queries to validate index usage:

-- 1. Test tenant user lookup (should use uk_users_tenant_username)
-- EXPLAIN SELECT * FROM users WHERE tenant_id = 'tenant-123' AND username = 'john';

-- 2. Test permission check (should use multiple indexes efficiently) 
-- EXPLAIN SELECT p.* FROM users u
-- JOIN user_roles ur ON u.id = ur.user_id  
-- JOIN role_permissions rp ON ur.role_id = rp.role_id
-- JOIN permissions p ON rp.permission_id = p.id
-- WHERE u.id = 'user-456' AND p.code = 'user.read';

-- 3. Test tenant audit logs (should use idx_audit_logs_tenant_timestamp)
-- EXPLAIN SELECT * FROM audit_logs 
-- WHERE tenant_id = 'tenant-123' 
-- AND timestamp BETWEEN '2023-01-01' AND '2023-12-31'
-- ORDER BY timestamp DESC;

-- 4. Test cross-tenant prevention (should use tenant_id indexes)
-- EXPLAIN SELECT * FROM users WHERE tenant_id = 'tenant-123';
-- EXPLAIN SELECT * FROM roles WHERE tenant_id = 'tenant-123';

-- INDEX MAINTENANCE RECOMMENDATIONS
-- 1. Monitor index usage with INFORMATION_SCHEMA.TABLE_STATISTICS
-- 2. Analyze slow query log for missing indexes
-- 3. Consider partitioning audit_logs by timestamp for very large datasets
-- 4. Review and update table statistics regularly
-- 5. Monitor index fragmentation and rebuild as needed