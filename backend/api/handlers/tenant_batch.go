package handlers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofromzero/project_temp/backend/application/tenant"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// TenantBatchHandler handles tenant batch operations
type TenantBatchHandler struct {
	batchService *tenant.BatchService
}

// NewTenantBatchHandler creates a new tenant batch handler
func NewTenantBatchHandler() *TenantBatchHandler {
	return &TenantBatchHandler{
		batchService: tenant.NewBatchService(),
	}
}

// ImportTenants handles POST /tenants/import requests
func (h *TenantBatchHandler) ImportTenants(r *ghttp.Request) {
	ctx := r.Context()

	// Parse file upload
	file := r.GetUploadFile("file")
	if file == nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "File upload is required",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Validate file size (50MB limit)
	if file.Size > 50*1024*1024 {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "File size exceeds 50MB limit",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Validate file type
	filename := file.Filename
	if !isValidFileType(filename) {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Invalid file type. Only CSV and Excel files are supported",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Parse additional request parameters
	var req tenant.ImportTenantsRequest
	req.FileName = filename
	req.FileSize = file.Size

	// Get skip duplicates option
	if skipDuplicates := r.Get("skipDuplicates").String(); skipDuplicates == "true" {
		req.SkipDuplicates = true
	}

	// Get dry run option
	if dryRun := r.Get("dryRun").String(); dryRun == "true" {
		req.DryRun = true
	}

	// Process file import
	result, err := h.batchService.ImportTenants(ctx, file, req)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid file format" || err.Error() == "file validation failed" {
			status = http.StatusBadRequest
		}

		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to import tenants: " + err.Error(),
		})
		r.Response.Status = status
		return
	}

	// Return batch operation result
	r.Response.WriteJson(g.Map{
		"success": true,
		"message": "Tenant import initiated successfully",
		"data":    result,
	})
	r.Response.Status = http.StatusOK
}

// ExportTenants handles GET /tenants/export requests
func (h *TenantBatchHandler) ExportTenants(r *ghttp.Request) {
	ctx := r.Context()

	var req tenant.ExportTenantsRequest

	// Parse query parameters
	req.Format = r.Get("format").String()
	if req.Format == "" {
		req.Format = "csv" // Default format
	}

	req.Status = r.Get("status").String()
	req.DateFrom = r.Get("dateFrom").String()
	req.DateTo = r.Get("dateTo").String()

	// Parse include statistics flag
	if includeStats := r.Get("includeStats").String(); includeStats == "true" {
		req.IncludeStatistics = true
	}

	// Process export request
	result, err := h.batchService.ExportTenants(ctx, req)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to export tenants: " + err.Error(),
		})
		r.Response.Status = http.StatusInternalServerError
		return
	}

	// Return batch operation result
	r.Response.WriteJson(g.Map{
		"success": true,
		"message": "Tenant export initiated successfully",
		"data":    result,
	})
	r.Response.Status = http.StatusOK
}

// BulkUpdateStatus handles PATCH /tenants/bulk-status requests
func (h *TenantBatchHandler) BulkUpdateStatus(r *ghttp.Request) {
	ctx := r.Context()

	var req tenant.BulkUpdateStatusRequest

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

	// Process bulk status update
	result, err := h.batchService.BulkUpdateStatus(ctx, req)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to update tenant status: " + err.Error(),
		})
		r.Response.Status = http.StatusInternalServerError
		return
	}

	// Return batch operation result
	r.Response.WriteJson(g.Map{
		"success": true,
		"message": "Bulk status update completed successfully",
		"data":    result,
	})
	r.Response.Status = http.StatusOK
}

// BulkUpdateConfig handles PATCH /tenants/bulk-config requests
func (h *TenantBatchHandler) BulkUpdateConfig(r *ghttp.Request) {
	ctx := r.Context()

	var req tenant.BulkUpdateConfigRequest

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

	// Process bulk config update
	result, err := h.batchService.BulkUpdateConfig(ctx, req)
	if err != nil {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to update tenant config: " + err.Error(),
		})
		r.Response.Status = http.StatusInternalServerError
		return
	}

	// Return batch operation result
	r.Response.WriteJson(g.Map{
		"success": true,
		"message": "Bulk config update completed successfully",
		"data":    result,
	})
	r.Response.Status = http.StatusOK
}

// GetBatchOperationStatus handles GET /tenants/batch/{taskId} requests
func (h *TenantBatchHandler) GetBatchOperationStatus(r *ghttp.Request) {
	ctx := r.Context()

	taskID := r.Get("taskId").String()
	if taskID == "" {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Task ID is required",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Get operation status
	result, err := h.batchService.GetBatchOperationStatus(ctx, taskID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "batch operation not found" {
			status = http.StatusNotFound
		}

		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to get operation status: " + err.Error(),
		})
		r.Response.Status = status
		return
	}

	r.Response.WriteJson(g.Map{
		"success": true,
		"data":    result,
	})
}

// DownloadBatchResult handles GET /tenants/batch/{taskId}/download requests
func (h *TenantBatchHandler) DownloadBatchResult(r *ghttp.Request) {
	ctx := r.Context()

	taskID := r.Get("taskId").String()
	if taskID == "" {
		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Task ID is required",
		})
		r.Response.Status = http.StatusBadRequest
		return
	}

	// Get download information
	downloadInfo, err := h.batchService.GetDownloadInfo(ctx, taskID)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "batch operation not found" || err.Error() == "download not available" {
			status = http.StatusNotFound
		}

		r.Response.WriteJson(g.Map{
			"success": false,
			"error":   "Failed to get download info: " + err.Error(),
		})
		r.Response.Status = status
		return
	}

	// Set content type based on file format
	contentType := "application/octet-stream"
	if downloadInfo.Format == "csv" {
		contentType = "text/csv"
	} else if downloadInfo.Format == "xlsx" {
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	}

	r.Response.Header().Set("Content-Type", contentType)
	r.Response.Header().Set("Content-Disposition", "attachment; filename=\""+downloadInfo.FileName+"\"")
	r.Response.Header().Set("Content-Length", strconv.FormatInt(downloadInfo.FileSize, 10))

	// Serve file for download
	r.Response.ServeFile(downloadInfo.FilePath)
}

// isValidFileType checks if the uploaded file type is supported
func isValidFileType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".csv":
		return true
	case ".xlsx", ".xls":
		return true
	default:
		return false
	}
}
