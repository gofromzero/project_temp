package audit

import "time"

// AuditLog represents the audit log domain entity
type AuditLog struct {
	ID       uint64    `json:"id"`
	TenantID uint64    `json:"tenant_id"`
	UserID   uint64    `json:"user_id"`
	Action   string    `json:"action"`
	Resource string    `json:"resource"`
	Details  string    `json:"details"`
	Created  time.Time `json:"created"`
}

// AuditRepository defines the repository interface for audit operations
type AuditRepository interface {
	Create(log *AuditLog) error
	GetByTenantID(tenantID uint64) ([]*AuditLog, error)
	GetByUserID(userID uint64) ([]*AuditLog, error)
}
