package relations

import (
	"context"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
)

// PostTaxonomyWriter assigns taxonomy and primary category on posts.
type PostTaxonomyWriter interface {
	SetCategories(ctx context.Context, postID ident.PostID, categoryIDs []string) error
	SetTags(ctx context.Context, postID ident.PostID, tagIDs []string) error
	SetPrimaryCategory(ctx context.Context, postID ident.PostID, categoryID *string) error
}

// PostFeaturedImageStore updates featured media columns on posts.
type PostFeaturedImageStore interface {
	SetFeaturedImage(ctx context.Context, postID ident.PostID, mediaID *string, alt *string, width *int, height *int, focalX *float32, focalY *float32, credit *string, license *string) error
}

// PostCoauthorsStore manages post_coauthors.
type PostCoauthorsStore interface {
	ReplaceCoauthors(ctx context.Context, postID ident.PostID, userIDs []string) error
}

// PostSeriesLinkStore manages post_series membership.
type PostSeriesLinkStore interface {
	SetPostSeries(ctx context.Context, postID ident.PostID, seriesID *string, partIndex *int, partLabel *string) error
}
