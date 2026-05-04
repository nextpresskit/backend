package application

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	posterr "github.com/nextpresskit/backend/internal/modules/posts/domain"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/extensions"
)

// TranslationGroupsService manages translation_groups rows.
type TranslationGroupsService struct {
	repo extensions.TranslationGroupRepository
}

// NewTranslationGroupsService constructs the translation groups service.
func NewTranslationGroupsService(repo extensions.TranslationGroupRepository) *TranslationGroupsService {
	return &TranslationGroupsService{repo: repo}
}

func (s *TranslationGroupsService) CreateTranslationGroup(ctx context.Context, explicitID *string) (string, error) {
	id := ""
	if explicitID != nil {
		id = strings.TrimSpace(*explicitID)
	}
	if id == "" {
		id = uuid.NewString()
	}
	if err := s.repo.CreateTranslationGroup(ctx, id); err != nil {
		if errors.Is(err, posterr.ErrConflict) {
			return "", ErrConflict
		}
		return "", err
	}
	return id, nil
}

func (s *TranslationGroupsService) TranslationGroupExists(ctx context.Context, id string) (bool, error) {
	return s.repo.FindTranslationGroup(ctx, strings.TrimSpace(id))
}

func (s *TranslationGroupsService) DeleteTranslationGroup(ctx context.Context, id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return ErrNotFound
	}
	return s.repo.DeleteTranslationGroup(ctx, id)
}
