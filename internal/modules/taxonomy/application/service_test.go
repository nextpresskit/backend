package application

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	taxDomain "github.com/nextpresskit/backend/internal/modules/taxonomy/domain"
)

type taxonomyRepoStub struct {
	categories map[taxDomain.CategoryID]*taxDomain.Category
	tags       map[taxDomain.TagID]*taxDomain.Tag
}

func (s *taxonomyRepoStub) CreateCategory(_ context.Context, c *taxDomain.Category) error {
	if s.categories == nil {
		s.categories = map[taxDomain.CategoryID]*taxDomain.Category{}
	}
	for _, v := range s.categories {
		if v.Slug == c.Slug {
			return gorm.ErrDuplicatedKey
		}
	}
	cp := *c
	s.categories[c.ID] = &cp
	return nil
}
func (s *taxonomyRepoStub) ListCategories(_ context.Context, _, _ int) ([]taxDomain.Category, error) {
	out := make([]taxDomain.Category, 0, len(s.categories))
	for _, v := range s.categories {
		out = append(out, *v)
	}
	return out, nil
}
func (s *taxonomyRepoStub) FindCategoryByID(_ context.Context, id taxDomain.CategoryID) (*taxDomain.Category, error) {
	return s.categories[id], nil
}
func (s *taxonomyRepoStub) FindCategoryByUUID(_ context.Context, uuid string) (*taxDomain.Category, error) {
	for _, c := range s.categories {
		if c != nil && c.UUID == uuid {
			cp := *c
			return &cp, nil
		}
	}
	return nil, nil
}
func (s *taxonomyRepoStub) UpdateCategory(_ context.Context, c *taxDomain.Category) error {
	s.categories[c.ID] = c
	return nil
}
func (s *taxonomyRepoStub) DeleteCategory(_ context.Context, uuid string) error {
	for id, c := range s.categories {
		if c != nil && c.UUID == uuid {
			delete(s.categories, id)
			return nil
		}
	}
	return nil
}
func (s *taxonomyRepoStub) CreateTag(_ context.Context, t *taxDomain.Tag) error {
	if s.tags == nil {
		s.tags = map[taxDomain.TagID]*taxDomain.Tag{}
	}
	for _, v := range s.tags {
		if v.Slug == t.Slug {
			return gorm.ErrDuplicatedKey
		}
	}
	cp := *t
	s.tags[t.ID] = &cp
	return nil
}
func (s *taxonomyRepoStub) ListTags(_ context.Context, _, _ int) ([]taxDomain.Tag, error) {
	out := make([]taxDomain.Tag, 0, len(s.tags))
	for _, v := range s.tags {
		out = append(out, *v)
	}
	return out, nil
}
func (s *taxonomyRepoStub) FindTagByID(_ context.Context, id taxDomain.TagID) (*taxDomain.Tag, error) {
	return s.tags[id], nil
}
func (s *taxonomyRepoStub) FindTagByUUID(_ context.Context, uuid string) (*taxDomain.Tag, error) {
	for _, t := range s.tags {
		if t != nil && t.UUID == uuid {
			cp := *t
			return &cp, nil
		}
	}
	return nil, nil
}
func (s *taxonomyRepoStub) UpdateTag(_ context.Context, t *taxDomain.Tag) error {
	s.tags[t.ID] = t
	return nil
}
func (s *taxonomyRepoStub) DeleteTag(_ context.Context, uuid string) error {
	for id, t := range s.tags {
		if t != nil && t.UUID == uuid {
			delete(s.tags, id)
			return nil
		}
	}
	return nil
}

func TestCreateCategory_DuplicateSlug(t *testing.T) {
	repo := &taxonomyRepoStub{
		categories: map[taxDomain.CategoryID]*taxDomain.Category{
			1: {ID: 1, UUID: "00000000-0000-0000-0000-0000000000c1", Name: "Go", Slug: "go"},
		},
	}
	svc := NewService(repo)

	_, err := svc.CreateCategory(context.Background(), "Go", "go")
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected ErrAlreadyExists, got %v", err)
	}
}

func TestCreateTag_Success(t *testing.T) {
	repo := &taxonomyRepoStub{
		categories: map[taxDomain.CategoryID]*taxDomain.Category{},
		tags:       map[taxDomain.TagID]*taxDomain.Tag{},
	}
	svc := NewService(repo)

	tag, err := svc.CreateTag(context.Background(), "GraphQL", " GraphQL ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.Slug != "graphql" {
		t.Fatalf("expected normalized slug graphql, got %q", tag.Slug)
	}
}

func TestUpdateCategory_NotFound(t *testing.T) {
	svc := NewService(&taxonomyRepoStub{
		categories: map[taxDomain.CategoryID]*taxDomain.Category{},
		tags:       map[taxDomain.TagID]*taxDomain.Tag{},
	})
	_, err := svc.UpdateCategory(context.Background(), "missing", "New", "new")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
