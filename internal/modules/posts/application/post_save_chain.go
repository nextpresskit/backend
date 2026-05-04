package application

import (
	"context"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ports"
)

// PostSaveChain composes multiple PostSave implementations.
// Nil entries are ignored.
type PostSaveChain struct {
	chain []ports.PostSave
}

func NewPostSaveChain(items ...ports.PostSave) *PostSaveChain {
	out := &PostSaveChain{chain: make([]ports.PostSave, 0, len(items))}
	for _, it := range items {
		if it == nil {
			continue
		}
		out.chain = append(out.chain, it)
	}
	return out
}

func (c *PostSaveChain) BeforePostSave(ctx context.Context, postID, slug string) error {
	for _, h := range c.chain {
		if err := h.BeforePostSave(ctx, postID, slug); err != nil {
			return err
		}
	}
	return nil
}

func (c *PostSaveChain) AfterPostSave(ctx context.Context, postID, slug string) error {
	for _, h := range c.chain {
		if err := h.AfterPostSave(ctx, postID, slug); err != nil {
			return err
		}
	}
	return nil
}

var _ ports.PostSave = (*PostSaveChain)(nil)
