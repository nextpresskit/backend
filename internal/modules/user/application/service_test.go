package application

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/nextpresskit/backend/internal/modules/user/domain"
)

type fakeRepo struct {
	byID   map[int64]domain.User
	byUUID map[string]domain.User
	err    error
}

func (f *fakeRepo) FindByID(id domain.UserID) (*domain.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if u, ok := f.byID[int64(id)]; ok {
		return &u, nil
	}
	return nil, nil
}

func (f *fakeRepo) FindByUUID(id string) (*domain.User, error) {
	if f.err != nil {
		return nil, f.err
	}
	if u, ok := f.byUUID[id]; ok {
		return &u, nil
	}
	return nil, nil
}

func (f *fakeRepo) FindByEmail(string) (*domain.User, error) { return nil, nil }
func (f *fakeRepo) Create(*domain.User) error                { return nil }
func (f *fakeRepo) Update(*domain.User) error                { return nil }
func (f *fakeRepo) Delete(domain.UserID) error               { return nil }

type fakeLister struct {
	users []domain.User
	total int64
	err   error
	got   domain.ListFilter
}

func (f *fakeLister) List(_ context.Context, filter domain.ListFilter) ([]domain.User, int64, error) {
	f.got = filter
	return f.users, f.total, f.err
}

type fakeRoles struct {
	roles map[int64][]string
	err   error
	calls int
}

func (f *fakeRoles) RoleNamesByUserIDs(_ context.Context, _ []int64) (map[int64][]string, error) {
	f.calls++
	return f.roles, f.err
}

func TestServiceList(t *testing.T) {
	users := []domain.User{{ID: 1, Email: "a@example.test"}, {ID: 2, Email: "b@example.test"}}
	errBoom := errors.New("boom")
	long := make([]byte, MaxQueryLength+20)
	for i := range long {
		long[i] = 'x'
	}

	tests := []struct {
		name       string
		query      string
		limit      int
		offset     int
		lister     *fakeLister
		roles      *fakeRoles
		wantErr    error
		wantFilter domain.ListFilter
		wantRoles  [][]string
	}{
		{
			name:       "defaults limit and attaches roles",
			lister:     &fakeLister{users: users, total: 2},
			roles:      &fakeRoles{roles: map[int64][]string{1: {"admin", "superadmin"}}},
			wantFilter: domain.ListFilter{Limit: DefaultListLimit},
			wantRoles:  [][]string{{"admin", "superadmin"}, {}},
		},
		{
			name:       "caps limit and trims query",
			query:      "  ada  ",
			limit:      500,
			offset:     40,
			lister:     &fakeLister{},
			roles:      &fakeRoles{},
			wantFilter: domain.ListFilter{Query: "ada", Limit: MaxListLimit, Offset: 40},
			wantRoles:  [][]string{},
		},
		{
			name:       "truncates very long query",
			query:      string(long),
			limit:      5,
			lister:     &fakeLister{},
			wantFilter: domain.ListFilter{Query: string(long[:MaxQueryLength]), Limit: 5},
			wantRoles:  [][]string{},
		},
		{
			name:      "nil role reader yields empty roles",
			lister:    &fakeLister{users: users[:1], total: 1},
			wantRoles: [][]string{{}},
			wantFilter: domain.ListFilter{
				Limit: DefaultListLimit,
			},
		},
		{name: "negative limit", limit: -1, lister: &fakeLister{}, wantErr: ErrInvalidListArg},
		{name: "negative offset", offset: -3, lister: &fakeLister{}, wantErr: ErrInvalidListArg},
		{name: "lister error", lister: &fakeLister{err: errBoom}, wantErr: errBoom},
		{name: "role error", lister: &fakeLister{users: users}, roles: &fakeRoles{err: errBoom}, wantErr: errBoom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var roles domain.RoleReader
			if tt.roles != nil {
				roles = tt.roles
			}
			svc := NewService(&fakeRepo{}, tt.lister, roles)
			res, err := svc.List(context.Background(), tt.query, tt.limit, tt.offset)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.lister.got != tt.wantFilter {
				t.Fatalf("filter = %+v, want %+v", tt.lister.got, tt.wantFilter)
			}
			if res.Limit != tt.wantFilter.Limit || res.Offset != tt.wantFilter.Offset || res.Total != tt.lister.total {
				t.Fatalf("page meta = %d/%d/%d", res.Limit, res.Offset, res.Total)
			}
			got := make([][]string, 0, len(res.Users))
			for _, u := range res.Users {
				got = append(got, u.Roles)
			}
			if !reflect.DeepEqual(got, tt.wantRoles) {
				t.Fatalf("roles = %v, want %v", got, tt.wantRoles)
			}
		})
	}
}

func TestServiceGet(t *testing.T) {
	const id = "8a7c3b8e-3f53-4d3a-9d0e-2b8c4f1a9e10"
	u := domain.User{ID: 7, UUID: id, Email: "g@example.test"}
	repo := &fakeRepo{byID: map[int64]domain.User{7: u}, byUUID: map[string]domain.User{id: u}}
	errBoom := errors.New("boom")

	tests := []struct {
		name    string
		raw     string
		repo    *fakeRepo
		wantErr error
		wantID  int64
	}{
		{name: "numeric id", raw: "7", repo: repo, wantID: 7},
		{name: "uuid", raw: id, repo: repo, wantID: 7},
		{name: "missing numeric", raw: "8", repo: repo, wantErr: ErrUserNotFound},
		{name: "missing uuid", raw: "00000000-0000-0000-0000-00000000ffff", repo: repo, wantErr: ErrUserNotFound},
		{name: "zero id", raw: "0", repo: repo, wantErr: ErrInvalidUserID},
		{name: "garbage", raw: "abc", repo: repo, wantErr: ErrInvalidUserID},
		{name: "repo error", raw: "7", repo: &fakeRepo{err: errBoom}, wantErr: errBoom},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles := &fakeRoles{roles: map[int64][]string{7: {"admin"}}}
			got, err := NewService(tt.repo, nil, roles).Get(context.Background(), tt.raw)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.User.ID != tt.wantID || !reflect.DeepEqual(got.Roles, []string{"admin"}) {
				t.Fatalf("got %+v", got)
			}
		})
	}
}
