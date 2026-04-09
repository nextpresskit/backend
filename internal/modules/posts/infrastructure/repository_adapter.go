package infrastructure

import (
	"context"
	"errors"

	postDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain"
	"gorm.io/gorm"
)

// RepositoryAdapter maps storage-specific errors (gorm) into domain errors,
// while delegating all persistence work to the underlying repository.
type RepositoryAdapter struct {
	inner postDomain.Repository
}

func NewRepositoryAdapter(inner postDomain.Repository) *RepositoryAdapter {
	return &RepositoryAdapter{inner: inner}
}

func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return postDomain.ErrNotFound
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return postDomain.ErrConflict
	}
	return err
}

// --- Core post CRUD/read ---

func (a *RepositoryAdapter) Create(ctx context.Context, post *postDomain.Post) error {
	return mapErr(a.inner.Create(ctx, post))
}

func (a *RepositoryAdapter) FindByID(ctx context.Context, id postDomain.PostID) (*postDomain.Post, error) {
	p, err := a.inner.FindByID(ctx, id)
	return p, mapErr(err)
}

func (a *RepositoryAdapter) FindBySlug(ctx context.Context, slug string) (*postDomain.Post, error) {
	p, err := a.inner.FindBySlug(ctx, slug)
	return p, mapErr(err)
}

func (a *RepositoryAdapter) List(ctx context.Context, includeDeleted bool, limit int, offset int) ([]postDomain.Post, error) {
	out, err := a.inner.List(ctx, includeDeleted, limit, offset)
	return out, mapErr(err)
}

func (a *RepositoryAdapter) ListFiltered(ctx context.Context, includeDeleted bool, limit int, offset int, status string, authorID string, q string) ([]postDomain.Post, error) {
	out, err := a.inner.ListFiltered(ctx, includeDeleted, limit, offset, status, authorID, q)
	return out, mapErr(err)
}

func (a *RepositoryAdapter) ListPublished(ctx context.Context, limit int, offset int, q string, categoryID string, tagID string) ([]postDomain.Post, error) {
	out, err := a.inner.ListPublished(ctx, limit, offset, q, categoryID, tagID)
	return out, mapErr(err)
}

func (a *RepositoryAdapter) FindPublishedBySlug(ctx context.Context, slug string) (*postDomain.Post, error) {
	p, err := a.inner.FindPublishedBySlug(ctx, slug)
	return p, mapErr(err)
}

func (a *RepositoryAdapter) Update(ctx context.Context, post *postDomain.Post) error {
	return mapErr(a.inner.Update(ctx, post))
}

func (a *RepositoryAdapter) Delete(ctx context.Context, id postDomain.PostID) error {
	return mapErr(a.inner.Delete(ctx, id))
}

// --- Taxonomy ---

func (a *RepositoryAdapter) SetCategories(ctx context.Context, postID postDomain.PostID, categoryIDs []string) error {
	return mapErr(a.inner.SetCategories(ctx, postID, categoryIDs))
}

func (a *RepositoryAdapter) SetTags(ctx context.Context, postID postDomain.PostID, tagIDs []string) error {
	return mapErr(a.inner.SetTags(ctx, postID, tagIDs))
}

func (a *RepositoryAdapter) SetPrimaryCategory(ctx context.Context, postID postDomain.PostID, categoryID *string) error {
	return mapErr(a.inner.SetPrimaryCategory(ctx, postID, categoryID))
}

// --- SEO / metrics ---

func (a *RepositoryAdapter) DeleteSEO(ctx context.Context, postID postDomain.PostID) error {
	return mapErr(a.inner.DeleteSEO(ctx, postID))
}

func (a *RepositoryAdapter) UpsertSEOOnly(ctx context.Context, postID postDomain.PostID, seo *postDomain.PostSEO) error {
	return mapErr(a.inner.UpsertSEOOnly(ctx, postID, seo))
}

func (a *RepositoryAdapter) GetMetrics(ctx context.Context, postID postDomain.PostID) (*postDomain.PostMetrics, error) {
	m, err := a.inner.GetMetrics(ctx, postID)
	return m, mapErr(err)
}

// --- Featured image ---

func (a *RepositoryAdapter) SetFeaturedImage(ctx context.Context, postID postDomain.PostID, mediaID *string, alt *string, width *int, height *int, focalX *float32, focalY *float32, credit *string, license *string) error {
	return mapErr(a.inner.SetFeaturedImage(ctx, postID, mediaID, alt, width, height, focalX, focalY, credit, license))
}

