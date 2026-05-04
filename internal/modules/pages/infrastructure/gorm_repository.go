package infrastructure

import (
	"context"
	"strconv"
	"time"

	pageDomain "github.com/nextpresskit/backend/internal/modules/pages/domain"
	pagep "github.com/nextpresskit/backend/internal/modules/pages/persistence"
	"gorm.io/gorm"
)

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
	var row pagep.Page
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
	var row pagep.Page
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
	var row pagep.Page
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
	dbq := r.db.WithContext(ctx).Model(&pagep.Page{}).Order("created_at DESC").Limit(limit).Offset(offset)
	if includeDeleted {
		dbq = dbq.Unscoped()
	}

	if status != "" {
		dbq = dbq.Where("status = ?", status)
	}
	if authorID != "" {
		if aid, err := strconv.ParseInt(authorID, 10, 64); err == nil {
			dbq = dbq.Where("author_id = ?", aid)
		}
	}
	if q != "" {
		like := "%" + q + "%"
		dbq = dbq.Where("(title ILIKE ? OR content ILIKE ?)", like, like)
	}

	var rows []pagep.Page
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
		Model(&pagep.Page{}).
		Where("id = ?", m.ID).
		Updates(m).Error
}

func (r *GormRepository) Delete(ctx context.Context, id pageDomain.PageID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", string(id)).
		Delete(&pagep.Page{}).Error
}

func toDomain(m *pagep.Page) *pageDomain.Page {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deletedAt = &t
	}
	return &pageDomain.Page{
		ID:          pageDomain.PageID(m.ID),
		AuthorID:    strconv.FormatInt(m.AuthorID, 10),
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

func fromDomain(p *pageDomain.Page) *pagep.Page {
	var deleted gorm.DeletedAt
	if p.DeletedAt != nil {
		deleted = gorm.DeletedAt{Time: *p.DeletedAt, Valid: true}
	}
	return &pagep.Page{
		ID:          string(p.ID),
		AuthorID:    parseInt64OrZero(p.AuthorID),
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

func parseInt64OrZero(v string) int64 {
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}

