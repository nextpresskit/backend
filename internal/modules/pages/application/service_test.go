package application

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	pageDomain "github.com/nextpresskit/backend/internal/modules/pages/domain"
)

type pagesRepoStub struct {
	byID   map[pageDomain.PageID]*pageDomain.Page
	bySlug map[string]*pageDomain.Page
}

func (s *pagesRepoStub) Create(_ context.Context, page *pageDomain.Page) error {
	if s.byID == nil {
		s.byID = map[pageDomain.PageID]*pageDomain.Page{}
	}
	if s.bySlug == nil {
		s.bySlug = map[string]*pageDomain.Page{}
	}
	if _, exists := s.bySlug[page.Slug]; exists {
		return gorm.ErrDuplicatedKey
	}
	cp := *page
	s.byID[cp.ID] = &cp
	s.bySlug[cp.Slug] = &cp
	return nil
}
func (s *pagesRepoStub) FindByID(_ context.Context, id pageDomain.PageID) (*pageDomain.Page, error) {
	return s.byID[id], nil
}
func (s *pagesRepoStub) FindBySlug(_ context.Context, slug string) (*pageDomain.Page, error) {
	return s.bySlug[slug], nil
}
func (s *pagesRepoStub) List(_ context.Context, _ bool, _, _ int) ([]pageDomain.Page, error) {
	return nil, nil
}
func (s *pagesRepoStub) ListFiltered(_ context.Context, _ bool, _, _ int, _, _, _ string) ([]pageDomain.Page, error) {
	return nil, nil
}
func (s *pagesRepoStub) FindPublishedBySlug(_ context.Context, slug string) (*pageDomain.Page, error) {
	return s.bySlug[slug], nil
}
func (s *pagesRepoStub) Update(_ context.Context, page *pageDomain.Page) error {
	if page == nil {
		return errors.New("nil")
	}
	s.byID[page.ID] = page
	s.bySlug[page.Slug] = page
	return nil
}
func (s *pagesRepoStub) Delete(_ context.Context, id pageDomain.PageID) error {
	delete(s.byID, id)
	return nil
}

func TestCreatePage_NormalizesSlugAndPersists(t *testing.T) {
	repo := &pagesRepoStub{byID: map[pageDomain.PageID]*pageDomain.Page{}, bySlug: map[string]*pageDomain.Page{}}
	svc := NewService(repo)

	p, err := svc.Create(context.Background(), "author-1", "About", " About Us ", "content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Slug != "about-us" {
		t.Fatalf("expected normalized slug about-us, got %q", p.Slug)
	}
}

func TestCreatePage_DuplicateSlug(t *testing.T) {
	repo := &pagesRepoStub{
		byID:   map[pageDomain.PageID]*pageDomain.Page{},
		bySlug: map[string]*pageDomain.Page{"about": {ID: "1", Slug: "about"}},
	}
	svc := NewService(repo)

	_, err := svc.Create(context.Background(), "author-1", "About", "about", "content")
	if !errors.Is(err, ErrSlugTaken) {
		t.Fatalf("expected ErrSlugTaken, got %v", err)
	}
}

func TestUpdatePage_InvalidStatus(t *testing.T) {
	repo := &pagesRepoStub{
		byID:   map[pageDomain.PageID]*pageDomain.Page{"1": {ID: "1", Slug: "about", Title: "About"}},
		bySlug: map[string]*pageDomain.Page{"about": {ID: "1", Slug: "about", Title: "About"}},
	}
	svc := NewService(repo)

	_, err := svc.Update(context.Background(), "1", "", "", "", "bad-status")
	if !errors.Is(err, ErrInvalidStatus) {
		t.Fatalf("expected ErrInvalidStatus, got %v", err)
	}
}
