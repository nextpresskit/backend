package ports

import (
	"context"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/model"
)

// PostLoadUpdater is the minimal persistence surface for PostSave hooks that reload and patch a post.
type PostLoadUpdater interface {
	FindByID(ctx context.Context, id ident.PostID) (*model.Post, error)
	FindByUUID(ctx context.Context, uuid string) (*model.Post, error)
	Update(ctx context.Context, post *model.Post) error
}
