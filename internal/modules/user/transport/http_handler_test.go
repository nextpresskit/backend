package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	userApp "github.com/nextpresskit/backend/internal/modules/user/application"
	"github.com/nextpresskit/backend/internal/modules/user/domain"
	platformMiddleware "github.com/nextpresskit/backend/internal/platform/middleware"
)

type fakeReader struct {
	list    userApp.ListResult
	listErr error
	get     *userApp.UserWithRoles
	getErr  error

	gotQuery  string
	gotLimit  int
	gotOffset int
	gotID     string
}

func (f *fakeReader) List(_ context.Context, q string, limit, offset int) (userApp.ListResult, error) {
	f.gotQuery, f.gotLimit, f.gotOffset = q, limit, offset
	return f.list, f.listErr
}

func (f *fakeReader) Get(_ context.Context, raw string) (*userApp.UserWithRoles, error) {
	f.gotID = raw
	return f.get, f.getErr
}

type fakeChecker struct {
	allowed bool
	gotCode string
}

func (f *fakeChecker) UserHasPermission(_ context.Context, _ string, code string) (bool, error) {
	f.gotCode = code
	return f.allowed, nil
}

func newRouter(svc UsersReader, checker *fakeChecker, authed bool) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	admin := r.Group("/admin", func(c *gin.Context) {
		if authed {
			c.Set(platformMiddleware.ContextUserIDKey, "1")
		}
		c.Next()
	})
	NewHandler(svc).RegisterRoutes(admin, func(code string) gin.HandlerFunc {
		return platformMiddleware.RequirePermission(checker, code)
	})
	return r
}

func sampleUser() userApp.UserWithRoles {
	return userApp.UserWithRoles{
		User: domain.User{
			ID: 1, UUID: "00000000-0000-0000-0100-000000000001", FirstName: "Super", LastName: "Admin",
			Email: "superadmin@example.test", Password: "hash-must-not-leak", Active: true,
			CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		},
		Roles: []string{"admin", "superadmin"},
	}
}

func TestAdminUsersRoutes(t *testing.T) {
	u := sampleUser()
	tests := []struct {
		name       string
		path       string
		authed     bool
		allowed    bool
		reader     *fakeReader
		wantStatus int
		wantError  string
		check      func(t *testing.T, body map[string]any, r *fakeReader)
	}{
		{
			name: "list ok", path: "/admin/users?q=adm&limit=10&offset=20", authed: true, allowed: true,
			reader:     &fakeReader{list: userApp.ListResult{Users: []userApp.UserWithRoles{u}, Total: 31, Limit: 10, Offset: 20}},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, body map[string]any, r *fakeReader) {
				if r.gotQuery != "adm" || r.gotLimit != 10 || r.gotOffset != 20 {
					t.Fatalf("args = %q %d %d", r.gotQuery, r.gotLimit, r.gotOffset)
				}
				if body["total"].(float64) != 31 || body["limit"].(float64) != 10 || body["offset"].(float64) != 20 {
					t.Fatalf("meta = %v", body)
				}
				users := body["users"].([]any)
				first := users[0].(map[string]any)
				if first["email"] != "superadmin@example.test" || first["firstName"] != "Super" {
					t.Fatalf("user = %v", first)
				}
				if roles := first["roles"].([]any); len(roles) != 2 || roles[0] != "admin" {
					t.Fatalf("roles = %v", roles)
				}
				if first["createdAt"] != "2026-01-02T03:04:05Z" {
					t.Fatalf("createdAt = %v", first["createdAt"])
				}
			},
		},
		{
			name: "list empty is an array", path: "/admin/users", authed: true, allowed: true,
			reader:     &fakeReader{list: userApp.ListResult{Limit: 20}},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, body map[string]any, _ *fakeReader) {
				if users, ok := body["users"].([]any); !ok || len(users) != 0 {
					t.Fatalf("users = %#v", body["users"])
				}
			},
		},
		{name: "list bad limit", path: "/admin/users?limit=ten", authed: true, allowed: true, reader: &fakeReader{}, wantStatus: http.StatusBadRequest, wantError: "invalid_pagination"},
		{name: "list negative offset", path: "/admin/users?offset=-1", authed: true, allowed: true, reader: &fakeReader{}, wantStatus: http.StatusBadRequest, wantError: "invalid_pagination"},
		{name: "list service error", path: "/admin/users", authed: true, allowed: true, reader: &fakeReader{listErr: errors.New("db down")}, wantStatus: http.StatusInternalServerError, wantError: "internal_error"},
		{name: "list unauthenticated", path: "/admin/users", reader: &fakeReader{}, wantStatus: http.StatusUnauthorized, wantError: "missing_user_context"},
		{name: "list forbidden", path: "/admin/users", authed: true, reader: &fakeReader{}, wantStatus: http.StatusForbidden, wantError: "forbidden"},
		{
			name: "get ok", path: "/admin/users/1", authed: true, allowed: true,
			reader:     &fakeReader{get: &u},
			wantStatus: http.StatusOK,
			check: func(t *testing.T, body map[string]any, r *fakeReader) {
				if r.gotID != "1" {
					t.Fatalf("id = %q", r.gotID)
				}
				user := body["user"].(map[string]any)
				if user["uuid"] != u.User.UUID || user["active"] != true {
					t.Fatalf("user = %v", user)
				}
			},
		},
		{name: "get not found", path: "/admin/users/99", authed: true, allowed: true, reader: &fakeReader{getErr: userApp.ErrUserNotFound}, wantStatus: http.StatusNotFound, wantError: "user_not_found"},
		{name: "get invalid id", path: "/admin/users/nope", authed: true, allowed: true, reader: &fakeReader{getErr: userApp.ErrInvalidUserID}, wantStatus: http.StatusBadRequest, wantError: "invalid_user_id"},
		{name: "get forbidden", path: "/admin/users/1", authed: true, reader: &fakeReader{}, wantStatus: http.StatusForbidden, wantError: "forbidden"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &fakeChecker{allowed: tt.allowed}
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			newRouter(tt.reader, checker, tt.authed).ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d (%s)", w.Code, tt.wantStatus, w.Body.String())
			}
			if strings.Contains(w.Body.String(), "hash-must-not-leak") {
				t.Fatal("password hash leaked into response")
			}
			if tt.authed && checker.gotCode != ReadPermission {
				t.Fatalf("permission checked = %q, want %q", checker.gotCode, ReadPermission)
			}
			var body map[string]any
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid json: %v", err)
			}
			if tt.wantError != "" && body["error"] != tt.wantError {
				t.Fatalf("error = %v, want %q", body["error"], tt.wantError)
			}
			if tt.check != nil {
				tt.check(t, body, tt.reader)
			}
		})
	}
}
