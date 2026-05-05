package domain

import "time"

type PageID int64

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
)

type Page struct {
	ID          PageID
	UUID        string
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
