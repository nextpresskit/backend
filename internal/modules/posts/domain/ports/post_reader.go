package ports

import (
	"context"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/model"
)

// PostReader loads posts (admin and public list/read paths).
type PostReader interface {
	FindByID(ctx context.Context, id ident.PostID) (*model.Post, error)
	FindByUUID(ctx context.Context, uuid string) (*model.Post, error)
	FindBySlug(ctx context.Context, slug string) (*model.Post, error)
	List(ctx context.Context, includeDeleted bool, limit int, offset int) ([]model.Post, error)
	ListFiltered(ctx context.Context, includeDeleted bool, limit int, offset int, status string, authorID string, q string) ([]model.Post, error)
	ListPublished(ctx context.Context, limit int, offset int, q string, categoryID string, tagID string) ([]model.Post, error)
	FindPublishedBySlug(ctx context.Context, slug string) (*model.Post, error)
}
