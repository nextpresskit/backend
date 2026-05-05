package ident

// PostID is the internal database primary key for posts (not exposed in public APIs).
type PostID int64

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
)
