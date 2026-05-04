package application

import (
	"context"
	"errors"
	"strings"

	posterr "github.com/nextpresskit/backend/internal/modules/posts/domain"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/series"
)

// SeriesService handles top-level series CRUD.
type SeriesService struct {
	repo series.SeriesRepository
}

// NewSeriesService constructs the series application service.
func NewSeriesService(repo series.SeriesRepository) *SeriesService {
	return &SeriesService{repo: repo}
}

func (s *SeriesService) ListSeries(ctx context.Context) ([]series.Series, error) {
	return s.repo.ListSeries(ctx)
}

func (s *SeriesService) CreateSeries(ctx context.Context, title, slug string) (*series.Series, error) {
	title = strings.TrimSpace(title)
	slug = strings.TrimSpace(slug)
	if title == "" || slug == "" {
		return nil, ErrInvalidArgument
	}
	sr := &series.Series{Title: title, Slug: slug}
	if err := s.repo.CreateSeries(ctx, sr); err != nil {
		if errors.Is(err, posterr.ErrConflict) {
			return nil, ErrSlugTaken
		}
		return nil, err
	}
	return sr, nil
}

func (s *SeriesService) GetSeries(ctx context.Context, id string) (*series.Series, error) {
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

func (s *SeriesService) UpdateSeries(ctx context.Context, id string, title, slug *string) (*series.Series, error) {
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
		if errors.Is(err, posterr.ErrNotFound) {
			return nil, ErrNotFound
		}
		if errors.Is(err, posterr.ErrConflict) {
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
