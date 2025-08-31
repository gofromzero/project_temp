package database

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTenantEntity(t *testing.T) {
	t.Run("创建租户实体", func(t *testing.T) {
		config := tenant.TenantConfig{
			MaxUsers: 100,
			Features: []string{"feature1", "feature2"},
			Theme:    stringPtr("dark"),
			Domain:   stringPtr("example.com"),
		}

		tn := &tenant.Tenant{
			ID:        "tenant-123",
			Name:      "Test Tenant",
			Code:      "test-tenant",
			Status:    tenant.StatusActive,
			Config:    config,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.Equal(t, "tenant-123", tn.ID)
		assert.Equal(t, "Test Tenant", tn.Name)
		assert.Equal(t, "test-tenant", tn.Code)
		assert.Equal(t, tenant.StatusActive, tn.Status)
		assert.True(t, tn.IsActive())
		assert.Equal(t, "tenants", tn.TableName())
	})

	t.Run("租户状态枚举验证", func(t *testing.T) {
		activeTenant := &tenant.Tenant{Status: tenant.StatusActive}
		suspendedTenant := &tenant.Tenant{Status: tenant.StatusSuspended}
		disabledTenant := &tenant.Tenant{Status: tenant.StatusDisabled}

		assert.True(t, activeTenant.IsActive())
		assert.False(t, suspendedTenant.IsActive())
		assert.False(t, disabledTenant.IsActive())
	})

	t.Run("租户配置JSON序列化", func(t *testing.T) {
		config := tenant.TenantConfig{
			MaxUsers: 50,
			Features: []string{"auth", "reporting"},
			Theme:    stringPtr("light"),
		}

		tn := &tenant.Tenant{
			ID:     "tenant-456",
			Config: config,
		}

		jsonBytes, err := tn.GetConfigJSON()
		require.NoError(t, err)

		var parsedConfig tenant.TenantConfig
		err = json.Unmarshal(jsonBytes, &parsedConfig)
		require.NoError(t, err)

		assert.Equal(t, 50, parsedConfig.MaxUsers)
		assert.Equal(t, []string{"auth", "reporting"}, parsedConfig.Features)
		assert.Equal(t, "light", *parsedConfig.Theme)
		assert.Nil(t, parsedConfig.Domain)
	})

	t.Run("设置租户配置", func(t *testing.T) {
		tn := &tenant.Tenant{ID: "tenant-789"}

		newConfig := tenant.TenantConfig{
			MaxUsers: 200,
			Features: []string{"premium"},
		}

		tn.SetConfig(newConfig)

		assert.Equal(t, 200, tn.Config.MaxUsers)
		assert.Equal(t, []string{"premium"}, tn.Config.Features)
	})

	t.Run("租户配置结构验证", func(t *testing.T) {
		config := tenant.TenantConfig{
			MaxUsers: 100,
			Features: []string{"feature1", "feature2"},
		}

		assert.Equal(t, 100, config.MaxUsers)
		assert.Len(t, config.Features, 2)
		assert.Nil(t, config.Theme)
		assert.Nil(t, config.Domain)
	})
}

func TestTenantConstants(t *testing.T) {
	t.Run("租户状态常量", func(t *testing.T) {
		assert.Equal(t, tenant.TenantStatus("active"), tenant.StatusActive)
		assert.Equal(t, tenant.TenantStatus("suspended"), tenant.StatusSuspended)
		assert.Equal(t, tenant.TenantStatus("disabled"), tenant.StatusDisabled)
	})
}

// Helper function for creating string pointers
func stringPtr(s string) *string {
	return &s
}
