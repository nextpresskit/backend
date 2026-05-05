package domain

import "context"

type Repository interface {
	// Categories
	CreateCategory(ctx context.Context, c *Category) error
	ListCategories(ctx context.Context, limit, offset int) ([]Category, error)
	FindCategoryByID(ctx context.Context, id CategoryID) (*Category, error)
	FindCategoryByUUID(ctx context.Context, uuid string) (*Category, error)
	UpdateCategory(ctx context.Context, c *Category) error
	DeleteCategory(ctx context.Context, uuid string) error

	// Tags
	CreateTag(ctx context.Context, t *Tag) error
	ListTags(ctx context.Context, limit, offset int) ([]Tag, error)
	FindTagByID(ctx context.Context, id TagID) (*Tag, error)
	FindTagByUUID(ctx context.Context, uuid string) (*Tag, error)
	UpdateTag(ctx context.Context, t *Tag) error
	DeleteTag(ctx context.Context, uuid string) error
}
