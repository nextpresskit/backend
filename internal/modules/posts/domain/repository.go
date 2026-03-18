package domain

import "context"

type Repository interface {
	Create(ctx context.Context, post *Post) error
	FindByID(ctx context.Context, id PostID) (*Post, error)
	FindBySlug(ctx context.Context, slug string) (*Post, error)
	List(ctx context.Context, includeDeleted bool, limit int, offset int) ([]Post, error)
	ListFiltered(ctx context.Context, includeDeleted bool, limit int, offset int, status string, authorID string, q string) ([]Post, error)
	ListPublished(ctx context.Context, limit int, offset int, q string, categoryID string, tagID string) ([]Post, error)
	FindPublishedBySlug(ctx context.Context, slug string) (*Post, error)
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id PostID) error

	// Taxonomy assignments
	SetCategories(ctx context.Context, postID PostID, categoryIDs []string) error
	SetTags(ctx context.Context, postID PostID, tagIDs []string) error
}

