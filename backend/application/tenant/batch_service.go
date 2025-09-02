package tenant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	domainTenant "github.com/gofromzero/project_temp/backend/domain/tenant"
	"github.com/gofromzero/project_temp/backend/pkg/files"
	"github.com/gofromzero/project_temp/backend/pkg/queue"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/guid"
)

// BatchService handles tenant batch operations
type BatchService struct {
	tenantService *TenantService
	fileProcessor *files.FileProcessor
	taskQueue     *queue.TaskQueue
}

// NewBatchService creates a new batch service
func NewBatchService() *BatchService {
	return &BatchService{
		tenantService: NewTenantService(),
		fileProcessor: files.NewFileProcessor(),
		taskQueue:     queue.NewTaskQueue(),
	}
}

// ImportTenantsRequest represents the request to import tenants
type ImportTenantsRequest struct {
	FileName       string `json:"fileName"`
	FileSize       int64  `json:"fileSize"`
	SkipDuplicates bool   `json:"skipDuplicates"`
	DryRun         bool   `json:"dryRun"`
}

// ExportTenantsRequest represents the request to export tenants
type ExportTenantsRequest struct {
	Format            string `json:"format" validate:"required,oneof=csv xlsx"`
	Status            string `json:"status,omitempty"`
	DateFrom          string `json:"dateFrom,omitempty"`
	DateTo            string `json:"dateTo,omitempty"`
	IncludeStatistics bool   `json:"includeStatistics"`
}

// BulkUpdateStatusRequest represents the request to bulk update tenant status
type BulkUpdateStatusRequest struct {
	TenantIDs []string                  `json:"tenantIds" validate:"required,min=1,max=1000"`
	Status    domainTenant.TenantStatus `json:"status" validate:"required"`
	Reason    string                    `json:"reason,omitempty"`
}

// BulkUpdateConfigRequest represents the request to bulk update tenant config
type BulkUpdateConfigRequest struct {
	TenantIDs []string                   `json:"tenantIds" validate:"required,min=1,max=1000"`
	Config    *domainTenant.TenantConfig `json:"config" validate:"required"`
	MergeMode string                     `json:"mergeMode,omitempty" validate:"omitempty,oneof=replace merge"`
}

