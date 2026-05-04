package persistence

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrate creates or updates the plugins table.
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Plugin{}); err != nil {
		return fmt.Errorf("plugins: %w", err)
	}
	return nil
}
