package persistence

import (
	"fmt"
	"time"

	"github.com/nextpresskit/backend/pkg/seed/helpers"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const demoSeedRows = 100

// SeedDemo inserts demo pages.
func SeedDemo(tx *gorm.DB) error {
	now := time.Now().UTC()
	for i := 1; i <= demoSeedRows; i++ {
		status := "published"
		var publishedAt *time.Time
		if i%4 == 0 {
			status = "draft"
		} else {
			t := now.Add(-time.Duration(i) * time.Minute)
			publishedAt = &t
		}
		pg := Page{
			UUID:        helpers.SeedUUID(0x0800, i),
			AuthorID:    helpers.UserPublicIDFromUUID(tx, "users", helpers.SeedUUID(0x0100, i)),
			Title:       fmt.Sprintf("Seed Page %03d", i),
			Slug:        fmt.Sprintf("seed-page-%03d", i),
			Content:     fmt.Sprintf("Seeded content body for page %03d.", i),
			Status:      status,
			PublishedAt: publishedAt,
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "slug"}},
			DoUpdates: clause.Assignments(map[string]any{
				"author_id":    pg.AuthorID,
				"title":        pg.Title,
				"content":      pg.Content,
				"status":       pg.Status,
				"published_at": pg.PublishedAt,
				"deleted_at":   nil,
				"updated_at":   time.Now().UTC(),
			}),
		}).Create(&pg).Error; err != nil {
			return fmt.Errorf("pages row %d: %w", i, err)
		}
	}
	return nil
}
