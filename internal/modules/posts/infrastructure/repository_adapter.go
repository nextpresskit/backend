package infrastructure

import (
	"context"
	"errors"

	posterr "github.com/nextpresskit/backend/internal/modules/posts/domain"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/metrics"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/model"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/ports"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/seo"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/series"
	"gorm.io/gorm"
)

// RepositoryAdapter maps storage-specific errors (gorm) into domain errors,
// while delegating all persistence work to the underlying repository.
type RepositoryAdapter struct {
	inner ports.Repository
}

func NewRepositoryAdapter(inner ports.Repository) *RepositoryAdapter {
	return &RepositoryAdapter{inner: inner}
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return posterr.ErrNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return posterr.ErrConflict
	}
	return err
}

// --- Core post CRUD/read ---

func (a *RepositoryAdapter) Create(ctx context.Context, post *model.Post) error {
	return mapErr(a.inner.Create(ctx, post))
}

func (a *RepositoryAdapter) FindByID(ctx context.Context, id ident.PostID) (*model.Post, error) {
	p, err := a.inner.FindByID(ctx, id)
	return p, mapErr(err)
}

func (a *RepositoryAdapter) FindByUUID(ctx context.Context, uuid string) (*model.Post, error) {
	p, err := a.inner.FindByUUID(ctx, uuid)
	return p, mapErr(err)
}

func (a *RepositoryAdapter) FindBySlug(ctx context.Context, slug string) (*model.Post, error) {
	p, err := a.inner.FindBySlug(ctx, slug)
	return p, mapErr(err)
}

func (a *RepositoryAdapter) List(ctx context.Context, includeDeleted bool, limit int, offset int) ([]model.Post, error) {
	out, err := a.inner.List(ctx, includeDeleted, limit, offset)
	return out, mapErr(err)
}

func (a *RepositoryAdapter) ListFiltered(ctx context.Context, includeDeleted bool, limit int, offset int, status string, authorID string, q string) ([]model.Post, error) {
	out, err := a.inner.ListFiltered(ctx, includeDeleted, limit, offset, status, authorID, q)
	return out, mapErr(err)
}

func (a *RepositoryAdapter) ListPublished(ctx context.Context, limit int, offset int, q string, categoryID string, tagID string) ([]model.Post, error) {
	out, err := a.inner.ListPublished(ctx, limit, offset, q, categoryID, tagID)
	return out, mapErr(err)
}

func (a *RepositoryAdapter) FindPublishedBySlug(ctx context.Context, slug string) (*model.Post, error) {
	p, err := a.inner.FindPublishedBySlug(ctx, slug)
	return p, mapErr(err)
}

func (a *RepositoryAdapter) Update(ctx context.Context, post *model.Post) error {
	return mapErr(a.inner.Update(ctx, post))
}

func (a *RepositoryAdapter) Delete(ctx context.Context, id ident.PostID) error {
	return mapErr(a.inner.Delete(ctx, id))
}

// --- Taxonomy ---

func (a *RepositoryAdapter) SetCategories(ctx context.Context, postID ident.PostID, categoryIDs []string) error {
	return mapErr(a.inner.SetCategories(ctx, postID, categoryIDs))
}

func (a *RepositoryAdapter) SetTags(ctx context.Context, postID ident.PostID, tagIDs []string) error {
	return mapErr(a.inner.SetTags(ctx, postID, tagIDs))
}

func (a *RepositoryAdapter) SetPrimaryCategory(ctx context.Context, postID ident.PostID, categoryID *string) error {
	return mapErr(a.inner.SetPrimaryCategory(ctx, postID, categoryID))
}

// --- SEO / metrics ---

func (a *RepositoryAdapter) DeleteSEO(ctx context.Context, postID ident.PostID) error {
	return mapErr(a.inner.DeleteSEO(ctx, postID))
}

func (a *RepositoryAdapter) UpsertSEOOnly(ctx context.Context, postID ident.PostID, seo *seo.PostSEO) error {
	return mapErr(a.inner.UpsertSEOOnly(ctx, postID, seo))
}

func (a *RepositoryAdapter) GetMetrics(ctx context.Context, postID ident.PostID) (*metrics.PostMetrics, error) {
	m, err := a.inner.GetMetrics(ctx, postID)
	return m, mapErr(err)
}

// --- Featured image ---

func (a *RepositoryAdapter) SetFeaturedImage(ctx context.Context, postID ident.PostID, mediaID *string, alt *string, width *int, height *int, focalX *float32, focalY *float32, credit *string, license *string) error {
	return mapErr(a.inner.SetFeaturedImage(ctx, postID, mediaID, alt, width, height, focalX, focalY, credit, license))
}

// --- Series link ---

func (a *RepositoryAdapter) SetPostSeries(ctx context.Context, postID ident.PostID, seriesID *string, partIndex *int, partLabel *string) error {
	return mapErr(a.inner.SetPostSeries(ctx, postID, seriesID, partIndex, partLabel))
}

