package infrastructure

import (
	"context"

	taxDomain "github.com/nextpresskit/backend/internal/modules/taxonomy/domain"
	taxp "github.com/nextpresskit/backend/internal/modules/taxonomy/persistence"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) CreateCategory(ctx context.Context, c *taxDomain.Category) error {
	row := taxp.Category{
		UUID:      c.UUID,
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *GormRepository) ListCategories(ctx context.Context, limit, offset int) ([]taxDomain.Category, error) {
	var rows []taxp.Category
	if err := r.db.WithContext(ctx).Order("name ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]taxDomain.Category, 0, len(rows))
	for _, row := range rows {
		out = append(out, taxDomain.Category{
			ID:        taxDomain.CategoryID(row.ID),
			UUID:      row.UUID,
			Name:      row.Name,
			Slug:      row.Slug,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *GormRepository) FindCategoryByUUID(ctx context.Context, uuid string) (*taxDomain.Category, error) {
	var row taxp.Category
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &taxDomain.Category{
		ID:        taxDomain.CategoryID(row.ID),
		UUID:      row.UUID,
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *GormRepository) FindCategoryByID(ctx context.Context, id taxDomain.CategoryID) (*taxDomain.Category, error) {
	var row taxp.Category
	if err := r.db.WithContext(ctx).Where("id = ?", int64(id)).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &taxDomain.Category{
		ID:        taxDomain.CategoryID(row.ID),
		UUID:      row.UUID,
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *GormRepository) UpdateCategory(ctx context.Context, c *taxDomain.Category) error {
	return r.db.WithContext(ctx).
		Model(&taxp.Category{}).
		Where("uuid = ?", c.UUID).
		Updates(map[string]any{
			"name":       c.Name,
			"slug":       c.Slug,
			"updated_at": c.UpdatedAt,
		}).Error
}

func (r *GormRepository) DeleteCategory(ctx context.Context, uuid string) error {
	return r.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&taxp.Category{}).Error
}

func (r *GormRepository) CreateTag(ctx context.Context, t *taxDomain.Tag) error {
	row := taxp.Tag{
		UUID:      t.UUID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *GormRepository) ListTags(ctx context.Context, limit, offset int) ([]taxDomain.Tag, error) {
	var rows []taxp.Tag
	if err := r.db.WithContext(ctx).Order("name ASC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]taxDomain.Tag, 0, len(rows))
	for _, row := range rows {
		out = append(out, taxDomain.Tag{
			ID:        taxDomain.TagID(row.ID),
			UUID:      row.UUID,
			Name:      row.Name,
			Slug:      row.Slug,
			CreatedAt: row.CreatedAt,
			UpdatedAt: row.UpdatedAt,
		})
	}
	return out, nil
}

func (r *GormRepository) FindTagByUUID(ctx context.Context, uuid string) (*taxDomain.Tag, error) {
	var row taxp.Tag
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &taxDomain.Tag{
		ID:        taxDomain.TagID(row.ID),
		UUID:      row.UUID,
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *GormRepository) FindTagByID(ctx context.Context, id taxDomain.TagID) (*taxDomain.Tag, error) {
	var row taxp.Tag
	if err := r.db.WithContext(ctx).Where("id = ?", int64(id)).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &taxDomain.Tag{
		ID:        taxDomain.TagID(row.ID),
		UUID:      row.UUID,
		Name:      row.Name,
		Slug:      row.Slug,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *GormRepository) UpdateTag(ctx context.Context, t *taxDomain.Tag) error {
	return r.db.WithContext(ctx).
		Model(&taxp.Tag{}).
		Where("uuid = ?", t.UUID).
		Updates(map[string]any{
			"name":       t.Name,
			"slug":       t.Slug,
			"updated_at": t.UpdatedAt,
		}).Error
}

func (r *GormRepository) DeleteTag(ctx context.Context, uuid string) error {
	return r.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&taxp.Tag{}).Error
}

// UpsertCategoryBySlug supports idempotent seeding.
func (r *GormRepository) UpsertCategoryBySlug(ctx context.Context, c *taxDomain.Category) error {
	row := taxp.Category{UUID: c.UUID, Name: c.Name, Slug: c.Slug, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "slug"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "updated_at"}),
	}).Create(&row).Error
}
