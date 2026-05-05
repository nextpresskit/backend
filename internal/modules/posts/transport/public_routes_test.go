package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/nextpresskit/backend/internal/config"
	postsApp "github.com/nextpresskit/backend/internal/modules/posts/application"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/model"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
)

type dummyAccessTokenParser struct{}

func (p dummyAccessTokenParser) ParseAccessToken(token string) (string, error) {
	if token == "good" {
		return "user-123", nil
	}
	return "", errors.New("invalid token")
}

type mockPermissionChecker struct {
	allowed bool
	err     error
}

func (m mockPermissionChecker) UserHasPermission(_ context.Context, _ string, _ string) (bool, error) {
	return m.allowed, m.err
}

// mockPostsCore implements PostsCore for route tests (only public list/slug behaviors are non-trivial).
type mockPostsCore struct {
	published []model.Post
}

func (m *mockPostsCore) Create(_ context.Context, _, _, _, _ string) (*model.Post, error) {
	return nil, postsApp.ErrInvalidPost
}

func (m *mockPostsCore) GetByID(_ context.Context, id string) (*model.Post, error) {
	for i := range m.published {
		if m.published[i].UUID == id {
			p := m.published[i]
			return &p, nil
		}
	}
	return nil, postsApp.ErrPostNotFound
}

func (m *mockPostsCore) ListFiltered(_ context.Context, _, _ int, _, _, _ string) ([]model.Post, error) {
	return nil, nil
}

func (m *mockPostsCore) PublicList(_ context.Context, limit, offset int, _ string, _, _ string) ([]model.Post, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = len(m.published)
	}
	start := offset
	if start > len(m.published) {
		return []model.Post{}, nil
	}
	end := start + limit
	if end > len(m.published) {
		end = len(m.published)
	}
	return m.published[start:end], nil
}

func (m *mockPostsCore) PublicGetBySlug(_ context.Context, slug string) (*model.Post, error) {
	for i := range m.published {
		if m.published[i].Slug == slug {
			p := m.published[i]
			return &p, nil
		}
	}
	return nil, postsApp.ErrPostNotFound
}

func (m *mockPostsCore) ReindexPublishedForSearch(_ context.Context, sync func(context.Context, *model.Post)) (int, error) {
	var n int
	for i := range m.published {
		if m.published[i].Status == ident.StatusPublished {
			sync(context.Background(), &m.published[i])
			n++
		}
	}
	return n, nil
}

func (m *mockPostsCore) Update(_ context.Context, _, _, _, _, _ string) (*model.Post, error) {
	return nil, postsApp.ErrPostNotFound
}

func (m *mockPostsCore) Save(_ context.Context, _ *model.Post) (*model.Post, error) {
	return nil, postsApp.ErrPostNotFound
}

func (m *mockPostsCore) Delete(_ context.Context, _ string) error {
	return nil
}

func (m *mockPostsCore) SetCategories(_ context.Context, _ string, _ []string) error {
	return nil
}

func (m *mockPostsCore) SetTags(_ context.Context, _ string, _ []string) error {
	return nil
}

func (m *mockPostsCore) SetPrimaryCategory(_ context.Context, _ string, _ *string) error {
	return nil
}

func TestPublicPostsRoute_ReturnsPublishedPosts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	core := &mockPostsCore{
		published: []model.Post{
			{
				ID:          1,
				UUID:        "00000000-0000-0000-0000-0000000000a1",
				AuthorID:    "a1",
				Title:       "Hello",
				Slug:        "hello",
				Content:     "content",
				Status:      ident.StatusPublished,
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
	}

	h := NewHandler(core, stubPostsSubresources{}, stubSeriesAdmin{}, stubTranslationGroupsAdmin{})

	router := gin.New()
	api := router.Group("")
	h.RegisterPublicRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/posts?limit=1", nil)
	req.RemoteAddr = "1.2.3.4:1234"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	type postResp struct {
		Slug   string `json:"slug"`
		Status string `json:"status"`
	}
	type resp struct {
		Posts []postResp `json:"posts"`
	}

	var out resp
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("failed decoding response: %v", err)
	}
	if len(out.Posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(out.Posts))
	}
	if out.Posts[0].Slug != "hello" {
		t.Fatalf(`expected slug "hello", got %q`, out.Posts[0].Slug)
	}
	if out.Posts[0].Status != string(ident.StatusPublished) {
		t.Fatalf(`expected status "published", got %q`, out.Posts[0].Status)
	}
}

type mockESSearch struct {
	ids []string
	err error
}

func (m *mockESSearch) SearchPostIDs(_ context.Context, _ string, _, _ int) ([]string, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.ids, nil
}

func (m *mockESSearch) SyncPost(_ context.Context, _ *model.Post) {}

func TestPublicPostsSearch_Disabled(t *testing.T) {
	gin.SetMode(gin.TestMode)
	core := &mockPostsCore{}
	h := NewHandler(core, stubPostsSubresources{}, stubSeriesAdmin{}, stubTranslationGroupsAdmin{})
	router := gin.New()
	api := router.Group("")
	h.RegisterPublicRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/posts/search?q=test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("expected 501, got %d", w.Code)
	}
}

func TestPublicPostsSearch_WithElasticsearch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	now := time.Now().UTC()
	core := &mockPostsCore{
		published: []model.Post{
			{
				ID:          1,
				UUID:        "00000000-0000-0000-0000-0000000000a1",
				AuthorID:    "a1",
				Title:       "Hello",
				Slug:        "hello",
				Content:     "content",
				Status:      ident.StatusPublished,
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
	}
	es := &mockESSearch{ids: []string{"00000000-0000-0000-0000-0000000000a1"}}
	h := NewHandlerWithOptionalSearch(core, stubPostsSubresources{}, stubSeriesAdmin{}, stubTranslationGroupsAdmin{}, es)
	router := gin.New()
	api := router.Group("")
	h.RegisterPublicRoutes(api)

	req := httptest.NewRequest(http.MethodGet, "/posts/search?q=hello", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out struct {
		Posts []struct {
			Slug string `json:"slug"`
		} `json:"posts"`
	}
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if len(out.Posts) != 1 || out.Posts[0].Slug != "hello" {
		t.Fatalf("unexpected posts: %+v", out.Posts)
	}
}

func TestAdminPostsRoute_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&mockPostsCore{}, stubPostsSubresources{}, stubSeriesAdmin{}, stubTranslationGroupsAdmin{})

	parser := dummyAccessTokenParser{}
	checker := mockPermissionChecker{allowed: true}
	jwtCfg := config.JWTConfig{AuthSource: "header", AccessCookieName: "access_token"}

	router := gin.New()
	api := router.Group("")
	admin := api.Group("/admin")
	admin.Use(platformMiddleware.AuthRequired(parser, jwtCfg))

	h.RegisterRoutes(
		admin,
		platformMiddleware.AuthRequired(parser, jwtCfg),
		func(code string) gin.HandlerFunc { return platformMiddleware.RequirePermission(checker, code) },
	)

	req := httptest.NewRequest(http.MethodGet, "/admin/posts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}
