package application

import (
	postDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain"
)

// PostSubresourceStores groups narrow persistence ports for post sub-resources.
type PostSubresourceStores struct {
	Reader       postDomain.PostReader
	SEO          postDomain.PostSEOStore
	Metrics      postDomain.PostMetricsStore
	Featured     postDomain.PostFeaturedImageStore
	SeriesLink   postDomain.PostSeriesLinkStore
	Coauthors    postDomain.PostCoauthorsStore
	Gallery      postDomain.PostGalleryStore
	Changelog    postDomain.PostChangelogStore
	Syndication  postDomain.PostSyndicationStore
	Translations postDomain.PostTranslationsStore
}

// PostSubresourceStoresFrom adapts a full Repository into store fields (same concrete value).
func PostSubresourceStoresFrom(r postDomain.Repository) PostSubresourceStores {
	return PostSubresourceStores{
		Reader:       r,
		SEO:          r,
		Metrics:      r,
		Featured:     r,
		SeriesLink:   r,
		Coauthors:    r,
		Gallery:      r,
		Changelog:    r,
		Syndication:  r,
		Translations: r,
	}
}
