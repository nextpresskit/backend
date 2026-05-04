package seed

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	mediap "github.com/nextpresskit/backend/internal/modules/media/persistence"
	pagep "github.com/nextpresskit/backend/internal/modules/pages/persistence"
	pluginp "github.com/nextpresskit/backend/internal/modules/plugins/persistence"
	postp "github.com/nextpresskit/backend/internal/modules/posts/persistence"
	rbacp "github.com/nextpresskit/backend/internal/modules/rbac/persistence"
	taxp "github.com/nextpresskit/backend/internal/modules/taxonomy/persistence"
	userp "github.com/nextpresskit/backend/internal/modules/user/persistence"
)

const (
	seedRows               = 100
	seedDefaultPermissions = 13

	superadminRoleName = "superadmin"
	superadminUserID   = "00000000-0000-0000-0100-000000000001"
	superadminRoleID   = "00000000-0000-0000-0200-000000000002"
)

// SeedFullDataset inserts deterministic demo records across all tables.
// It also ensures a superadmin user exists and is linked to admin/superadmin roles.
func SeedFullDataset(db *gorm.DB) error {
	superadminEmail := envOrDefault("SEED_SUPERADMIN_EMAIL", "superadmin@nextpresskit.local")
	superadminPassword := envOrDefault("SEED_SUPERADMIN_PASSWORD", "SuperAdmin123!")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(superadminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash superadmin password: %w", err)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := seedUsers(tx, superadminEmail, string(passwordHash)); err != nil {
			return err
		}
		if err := seedRoles(tx); err != nil {
			return err
		}
		if err := seedPermissions(tx); err != nil {
			return err
		}
		if err := seedRolePermissions(tx); err != nil {
			return err
		}
		if err := seedUserRoles(tx); err != nil {
			return err
		}
		if err := seedCategories(tx); err != nil {
			return err
		}
		if err := seedTags(tx); err != nil {
			return err
		}
		if err := seedMedia(tx); err != nil {
			return err
		}
		if err := seedPosts(tx); err != nil {
			return err
		}
		if err := seedPages(tx); err != nil {
			return err
		}
		if err := seedPlugins(tx); err != nil {
			return err
		}
		if err := seedPostCategories(tx); err != nil {
			return err
		}
		if err := seedPostTags(tx); err != nil {
			return err
		}
		if err := seedPostSEO(tx); err != nil {
			return err
		}
		if err := seedPostMetrics(tx); err != nil {
			return err
		}
		if err := seedSeries(tx); err != nil {
			return err
		}
		if err := seedPostSeries(tx); err != nil {
			return err
		}
		if err := seedPostCoauthors(tx); err != nil {
			return err
		}
		if err := seedPostGalleryItems(tx); err != nil {
			return err
		}
		if err := seedPostChangelog(tx); err != nil {
			return err
		}
		if err := seedPostSyndication(tx); err != nil {
			return err
		}
		if err := seedTranslationGroups(tx); err != nil {
			return err
		}
		if err := seedPostTranslations(tx); err != nil {
			return err
		}
		return nil
	})
}

func seedUsers(tx *gorm.DB, superadminEmail, passwordHash string) error {
	for i := 1; i <= seedRows; i++ {
		id := seedUUID(0x0100, i)
		firstName := fmt.Sprintf("User%03d", i)
		lastName := "Seed"
		email := fmt.Sprintf("user%03d@nextpresskit.local", i)
		if i == 1 {
			id = superadminUserID
			firstName = "Super"
			lastName = "Admin"
			email = superadminEmail
		}
		u := userp.User{
			UUID:      id,
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Password:  passwordHash,
			Active:    true,
		}
		now := time.Now().UTC()
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "email"}},
			DoUpdates: clause.Assignments(map[string]any{
				"first_name": firstName,
				"last_name":  lastName,
				"password":   passwordHash,
				"active":     true,
				"deleted_at": nil,
				"updated_at": now,
			}),
		}).Omit("public_id").Create(&u).Error; err != nil {
			return fmt.Errorf("users row %d: %w", i, err)
		}
	}
	return nil
}

func seedRoles(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		id := seedUUID(0x0200, i)
		name := fmt.Sprintf("role-%03d", i)
		if i == 1 {
			id = RoleAdminID
			name = "admin"
		} else if i == 2 {
			id = superadminRoleID
			name = superadminRoleName
		}
		r := rbacp.Role{ID: id, Name: name}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "name"}},
			DoUpdates: clause.Assignments(map[string]any{
				"updated_at": time.Now().UTC(),
			}),
		}).Create(&r).Error; err != nil {
			return fmt.Errorf("roles row %d: %w", i, err)
		}
	}
	return nil
}

