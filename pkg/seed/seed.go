// Package seed runs database seeders (reference data and optional dev data).
package seed

import (
	"fmt"
	"log"

	"gorm.io/gorm"
)

// Run runs all seeders in the correct order.
// Seeders must be idempotent: safe to run multiple times.
func Run(db *gorm.DB) error {
	for _, s := range []struct {
		name string
		fn   func(*gorm.DB) error
	}{
		{"rbac_defaults", SeedRBACDefaults},
	} {
		log.Printf("seeding %s...", s.name)
		if err := s.fn(db); err != nil {
			return fmt.Errorf("seed %s: %w", s.name, err)
		}
		log.Printf("seeded %s ok", s.name)
	}
	return nil
}

