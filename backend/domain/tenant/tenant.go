package tenant

// Tenant represents the tenant domain entity
type Tenant struct {
	ID     uint64 `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// TenantRepository defines the repository interface for tenant operations
type TenantRepository interface {
	Create(tenant *Tenant) error
	GetByID(id uint64) (*Tenant, error)
	Update(tenant *Tenant) error
	Delete(id uint64) error
}