func seedPermissions(tx *gorm.DB) error {
	for i := 1; i <= seedRows-seedDefaultPermissions; i++ {
		p := rbacp.Permission{
			ID:   seedUUID(0x0300, i),
			Code: fmt.Sprintf("seed:permission:%03d", i),
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "code"}},
			DoUpdates: clause.Assignments(map[string]any{
				"updated_at": time.Now().UTC(),
			}),
		}).Create(&p).Error; err != nil {
			return fmt.Errorf("permissions row %d: %w", i, err)
		}
	}
	return nil
}

func seedUserRoles(tx *gorm.DB) error {
	superPub := userPublicIDFromUUID(tx, superadminUserID)
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rbacp.UserRole{
		UserID: superPub,
		RoleID: superadminRoleID,
	}).Error; err != nil {
		return fmt.Errorf("user_roles superadmin link: %w", err)
	}
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rbacp.UserRole{
		UserID: superPub,
		RoleID: RoleAdminID,
	}).Error; err != nil {
		return fmt.Errorf("user_roles admin link: %w", err)
	}
	for i := 2; i <= 99; i++ {
		ur := rbacp.UserRole{
			UserID: userPublicIDFromUUID(tx, seedUUID(0x0100, i)),
			RoleID: seedUUID(0x0200, i+1),
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&ur).Error; err != nil {
			return fmt.Errorf("user_roles row %d: %w", i, err)
		}
	}
	return nil
}

func seedRolePermissions(tx *gorm.DB) error {
	defaultCodes := []string{
		"admin:ping",
		"rbac:manage",
		"posts:read",
		"posts:write",
		"pages:read",
		"pages:write",
		"categories:read",
		"categories:write",
		"tags:read",
		"tags:write",
		"media:read",
		"media:write",
		"plugins:manage",
	}
	for _, code := range defaultCodes {
		var permID string
		if err := tx.Model(&rbacp.Permission{}).Select("id").Where("code = ?", code).Scan(&permID).Error; err != nil {
			return err
		}
		if permID == "" {
			continue
		}
		rp := rbacp.RolePermission{RoleID: RoleAdminID, PermissionID: permID}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rp).Error; err != nil {
			return fmt.Errorf("role_permissions default %s: %w", code, err)
		}
	}
	for i := 1; i <= seedRows-seedDefaultPermissions; i++ {
		var permID, roleID string
		code := fmt.Sprintf("seed:permission:%03d", i)
		roleName := fmt.Sprintf("role-%03d", i+seedDefaultPermissions)
		if err := tx.Model(&rbacp.Permission{}).Select("id").Where("code = ?", code).Scan(&permID).Error; err != nil {
			return err
		}
		if err := tx.Model(&rbacp.Role{}).Select("id").Where("name = ?", roleName).Scan(&roleID).Error; err != nil {
			return err
		}
		if permID == "" || roleID == "" {
			continue
		}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rbacp.RolePermission{
			RoleID: roleID, PermissionID: permID,
		}).Error; err != nil {
			return fmt.Errorf("role_permissions row %d: %w", i, err)
		}
	}
	return nil
}

