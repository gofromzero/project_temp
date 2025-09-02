package routes

import (
	"github.com/gofromzero/project_temp/backend/api/handlers"
	"github.com/gogf/gf/v2/net/ghttp"
)

// RegisterUserRoutes registers all user management routes
func RegisterUserRoutes(group *ghttp.RouterGroup) {
	userHandler := handlers.NewUserHandler()

	// User CRUD operations
	group.POST("/users", userHandler.CreateUser)
	group.GET("/users", userHandler.ListUsers)
	group.GET("/users/{id}", userHandler.GetUser)
	group.PUT("/users/{id}", userHandler.UpdateUser)
	group.DELETE("/users/{id}", userHandler.DeleteUser)
}