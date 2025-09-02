package user

import (
	"testing"
	"time"

	"github.com/gofromzero/project_temp/backend/domain/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUser_ValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
		errCode  string
	}{
		{
			name:     "Valid username",
			username: "validuser123",
			wantErr:  false,
		},
		{
			name:     "Empty username",
			username: "",
			wantErr:  true,
			errCode:  "REQUIRED",
		},
		{
			name:     "Username too short",
			username: "ab",
			wantErr:  true,
			errCode:  "MIN_LENGTH",
		},
		{
			name:     "Username too long",
			username: "this_is_a_very_long_username_that_exceeds_fifty_chars_limit",
			wantErr:  true,
			errCode:  "MAX_LENGTH",
		},
		{
			name:     "Username with invalid characters",
			username: "user@name",
			wantErr:  true,
			errCode:  "INVALID_FORMAT",
		},
		{
			name:     "Reserved username",
			username: "admin",
			wantErr:  true,
			errCode:  "RESERVED_NAME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &user.User{Username: tt.username}
			err := u.ValidateUsername()

			if tt.wantErr {
				require.Error(t, err)
				var validationErr *user.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Equal(t, tt.errCode, validationErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_ValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
		errCode string
	}{
		{
			name:    "Valid email",
			email:   "user@example.com",
			wantErr: false,
		},
		{
			name:    "Empty email",
			email:   "",
			wantErr: true,
			errCode: "REQUIRED",
		},
		{
			name:    "Invalid email format",
			email:   "invalid-email",
			wantErr: true,
			errCode: "INVALID_FORMAT",
		},
		{
			name:    "Valid email lowercase",
			email:   "user@example.com",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &user.User{Email: tt.email}
			err := u.ValidateEmail()

			if tt.wantErr {
				require.Error(t, err)
				var validationErr *user.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Equal(t, tt.errCode, validationErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
		errCode  string
	}{
		{
			name:     "Valid password",
			password: "validpass123",
			wantErr:  false,
		},
		{
			name:     "Empty password",
			password: "",
			wantErr:  true,
			errCode:  "REQUIRED",
		},
		{
			name:     "Password too short",
			password: "short1",
			wantErr:  true,
			errCode:  "MIN_LENGTH",
		},
		{
			name:     "Password without letter",
			password: "12345678",
			wantErr:  true,
			errCode:  "WEAK_PASSWORD",
		},
		{
			name:     "Password without number",
			password: "onlyletters",
			wantErr:  true,
			errCode:  "WEAK_PASSWORD",
		},
		{
			name:     "Common weak password",
			password: "admin123",
			wantErr:  true,
			errCode:  "COMMON_PASSWORD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := user.ValidatePassword(tt.password)

			if tt.wantErr {
				require.Error(t, err)
				var validationErr *user.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Equal(t, tt.errCode, validationErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_ValidateProfile(t *testing.T) {
	tests := []struct {
		name      string
		user      *user.User
		wantErr   bool
		errCode   string
		errField  string
	}{
		{
			name: "Valid profile",
			user: &user.User{
				FirstName: "John",
				LastName:  "Doe",
			},
			wantErr: false,
		},
		{
			name: "Empty first name",
			user: &user.User{
				FirstName: "",
				LastName:  "Doe",
			},
			wantErr:  true,
			errCode:  "REQUIRED",
			errField: "firstName",
		},
		{
			name: "Empty last name",
			user: &user.User{
				FirstName: "John",
				LastName:  "",
			},
			wantErr:  true,
			errCode:  "REQUIRED",
			errField: "lastName",
		},
		{
			name: "Invalid phone format",
			user: &user.User{
				FirstName: "John",
				LastName:  "Doe",
				Phone:     stringPtr("invalid-phone"),
			},
			wantErr:  true,
			errCode:  "INVALID_FORMAT",
			errField: "phone",
		},
		{
			name: "Invalid avatar URL",
			user: &user.User{
				FirstName: "John",
				LastName:  "Doe",
				Avatar:    stringPtr("invalid-url"),
			},
			wantErr:  true,
			errCode:  "INVALID_FORMAT",
			errField: "avatar",
		},
		{
			name: "Valid profile with phone and avatar",
			user: &user.User{
				FirstName: "John",
				LastName:  "Doe",
				Phone:     stringPtr("+1234567890"),
				Avatar:    stringPtr("https://example.com/avatar.jpg"),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.ValidateProfile()

			if tt.wantErr {
				require.Error(t, err)
				var validationErr *user.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Equal(t, tt.errCode, validationErr.Code)
				assert.Equal(t, tt.errField, validationErr.Field)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_ValidateStatusTransition(t *testing.T) {
	tests := []struct {
		name      string
		fromStatus user.UserStatus
		toStatus   user.UserStatus
		wantErr    bool
		errCode    string
	}{
		{
			name:       "Active to Inactive - valid",
			fromStatus: user.StatusActive,
			toStatus:   user.StatusInactive,
			wantErr:    false,
		},
		{
			name:       "Active to Locked - valid",
			fromStatus: user.StatusActive,
			toStatus:   user.StatusLocked,
			wantErr:    false,
		},
		{
			name:       "Locked to Active - valid",
			fromStatus: user.StatusLocked,
			toStatus:   user.StatusActive,
			wantErr:    false,
		},
		{
			name:       "Same status - valid",
			fromStatus: user.StatusActive,
			toStatus:   user.StatusActive,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := &user.User{Status: tt.fromStatus}
			err := u.ValidateStatusTransition(tt.toStatus)

			if tt.wantErr {
				require.Error(t, err)
				var validationErr *user.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Equal(t, tt.errCode, validationErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_CanBeDeleted(t *testing.T) {
	tests := []struct {
		name    string
		user    *user.User
		wantErr bool
		errCode string
	}{
		{
			name: "Regular user can be deleted",
			user: &user.User{
				TenantID: stringPtr("tenant-1"),
				Status:   user.StatusActive,
			},
			wantErr: false,
		},
		{
			name: "System admin cannot be deleted",
			user: &user.User{
				TenantID: nil,
				Status:   user.StatusActive,
			},
			wantErr: true,
			errCode: "SYSTEM_ADMIN_UNDELETABLE",
		},
		{
			name: "Locked user cannot be deleted",
			user: &user.User{
				TenantID: stringPtr("tenant-1"),
				Status:   user.StatusLocked,
			},
			wantErr: true,
			errCode: "LOCKED_USER_UNDELETABLE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.CanBeDeleted()

			if tt.wantErr {
				require.Error(t, err)
				var validationErr *user.ValidationError
				require.ErrorAs(t, err, &validationErr)
				assert.Equal(t, tt.errCode, validationErr.Code)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUser_SetPasswordWithValidation(t *testing.T) {
	u := &user.User{}

	// Test valid password
	err := u.SetPasswordWithValidation("validpass123")
	assert.NoError(t, err)
	assert.NotEmpty(t, u.HashedPassword)

	// Test password verification
	assert.True(t, u.CheckPassword("validpass123"))
	assert.False(t, u.CheckPassword("wrongpassword"))

	// Test invalid password
	err = u.SetPasswordWithValidation("weak")
	assert.Error(t, err)
}

func TestUser_StatusTransitionMethods(t *testing.T) {
	u := &user.User{Status: user.StatusActive, UpdatedAt: time.Now().Add(-time.Hour)}
	originalTime := u.UpdatedAt

	// Test Lock
	err := u.Lock()
	assert.NoError(t, err)
	assert.Equal(t, user.StatusLocked, u.Status)
	assert.True(t, u.UpdatedAt.After(originalTime))

	// Test Unlock
	err = u.Unlock()
	assert.NoError(t, err)
	assert.Equal(t, user.StatusActive, u.Status)

	// Test Deactivate
	err = u.Deactivate()
	assert.NoError(t, err)
	assert.Equal(t, user.StatusInactive, u.Status)

	// Test Activate
	err = u.Activate()
	assert.NoError(t, err)
	assert.Equal(t, user.StatusActive, u.Status)
}

func TestUser_HelperMethods(t *testing.T) {
	u := &user.User{
		FirstName: "John",
		LastName:  "Doe",
		Status:    user.StatusActive,
		TenantID:  stringPtr("tenant-1"),
	}

	assert.Equal(t, "John Doe", u.GetFullName())
	assert.True(t, u.IsActive())
	assert.False(t, u.IsSystemAdmin())

	u.UpdateLastLogin()
	assert.NotNil(t, u.LastLoginAt)
	assert.True(t, time.Since(*u.LastLoginAt) < time.Second)

	profile := u.GetProfile()
	assert.Equal(t, "John", profile.FirstName)
	assert.Equal(t, "Doe", profile.LastName)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}