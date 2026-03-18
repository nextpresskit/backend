package infrastructure

import (
	"context"
	"time"

	pageDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/pages/domain"
	"gorm.io/gorm"
)

type gormPage struct {
	ID          string         `gorm:"column:id;type:uuid;primaryKey"`
	AuthorID    string         `gorm:"column:author_id;type:uuid;not null;index"`
	Title       string         `gorm:"column:title;not null"`
	Slug        string         `gorm:"column:slug;not null;uniqueIndex"`
	Content     string         `gorm:"column:content;not null"`
	Status      string         `gorm:"column:status;not null"`
	PublishedAt *time.Time     `gorm:"column:published_at"`
	CreatedAt   time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;not null"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (gormPage) TableName() string { return "pages" }

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, page *pageDomain.Page) error {
	m := fromDomain(page)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	*page = *toDomain(m)
	return nil
}

func (r *GormRepository) FindByID(ctx context.Context, id pageDomain.PageID) (*pageDomain.Page, error) {
	var row gormPage
	if err := r.db.WithContext(ctx).
		Where("id = ?", string(id)).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(&row), nil
}

func (r *GormRepository) FindBySlug(ctx context.Context, slug string) (*pageDomain.Page, error) {
	var row gormPage
	if err := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(&row), nil
}

func (r *GormRepository) FindPublishedBySlug(ctx context.Context, slug string) (*pageDomain.Page, error) {
	var row gormPage
	if err := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		Where("status = ?", string(pageDomain.StatusPublished)).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(&row), nil
}

func (r *GormRepository) List(ctx context.Context, includeDeleted bool, limit int, offset int) ([]pageDomain.Page, error) {
	return r.ListFiltered(ctx, includeDeleted, limit, offset, "", "", "")
}

func (r *GormRepository) ListFiltered(ctx context.Context, includeDeleted bool, limit int, offset int, status string, authorID string, q string) ([]pageDomain.Page, error) {
	dbq := r.db.WithContext(ctx).Model(&gormPage{}).Order("created_at DESC").Limit(limit).Offset(offset)
	if includeDeleted {
		dbq = dbq.Unscoped()
	}

	if status != "" {
		dbq = dbq.Where("status = ?", status)
	}
	if authorID != "" {
		dbq = dbq.Where("author_id = ?", authorID)
	}
	if q != "" {
		like := "%" + q + "%"
		dbq = dbq.Where("(title ILIKE ? OR content ILIKE ?)", like, like)
	}

	var rows []gormPage
	if err := dbq.Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]pageDomain.Page, 0, len(rows))
	for i := range rows {
		p := toDomain(&rows[i])
		out = append(out, *p)
	}
	return out, nil
}

func (r *GormRepository) Update(ctx context.Context, page *pageDomain.Page) error {
	m := fromDomain(page)
	return r.db.WithContext(ctx).
		Model(&gormPage{}).
		Where("id = ?", m.ID).
		Updates(m).Error
}

func (r *GormRepository) Delete(ctx context.Context, id pageDomain.PageID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", string(id)).
		Delete(&gormPage{}).Error
}

func toDomain(m *gormPage) *pageDomain.Page {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deletedAt = &t
	}
	return &pageDomain.Page{
		ID:          pageDomain.PageID(m.ID),
		AuthorID:    m.AuthorID,
		Title:       m.Title,
		Slug:        m.Slug,
		Content:     m.Content,
		Status:      pageDomain.Status(m.Status),
		PublishedAt: m.PublishedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}

func fromDomain(p *pageDomain.Page) *gormPage {
	var deleted gorm.DeletedAt
	if p.DeletedAt != nil {
		deleted = gorm.DeletedAt{Time: *p.DeletedAt, Valid: true}
	}
	return &gormPage{
		ID:          string(p.ID),
		AuthorID:    p.AuthorID,
		Title:       p.Title,
		Slug:        p.Slug,
		Content:     p.Content,
		Status:      string(p.Status),
		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
		DeletedAt:   deleted,
	}
}

