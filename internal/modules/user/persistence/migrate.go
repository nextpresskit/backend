package persistence

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrate creates or updates the users table.
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("users: %w", err)
	}
	return nil
}
