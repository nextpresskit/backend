package persistence

import (
	"fmt"
	"time"

	"github.com/nextpresskit/backend/pkg/seed/helpers"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const demoSeedRows = 100

const superadminUserUUID = "00000000-0000-0000-0100-000000000001"

// SeedDemo inserts deterministic demo users (including superadmin).
func SeedDemo(tx *gorm.DB) error {
	superadminEmail := helpers.EnvOrDefault("SEED_SUPERADMIN_EMAIL", "superadmin@nextpresskit.local")
	superadminPassword := helpers.EnvOrDefault("SEED_SUPERADMIN_PASSWORD", "SuperAdmin123!")
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(superadminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash superadmin password: %w", err)
	}
	return seedUsers(tx, superadminEmail, string(passwordHash))
}

func seedUsers(tx *gorm.DB, superadminEmail, passwordHash string) error {
	for i := 1; i <= demoSeedRows; i++ {
		id := helpers.SeedUUID(0x0100, i)
		firstName := fmt.Sprintf("User%03d", i)
		lastName := "Seed"
		email := fmt.Sprintf("user%03d@nextpresskit.local", i)
		if i == 1 {
			id = superadminUserUUID
			firstName = "Super"
			lastName = "Admin"
			email = superadminEmail
		}
		u := User{
			UUID:      id,
			FirstName: firstName,
			LastName:  lastName,
			Email:     email,
			Password:  passwordHash,
			Active:    true,
		}
		now := time.Now().UTC()
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "uuid"}},
			DoUpdates: clause.Assignments(map[string]any{
				"email":      email,
				"first_name": firstName,
				"last_name":  lastName,
				"password":   passwordHash,
				"active":     true,
				"deleted_at": nil,
				"updated_at": now,
			}),
		}).Create(&u).Error; err != nil {
			return fmt.Errorf("users row %d: %w", i, err)
		}
	}
	return nil
}
