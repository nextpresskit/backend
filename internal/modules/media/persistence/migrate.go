package persistence

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrate creates or updates the media table.
func AutoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&Media{}); err != nil {
		return fmt.Errorf("media: %w", err)
	}
	return nil
}
