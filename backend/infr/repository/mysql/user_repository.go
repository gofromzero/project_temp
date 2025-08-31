package mysql

import (
	"context"

	"github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/gofromzero/project_temp/backend/infr/database"
)

// UserRepository implements the user repository interface using MySQL
type UserRepository struct {
	connection *database.Connection
}

// NewUserRepository creates a new MySQL user repository
func NewUserRepository() user.UserRepository {
	return &UserRepository{
		connection: database.NewConnection(),
	}
}

// Create creates a new user
func (r *UserRepository) Create(user *user.User) error {
	ctx := context.Background()
	_, err := r.connection.InsertWithTenantFilter(ctx, "users", user)
	return err
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id string) (*user.User, error) {
	ctx := context.Background()
	model, err := r.connection.GetTenantAwareModel(ctx, "users")
	if err != nil {
		return nil, err
	}

	var result user.User
	err = model.Where("id", id).Scan(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetByUsername retrieves a user by username within a tenant
func (r *UserRepository) GetByUsername(tenantID *string, username string) (*user.User, error) {
	ctx := context.Background()

	// Create tenant-aware context
	if tenantID != nil {
		ctx = r.connection.WithTenantContext(ctx, tenantID, false)
	} else {
		ctx = r.connection.WithTenantContext(ctx, nil, true) // System admin
	}

	model, err := r.connection.GetTenantAwareModel(ctx, "users")
	if err != nil {
		return nil, err
	}

	query := model.Where("username", username)
	if tenantID != nil {
		query = query.Where("tenant_id", *tenantID)
	} else {
		query = query.WhereNull("tenant_id")
	}

	var result user.User
	err = query.Scan(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetByEmail retrieves a user by email within a tenant
func (r *UserRepository) GetByEmail(tenantID *string, email string) (*user.User, error) {
	ctx := context.Background()

	// Create tenant-aware context
	if tenantID != nil {
		ctx = r.connection.WithTenantContext(ctx, tenantID, false)
	} else {
		ctx = r.connection.WithTenantContext(ctx, nil, true) // System admin
	}

	model, err := r.connection.GetTenantAwareModel(ctx, "users")
	if err != nil {
		return nil, err
	}

	query := model.Where("email", email)
	if tenantID != nil {
		query = query.Where("tenant_id", *tenantID)
	} else {
		query = query.WhereNull("tenant_id")
	}

	var result user.User
	err = query.Scan(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// GetByTenantID retrieves users by tenant ID with pagination
func (r *UserRepository) GetByTenantID(tenantID string, offset, limit int) ([]*user.User, error) {
	ctx := context.Background()
	ctx = r.connection.WithTenantContext(ctx, &tenantID, false)

	model, err := r.connection.GetTenantAwareModel(ctx, "users")
	if err != nil {
		return nil, err
	}

	var results []*user.User
	err = model.Where("tenant_id", tenantID).
		Offset(offset).
		Limit(limit).
		Scan(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// GetSystemAdmins retrieves system administrators with pagination
func (r *UserRepository) GetSystemAdmins(offset, limit int) ([]*user.User, error) {
	ctx := context.Background()
	ctx = r.connection.WithTenantContext(ctx, nil, true)

	model, err := r.connection.GetTenantAwareModel(ctx, "users")
	if err != nil {
		return nil, err
	}

	var results []*user.User
	err = model.WhereNull("tenant_id").
		Offset(offset).
		Limit(limit).
		Scan(&results)
	if err != nil {
		return nil, err
	}
	return results, nil
}

// Update updates a user
func (r *UserRepository) Update(user *user.User) error {
	ctx := context.Background()

	// Create tenant-aware context based on user's tenant
	if user.TenantID != nil {
		ctx = r.connection.WithTenantContext(ctx, user.TenantID, false)
	} else {
		ctx = r.connection.WithTenantContext(ctx, nil, true)
	}

	model, err := r.connection.GetTenantAwareModel(ctx, "users")
	if err != nil {
		return err
	}

	_, err = model.Where("id", user.ID).Update(user)
	return err
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(id string) error {
	ctx := context.Background()
	model, err := r.connection.GetTenantAwareModel(ctx, "users")
	if err != nil {
		return err
	}

	_, err = model.Where("id", id).Delete()
	return err
}

// Count returns the count of users for a tenant
func (r *UserRepository) Count(tenantID *string) (int64, error) {
	ctx := context.Background()

	if tenantID != nil {
		ctx = r.connection.WithTenantContext(ctx, tenantID, false)
	} else {
		ctx = r.connection.WithTenantContext(ctx, nil, true)
	}

	model, err := r.connection.GetTenantAwareModel(ctx, "users")
	if err != nil {
		return 0, err
	}

	query := model
	if tenantID != nil {
		query = query.Where("tenant_id", *tenantID)
	} else {
		query = query.WhereNull("tenant_id")
	}

	count, err := query.Count()
	return int64(count), err
}

// ExistsByUsername checks if a user exists by username within a tenant
func (r *UserRepository) ExistsByUsername(tenantID *string, username string) (bool, error) {
	ctx := context.Background()

	if tenantID != nil {
		ctx = r.connection.WithTenantContext(ctx, tenantID, false)
	} else {
		ctx = r.connection.WithTenantContext(ctx, nil, true)
	}

	model, err := r.connection.GetTenantAwareModel(ctx, "users")
	if err != nil {
		return false, err
	}

	query := model.Where("username", username)
	if tenantID != nil {
		query = query.Where("tenant_id", *tenantID)
	} else {
		query = query.WhereNull("tenant_id")
	}

	count, err := query.Count()
	return count > 0, err
}

// ExistsByEmail checks if a user exists by email within a tenant
func (r *UserRepository) ExistsByEmail(tenantID *string, email string) (bool, error) {
	ctx := context.Background()

	if tenantID != nil {
		ctx = r.connection.WithTenantContext(ctx, tenantID, false)
	} else {
		ctx = r.connection.WithTenantContext(ctx, nil, true)
	}

	model, err := r.connection.GetTenantAwareModel(ctx, "users")
	if err != nil {
		return false, err
	}

	query := model.Where("email", email)
	if tenantID != nil {
		query = query.Where("tenant_id", *tenantID)
	} else {
		query = query.WhereNull("tenant_id")
	}

	count, err := query.Count()
	return count > 0, err
}
