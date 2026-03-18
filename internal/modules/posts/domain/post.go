package domain

import "time"

type PostID string

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
)

type Post struct {
	ID          PostID
	AuthorID    string
	Title       string
	Slug        string
	Content     string
	Status      Status
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

