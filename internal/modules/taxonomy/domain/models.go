package domain

import "time"

type CategoryID string
type TagID string

type Category struct {
	ID        CategoryID
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Tag struct {
	ID        TagID
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

