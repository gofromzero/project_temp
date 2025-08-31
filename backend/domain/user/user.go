package user

// User represents the user domain entity
type User struct {
	ID       uint64 `json:"id"`
	TenantID uint64 `json:"tenant_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

// UserRepository defines the repository interface for user operations
type UserRepository interface {
	Create(user *User) error
	GetByID(id uint64) (*User, error)
	GetByTenantID(tenantID uint64) ([]*User, error)
	Update(user *User) error
	Delete(id uint64) error
}
