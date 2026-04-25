package seed

import (
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
	superadminEmail := envOrDefault("SEED_SUPERADMIN_EMAIL", "superadmin@nextpress.local")
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
		email := fmt.Sprintf("user%03d@nextpress.local", i)
		if i == 1 {
			id = superadminUserID
			firstName = "Super"
			lastName = "Admin"
			email = superadminEmail
		}

		if err := tx.Exec(
			`INSERT INTO users (id, first_name, last_name, email, password, active, deleted_at)
			 VALUES (?, ?, ?, ?, ?, TRUE, NULL)
			 ON CONFLICT (email) DO UPDATE
			 SET first_name = EXCLUDED.first_name,
			     last_name = EXCLUDED.last_name,
			     password = EXCLUDED.password,
			     active = EXCLUDED.active,
			     deleted_at = NULL,
			     updated_at = NOW()`,
			id, firstName, lastName, email, passwordHash,
		).Error; err != nil {
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

		if err := tx.Exec(
			`INSERT INTO roles (id, name)
			 VALUES (?, ?)
			 ON CONFLICT (name) DO UPDATE
			 SET updated_at = NOW()`,
			id, name,
		).Error; err != nil {
			return fmt.Errorf("roles row %d: %w", i, err)
		}
	}
	return nil
}

func seedPermissions(tx *gorm.DB) error {
	for i := 1; i <= seedRows-seedDefaultPermissions; i++ {
		if err := tx.Exec(
			`INSERT INTO permissions (id, code)
			 VALUES (?, ?)
			 ON CONFLICT (code) DO UPDATE
			 SET updated_at = NOW()`,
			seedUUID(0x0300, i), fmt.Sprintf("seed:permission:%03d", i),
		).Error; err != nil {
			return fmt.Errorf("permissions row %d: %w", i, err)
		}
	}
	return nil
}

func seedUserRoles(tx *gorm.DB) error {
	if err := tx.Exec(
		`INSERT INTO user_roles (user_id, role_id)
		 SELECT (SELECT public_id FROM users WHERE id = ?), id FROM roles WHERE name = ?
		 ON CONFLICT DO NOTHING`,
		superadminUserID, superadminRoleName,
	).Error; err != nil {
		return fmt.Errorf("user_roles superadmin link: %w", err)
	}

	if err := tx.Exec(
		`INSERT INTO user_roles (user_id, role_id)
		 VALUES ((SELECT public_id FROM users WHERE id = ?), ?)
		 ON CONFLICT DO NOTHING`,
		superadminUserID, RoleAdminID,
	).Error; err != nil {
		return fmt.Errorf("user_roles admin link: %w", err)
	}

	for i := 2; i <= 99; i++ {
		if err := tx.Exec(
			`INSERT INTO user_roles (user_id, role_id)
			 SELECT (SELECT public_id FROM users WHERE id = ?), id FROM roles WHERE name = ?
			 ON CONFLICT DO NOTHING`,
			seedUUID(0x0100, i), fmt.Sprintf("role-%03d", i+1),
		).Error; err != nil {
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
		if err := tx.Exec(
			`INSERT INTO role_permissions (role_id, permission_id)
			 SELECT r.id, p.id
			 FROM roles r
			 JOIN permissions p ON p.code = ?
			 WHERE r.name = 'admin'
			 ON CONFLICT DO NOTHING`,
			code,
		).Error; err != nil {
			return fmt.Errorf("role_permissions default %s: %w", code, err)
		}
	}

	for i := 1; i <= seedRows-seedDefaultPermissions; i++ {
		if err := tx.Exec(
			`INSERT INTO role_permissions (role_id, permission_id)
			 SELECT r.id, p.id
			 FROM roles r
			 JOIN permissions p ON p.code = ?
			 WHERE r.name = ?
			 ON CONFLICT DO NOTHING`,
			fmt.Sprintf("seed:permission:%03d", i),
			fmt.Sprintf("role-%03d", i+seedDefaultPermissions),
		).Error; err != nil {
			return fmt.Errorf("role_permissions row %d: %w", i, err)
		}
	}
	return nil
}

func seedCategories(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO categories (id, name, slug)
			 VALUES (?, ?, ?)
			 ON CONFLICT (slug) DO UPDATE
			 SET name = EXCLUDED.name, updated_at = NOW()`,
			seedUUID(0x0400, i), fmt.Sprintf("Category %03d", i), fmt.Sprintf("category-%03d", i),
		).Error; err != nil {
			return fmt.Errorf("categories row %d: %w", i, err)
		}
	}
	return nil
}

func seedTags(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO tags (id, name, slug)
			 VALUES (?, ?, ?)
			 ON CONFLICT (slug) DO UPDATE
			 SET name = EXCLUDED.name, updated_at = NOW()`,
			seedUUID(0x0500, i), fmt.Sprintf("Tag %03d", i), fmt.Sprintf("tag-%03d", i),
		).Error; err != nil {
			return fmt.Errorf("tags row %d: %w", i, err)
		}
	}
	return nil
}

