package database

import (
	"testing"
	"time"

	"github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserEntity(t *testing.T) {
	t.Run("创建用户实体", func(t *testing.T) {
		tenantID := "tenant-123"
		u := &user.User{
			ID:        "user-456",
			TenantID:  &tenantID,
			Username:  "testuser",
			Email:     "test@example.com",
			FirstName: "Test",
			LastName:  "User",
			Status:    user.StatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		assert.Equal(t, "user-456", u.ID)
		assert.Equal(t, "tenant-123", *u.TenantID)
		assert.Equal(t, "testuser", u.Username)
		assert.Equal(t, "test@example.com", u.Email)
		assert.Equal(t, user.StatusActive, u.Status)
		assert.True(t, u.IsActive())
		assert.False(t, u.IsSystemAdmin())
		assert.Equal(t, "users", u.TableName())
	})

	t.Run("系统管理员用户", func(t *testing.T) {
		admin := &user.User{
			ID:        "admin-123",
			TenantID:  nil, // 系统管理员没有租户ID
			Username:  "admin",
			Email:     "admin@system.com",
			FirstName: "System",
			LastName:  "Administrator",
			Status:    user.StatusActive,
		}

		assert.True(t, admin.IsSystemAdmin())
		assert.Nil(t, admin.TenantID)
	})

	t.Run("用户状态验证", func(t *testing.T) {
		activeUser := &user.User{Status: user.StatusActive}
		inactiveUser := &user.User{Status: user.StatusInactive}
		lockedUser := &user.User{Status: user.StatusLocked}

		assert.True(t, activeUser.IsActive())
		assert.False(t, inactiveUser.IsActive())
		assert.False(t, lockedUser.IsActive())
	})

	t.Run("密码哈希和验证", func(t *testing.T) {
		u := &user.User{ID: "user-789"}
		
		password := "testpassword123"
		err := u.SetPassword(password)
		require.NoError(t, err)
		
		assert.NotEmpty(t, u.HashedPassword)
		assert.NotEqual(t, password, u.HashedPassword)
		
		// 验证正确密码
		assert.True(t, u.CheckPassword(password))
		
		// 验证错误密码
		assert.False(t, u.CheckPassword("wrongpassword"))
	})

	t.Run("用户资料获取", func(t *testing.T) {
		avatar := "https://example.com/avatar.jpg"
		phone := "+1234567890"
		
		u := &user.User{
			FirstName: "John",
			LastName:  "Doe",
			Avatar:    &avatar,
			Phone:     &phone,
		}

		profile := u.GetProfile()
		assert.Equal(t, "John", profile.FirstName)
		assert.Equal(t, "Doe", profile.LastName)
		assert.Equal(t, "https://example.com/avatar.jpg", *profile.Avatar)
		assert.Equal(t, "+1234567890", *profile.Phone)
	})

	t.Run("获取用户全名", func(t *testing.T) {
		u := &user.User{
			FirstName: "Jane",
			LastName:  "Smith",
		}

		assert.Equal(t, "Jane Smith", u.GetFullName())
	})

	t.Run("更新最后登录时间", func(t *testing.T) {
		u := &user.User{ID: "user-login-test"}
		
		assert.Nil(t, u.LastLoginAt)
		
		u.UpdateLastLogin()
		
		require.NotNil(t, u.LastLoginAt)
		assert.True(t, time.Since(*u.LastLoginAt) < time.Second)
	})

	t.Run("多租户隔离验证", func(t *testing.T) {
		tenant1ID := "tenant-1"
		tenant2ID := "tenant-2"

		user1 := &user.User{
			ID:       "user-1",
			TenantID: &tenant1ID,
			Username: "user1",
			Email:    "user1@tenant1.com",
		}

		user2 := &user.User{
			ID:       "user-2", 
			TenantID: &tenant2ID,
			Username: "user1", // 同名用户在不同租户下应该被允许
			Email:    "user1@tenant2.com",
		}

		// 验证两个用户属于不同租户
		assert.NotEqual(t, *user1.TenantID, *user2.TenantID)
		assert.Equal(t, user1.Username, user2.Username) // 允许相同用户名在不同租户
	})
}

func TestUserConstants(t *testing.T) {
	t.Run("用户状态常量", func(t *testing.T) {
		assert.Equal(t, user.UserStatus("active"), user.StatusActive)
		assert.Equal(t, user.UserStatus("inactive"), user.StatusInactive)
		assert.Equal(t, user.UserStatus("locked"), user.StatusLocked)
	})
}

func TestUserProfile(t *testing.T) {
	t.Run("用户资料结构", func(t *testing.T) {
		avatar := "avatar.jpg"
		phone := "123-456-7890"
		
		profile := user.UserProfile{
			FirstName: "Alice",
			LastName:  "Johnson",
			Avatar:    &avatar,
			Phone:     &phone,
		}

		assert.Equal(t, "Alice", profile.FirstName)
		assert.Equal(t, "Johnson", profile.LastName)
		assert.Equal(t, "avatar.jpg", *profile.Avatar)
		assert.Equal(t, "123-456-7890", *profile.Phone)
	})

	t.Run("用户资料可选字段", func(t *testing.T) {
		profile := user.UserProfile{
			FirstName: "Bob",
			LastName:  "Wilson",
			// Avatar和Phone为nil
		}

		assert.Equal(t, "Bob", profile.FirstName)
		assert.Equal(t, "Wilson", profile.LastName)
		assert.Nil(t, profile.Avatar)
		assert.Nil(t, profile.Phone)
	})
}