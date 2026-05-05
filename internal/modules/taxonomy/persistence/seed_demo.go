package persistence

import (
	"fmt"
	"time"

	"github.com/nextpresskit/backend/pkg/seed/helpers"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const demoSeedRows = 100

// SeedDemo inserts demo categories and tags.
func SeedDemo(tx *gorm.DB) error {
	if err := seedCategories(tx); err != nil {
		return err
	}
	return seedTags(tx)
}

func seedCategories(tx *gorm.DB) error {
	now := time.Now().UTC()
	for i := 1; i <= demoSeedRows; i++ {
		c := Category{
			UUID:      helpers.SeedUUID(0x0400, i),
			Name:      fmt.Sprintf("Category %03d", i),
			Slug:      fmt.Sprintf("category-%03d", i),
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "slug"}},
			DoUpdates: clause.Assignments(map[string]any{
				"name":       c.Name,
				"updated_at": time.Now().UTC(),
			}),
		}).Create(&c).Error; err != nil {
			return fmt.Errorf("categories row %d: %w", i, err)
		}
	}
	return nil
}

func seedTags(tx *gorm.DB) error {
	now := time.Now().UTC()
	for i := 1; i <= demoSeedRows; i++ {
		t := Tag{
			UUID:      helpers.SeedUUID(0x0500, i),
			Name:      fmt.Sprintf("Tag %03d", i),
			Slug:      fmt.Sprintf("tag-%03d", i),
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "slug"}},
			DoUpdates: clause.Assignments(map[string]any{
				"name":       t.Name,
				"updated_at": time.Now().UTC(),
			}),
		}).Create(&t).Error; err != nil {
			return fmt.Errorf("tags row %d: %w", i, err)
		}
	}
	return nil
}
