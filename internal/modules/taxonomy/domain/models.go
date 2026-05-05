package domain

import "time"

type CategoryID int64
type TagID int64

type Category struct {
	ID        CategoryID
	UUID      string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Tag struct {
	ID        TagID
	UUID      string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
