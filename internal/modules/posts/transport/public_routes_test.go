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

	platformMiddleware "github.com/Petar-V-Nikolov/nextpress-backend/internal/platform/middleware"

	postsApp "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/application"
	postDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain"
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

type mockPostRepo struct {
	published []postDomain.Post
}

func (m *mockPostRepo) Create(ctx context.Context, post *postDomain.Post) error {
	return nil
}
func (m *mockPostRepo) FindByID(ctx context.Context, id postDomain.PostID) (*postDomain.Post, error) {
	return nil, nil
}
func (m *mockPostRepo) FindBySlug(ctx context.Context, slug string) (*postDomain.Post, error) {
	return nil, nil
}
func (m *mockPostRepo) List(ctx context.Context, includeDeleted bool, limit int, offset int) ([]postDomain.Post, error) {
	return nil, nil
}
func (m *mockPostRepo) ListFiltered(ctx context.Context, includeDeleted bool, limit int, offset int, status string, authorID string, q string) ([]postDomain.Post, error) {
	return nil, nil
}
func (m *mockPostRepo) ListPublished(ctx context.Context, limit int, offset int, q string, categoryID string, tagID string) ([]postDomain.Post, error) {
	// Simple behavior for tests: ignore filters and return the first `limit` items.
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 {
		limit = len(m.published)
	}
	start := offset
	if start > len(m.published) {
		return []postDomain.Post{}, nil
	}
	end := start + limit
	if end > len(m.published) {
		end = len(m.published)
	}
	return m.published[start:end], nil
}
func (m *mockPostRepo) FindPublishedBySlug(ctx context.Context, slug string) (*postDomain.Post, error) {
	for i := range m.published {
		if m.published[i].Slug == slug {
			p := m.published[i]
			return &p, nil
		}
	}
	return nil, nil
}
func (m *mockPostRepo) Update(ctx context.Context, post *postDomain.Post) error {
	return nil
}
func (m *mockPostRepo) Delete(ctx context.Context, id postDomain.PostID) error {
	return nil
}
func (m *mockPostRepo) SetCategories(ctx context.Context, postID postDomain.PostID, categoryIDs []string) error {
	return nil
}
func (m *mockPostRepo) SetTags(ctx context.Context, postID postDomain.PostID, tagIDs []string) error {
	return nil
}

func TestPublicPostsRoute_ReturnsPublishedPosts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Now().UTC()
	repo := &mockPostRepo{
		published: []postDomain.Post{
			{
				ID:          postDomain.PostID("p1"),
				AuthorID:    "a1",
				Title:       "Hello",
				Slug:        "hello",
				Content:     "content",
				Status:      postDomain.StatusPublished,
				PublishedAt: &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			},
		},
	}

	svc := postsApp.NewService(repo)
	h := NewHandler(svc)

	router := gin.New()
	v1 := router.Group("/v1")
	h.RegisterPublicRoutes(v1)

	req := httptest.NewRequest(http.MethodGet, "/v1/posts?limit=1", nil)
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
	if out.Posts[0].Status != string(postDomain.StatusPublished) {
		t.Fatalf(`expected status "published", got %q`, out.Posts[0].Status)
	}
}

func TestAdminPostsRoute_RequiresAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	repo := &mockPostRepo{published: []postDomain.Post{}}
	svc := postsApp.NewService(repo)
	h := NewHandler(svc)

	parser := dummyAccessTokenParser{}
	checker := mockPermissionChecker{allowed: true}

	router := gin.New()
	v1 := router.Group("/v1")
	admin := v1.Group("/admin")
	admin.Use(platformMiddleware.AuthRequired(parser))

	h.RegisterRoutes(
		admin,
		platformMiddleware.AuthRequired(parser),
		func(code string) gin.HandlerFunc { return platformMiddleware.RequirePermission(checker, code) },
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/admin/posts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

