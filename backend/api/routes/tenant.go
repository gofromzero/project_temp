package routes

import (
	"github.com/gofromzero/project_temp/backend/api/handlers"
	"github.com/gofromzero/project_temp/backend/api/middleware"
	"github.com/gofromzero/project_temp/backend/application/auth"
	"github.com/gofromzero/project_temp/backend/infr/repository/mysql"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// TenantRoutes sets up tenant-related routes
func TenantRoutes(s *ghttp.Server) error {
	// Create tenant handler
	tenantHandler := handlers.NewTenantHandler()

	// Create repositories for auth middleware
	userRepo := mysql.NewUserRepository()
	tenantRepo := mysql.NewTenantRepository()

	// Create auth service for middleware
	authService, err := auth.NewAuthService(userRepo, tenantRepo)
	if err != nil {
		return err
	}

	// Create authentication middleware
	publicPaths := []string{
		"/auth/login",
		"/auth/register",
		"/health",
		"/ping",
	}
	
	authMiddleware, err := middleware.NewAuthMiddleware(authService, publicPaths)
	if err != nil {
		return err
	}

	// Tenant management routes - System Admin only
	s.Group("/tenants", func(group *ghttp.RouterGroup) {
		// All tenant operations require system admin privileges
		group.Middleware(authMiddleware.Authenticate, authMiddleware.RequireSystemAdmin)
		
		// CRUD operations
		group.POST("/", tenantHandler.CreateTenant)           // POST /tenants
		group.GET("/:id", tenantHandler.GetTenant)            // GET /tenants/{id}
		group.PUT("/:id", tenantHandler.UpdateTenant)         // PUT /tenants/{id}
		
		// Status management operations
		group.PUT("/:id/activate", tenantHandler.ActivateTenant)  // PUT /tenants/{id}/activate
		group.PUT("/:id/suspend", tenantHandler.SuspendTenant)    // PUT /tenants/{id}/suspend
		group.PUT("/:id/disable", tenantHandler.DisableTenant)    // PUT /tenants/{id}/disable
	})

	// Tenant self-service routes - Tenant admin only
	s.Group("/tenant", func(group *ghttp.RouterGroup) {
		// These routes allow tenant admins to manage their own tenant
		group.Middleware(authMiddleware.Authenticate, authMiddleware.RequireRole("admin"))
		
		// Get current tenant info
		group.GET("/", func(r *ghttp.Request) {
			// Extract tenant ID from context (set by auth middleware)
			tenantID := r.GetCtx().Value("tenantId")
			if tenantID == nil {
				r.Response.WriteJson(g.Map{
					"success": false,
					"error":   "Tenant context not found",
				})
				r.Response.Status = 400
				return
			}

			// Use the GetTenant handler by setting the ID parameter
			r.SetParam("id", tenantID.(string))
			tenantHandler.GetTenant(r)
		})
		
		// Update current tenant (limited fields)
		group.PUT("/", func(r *ghttp.Request) {
			// Extract tenant ID from context
			tenantID := r.GetCtx().Value("tenantId")
			if tenantID == nil {
				r.Response.WriteJson(g.Map{
					"success": false,
					"error":   "Tenant context not found",
				})
				r.Response.Status = 400
				return
			}

			// Use the UpdateTenant handler but restrict to current tenant
			r.SetParam("id", tenantID.(string))
			tenantHandler.UpdateTenant(r)
		})
	})

	return nil
}