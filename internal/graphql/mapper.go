package graphql

import (
	"time"

	"github.com/Petar-V-Nikolov/nextpress-backend/internal/graphql/model"
	userdomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/user/domain"
	menudomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/menus/domain"
	pagedomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/pages/domain"
	domainmodel "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain/model"
	taxdomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/taxonomy/domain"
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

func domainMenuItemToGQL(i *menudomain.MenuItem) *model.MenuItem {
	if i == nil {
		return nil
	}
	var parentID *string
	if i.ParentID != nil {
		v := string(*i.ParentID)
		parentID = &v
	}
	return &model.MenuItem{
		ID:       string(i.ID),
		ParentID: parentID,
		Label:    i.Label,
		ItemType: string(i.ItemType),
		RefID:    i.RefID,
		URL:      i.URL,
		SortOrder: i.SortOrder,
	}
}

func domainMenuToGQL(m *menudomain.Menu, items []menudomain.MenuItem) *model.Menu {
	if m == nil {
		return nil
	}
	out := make([]*model.MenuItem, 0, len(items))
	for i := range items {
		out = append(out, domainMenuItemToGQL(&items[i]))
	}
	return &model.Menu{
		ID:    string(m.ID),
		Name:  m.Name,
		Slug:  m.Slug,
		Items: out,
	}
}

func domainAuthUserToGQL(u *userdomain.User) *model.AuthUser {
	if u == nil {
		return nil
	}
	return &model.AuthUser{
		ID:        string(u.ID),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
	}
}
