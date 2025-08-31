package tenant

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/gogf/gf/v2/util/guid"
)

// TenantStatus represents the possible status values for a tenant
type TenantStatus string

const (
	StatusActive    TenantStatus = "active"
	StatusSuspended TenantStatus = "suspended"
	StatusDisabled  TenantStatus = "disabled"
)

// IsValid checks if the tenant status is valid
func (s TenantStatus) IsValid() bool {
	return s == StatusActive || s == StatusSuspended || s == StatusDisabled
}

// String returns string representation of status
func (s TenantStatus) String() string {
	return string(s)
}

// TenantConfig represents the configuration settings for a tenant
type TenantConfig struct {
	MaxUsers int      `json:"maxUsers"`
	Features []string `json:"features"`
	Theme    *string  `json:"theme,omitempty"`
	Domain   *string  `json:"domain,omitempty"`
}

// Tenant represents the tenant domain entity
type Tenant struct {
	ID          string       `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name        string       `json:"name" gorm:"type:varchar(255);not null;uniqueIndex"`
	Code        string       `json:"code" gorm:"type:varchar(100);not null;uniqueIndex"`
	Status      TenantStatus `json:"status" gorm:"type:enum('active','suspended','disabled');default:'active'"`
	Config      TenantConfig `json:"config" gorm:"type:json"`
	AdminUserID *string      `json:"adminUserId,omitempty" gorm:"type:varchar(36);index"`
	CreatedAt   time.Time    `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt   time.Time    `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName returns the table name for GORM
func (Tenant) TableName() string {
	return "tenants"
}

// NewTenant creates a new tenant with default values
func NewTenant(name, code string) *Tenant {
	return &Tenant{
		ID:     guid.S(),
		Name:   name,
		Code:   code,
		Status: StatusActive,
		Config: TenantConfig{
			MaxUsers: 100, // Default max users
			Features: []string{},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// Validate validates the tenant entity
func (t *Tenant) Validate() error {
	if t.Name == "" {
		return errors.New("tenant name is required")
	}
	if t.Code == "" {
		return errors.New("tenant code is required")
	}
	if !t.Status.IsValid() {
		return errors.New("invalid tenant status")
	}
	if t.Config.MaxUsers <= 0 {
		return errors.New("max users must be greater than 0")
	}
	return nil
}

// IsActive checks if the tenant is active
func (t *Tenant) IsActive() bool {
	return t.Status == StatusActive
}

// IsSuspended checks if the tenant is suspended
func (t *Tenant) IsSuspended() bool {
	return t.Status == StatusSuspended
}

// IsDisabled checks if the tenant is disabled
func (t *Tenant) IsDisabled() bool {
	return t.Status == StatusDisabled
}

// UpdateStatus updates the tenant status
func (t *Tenant) UpdateStatus(status TenantStatus) error {
	if !status.IsValid() {
		return errors.New("invalid tenant status")
	}
	t.Status = status
	t.UpdatedAt = time.Now()
	return nil
}

// UpdateName updates the tenant name
func (t *Tenant) UpdateName(name string) error {
	if name == "" {
		return errors.New("tenant name is required")
	}
	t.Name = name
	t.UpdatedAt = time.Now()
	return nil
}

// SetConfig sets the tenant configuration
func (t *Tenant) SetConfig(config TenantConfig) error {
	if config.MaxUsers <= 0 {
		return errors.New("max users must be greater than 0")
	}
	t.Config = config
	t.UpdatedAt = time.Now()
	return nil
}

// GetConfigJSON returns the configuration as JSON bytes
func (t *Tenant) GetConfigJSON() ([]byte, error) {
	return json.Marshal(t.Config)
}

// CanCreateUsers checks if tenant can create more users
func (t *Tenant) CanCreateUsers(currentUserCount int) bool {
	return t.IsActive() && currentUserCount < t.Config.MaxUsers
}

// TenantRepository defines the repository interface for tenant operations
type TenantRepository interface {
	Create(tenant *Tenant) error
	GetByID(id string) (*Tenant, error)
	GetByCode(code string) (*Tenant, error)
	Update(tenant *Tenant) error
	Delete(id string) error
	List(offset, limit int) ([]*Tenant, error)
	ListWithFilters(filters map[string]interface{}, offset, limit int) ([]*Tenant, int, error)
	Count() (int64, error)
}
