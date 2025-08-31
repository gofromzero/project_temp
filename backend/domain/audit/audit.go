package audit

import (
	"encoding/json"
	"time"
)

// AuditLogDetails represents the details structure for audit logs
type AuditLogDetails struct {
	Before   interface{}            `json:"before,omitempty"`
	After    interface{}            `json:"after,omitempty"`
	Metadata map[string]interface{} `json:"metadata"`
}

// AuditLog represents the audit log domain entity
type AuditLog struct {
	ID         string           `json:"id" gorm:"type:varchar(36);primaryKey"`
	TenantID   *string          `json:"tenantId,omitempty" gorm:"type:varchar(36);index"`
	UserID     string           `json:"userId" gorm:"type:varchar(36);not null;index"`
	Action     string           `json:"action" gorm:"type:varchar(50);not null;index;comment:create,update,delete,login,logout"`
	Resource   string           `json:"resource" gorm:"type:varchar(50);not null;index;comment:user,tenant,role,permission"`
	ResourceID *string          `json:"resourceId,omitempty" gorm:"type:varchar(36);index"`
	Details    AuditLogDetails  `json:"details" gorm:"type:json"`
	IPAddress  *string          `json:"ipAddress,omitempty" gorm:"type:varchar(45)"`
	UserAgent  *string          `json:"userAgent,omitempty" gorm:"type:varchar(500)"`
	Timestamp  time.Time        `json:"timestamp" gorm:"autoCreateTime;index"`
}

// TableName returns the table name for GORM
func (AuditLog) TableName() string {
	return "audit_logs"
}

// SetDetails sets the audit log details
func (a *AuditLog) SetDetails(before, after interface{}, metadata map[string]interface{}) {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	a.Details = AuditLogDetails{
		Before:   before,
		After:    after,
		Metadata: metadata,
	}
}

// GetDetailsJSON returns the details as JSON bytes
func (a *AuditLog) GetDetailsJSON() ([]byte, error) {
	return json.Marshal(a.Details)
}

// IsSystemAction checks if the audit log is for a system-level action
func (a *AuditLog) IsSystemAction() bool {
	return a.TenantID == nil
}

// AuditRepository defines the repository interface for audit operations
type AuditRepository interface {
	Create(log *AuditLog) error
	GetByID(id string) (*AuditLog, error)
	GetByTenantID(tenantID string, offset, limit int) ([]*AuditLog, error)
	GetByUserID(userID string, offset, limit int) ([]*AuditLog, error)
	GetByResource(resource string, resourceID *string, offset, limit int) ([]*AuditLog, error)
	GetByAction(action string, offset, limit int) ([]*AuditLog, error)
	GetByTimeRange(tenantID *string, startTime, endTime time.Time, offset, limit int) ([]*AuditLog, error)
	Count(tenantID *string) (int64, error)
	Delete(id string) error
	DeleteOldEntries(before time.Time) (int64, error)
}
