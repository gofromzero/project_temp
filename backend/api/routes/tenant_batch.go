package routes

import (
	"github.com/gofromzero/project_temp/backend/api/handlers"
	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterTenantBatchRoutes registers all tenant batch operation routes
func RegisterTenantBatchRoutes(group *ghttp.RouterGroup) {
	batchHandler := handlers.NewTenantBatchHandler()

	// Batch operations
	group.POST("/tenants/import", batchHandler.ImportTenants)
	group.GET("/tenants/export", batchHandler.ExportTenants)
	group.PATCH("/tenants/bulk-status", batchHandler.BulkUpdateStatus)
	group.PATCH("/tenants/bulk-config", batchHandler.BulkUpdateConfig)

	// Batch operation status and results
	group.GET("/tenants/batch/{taskId}", batchHandler.GetBatchOperationStatus)
	group.GET("/tenants/batch/{taskId}/download", batchHandler.DownloadBatchResult)
}
