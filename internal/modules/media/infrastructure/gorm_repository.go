package infrastructure

import (
	"context"
	"strconv"
	"time"

	mediaDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/media/domain"
	"gorm.io/gorm"
)

type gormMedia struct {
	ID           string    `gorm:"column:id;type:uuid;primaryKey"`
	UploaderID   int64     `gorm:"column:uploader_id;not null;index"`
	OriginalName string    `gorm:"column:original_name;not null"`
	StorageName  string    `gorm:"column:storage_name;not null;uniqueIndex"`
	MimeType     string    `gorm:"column:mime_type;not null"`
	SizeBytes    int64     `gorm:"column:size_bytes;not null"`
	StoragePath  string    `gorm:"column:storage_path;not null"`
	PublicURL    string    `gorm:"column:public_url;not null"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
}

func (gormMedia) TableName() string { return "media" }

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
	var row gormMedia
	if err := r.db.WithContext(ctx).Where("id = ?", string(id)).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(&row), nil
}

func (r *GormRepository) List(ctx context.Context, limit, offset int) ([]mediaDomain.Media, error) {
	var rows []gormMedia
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

func toDomain(m *gormMedia) *mediaDomain.Media {
	return &mediaDomain.Media{
		ID:           mediaDomain.MediaID(m.ID),
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

func fromDomain(m *mediaDomain.Media) *gormMedia {
	return &gormMedia{
		ID:           string(m.ID),
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

