package tenant

import (
	"errors"
	"fmt"
)

// Service provides tenant domain business logic
type Service struct {
	repo TenantRepository
}

// NewService creates a new tenant domain service
func NewService(repo TenantRepository) *Service {
	return &Service{
		repo: repo,
	}
}

// CreateTenant creates a new tenant with business validation
func (s *Service) CreateTenant(name, code string, config *TenantConfig) (*Tenant, error) {
	// Check if tenant code is unique
	existing, err := s.repo.GetByCode(code)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("tenant with code '%s' already exists", code)
	}

	// Create new tenant
	tenant := NewTenant(name, code)
	
	// Set custom config if provided
	if config != nil {
		if err := tenant.SetConfig(*config); err != nil {
			return nil, fmt.Errorf("invalid tenant config: %w", err)
		}
	}

	// Validate tenant
	if err := tenant.Validate(); err != nil {
		return nil, fmt.Errorf("tenant validation failed: %w", err)
	}

	// Save to repository
	if err := s.repo.Create(tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenant, nil
}

// UpdateTenant updates an existing tenant
func (s *Service) UpdateTenant(id string, updates TenantUpdates) (*Tenant, error) {
	// Get existing tenant
	tenant, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Apply updates
	if updates.Name != nil {
		if err := tenant.UpdateName(*updates.Name); err != nil {
			return nil, fmt.Errorf("failed to update name: %w", err)
		}
	}

	if updates.Status != nil {
		if err := tenant.UpdateStatus(*updates.Status); err != nil {
			return nil, fmt.Errorf("failed to update status: %w", err)
		}
	}

	if updates.Config != nil {
		if err := tenant.SetConfig(*updates.Config); err != nil {
			return nil, fmt.Errorf("failed to update config: %w", err)
		}
	}

	// Validate updated tenant
	if err := tenant.Validate(); err != nil {
		return nil, fmt.Errorf("tenant validation failed: %w", err)
	}

	// Save changes
	if err := s.repo.Update(tenant); err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (s *Service) GetTenant(id string) (*Tenant, error) {
	tenant, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	return tenant, nil
}

// GetTenantByCode retrieves a tenant by code
func (s *Service) GetTenantByCode(code string) (*Tenant, error) {
	tenant, err := s.repo.GetByCode(code)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}
	return tenant, nil
}

// ActivateTenant activates a tenant
func (s *Service) ActivateTenant(id string) (*Tenant, error) {
	status := StatusActive
	return s.UpdateTenant(id, TenantUpdates{
		Status: &status,
	})
}

// SuspendTenant suspends a tenant
func (s *Service) SuspendTenant(id string) (*Tenant, error) {
	status := StatusSuspended
	return s.UpdateTenant(id, TenantUpdates{
		Status: &status,
	})
}

// DisableTenant disables a tenant
func (s *Service) DisableTenant(id string) (*Tenant, error) {
	status := StatusDisabled
	return s.UpdateTenant(id, TenantUpdates{
		Status: &status,
	})
}

// ValidateConfig validates tenant configuration
func (s *Service) ValidateConfig(config TenantConfig) error {
	if config.MaxUsers <= 0 {
		return errors.New("max users must be greater than 0")
	}
	
	// Validate domain format if provided
	if config.Domain != nil && *config.Domain != "" {
		// Basic domain validation (can be enhanced)
		domain := *config.Domain
		if len(domain) < 3 || len(domain) > 253 {
			return errors.New("domain must be between 3 and 253 characters")
		}
	}
	
	return nil
}

// CanTenantCreateUsers checks if tenant can create more users
func (s *Service) CanTenantCreateUsers(tenantID string, currentUserCount int) (bool, error) {
	tenant, err := s.repo.GetByID(tenantID)
	if err != nil {
		return false, fmt.Errorf("tenant not found: %w", err)
	}
	
	return tenant.CanCreateUsers(currentUserCount), nil
}

// TenantUpdates represents fields that can be updated for a tenant
type TenantUpdates struct {
	Name   *string       `json:"name,omitempty"`
	Status *TenantStatus `json:"status,omitempty"`
	Config *TenantConfig `json:"config,omitempty"`
}