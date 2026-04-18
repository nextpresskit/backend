package infrastructure

import (
	"context"
	"time"

	"github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/user/domain"
	"gorm.io/gorm"
)

type gormUser struct {
	ID        string         `gorm:"type:uuid;primaryKey"`
	FirstName string         `gorm:"not null"`
	LastName  string         `gorm:"not null"`
	Email     string         `gorm:"not null;uniqueIndex"`
	Password  string         `gorm:"not null"`
	Active    bool           `gorm:"not null;default:true"`
	CreatedAt time.Time      `gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time      `gorm:"not null;autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (gormUser) TableName() string {
	return "users"
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) FindByID(id domain.UserID) (*domain.User, error) {
	var u gormUser
	if err := r.db.WithContext(context.Background()).
		Where("id = ?", string(id)).
		First(&u).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(&u), nil
}

func (r *GormRepository) FindByEmail(email string) (*domain.User, error) {
	var u gormUser
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
		Model(&gormUser{}).
		Where("id = ?", u.ID).
		Updates(&u).Error
}

func (r *GormRepository) Delete(id domain.UserID) error {
	return r.db.WithContext(context.Background()).
		Where("id = ?", string(id)).
		Delete(&gormUser{}).Error
}

func toDomain(u *gormUser) *domain.User {
	var deletedAt *time.Time
	if u.DeletedAt.Valid {
		t := u.DeletedAt.Time
		deletedAt = &t
	}
	return &domain.User{
		ID:        domain.UserID(u.ID),
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

func fromDomain(u *domain.User) *gormUser {
	var deleted gorm.DeletedAt
	if u.DeletedAt != nil {
		deleted = gorm.DeletedAt{Time: *u.DeletedAt, Valid: true}
	}
	return &gormUser{
		ID:        string(u.ID),
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
