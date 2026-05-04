package application

import (
	"context"
	"fmt"
	"log"
	"strings"

	pluginsDomain "github.com/nextpresskit/backend/internal/modules/plugins/domain"
	postPorts "github.com/nextpresskit/backend/internal/modules/posts/domain/ports"
)

// HookRegistry is the in-process registry for hook invocations.
// Handlers are invoked in registration order (typically one slot per enabled
// plugin row from the database). Phase 5 can replace noopPluginSlot with real
// implementations without changing the posts service contract.
type HookRegistry struct {
	chain []registeredPostHook
}

func NewHookRegistry() *HookRegistry {
	return &HookRegistry{chain: nil}
}

type hookErrorPolicy string

const (
	hookErrorPolicyFail     hookErrorPolicy = "fail"
	hookErrorPolicyContinue hookErrorPolicy = "continue"
)

type registeredPostHook struct {
	slug        string
	beforeError hookErrorPolicy
	afterError  hookErrorPolicy
	hook        pluginsDomain.PostHooks
}

// RegisterPostHooks appends a hook implementation to the chain. Nil is ignored.
func (r *HookRegistry) RegisterPostHooks(h pluginsDomain.PostHooks) {
	r.registerPostHooksWithPolicy("anonymous", h, hookErrorPolicyFail, hookErrorPolicyFail)
}

func (r *HookRegistry) registerPostHooksWithPolicy(
	slug string,
	h pluginsDomain.PostHooks,
	beforePolicy hookErrorPolicy,
	afterPolicy hookErrorPolicy,
) {
	if h == nil || r == nil {
		return
	}
	r.chain = append(r.chain, registeredPostHook{
		slug:        slug,
		beforeError: beforePolicy,
		afterError:  afterPolicy,
		hook:        h,
	})
}

// Policy:
// - before-save hook errors fail the operation by default (`fail`)
// - after-save hook errors continue by default (`continue`) because the post is already persisted
// Per-plugin overrides are supported via config keys:
// - `before_error_policy`: `fail` | `continue`
// - `after_error_policy`: `fail` | `continue`
func (r *HookRegistry) BeforePostSave(ctx context.Context, postID string, slug string) error {
	for _, item := range r.chain {
		if err := item.hook.BeforePostSave(ctx, postID, slug); err != nil {
			if item.beforeError == hookErrorPolicyContinue {
				log.Printf("plugin hook before-save error ignored (slug=%s): %v", item.slug, err)
				continue
			}
			return err
		}
	}
	return nil
}

func (r *HookRegistry) AfterPostSave(ctx context.Context, postID string, slug string) error {
	for _, item := range r.chain {
		if err := item.hook.AfterPostSave(ctx, postID, slug); err != nil {
			if item.afterError == hookErrorPolicyContinue {
				log.Printf("plugin hook after-save error ignored (slug=%s): %v", item.slug, err)
				continue
			}
			return err
		}
	}
	return nil
}

var _ pluginsDomain.PostHooks = (*HookRegistry)(nil)

var _ postPorts.PostSave = (*HookRegistry)(nil)

type pluginDispatchHook struct {
	pluginSlug       string
	blockedBeforeSet map[string]struct{}
	blockedAfterSet  map[string]struct{}
}

func (h pluginDispatchHook) BeforePostSave(_ context.Context, _ string, slug string) error {
	if _, blocked := h.blockedBeforeSet[strings.ToLower(strings.TrimSpace(slug))]; !blocked {
		return nil
	}
	return fmt.Errorf("plugin %q blocked slug before save: %s", h.pluginSlug, slug)
}

func (h pluginDispatchHook) AfterPostSave(_ context.Context, _ string, slug string) error {
	if _, blocked := h.blockedAfterSet[strings.ToLower(strings.TrimSpace(slug))]; !blocked {
		return nil
	}
	return fmt.Errorf("plugin %q blocked slug after save: %s", h.pluginSlug, slug)
}

type noopPluginSlot struct{}

func (noopPluginSlot) BeforePostSave(_ context.Context, _ string, _ string) error { return nil }

func (noopPluginSlot) AfterPostSave(_ context.Context, _ string, _ string) error { return nil }

func buildHookForPlugin(p pluginsDomain.Plugin) pluginsDomain.PostHooks {
	switch strings.ToLower(strings.TrimSpace(p.Slug)) {
	case "post-slug-guard":
		return pluginDispatchHook{
			pluginSlug:       p.Slug,
			blockedBeforeSet: stringSetFromConfig(p.Config, "before_blocked_slugs"),
			blockedAfterSet:  stringSetFromConfig(p.Config, "after_blocked_slugs"),
		}
	default:
		return noopPluginSlot{}
	}
}

func hookErrorPolicyFromConfig(config map[string]any, key string, fallback hookErrorPolicy) hookErrorPolicy {
	if config == nil {
		return fallback
	}
	raw, ok := config[key]
	if !ok {
		return fallback
	}
	text, ok := raw.(string)
	if !ok {
		return fallback
	}
	switch strings.ToLower(strings.TrimSpace(text)) {
	case string(hookErrorPolicyFail):
		return hookErrorPolicyFail
	case string(hookErrorPolicyContinue):
		return hookErrorPolicyContinue
	default:
		return fallback
	}
}

func stringSetFromConfig(config map[string]any, key string) map[string]struct{} {
	if config == nil {
		return map[string]struct{}{}
	}

	raw, ok := config[key]
	if !ok {
		return map[string]struct{}{}
	}

	out := make(map[string]struct{})
	switch vals := raw.(type) {
	case []any:
		for _, v := range vals {
			s, ok := v.(string)
			if !ok {
				continue
			}
			normalized := strings.ToLower(strings.TrimSpace(s))
			if normalized != "" {
				out[normalized] = struct{}{}
			}
		}
	case []string:
		for _, s := range vals {
			normalized := strings.ToLower(strings.TrimSpace(s))
			if normalized != "" {
				out[normalized] = struct{}{}
			}
		}
	}
	return out
}

// BootstrapPostHooks reads enabled plugins and registers one chain slot per row.
func BootstrapPostHooks(ctx context.Context, repo pluginsDomain.Repository) (*HookRegistry, int, error) {
	enabled, err := repo.ListEnabled(ctx)
	if err != nil {
		return nil, 0, err
	}

	reg := NewHookRegistry()
	for _, plugin := range enabled {
		reg.registerPostHooksWithPolicy(
			plugin.Slug,
			buildHookForPlugin(plugin),
			hookErrorPolicyFromConfig(plugin.Config, "before_error_policy", hookErrorPolicyFail),
			hookErrorPolicyFromConfig(plugin.Config, "after_error_policy", hookErrorPolicyContinue),
		)
	}

	return reg, len(enabled), nil
}
