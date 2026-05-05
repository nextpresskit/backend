package infrastructure

import (
	"context"
	"strconv"

	mediaDomain "github.com/nextpresskit/backend/internal/modules/media/domain"
	mediap "github.com/nextpresskit/backend/internal/modules/media/persistence"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, m *mediaDomain.Media) error {
	row := fromDomain(m)
	if err := r.db.WithContext(ctx).Create(row).Error; err != nil {
		return err
	}
	*m = *toDomain(row)
	return nil
}

func (r *GormRepository) FindByID(ctx context.Context, id mediaDomain.MediaID) (*mediaDomain.Media, error) {
	var row mediap.Media
	if err := r.db.WithContext(ctx).Where("id = ?", int64(id)).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(&row), nil
}

func (r *GormRepository) FindByUUID(ctx context.Context, uuid string) (*mediaDomain.Media, error) {
	var row mediap.Media
	if err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(&row), nil
}

func (r *GormRepository) List(ctx context.Context, limit, offset int) ([]mediaDomain.Media, error) {
	var rows []mediap.Media
	if err := r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]mediaDomain.Media, 0, len(rows))
	for i := range rows {
		m := toDomain(&rows[i])
		out = append(out, *m)
	}
	return out, nil
}

func toDomain(m *mediap.Media) *mediaDomain.Media {
	return &mediaDomain.Media{
		ID:           mediaDomain.MediaID(m.ID),
		UUID:         m.UUID,
		UploaderID:   strconv.FormatInt(m.UploaderID, 10),
		OriginalName: m.OriginalName,
		StorageName:  m.StorageName,
		MimeType:     m.MimeType,
		SizeBytes:    m.SizeBytes,
		StoragePath:  m.StoragePath,
		PublicURL:    m.PublicURL,
		CreatedAt:    m.CreatedAt,
	}
}

func fromDomain(m *mediaDomain.Media) *mediap.Media {
	return &mediap.Media{
		ID:           int64(m.ID),
		UUID:         m.UUID,
		UploaderID:   parseInt64OrZero(m.UploaderID),
		OriginalName: m.OriginalName,
		StorageName:  m.StorageName,
		MimeType:     m.MimeType,
		SizeBytes:    m.SizeBytes,
		StoragePath:  m.StoragePath,
		PublicURL:    m.PublicURL,
		CreatedAt:    m.CreatedAt,
	}
}

func parseInt64OrZero(v string) int64 {
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
