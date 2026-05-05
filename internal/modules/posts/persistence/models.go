package persistence

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// Post maps to posts (author and editor IDs are users.id).
type Post struct {
	ID                   int64           `gorm:"column:id;primaryKey;autoIncrement"`
	UUID                 string          `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	AuthorID             int64           `gorm:"column:author_id;not null;index"`
	Title                string          `gorm:"column:title;not null"`
	Slug                 string          `gorm:"column:slug;not null;unique"`
	Subtitle             string          `gorm:"column:subtitle"`
	Excerpt              string          `gorm:"column:excerpt"`
	PostType             string          `gorm:"column:post_type"`
	Format               string          `gorm:"column:format"`
	Visibility           string          `gorm:"column:visibility;not null"`
	Locale               string          `gorm:"column:locale;not null"`
	Timezone             string          `gorm:"column:timezone;not null"`
	Content              string          `gorm:"column:content;not null"`
	Status               string          `gorm:"column:status;not null"`
	WorkflowStage        string          `gorm:"column:workflow_stage;not null"`
	Revision             int             `gorm:"column:revision;not null"`
	ReviewerUserID       *int64          `gorm:"column:reviewer_user_id"`
	LastEditedByUserID   *int64          `gorm:"column:last_edited_by_user_id"`
	ScheduledPublishAt   *time.Time      `gorm:"column:scheduled_publish_at"`
	PublishedAt          *time.Time      `gorm:"column:published_at"`
	FirstIndexedAt       *time.Time      `gorm:"column:first_indexed_at"`
	CustomFields         json.RawMessage `gorm:"column:custom_fields;type:jsonb;not null"`
	Flags                json.RawMessage `gorm:"column:flags;type:jsonb;not null"`
	Engagement           json.RawMessage `gorm:"column:engagement;type:jsonb;not null"`
	Workflow             json.RawMessage `gorm:"column:workflow;type:jsonb;not null"`
	FeaturedMediaID      *int64          `gorm:"column:featured_media_id"`
	FeaturedAlt          *string         `gorm:"column:featured_alt"`
	FeaturedWidth        *int            `gorm:"column:featured_width"`
	FeaturedHeight       *int            `gorm:"column:featured_height"`
	FeaturedFocalX       *float32        `gorm:"column:featured_focal_x"`
	FeaturedFocalY       *float32        `gorm:"column:featured_focal_y"`
	FeaturedCredit       *string         `gorm:"column:featured_credit"`
	FeaturedLicense      *string         `gorm:"column:featured_license"`
	PrimaryCategoryID    *int64          `gorm:"column:primary_category_id"`
	CreatedAt            time.Time       `gorm:"column:created_at;not null"`
	UpdatedAt            time.Time       `gorm:"column:updated_at;not null"`
	DeletedAt            gorm.DeletedAt  `gorm:"column:deleted_at;index"`
}

func (Post) TableName() string { return "posts" }

// PostCategory maps to post_categories.
type PostCategory struct {
	PostID     int64 `gorm:"column:post_id;primaryKey"`
	CategoryID int64 `gorm:"column:category_id;primaryKey"`
}

func (PostCategory) TableName() string { return "post_categories" }

// PostTag maps to post_tags.
type PostTag struct {
	PostID int64 `gorm:"column:post_id;primaryKey"`
	TagID  int64 `gorm:"column:tag_id;primaryKey"`
}

func (PostTag) TableName() string { return "post_tags" }

// PostSEO maps to post_seo.
type PostSEO struct {
	PostID         int64           `gorm:"column:post_id;primaryKey"`
	Title          *string         `gorm:"column:title"`
	Description    *string         `gorm:"column:description"`
	CanonicalURL   *string         `gorm:"column:canonical_url"`
	Robots         *string         `gorm:"column:robots"`
	OGType         *string         `gorm:"column:og_type"`
	OGImageURL     *string         `gorm:"column:og_image_url"`
	TwitterCard    *string         `gorm:"column:twitter_card"`
	StructuredData json.RawMessage `gorm:"column:structured_data;type:jsonb;not null"`
	UpdatedAt      time.Time       `gorm:"column:updated_at;not null"`
}

func (PostSEO) TableName() string { return "post_seo" }

// PostMetrics maps to post_metrics.
type PostMetrics struct {
	PostID                int64     `gorm:"column:post_id;primaryKey"`
	WordCount             int       `gorm:"column:word_count;not null"`
	CharacterCount        int       `gorm:"column:character_count;not null"`
	ReadingTimeMinutes    int       `gorm:"column:reading_time_minutes;not null"`
	EstReadTimeSeconds    int       `gorm:"column:est_read_time_seconds;not null"`
	ViewCount             int64     `gorm:"column:view_count;not null"`
	UniqueVisitors7d      int64     `gorm:"column:unique_visitors_7d;not null"`
	ScrollDepthAvgPercent float32   `gorm:"column:scroll_depth_avg_percent;not null"`
	BounceRatePercent     float32   `gorm:"column:bounce_rate_percent;not null"`
	AvgTimeOnPageSeconds  int       `gorm:"column:avg_time_on_page_seconds;not null"`
	CommentCount          int       `gorm:"column:comment_count;not null"`
	LikeCount             int       `gorm:"column:like_count;not null"`
	ShareCount            int       `gorm:"column:share_count;not null"`
	BookmarkCount         int       `gorm:"column:bookmark_count;not null"`
	UpdatedAt             time.Time `gorm:"column:updated_at;not null"`
}

func (PostMetrics) TableName() string { return "post_metrics" }

// Series maps to series.
type Series struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UUID      string    `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	Title     string    `gorm:"column:title;not null"`
	Slug      string    `gorm:"column:slug;not null;unique"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (Series) TableName() string { return "series" }

// PostSeries maps to post_series.
type PostSeries struct {
	PostID    int64   `gorm:"column:post_id;primaryKey"`
	SeriesID  int64   `gorm:"column:series_id;primaryKey"`
	PartIndex *int    `gorm:"column:part_index"`
	PartLabel *string `gorm:"column:part_label"`
}

func (PostSeries) TableName() string { return "post_series" }

// PostCoauthor maps to post_coauthors (user_id is users.id).
type PostCoauthor struct {
	PostID    int64 `gorm:"column:post_id;primaryKey"`
	UserID    int64 `gorm:"column:user_id;primaryKey"`
	SortOrder int   `gorm:"column:sort_order;not null"`
}

func (PostCoauthor) TableName() string { return "post_coauthors" }

// PostGalleryItem maps to post_gallery_items.
type PostGalleryItem struct {
	ID        int64   `gorm:"column:id;primaryKey;autoIncrement"`
	UUID      string  `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	PostID    int64   `gorm:"column:post_id;not null;index"`
	MediaID   int64   `gorm:"column:media_id;not null"`
	SortOrder int     `gorm:"column:sort_order;not null"`
	Caption   *string `gorm:"column:caption"`
	Alt       *string `gorm:"column:alt"`
}

