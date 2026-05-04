package extensions

import (
	"context"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
)

// PostTranslationsStore manages post_translations and translation group linkage.
type PostTranslationsStore interface {
	PutPostTranslation(ctx context.Context, postID ident.PostID, groupID *string, locale string) (resolvedGroupID string, err error)
	ClearPostTranslation(ctx context.Context, postID ident.PostID) error
}
