package model

import (
	"encoding/json"
	"time"

	"github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain/ident"
	"github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain/metrics"
	"github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain/seo"
)

// Post is the aggregate root for posts module read/write paths.
type Post struct {
	ID            ident.PostID
	UUID          *string
	AuthorID      string
	Title         string
	Slug          string
	Subtitle      string
	Excerpt       string
	PostType      string
	Format        string
	Visibility    string
	Locale        string
	Timezone      string
	Content       string // markdown
	Status        ident.Status
	WorkflowStage string
	Revision      int

	ReviewerUserID     *string
	LastEditedByUserID *string
	EditorUserIDs      []string

	Author       *UserSummary
	Reviewer     *UserSummary
	LastEditedBy *UserSummary
	Editors      []UserSummary

	ScheduledPublishAt *time.Time
	PublishedAt        *time.Time
	FirstIndexedAt     *time.Time

	CustomFields json.RawMessage
	Flags        json.RawMessage
	Engagement   json.RawMessage
	Workflow     json.RawMessage

	FeaturedMediaID        *string
	FeaturedMediaPublicURL *string
	FeaturedAlt            *string
	FeaturedWidth          *int
	FeaturedHeight         *int
	FeaturedFocalX         *float32
	FeaturedFocalY         *float32
	FeaturedCredit         *string
	FeaturedLicense        *string

	SEO     *seo.PostSEO
	Metrics *metrics.PostMetrics

	Categories   []PostCategory
	Tags         []PostTag
	Series       *PostSeries
	CoAuthors    []UserSummary
	Gallery      []PostGalleryItem
	Syndication  []PostSyndication
	Changelog    []PostChangelogEntry
	Translations *PostTranslations

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// UserSummary is a minimal user projection on posts.
type UserSummary struct {
	ID          string  `json:"id"`
	UUID        string  `json:"uuid"`
	DisplayName string  `json:"displayName"`
	Email       *string `json:"email,omitempty"`
	Role        *string `json:"role,omitempty"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
}

type PostCategory struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	IsPrimary bool   `json:"isPrimary"`
}

type PostTag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type PostSeries struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Slug      string  `json:"slug"`
	PartIndex *int    `json:"partIndex,omitempty"`
	PartLabel *string `json:"partLabel,omitempty"`
}

type PostGalleryItem struct {
	ID        string  `json:"id"`
	MediaID   string  `json:"mediaId"`
	URL       *string `json:"url,omitempty"`
	SortOrder int     `json:"sortOrder"`
	Caption   *string `json:"caption,omitempty"`
	Alt       *string `json:"alt,omitempty"`
}

type PostSyndication struct {
	ID       string `json:"id"`
	Platform string `json:"platform"`
	URL      string `json:"url"`
	Status   string `json:"status"`
}

type PostChangelogEntry struct {
	ID   string       `json:"id"`
	At   time.Time    `json:"at"`
	User *UserSummary `json:"user,omitempty"`
	Note string       `json:"note"`
}

type PostTranslations struct {
	GroupID      *string                `json:"translationGroupId,omitempty"`
	Translations []PostTranslationEntry `json:"translations"`
}

type PostTranslationEntry struct {
	PostID string `json:"postId"`
	Locale string `json:"locale"`
	Slug   string `json:"slug"`
}
