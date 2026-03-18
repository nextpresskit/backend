package infrastructure

import (
	"context"
	"time"

	taxDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/taxonomy/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

type gormCategory struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	Name      string    `gorm:"column:name;not null"`
	Slug      string    `gorm:"column:slug;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (gormCategory) TableName() string { return "categories" }

type gormTag struct {
	ID        string    `gorm:"column:id;type:uuid;primaryKey"`
	Name      string    `gorm:"column:name;not null"`
	Slug      string    `gorm:"column:slug;not null;uniqueIndex"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (gormTag) TableName() string { return "tags" }

func (r *GormRepository) CreateCategory(ctx context.Context, c *taxDomain.Category) error {
	row := gormCategory{
		ID:        string(c.ID),
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *GormRepository) ListCategories(ctx context.Context, limit, offset int) ([]taxDomain.Category, error) {
	var rows []gormCategory
	if err := r.db.WithContext(ctx).Order("name ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]taxDomain.Category, 0, len(rows))
	for _, row := range rows {
		out = append(out, taxDomain.Category{
			ID:        taxDomain.CategoryID(row.ID),
			Name:      row.Name,
			Slug:      row.Slug,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *GormRepository) FindCategoryByID(ctx context.Context, id taxDomain.CategoryID) (*taxDomain.Category, error) {
	var row gormCategory
	if err := r.db.WithContext(ctx).Where("id = ?", string(id)).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &taxDomain.Category{
		ID:        taxDomain.CategoryID(row.ID),
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *GormRepository) UpdateCategory(ctx context.Context, c *taxDomain.Category) error {
	row := gormCategory{
		ID:        string(c.ID),
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	return r.db.WithContext(ctx).
		Model(&gormCategory{}).
		Where("id = ?", row.ID).
		Updates(&row).Error
}

func (r *GormRepository) DeleteCategory(ctx context.Context, id taxDomain.CategoryID) error {
	return r.db.WithContext(ctx).Where("id = ?", string(id)).Delete(&gormCategory{}).Error
}

func (r *GormRepository) CreateTag(ctx context.Context, t *taxDomain.Tag) error {
	row := gormTag{
		ID:        string(t.ID),
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *GormRepository) ListTags(ctx context.Context, limit, offset int) ([]taxDomain.Tag, error) {
	var rows []gormTag
	if err := r.db.WithContext(ctx).Order("name ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]taxDomain.Tag, 0, len(rows))
	for _, row := range rows {
		out = append(out, taxDomain.Tag{
			ID:        taxDomain.TagID(row.ID),
			Name:      row.Name,
			Slug:      row.Slug,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *GormRepository) FindTagByID(ctx context.Context, id taxDomain.TagID) (*taxDomain.Tag, error) {
	var row gormTag
	if err := r.db.WithContext(ctx).Where("id = ?", string(id)).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &taxDomain.Tag{
		ID:        taxDomain.TagID(row.ID),
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *GormRepository) UpdateTag(ctx context.Context, t *taxDomain.Tag) error {
	row := gormTag{
		ID:        string(t.ID),
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
	return r.db.WithContext(ctx).
		Model(&gormTag{}).
		Where("id = ?", row.ID).
		Updates(&row).Error
}

func (r *GormRepository) DeleteTag(ctx context.Context, id taxDomain.TagID) error {
	return r.db.WithContext(ctx).Where("id = ?", string(id)).Delete(&gormTag{}).Error
}

// Optional: helpers to support idempotent seeding/upserts later.
func (r *GormRepository) UpsertCategoryBySlug(ctx context.Context, c *taxDomain.Category) error {
	row := gormCategory{ID: string(c.ID), Name: c.Name, Slug: c.Slug, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "updated_at"}),
	}).Create(&row).Error
}

