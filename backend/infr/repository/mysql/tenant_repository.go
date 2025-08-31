package mysql

import (
	"context"

	"github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/gofromzero/project_temp/backend/infr/database"
)

// TenantRepository implements the tenant repository interface using MySQL
type TenantRepository struct {
	connection *database.Connection
}

// NewTenantRepository creates a new MySQL tenant repository
func NewTenantRepository() tenant.TenantRepository {
	return &TenantRepository{
		connection: database.NewConnection(),
	}
}

// Create creates a new tenant
func (r *TenantRepository) Create(tenant *tenant.Tenant) error {
	ctx := context.Background()
	_, err := r.connection.InsertWithTenantFilter(ctx, "tenants", tenant)
	return err
}

// GetByID retrieves a tenant by ID
func (r *TenantRepository) GetByID(id string) (*tenant.Tenant, error) {
	ctx := context.Background()
	ctx = r.connection.WithTenantContext(ctx, nil, true) // System admin context for tenant access
	
	model, err := r.connection.GetTenantAwareModel(ctx, "tenants")
	if err != nil {
		return nil, err
	}

	var result tenant.Tenant
	err = model.Where("id", id).Scan(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetByCode retrieves a tenant by code
func (r *TenantRepository) GetByCode(code string) (*tenant.Tenant, error) {
	ctx := context.Background()
	ctx = r.connection.WithTenantContext(ctx, nil, true) // System admin context for tenant access
	
	model, err := r.connection.GetTenantAwareModel(ctx, "tenants")
	if err != nil {
		return nil, err
	}

	var result tenant.Tenant
	err = model.Where("code", code).Scan(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Update updates a tenant
func (r *TenantRepository) Update(tenant *tenant.Tenant) error {
	ctx := context.Background()
	ctx = r.connection.WithTenantContext(ctx, nil, true) // System admin context for tenant updates
	
	model, err := r.connection.GetTenantAwareModel(ctx, "tenants")
	if err != nil {
		return err
	}

	_, err = model.Where("id", tenant.ID).Update(tenant)
	return err
}

// Delete deletes a tenant by ID
func (r *TenantRepository) Delete(id string) error {
	ctx := context.Background()
	ctx = r.connection.WithTenantContext(ctx, nil, true) // System admin context
	
	model, err := r.connection.GetTenantAwareModel(ctx, "tenants")
	if err != nil {
		return err
	}

	_, err = model.Where("id", id).Delete()
	return err
}

// List retrieves tenants with pagination
func (r *TenantRepository) List(offset, limit int) ([]*tenant.Tenant, error) {
	ctx := context.Background()
	ctx = r.connection.WithTenantContext(ctx, nil, true) // System admin context
	
	model, err := r.connection.GetTenantAwareModel(ctx, "tenants")
	if err != nil {
		return nil, err
	}

	var results []*tenant.Tenant
	err = model.Offset(offset).Limit(limit).Scan(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// Count returns the total count of tenants
func (r *TenantRepository) Count() (int64, error) {
	ctx := context.Background()
	ctx = r.connection.WithTenantContext(ctx, nil, true) // System admin context
	
	model, err := r.connection.GetTenantAwareModel(ctx, "tenants")
	if err != nil {
		return 0, err
	}

	count, err := model.Count()
	return int64(count), err
}