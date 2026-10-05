package application

import (
	"context"
	"errors"
	"testing"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/model"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/ports"
)

// lookupRepo fakes only the two finders; any other call panics via the nil embed.
type lookupRepo struct {
	ports.CorePostsPersistence
	byID      map[int64]*model.Post
	byUUID    map[string]*model.Post
	uuidCalls int
}

func (r *lookupRepo) FindByID(_ context.Context, id ident.PostID) (*model.Post, error) {
	return r.byID[int64(id)], nil
}

func (r *lookupRepo) FindByUUID(_ context.Context, u string) (*model.Post, error) {
	r.uuidCalls++
	return r.byUUID[u], nil
}

func TestCorePostsServiceGetByIDLookup(t *testing.T) {
	const knownUUID = "00000000-0000-0000-0700-000000000001"
	post := &model.Post{ID: 1, UUID: knownUUID, Title: "One"}

	tests := []struct {
		name          string
		id            string
		wantFound     bool
		wantUUIDCalls int
	}{
		{name: "numeric id", id: "1", wantFound: true},
		{name: "missing numeric id is not retried as uuid", id: "999999"},
		{name: "zero id", id: "0"},
		{name: "uuid", id: knownUUID, wantFound: true, wantUUIDCalls: 1},
		{name: "missing uuid", id: "00000000-0000-0000-0700-0000000fffff", wantUUIDCalls: 1},
		{name: "garbage never reaches the uuid column", id: "abc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &lookupRepo{
				byID:   map[int64]*model.Post{1: post},
				byUUID: map[string]*model.Post{knownUUID: post},
			}
			svc := NewCorePostsService(repo, nil)
			got, err := svc.GetByID(context.Background(), tt.id)
			if tt.wantFound {
				if err != nil || got != post {
					t.Fatalf("got %v, %v; want the post", got, err)
				}
			} else if !errors.Is(err, ErrPostNotFound) {
				t.Fatalf("err = %v, want ErrPostNotFound", err)
			}
			if repo.uuidCalls != tt.wantUUIDCalls {
				t.Fatalf("FindByUUID calls = %d, want %d", repo.uuidCalls, tt.wantUUIDCalls)
			}
		})
	}
}
