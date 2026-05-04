package infrastructure

import (
	"context"
	"time"

	"github.com/nextpresskit/backend/internal/modules/user/domain"
	userp "github.com/nextpresskit/backend/internal/modules/user/persistence"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindByID(id domain.UserID) (*domain.User, error) {
	var u userp.User
	if err := r.db.WithContext(context.Background()).
		Where("public_id = ?", int64(id)).
		First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(&u), nil
}

func (r *GormRepository) FindByUUID(uuid string) (*domain.User, error) {
	var u userp.User
	if err := r.db.WithContext(context.Background()).
		Where("id = ?", uuid).
		First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(&u), nil
}

func (r *GormRepository) FindByEmail(email string) (*domain.User, error) {
	var u userp.User
	if err := r.db.WithContext(context.Background()).
		Where("LOWER(email) = LOWER(?)", email).
		First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(&u), nil
}

func (r *GormRepository) Create(user *domain.User) error {
	u := fromDomain(user)
	if err := r.db.WithContext(context.Background()).Create(u).Error; err != nil {
		return err
	}
	*user = *toDomain(u)
	return nil
}

func (r *GormRepository) Update(user *domain.User) error {
	u := fromDomain(user)
	return r.db.WithContext(context.Background()).
		Model(&userp.User{}).
		Where("id = ?", u.UUID).
		Updates(&u).Error
}

func (r *GormRepository) Delete(id domain.UserID) error {
	return r.db.WithContext(context.Background()).
		Where("public_id = ?", int64(id)).
		Delete(&userp.User{}).Error
}

func toDomain(u *userp.User) *domain.User {
	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		t := u.DeletedAt.Time
		deletedAt = &t
	}
	return &domain.User{
		ID:        u.PublicID,
		UUID:      u.UUID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Password:  u.Password,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		DeletedAt: deletedAt,
	}
}

func fromDomain(u *domain.User) *userp.User {
	var deleted gorm.DeletedAt
	if u.DeletedAt != nil {
		deleted = gorm.DeletedAt{Time: *u.DeletedAt, Valid: true}
	}
	return &userp.User{
		PublicID:  u.ID,
		UUID:      u.UUID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Password:  u.Password,
		Active:    u.Active,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
		DeletedAt: deleted,
	}
}
