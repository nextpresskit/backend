package graphql

import (
	"time"

	"github.com/Petar-V-Nikolov/nextpress-backend/internal/graphql/model"
	pagedomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/pages/domain"
	domainmodel "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain/model"
)

func domainPageToGQL(p *pagedomain.Page) *model.Page {
	if p == nil {
		return nil
	}
	var pub *string
	if p.PublishedAt != nil {
		s := p.PublishedAt.UTC().Format(time.RFC3339)
		pub = &s
	}
	return &model.Page{
		ID:          string(p.ID),
		Title:       p.Title,
		Slug:        p.Slug,
		Status:      string(p.Status),
		PublishedAt: pub,
	}
}

func domainPostToGQL(p *domainmodel.Post) *model.Post {
	if p == nil {
		return nil
	}
	var pub *string
	if p.PublishedAt != nil {
		s := p.PublishedAt.UTC().Format(time.RFC3339)
		pub = &s
	}
	var ex *string
	if p.Excerpt != "" {
		e := p.Excerpt
		ex = &e
	}
	return &model.Post{
		ID:          string(p.ID),
		Title:       p.Title,
		Slug:        p.Slug,
		Excerpt:     ex,
		Status:      string(p.Status),
		PublishedAt: pub,
	}
}
