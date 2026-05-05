package domain

import "context"

type Repository interface {
	Create(ctx context.Context, m *Media) error
	FindByID(ctx context.Context, id MediaID) (*Media, error)
	FindByUUID(ctx context.Context, uuid string) (*Media, error)
	List(ctx context.Context, limit, offset int) ([]Media, error)
}
