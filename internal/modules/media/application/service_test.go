package application

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	mediaDomain "github.com/nextpresskit/backend/internal/modules/media/domain"
)

type mediaRepoStub struct {
	byID map[mediaDomain.MediaID]*mediaDomain.Media
}

func (s *mediaRepoStub) Create(_ context.Context, m *mediaDomain.Media) error {
	if s.byID == nil {
		s.byID = map[mediaDomain.MediaID]*mediaDomain.Media{}
	}
	cp := *m
	s.byID[m.ID] = &cp
	return nil
}

func (s *mediaRepoStub) FindByID(_ context.Context, id mediaDomain.MediaID) (*mediaDomain.Media, error) {
	return s.byID[id], nil
}

func (s *mediaRepoStub) List(_ context.Context, _, _ int) ([]mediaDomain.Media, error) {
	out := make([]mediaDomain.Media, 0, len(s.byID))
	for _, v := range s.byID {
		out = append(out, *v)
	}
	return out, nil
}

type mediaStorageStub struct {
	saveErr error
}

func (s mediaStorageStub) Save(_ context.Context, _ string, _ string, _ io.Reader) (*StoredFile, error) {
	if s.saveErr != nil {
		return nil, s.saveErr
	}
	return &StoredFile{
		StorageName: "stored.bin",
		StoragePath: "/tmp/stored.bin",
		PublicURL:   "/uploads/stored.bin",
		SizeBytes:   5,
	}, nil
}

func TestUpload_Validation(t *testing.T) {
	svc := NewService(&mediaRepoStub{}, mediaStorageStub{})

	_, err := svc.Upload(context.Background(), "", "file.png", "image/png", strings.NewReader("data"))
	if !errors.Is(err, ErrInvalidUpload) {
		t.Fatalf("expected ErrInvalidUpload, got %v", err)
	}
}

func TestUpload_Success(t *testing.T) {
	repo := &mediaRepoStub{byID: map[mediaDomain.MediaID]*mediaDomain.Media{}}
	svc := NewService(repo, mediaStorageStub{})

	m, err := svc.Upload(context.Background(), "user-1", "file.png", "image/png", strings.NewReader("data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m == nil || m.StorageName != "stored.bin" {
		t.Fatalf("expected stored media result, got %#v", m)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := NewService(&mediaRepoStub{byID: map[mediaDomain.MediaID]*mediaDomain.Media{}}, mediaStorageStub{})
	_, err := svc.GetByID(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

