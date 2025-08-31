package documentation_test

import (
	"testing"

	"github.com/gofromzero/project_temp/backend/api/handlers"
	"github.com/gofromzero/project_temp/backend/application/tenant"
	domainTenant "github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/gofromzero/project_temp/backend/pkg/response"
	"github.com/gogf/gf/v2/test/gtest"
)

// TestAPIDocumentationAccuracy verifies that the API documentation matches the implementation
func TestAPIDocumentationAccuracy(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// Test that all documented request/response structures exist and match documentation

		// Test handler methods exist
		handler := handlers.NewTenantHandler()
		t.Assert(handler.ListTenants != nil, true)
		t.Assert(handler.GetTenant != nil, true)
		t.Assert(handler.CreateTenant != nil, true)
		t.Assert(handler.UpdateTenant != nil, true)
		t.Assert(handler.DeleteTenant != nil, true)
		t.Assert(handler.ActivateTenant != nil, true)
		t.Assert(handler.SuspendTenant != nil, true)
		t.Assert(handler.DisableTenant != nil, true)
	})
}

func TestDocumentedRequestStructures(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// Test that all documented request structures match implementation

		// 1. CreateTenantRequest structure matches documentation
		createReq := tenant.CreateTenantRequest{
			Name: "Test Tenant",
			Code: "TEST001",
			Config: &domainTenant.TenantConfig{
				MaxUsers: 100,
				Features: []string{"basic", "analytics"},
				Theme:    stringPtr("default"),
				Domain:   stringPtr("test.example.com"),
			},
			AdminUser: &tenant.AdminUserData{
				Email:    "admin@test.com",
				Name:     "Admin User",
				Password: "password123",
			},
		}

		// Verify fields exist and have correct types
		t.Assert(createReq.Name, "Test Tenant")
		t.Assert(createReq.Code, "TEST001")
		t.Assert(createReq.Config != nil, true)
		t.Assert(createReq.AdminUser != nil, true)

		// 2. UpdateTenantRequest structure matches documentation
		updatedName := "Updated Name"
		status := domainTenant.StatusActive
		updateReq := tenant.UpdateTenantRequest{
			Name:   &updatedName,
			Status: &status,
			Config: &domainTenant.TenantConfig{
				MaxUsers: 200,
				Features: []string{"basic", "premium"},
			},
		}

		t.Assert(updateReq.Name != nil, true)
		t.Assert(*updateReq.Name, "Updated Name")
		t.Assert(updateReq.Status != nil, true)
		t.Assert(*updateReq.Status, domainTenant.StatusActive)

		// 3. ListTenantsRequest structure matches documentation
		listReq := tenant.ListTenantsRequest{
			Page:   1,
			Limit:  10,
			Name:   "search",
			Code:   "CODE",
			Status: "active",
		}

		t.Assert(listReq.Page, 1)
		t.Assert(listReq.Limit, 10)
		t.Assert(listReq.Name, "search")
		t.Assert(listReq.Code, "CODE")
		t.Assert(listReq.Status, "active")

		// 4. DeleteTenantRequest structure matches documentation
		deleteReq := handlers.DeleteTenantRequest{
			Confirmation: "DELETE_TENANT_12345",
			Reason:       "user_request",
			CreateBackup: true,
		}

		t.Assert(deleteReq.Confirmation, "DELETE_TENANT_12345")
		t.Assert(deleteReq.Reason, "user_request")
		t.Assert(deleteReq.CreateBackup, true)
	})
}

func TestDocumentedResponseStructures(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// Test that all documented response structures match implementation

		// 1. Standard success response structure
		successResp := response.SuccessResponse{
			Success:   true,
			Message:   "Operation successful",
			Data:      map[string]string{"key": "value"},
			RequestID: "req-12345",
		}

		t.Assert(successResp.Success, true)
		t.Assert(successResp.Message, "Operation successful")
		t.Assert(successResp.Data != nil, true)
		t.Assert(successResp.RequestID, "req-12345")

		// 2. Standard error response structure
		errorResp := response.ErrorResponse{
			Success: false,
			Error: response.ErrorDetail{
				Code:    response.ErrCodeValidationFailed,
				Message: "Validation failed",
				Field:   "name",
				Details: map[string]string{"reason": "required"},
			},
			RequestID: "req-12346",
			Path:      "/v1/tenants",
		}

		t.Assert(errorResp.Success, false)
		t.Assert(errorResp.Error.Code, response.ErrCodeValidationFailed)
		t.Assert(errorResp.Error.Message, "Validation failed")
		t.Assert(errorResp.Error.Field, "name")
		t.Assert(errorResp.RequestID, "req-12346")
		t.Assert(errorResp.Path, "/v1/tenants")

		// 3. Pagination response structure
		pagination := tenant.Pagination{
			Page:  1,
			Limit: 10,
			Total: 50,
			Pages: 5,
		}

		t.Assert(pagination.Page, 1)
		t.Assert(pagination.Limit, 10)
		t.Assert(pagination.Total, 50)
		t.Assert(pagination.Pages, 5)

		// 4. ListTenantsResponse structure
		listResp := tenant.ListTenantsResponse{
			Tenants: []*domainTenant.Tenant{
				{
					ID:     "tenant-1",
					Name:   "Test Tenant",
					Code:   "TEST001",
					Status: domainTenant.StatusActive,
				},
			},
			Pagination: pagination,
		}

		t.Assert(len(listResp.Tenants), 1)
		t.Assert(listResp.Tenants[0].ID, "tenant-1")
		t.Assert(listResp.Pagination.Total, 50)
	})
}

