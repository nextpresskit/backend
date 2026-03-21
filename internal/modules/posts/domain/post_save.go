package domain

import "context"

// PostSave is a port invoked by the posts application service around persistence.
// Implementations are provided at the composition root (e.g. plugin HookRegistry).
type PostSave interface {
	BeforePostSave(ctx context.Context, postID, slug string) error
	AfterPostSave(ctx context.Context, postID, slug string) error
}
