package application

import (
	"context"
	"errors"
	"strings"

	posterr "github.com/nextpresskit/backend/internal/modules/posts/domain"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/metrics"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/seo"
)

// PostSubresourcesService validates and orchestrates post sub-resource persistence.
type PostSubresourcesService struct {
	stores PostSubresourceStores
}

// NewPostSubresourcesService constructs the sub-resource service.
func NewPostSubresourcesService(stores PostSubresourceStores) *PostSubresourcesService {
	return &PostSubresourcesService{stores: stores}
}

func (s *PostSubresourcesService) requirePost(ctx context.Context, postID string) error {
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return ErrPostNotFound
	}
	p, err := s.stores.Reader.FindByID(ctx, ident.PostID(postID))
	if err != nil {
		return err
	}
	if p == nil {
		return ErrPostNotFound
	}
	return nil
}

func (s *PostSubresourcesService) GetMetricsForPost(ctx context.Context, postID string) (*metrics.PostMetrics, error) {
	postID = strings.TrimSpace(postID)
	if err := s.requirePost(ctx, postID); err != nil {
		return nil, err
	}
	return s.stores.Metrics.GetMetrics(ctx, ident.PostID(postID))
}

func (s *PostSubresourcesService) DeleteSEO(ctx context.Context, postID string) error {
	postID = strings.TrimSpace(postID)
	if err := s.requirePost(ctx, postID); err != nil {
		return err
	}
	return s.stores.SEO.DeleteSEO(ctx, ident.PostID(postID))
}

func (s *PostSubresourcesService) UpsertSEO(ctx context.Context, postID string, doc *seo.PostSEO) error {
	postID = strings.TrimSpace(postID)
	if err := s.requirePost(ctx, postID); err != nil {
		return err
	}
	if doc == nil {
		return ErrInvalidArgument
	}
	return s.stores.SEO.UpsertSEOOnly(ctx, ident.PostID(postID), doc)
}

func (s *PostSubresourcesService) SetFeaturedImage(ctx context.Context, postID string, mediaID *string, alt *string, width *int, height *int, focalX *float32, focalY *float32, credit *string, license *string) error {
	postID = strings.TrimSpace(postID)
	if err := s.requirePost(ctx, postID); err != nil {
		return err
	}
	return s.stores.Featured.SetFeaturedImage(ctx, ident.PostID(postID), mediaID, alt, width, height, focalX, focalY, credit, license)
}

func (s *PostSubresourcesService) SetPostSeries(ctx context.Context, postID string, seriesID *string, partIndex *int, partLabel *string) error {
	postID = strings.TrimSpace(postID)
	if err := s.requirePost(ctx, postID); err != nil {
		return err
	}
	return s.stores.SeriesLink.SetPostSeries(ctx, ident.PostID(postID), seriesID, partIndex, partLabel)
}

func (s *PostSubresourcesService) ReplaceCoauthors(ctx context.Context, postID string, userIDs []string) error {
	postID = strings.TrimSpace(postID)
	if err := s.requirePost(ctx, postID); err != nil {
		return err
	}
	return s.stores.Coauthors.ReplaceCoauthors(ctx, ident.PostID(postID), userIDs)
}

func (s *PostSubresourcesService) CreateGalleryItem(ctx context.Context, postID, mediaID string, sortOrder int, caption *string, alt *string) (string, error) {
	postID = strings.TrimSpace(postID)
	if err := s.requirePost(ctx, postID); err != nil {
		return "", err
	}
	mediaID = strings.TrimSpace(mediaID)
	if mediaID == "" {
		return "", ErrInvalidArgument
	}
	id, err := s.stores.Gallery.CreateGalleryItem(ctx, ident.PostID(postID), mediaID, sortOrder, caption, alt)
	if err != nil {
		if err.Error() == "media_id_required" {
			return "", ErrInvalidArgument
		}
		return "", err
	}
	return id, nil
}

