package persistence

import "time"

// Category maps to categories.
type Category struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	Name      string    `gorm:"column:name;not null"`
	Slug      string    `gorm:"column:slug;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (Category) TableName() string { return "categories" }

// Tag maps to tags.
type Tag struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	Name      string    `gorm:"column:name;not null"`
	Slug      string    `gorm:"column:slug;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (Tag) TableName() string { return "tags" }
