package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	menuDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/menus/domain"
)

var (
	ErrInvalidInput  = errors.New("invalid_input")
	ErrAlreadyExists = errors.New("already_exists")
	ErrNotFound      = errors.New("not_found")
)

type Service struct {
	repo menuDomain.Repository
}

func NewService(repo menuDomain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateMenu(ctx context.Context, name, slug string) (*menuDomain.Menu, error) {
	name = strings.TrimSpace(name)
	slug = normalizeSlug(slug)
	if name == "" || slug == "" {
		return nil, ErrInvalidInput
	}

	now := time.Now().UTC()
	m := &menuDomain.Menu{
		ID:        menuDomain.MenuID(uuid.NewString()),
		Name:      name,
		Slug:      slug,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.CreateMenu(ctx, m); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}
	return m, nil
}

func (s *Service) ListMenus(ctx context.Context, limit, offset int) ([]menuDomain.Menu, error) {
	limit, offset = normalizeList(limit, offset)
	return s.repo.ListMenus(ctx, limit, offset)
}

func (s *Service) GetMenu(ctx context.Context, id string) (*menuDomain.Menu, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrNotFound
	}
	m, err := s.repo.FindMenuByID(ctx, menuDomain.MenuID(id))
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrNotFound
	}
	return m, nil
}

func (s *Service) PublicGetMenuBySlug(ctx context.Context, slug string) (*menuDomain.Menu, []menuDomain.MenuItem, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, nil, ErrNotFound
	}
	m, err := s.repo.FindMenuBySlug(ctx, slug)
	if err != nil {
		return nil, nil, err
	}
	if m == nil {
		return nil, nil, ErrNotFound
	}
	items, err := s.repo.ListMenuItems(ctx, m.ID)
	if err != nil {
		return nil, nil, err
	}
	return m, items, nil
}

func (s *Service) UpdateMenu(ctx context.Context, id, name, slug string) (*menuDomain.Menu, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrNotFound
	}
	m, err := s.repo.FindMenuByID(ctx, menuDomain.MenuID(id))
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrNotFound
	}

	if v := strings.TrimSpace(name); v != "" {
		m.Name = v
	}
	if v := normalizeSlug(slug); v != "" {
		m.Slug = v
	}
	m.UpdatedAt = time.Now().UTC()

	if err := s.repo.UpdateMenu(ctx, m); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}
	return m, nil
}

func (s *Service) DeleteMenu(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrNotFound
	}
	return s.repo.DeleteMenu(ctx, menuDomain.MenuID(id))
}

type ItemInput struct {
	ID        string
	ParentID  *string
	Label     string
	ItemType  string
	RefID     *string
	URL       *string
	SortOrder int
}

func (s *Service) ListItems(ctx context.Context, menuID string) ([]menuDomain.MenuItem, error) {
	menuID = strings.TrimSpace(menuID)
	if menuID == "" {
		return nil, ErrNotFound
	}
	// Ensure menu exists
	m, err := s.repo.FindMenuByID(ctx, menuDomain.MenuID(menuID))
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrNotFound
	}
	return s.repo.ListMenuItems(ctx, menuDomain.MenuID(menuID))
}

func (s *Service) ReplaceItems(ctx context.Context, menuID string, inputs []ItemInput) error {
	menuID = strings.TrimSpace(menuID)
	if menuID == "" {
		return ErrNotFound
	}
	// Ensure menu exists
	m, err := s.repo.FindMenuByID(ctx, menuDomain.MenuID(menuID))
	if err != nil {
		return err
	}
	if m == nil {
		return ErrNotFound
	}

	now := time.Now().UTC()
	items := make([]menuDomain.MenuItem, 0, len(inputs))
	for _, in := range inputs {
		label := strings.TrimSpace(in.Label)
		if label == "" {
			return ErrInvalidInput
		}

		t := menuDomain.ItemType(strings.ToLower(strings.TrimSpace(in.ItemType)))
		switch t {
		case menuDomain.ItemTypeURL:
			if in.URL == nil || strings.TrimSpace(*in.URL) == "" {
				return ErrInvalidInput
			}
		case menuDomain.ItemTypePage, menuDomain.ItemTypePost:
			if in.RefID == nil || strings.TrimSpace(*in.RefID) == "" {
				return ErrInvalidInput
			}
		default:
			return ErrInvalidInput
		}

		id := strings.TrimSpace(in.ID)
		if id == "" {
			id = uuid.NewString()
		}

		var parent *menuDomain.MenuItemID
		if in.ParentID != nil && strings.TrimSpace(*in.ParentID) != "" {
			p := menuDomain.MenuItemID(strings.TrimSpace(*in.ParentID))
			parent = &p
		}

		items = append(items, menuDomain.MenuItem{
			ID:        menuDomain.MenuItemID(id),
			MenuID:    menuDomain.MenuID(menuID),
			ParentID:  parent,
			Label:     label,
			ItemType:  t,
			RefID:     in.RefID,
			URL:       in.URL,
			SortOrder: in.SortOrder,
			CreatedAt: now,
			UpdatedAt: now,
		})
	}

	return s.repo.ReplaceMenuItems(ctx, menuDomain.MenuID(menuID), items)
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

