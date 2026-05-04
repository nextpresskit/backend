package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	taxDomain "github.com/nextpresskit/backend/internal/modules/taxonomy/domain"
)

var (
	ErrInvalidInput  = errors.New("invalid_input")
	ErrAlreadyExists = errors.New("already_exists")
	ErrNotFound      = errors.New("not_found")
)

type Service struct {
	repo taxDomain.Repository
}

func NewService(repo taxDomain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateCategory(ctx context.Context, name, slug string) (*taxDomain.Category, error) {
	name = strings.TrimSpace(name)
	slug = normalizeSlug(slug)
	if name == "" || slug == "" {
		return nil, ErrInvalidInput
	}
	now := time.Now().UTC()
	c := &taxDomain.Category{
		ID:        taxDomain.CategoryID(uuid.NewString()),
		Name:      name,
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateCategory(ctx, c); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}
	return c, nil
}

func (s *Service) ListCategories(ctx context.Context, limit, offset int) ([]taxDomain.Category, error) {
	limit, offset = normalizeList(limit, offset)
	return s.repo.ListCategories(ctx, limit, offset)
}

func (s *Service) UpdateCategory(ctx context.Context, id, name, slug string) (*taxDomain.Category, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrNotFound
	}
	existing, err := s.repo.FindCategoryByID(ctx, taxDomain.CategoryID(id))
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	if v := strings.TrimSpace(name); v != "" {
		existing.Name = v
	}
	if v := normalizeSlug(slug); v != "" {
		existing.Slug = v
	}
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateCategory(ctx, existing); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrNotFound
	}
	return s.repo.DeleteCategory(ctx, taxDomain.CategoryID(id))
}

func (s *Service) CreateTag(ctx context.Context, name, slug string) (*taxDomain.Tag, error) {
	name = strings.TrimSpace(name)
	slug = normalizeSlug(slug)
	if name == "" || slug == "" {
		return nil, ErrInvalidInput
	}
	now := time.Now().UTC()
	t := &taxDomain.Tag{
		ID:        taxDomain.TagID(uuid.NewString()),
		Name:      name,
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateTag(ctx, t); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}
	return t, nil
}

func (s *Service) ListTags(ctx context.Context, limit, offset int) ([]taxDomain.Tag, error) {
	limit, offset = normalizeList(limit, offset)
	return s.repo.ListTags(ctx, limit, offset)
}

func (s *Service) UpdateTag(ctx context.Context, id, name, slug string) (*taxDomain.Tag, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrNotFound
	}
	existing, err := s.repo.FindTagByID(ctx, taxDomain.TagID(id))
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	if v := strings.TrimSpace(name); v != "" {
		existing.Name = v
	}
	if v := normalizeSlug(slug); v != "" {
		existing.Slug = v
	}
	existing.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateTag(ctx, existing); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}
	return existing, nil
}

func (s *Service) DeleteTag(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrNotFound
	}
	return s.repo.DeleteTag(ctx, taxDomain.TagID(id))
}

func normalizeSlug(slug string) string {
	s := strings.ToLower(strings.TrimSpace(slug))
	s = strings.ReplaceAll(s, " ", "-")
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	s = strings.Trim(s, "-")
	return s
}

func normalizeList(limit, offset int) (int, int) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

