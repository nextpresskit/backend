package domain

import "time"

type MenuID string
type MenuItemID string

type Menu struct {
	ID        MenuID
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ItemType string

const (
	ItemTypeURL  ItemType = "url"
	ItemTypePage ItemType = "page"
	ItemTypePost ItemType = "post"
)

type MenuItem struct {
	ID        MenuItemID
	MenuID    MenuID
	ParentID  *MenuItemID
	Label     string
	ItemType  ItemType
	RefID     *string
	URL       *string
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}

