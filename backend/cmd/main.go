package main

import (
	"context"

	"github.com/gofromzero/project_temp/backend/api/handlers"
	"github.com/gofromzero/project_temp/backend/api/middleware"
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

			s.Group("/", func(group *ghttp.RouterGroup) {
				// Apply middleware to all routes
				group.Middleware(loggingMiddleware.RequestLogger, loggingMiddleware.ErrorLogger, monitoringMiddleware.MetricsCollector)

				group.ALL("/", func(r *ghttp.Request) {
					r.Response.Write("Hello World!")
				})
				group.GET("/health", healthHandler.Health)
				group.GET("/metrics", metricsHandler.Metrics)
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
