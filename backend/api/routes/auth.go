package routes

import (
	"time"

	"github.com/gofromzero/project_temp/backend/api/handlers"
	"github.com/gofromzero/project_temp/backend/api/middleware"
	"github.com/gofromzero/project_temp/backend/application/auth"
	"github.com/gofromzero/project_temp/backend/infr/repository/mysql"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// AuthRoutes sets up authentication-related routes
func AuthRoutes(s *ghttp.Server) error {
	// Create repositories
	userRepo := mysql.NewUserRepository()
	tenantRepo := mysql.NewTenantRepository()

	// Create auth handler
	authHandler, err := handlers.NewAuthHandler(userRepo, tenantRepo)
	if err != nil {
		return err
	}

	// Create auth service for middleware
	authService, err := auth.NewAuthService(userRepo, tenantRepo)
	if err != nil {
		return err
	}

	// Create authentication middleware with public paths
	publicPaths := []string{
		"/auth/login",
		"/auth/register", // Note: register still requires admin token, but skips middleware check
		"/health",
		"/ping",
	}
	
	authMiddleware, err := middleware.NewAuthMiddleware(authService, publicPaths)
	if err != nil {
		return err
	}

	// Public authentication routes (no middleware)
	s.Group("/auth", func(group *ghttp.RouterGroup) {
		group.POST("/login", authHandler.Login)
		// Note: Register requires admin authentication but is handled within the handler
		group.POST("/register", authHandler.Register) 
		group.POST("/refresh", authHandler.RefreshToken)
		group.POST("/logout", authHandler.Logout)
	})

	// Protected routes (require authentication)
	s.Group("/protected", func(group *ghttp.RouterGroup) {
		group.Middleware(authMiddleware.Authenticate)
		group.GET("/profile", authHandler.Profile)
		group.GET("/data", func(r *ghttp.Request) {
			r.Response.WriteJson(g.Map{
				"success": true,
				"data":    "Protected data accessed successfully",
			})
		})
	})

	// Example of system admin only routes
	s.Group("/admin", func(group *ghttp.RouterGroup) {
		group.Middleware(authMiddleware.Authenticate, authMiddleware.RequireSystemAdmin)
		group.GET("/users", func(r *ghttp.Request) {
			// System admin only endpoint
			r.Response.WriteJson(g.Map{
				"success": true,
				"message": "Admin access granted",
				"data":    "System admin users list would be here",
			})
		})
	})

	// Example of role-based protected routes
	s.Group("/api", func(group *ghttp.RouterGroup) {
		group.Middleware(authMiddleware.Authenticate)
		
		// User management endpoints that require admin role
		group.Group("/users", func(userGroup *ghttp.RouterGroup) {
			userGroup.Middleware(authMiddleware.RequireRole("admin"))
			userGroup.GET("/", func(r *ghttp.Request) {
				r.Response.WriteJson(g.Map{
					"success": true,
					"message": "User admin access granted",
					"data":    "User list would be here",
				})
			})
		})

		// Regular user endpoints
		group.Group("/user", func(userGroup *ghttp.RouterGroup) {
			userGroup.GET("/dashboard", func(r *ghttp.Request) {
				r.Response.WriteJson(g.Map{
					"success": true,
					"message": "User dashboard access granted",
					"data":    "User dashboard data would be here",
				})
			})
		})
	})

	return nil
}

// HealthRoutes sets up health check routes (public)
func HealthRoutes(s *ghttp.Server) {
	s.BindHandler("/health", func(r *ghttp.Request) {
		r.Response.WriteJson(g.Map{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	s.BindHandler("/ping", func(r *ghttp.Request) {
		r.Response.Write("pong")
	})
}