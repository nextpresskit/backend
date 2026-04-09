package application

import (
	"context"
	"errors"
	"strings"

	postDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain"
)

// SeriesService handles top-level series CRUD.
type SeriesService struct {
	repo postDomain.SeriesRepository
}

// NewSeriesService constructs the series application service.
func NewSeriesService(repo postDomain.SeriesRepository) *SeriesService {
	return &SeriesService{repo: repo}
}

func (s *SeriesService) ListSeries(ctx context.Context) ([]postDomain.Series, error) {
	return s.repo.ListSeries(ctx)
}

func (s *SeriesService) CreateSeries(ctx context.Context, title, slug string) (*postDomain.Series, error) {
	title = strings.TrimSpace(title)
	slug = strings.TrimSpace(slug)
	if title == "" || slug == "" {
		return nil, ErrInvalidArgument
	}
	sr := &postDomain.Series{Title: title, Slug: slug}
	if err := s.repo.CreateSeries(ctx, sr); err != nil {
		if errors.Is(err, postDomain.ErrConflict) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	return sr, nil
}

func (s *SeriesService) GetSeries(ctx context.Context, id string) (*postDomain.Series, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrNotFound
	}
	sr, err := s.repo.FindSeriesByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if sr == nil {
		return nil, ErrNotFound
	}
	return sr, nil
}

func (s *SeriesService) UpdateSeries(ctx context.Context, id string, title, slug *string) (*postDomain.Series, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrNotFound
	}
	ex, err := s.repo.FindSeriesByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, ErrNotFound
	}
	if title != nil {
		ex.Title = strings.TrimSpace(*title)
	}
	if slug != nil {
		ex.Slug = strings.TrimSpace(*slug)
	}
	if ex.Title == "" || ex.Slug == "" {
		return nil, ErrInvalidArgument
	}
	if err := s.repo.UpdateSeries(ctx, ex); err != nil {
		if errors.Is(err, postDomain.ErrNotFound) {
			return nil, ErrNotFound
		}
		if errors.Is(err, postDomain.ErrConflict) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	return s.repo.FindSeriesByID(ctx, id)
}

func (s *SeriesService) DeleteSeries(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrNotFound
	}
	return s.repo.DeleteSeries(ctx, id)
}
