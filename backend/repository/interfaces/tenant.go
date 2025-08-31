package interfaces

import "github.com/gofromzero/project_temp/backend/domain/tenant"

// TenantRepository defines the repository interface for tenant operations
// This interface is implemented by the MySQL repository layer
type TenantRepository interface {
	tenant.TenantRepository
}