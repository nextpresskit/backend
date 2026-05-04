package persistence

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrate creates or updates RBAC tables (after users).
func AutoMigrate(db *gorm.DB) error {
	for _, m := range []any{&Role{}, &Permission{}, &UserRole{}, &RolePermission{}} {
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("rbac %T: %w", m, err)
		}
	}
	return nil
}
