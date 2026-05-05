package application

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	pageDomain "github.com/nextpresskit/backend/internal/modules/pages/domain"
)

var (
	ErrInvalidPage   = errors.New("invalid_page")
	ErrSlugTaken     = errors.New("slug_taken")
	ErrPageNotFound  = errors.New("page_not_found")
	ErrInvalidStatus = errors.New("invalid_status")
)

type Service struct {
	repo pageDomain.Repository
}

func NewService(repo pageDomain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, authorID, title, slug, content string) (*pageDomain.Page, error) {
	title = strings.TrimSpace(title)
	slug = normalizeSlug(slug)
	content = strings.TrimSpace(content)
	if authorID == "" || title == "" || slug == "" {
		return nil, ErrInvalidPage
	}

	existing, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrSlugTaken
	}

	now := time.Now().UTC()
	p := &pageDomain.Page{
		UUID:      uuid.NewString(),
		AuthorID:  authorID,
		Title:     title,
		Slug:      slug,
		Content:   content,
		Status:    pageDomain.StatusDraft,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.repo.Create(ctx, p); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}

	return p, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*pageDomain.Page, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrPageNotFound
	}
	p, err := s.resolveByIDOrUUID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPageNotFound
	}
	return p, nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]pageDomain.Page, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListFiltered(ctx, false, limit, offset, "", "", "")
}

func (s *Service) ListFiltered(ctx context.Context, limit, offset int, status string, authorID string, q string) ([]pageDomain.Page, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" {
		switch pageDomain.Status(status) {
		case pageDomain.StatusDraft, pageDomain.StatusPublished, pageDomain.StatusArchived:
		default:
			return nil, ErrInvalidStatus
		}
	}
	authorID = strings.TrimSpace(authorID)
	q = strings.TrimSpace(q)
	return s.repo.ListFiltered(ctx, false, limit, offset, status, authorID, q)
}

func (s *Service) PublicGetBySlug(ctx context.Context, slug string) (*pageDomain.Page, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, ErrPageNotFound
	}
	p, err := s.repo.FindPublishedBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPageNotFound
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id, title, slug, content, status string) (*pageDomain.Page, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrPageNotFound
	}

	p, err := s.resolveByIDOrUUID(ctx, id)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPageNotFound
	}

	if t := strings.TrimSpace(title); t != "" {
		p.Title = t
	}
	if sSlug := normalizeSlug(slug); sSlug != "" && sSlug != p.Slug {
		existing, err := s.repo.FindBySlug(ctx, sSlug)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != p.ID {
			return nil, ErrSlugTaken
		}
		p.Slug = sSlug
	}
	if c := strings.TrimSpace(content); c != "" {
		p.Content = c
	}
	if status != "" {
		st := pageDomain.Status(strings.ToLower(strings.TrimSpace(status)))
		switch st {
		case pageDomain.StatusDraft, pageDomain.StatusPublished, pageDomain.StatusArchived:
			p.Status = st
			if st == pageDomain.StatusPublished && p.PublishedAt == nil {
				now := time.Now().UTC()
				p.PublishedAt = &now
			}
		default:
			return nil, ErrInvalidStatus
		}
	}

	p.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, p); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}

	return p, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrPageNotFound
	}
	pg, err := s.resolveByIDOrUUID(ctx, id)
	if err != nil {
		return err
	}
	if pg == nil {
		return ErrPageNotFound
	}
	return s.repo.Delete(ctx, pg.ID)
}

func (s *Service) resolveByIDOrUUID(ctx context.Context, idOrUUID string) (*pageDomain.Page, error) {
	if idNum, err := strconv.ParseInt(idOrUUID, 10, 64); err == nil && idNum > 0 {
		p, err := s.repo.FindByID(ctx, pageDomain.PageID(idNum))
		if err != nil {
			return nil, err
		}
		if p != nil {
			return p, nil
		}
	}
	return s.repo.FindByUUID(ctx, idOrUUID)
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

