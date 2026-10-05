//go:build integration

package infrastructure_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	rbacinfra "github.com/nextpresskit/backend/internal/modules/rbac/infrastructure"
	rbacp "github.com/nextpresskit/backend/internal/modules/rbac/persistence"
	userApp "github.com/nextpresskit/backend/internal/modules/user/application"
	userinfra "github.com/nextpresskit/backend/internal/modules/user/infrastructure"
	userp "github.com/nextpresskit/backend/internal/modules/user/persistence"
	"github.com/nextpresskit/backend/internal/platform/database"
)

// TestAdminUserListing_Integration runs the real GORM queries against Postgres inside a
// transaction that is always rolled back, so it never leaves rows behind.
func TestAdminUserListing_Integration(t *testing.T) {
	cfg := database.Config{
		Driver:   "postgres",
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}
	if cfg.Host == "" || cfg.Port == "" || cfg.User == "" || cfg.Name == "" {
		t.Skip("integration database env vars not set")
	}
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}
	db, err := database.New(cfg)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	tx := db.Begin()
	defer tx.Rollback()

	if err := userp.AutoMigrate(tx); err != nil {
		t.Fatalf("migrate users: %v", err)
	}
	if err := rbacp.AutoMigrate(tx); err != nil {
		t.Fatalf("migrate rbac: %v", err)
	}
	// Isolate from any rows already in the database.
	if err := tx.Exec("DELETE FROM user_roles").Error; err != nil {
		t.Fatalf("clear user_roles: %v", err)
	}
	if err := tx.Exec("DELETE FROM users").Error; err != nil {
		t.Fatalf("clear users: %v", err)
	}

	names := [][2]string{{"Ada", "Lovelace"}, {"Grace", "Hopper"}, {"Alan", "Turing"}, {"Under_Score", "Person"}}
	ids := make([]int64, 0, len(names))
	for i, n := range names {
		u := userp.User{
			UUID:      fmt.Sprintf("00000000-0000-4000-8000-%012d", i+1),
			FirstName: n[0],
			LastName:  n[1],
			Email:     fmt.Sprintf("%s@example.test", n[0]),
			Password:  "x",
			Active:    true,
		}
		if err := tx.Create(&u).Error; err != nil {
			t.Fatalf("insert user: %v", err)
		}
		ids = append(ids, u.ID)
	}
	if err := tx.Delete(&userp.User{}, ids[2]).Error; err != nil { // soft-delete Alan
		t.Fatalf("soft delete: %v", err)
	}
	role := rbacp.Role{UUID: "00000000-0000-4000-8000-0000000000aa", Name: "editor"}
	if err := tx.Create(&role).Error; err != nil {
		t.Fatalf("insert role: %v", err)
	}
	if err := tx.Create(&rbacp.UserRole{UserID: ids[0], RoleID: role.ID}).Error; err != nil {
		t.Fatalf("insert user_role: %v", err)
	}

	repo := userinfra.NewGormRepository(tx)
	svc := userApp.NewService(repo, repo, rbacinfra.NewGormRepository(tx))
	ctx := context.Background()

	tests := []struct {
		name       string
		query      string
		limit      int
		offset     int
		wantTotal  int64
		wantEmails []string
	}{
		{name: "all non-deleted", limit: 10, wantTotal: 3, wantEmails: []string{"Ada@example.test", "Grace@example.test", "Under_Score@example.test"}},
		{name: "page 2 of size 2", limit: 2, offset: 2, wantTotal: 3, wantEmails: []string{"Under_Score@example.test"}},
		{name: "search by last name, case-insensitive", query: "HOPPER", limit: 10, wantTotal: 1, wantEmails: []string{"Grace@example.test"}},
		{name: "search by full name", query: "ada love", limit: 10, wantTotal: 1, wantEmails: []string{"Ada@example.test"}},
		{name: "underscore is literal, not a wildcard", query: "_", limit: 10, wantTotal: 1, wantEmails: []string{"Under_Score@example.test"}},
		{name: "soft-deleted user hidden", query: "turing", limit: 10, wantTotal: 0, wantEmails: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := svc.List(ctx, tt.query, tt.limit, tt.offset)
			if err != nil {
				t.Fatalf("list: %v", err)
			}
			if res.Total != tt.wantTotal {
				t.Fatalf("total = %d, want %d", res.Total, tt.wantTotal)
			}
			got := make([]string, 0, len(res.Users))
			for _, u := range res.Users {
				got = append(got, u.User.Email)
			}
			if fmt.Sprint(got) != fmt.Sprint(tt.wantEmails) {
				t.Fatalf("emails = %v, want %v", got, tt.wantEmails)
			}
		})
	}

	got, err := svc.Get(ctx, fmt.Sprint(ids[0]))
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(got.Roles) != 1 || got.Roles[0] != "editor" {
		t.Fatalf("roles = %v, want [editor]", got.Roles)
	}
	if _, err := svc.Get(ctx, fmt.Sprint(ids[2])); err != userApp.ErrUserNotFound {
		t.Fatalf("soft-deleted get err = %v, want ErrUserNotFound", err)
	}
}
