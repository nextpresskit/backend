package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	posterr "github.com/nextpresskit/backend/internal/modules/posts/domain"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/model"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/ports"
)

// CorePostsService handles core post CRUD, public reads, and taxonomy assignments.
type CorePostsService struct {
	repo  ports.CorePostsPersistence
	hooks ports.PostSave
}

// NewCorePostsService constructs the core posts application service.
func NewCorePostsService(repo ports.CorePostsPersistence, hooks ports.PostSave) *CorePostsService {
	return &CorePostsService{repo: repo, hooks: hooks}
}

func (s *CorePostsService) Create(ctx context.Context, authorID, title, slug, content string) (*model.Post, error) {
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
	p := &model.Post{
		ID:            ident.PostID(uuid.NewString()),
		AuthorID:      authorID,
		Title:         title,
		Slug:          slug,
		Content:       content,
		Status:        ident.StatusDraft,
		Visibility:    "public",
		Locale:        "en-US",
		Timezone:      "UTC",
		WorkflowStage: "draft",
		Revision:      1,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if s.hooks != nil {
		if err := s.hooks.BeforePostSave(ctx, string(p.ID), p.Slug); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Create(ctx, p); err != nil {
		if errors.Is(err, posterr.ErrConflict) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}

	if s.hooks != nil {
		if err := s.hooks.AfterPostSave(ctx, string(p.ID), p.Slug); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (s *CorePostsService) GetByID(ctx context.Context, id string) (*model.Post, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrPostNotFound
	}
	p, err := s.repo.FindByID(ctx, ident.PostID(id))
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPostNotFound
	}
	return p, nil
}

func (s *CorePostsService) List(ctx context.Context, limit, offset int) ([]model.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListFiltered(ctx, false, limit, offset, "", "", "")
}

func (s *CorePostsService) ListFiltered(ctx context.Context, limit, offset int, status string, authorID string, q string) ([]model.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status != "" {
		switch ident.Status(status) {
		case ident.StatusDraft, ident.StatusPublished, ident.StatusArchived:
		default:
			return nil, ErrInvalidStatus
		}
	}
	authorID = strings.TrimSpace(authorID)
	q = strings.TrimSpace(q)
	return s.repo.ListFiltered(ctx, false, limit, offset, status, authorID, q)
}

func (s *CorePostsService) PublicList(ctx context.Context, limit, offset int, q string, categoryID string, tagID string) ([]model.Post, error) {
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

func (s *CorePostsService) PublicGetBySlug(ctx context.Context, slug string) (*model.Post, error) {
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

// ReindexPublishedForSearch walks all published posts in pages of pageSize and calls sync for each (e.g. Elasticsearch).
func (s *CorePostsService) ReindexPublishedForSearch(ctx context.Context, sync func(context.Context, *model.Post)) (int, error) {
	const pageSize = 100
	var total int
	offset := 0
	for {
		list, err := s.repo.ListPublished(ctx, pageSize, offset, "", "", "")
		if err != nil {
			return total, err
		}
		if len(list) == 0 {
			break
		}
		for i := range list {
			sync(ctx, &list[i])
			total++
		}
		if len(list) < pageSize {
			break
		}
		offset += pageSize
	}
	return total, nil
}

func (s *CorePostsService) Update(ctx context.Context, id, title, slug, content, status string) (*model.Post, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrPostNotFound
	}

	p, err := s.repo.FindByID(ctx, ident.PostID(id))
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
		st := ident.Status(strings.ToLower(strings.TrimSpace(status)))
		switch st {
		case ident.StatusDraft, ident.StatusPublished, ident.StatusArchived:
			p.Status = st
			if st == ident.StatusPublished && p.PublishedAt == nil {
				now := time.Now().UTC()
				p.PublishedAt = &now
			}
		default:
			return nil, ErrInvalidStatus
		}
	}

	if p.Visibility == "" {
		p.Visibility = "public"
	}
	if p.Locale == "" {
		p.Locale = "en-US"
	}
	if p.Timezone == "" {
		p.Timezone = "UTC"
	}
	if p.WorkflowStage == "" {
		p.WorkflowStage = "draft"
	}
	if p.Revision <= 0 {
		p.Revision = 1
	}

	p.UpdatedAt = time.Now().UTC()

	if s.hooks != nil {
		if err := s.hooks.BeforePostSave(ctx, string(p.ID), p.Slug); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, p); err != nil {
		if errors.Is(err, posterr.ErrConflict) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}

	if s.hooks != nil {
		if err := s.hooks.AfterPostSave(ctx, string(p.ID), p.Slug); err != nil {
			return nil, err
		}
	}

	return p, nil
}

// Save persists an already-loaded post with any fields updated by callers.
func (s *CorePostsService) Save(ctx context.Context, p *model.Post) (*model.Post, error) {
	if p == nil || strings.TrimSpace(string(p.ID)) == "" {
		return nil, ErrPostNotFound
	}

	if p.Visibility == "" {
		p.Visibility = "public"
	}
	if p.Locale == "" {
		p.Locale = "en-US"
	}
	if p.Timezone == "" {
		p.Timezone = "UTC"
	}
	if p.WorkflowStage == "" {
		p.WorkflowStage = "draft"
	}
	if p.Revision <= 0 {
		p.Revision = 1
	}

	p.UpdatedAt = time.Now().UTC()

	if s.hooks != nil {
		if err := s.hooks.BeforePostSave(ctx, string(p.ID), p.Slug); err != nil {
			return nil, err
		}
	}

	if err := s.repo.Update(ctx, p); err != nil {
		if errors.Is(err, posterr.ErrConflict) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}

	if s.hooks != nil {
		if err := s.hooks.AfterPostSave(ctx, string(p.ID), p.Slug); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (s *CorePostsService) Delete(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrPostNotFound
	}
	p, err := s.repo.FindByID(ctx, ident.PostID(id))
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}
	slug := p.Slug
	if err := s.repo.Delete(ctx, ident.PostID(id)); err != nil {
		return err
	}
	if s.hooks != nil {
		if err := s.hooks.AfterPostSave(ctx, id, slug); err != nil {
			return err
		}
	}
	return nil
}

func (s *CorePostsService) SetCategories(ctx context.Context, postID string, categoryIDs []string) error {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return ErrPostNotFound
	}

	p, err := s.repo.FindByID(ctx, ident.PostID(postID))
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}

	return s.repo.SetCategories(ctx, ident.PostID(postID), categoryIDs)
}

func (s *CorePostsService) SetTags(ctx context.Context, postID string, tagIDs []string) error {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return ErrPostNotFound
	}

	p, err := s.repo.FindByID(ctx, ident.PostID(postID))
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}

	return s.repo.SetTags(ctx, ident.PostID(postID), tagIDs)
}

func (s *CorePostsService) SetPrimaryCategory(ctx context.Context, postID string, categoryID *string) error {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return ErrPostNotFound
	}
	p, err := s.repo.FindByID(ctx, ident.PostID(postID))
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}
	return s.repo.SetPrimaryCategory(ctx, ident.PostID(postID), categoryID)
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
