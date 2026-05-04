package extensions

import (
	"context"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
)

// PostGalleryStore manages post_gallery_items.
type PostGalleryStore interface {
	CreateGalleryItem(ctx context.Context, postID ident.PostID, mediaID string, sortOrder int, caption *string, alt *string) (itemID string, err error)
	UpdateGalleryItem(ctx context.Context, postID ident.PostID, itemID string, sortOrder *int, caption *string, alt *string) error
	DeleteGalleryItem(ctx context.Context, postID ident.PostID, itemID string) error
}