func TestDocumentedErrorCodes(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// Verify all documented error codes exist in implementation

		documentedErrorCodes := []response.ErrorCode{
			response.ErrCodeBadRequest,
			response.ErrCodeUnauthorized,
			response.ErrCodeForbidden,
			response.ErrCodeNotFound,
			response.ErrCodeConflict,
			response.ErrCodeValidationFailed,
			response.ErrCodeRateLimitExceeded,
			response.ErrCodeInternalServer,
			response.ErrCodeServiceUnavailable,
			response.ErrCodeDatabaseError,
		}

		for _, errorCode := range documentedErrorCodes {
			// Verify error code is not empty and follows expected pattern
			t.Assert(string(errorCode) != "", true)
			t.Assert(len(string(errorCode)) > 0, true)
		}

		// Test specific error codes match documentation
		t.Assert(string(response.ErrCodeBadRequest), "BAD_REQUEST")
		t.Assert(string(response.ErrCodeNotFound), "NOT_FOUND")
		t.Assert(string(response.ErrCodeValidationFailed), "VALIDATION_FAILED")
		t.Assert(string(response.ErrCodeRateLimitExceeded), "RATE_LIMIT_EXCEEDED")
	})
}

func TestDocumentedTenantModel(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// Verify tenant model matches documentation specification

		tenant := domainTenant.Tenant{
			ID:     "tenant-uuid-1",
			Name:   "Test Tenant",
			Code:   "TEST001",
			Status: domainTenant.StatusActive,
			Config: domainTenant.TenantConfig{
				MaxUsers: 100,
				Features: []string{"basic", "analytics"},
				Theme:    stringPtr("default"),
				Domain:   stringPtr("tenant.example.com"),
			},
			AdminUserID: stringPtr("admin-uuid-1"),
		}

		// Verify all documented fields exist
		t.Assert(tenant.ID, "tenant-uuid-1")
		t.Assert(tenant.Name, "Test Tenant")
		t.Assert(tenant.Code, "TEST001")
		t.Assert(tenant.Status, domainTenant.StatusActive)
		t.Assert(tenant.Config.MaxUsers, 100)
		t.Assert(len(tenant.Config.Features), 2)
		t.Assert(tenant.AdminUserID != nil, true)

		// Verify status enum values match documentation
		validStatuses := []domainTenant.TenantStatus{
			domainTenant.StatusActive,
			domainTenant.StatusSuspended,
			domainTenant.StatusDisabled,
		}

		for _, status := range validStatuses {
			t.Assert(status.IsValid(), true)
		}

		// Verify status string representations
		t.Assert(string(domainTenant.StatusActive), "active")
		t.Assert(string(domainTenant.StatusSuspended), "suspended")
		t.Assert(string(domainTenant.StatusDisabled), "disabled")
	})
}

func TestDocumentedValidationRules(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// Test that validation rules match documentation

		// Test tenant name validation (documented as required, max 255 chars)
		validTenant := domainTenant.NewTenant("Valid Name", "VALID123")
		t.Assert(validTenant.Validate() == nil, true)

		// Test empty name (should fail)
		invalidTenant := domainTenant.NewTenant("", "VALID123")
		t.Assert(invalidTenant.Validate() != nil, true)

		// Test tenant code validation (documented as required, max 100 chars, alphanumeric)
		validTenant2 := domainTenant.NewTenant("Valid Tenant", "VALID123")
		t.Assert(validTenant2.Validate() == nil, true)

		// Test empty code (should fail)
		invalidTenant2 := domainTenant.NewTenant("Valid Name", "")
		t.Assert(invalidTenant2.Validate() != nil, true)

		// Test config validation (documented as maxUsers > 0)
		validConfig := domainTenant.TenantConfig{
			MaxUsers: 100,
			Features: []string{"basic"},
		}

		invalidConfig := domainTenant.TenantConfig{
			MaxUsers: -1, // Should be > 0
			Features: []string{"basic"},
		}

		validTenant3 := domainTenant.NewTenant("Test", "TEST123")
		validTenant3.SetConfig(validConfig)
		t.Assert(validTenant3.Validate() == nil, true)

		validTenant4 := domainTenant.NewTenant("Test", "TEST123")
		err := validTenant4.SetConfig(invalidConfig)
		t.Assert(err != nil, true)
	})
}

func TestDocumentedAPIEndpoints(t *testing.T) {
	gtest.C(t, func(t *gtest.T) {
		// Verify that all documented API endpoints have corresponding handler methods

		// This test ensures the API specification matches the implementation
		handler := handlers.NewTenantHandler()

		// Documented endpoints and their corresponding methods:
		endpoints := map[string]interface{}{
			"GET /tenants":               handler.ListTenants,
			"GET /tenants/{id}":          handler.GetTenant,
			"POST /tenants":              handler.CreateTenant,
			"PUT /tenants/{id}":          handler.UpdateTenant,
			"DELETE /tenants/{id}":       handler.DeleteTenant,
			"PUT /tenants/{id}/activate": handler.ActivateTenant,
			"PUT /tenants/{id}/suspend":  handler.SuspendTenant,
			"PUT /tenants/{id}/disable":  handler.DisableTenant,
		}

		for endpoint, method := range endpoints {
			t.Assert(method != nil, true, "Handler method missing for endpoint: "+endpoint)
		}
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
