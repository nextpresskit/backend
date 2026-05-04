package transport

import (
	"context"

	postApp "github.com/nextpresskit/backend/internal/modules/posts/application"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/metrics"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/model"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/seo"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/series"
)

// PostsElasticsearchBackend is optional Elasticsearch search + post document sync.
type PostsElasticsearchBackend interface {
	SearchPostIDs(ctx context.Context, q string, limit, offset int) ([]string, error)
	SyncPost(ctx context.Context, p *model.Post)
}

// PostsCore is the application surface for core post CRUD, public reads, and taxonomy.
type PostsCore interface {
	Create(ctx context.Context, authorID, title, slug, content string) (*model.Post, error)
	GetByID(ctx context.Context, id string) (*model.Post, error)
	ListFiltered(ctx context.Context, limit, offset int, status string, authorID string, q string) ([]model.Post, error)
	PublicList(ctx context.Context, limit, offset int, q string, categoryID string, tagID string) ([]model.Post, error)
	PublicGetBySlug(ctx context.Context, slug string) (*model.Post, error)
	ReindexPublishedForSearch(ctx context.Context, sync func(context.Context, *model.Post)) (int, error)
	Update(ctx context.Context, id, title, slug, content, status string) (*model.Post, error)
	Save(ctx context.Context, p *model.Post) (*model.Post, error)
	Delete(ctx context.Context, id string) error
	SetCategories(ctx context.Context, postID string, categoryIDs []string) error
	SetTags(ctx context.Context, postID string, tagIDs []string) error
	SetPrimaryCategory(ctx context.Context, postID string, categoryID *string) error
}

// PostsSubresources is the application surface for nested post sub-resources.
type PostsSubresources interface {
	GetMetricsForPost(ctx context.Context, postID string) (*metrics.PostMetrics, error)
	DeleteSEO(ctx context.Context, postID string) error
	UpsertSEO(ctx context.Context, postID string, data *seo.PostSEO) error
	SetFeaturedImage(ctx context.Context, postID string, mediaID *string, alt *string, width *int, height *int, focalX *float32, focalY *float32, credit *string, license *string) error
	SetPostSeries(ctx context.Context, postID string, seriesID *string, partIndex *int, partLabel *string) error
	ReplaceCoauthors(ctx context.Context, postID string, userIDs []string) error
	CreateGalleryItem(ctx context.Context, postID, mediaID string, sortOrder int, caption *string, alt *string) (string, error)
	UpdateGalleryItem(ctx context.Context, postID, itemID string, sortOrder *int, caption *string, alt *string) error
	DeleteGalleryItem(ctx context.Context, postID, itemID string) error
	CreateChangelog(ctx context.Context, postID string, userID *string, note string) (string, error)
	DeleteChangelog(ctx context.Context, postID, changelogID string) error
	CreateSyndication(ctx context.Context, postID, platform, url, status string) (string, error)
	UpdateSyndication(ctx context.Context, postID, syndicationID string, platform, url, status *string) error
	DeleteSyndication(ctx context.Context, postID, syndicationID string) error
	UpdateSyndicationGlobal(ctx context.Context, syndicationID string, platform, url, status *string) error
	DeleteSyndicationGlobal(ctx context.Context, syndicationID string) error
	PutPostTranslation(ctx context.Context, postID string, groupID *string, locale string) (string, error)
	ClearPostTranslation(ctx context.Context, postID string) error
}

// SeriesAdmin is the application surface for top-level series CRUD.
type SeriesAdmin interface {
	ListSeries(ctx context.Context) ([]series.Series, error)
	CreateSeries(ctx context.Context, title, slug string) (*series.Series, error)
	GetSeries(ctx context.Context, id string) (*series.Series, error)
	UpdateSeries(ctx context.Context, id string, title, slug *string) (*series.Series, error)
	DeleteSeries(ctx context.Context, id string) error
}

// TranslationGroupsAdmin is the application surface for translation groups.
type TranslationGroupsAdmin interface {
	CreateTranslationGroup(ctx context.Context, explicitID *string) (string, error)
	TranslationGroupExists(ctx context.Context, id string) (bool, error)
	DeleteTranslationGroup(ctx context.Context, id string) error
}

// NewHandler wires the posts HTTP layer to focused application services.
func NewHandler(core PostsCore, sub PostsSubresources, series SeriesAdmin, groups TranslationGroupsAdmin) *Handler {
	return NewHandlerWithOptionalSearch(core, sub, series, groups, nil)
}

// NewHandlerWithOptionalSearch adds an Elasticsearch-backed search implementation when non-nil.
func NewHandlerWithOptionalSearch(core PostsCore, sub PostsSubresources, series SeriesAdmin, groups TranslationGroupsAdmin, es PostsElasticsearchBackend) *Handler {
	return &Handler{core: core, sub: sub, series: series, groups: groups, esBackend: es}
}

// NewHandlerFromService adapts the façade service to transport dependencies.
func NewHandlerFromService(svc *postApp.Service) *Handler {
	return NewHandlerFromServiceWithOptionalSearch(svc, nil)
}

// NewHandlerFromServiceWithOptionalSearch passes a non-nil es value when Elasticsearch is enabled.
func NewHandlerFromServiceWithOptionalSearch(svc *postApp.Service, es PostsElasticsearchBackend) *Handler {
	return NewHandlerWithOptionalSearch(svc.CorePostsService, svc.PostSubresourcesService, svc.SeriesService, svc.TranslationGroupsService, es)
}
