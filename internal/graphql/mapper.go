package graphql

import (
	"time"

	"github.com/nextpresskit/backend/internal/graphql/model"
	userdomain "github.com/nextpresskit/backend/internal/modules/user/domain"
	pagedomain "github.com/nextpresskit/backend/internal/modules/pages/domain"
	domainmodel "github.com/nextpresskit/backend/internal/modules/posts/domain/model"
	taxdomain "github.com/nextpresskit/backend/internal/modules/taxonomy/domain"
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

func domainCategoryToGQL(c *taxdomain.Category) *model.Category {
	if c == nil {
		return nil
	}
	return &model.Category{
		ID:   string(c.ID),
		Name: c.Name,
		Slug: c.Slug,
	}
}

func domainTagToGQL(t *taxdomain.Tag) *model.Tag {
	if t == nil {
		return nil
	}
	return &model.Tag{
		ID:   string(t.ID),
		Name: t.Name,
		Slug: t.Slug,
	}
}

func domainAuthUserToGQL(u *userdomain.User) *model.AuthUser {
	if u == nil {
		return nil
	}
	out := &model.AuthUser{
		ID:        string(u.UUID),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Active:    u.Active,
	}
	if !u.CreatedAt.IsZero() {
		s := u.CreatedAt.UTC().Format(time.RFC3339Nano)
		out.CreatedAt = &s
	}
	if !u.UpdatedAt.IsZero() {
		s := u.UpdatedAt.UTC().Format(time.RFC3339Nano)
		out.UpdatedAt = &s
	}
	if u.DeletedAt != nil && !u.DeletedAt.IsZero() {
		s := u.DeletedAt.UTC().Format(time.RFC3339Nano)
		out.DeletedAt = &s
	}
	return out
}
