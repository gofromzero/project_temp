package interfaces

import "github.com/gofromzero/project_temp/backend/domain/user"

// UserRepository defines the repository interface for user operations
// This interface is implemented by the MySQL repository layer
type UserRepository interface {
	user.UserRepository
}
