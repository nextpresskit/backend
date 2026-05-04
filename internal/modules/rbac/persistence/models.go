package persistence

import "time"

// Role maps to roles.
type Role struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	Name      string    `gorm:"column:name;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

func (Role) TableName() string { return "roles" }

// Permission maps to permissions.
type Permission struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	Code      string    `gorm:"column:code;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

func (Permission) TableName() string { return "permissions" }

// UserRole maps to user_roles (user_id is users.public_id).
type UserRole struct {
	UserID int64  `gorm:"column:user_id;primaryKey"`
	RoleID string `gorm:"column:role_id;type:uuid;primaryKey"`
}

func (UserRole) TableName() string { return "user_roles" }

// RolePermission maps to role_permissions.
type RolePermission struct {
	RoleID       string `gorm:"column:role_id;type:uuid;primaryKey"`
	PermissionID string `gorm:"column:permission_id;type:uuid;primaryKey"`
}

func (RolePermission) TableName() string { return "role_permissions" }
