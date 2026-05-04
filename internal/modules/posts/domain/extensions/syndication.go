package extensions

import (
	"context"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
)

// PostSyndicationStore manages post_syndication rows.
type PostSyndicationStore interface {
	CreateSyndication(ctx context.Context, postID ident.PostID, platform, url, status string) (id string, err error)
	UpdateSyndication(ctx context.Context, postID ident.PostID, id string, platform, url, status *string) error
	DeleteSyndication(ctx context.Context, postID ident.PostID, id string) error
	UpdateSyndicationByID(ctx context.Context, id string, platform, url, status *string) error
	DeleteSyndicationByID(ctx context.Context, id string) error
}