func seedMedia(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO media (id, uploader_id, original_name, storage_name, mime_type, size_bytes, storage_path, public_url)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (storage_name) DO UPDATE
			 SET uploader_id = EXCLUDED.uploader_id,
			     original_name = EXCLUDED.original_name,
			     mime_type = EXCLUDED.mime_type,
			     size_bytes = EXCLUDED.size_bytes,
			     storage_path = EXCLUDED.storage_path,
			     public_url = EXCLUDED.public_url`,
			seedUUID(0x0600, i),
			userPublicIDFromUUID(tx, seedUUID(0x0100, i)),
			fmt.Sprintf("image-%03d.jpg", i),
			fmt.Sprintf("seed-image-%03d.jpg", i),
			"image/jpeg",
			int64(1024+i),
			fmt.Sprintf("uploads/seed-image-%03d.jpg", i),
			fmt.Sprintf("/uploads/seed-image-%03d.jpg", i),
		).Error; err != nil {
			return fmt.Errorf("media row %d: %w", i, err)
		}
	}
	return nil
}

func seedPosts(tx *gorm.DB) error {
	now := time.Now().UTC()
	for i := 1; i <= seedRows; i++ {
		status := "draft"
		var publishedAt any
		if i%3 == 0 {
			status = "published"
			publishedAt = now.Add(-time.Duration(i) * time.Hour)
		}

		postID := seedUUID(0x0700, i)
		if err := tx.Exec(
			`INSERT INTO posts (
			    id, author_id, title, slug, content, status, published_at,
			    uuid, post_type, format, subtitle, excerpt, visibility, locale, timezone,
			    scheduled_publish_at, first_indexed_at, reviewer_user_id, last_edited_by_user_id,
			    workflow_stage, revision, custom_fields, flags, engagement, workflow,
			    featured_media_id, featured_alt, featured_width, featured_height, featured_focal_x, featured_focal_y,
			    featured_credit, featured_license, primary_category_id, deleted_at
			 )
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?::jsonb, ?::jsonb, ?::jsonb, ?::jsonb, ?, ?, ?, ?, ?, ?, ?, ?, ?, NULL)
			 ON CONFLICT (slug) DO UPDATE
			 SET author_id = EXCLUDED.author_id,
			     title = EXCLUDED.title,
			     content = EXCLUDED.content,
			     status = EXCLUDED.status,
			     published_at = EXCLUDED.published_at,
			     uuid = EXCLUDED.uuid,
			     post_type = EXCLUDED.post_type,
			     format = EXCLUDED.format,
			     subtitle = EXCLUDED.subtitle,
			     excerpt = EXCLUDED.excerpt,
			     visibility = EXCLUDED.visibility,
			     locale = EXCLUDED.locale,
			     timezone = EXCLUDED.timezone,
			     scheduled_publish_at = EXCLUDED.scheduled_publish_at,
			     first_indexed_at = EXCLUDED.first_indexed_at,
			     reviewer_user_id = EXCLUDED.reviewer_user_id,
			     last_edited_by_user_id = EXCLUDED.last_edited_by_user_id,
			     workflow_stage = EXCLUDED.workflow_stage,
			     revision = EXCLUDED.revision,
			     custom_fields = EXCLUDED.custom_fields,
			     flags = EXCLUDED.flags,
			     engagement = EXCLUDED.engagement,
			     workflow = EXCLUDED.workflow,
			     featured_media_id = EXCLUDED.featured_media_id,
			     featured_alt = EXCLUDED.featured_alt,
			     featured_width = EXCLUDED.featured_width,
			     featured_height = EXCLUDED.featured_height,
			     featured_focal_x = EXCLUDED.featured_focal_x,
			     featured_focal_y = EXCLUDED.featured_focal_y,
			     featured_credit = EXCLUDED.featured_credit,
			     featured_license = EXCLUDED.featured_license,
			     primary_category_id = EXCLUDED.primary_category_id,
			     deleted_at = NULL,
			     updated_at = NOW()`,
			postID,
			userPublicIDFromUUID(tx, seedUUID(0x0100, i)),
			fmt.Sprintf("Seed Post %03d", i),
			fmt.Sprintf("seed-post-%03d", i),
			fmt.Sprintf("Seeded content body for post %03d.", i),
			status,
			publishedAt,
			postID,
			"article",
			"standard",
			fmt.Sprintf("Subtitle for post %03d", i),
			fmt.Sprintf("Short excerpt for post %03d", i),
			"public",
			"en-US",
			"UTC",
			now.Add(time.Duration(i)*time.Hour),
			now,
			userPublicIDFromUUID(tx, seedUUID(0x0100, ((i)%seedRows)+1)),
			userPublicIDFromUUID(tx, seedUUID(0x0100, ((i+1)%seedRows)+1)),
			"draft",
			1,
			fmt.Sprintf(`{"seed_index": %d}`, i),
			`{"featured": false}`,
			fmt.Sprintf(`{"score": %d}`, i),
			fmt.Sprintf(`{"state": "draft-%03d"}`, i),
			seedUUID(0x0600, i),
			fmt.Sprintf("Featured alt %03d", i),
			1920,
			1080,
			0.5,
			0.5,
			"Seed Generator",
			"CC-BY",
			seedUUID(0x0400, i),
		).Error; err != nil {
			return fmt.Errorf("posts row %d: %w", i, err)
		}
	}
	return nil
}

func seedPages(tx *gorm.DB) error {
	now := time.Now().UTC()
	for i := 1; i <= seedRows; i++ {
		status := "published"
		var publishedAt any = now.Add(-time.Duration(i) * time.Minute)
		if i%4 == 0 {
			status = "draft"
			publishedAt = nil
		}

		if err := tx.Exec(
			`INSERT INTO pages (id, author_id, title, slug, content, status, published_at, deleted_at)
			 VALUES (?, ?, ?, ?, ?, ?, ?, NULL)
			 ON CONFLICT (slug) DO UPDATE
			 SET author_id = EXCLUDED.author_id,
			     title = EXCLUDED.title,
			     content = EXCLUDED.content,
			     status = EXCLUDED.status,
			     published_at = EXCLUDED.published_at,
			     deleted_at = NULL,
			     updated_at = NOW()`,
			seedUUID(0x0800, i),
			userPublicIDFromUUID(tx, seedUUID(0x0100, i)),
			fmt.Sprintf("Seed Page %03d", i),
			fmt.Sprintf("seed-page-%03d", i),
			fmt.Sprintf("Seeded content body for page %03d.", i),
			status,
			publishedAt,
		).Error; err != nil {
			return fmt.Errorf("pages row %d: %w", i, err)
		}
	}
	return nil
}

func seedPlugins(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO plugins (id, name, slug, enabled, version, config)
			 VALUES (?, ?, ?, ?, ?, ?::jsonb)
			 ON CONFLICT (slug) DO UPDATE
			 SET name = EXCLUDED.name,
			     enabled = EXCLUDED.enabled,
			     version = EXCLUDED.version,
			     config = EXCLUDED.config,
			     updated_at = NOW()`,
			seedUUID(0x0b00, i),
			fmt.Sprintf("Seed Plugin %03d", i),
			fmt.Sprintf("seed-plugin-%03d", i),
			i%2 == 0,
			"1.0.0",
			fmt.Sprintf(`{"seed_index": %d}`, i),
		).Error; err != nil {
			return fmt.Errorf("plugins row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostCategories(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO post_categories (post_id, category_id)
			 VALUES (?, ?)
			 ON CONFLICT DO NOTHING`,
			seedUUID(0x0700, i), seedUUID(0x0400, i),
		).Error; err != nil {
			return fmt.Errorf("post_categories row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostTags(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO post_tags (post_id, tag_id)
			 VALUES (?, ?)
			 ON CONFLICT DO NOTHING`,
			seedUUID(0x0700, i), seedUUID(0x0500, i),
		).Error; err != nil {
			return fmt.Errorf("post_tags row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostSEO(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO post_seo (post_id, title, description, canonical_url, robots, og_type, og_image_url, twitter_card, structured_data)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?::jsonb)
			 ON CONFLICT (post_id) DO UPDATE
			 SET title = EXCLUDED.title,
			     description = EXCLUDED.description,
			     canonical_url = EXCLUDED.canonical_url,
			     robots = EXCLUDED.robots,
			     og_type = EXCLUDED.og_type,
			     og_image_url = EXCLUDED.og_image_url,
			     twitter_card = EXCLUDED.twitter_card,
			     structured_data = EXCLUDED.structured_data,
			     updated_at = NOW()`,
			seedUUID(0x0700, i),
			fmt.Sprintf("SEO Title %03d", i),
			fmt.Sprintf("SEO Description %03d", i),
			fmt.Sprintf("https://example.local/seed-post-%03d", i),
			"index,follow",
			"article",
			fmt.Sprintf("https://example.local/media/seed-image-%03d.jpg", i),
			"summary_large_image",
			fmt.Sprintf(`{"seed_index": %d}`, i),
		).Error; err != nil {
			return fmt.Errorf("post_seo row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostMetrics(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO post_metrics (
			    post_id, word_count, character_count, reading_time_minutes, est_read_time_seconds, view_count,
			    unique_visitors_7d, scroll_depth_avg_percent, bounce_rate_percent, avg_time_on_page_seconds,
			    comment_count, like_count, share_count, bookmark_count
			 )
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			 ON CONFLICT (post_id) DO UPDATE
			 SET word_count = EXCLUDED.word_count,
			     character_count = EXCLUDED.character_count,
			     reading_time_minutes = EXCLUDED.reading_time_minutes,
			     est_read_time_seconds = EXCLUDED.est_read_time_seconds,
			     view_count = EXCLUDED.view_count,
			     unique_visitors_7d = EXCLUDED.unique_visitors_7d,
			     scroll_depth_avg_percent = EXCLUDED.scroll_depth_avg_percent,
			     bounce_rate_percent = EXCLUDED.bounce_rate_percent,
			     avg_time_on_page_seconds = EXCLUDED.avg_time_on_page_seconds,
			     comment_count = EXCLUDED.comment_count,
			     like_count = EXCLUDED.like_count,
			     share_count = EXCLUDED.share_count,
			     bookmark_count = EXCLUDED.bookmark_count,
			     updated_at = NOW()`,
			seedUUID(0x0700, i),
			800+i,
			5000+i*20,
			5,
			300,
			int64(i*100),
			int64(i*10),
			60.5,
			35.0,
			240,
			i%20,
			i%25,
			i%15,
			i%10,
		).Error; err != nil {
			return fmt.Errorf("post_metrics row %d: %w", i, err)
		}
	}
	return nil
}

