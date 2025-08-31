package main

import (
	"context"

	"github.com/gofromzero/project_temp/backend/pkg/utils"
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
			s.Group("/", func(group *ghttp.RouterGroup) {
				group.ALL("/", func(r *ghttp.Request) {
					r.Response.Write("Hello World!")
				})
				group.GET("/health", func(r *ghttp.Request) {
					healthStatus := g.Map{
						"status":    "ok",
						"message":   "Multi-tenant admin backend is running",
						"database":  "unknown",
						"redis":     "unknown",
						"timestamp": g.NewVar(nil).Time(),
					}

					// Test database connection
					if err := utils.TestDatabaseConnection(); err != nil {
						healthStatus["database"] = "error: " + err.Error()
						healthStatus["status"] = "degraded"
					} else {
						healthStatus["database"] = "healthy"
					}

					// Test Redis connection
					if err := utils.TestRedisConnection(); err != nil {
						healthStatus["redis"] = "error: " + err.Error()
						healthStatus["status"] = "degraded"
					} else {
						healthStatus["redis"] = "healthy"
					}

					r.Response.WriteJson(healthStatus)
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
