package domain

import "context"

// PostHooks defines the internal "hook points" that content modules can
// trigger. Phase 5 adds real plugin implementations later; Phase 5 A0 only
// provides the infrastructure.
type PostHooks interface {
	BeforePostSave(ctx context.Context, postID string, slug string) error
	AfterPostSave(ctx context.Context, postID string, slug string) error
}

