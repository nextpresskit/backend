package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	postDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain"
)

var (
	ErrInvalidPost   = errors.New("invalid_post")
	ErrSlugTaken     = errors.New("slug_taken")
	ErrPostNotFound  = errors.New("post_not_found")
	ErrInvalidStatus = errors.New("invalid_status")
)

type Service struct {
	repo postDomain.Repository
}

func NewService(repo postDomain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, authorID, title, slug, content string) (*postDomain.Post, error) {
	title = strings.TrimSpace(title)
	slug = normalizeSlug(slug)
	content = strings.TrimSpace(content)
	if authorID == "" || title == "" || slug == "" {
		return nil, ErrInvalidPost
	}

	existing, err := s.repo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrSlugTaken
	}

	now := time.Now().UTC()
	p := &postDomain.Post{
		ID:        postDomain.PostID(uuid.NewString()),
		AuthorID:  authorID,
		Title:     title,
		Slug:      slug,
		Content:   content,
		Status:    postDomain.StatusDraft,
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

func (s *Service) GetByID(ctx context.Context, id string) (*postDomain.Post, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrPostNotFound
	}
	p, err := s.repo.FindByID(ctx, postDomain.PostID(id))
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPostNotFound
	}
	return p, nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]postDomain.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListFiltered(ctx, false, limit, offset, "", "", "")
}

func (s *Service) ListFiltered(ctx context.Context, limit, offset int, status string, authorID string, q string) ([]postDomain.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" {
		switch postDomain.Status(status) {
		case postDomain.StatusDraft, postDomain.StatusPublished, postDomain.StatusArchived:
		default:
			return nil, ErrInvalidStatus
		}
	}
	authorID = strings.TrimSpace(authorID)
	q = strings.TrimSpace(q)
	return s.repo.ListFiltered(ctx, false, limit, offset, status, authorID, q)
}

func (s *Service) PublicList(ctx context.Context, limit, offset int, q string, categoryID string, tagID string) ([]postDomain.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	q = strings.TrimSpace(q)
	categoryID = strings.TrimSpace(categoryID)
	tagID = strings.TrimSpace(tagID)
	return s.repo.ListPublished(ctx, limit, offset, q, categoryID, tagID)
}

func (s *Service) PublicGetBySlug(ctx context.Context, slug string) (*postDomain.Post, error) {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil, ErrPostNotFound
	}
	p, err := s.repo.FindPublishedBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPostNotFound
	}
	return p, nil
}

func (s *Service) Update(ctx context.Context, id, title, slug, content, status string) (*postDomain.Post, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrPostNotFound
	}

	p, err := s.repo.FindByID(ctx, postDomain.PostID(id))
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPostNotFound
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
		st := postDomain.Status(strings.ToLower(strings.TrimSpace(status)))
		switch st {
		case postDomain.StatusDraft, postDomain.StatusPublished, postDomain.StatusArchived:
			p.Status = st
			if st == postDomain.StatusPublished && p.PublishedAt == nil {
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
		return ErrPostNotFound
	}
	return s.repo.Delete(ctx, postDomain.PostID(id))
}

func (s *Service) SetCategories(ctx context.Context, postID string, categoryIDs []string) error {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return ErrPostNotFound
	}

	p, err := s.repo.FindByID(ctx, postDomain.PostID(postID))
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}

	return s.repo.SetCategories(ctx, postDomain.PostID(postID), categoryIDs)
}

func (s *Service) SetTags(ctx context.Context, postID string, tagIDs []string) error {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return ErrPostNotFound
	}

	p, err := s.repo.FindByID(ctx, postDomain.PostID(postID))
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}

	return s.repo.SetTags(ctx, postDomain.PostID(postID), tagIDs)
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

