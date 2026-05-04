package persistence

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrate creates or updates the pages table.
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Page{}); err != nil {
		return fmt.Errorf("pages: %w", err)
	}
	return nil
}
