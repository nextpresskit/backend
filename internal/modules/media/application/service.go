package application

import (
	"context"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	mediaDomain "github.com/nextpresskit/backend/internal/modules/media/domain"
)

var (
	ErrInvalidUpload = errors.New("invalid_upload")
	ErrNotFound      = errors.New("media_not_found")
)

type Service struct {
	repo    mediaDomain.Repository
	storage Storage
}

func NewService(repo mediaDomain.Repository, storage Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

func (s *Service) Upload(ctx context.Context, uploaderID, originalName, mimeType string, r io.Reader) (*mediaDomain.Media, error) {
	uploaderID = strings.TrimSpace(uploaderID)
	originalName = strings.TrimSpace(originalName)
	mimeType = strings.TrimSpace(mimeType)
	if uploaderID == "" || originalName == "" || mimeType == "" || r == nil {
		return nil, ErrInvalidUpload
	}

	sf, err := s.storage.Save(ctx, originalName, mimeType, r)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	m := &mediaDomain.Media{
		UUID:         uuid.NewString(),
		UploaderID:   uploaderID,
		OriginalName: originalName,
		StorageName:  sf.StorageName,
		MimeType:     mimeType,
		SizeBytes:    sf.SizeBytes,
		StoragePath:  sf.StoragePath,
		PublicURL:    sf.PublicURL,
		CreatedAt:    now,
	}

	if err := s.repo.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Service) GetByID(ctx context.Context, id string) (*mediaDomain.Media, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil, ErrNotFound
	}
	if idNum, err := strconv.ParseInt(id, 10, 64); err == nil && idNum > 0 {
		m, err := s.repo.FindByID(ctx, mediaDomain.MediaID(idNum))
		if err != nil {
			return nil, err
		}
		if m != nil {
			return m, nil
		}
	}
	m, err := s.repo.FindByUUID(ctx, id)
	if err != nil {
		return nil, err
	}
	if m == nil {
		return nil, ErrNotFound
	}
	return m, nil
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]mediaDomain.Media, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.List(ctx, limit, offset)
}

