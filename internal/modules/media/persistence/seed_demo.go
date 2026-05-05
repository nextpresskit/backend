package persistence

import (
	"fmt"
	"time"

	"github.com/nextpresskit/backend/pkg/seed/helpers"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const demoSeedRows = 100

// SeedDemo inserts demo media rows.
func SeedDemo(tx *gorm.DB) error {
	now := time.Now().UTC()
	for i := 1; i <= demoSeedRows; i++ {
		m := Media{
			UUID:         helpers.SeedUUID(0x0600, i),
			UploaderID:   helpers.UserPublicIDFromUUID(tx, "users", helpers.SeedUUID(0x0100, i)),
			OriginalName: fmt.Sprintf("image-%03d.jpg", i),
			StorageName:  fmt.Sprintf("seed-image-%03d.jpg", i),
			MimeType:     "image/jpeg",
			SizeBytes:    int64(1024 + i),
			StoragePath:  fmt.Sprintf("uploads/seed-image-%03d.jpg", i),
			PublicURL:    fmt.Sprintf("/uploads/seed-image-%03d.jpg", i),
			CreatedAt:    now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "storage_name"}},
			DoUpdates: clause.Assignments(map[string]any{
				"uploader_id":   m.UploaderID,
				"original_name": m.OriginalName,
				"mime_type":     m.MimeType,
				"size_bytes":    m.SizeBytes,
				"storage_path":  m.StoragePath,
				"public_url":    m.PublicURL,
			}),
		}).Create(&m).Error; err != nil {
			return fmt.Errorf("media row %d: %w", i, err)
		}
	}
	return nil
}
