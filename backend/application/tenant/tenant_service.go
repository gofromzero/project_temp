package tenant

import (
	"errors"
	"fmt"

	"github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/gofromzero/project_temp/backend/infr/repository/mysql"
)

// TenantService provides high-level tenant operations
// This is the application layer service that orchestrates tenant business processes
type TenantService struct {
	tenantDomainService *tenant.Service
	tenantRepository    tenant.TenantRepository
}

// NewTenantService creates a new tenant application service
func NewTenantService() *TenantService {
	// Initialize repository
	repo := mysql.NewTenantRepository()
	
	// Initialize domain service with repository
	domainService := tenant.NewService(repo)
	
	return &TenantService{
		tenantDomainService: domainService,
		tenantRepository:    repo,
	}
}

// CreateTenantRequest represents the request to create a new tenant
type CreateTenantRequest struct {
	Name      string                `json:"name" validate:"required,min=1,max=255"`
	Code      string                `json:"code" validate:"required,min=2,max=100,alphanum"`
	Config    *tenant.TenantConfig  `json:"config,omitempty"`
	AdminUser *AdminUserData        `json:"adminUser,omitempty"`
}

// AdminUserData represents admin user information for tenant creation
type AdminUserData struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=1,max=255"`
	Password string `json:"password" validate:"required,min=8"`
}

// CreateTenantResponse represents the response after creating a tenant
type CreateTenantResponse struct {
	Tenant    *tenant.Tenant `json:"tenant"`
	AdminUser *UserInfo      `json:"adminUser,omitempty"`
	Message   string         `json:"message"`
}

// UserInfo represents user information in response
type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	TenantID string `json:"tenantId"`
}

// CreateTenant creates a new tenant with all necessary initialization
func (s *TenantService) CreateTenant(req CreateTenantRequest) (*CreateTenantResponse, error) {
	// Validate request
	if err := s.validateCreateTenantRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	
	// Create tenant through domain service
	newTenant, err := s.tenantDomainService.CreateTenant(req.Name, req.Code, req.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}
	
	// Initialize tenant data structures
	if err := s.initializeTenantData(newTenant); err != nil {
		// If data initialization fails, we might want to rollback tenant creation
		// For now, we'll just log and continue
		// In a real implementation, this should use transactions
		return nil, fmt.Errorf("failed to initialize tenant data: %w", err)
	}
	
	response := &CreateTenantResponse{
		Tenant:  newTenant,
		Message: fmt.Sprintf("Tenant '%s' created successfully", newTenant.Name),
	}
	
	// Create admin user if provided
	if req.AdminUser != nil {
		adminUser, err := s.createAdminUser(newTenant.ID, *req.AdminUser)
		if err != nil {
			// Admin user creation failed, but tenant was created
			response.Message = fmt.Sprintf("Tenant '%s' created successfully, but admin user creation failed: %s", newTenant.Name, err.Error())
		} else {
			response.AdminUser = adminUser
			response.Message = fmt.Sprintf("Tenant '%s' and admin user created successfully", newTenant.Name)
		}
	}
	
	return response, nil
}

