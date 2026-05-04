package transport

import (
	"context"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/metrics"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/seo"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/series"
)

// stubPostsSubresources is a test double for PostsSubresources (all no-ops / empty).
type stubPostsSubresources struct{}

func (stubPostsSubresources) GetMetricsForPost(_ context.Context, _ string) (*metrics.PostMetrics, error) {
	return nil, nil
}
func (stubPostsSubresources) DeleteSEO(_ context.Context, _ string) error { return nil }
func (stubPostsSubresources) UpsertSEO(_ context.Context, _ string, _ *seo.PostSEO) error {
	return nil
}
func (stubPostsSubresources) SetFeaturedImage(_ context.Context, _ string, _, _ *string, _, _ *int, _, _ *float32, _, _ *string) error {
	return nil
}
func (stubPostsSubresources) SetPostSeries(_ context.Context, _ string, _ *string, _ *int, _ *string) error {
	return nil
}
func (stubPostsSubresources) ReplaceCoauthors(_ context.Context, _ string, _ []string) error {
	return nil
}
func (stubPostsSubresources) CreateGalleryItem(_ context.Context, _, _ string, _ int, _, _ *string) (string, error) {
	return "", nil
}
func (stubPostsSubresources) UpdateGalleryItem(_ context.Context, _, _ string, _ *int, _, _ *string) error {
	return nil
}
func (stubPostsSubresources) DeleteGalleryItem(_ context.Context, _, _ string) error { return nil }
func (stubPostsSubresources) CreateChangelog(_ context.Context, _ string, _ *string, _ string) (string, error) {
	return "", nil
}
func (stubPostsSubresources) DeleteChangelog(_ context.Context, _, _ string) error { return nil }
func (stubPostsSubresources) CreateSyndication(_ context.Context, _, _, _, _ string) (string, error) {
	return "", nil
}
func (stubPostsSubresources) UpdateSyndication(_ context.Context, _, _ string, _, _, _ *string) error {
	return nil
}
func (stubPostsSubresources) DeleteSyndication(_ context.Context, _, _ string) error { return nil }
func (stubPostsSubresources) UpdateSyndicationGlobal(_ context.Context, _ string, _, _, _ *string) error {
	return nil
}
func (stubPostsSubresources) DeleteSyndicationGlobal(_ context.Context, _ string) error { return nil }
func (stubPostsSubresources) PutPostTranslation(_ context.Context, _ string, _ *string, _ string) (string, error) {
	return "", nil
}
func (stubPostsSubresources) ClearPostTranslation(_ context.Context, _ string) error { return nil }

type stubSeriesAdmin struct{}

func (stubSeriesAdmin) ListSeries(_ context.Context) ([]series.Series, error) { return nil, nil }
func (stubSeriesAdmin) CreateSeries(_ context.Context, _, _ string) (*series.Series, error) {
	return nil, nil
}
func (stubSeriesAdmin) GetSeries(_ context.Context, _ string) (*series.Series, error) {
	return nil, nil
}
func (stubSeriesAdmin) UpdateSeries(_ context.Context, _ string, _, _ *string) (*series.Series, error) {
	return nil, nil
}
func (stubSeriesAdmin) DeleteSeries(_ context.Context, _ string) error { return nil }

type stubTranslationGroupsAdmin struct{}

func (stubTranslationGroupsAdmin) CreateTranslationGroup(_ context.Context, _ *string) (string, error) {
	return "", nil
}
func (stubTranslationGroupsAdmin) TranslationGroupExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (stubTranslationGroupsAdmin) DeleteTranslationGroup(_ context.Context, _ string) error {
	return nil
}