// --- Series link ---

func (a *RepositoryAdapter) SetPostSeries(ctx context.Context, postID postDomain.PostID, seriesID *string, partIndex *int, partLabel *string) error {
	return mapErr(a.inner.SetPostSeries(ctx, postID, seriesID, partIndex, partLabel))
}

// --- Coauthors ---

func (a *RepositoryAdapter) ReplaceCoauthors(ctx context.Context, postID postDomain.PostID, userIDs []string) error {
	return mapErr(a.inner.ReplaceCoauthors(ctx, postID, userIDs))
}

// --- Gallery ---

func (a *RepositoryAdapter) CreateGalleryItem(ctx context.Context, postID postDomain.PostID, mediaID string, sortOrder int, caption *string, alt *string) (string, error) {
	id, err := a.inner.CreateGalleryItem(ctx, postID, mediaID, sortOrder, caption, alt)
	return id, mapErr(err)
}

func (a *RepositoryAdapter) UpdateGalleryItem(ctx context.Context, postID postDomain.PostID, itemID string, sortOrder *int, caption *string, alt *string) error {
	return mapErr(a.inner.UpdateGalleryItem(ctx, postID, itemID, sortOrder, caption, alt))
}

func (a *RepositoryAdapter) DeleteGalleryItem(ctx context.Context, postID postDomain.PostID, itemID string) error {
	return mapErr(a.inner.DeleteGalleryItem(ctx, postID, itemID))
}

// --- Changelog ---

func (a *RepositoryAdapter) CreateChangelog(ctx context.Context, postID postDomain.PostID, userID *string, note string) (string, error) {
	id, err := a.inner.CreateChangelog(ctx, postID, userID, note)
	return id, mapErr(err)
}

func (a *RepositoryAdapter) DeleteChangelog(ctx context.Context, postID postDomain.PostID, changelogID string) error {
	return mapErr(a.inner.DeleteChangelog(ctx, postID, changelogID))
}

// --- Syndication ---

func (a *RepositoryAdapter) CreateSyndication(ctx context.Context, postID postDomain.PostID, platform, url, status string) (string, error) {
	id, err := a.inner.CreateSyndication(ctx, postID, platform, url, status)
	return id, mapErr(err)
}

func (a *RepositoryAdapter) UpdateSyndication(ctx context.Context, postID postDomain.PostID, id string, platform, url, status *string) error {
	return mapErr(a.inner.UpdateSyndication(ctx, postID, id, platform, url, status))
}

func (a *RepositoryAdapter) DeleteSyndication(ctx context.Context, postID postDomain.PostID, id string) error {
	return mapErr(a.inner.DeleteSyndication(ctx, postID, id))
}

func (a *RepositoryAdapter) UpdateSyndicationByID(ctx context.Context, id string, platform, url, status *string) error {
	return mapErr(a.inner.UpdateSyndicationByID(ctx, id, platform, url, status))
}

func (a *RepositoryAdapter) DeleteSyndicationByID(ctx context.Context, id string) error {
	return mapErr(a.inner.DeleteSyndicationByID(ctx, id))
}

// --- Translations ---

func (a *RepositoryAdapter) PutPostTranslation(ctx context.Context, postID postDomain.PostID, groupID *string, locale string) (string, error) {
	gid, err := a.inner.PutPostTranslation(ctx, postID, groupID, locale)
	return gid, mapErr(err)
}

func (a *RepositoryAdapter) ClearPostTranslation(ctx context.Context, postID postDomain.PostID) error {
	return mapErr(a.inner.ClearPostTranslation(ctx, postID))
}

// --- Series (top-level) ---

func (a *RepositoryAdapter) ListSeries(ctx context.Context) ([]postDomain.Series, error) {
	out, err := a.inner.ListSeries(ctx)
	return out, mapErr(err)
}

func (a *RepositoryAdapter) CreateSeries(ctx context.Context, s *postDomain.Series) error {
	return mapErr(a.inner.CreateSeries(ctx, s))
}

func (a *RepositoryAdapter) FindSeriesByID(ctx context.Context, id string) (*postDomain.Series, error) {
	s, err := a.inner.FindSeriesByID(ctx, id)
	return s, mapErr(err)
}

func (a *RepositoryAdapter) UpdateSeries(ctx context.Context, s *postDomain.Series) error {
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

var _ postDomain.Repository = (*RepositoryAdapter)(nil)

