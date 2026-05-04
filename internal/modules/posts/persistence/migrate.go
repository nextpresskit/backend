package persistence

import (
	"fmt"

	"gorm.io/gorm"
)

// AutoMigrate creates or updates all post-related tables in FK-safe order.
func AutoMigrate(db *gorm.DB) error {
	models := []any{
		&Post{},
		&PostCategory{},
		&PostTag{},
		&PostSEO{},
		&PostMetrics{},
		&Series{},
		&PostSeries{},
		&PostCoauthor{},
		&PostGalleryItem{},
		&PostChangelog{},
		&PostSyndication{},
		&TranslationGroup{},
		&PostTranslation{},
	}
	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("posts %T: %w", m, err)
		}
	}
	return nil
}
