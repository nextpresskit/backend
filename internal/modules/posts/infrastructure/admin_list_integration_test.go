//go:build integration

package infrastructure_test

import (
	"context"
	"os"
	"testing"
	"time"

	postsinfra "github.com/nextpresskit/backend/internal/modules/posts/infrastructure"
	postp "github.com/nextpresskit/backend/internal/modules/posts/persistence"
	taxp "github.com/nextpresskit/backend/internal/modules/taxonomy/persistence"
	userp "github.com/nextpresskit/backend/internal/modules/user/persistence"
	"github.com/nextpresskit/backend/internal/platform/database"
)

// TestAdminListHydratesSummaries_Integration checks that the admin list fills author,
// categories (with primary flag) and tags. Runs in a rolled-back transaction.
func TestAdminListHydratesSummaries_Integration(t *testing.T) {
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
	for name, migrate := range map[string]func() error{
		"users":    func() error { return userp.AutoMigrate(tx) },
		"taxonomy": func() error { return taxp.AutoMigrate(tx) },
		"posts":    func() error { return postp.AutoMigrate(tx) },
	} {
		if err := migrate(); err != nil {
			t.Fatalf("migrate %s: %v", name, err)
		}
	}

	now := time.Now().UTC()
	author := userp.User{UUID: "00000000-0000-4000-8000-00000000a001", FirstName: "Ada", LastName: "Lovelace", Email: "ada-it@example.test", Password: "x", Active: true}
	if err := tx.Create(&author).Error; err != nil {
		t.Fatalf("author: %v", err)
	}
	cat := taxp.Category{UUID: "00000000-0000-4000-8000-00000000c001", Name: "IT Category", Slug: "it-category-x", CreatedAt: now, UpdatedAt: now}
	tag := taxp.Tag{UUID: "00000000-0000-4000-8000-00000000d001", Name: "IT Tag", Slug: "it-tag-x", CreatedAt: now, UpdatedAt: now}
	if err := tx.Create(&cat).Error; err != nil {
		t.Fatalf("category: %v", err)
	}
	if err := tx.Create(&tag).Error; err != nil {
		t.Fatalf("tag: %v", err)
	}
	post := postp.Post{
		UUID: "00000000-0000-4000-8000-00000000e001", AuthorID: author.ID, Title: "Integration hydrate", Slug: "integration-hydrate-x",
		Visibility: "public", Locale: "en-US", Timezone: "UTC", Content: "body", Status: "draft", WorkflowStage: "draft", Revision: 1,
		CustomFields: []byte("{}"), Flags: []byte("{}"), Engagement: []byte("{}"), Workflow: []byte("{}"),
		PrimaryCategoryID: &cat.ID, CreatedAt: now.Add(time.Hour), UpdatedAt: now,
	}
	if err := tx.Create(&post).Error; err != nil {
		t.Fatalf("post: %v", err)
	}
	if err := tx.Create(&postp.PostCategory{PostID: post.ID, CategoryID: cat.ID}).Error; err != nil {
		t.Fatalf("post_category: %v", err)
	}
	if err := tx.Create(&postp.PostTag{PostID: post.ID, TagID: tag.ID}).Error; err != nil {
		t.Fatalf("post_tag: %v", err)
	}

	repo := postsinfra.NewGormRepository(tx)
	posts, err := repo.ListFiltered(context.Background(), false, 50, 0, "", "", "Integration hydrate")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("got %d posts, want 1", len(posts))
	}
	p := posts[0]
	if p.Author == nil || p.Author.DisplayName != "Ada Lovelace" {
		t.Fatalf("author = %+v", p.Author)
	}
	if len(p.Categories) != 1 || p.Categories[0].Name != "IT Category" || !p.Categories[0].IsPrimary {
		t.Fatalf("categories = %+v", p.Categories)
	}
	if len(p.Tags) != 1 || p.Tags[0].Slug != "it-tag-x" {
		t.Fatalf("tags = %+v", p.Tags)
	}
}
