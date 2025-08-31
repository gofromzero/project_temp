package handlers

import (
	"net/http"

	"github.com/gofromzero/project_temp/backend/application/tenant"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// TenantHandler handles tenant-related HTTP requests
type TenantHandler struct {
	tenantService *tenant.TenantService
}

// NewTenantHandler creates a new tenant handler instance
func NewTenantHandler() *TenantHandler {
	return &TenantHandler{
		tenantService: tenant.NewTenantService(),
	}
}

// CreateTenant handles POST /tenants requests
func (h *TenantHandler) CreateTenant(r *ghttp.Request) {
	ctx := r.Context()
	
	var req tenant.CreateTenantRequest
	
	// Parse request body
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Validate request
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Validation failed: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Create tenant
	response, err := h.tenantService.CreateTenant(req)
	if err != nil {
		// Determine status code based on error type
		status := http.StatusInternalServerError
		if err.Error() == "tenant with code already exists" {
			status = http.StatusConflict
		}
		
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to create tenant: " + err.Error(),
		})
		r.Response.Status = status
		return
	}

	// Return success response
	r.Response.WriteJson(g.Map{
		"success": true,
		"message": "Tenant created successfully",
		"data":    response,
	})
	r.Response.Status = http.StatusCreated
}

// GetTenant handles GET /tenants/{id} requests
func (h *TenantHandler) GetTenant(r *ghttp.Request) {
	tenantID := r.Get("id").String()
	if tenantID == "" {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Tenant ID is required",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	tenant, err := h.tenantService.GetTenant(tenantID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "tenant not found" {
			status = http.StatusNotFound
		}
		
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to get tenant: " + err.Error(),
		})
		r.Response.Status = status
		return
	}

	r.Response.WriteJson(g.Map{
		"success": true,
		"data":    tenant,
	})
}

// UpdateTenant handles PUT /tenants/{id} requests
func (h *TenantHandler) UpdateTenant(r *ghttp.Request) {
	ctx := r.Context()
	
	tenantID := r.Get("id").String()
	if tenantID == "" {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Tenant ID is required",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	var req tenant.UpdateTenantRequest
	
	// Parse request body
	if err := r.Parse(&req); err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Invalid request format: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Validate request
	if err := g.Validator().Data(req).Run(ctx); err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Validation failed: " + err.Error(),
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Update tenant
	updatedTenant, err := h.tenantService.UpdateTenant(tenantID, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "tenant not found" {
			status = http.StatusNotFound
		}
		
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to update tenant: " + err.Error(),
		})
		r.Response.Status = status
		return
	}

	r.Response.WriteJson(g.Map{
		"success": true,
		"message": "Tenant updated successfully",
		"data":    updatedTenant,
	})
}

// ActivateTenant handles PUT /tenants/{id}/activate requests
func (h *TenantHandler) ActivateTenant(r *ghttp.Request) {
	tenantID := r.Get("id").String()
	if tenantID == "" {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Tenant ID is required",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	tenant, err := h.tenantService.ActivateTenant(tenantID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "tenant not found" {
			status = http.StatusNotFound
		}
		
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to activate tenant: " + err.Error(),
		})
		r.Response.Status = status
		return
	}

	r.Response.WriteJson(g.Map{
		"success": true,
		"message": "Tenant activated successfully",
		"data":    tenant,
	})
}

// SuspendTenant handles PUT /tenants/{id}/suspend requests
func (h *TenantHandler) SuspendTenant(r *ghttp.Request) {
	tenantID := r.Get("id").String()
	if tenantID == "" {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Tenant ID is required",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	tenant, err := h.tenantService.SuspendTenant(tenantID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "tenant not found" {
			status = http.StatusNotFound
		}
		
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to suspend tenant: " + err.Error(),
		})
		r.Response.Status = status
		return
	}

	r.Response.WriteJson(g.Map{
		"success": true,
		"message": "Tenant suspended successfully",
		"data":    tenant,
	})
}

// DisableTenant handles PUT /tenants/{id}/disable requests
func (h *TenantHandler) DisableTenant(r *ghttp.Request) {
	tenantID := r.Get("id").String()
	if tenantID == "" {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Tenant ID is required",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	tenant, err := h.tenantService.DisableTenant(tenantID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "tenant not found" {
			status = http.StatusNotFound
		}
		
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to disable tenant: " + err.Error(),
		})
		r.Response.Status = status
		return
	}

	r.Response.WriteJson(g.Map{
		"success": true,
		"message": "Tenant disabled successfully",
		"data":    tenant,
	})
}