func seedCategories(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		c := taxp.Category{
			ID:   seedUUID(0x0400, i),
			Name: fmt.Sprintf("Category %03d", i),
			Slug: fmt.Sprintf("category-%03d", i),
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
	for i := 1; i <= seedRows; i++ {
		t := taxp.Tag{
			ID:   seedUUID(0x0500, i),
			Name: fmt.Sprintf("Tag %03d", i),
			Slug: fmt.Sprintf("tag-%03d", i),
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

func seedMedia(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		m := mediap.Media{
			ID:           seedUUID(0x0600, i),
			UploaderID:   userPublicIDFromUUID(tx, seedUUID(0x0100, i)),
			OriginalName: fmt.Sprintf("image-%03d.jpg", i),
			StorageName:  fmt.Sprintf("seed-image-%03d.jpg", i),
			MimeType:     "image/jpeg",
			SizeBytes:    int64(1024 + i),
			StoragePath:  fmt.Sprintf("uploads/seed-image-%03d.jpg", i),
			PublicURL:    fmt.Sprintf("/uploads/seed-image-%03d.jpg", i),
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

func seedPosts(tx *gorm.DB) error {
	now := time.Now().UTC()
	for i := 1; i <= seedRows; i++ {
		status := "draft"
		var publishedAt *time.Time
		if i%3 == 0 {
			status = "published"
			t := now.Add(-time.Duration(i) * time.Hour)
			publishedAt = &t
		}
		postID := seedUUID(0x0700, i)
		pid := postID
		focalX := float32(0.5)
		focalY := float32(0.5)
		w, h := 1920, 1080
		alt := fmt.Sprintf("Featured alt %03d", i)
		credit := "Seed Generator"
		license := "CC-BY"
		catID := seedUUID(0x0400, i)
		mediaID := seedUUID(0x0600, i)
		p := postp.Post{
			ID:                 postID,
			UUID:               &pid,
			AuthorID:           userPublicIDFromUUID(tx, seedUUID(0x0100, i)),
			Title:              fmt.Sprintf("Seed Post %03d", i),
			Slug:               fmt.Sprintf("seed-post-%03d", i),
			Content:            fmt.Sprintf("Seeded content body for post %03d.", i),
			Status:             status,
			PublishedAt:        publishedAt,
			PostType:           "article",
			Format:             "standard",
			Subtitle:           fmt.Sprintf("Subtitle for post %03d", i),
			Excerpt:            fmt.Sprintf("Short excerpt for post %03d", i),
			Visibility:         "public",
			Locale:             "en-US",
			Timezone:           "UTC",
			ScheduledPublishAt: ptrTime(now.Add(time.Duration(i) * time.Hour)),
			FirstIndexedAt:     &now,
			ReviewerUserID:     int64Ptr(userPublicIDFromUUID(tx, seedUUID(0x0100, ((i)%seedRows)+1))),
			LastEditedByUserID: int64Ptr(userPublicIDFromUUID(tx, seedUUID(0x0100, ((i+1)%seedRows)+1))),
			WorkflowStage:      "draft",
			Revision:           1,
			CustomFields:       jf(`{"seed_index": %d}`, i),
			Flags:              j(`{"featured": false}`),
			Engagement:         jf(`{"score": %d}`, i),
			Workflow:           jf(`{"state": "draft-%03d"}`, i),
			FeaturedMediaID:    &mediaID,
			FeaturedAlt:        &alt,
			FeaturedWidth:      &w,
			FeaturedHeight:     &h,
			FeaturedFocalX:     &focalX,
			FeaturedFocalY:     &focalY,
			FeaturedCredit:     &credit,
			FeaturedLicense:    &license,
			PrimaryCategoryID:  &catID,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "slug"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"author_id", "title", "content", "status", "published_at", "uuid", "post_type", "format",
				"subtitle", "excerpt", "visibility", "locale", "timezone", "scheduled_publish_at", "first_indexed_at",
				"reviewer_user_id", "last_edited_by_user_id", "workflow_stage", "revision",
				"custom_fields", "flags", "engagement", "workflow",
				"featured_media_id", "featured_alt", "featured_width", "featured_height",
				"featured_focal_x", "featured_focal_y", "featured_credit", "featured_license",
				"primary_category_id", "deleted_at", "updated_at",
			}),
		}).Create(&p).Error; err != nil {
			return fmt.Errorf("posts row %d: %w", i, err)
		}
	}
	return nil
}

func seedPages(tx *gorm.DB) error {
	now := time.Now().UTC()
	for i := 1; i <= seedRows; i++ {
		status := "published"
		var publishedAt *time.Time
		if i%4 == 0 {
			status = "draft"
		} else {
			t := now.Add(-time.Duration(i) * time.Minute)
			publishedAt = &t
		}
		pg := pagep.Page{
			ID:          seedUUID(0x0800, i),
			AuthorID:    userPublicIDFromUUID(tx, seedUUID(0x0100, i)),
			Title:       fmt.Sprintf("Seed Page %03d", i),
			Slug:        fmt.Sprintf("seed-page-%03d", i),
			Content:     fmt.Sprintf("Seeded content body for page %03d.", i),
			Status:      status,
			PublishedAt: publishedAt,
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

func seedPlugins(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		cfg := jf(`{"seed_index": %d}`, i)
		pl := pluginp.Plugin{
			ID:        seedUUID(0x0b00, i),
			Name:      fmt.Sprintf("Seed Plugin %03d", i),
			Slug:      fmt.Sprintf("seed-plugin-%03d", i),
			Enabled:   i%2 == 0,
			Version:   "1.0.0",
			ConfigRaw: cfg,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "slug"}},
			DoUpdates: clause.Assignments(map[string]any{
				"name":       pl.Name,
				"enabled":    pl.Enabled,
				"version":    pl.Version,
				"config":     pl.ConfigRaw,
				"updated_at": time.Now().UTC(),
			}),
		}).Create(&pl).Error; err != nil {
			return fmt.Errorf("plugins row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostCategories(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		row := postp.PostCategory{PostID: seedUUID(0x0700, i), CategoryID: seedUUID(0x0400, i)}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "post_id"}, {Name: "category_id"}},
			DoNothing: true,
		}).Create(&row).Error; err != nil {
			return fmt.Errorf("post_categories row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostTags(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		row := postp.PostTag{PostID: seedUUID(0x0700, i), TagID: seedUUID(0x0500, i)}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "post_id"}, {Name: "tag_id"}},
			DoNothing: true,
		}).Create(&row).Error; err != nil {
			return fmt.Errorf("post_tags row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostSEO(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		t := fmt.Sprintf("SEO Title %03d", i)
		d := fmt.Sprintf("SEO Description %03d", i)
		canon := fmt.Sprintf("https://example.local/seed-post-%03d", i)
		robots := "index,follow"
		ogt := "article"
		ogu := fmt.Sprintf("https://example.local/media/seed-image-%03d.jpg", i)
		tw := "summary_large_image"
		row := postp.PostSEO{
			PostID:         seedUUID(0x0700, i),
			Title:          &t,
			Description:    &d,
			CanonicalURL:   &canon,
			Robots:         &robots,
			OGType:         &ogt,
			OGImageURL:     &ogu,
			TwitterCard:    &tw,
			StructuredData: jf(`{"seed_index": %d}`, i),
			UpdatedAt:      time.Now().UTC(),
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "post_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"title":           row.Title,
				"description":     row.Description,
				"canonical_url":   row.CanonicalURL,
				"robots":          row.Robots,
				"og_type":         row.OGType,
				"og_image_url":    row.OGImageURL,
				"twitter_card":    row.TwitterCard,
				"structured_data": row.StructuredData,
				"updated_at":      time.Now().UTC(),
			}),
		}).Create(&row).Error; err != nil {
			return fmt.Errorf("post_seo row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostMetrics(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		row := postp.PostMetrics{
			PostID:                seedUUID(0x0700, i),
			WordCount:             800 + i,
			CharacterCount:        5000 + i*20,
			ReadingTimeMinutes:    5,
			EstReadTimeSeconds:    300,
			ViewCount:             int64(i * 100),
			UniqueVisitors7d:      int64(i * 10),
			ScrollDepthAvgPercent: 60.5,
			BounceRatePercent:     35.0,
			AvgTimeOnPageSeconds:  240,
			CommentCount:          i % 20,
			LikeCount:             i % 25,
			ShareCount:            i % 15,
			BookmarkCount:         i % 10,
			UpdatedAt:             time.Now().UTC(),
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "post_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"word_count": row.WordCount, "character_count": row.CharacterCount,
				"reading_time_minutes": row.ReadingTimeMinutes, "est_read_time_seconds": row.EstReadTimeSeconds,
				"view_count": row.ViewCount, "unique_visitors_7d": row.UniqueVisitors7d,
				"scroll_depth_avg_percent": row.ScrollDepthAvgPercent, "bounce_rate_percent": row.BounceRatePercent,
				"avg_time_on_page_seconds": row.AvgTimeOnPageSeconds, "comment_count": row.CommentCount,
				"like_count": row.LikeCount, "share_count": row.ShareCount, "bookmark_count": row.BookmarkCount,
				"updated_at": time.Now().UTC(),
			}),
		}).Create(&row).Error; err != nil {
			return fmt.Errorf("post_metrics row %d: %w", i, err)
		}
	}
	return nil
}