func (PostGalleryItem) TableName() string { return "post_gallery_items" }

// PostChangelog maps to post_changelog (user_id is users.id when set).
type PostChangelog struct {
	ID     int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UUID   string    `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	PostID int64     `gorm:"column:post_id;not null;index"`
	At     time.Time `gorm:"column:at;not null"`
	UserID *int64    `gorm:"column:user_id"`
	Note   string    `gorm:"column:note;not null"`
}

func (PostChangelog) TableName() string { return "post_changelog" }

// PostSyndication maps to post_syndication.
type PostSyndication struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UUID      string    `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	PostID    int64     `gorm:"column:post_id;not null;index"`
	Platform  string    `gorm:"column:platform;not null"`
	URL       string    `gorm:"column:url;not null"`
	Status    string    `gorm:"column:status;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (PostSyndication) TableName() string { return "post_syndication" }

// TranslationGroup maps to translation_groups.
type TranslationGroup struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	UUID      string    `gorm:"column:uuid;type:uuid;uniqueIndex;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

func (TranslationGroup) TableName() string { return "translation_groups" }

// PostTranslation maps to post_translations.
type PostTranslation struct {
	PostID  int64  `gorm:"column:post_id;primaryKey"`
	GroupID int64  `gorm:"column:group_id;not null;index"`
	Locale  string `gorm:"column:locale;not null"`
}

func (PostTranslation) TableName() string { return "post_translations" }
