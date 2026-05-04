package extensions

import (
	"context"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
)

// PostChangelogStore manages post_changelog.
type PostChangelogStore interface {
	CreateChangelog(ctx context.Context, postID ident.PostID, userID *string, note string) (changelogID string, err error)
	DeleteChangelog(ctx context.Context, postID ident.PostID, changelogID string) error
}