func seedSeries(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		s := postp.Series{
			ID:    seedUUID(0x0c00, i),
			Title: fmt.Sprintf("Series %03d", i),
			Slug:  fmt.Sprintf("series-%03d", i),
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "slug"}},
			DoUpdates: clause.Assignments(map[string]any{
				"title": s.Title, "updated_at": time.Now().UTC(),
			}),
		}).Create(&s).Error; err != nil {
			return fmt.Errorf("series row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostSeries(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		pi := i
		lbl := fmt.Sprintf("Part %03d", i)
		row := postp.PostSeries{
			PostID:    seedUUID(0x0700, i),
			SeriesID:  seedUUID(0x0c00, i),
			PartIndex: &pi,
			PartLabel: &lbl,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "post_id"}, {Name: "series_id"}},
			DoNothing: true,
		}).Create(&row).Error; err != nil {
			return fmt.Errorf("post_series row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostCoauthors(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		uid := userPublicIDFromUUID(tx, seedUUID(0x0100, ((i)%seedRows)+1))
		row := postp.PostCoauthor{PostID: seedUUID(0x0700, i), UserID: uid, SortOrder: 1}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "post_id"}, {Name: "user_id"}},
			DoNothing: true,
		}).Create(&row).Error; err != nil {
			return fmt.Errorf("post_coauthors row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostGalleryItems(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		cap := fmt.Sprintf("Gallery caption %03d", i)
		alt := fmt.Sprintf("Gallery alt %03d", i)
		row := postp.PostGalleryItem{
			ID:        seedUUID(0x0d00, i),
			PostID:    seedUUID(0x0700, i),
			MediaID:   seedUUID(0x0600, i),
			SortOrder: i,
			Caption:   &cap,
			Alt:       &alt,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"post_id": row.PostID, "media_id": row.MediaID, "sort_order": row.SortOrder,
				"caption": row.Caption, "alt": row.Alt,
			}),
		}).Create(&row).Error; err != nil {
			return fmt.Errorf("post_gallery_items row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostChangelog(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		u := userPublicIDFromUUID(tx, seedUUID(0x0100, ((i+2)%seedRows)+1))
		note := fmt.Sprintf("Seed changelog entry %03d", i)
		row := postp.PostChangelog{
			ID:     seedUUID(0x0e00, i),
			PostID: seedUUID(0x0700, i),
			At:     time.Now().UTC().Add(-time.Duration(i) * time.Hour),
			UserID: int64Ptr(u),
			Note:   note,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"post_id": row.PostID, "at": row.At, "user_id": row.UserID, "note": row.Note,
			}),
		}).Create(&row).Error; err != nil {
			return fmt.Errorf("post_changelog row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostSyndication(tx *gorm.DB) error {
	now := time.Now().UTC()
	for i := 1; i <= seedRows; i++ {
		row := postp.PostSyndication{
			ID:        seedUUID(0x0f00, i),
			PostID:    seedUUID(0x0700, i),
			Platform:  "medium",
			URL:       fmt.Sprintf("https://medium.example/seed-post-%03d", i),
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"post_id": row.PostID, "platform": row.Platform, "url": row.URL, "status": row.Status,
				"updated_at": time.Now().UTC(),
			}),
		}).Create(&row).Error; err != nil {
			return fmt.Errorf("post_syndication row %d: %w", i, err)
		}
	}
	return nil
}

