package domain

import "context"

type Repository interface {
	Create(ctx context.Context, plugin *Plugin) error
	FindByID(ctx context.Context, id PluginID) (*Plugin, error)
	FindBySlug(ctx context.Context, slug string) (*Plugin, error)
	List(ctx context.Context) ([]Plugin, error)
	ListEnabled(ctx context.Context) ([]Plugin, error)
	Update(ctx context.Context, plugin *Plugin) error
}

