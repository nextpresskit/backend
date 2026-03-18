package domain

import "context"

type Repository interface {
	// Menus
	CreateMenu(ctx context.Context, m *Menu) error
	ListMenus(ctx context.Context, limit, offset int) ([]Menu, error)
	FindMenuByID(ctx context.Context, id MenuID) (*Menu, error)
	FindMenuBySlug(ctx context.Context, slug string) (*Menu, error)
	UpdateMenu(ctx context.Context, m *Menu) error
	DeleteMenu(ctx context.Context, id MenuID) error

	// Items
	ListMenuItems(ctx context.Context, menuID MenuID) ([]MenuItem, error)
	ReplaceMenuItems(ctx context.Context, menuID MenuID, items []MenuItem) error
}

