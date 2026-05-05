package application

import (
	"context"
	"errors"
	"strconv"
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

func (s *PostSubresourcesService) requirePostID(ctx context.Context, postUUID string) (ident.PostID, error) {
	postUUID = strings.TrimSpace(postUUID)
	if postUUID == "" {
		return 0, ErrPostNotFound
	}
	if postID, parseErr := strconv.ParseInt(postUUID, 10, 64); parseErr == nil && postID > 0 {
		p, err := s.stores.Reader.FindByID(ctx, ident.PostID(postID))
		if err != nil {
			return 0, err
		}
		if p != nil {
			return p.ID, nil
		}
	}
	p, err := s.stores.Reader.FindByUUID(ctx, postUUID)
	if err != nil {
		return 0, err
	}
	if p == nil {
		return 0, ErrPostNotFound
	}
	return p.ID, nil
}

func (s *PostSubresourcesService) GetMetricsForPost(ctx context.Context, postID string) (*metrics.PostMetrics, error) {
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return nil, err
	}
	return s.stores.Metrics.GetMetrics(ctx, pid)
}

func (s *PostSubresourcesService) DeleteSEO(ctx context.Context, postID string) error {
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return err
	}
	return s.stores.SEO.DeleteSEO(ctx, pid)
}

func (s *PostSubresourcesService) UpsertSEO(ctx context.Context, postID string, doc *seo.PostSEO) error {
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return err
	}
	if doc == nil {
		return ErrInvalidArgument
	}
	return s.stores.SEO.UpsertSEOOnly(ctx, pid, doc)
}

func (s *PostSubresourcesService) SetFeaturedImage(ctx context.Context, postID string, mediaID *string, alt *string, width *int, height *int, focalX *float32, focalY *float32, credit *string, license *string) error {
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return err
	}
	return s.stores.Featured.SetFeaturedImage(ctx, pid, mediaID, alt, width, height, focalX, focalY, credit, license)
}

func (s *PostSubresourcesService) SetPostSeries(ctx context.Context, postID string, seriesID *string, partIndex *int, partLabel *string) error {
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return err
	}
	return s.stores.SeriesLink.SetPostSeries(ctx, pid, seriesID, partIndex, partLabel)
}

func (s *PostSubresourcesService) ReplaceCoauthors(ctx context.Context, postID string, userIDs []string) error {
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return err
	}
	return s.stores.Coauthors.ReplaceCoauthors(ctx, pid, userIDs)
}

func (s *PostSubresourcesService) CreateGalleryItem(ctx context.Context, postID, mediaID string, sortOrder int, caption *string, alt *string) (string, error) {
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return "", err
	}
	mediaID = strings.TrimSpace(mediaID)
	if mediaID == "" {
		return "", ErrInvalidArgument
	}
	id, err := s.stores.Gallery.CreateGalleryItem(ctx, pid, mediaID, sortOrder, caption, alt)
	if err != nil {
		if err.Error() == "media_id_required" {
			return "", ErrInvalidArgument
		}
		return "", err
	}
	return id, nil
}

func (s *PostSubresourcesService) UpdateGalleryItem(ctx context.Context, postID, itemID string, sortOrder *int, caption *string, alt *string) error {
	itemID = strings.TrimSpace(itemID)
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return err
	}
	if sortOrder == nil && caption == nil && alt == nil {
		return ErrInvalidArgument
	}
	err = s.stores.Gallery.UpdateGalleryItem(ctx, pid, itemID, sortOrder, caption, alt)
	if errors.Is(err, posterr.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *PostSubresourcesService) DeleteGalleryItem(ctx context.Context, postID, itemID string) error {
	itemID = strings.TrimSpace(itemID)
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return err
	}
	return s.stores.Gallery.DeleteGalleryItem(ctx, pid, itemID)
}

func (s *PostSubresourcesService) CreateChangelog(ctx context.Context, postID string, userID *string, note string) (string, error) {
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return "", err
	}
	note = strings.TrimSpace(note)
	if note == "" {
		return "", ErrInvalidArgument
	}
	id, err := s.stores.Changelog.CreateChangelog(ctx, pid, userID, note)
	if err != nil {
		if err.Error() == "note_required" {
			return "", ErrInvalidArgument
		}
		return "", err
	}
	return id, nil
}

func (s *PostSubresourcesService) DeleteChangelog(ctx context.Context, postID, changelogID string) error {
	changelogID = strings.TrimSpace(changelogID)
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return err
	}
	err = s.stores.Changelog.DeleteChangelog(ctx, pid, changelogID)
	if errors.Is(err, posterr.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *PostSubresourcesService) CreateSyndication(ctx context.Context, postID, platform, url, status string) (string, error) {
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return "", err
	}
	id, err := s.stores.Syndication.CreateSyndication(ctx, pid, platform, url, status)
	if err != nil {
		if err.Error() == "platform_url_required" {
			return "", ErrInvalidArgument
		}
		return "", err
	}
	return id, nil
}

func (s *PostSubresourcesService) UpdateSyndication(ctx context.Context, postID, syndicationID string, platform, url, status *string) error {
	syndicationID = strings.TrimSpace(syndicationID)
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return err
	}
	err = s.stores.Syndication.UpdateSyndication(ctx, pid, syndicationID, platform, url, status)
	if errors.Is(err, posterr.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *PostSubresourcesService) DeleteSyndication(ctx context.Context, postID, syndicationID string) error {
	syndicationID = strings.TrimSpace(syndicationID)
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return err
	}
	err = s.stores.Syndication.DeleteSyndication(ctx, pid, syndicationID)
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
	locale = strings.TrimSpace(locale)
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return "", err
	}
	if locale == "" {
		return "", ErrInvalidArgument
	}
	resolved, err := s.stores.Translations.PutPostTranslation(ctx, pid, groupID, locale)
	if errors.Is(err, posterr.ErrNotFound) {
		return "", ErrNotFound
	}
	return resolved, err
}

func (s *PostSubresourcesService) ClearPostTranslation(ctx context.Context, postID string) error {
	pid, err := s.requirePostID(ctx, postID)
	if err != nil {
		return err
	}
	return s.stores.Translations.ClearPostTranslation(ctx, pid)
}
