package infrastructure

import (
	"github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/user/domain"
	"gorm.io/gorm"
)

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

// Methods will be implemented when we add real user persistence.
func (r *GormRepository) FindByID(id domain.UserID) (*domain.User, error) { return nil, nil }
func (r *GormRepository) FindByEmail(email string) (*domain.User, error)  { return nil, nil }
func (r *GormRepository) Create(user *domain.User) error                  { return nil }
func (r *GormRepository) Update(user *domain.User) error                  { return nil }
func (r *GormRepository) Delete(id domain.UserID) error                   { return nil }