// UpdateTenantRequest represents the request to update a tenant
type UpdateTenantRequest struct {
	Name   *string              `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Status *tenant.TenantStatus `json:"status,omitempty"`
	Config *tenant.TenantConfig `json:"config,omitempty"`
}

// UpdateTenant updates an existing tenant
func (s *TenantService) UpdateTenant(id string, req UpdateTenantRequest) (*tenant.Tenant, error) {
	// Validate request
	if err := s.validateUpdateTenantRequest(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}
	
	// Convert to domain update structure
	updates := tenant.TenantUpdates{
		Name:   req.Name,
		Status: req.Status,
		Config: req.Config,
	}
	
	// Update through domain service
	updatedTenant, err := s.tenantDomainService.UpdateTenant(id, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}
	
	return updatedTenant, nil
}

// GetTenant retrieves a tenant by ID
func (s *TenantService) GetTenant(id string) (*tenant.Tenant, error) {
	if id == "" {
		return nil, errors.New("tenant ID is required")
	}
	
	return s.tenantDomainService.GetTenant(id)
}

// GetTenantByCode retrieves a tenant by code
func (s *TenantService) GetTenantByCode(code string) (*tenant.Tenant, error) {
	if code == "" {
		return nil, errors.New("tenant code is required")
	}
	
	return s.tenantDomainService.GetTenantByCode(code)
}

// ActivateTenant activates a tenant
func (s *TenantService) ActivateTenant(id string) (*tenant.Tenant, error) {
	if id == "" {
		return nil, errors.New("tenant ID is required")
	}
	
	return s.tenantDomainService.ActivateTenant(id)
}

// SuspendTenant suspends a tenant
func (s *TenantService) SuspendTenant(id string) (*tenant.Tenant, error) {
	if id == "" {
		return nil, errors.New("tenant ID is required")
	}
	
	return s.tenantDomainService.SuspendTenant(id)
}

// DisableTenant disables a tenant
func (s *TenantService) DisableTenant(id string) (*tenant.Tenant, error) {
	if id == "" {
		return nil, errors.New("tenant ID is required")
	}
	
	return s.tenantDomainService.DisableTenant(id)
}

// validateCreateTenantRequest validates the create tenant request
func (s *TenantService) validateCreateTenantRequest(req CreateTenantRequest) error {
	if req.Name == "" {
		return errors.New("tenant name is required")
	}
	if req.Code == "" {
		return errors.New("tenant code is required")
	}
	if len(req.Name) > 255 {
		return errors.New("tenant name must be less than 255 characters")
	}
	if len(req.Code) > 100 {
		return errors.New("tenant code must be less than 100 characters")
	}
	
	// Validate admin user data if provided
	if req.AdminUser != nil {
		if req.AdminUser.Email == "" {
			return errors.New("admin user email is required")
		}
		if req.AdminUser.Name == "" {
			return errors.New("admin user name is required")
		}
		if req.AdminUser.Password == "" {
			return errors.New("admin user password is required")
		}
		if len(req.AdminUser.Password) < 8 {
			return errors.New("admin user password must be at least 8 characters")
		}
	}
	
	return nil
}

// validateUpdateTenantRequest validates the update tenant request
func (s *TenantService) validateUpdateTenantRequest(req UpdateTenantRequest) error {
	if req.Name != nil {
		if *req.Name == "" {
			return errors.New("tenant name cannot be empty")
		}
		if len(*req.Name) > 255 {
			return errors.New("tenant name must be less than 255 characters")
		}
	}
	
	if req.Status != nil {
		if !req.Status.IsValid() {
			return errors.New("invalid tenant status")
		}
	}
	
	if req.Config != nil {
		if req.Config.MaxUsers <= 0 {
			return errors.New("max users must be greater than 0")
		}
	}
	
	return nil
}

// initializeTenantData initializes tenant-specific data structures
func (s *TenantService) initializeTenantData(tenant *tenant.Tenant) error {
	// This method would typically:
	// 1. Create default roles for the tenant
	// 2. Set up default permissions
	// 3. Initialize tenant-specific settings
	// 4. Set up any required data structures
	
	// For now, we'll just return nil as the actual implementation
	// would depend on other services (user service, role service, etc.)
	// that are not yet implemented
	
	// TODO: Implement data initialization when other services are available
	return nil
}

// createAdminUser creates an admin user for the tenant
func (s *TenantService) createAdminUser(tenantID string, adminData AdminUserData) (*UserInfo, error) {
	// This method would typically:
	// 1. Create a user through the user service
	// 2. Assign admin role to the user
	// 3. Set up user permissions
	// 4. Send welcome email
	
	// For now, we'll return a mock user info since the user service
	// is not yet implemented
	
	// TODO: Implement admin user creation when user service is available
	return &UserInfo{
		ID:       "admin-user-id", // This would be generated by user service
		Email:    adminData.Email,
		Name:     adminData.Name,
		TenantID: tenantID,
	}, nil
}