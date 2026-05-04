package dbmigrate

import (
	"fmt"

	"gorm.io/gorm"

	mediap "github.com/nextpresskit/backend/internal/modules/media/persistence"
	pagesp "github.com/nextpresskit/backend/internal/modules/pages/persistence"
	pluginsp "github.com/nextpresskit/backend/internal/modules/plugins/persistence"
	postsp "github.com/nextpresskit/backend/internal/modules/posts/persistence"
	rbacp "github.com/nextpresskit/backend/internal/modules/rbac/persistence"
	taxp "github.com/nextpresskit/backend/internal/modules/taxonomy/persistence"
	userp "github.com/nextpresskit/backend/internal/modules/user/persistence"
)

// AutoMigrate runs module persistence AutoMigrate hooks in FK-safe order.
func AutoMigrate(db *gorm.DB) error {
	steps := []struct {
		name string
		fn   func(*gorm.DB) error
	}{
		{"user", userp.AutoMigrate},
		{"rbac", rbacp.AutoMigrate},
		{"taxonomy", taxp.AutoMigrate},
		{"media", mediap.AutoMigrate},
		{"posts", postsp.AutoMigrate},
		{"pages", pagesp.AutoMigrate},
		{"plugins", pluginsp.AutoMigrate},
	}
	for _, s := range steps {
		if err := s.fn(db); err != nil {
			return fmt.Errorf("%s: %w", s.name, err)
		}
	}
	ensureUserPublicIDDefault(db)
	return nil
}

func ensureUserPublicIDDefault(db *gorm.DB) {
	if err := db.Exec(`CREATE SEQUENCE IF NOT EXISTS users_public_id_seq`).Error; err != nil {
		return
	}
	_ = db.Exec(`SELECT setval('users_public_id_seq', GREATEST(COALESCE((SELECT MAX(public_id) FROM users), 0), 1))`).Error
	_ = db.Exec(`ALTER TABLE users ALTER COLUMN public_id SET DEFAULT nextval('users_public_id_seq'::regclass)`).Error
}

// DropPublicSchema drops every table in public (dev reset). Destructive.
func DropPublicSchema(db *gorm.DB) error {
	return db.Exec(`
		DO $$ DECLARE r RECORD;
		BEGIN
		  FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public')
		  LOOP
		    EXECUTE 'DROP TABLE IF EXISTS public.' || quote_ident(r.tablename) || ' CASCADE';
		  END LOOP;
		END $$;
	`).Error
}
