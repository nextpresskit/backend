package application

import (
	"errors"

	posterr "github.com/nextpresskit/backend/internal/modules/posts/domain"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/ports"
)

var (
	ErrInvalidPost   = errors.New("invalid_post")
	ErrSlugTaken     = errors.New("slug_taken")
	ErrPostNotFound  = errors.New("post_not_found")
	ErrInvalidStatus = errors.New("invalid_status")
	// Prefer these domain errors for application/storage boundary mapping.
	ErrNotFound        = posterr.ErrNotFound
	ErrInvalidArgument = posterr.ErrInvalidArgument
	ErrConflict        = posterr.ErrConflict
	// ErrInvalidSubresource is an alias for ErrInvalidArgument (legacy name).
	ErrInvalidSubresource = posterr.ErrInvalidArgument
)

// Service is the façade over focused posts sub-services. Method names are promoted via embedding.
type Service struct {
	*CorePostsService
	*PostSubresourcesService
	*SeriesService
	*TranslationGroupsService
}

// NewService constructs the posts module application layer.
func NewService(repo ports.Repository, hooks ports.PostSave) *Service {
	return &Service{
		CorePostsService:         NewCorePostsService(repo, hooks),
		PostSubresourcesService:  NewPostSubresourcesService(PostSubresourceStoresFrom(repo)),
		SeriesService:            NewSeriesService(repo),
		TranslationGroupsService: NewTranslationGroupsService(repo),
	}
}
