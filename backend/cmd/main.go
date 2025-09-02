package main

import (
	"context"

	"github.com/gofromzero/project_temp/backend/api/handlers"
	"github.com/gofromzero/project_temp/backend/api/middleware"
	"github.com/gofromzero/project_temp/backend/api/routes"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcmd"
)

var (
	Main = gcmd.Command{
		Name:  "main",
		Usage: "main",
		Brief: "start http server",
		Func: func(ctx context.Context, parser *gcmd.Parser) (err error) {
			s := g.Server()

			// Initialize handlers
			healthHandler := handlers.NewHealthHandler()
			metricsHandler := handlers.NewMetricsHandler()

			// Initialize middleware
			loggingMiddleware := middleware.NewLoggingMiddleware()
			monitoringMiddleware := middleware.NewMonitoringMiddleware()

			// Initialize authentication middleware
			// Note: In production, this should use proper repository injection
			publicPaths := []string{"/health", "/metrics", "/"}

			// Create auth middleware with minimal setup for development
			authMiddleware := middleware.NewMinimalAuthMiddleware(publicPaths)
			g.Log().Info(ctx, "Auth middleware initialized for batch routes")

			// Public routes
			s.Group("/", func(group *ghttp.RouterGroup) {
				// Apply middleware to all routes
				group.Middleware(loggingMiddleware.RequestLogger, loggingMiddleware.ErrorLogger, monitoringMiddleware.MetricsCollector)

				group.ALL("/", func(r *ghttp.Request) {
					r.Response.Write("Hello World!")
				})
				group.GET("/health", healthHandler.Health)
				group.GET("/metrics", metricsHandler.Metrics)
			})

			// API v1 routes with authentication
			s.Group("/v1", func(group *ghttp.RouterGroup) {
				// Apply middleware to all API routes
				group.Middleware(loggingMiddleware.RequestLogger, loggingMiddleware.ErrorLogger, monitoringMiddleware.MetricsCollector)

				// Protected user management routes - require authentication (tenant admin or system admin)
				group.Group("/", func(userGroup *ghttp.RouterGroup) {
					userGroup.Middleware(authMiddleware.Authenticate)
					routes.RegisterUserRoutes(userGroup)
				})

				// Protected batch operations - require system admin access
				group.Group("/", func(batchGroup *ghttp.RouterGroup) {
					batchGroup.Middleware(authMiddleware.Authenticate, authMiddleware.RequireSystemAdmin)
					routes.RegisterTenantBatchRoutes(batchGroup)
				})
			})
			s.SetPort(8000)
			s.Run()
			return nil
		},
	}
)

func main() {
	Main.Run(context.Background())
}
