package persistence

import (
	"time"

	"gorm.io/gorm"
)

// Page maps to pages.
type Page struct {
	ID          string         `gorm:"column:id;type:uuid;primaryKey"`
	AuthorID    int64          `gorm:"column:author_id;not null;index"`
	Title       string         `gorm:"column:title;not null"`
	Slug        string         `gorm:"column:slug;not null;uniqueIndex"`
	Content     string         `gorm:"column:content;not null"`
	Status      string         `gorm:"column:status;not null"`
	PublishedAt *time.Time     `gorm:"column:published_at"`
	CreatedAt   time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Page) TableName() string { return "pages" }
