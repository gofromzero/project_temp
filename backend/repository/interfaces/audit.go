package interfaces

import "github.com/gofromzero/project_temp/backend/domain/audit"

// AuditRepository defines the repository interface for audit log operations
type AuditRepository interface {
	audit.AuditRepository
}
