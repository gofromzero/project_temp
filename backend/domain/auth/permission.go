package auth

// PermissionScope represents the scope of a permission
type PermissionScope string

const (
	ScopeSystem PermissionScope = "system"
	ScopeTenant PermissionScope = "tenant"
	ScopeSelf   PermissionScope = "self"
)

// Permission represents the permission domain entity in RBAC model
type Permission struct {
	ID          string          `json:"id" gorm:"type:varchar(36);primaryKey"`
	Name        string          `json:"name" gorm:"type:varchar(100);not null"`
	Code        string          `json:"code" gorm:"type:varchar(100);not null;uniqueIndex;comment:format: resource.action"`
	Resource    string          `json:"resource" gorm:"type:varchar(50);not null;index"`
	Action      string          `json:"action" gorm:"type:varchar(50);not null;index"`
	Scope       PermissionScope `json:"scope" gorm:"type:enum('system','tenant','self');not null"`
	Description *string         `json:"description,omitempty" gorm:"type:text"`
	IsSystem    bool            `json:"isSystem" gorm:"default:false;comment:system built-in permission"`
}

// TableName returns the table name for GORM
func (Permission) TableName() string {
	return "permissions"
}

// IsSystemPermission checks if the permission is a system built-in permission
func (p *Permission) IsSystemPermission() bool {
	return p.IsSystem
}

// GetFullCode returns the full permission code (resource.action format)
func (p *Permission) GetFullCode() string {
	return p.Resource + "." + p.Action
}

// PermissionRepository defines the repository interface for permission operations
type PermissionRepository interface {
	Create(permission *Permission) error
	GetByID(id string) (*Permission, error)
	GetByCode(code string) (*Permission, error)
	GetByResource(resource string) ([]*Permission, error)
	GetByScope(scope PermissionScope) ([]*Permission, error)
	List(offset, limit int) ([]*Permission, error)
	Update(permission *Permission) error
	Delete(id string) error
	Count() (int64, error)
	ExistsByCode(code string) (bool, error)
}