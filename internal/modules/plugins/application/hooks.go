package application

import (
	"context"

	pluginsDomain "github.com/Petar-V-Nikolov/nextpress-backend/internal/modules/plugins/domain"
)

// HookRegistry is the in-process registry for hook invocations.
// Phase 5 will register real plugin-backed hook implementations; A0 only
// provides a no-op registry with the calling infrastructure.
type HookRegistry struct {
	// In later phases we will store enabled plugin hook implementations here.
}

func NewHookRegistry() *HookRegistry {
	return &HookRegistry{}
}

func (r *HookRegistry) BeforePostSave(ctx context.Context, postID string, slug string) error {
	// No-op for A0.
	return nil
}

func (r *HookRegistry) AfterPostSave(ctx context.Context, postID string, slug string) error {
	// No-op for A0.
	return nil
}

var _ pluginsDomain.PostHooks = (*HookRegistry)(nil)

// BootstrapPostHooks reads enabled plugins and wires their hook implementations.
// For A0, the registry is a no-op, but we still perform the enabled plugins
// lookup so later phases can extend the mapping.
func BootstrapPostHooks(ctx context.Context, repo pluginsDomain.Repository) (*HookRegistry, int, error) {
	enabled, err := repo.ListEnabled(ctx)
	if err != nil {
		return nil, 0, err
	}

	_ = enabled
	return NewHookRegistry(), len(enabled), nil
}

