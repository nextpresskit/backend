package persistence

import (
	"time"

	"gorm.io/gorm"
)

// Page maps to pages (bigint id + public uuid).
type Page struct {
	ID          int64          `gorm:"column:id;primaryKey;autoIncrement"`
	UUID        string         `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	AuthorID    int64          `gorm:"column:author_id;not null;index"`
	Title       string         `gorm:"column:title;not null"`
	Slug        string         `gorm:"column:slug;not null;unique"`
	Content     string         `gorm:"column:content;not null"`
	Status      string         `gorm:"column:status;not null"`
	PublishedAt *time.Time     `gorm:"column:published_at"`
	CreatedAt   time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (Page) TableName() string { return "pages" }
