package dbmigrate

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/nextpresskit/backend/internal/kit"
)

// AutoMigrateAll runs module AutoMigrate hooks in invocation order (caller must pass FK-safe order).
func AutoMigrateAll(db *gorm.DB, modules []kit.Module) error {
	for _, m := range modules {
		if err := m.AutoMigrate(db); err != nil {
			return fmt.Errorf("%s: %w", m.ID(), err)
		}
	}
	return nil
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
