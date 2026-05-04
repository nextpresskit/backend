package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	pluginsDomain "github.com/nextpresskit/backend/internal/modules/plugins/domain"
)

type Service struct {
	repo pluginsDomain.Repository
}

func NewService(repo pluginsDomain.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]pluginsDomain.Plugin, error) {
	return s.repo.List(ctx)
}

func (s *Service) Register(ctx context.Context, name, slug string, enabled bool, version string, config map[string]any) (*pluginsDomain.Plugin, error) {
	name = strings.TrimSpace(name)
	slug = normalizeSlug(slug)

	if name == "" || slug == "" {
		return nil, ErrInvalidPluginInput
	}

	version = strings.TrimSpace(version)
	if version == "" {
		version = "1.0.0"
	}

	if config == nil {
		config = map[string]any{}
	}

	p := &pluginsDomain.Plugin{
		ID:        pluginsDomain.PluginID(uuid.NewString()),
		Name:      name,
		Slug:      slug,
		Enabled:   enabled,
		Version:   version,
		Config:    config,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, p); err != nil {
		// Preserve a stable, domain-level error contract for handlers.
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrPluginAlreadyExists
		}
		return nil, err
	}

	return p, nil
}

func (s *Service) Update(ctx context.Context, id string, enabled *bool, version *string, config *map[string]any) (*pluginsDomain.Plugin, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrPluginNotFound
	}

	p, err := s.repo.FindByID(ctx, pluginsDomain.PluginID(id))
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ErrPluginNotFound
	}

	// Only apply fields explicitly provided by the caller.
	if enabled != nil {
		p.Enabled = *enabled
	}
	if version != nil {
		v := strings.TrimSpace(*version)
		if v != "" {
			p.Version = v
		}
	}
	if config != nil {
		if *config == nil {
			p.Config = map[string]any{}
		} else {
			p.Config = *config
		}
	}

	p.UpdatedAt = time.Now().UTC()

	if err := s.repo.Update(ctx, p); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, ErrPluginAlreadyExists
		}
		return nil, err
	}

	return p, nil
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

