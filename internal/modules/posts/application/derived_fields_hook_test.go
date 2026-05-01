package application

import (
	"strings"
	"testing"
)

func TestDeriveExcerpt_TrimsAndEllipsis(t *testing.T) {
	in := "## Title\n\nThis is a test post with some content that should be turned into an excerpt.\n\nMore text here."
	out := deriveExcerpt(in, 30)
	if out == "" {
		t.Fatalf("expected non-empty excerpt")
	}
	if len(out) < 10 {
		t.Fatalf("excerpt too short: %q", out)
	}
	if !strings.HasSuffix(out, "…") {
		t.Fatalf("expected ellipsis, got %q", out)
	}
}

func TestCountWordsAndChars(t *testing.T) {
	wc, cc := countWordsAndChars("hello   world\nnextpresskit")
	if wc != 3 {
		t.Fatalf("wordCount=%d, want 3", wc)
	}
	if cc == 0 {
		t.Fatalf("charCount should be >0")
	}
}

func TestReadingTimeMinutes(t *testing.T) {
	if got := readingTimeMinutes(0, 200); got != 0 {
		t.Fatalf("got=%d want=0", got)
	}
	if got := readingTimeMinutes(1, 200); got != 1 {
		t.Fatalf("got=%d want=1", got)
	}
	if got := readingTimeMinutes(201, 200); got != 2 {
		t.Fatalf("got=%d want=2", got)
	}
}
