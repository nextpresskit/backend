package application

import (
	"errors"

	postDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain"
)

var (
	ErrInvalidPost       = errors.New("invalid_post")
	ErrSlugTaken         = errors.New("slug_taken")
	ErrPostNotFound      = errors.New("post_not_found")
	ErrInvalidStatus     = errors.New("invalid_status")
	// Prefer these domain errors for application/storage boundary mapping.
	ErrNotFound        = postDomain.ErrNotFound
	ErrInvalidArgument = postDomain.ErrInvalidArgument
	ErrConflict        = postDomain.ErrConflict
	// ErrInvalidSubresource is an alias for ErrInvalidArgument (legacy name).
	ErrInvalidSubresource = postDomain.ErrInvalidArgument
)

// Service is the façade over focused posts sub-services. Method names are promoted via embedding.
type Service struct {
	*CorePostsService
	*PostSubresourcesService
	*SeriesService
	*TranslationGroupsService
}

// NewService constructs the posts module application layer.
func NewService(repo postDomain.Repository, hooks postDomain.PostSave) *Service {
	return &Service{
		CorePostsService:          NewCorePostsService(repo, hooks),
		PostSubresourcesService:   NewPostSubresourcesService(PostSubresourceStoresFrom(repo)),
		SeriesService:             NewSeriesService(repo),
		TranslationGroupsService:  NewTranslationGroupsService(repo),
	}
}
