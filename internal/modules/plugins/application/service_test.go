package application

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	pluginsDomain "github.com/nextpresskit/backend/internal/modules/plugins/domain"
)

type pluginsRepoStub struct {
	byID map[pluginsDomain.PluginID]*pluginsDomain.Plugin
}

func (s *pluginsRepoStub) Create(_ context.Context, plugin *pluginsDomain.Plugin) error {
	if s.byID == nil {
		s.byID = map[pluginsDomain.PluginID]*pluginsDomain.Plugin{}
	}
	for _, v := range s.byID {
		if v.Slug == plugin.Slug {
			return gorm.ErrDuplicatedKey
		}
	}
	cp := *plugin
	s.byID[plugin.ID] = &cp
	return nil
}
func (s *pluginsRepoStub) FindByID(_ context.Context, id pluginsDomain.PluginID) (*pluginsDomain.Plugin, error) {
	return s.byID[id], nil
}
func (s *pluginsRepoStub) FindBySlug(_ context.Context, slug string) (*pluginsDomain.Plugin, error) {
	for _, v := range s.byID {
		if v.Slug == slug {
			return v, nil
		}
	}
	return nil, nil
}
func (s *pluginsRepoStub) List(_ context.Context) ([]pluginsDomain.Plugin, error) {
	out := make([]pluginsDomain.Plugin, 0, len(s.byID))
	for _, v := range s.byID {
		out = append(out, *v)
	}
	return out, nil
}
func (s *pluginsRepoStub) ListEnabled(_ context.Context) ([]pluginsDomain.Plugin, error) {
	return nil, nil
}
func (s *pluginsRepoStub) Update(_ context.Context, plugin *pluginsDomain.Plugin) error {
	s.byID[plugin.ID] = plugin
	return nil
}

func TestRegisterPlugin_DefaultVersionAndSlug(t *testing.T) {
	repo := &pluginsRepoStub{byID: map[pluginsDomain.PluginID]*pluginsDomain.Plugin{}}
	svc := NewService(repo)

	p, err := svc.Register(context.Background(), "SEO Plugin", " SEO PLUGIN ", true, "", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Slug != "seo-plugin" {
		t.Fatalf("expected normalized slug seo-plugin, got %q", p.Slug)
	}
	if p.Version != "1.0.0" {
		t.Fatalf("expected default version 1.0.0, got %q", p.Version)
	}
}

func TestRegisterPlugin_Duplicate(t *testing.T) {
	repo := &pluginsRepoStub{
		byID: map[pluginsDomain.PluginID]*pluginsDomain.Plugin{
			"id1": {ID: "id1", Name: "A", Slug: "dup", Version: "1.0.0"},
		},
	}
	svc := NewService(repo)

	_, err := svc.Register(context.Background(), "B", "dup", true, "1.0.0", nil)
	if !errors.Is(err, ErrPluginAlreadyExists) {
		t.Fatalf("expected ErrPluginAlreadyExists, got %v", err)
	}
}

func TestUpdatePlugin_NotFound(t *testing.T) {
	svc := NewService(&pluginsRepoStub{byID: map[pluginsDomain.PluginID]*pluginsDomain.Plugin{}})
	_, err := svc.Update(context.Background(), "missing", nil, nil, nil)
	if !errors.Is(err, ErrPluginNotFound) {
		t.Fatalf("expected ErrPluginNotFound, got %v", err)
	}
}

