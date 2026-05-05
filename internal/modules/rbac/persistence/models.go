package persistence

import "time"

// Role maps to roles (bigint id + public uuid).
type Role struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UUID      string    `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	Name      string    `gorm:"column:name;not null;unique"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

func (Role) TableName() string { return "roles" }

// Permission maps to permissions.
type Permission struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UUID      string    `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	Code      string    `gorm:"column:code;not null;unique"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;autoUpdateTime"`
}

func (Permission) TableName() string { return "permissions" }

// UserRole maps to user_roles (user_id is users.id).
type UserRole struct {
	UserID int64 `gorm:"column:user_id;primaryKey"`
	RoleID int64 `gorm:"column:role_id;primaryKey"`
}

func (UserRole) TableName() string { return "user_roles" }

// RolePermission maps to role_permissions.
type RolePermission struct {
	RoleID       int64 `gorm:"column:role_id;primaryKey"`
	PermissionID int64 `gorm:"column:permission_id;primaryKey"`
}

func (RolePermission) TableName() string { return "role_permissions" }
