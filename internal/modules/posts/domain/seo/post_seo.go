package seo

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
)

// PostSEO is persisted SEO metadata for a post.
type PostSEO struct {
	Title          *string
	Description    *string
	CanonicalURL   *string
	Robots         *string
	OGType         *string
	OGImageURL     *string
	TwitterCard    *string
	StructuredData json.RawMessage
	UpdatedAt      time.Time
}

// PostSEOStore manages post_seo rows.
type PostSEOStore interface {
	DeleteSEO(ctx context.Context, postID ident.PostID) error
	UpsertSEOOnly(ctx context.Context, postID ident.PostID, seo *PostSEO) error
}