// BatchOperationResult represents the result of a batch operation
type BatchOperationResult struct {
	TaskID      string                 `json:"taskId"`
	Status      string                 `json:"status"`
	Summary     BatchOperationSummary  `json:"summary"`
	Errors      []BatchError           `json:"errors,omitempty"`
	DownloadURL *string                `json:"downloadUrl,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	CompletedAt *time.Time             `json:"completedAt,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// BatchOperationSummary represents the summary of batch operation results
type BatchOperationSummary struct {
	Total   int `json:"total"`
	Success int `json:"success"`
	Failed  int `json:"failed"`
	Skipped int `json:"skipped"`
}

// BatchError represents an error that occurred during batch processing
type BatchError struct {
	Row      *int    `json:"row,omitempty"`
	TenantID *string `json:"tenantId,omitempty"`
	Field    *string `json:"field,omitempty"`
	Message  string  `json:"message"`
	Code     string  `json:"code"`
}

// DownloadInfo represents download information for batch results
type DownloadInfo struct {
	TaskID    string    `json:"taskId"`
	FileName  string    `json:"fileName"`
	FilePath  string    `json:"filePath"`
	Format    string    `json:"format"`
	FileSize  int64     `json:"fileSize"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// ImportTenants processes tenant import from uploaded file
func (s *BatchService) ImportTenants(ctx context.Context, file *ghttp.UploadFile, req ImportTenantsRequest) (*BatchOperationResult, error) {
	taskID := guid.S()

	// Create initial batch operation result
	result := &BatchOperationResult{
		TaskID:    taskID,
		Status:    "pending",
		CreatedAt: time.Now(),
		Summary: BatchOperationSummary{
			Total:   0,
			Success: 0,
			Failed:  0,
			Skipped: 0,
		},
		Metadata: map[string]interface{}{
			"operation": "import",
			"fileName":  req.FileName,
			"fileSize":  req.FileSize,
			"dryRun":    req.DryRun,
		},
	}

	// Parse file content
	tenantData, err := s.parseImportFile(file, req.FileName)
	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, BatchError{
			Message: "Failed to parse file: " + err.Error(),
			Code:    "FILE_PARSE_ERROR",
		})
		return result, err
	}

	result.Summary.Total = len(tenantData)

	// For small batches (< 100), process synchronously
	if len(tenantData) < 100 {
		return s.processImportSync(ctx, tenantData, req, result)
	}

	// For large batches, queue for async processing
	return s.queueImportAsync(ctx, tenantData, req, result)
}

// ExportTenants processes tenant export request
func (s *BatchService) ExportTenants(ctx context.Context, req ExportTenantsRequest) (*BatchOperationResult, error) {
	taskID := guid.S()

	result := &BatchOperationResult{
		TaskID:    taskID,
		Status:    "pending",
		CreatedAt: time.Now(),
		Metadata: map[string]interface{}{
			"operation":    "export",
			"format":       req.Format,
			"includeStats": req.IncludeStatistics,
		},
	}

	// Build filters for export
	filters := s.buildExportFilters(req)

	// Get tenant count for the export
	tenants, total, err := s.tenantService.tenantDomainService.ListTenants(filters, -1, 0)
	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, BatchError{
			Message: "Failed to fetch tenants: " + err.Error(),
			Code:    "DATA_FETCH_ERROR",
		})
		return result, err
	}

	result.Summary.Total = int(total)

	// For small exports (< 1000), process synchronously
	if total < 1000 {
		return s.processExportSync(ctx, tenants, req, result)
	}

	// For large exports, queue for async processing
	return s.queueExportAsync(ctx, filters, req, result)
}

// BulkUpdateStatus processes bulk tenant status updates
func (s *BatchService) BulkUpdateStatus(ctx context.Context, req BulkUpdateStatusRequest) (*BatchOperationResult, error) {
	taskID := guid.S()

	result := &BatchOperationResult{
		TaskID:    taskID,
		Status:    "processing",
		CreatedAt: time.Now(),
		Summary: BatchOperationSummary{
			Total: len(req.TenantIDs),
		},
		Metadata: map[string]interface{}{
			"operation": "bulk_status_update",
			"status":    req.Status,
			"reason":    req.Reason,
		},
	}

	// Process updates
	for i, tenantID := range req.TenantIDs {
		// Validate tenant exists
		tenant, err := s.tenantService.GetTenant(tenantID)
		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, BatchError{
				TenantID: &tenantID,
				Message:  "Tenant not found: " + err.Error(),
				Code:     "TENANT_NOT_FOUND",
			})
			continue
		}

		// Skip if already in target status
		if tenant.Status == req.Status {
			result.Summary.Skipped++
			continue
		}

		// Update status through service
		var updatedTenant *domainTenant.Tenant
		switch req.Status {
		case domainTenant.StatusActive:
			updatedTenant, err = s.tenantService.ActivateTenant(tenantID)
		case domainTenant.StatusSuspended:
			updatedTenant, err = s.tenantService.SuspendTenant(tenantID)
		case domainTenant.StatusDisabled:
			updatedTenant, err = s.tenantService.DisableTenant(tenantID)
		default:
			err = errors.New("invalid status")
		}

		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, BatchError{
				TenantID: &tenantID,
				Message:  "Failed to update status: " + err.Error(),
				Code:     "STATUS_UPDATE_ERROR",
			})
		} else {
			result.Summary.Success++
			result.Metadata[fmt.Sprintf("tenant_%d_updated", i)] = updatedTenant.ID
		}
	}

	result.Status = "completed"
	completedAt := time.Now()
	result.CompletedAt = &completedAt

	return result, nil
}

// BulkUpdateConfig processes bulk tenant config updates
func (s *BatchService) BulkUpdateConfig(ctx context.Context, req BulkUpdateConfigRequest) (*BatchOperationResult, error) {
	taskID := guid.S()

	result := &BatchOperationResult{
		TaskID:    taskID,
		Status:    "processing",
		CreatedAt: time.Now(),
		Summary: BatchOperationSummary{
			Total: len(req.TenantIDs),
		},
		Metadata: map[string]interface{}{
			"operation": "bulk_config_update",
			"mergeMode": req.MergeMode,
		},
	}

	// Process updates
	for i, tenantID := range req.TenantIDs {
		// Get current tenant
		tenant, err := s.tenantService.GetTenant(tenantID)
		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, BatchError{
				TenantID: &tenantID,
				Message:  "Tenant not found: " + err.Error(),
				Code:     "TENANT_NOT_FOUND",
			})
			continue
		}

		// Prepare config update
		var updateConfig *domainTenant.TenantConfig
		if req.MergeMode == "merge" {
			// Merge with existing config
			updateConfig = s.mergeConfigs(&tenant.Config, req.Config)
		} else {
			// Replace config
			updateConfig = req.Config
		}

		// Update tenant config
		updateRequest := UpdateTenantRequest{
			Config: updateConfig,
		}

		_, err = s.tenantService.UpdateTenant(tenantID, updateRequest)
		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, BatchError{
				TenantID: &tenantID,
				Message:  "Failed to update config: " + err.Error(),
				Code:     "CONFIG_UPDATE_ERROR",
			})
		} else {
			result.Summary.Success++
			result.Metadata[fmt.Sprintf("tenant_%d_updated", i)] = tenantID
		}
	}

	result.Status = "completed"
	completedAt := time.Now()
	result.CompletedAt = &completedAt

	return result, nil
}

// GetBatchOperationStatus retrieves the status of a batch operation
func (s *BatchService) GetBatchOperationStatus(ctx context.Context, taskID string) (*BatchOperationResult, error) {
	// Get task status from queue/storage
	task, err := s.taskQueue.GetTask(taskID)
	if err != nil {
		return nil, errors.New("batch operation not found")
	}

	// Convert task to BatchOperationResult
	result := &BatchOperationResult{
		TaskID:    task.ID,
		Status:    task.Status,
		CreatedAt: task.CreatedAt,
	}

	if task.CompletedAt != nil {
		result.CompletedAt = task.CompletedAt
	}

	// Parse task result if available
	if task.Result != nil {
		if err := json.Unmarshal(task.Result, &result.Summary); err == nil {
			// Successfully parsed summary
		}
	}

	// Parse errors if available
	if task.Errors != nil {
		if err := json.Unmarshal(task.Errors, &result.Errors); err == nil {
			// Successfully parsed errors
		}
	}

	return result, nil
}

// GetDownloadInfo retrieves download information for a batch operation result
func (s *BatchService) GetDownloadInfo(ctx context.Context, taskID string) (*DownloadInfo, error) {
	// Get task to verify it exists and is completed
	task, err := s.taskQueue.GetTask(taskID)
	if err != nil {
		return nil, errors.New("batch operation not found")
	}

	if task.Status != "completed" {
		return nil, errors.New("download not available")
	}

	// Check if download file exists
	downloadPath := fmt.Sprintf("/tmp/batch_results/%s", taskID)
	// This would typically check file system or storage service

	return &DownloadInfo{
		TaskID:    taskID,
		FileName:  fmt.Sprintf("batch_result_%s.xlsx", taskID),
		FilePath:  downloadPath + ".xlsx",
		Format:    "xlsx",
		FileSize:  0,                                  // Would be determined from actual file
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}, nil
}

// Helper methods for internal processing

func (s *BatchService) parseImportFile(file *ghttp.UploadFile, filename string) ([]map[string]interface{}, error) {
	// This would delegate to file processor service
	return s.fileProcessor.ParseFile(file, filename)
}

func (s *BatchService) processImportSync(ctx context.Context, data []map[string]interface{}, req ImportTenantsRequest, result *BatchOperationResult) (*BatchOperationResult, error) {
	result.Status = "processing"

	for i, tenantData := range data {
		if req.DryRun {
			// Just validate, don't create
			if err := s.validateTenantData(tenantData); err != nil {
				result.Summary.Failed++
				result.Errors = append(result.Errors, BatchError{
					Row:     &i,
					Message: "Validation failed: " + err.Error(),
					Code:    "VALIDATION_ERROR",
				})
			} else {
				result.Summary.Success++
			}
			continue
		}

		// Create tenant
		err := s.createTenantFromData(tenantData, req.SkipDuplicates)
		if err != nil {
			result.Summary.Failed++
			result.Errors = append(result.Errors, BatchError{
				Row:     &i,
				Message: "Creation failed: " + err.Error(),
				Code:    "CREATION_ERROR",
			})
		} else {
			result.Summary.Success++
		}
	}

	result.Status = "completed"
	completedAt := time.Now()
	result.CompletedAt = &completedAt

	return result, nil
}

func (s *BatchService) queueImportAsync(ctx context.Context, data []map[string]interface{}, req ImportTenantsRequest, result *BatchOperationResult) (*BatchOperationResult, error) {
	// Queue task for async processing
	task := &queue.Task{
		ID:        result.TaskID,
		Type:      "tenant_import",
		Status:    "pending",
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"tenantData": data,
			"request":    req,
		},
	}

	if err := s.taskQueue.EnqueueTask(task); err != nil {
		result.Status = "failed"
		return result, err
	}

	return result, nil
}

func (s *BatchService) processExportSync(ctx context.Context, tenants []*domainTenant.Tenant, req ExportTenantsRequest, result *BatchOperationResult) (*BatchOperationResult, error) {
	result.Status = "processing"

	// Generate export file
	filePath, err := s.fileProcessor.GenerateExportFile(tenants, req.Format, req.IncludeStatistics)
	if err != nil {
		result.Status = "failed"
		result.Errors = append(result.Errors, BatchError{
			Message: "Export generation failed: " + err.Error(),
			Code:    "EXPORT_ERROR",
		})
		return result, err
	}

	result.Status = "completed"
	result.Summary.Success = len(tenants)
	completedAt := time.Now()
	result.CompletedAt = &completedAt

	// Set download URL
	downloadURL := fmt.Sprintf("/v1/tenants/batch/%s/download", result.TaskID)
	result.DownloadURL = &downloadURL
	result.Metadata["filePath"] = filePath

	return result, nil
}

func (s *BatchService) queueExportAsync(ctx context.Context, filters map[string]interface{}, req ExportTenantsRequest, result *BatchOperationResult) (*BatchOperationResult, error) {
	// Queue task for async processing
	task := &queue.Task{
		ID:        result.TaskID,
		Type:      "tenant_export",
		Status:    "pending",
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"filters": filters,
			"request": req,
		},
	}

	if err := s.taskQueue.EnqueueTask(task); err != nil {
		result.Status = "failed"
		return result, err
	}

	return result, nil
}

func (s *BatchService) buildExportFilters(req ExportTenantsRequest) map[string]interface{} {
	filters := make(map[string]interface{})

	if req.Status != "" {
		filters["status"] = req.Status
	}

	if req.DateFrom != "" {
		filters["dateFrom"] = req.DateFrom
	}

	if req.DateTo != "" {
		filters["dateTo"] = req.DateTo
	}

	return filters
}

func (s *BatchService) validateTenantData(data map[string]interface{}) error {
	name, ok := data["name"].(string)
	if !ok || name == "" {
		return errors.New("name is required")
	}

	code, ok := data["code"].(string)
	if !ok || code == "" {
		return errors.New("code is required")
	}

	return nil
}

func (s *BatchService) createTenantFromData(data map[string]interface{}, skipDuplicates bool) error {
	// Parse tenant data
	name := data["name"].(string)
	code := data["code"].(string)

	// Check for duplicates if needed
	if skipDuplicates {
		if _, err := s.tenantService.GetTenantByCode(code); err == nil {
			// Tenant already exists, skip
			return nil
		}
	}

	// Build create request
	req := CreateTenantRequest{
		Name: name,
		Code: code,
	}

	// Parse config if provided
	if configData, ok := data["config"].(map[string]interface{}); ok {
		config := &domainTenant.TenantConfig{}
		if maxUsers, ok := configData["maxUsers"].(float64); ok {
			config.MaxUsers = int(maxUsers)
		}
		if features, ok := configData["features"].([]interface{}); ok {
			for _, f := range features {
				if feature, ok := f.(string); ok {
					config.Features = append(config.Features, feature)
				}
			}
		}
		req.Config = config
	}

	_, err := s.tenantService.CreateTenant(req)
	return err
}

func (s *BatchService) mergeConfigs(existing, new *domainTenant.TenantConfig) *domainTenant.TenantConfig {
	result := *existing // Copy existing config

	// Merge fields from new config
	if new.MaxUsers > 0 {
		result.MaxUsers = new.MaxUsers
	}

	if len(new.Features) > 0 {
		result.Features = new.Features
	}

	if new.Theme != nil {
		result.Theme = new.Theme
	}

	if new.Domain != nil {
		result.Domain = new.Domain
	}

	return &result
}
