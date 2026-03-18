package domain

import "context"

type Repository interface {
	Create(ctx context.Context, page *Page) error
	FindByID(ctx context.Context, id PageID) (*Page, error)
	FindBySlug(ctx context.Context, slug string) (*Page, error)
	List(ctx context.Context, includeDeleted bool, limit int, offset int) ([]Page, error)
	ListFiltered(ctx context.Context, includeDeleted bool, limit int, offset int, status string, authorID string, q string) ([]Page, error)
	FindPublishedBySlug(ctx context.Context, slug string) (*Page, error)
	Update(ctx context.Context, page *Page) error
	Delete(ctx context.Context, id PageID) error
}

