package persistence

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrate creates or updates taxonomy tables.
func AutoMigrate(db *gorm.DB) error {
	for _, m := range []any{&Category{}, &Tag{}} {
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("taxonomy %T: %w", m, err)
		}
	}
	return nil
}
