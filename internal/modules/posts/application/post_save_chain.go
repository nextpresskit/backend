package application

import (
	"context"

	postDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain"
)

// PostSaveChain composes multiple PostSave implementations.
// Nil entries are ignored.
type PostSaveChain struct {
	chain []postDomain.PostSave
}

func NewPostSaveChain(items ...postDomain.PostSave) *PostSaveChain {
	out := &PostSaveChain{chain: make([]postDomain.PostSave, 0, len(items))}
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

var _ postDomain.PostSave = (*PostSaveChain)(nil)

