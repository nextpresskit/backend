package application

import (
	"context"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/nextpresskit/backend/internal/modules/posts/domain/ident"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/metrics"
	"github.com/nextpresskit/backend/internal/modules/posts/domain/ports"
)

// DerivedFieldsHook computes and persists derived editorial fields (excerpt + metrics)
// using the PostSave seam. It operates after the core post write succeeds.
type DerivedFieldsHook struct {
	repo ports.PostLoadUpdater
}

func NewDerivedFieldsHook(repo ports.PostLoadUpdater) *DerivedFieldsHook {
	return &DerivedFieldsHook{repo: repo}
}

func (h *DerivedFieldsHook) BeforePostSave(_ context.Context, _ string, _ string) error { return nil }

func (h *DerivedFieldsHook) AfterPostSave(ctx context.Context, postID string, _ string) error {
	if h == nil || h.repo == nil {
		return nil
	}
	p, err := h.repo.FindByID(ctx, ident.PostID(strings.TrimSpace(postID)))
	if err != nil || p == nil {
		return err
	}

	changed := false

	if strings.TrimSpace(p.Excerpt) == "" && strings.TrimSpace(p.Content) != "" {
		p.Excerpt = deriveExcerpt(p.Content, 240)
		changed = true
	}

	wc, cc := countWordsAndChars(p.Content)
	rtMin := readingTimeMinutes(wc, 200)
	if p.Metrics == nil {
		p.Metrics = &metrics.PostMetrics{}
	}
	if p.Metrics.WordCount != wc || p.Metrics.CharacterCount != cc || p.Metrics.ReadingTimeMinutes != rtMin || p.Metrics.EstReadTimeSeconds != rtMin*60 {
		p.Metrics.WordCount = wc
		p.Metrics.CharacterCount = cc
		p.Metrics.ReadingTimeMinutes = rtMin
		p.Metrics.EstReadTimeSeconds = rtMin * 60
		p.Metrics.UpdatedAt = time.Now().UTC()
		changed = true
	}

	if !changed {
		return nil
	}

	// Persist without re-invoking hooks (we call repo.Update directly).
	p.UpdatedAt = time.Now().UTC()
	return h.repo.Update(ctx, p)
}

var _ ports.PostSave = (*DerivedFieldsHook)(nil)

var markdownNoise = regexp.MustCompile(`(?m)^[#>\-\*\+]\s+`)
var whitespace = regexp.MustCompile(`\s+`)

func deriveExcerpt(markdown string, maxChars int) string {
	s := strings.TrimSpace(markdown)
	s = markdownNoise.ReplaceAllString(s, "")
	s = whitespace.ReplaceAllString(s, " ")
	if maxChars <= 0 {
		maxChars = 240
	}
	if len(s) <= maxChars {
		return s
	}
	out := strings.TrimSpace(s[:maxChars])
	out = strings.TrimRight(out, ".,;:- ")
	return out + "…"
}

func countWordsAndChars(markdown string) (wordCount int, charCount int) {
	s := strings.TrimSpace(markdown)
	if s == "" {
		return 0, 0
	}
	charCount = len(s)
	normalized := whitespace.ReplaceAllString(s, " ")
	parts := strings.Split(normalized, " ")
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			wordCount++
		}
	}
	return wordCount, charCount
}

func readingTimeMinutes(wordCount int, wpm int) int {
	if wordCount <= 0 {
		return 0
	}
	if wpm <= 0 {
		wpm = 200
	}
	min := int(math.Ceil(float64(wordCount) / float64(wpm)))
	if min < 1 {
		return 1
	}
	return min
}
