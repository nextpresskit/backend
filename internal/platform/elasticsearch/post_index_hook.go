package elasticsearch

import (
	"context"
	"strings"

	"go.uber.org/zap"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/ports"
)

// PostIndexHook syncs post documents to Elasticsearch after successful writes.
// Errors are logged only; hooks must not fail the HTTP transaction (plan).
type PostIndexHook struct {
	log   *zap.SugaredLogger
	idx   *PostsIndex
	read  ports.PostReader
}

// NewPostIndexHook returns nil when idx is nil.
func NewPostIndexHook(log *zap.SugaredLogger, idx *PostsIndex, read ports.PostReader) *PostIndexHook {
	if idx == nil || read == nil {
		return nil
	}
	return &PostIndexHook{log: log, idx: idx, read: read}
}

func (h *PostIndexHook) BeforePostSave(_ context.Context, _ string, _ string) error {
	return nil
}

func (h *PostIndexHook) AfterPostSave(ctx context.Context, postID string, _ string) error {
	if h == nil || h.idx == nil {
		return nil
	}
	postID = strings.TrimSpace(postID)
	if postID == "" {
		return nil
	}
	p, err := h.read.FindByID(ctx, ident.PostID(postID))
	if err != nil {
		h.log.Warnw("elasticsearch hook load post failed", "post_id", postID, "error", err)
		return nil
	}
	if p == nil {
		h.idx.DeletePost(ctx, postID)
		return nil
	}
	h.idx.SyncPost(ctx, p)
	return nil
}

var _ ports.PostSave = (*PostIndexHook)(nil)