// --- Coauthors ---

func (a *RepositoryAdapter) ReplaceCoauthors(ctx context.Context, postID ident.PostID, userIDs []string) error {
	return mapErr(a.inner.ReplaceCoauthors(ctx, postID, userIDs))
}

// --- Gallery ---

func (a *RepositoryAdapter) CreateGalleryItem(ctx context.Context, postID ident.PostID, mediaID string, sortOrder int, caption *string, alt *string) (string, error) {
	id, err := a.inner.CreateGalleryItem(ctx, postID, mediaID, sortOrder, caption, alt)
	return id, mapErr(err)
}

func (a *RepositoryAdapter) UpdateGalleryItem(ctx context.Context, postID ident.PostID, itemID string, sortOrder *int, caption *string, alt *string) error {
	return mapErr(a.inner.UpdateGalleryItem(ctx, postID, itemID, sortOrder, caption, alt))
}

func (a *RepositoryAdapter) DeleteGalleryItem(ctx context.Context, postID ident.PostID, itemID string) error {
	return mapErr(a.inner.DeleteGalleryItem(ctx, postID, itemID))
}

// --- Changelog ---

func (a *RepositoryAdapter) CreateChangelog(ctx context.Context, postID ident.PostID, userID *string, note string) (string, error) {
	id, err := a.inner.CreateChangelog(ctx, postID, userID, note)
	return id, mapErr(err)
}

func (a *RepositoryAdapter) DeleteChangelog(ctx context.Context, postID ident.PostID, changelogID string) error {
	return mapErr(a.inner.DeleteChangelog(ctx, postID, changelogID))
}

// --- Syndication ---

func (a *RepositoryAdapter) CreateSyndication(ctx context.Context, postID ident.PostID, platform, url, status string) (string, error) {
	id, err := a.inner.CreateSyndication(ctx, postID, platform, url, status)
	return id, mapErr(err)
}

func (a *RepositoryAdapter) UpdateSyndication(ctx context.Context, postID ident.PostID, id string, platform, url, status *string) error {
	return mapErr(a.inner.UpdateSyndication(ctx, postID, id, platform, url, status))
}

func (a *RepositoryAdapter) DeleteSyndication(ctx context.Context, postID ident.PostID, id string) error {
	return mapErr(a.inner.DeleteSyndication(ctx, postID, id))
}

func (a *RepositoryAdapter) UpdateSyndicationByID(ctx context.Context, id string, platform, url, status *string) error {
	return mapErr(a.inner.UpdateSyndicationByID(ctx, id, platform, url, status))
}

func (a *RepositoryAdapter) DeleteSyndicationByID(ctx context.Context, id string) error {
	return mapErr(a.inner.DeleteSyndicationByID(ctx, id))
}

// --- Translations ---

func (a *RepositoryAdapter) PutPostTranslation(ctx context.Context, postID ident.PostID, groupID *string, locale string) (string, error) {
	gid, err := a.inner.PutPostTranslation(ctx, postID, groupID, locale)
	return gid, mapErr(err)
}

func (a *RepositoryAdapter) ClearPostTranslation(ctx context.Context, postID ident.PostID) error {
	return mapErr(a.inner.ClearPostTranslation(ctx, postID))
}

// --- Series (top-level) ---

func (a *RepositoryAdapter) ListSeries(ctx context.Context) ([]series.Series, error) {
	out, err := a.inner.ListSeries(ctx)
	return out, mapErr(err)
}

func (a *RepositoryAdapter) CreateSeries(ctx context.Context, s *series.Series) error {
	return mapErr(a.inner.CreateSeries(ctx, s))
}

func (a *RepositoryAdapter) FindSeriesByID(ctx context.Context, id string) (*series.Series, error) {
	s, err := a.inner.FindSeriesByID(ctx, id)
	return s, mapErr(err)
}

func (a *RepositoryAdapter) UpdateSeries(ctx context.Context, s *series.Series) error {
	return mapErr(a.inner.UpdateSeries(ctx, s))
}

func (a *RepositoryAdapter) DeleteSeries(ctx context.Context, id string) error {
	return mapErr(a.inner.DeleteSeries(ctx, id))
}

// --- Translation groups (top-level) ---

func (a *RepositoryAdapter) CreateTranslationGroup(ctx context.Context, id string) error {
	return mapErr(a.inner.CreateTranslationGroup(ctx, id))
}

func (a *RepositoryAdapter) FindTranslationGroup(ctx context.Context, id string) (bool, error) {
	ok, err := a.inner.FindTranslationGroup(ctx, id)
	return ok, mapErr(err)
}

func (a *RepositoryAdapter) DeleteTranslationGroup(ctx context.Context, id string) error {
	return mapErr(a.inner.DeleteTranslationGroup(ctx, id))
}

var _ ports.Repository = (*RepositoryAdapter)(nil)