func seedSeries(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO series (id, title, slug)
			 VALUES (?, ?, ?)
			 ON CONFLICT (slug) DO UPDATE
			 SET title = EXCLUDED.title, updated_at = NOW()`,
			seedUUID(0x0c00, i),
			fmt.Sprintf("Series %03d", i),
			fmt.Sprintf("series-%03d", i),
		).Error; err != nil {
			return fmt.Errorf("series row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostSeries(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO post_series (post_id, series_id, part_index, part_label)
			 VALUES (?, ?, ?, ?)
			 ON CONFLICT DO NOTHING`,
			seedUUID(0x0700, i),
			seedUUID(0x0c00, i),
			i,
			fmt.Sprintf("Part %03d", i),
		).Error; err != nil {
			return fmt.Errorf("post_series row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostCoauthors(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO post_coauthors (post_id, user_id, sort_order)
			 VALUES (?, ?, ?)
			 ON CONFLICT DO NOTHING`,
			seedUUID(0x0700, i),
			userPublicIDFromUUID(tx, seedUUID(0x0100, ((i)%seedRows)+1)),
			1,
		).Error; err != nil {
			return fmt.Errorf("post_coauthors row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostGalleryItems(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO post_gallery_items (id, post_id, media_id, sort_order, caption, alt)
			 VALUES (?, ?, ?, ?, ?, ?)
			 ON CONFLICT (id) DO UPDATE
			 SET post_id = EXCLUDED.post_id,
			     media_id = EXCLUDED.media_id,
			     sort_order = EXCLUDED.sort_order,
			     caption = EXCLUDED.caption,
			     alt = EXCLUDED.alt`,
			seedUUID(0x0d00, i),
			seedUUID(0x0700, i),
			seedUUID(0x0600, i),
			i,
			fmt.Sprintf("Gallery caption %03d", i),
			fmt.Sprintf("Gallery alt %03d", i),
		).Error; err != nil {
			return fmt.Errorf("post_gallery_items row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostChangelog(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO post_changelog (id, post_id, at, user_id, note)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT (id) DO UPDATE
			 SET post_id = EXCLUDED.post_id,
			     at = EXCLUDED.at,
			     user_id = EXCLUDED.user_id,
			     note = EXCLUDED.note`,
			seedUUID(0x0e00, i),
			seedUUID(0x0700, i),
			time.Now().UTC().Add(-time.Duration(i)*time.Hour),
			userPublicIDFromUUID(tx, seedUUID(0x0100, ((i+2)%seedRows)+1)),
			fmt.Sprintf("Seed changelog entry %03d", i),
		).Error; err != nil {
			return fmt.Errorf("post_changelog row %d: %w", i, err)
		}
	}
	return nil
}

func seedPostSyndication(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO post_syndication (id, post_id, platform, url, status)
			 VALUES (?, ?, ?, ?, ?)
			 ON CONFLICT (id) DO UPDATE
			 SET post_id = EXCLUDED.post_id,
			     platform = EXCLUDED.platform,
			     url = EXCLUDED.url,
			     status = EXCLUDED.status,
			     updated_at = NOW()`,
			seedUUID(0x0f00, i),
			seedUUID(0x0700, i),
			"medium",
			fmt.Sprintf("https://medium.example/seed-post-%03d", i),
			"active",
		).Error; err != nil {
			return fmt.Errorf("post_syndication row %d: %w", i, err)
		}
	}
	return nil
}

func seedTranslationGroups(tx *gorm.DB) error {
	for i := 1; i <= seedRows; i++ {
		if err := tx.Exec(
			`INSERT INTO translation_groups (id)
			 VALUES (?)
			 ON CONFLICT (id) DO NOTHING`,
			seedUUID(0x1000, i),
		).Error; err != nil {
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
		if err := tx.Exec(
			`INSERT INTO post_translations (post_id, group_id, locale)
			 VALUES (?, ?, ?)
			 ON CONFLICT (post_id) DO UPDATE
			 SET group_id = EXCLUDED.group_id, locale = EXCLUDED.locale`,
			seedUUID(0x0700, i),
			seedUUID(0x1000, i),
			locale,
		).Error; err != nil {
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
	_ = tx.Raw(`SELECT public_id FROM users WHERE id = ?`, userUUID).Scan(&id).Error
	return id
}
