package infrastructure

import (
	"context"
	"time"

	menuDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/menus/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

type gormMenu struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	Name      string    `gorm:"column:name;not null"`
	Slug      string    `gorm:"column:slug;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (gormMenu) TableName() string { return "menus" }

type gormMenuItem struct {
	ID        string     `gorm:"column:id;type:uuid;primaryKey"`
	MenuID    string     `gorm:"column:menu_id;type:uuid;not null;index"`
	ParentID  *string    `gorm:"column:parent_id;type:uuid;index"`
	Label     string     `gorm:"column:label;not null"`
	ItemType  string     `gorm:"column:item_type;not null"`
	RefID     *string    `gorm:"column:ref_id;type:uuid"`
	URL       *string    `gorm:"column:url"`
	SortOrder int        `gorm:"column:sort_order;not null"`
	CreatedAt time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null"`
}

func (gormMenuItem) TableName() string { return "menu_items" }

func (r *GormRepository) CreateMenu(ctx context.Context, m *menuDomain.Menu) error {
	row := gormMenu{
		ID:        string(m.ID),
		Name:      m.Name,
		Slug:      m.Slug,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *GormRepository) ListMenus(ctx context.Context, limit, offset int) ([]menuDomain.Menu, error) {
	var rows []gormMenu
	if err := r.db.WithContext(ctx).Order("name ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]menuDomain.Menu, 0, len(rows))
	for _, row := range rows {
		out = append(out, menuDomain.Menu{
			ID:        menuDomain.MenuID(row.ID),
			Name:      row.Name,
			Slug:      row.Slug,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *GormRepository) FindMenuByID(ctx context.Context, id menuDomain.MenuID) (*menuDomain.Menu, error) {
	var row gormMenu
	if err := r.db.WithContext(ctx).Where("id = ?", string(id)).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &menuDomain.Menu{
		ID:        menuDomain.MenuID(row.ID),
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *GormRepository) FindMenuBySlug(ctx context.Context, slug string) (*menuDomain.Menu, error) {
	var row gormMenu
	if err := r.db.WithContext(ctx).Where("slug = ?", slug).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &menuDomain.Menu{
		ID:        menuDomain.MenuID(row.ID),
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *GormRepository) UpdateMenu(ctx context.Context, m *menuDomain.Menu) error {
	row := gormMenu{
		ID:        string(m.ID),
		Name:      m.Name,
		Slug:      m.Slug,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
	return r.db.WithContext(ctx).
		Model(&gormMenu{}).
		Where("id = ?", row.ID).
		Updates(&row).Error
}

func (r *GormRepository) DeleteMenu(ctx context.Context, id menuDomain.MenuID) error {
	return r.db.WithContext(ctx).Where("id = ?", string(id)).Delete(&gormMenu{}).Error
}

func (r *GormRepository) ListMenuItems(ctx context.Context, menuID menuDomain.MenuID) ([]menuDomain.MenuItem, error) {
	var rows []gormMenuItem
	if err := r.db.WithContext(ctx).
		Where("menu_id = ?", string(menuID)).
		Order("sort_order ASC").
		Order("created_at ASC").
		Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]menuDomain.MenuItem, 0, len(rows))
	for _, row := range rows {
		var parent *menuDomain.MenuItemID
		if row.ParentID != nil {
			p := menuDomain.MenuItemID(*row.ParentID)
			parent = &p
		}
		out = append(out, menuDomain.MenuItem{
			ID:        menuDomain.MenuItemID(row.ID),
			MenuID:    menuDomain.MenuID(row.MenuID),
			ParentID:  parent,
			Label:     row.Label,
			ItemType:  menuDomain.ItemType(row.ItemType),
			RefID:     row.RefID,
			URL:       row.URL,
			SortOrder: row.SortOrder,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *GormRepository) ReplaceMenuItems(ctx context.Context, menuID menuDomain.MenuID, items []menuDomain.MenuItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("menu_id = ?", string(menuID)).Delete(&gormMenuItem{}).Error; err != nil {
			return err
		}
		if len(items) == 0 {
			return nil
		}

		rows := make([]gormMenuItem, 0, len(items))
		for _, it := range items {
			var parent *string
			if it.ParentID != nil {
				p := string(*it.ParentID)
				parent = &p
			}
			rows = append(rows, gormMenuItem{
				ID:        string(it.ID),
				MenuID:    string(it.MenuID),
				ParentID:  parent,
				Label:     it.Label,
				ItemType:  string(it.ItemType),
				RefID:     it.RefID,
				URL:       it.URL,
				SortOrder: it.SortOrder,
				CreatedAt: it.CreatedAt,
				UpdatedAt: it.UpdatedAt,
			})
		}

		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
	})
}

