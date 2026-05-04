package ports

import (
	"github.com/nextpresskit/backend/internal/modules/posts/domain/extensions"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/metrics"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/relations"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/seo"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/series"
)

// Repository is the composed persistence port for the posts module.
// Prefer depending on smaller interfaces (PostReader, PostSEOStore, …) in new code.
type Repository interface {
	PostReader
	PostWriter
	relations.PostTaxonomyWriter
	seo.PostSEOStore
	metrics.PostMetricsStore
	relations.PostFeaturedImageStore
	relations.PostSeriesLinkStore
	relations.PostCoauthorsStore
	extensions.PostGalleryStore
	extensions.PostChangelogStore
	extensions.PostSyndicationStore
	extensions.PostTranslationsStore
	series.SeriesRepository
	extensions.TranslationGroupRepository
}