func seedTranslationGroups(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		row := postp.TranslationGroup{ID: seedUUID(0x1000, i), CreatedAt: time.Now().UTC()}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&row).Error; err != nil {
			return fmt.Errorf("translation_groups row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostTranslations(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		locale := "en-US"
		if i%2 == 0 {
			locale = "bg-BG"
		}
		row := postp.PostTranslation{
			PostID:  seedUUID(0x0700, i),
			GroupID: seedUUID(0x1000, i),
			Locale:  locale,
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "post_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"group_id": row.GroupID, "locale": row.Locale,
			}),
		}).Create(&row).Error; err != nil {
			return fmt.Errorf("post_translations row %d: %w", i, err)
		}
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func seedUUID(namespace, index int) string {
	return fmt.Sprintf("00000000-0000-0000-%04x-%012x", namespace, index)
}

func userPublicIDFromUUID(tx *gorm.DB, userUUID string) int64 {
	var id int64
	_ = tx.Model(&userp.User{}).Select("public_id").Where("id = ?", userUUID).Scan(&id).Error
	return id
}

func j(s string) json.RawMessage { return json.RawMessage([]byte(s)) }

func jf(format string, a ...any) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(format, a...))
}

func int64Ptr(v int64) *int64 {
	if v <= 0 {
		return nil
	}
	return &v
}

func ptrTime(t time.Time) *time.Time { return &t }
