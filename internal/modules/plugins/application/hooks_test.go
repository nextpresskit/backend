package application

import (
	"context"
	"strings"
	"testing"

	pluginsDomain "github.com/nextpresskit/backend/internal/modules/plugins/domain"
)

type countingHook struct {
	before, after *int
}

func (c countingHook) BeforePostSave(_ context.Context, _, _ string) error {
	*c.before++
	return nil
}

func (c countingHook) AfterPostSave(_ context.Context, _, _ string) error {
	*c.after++
	return nil
}

func TestHookRegistry_ChainInvokesInOrder(t *testing.T) {
	var beforeCount, afterCount int
	r := NewHookRegistry()
	r.RegisterPostHooks(countingHook{before: &beforeCount, after: &afterCount})
	r.RegisterPostHooks(countingHook{before: &beforeCount, after: &afterCount})

	ctx := context.Background()
	if err := r.BeforePostSave(ctx, "id", "slug"); err != nil {
		t.Fatal(err)
	}
	if beforeCount != 2 {
		t.Fatalf("expected 2 before calls, got %d", beforeCount)
	}
	if err := r.AfterPostSave(ctx, "id", "slug"); err != nil {
		t.Fatal(err)
	}
	if afterCount != 2 {
		t.Fatalf("expected 2 after calls, got %d", afterCount)
	}
}

type bootstrapRepoStub struct {
	enabled []pluginsDomain.Plugin
}

func (s bootstrapRepoStub) Create(_ context.Context, _ *pluginsDomain.Plugin) error { return nil }

func (s bootstrapRepoStub) FindByID(_ context.Context, _ pluginsDomain.PluginID) (*pluginsDomain.Plugin, error) {
	return nil, nil
}

func (s bootstrapRepoStub) FindBySlug(_ context.Context, _ string) (*pluginsDomain.Plugin, error) {
	return nil, nil
}

func (s bootstrapRepoStub) List(_ context.Context) ([]pluginsDomain.Plugin, error) { return nil, nil }

func (s bootstrapRepoStub) ListEnabled(_ context.Context) ([]pluginsDomain.Plugin, error) {
	return s.enabled, nil
}

func (s bootstrapRepoStub) Update(_ context.Context, _ *pluginsDomain.Plugin) error { return nil }

func TestBootstrapPostHooks_DispatchesByPluginSlugAndConfig(t *testing.T) {
	ctx := context.Background()
	repo := bootstrapRepoStub{
		enabled: []pluginsDomain.Plugin{
			{
				Slug: "post-slug-guard",
				Config: map[string]any{
					"before_blocked_slugs": []any{"blocked-before"},
					"after_blocked_slugs":  []any{"blocked-after"},
					"after_error_policy":   "fail",
				},
			},
			{Slug: "some-unknown-plugin"},
		},
	}

	reg, count, err := BootstrapPostHooks(ctx, repo)
	if err != nil {
		t.Fatalf("unexpected bootstrap error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected enabled count 2, got %d", count)
	}

	if err := reg.BeforePostSave(ctx, "id", "allowed"); err != nil {
		t.Fatalf("expected allow for before save slug, got err: %v", err)
	}
	if err := reg.AfterPostSave(ctx, "id", "allowed"); err != nil {
		t.Fatalf("expected allow for after save slug, got err: %v", err)
	}

	beforeErr := reg.BeforePostSave(ctx, "id", "blocked-before")
	if beforeErr == nil || !strings.Contains(beforeErr.Error(), "post-slug-guard") {
		t.Fatalf("expected post-slug-guard before error, got: %v", beforeErr)
	}

	afterErr := reg.AfterPostSave(ctx, "id", "blocked-after")
	if afterErr == nil || !strings.Contains(afterErr.Error(), "post-slug-guard") {
		t.Fatalf("expected post-slug-guard after error, got: %v", afterErr)
	}
}

func TestBootstrapPostHooks_AfterErrorPolicyContinue(t *testing.T) {
	ctx := context.Background()
	repo := bootstrapRepoStub{
		enabled: []pluginsDomain.Plugin{
			{
				Slug: "post-slug-guard",
				Config: map[string]any{
					"after_blocked_slugs": []any{"blocked-after"},
					"after_error_policy":  "continue",
				},
			},
		},
	}

	reg, _, err := BootstrapPostHooks(ctx, repo)
	if err != nil {
		t.Fatalf("unexpected bootstrap error: %v", err)
	}

	if err := reg.AfterPostSave(ctx, "id", "blocked-after"); err != nil {
		t.Fatalf("expected after error to be ignored with continue policy, got: %v", err)
	}
}

func TestBootstrapPostHooks_BeforeErrorPolicyFail(t *testing.T) {
	ctx := context.Background()
	repo := bootstrapRepoStub{
		enabled: []pluginsDomain.Plugin{
			{
				Slug: "post-slug-guard",
				Config: map[string]any{
					"before_blocked_slugs": []any{"blocked-before"},
					"before_error_policy":  "fail",
				},
			},
		},
	}

	reg, _, err := BootstrapPostHooks(ctx, repo)
	if err != nil {
		t.Fatalf("unexpected bootstrap error: %v", err)
	}

	if err := reg.BeforePostSave(ctx, "id", "blocked-before"); err == nil {
		t.Fatal("expected before error with fail policy")
	}
}
