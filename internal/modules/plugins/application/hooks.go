package application

import (
	"context"

	pluginsDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/plugins/domain"
	postDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/posts/domain"
)

// HookRegistry is the in-process registry for hook invocations.
// Handlers are invoked in registration order (typically one slot per enabled
// plugin row from the database). Phase 5 can replace noopPluginSlot with real
// implementations without changing the posts service contract.
type HookRegistry struct {
	chain []pluginsDomain.PostHooks
}

func NewHookRegistry() *HookRegistry {
	return &HookRegistry{chain: nil}
}

// RegisterPostHooks appends a hook implementation to the chain. Nil is ignored.
func (r *HookRegistry) RegisterPostHooks(h pluginsDomain.PostHooks) {
	if h == nil || r == nil {
		return
	}
	r.chain = append(r.chain, h)
}

func (r *HookRegistry) BeforePostSave(ctx context.Context, postID string, slug string) error {
	for _, h := range r.chain {
		if err := h.BeforePostSave(ctx, postID, slug); err != nil {
			return err
		}
	}
	return nil
}

func (r *HookRegistry) AfterPostSave(ctx context.Context, postID string, slug string) error {
	for _, h := range r.chain {
		if err := h.AfterPostSave(ctx, postID, slug); err != nil {
			return err
		}
	}
	return nil
}

var _ pluginsDomain.PostHooks = (*HookRegistry)(nil)

var _ postDomain.PostSave = (*HookRegistry)(nil)

// noopPluginSlot reserves one hook chain entry per enabled plugin. Real
// plugin logic will replace or wrap this in later Phase 5 work.
type noopPluginSlot struct{}

func (noopPluginSlot) BeforePostSave(_ context.Context, _ string, _ string) error { return nil }

func (noopPluginSlot) AfterPostSave(_ context.Context, _ string, _ string) error { return nil }

// BootstrapPostHooks reads enabled plugins and registers one chain slot per row.
func BootstrapPostHooks(ctx context.Context, repo pluginsDomain.Repository) (*HookRegistry, int, error) {
	enabled, err := repo.ListEnabled(ctx)
	if err != nil {
		return nil, 0, err
	}

	reg := NewHookRegistry()
	for range enabled {
		reg.RegisterPostHooks(noopPluginSlot{})
	}

	return reg, len(enabled), nil
}

