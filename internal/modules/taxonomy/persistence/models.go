package persistence

import "time"

// Category maps to categories (bigint id + public uuid).
type Category struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UUID      string    `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	Name      string    `gorm:"column:name;not null"`
	Slug      string    `gorm:"column:slug;not null;unique"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (Category) TableName() string { return "categories" }

// Tag maps to tags.
type Tag struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UUID      string    `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	Name      string    `gorm:"column:name;not null"`
	Slug      string    `gorm:"column:slug;not null;unique"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (Tag) TableName() string { return "tags" }
