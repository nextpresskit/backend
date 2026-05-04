package metrics

import (
	"context"
	"time"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
)

// PostMetrics is read model / analytics snapshot for a post.
type PostMetrics struct {
	WordCount             int
	CharacterCount        int
	ReadingTimeMinutes    int
	EstReadTimeSeconds    int
	ViewCount             int64
	UniqueVisitors7d      int64
	ScrollDepthAvgPercent float32
	BounceRatePercent     float32
	AvgTimeOnPageSeconds  int
	CommentCount          int
	LikeCount             int
	ShareCount            int
	BookmarkCount         int
	UpdatedAt             time.Time
}

// PostMetricsStore reads/writes post_metrics (read path; writes often via full post upsert).
type PostMetricsStore interface {
	GetMetrics(ctx context.Context, postID ident.PostID) (*PostMetrics, error)
}
