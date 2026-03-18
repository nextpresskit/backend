package infrastructure

import (
	"context"
	"time"

	postDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type gormPost struct {
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

func (gormPost) TableName() string { return "posts" }

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, post *postDomain.Post) error {
	m := fromDomain(post)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	*post = *toDomain(m)
	return nil
}

func (r *GormRepository) FindByID(ctx context.Context, id postDomain.PostID) (*postDomain.Post, error) {
	var row gormPost
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

func (r *GormRepository) FindBySlug(ctx context.Context, slug string) (*postDomain.Post, error) {
	var row gormPost
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

func (r *GormRepository) List(ctx context.Context, includeDeleted bool, limit int, offset int) ([]postDomain.Post, error) {
	return r.ListFiltered(ctx, includeDeleted, limit, offset, "", "", "")
}

func (r *GormRepository) ListFiltered(ctx context.Context, includeDeleted bool, limit int, offset int, status string, authorID string, q string) ([]postDomain.Post, error) {
	dbq := r.db.WithContext(ctx).Model(&gormPost{}).Order("created_at DESC").Limit(limit).Offset(offset)
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

	var rows []gormPost
	if err := dbq.Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]postDomain.Post, 0, len(rows))
	for i := range rows {
		p := toDomain(&rows[i])
		out = append(out, *p)
	}
	return out, nil
}

func (r *GormRepository) ListPublished(ctx context.Context, limit int, offset int, q string, categoryID string, tagID string) ([]postDomain.Post, error) {
	dbq := r.db.WithContext(ctx).
		Model(&gormPost{}).
		Where("status = ?", string(postDomain.StatusPublished)).
		Order("published_at DESC").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if q != "" {
		like := "%" + q + "%"
		dbq = dbq.Where("(title ILIKE ? OR content ILIKE ?)", like, like)
	}
	if categoryID != "" {
		dbq = dbq.Joins("JOIN post_categories pc ON pc.post_id = posts.id").
			Where("pc.category_id = ?", categoryID)
	}
	if tagID != "" {
		dbq = dbq.Joins("JOIN post_tags pt ON pt.post_id = posts.id").
			Where("pt.tag_id = ?", tagID)
	}

	var rows []gormPost
	if err := dbq.Find(&rows).Error; err != nil {
		return nil, err
	}

	out := make([]postDomain.Post, 0, len(rows))
	for i := range rows {
		p := toDomain(&rows[i])
		out = append(out, *p)
	}
	return out, nil
}

func (r *GormRepository) FindPublishedBySlug(ctx context.Context, slug string) (*postDomain.Post, error) {
	var row gormPost
	if err := r.db.WithContext(ctx).
		Where("slug = ?", slug).
		Where("status = ?", string(postDomain.StatusPublished)).
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(&row), nil
}

func (r *GormRepository) Update(ctx context.Context, post *postDomain.Post) error {
	m := fromDomain(post)
	if err := r.db.WithContext(ctx).
		Model(&gormPost{}).
		Where("id = ?", m.ID).
		Updates(m).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormRepository) Delete(ctx context.Context, id postDomain.PostID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", string(id)).
		Delete(&gormPost{}).Error
}

type gormPostCategory struct {
	PostID     string `gorm:"column:post_id;type:uuid;primaryKey"`
	CategoryID string `gorm:"column:category_id;type:uuid;primaryKey"`
}

func (gormPostCategory) TableName() string { return "post_categories" }

type gormPostTag struct {
	PostID string `gorm:"column:post_id;type:uuid;primaryKey"`
	TagID  string `gorm:"column:tag_id;type:uuid;primaryKey"`
}

func (gormPostTag) TableName() string { return "post_tags" }

func (r *GormRepository) SetCategories(ctx context.Context, postID postDomain.PostID, categoryIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", string(postID)).Delete(&gormPostCategory{}).Error; err != nil {
			return err
		}
		if len(categoryIDs) == 0 {
			return nil
		}

		rows := make([]gormPostCategory, 0, len(categoryIDs))
		for _, id := range categoryIDs {
			if id == "" {
				continue
			}
			rows = append(rows, gormPostCategory{PostID: string(postID), CategoryID: id})
		}
		if len(rows) == 0 {
			return nil
		}

		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
	})
}

func (r *GormRepository) SetTags(ctx context.Context, postID postDomain.PostID, tagIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("post_id = ?", string(postID)).Delete(&gormPostTag{}).Error; err != nil {
			return err
		}
		if len(tagIDs) == 0 {
			return nil
		}

		rows := make([]gormPostTag, 0, len(tagIDs))
		for _, id := range tagIDs {
			if id == "" {
				continue
			}
			rows = append(rows, gormPostTag{PostID: string(postID), TagID: id})
		}
		if len(rows) == 0 {
			return nil
		}

		return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
	})
}

func toDomain(m *gormPost) *postDomain.Post {
	var deletedAt *time.Time
	if m.DeletedAt.Valid {
		t := m.DeletedAt.Time
		deletedAt = &t
	}
	return &postDomain.Post{
		ID:          postDomain.PostID(m.ID),
		AuthorID:    m.AuthorID,
		Title:       m.Title,
		Slug:        m.Slug,
		Content:     m.Content,
		Status:      postDomain.Status(m.Status),
		PublishedAt: m.PublishedAt,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   deletedAt,
	}
}

func fromDomain(p *postDomain.Post) *gormPost {
	var deleted gorm.DeletedAt
	if p.DeletedAt != nil {
		deleted = gorm.DeletedAt{Time: *p.DeletedAt, Valid: true}
	}
	return &gormPost{
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

