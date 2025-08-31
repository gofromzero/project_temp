package auth

import (
	"github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of user.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *user.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id string) (*user.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByUsername(tenantID *string, username string) (*user.User, error) {
	args := m.Called(tenantID, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(tenantID *string, email string) (*user.User, error) {
	args := m.Called(tenantID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockUserRepository) GetByTenantID(tenantID string, offset, limit int) ([]*user.User, error) {
	args := m.Called(tenantID, offset, limit)
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *MockUserRepository) GetSystemAdmins(offset, limit int) ([]*user.User, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]*user.User), args.Error(1)
}

func (m *MockUserRepository) Update(user *user.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) Count(tenantID *string) (int64, error) {
	args := m.Called(tenantID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserRepository) ExistsByUsername(tenantID *string, username string) (bool, error) {
	args := m.Called(tenantID, username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmail(tenantID *string, email string) (bool, error) {
	args := m.Called(tenantID, email)
	return args.Bool(0), args.Error(1)
}

// MockTenantRepository is a mock implementation of tenant.TenantRepository
type MockTenantRepository struct {
	mock.Mock
}

func (m *MockTenantRepository) Create(tenant *tenant.Tenant) error {
	args := m.Called(tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) GetByID(id string) (*tenant.Tenant, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tenant.Tenant), args.Error(1)
}

func (m *MockTenantRepository) GetByCode(code string) (*tenant.Tenant, error) {
	args := m.Called(code)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tenant.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Update(tenant *tenant.Tenant) error {
	args := m.Called(tenant)
	return args.Error(0)
}

func (m *MockTenantRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTenantRepository) List(offset, limit int) ([]*tenant.Tenant, error) {
	args := m.Called(offset, limit)
	return args.Get(0).([]*tenant.Tenant), args.Error(1)
}

func (m *MockTenantRepository) Count() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTenantRepository) ListWithFilters(filters map[string]interface{}, offset, limit int) ([]*tenant.Tenant, int, error) {
	args := m.Called(filters, offset, limit)
	return args.Get(0).([]*tenant.Tenant), args.Get(1).(int), args.Error(2)
}