func (s *PostSubresourcesService) UpdateGalleryItem(ctx context.Context, postID, itemID string, sortOrder *int, caption *string, alt *string) error {
	postID = strings.TrimSpace(postID)
	itemID = strings.TrimSpace(itemID)
	if err := s.requirePost(ctx, postID); err != nil {
		return err
	}
	if sortOrder == nil && caption == nil && alt == nil {
		return ErrInvalidArgument
	}
	err := s.stores.Gallery.UpdateGalleryItem(ctx, ident.PostID(postID), itemID, sortOrder, caption, alt)
	if errors.Is(err, posterr.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *PostSubresourcesService) DeleteGalleryItem(ctx context.Context, postID, itemID string) error {
	postID = strings.TrimSpace(postID)
	itemID = strings.TrimSpace(itemID)
	if err := s.requirePost(ctx, postID); err != nil {
		return err
	}
	return s.stores.Gallery.DeleteGalleryItem(ctx, ident.PostID(postID), itemID)
}

func (s *PostSubresourcesService) CreateChangelog(ctx context.Context, postID string, userID *string, note string) (string, error) {
	postID = strings.TrimSpace(postID)
	if err := s.requirePost(ctx, postID); err != nil {
		return "", err
	}
	note = strings.TrimSpace(note)
	if note == "" {
		return "", ErrInvalidArgument
	}
	id, err := s.stores.Changelog.CreateChangelog(ctx, ident.PostID(postID), userID, note)
	if err != nil {
		if err.Error() == "note_required" {
			return "", ErrInvalidArgument
		}
		return "", err
	}
	return id, nil
}

func (s *PostSubresourcesService) DeleteChangelog(ctx context.Context, postID, changelogID string) error {
	postID = strings.TrimSpace(postID)
	changelogID = strings.TrimSpace(changelogID)
	if err := s.requirePost(ctx, postID); err != nil {
		return err
	}
	err := s.stores.Changelog.DeleteChangelog(ctx, ident.PostID(postID), changelogID)
	if errors.Is(err, posterr.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *PostSubresourcesService) CreateSyndication(ctx context.Context, postID, platform, url, status string) (string, error) {
	postID = strings.TrimSpace(postID)
	if err := s.requirePost(ctx, postID); err != nil {
		return "", err
	}
	id, err := s.stores.Syndication.CreateSyndication(ctx, ident.PostID(postID), platform, url, status)
	if err != nil {
		if err.Error() == "platform_url_required" {
			return "", ErrInvalidArgument
		}
		return "", err
	}
	return id, nil
}

func (s *PostSubresourcesService) UpdateSyndication(ctx context.Context, postID, syndicationID string, platform, url, status *string) error {
	postID = strings.TrimSpace(postID)
	syndicationID = strings.TrimSpace(syndicationID)
	if err := s.requirePost(ctx, postID); err != nil {
		return err
	}
	err := s.stores.Syndication.UpdateSyndication(ctx, ident.PostID(postID), syndicationID, platform, url, status)
	if errors.Is(err, posterr.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *PostSubresourcesService) DeleteSyndication(ctx context.Context, postID, syndicationID string) error {
	postID = strings.TrimSpace(postID)
	syndicationID = strings.TrimSpace(syndicationID)
	if err := s.requirePost(ctx, postID); err != nil {
		return err
	}
	err := s.stores.Syndication.DeleteSyndication(ctx, ident.PostID(postID), syndicationID)
	if errors.Is(err, posterr.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *PostSubresourcesService) UpdateSyndicationGlobal(ctx context.Context, syndicationID string, platform, url, status *string) error {
	syndicationID = strings.TrimSpace(syndicationID)
	err := s.stores.Syndication.UpdateSyndicationByID(ctx, syndicationID, platform, url, status)
	if errors.Is(err, posterr.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *PostSubresourcesService) DeleteSyndicationGlobal(ctx context.Context, syndicationID string) error {
	syndicationID = strings.TrimSpace(syndicationID)
	err := s.stores.Syndication.DeleteSyndicationByID(ctx, syndicationID)
	if errors.Is(err, posterr.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *PostSubresourcesService) PutPostTranslation(ctx context.Context, postID string, groupID *string, locale string) (resolvedGroupID string, err error) {
	postID = strings.TrimSpace(postID)
	locale = strings.TrimSpace(locale)
	if err := s.requirePost(ctx, postID); err != nil {
		return "", err
	}
	if locale == "" {
		return "", ErrInvalidArgument
	}
	resolved, err := s.stores.Translations.PutPostTranslation(ctx, ident.PostID(postID), groupID, locale)
	if errors.Is(err, posterr.ErrNotFound) {
		return "", ErrNotFound
	}
	return resolved, err
}

func (s *PostSubresourcesService) ClearPostTranslation(ctx context.Context, postID string) error {
	postID = strings.TrimSpace(postID)
	if err := s.requirePost(ctx, postID); err != nil {
		return err
	}
	return s.stores.Translations.ClearPostTranslation(ctx, ident.PostID(postID))
}
