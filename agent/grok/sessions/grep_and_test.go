package sessions

import (
	"strings"
	"testing"
)

func TestTextContainsAllLiteralCI(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		s        string
		patterns []string
		want     bool
	}{
		{name: "empty patterns", s: "anything", patterns: nil, want: true},
		{name: "single hit", s: "hello world", patterns: []string{"WORLD"}, want: true},
		{name: "single miss", s: "hello world", patterns: []string{"xyz"}, want: false},
		{name: "and both present", s: "fix the timeout path", patterns: []string{"fix", "timeout"}, want: true},
		{name: "and missing one", s: "fix the retry path", patterns: []string{"fix", "timeout"}, want: false},
		{name: "order independent", s: "a b c", patterns: []string{"c", "a"}, want: true},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := textContainsAllLiteralCI(tc.s, tc.patterns)
			if got != tc.want {
				t.Fatalf("textContainsAllLiteralCI(%q, %v) = %v, want %v", tc.s, tc.patterns, got, tc.want)
			}
		})
	}
}

func TestNormalizeGrepPatterns(t *testing.T) {
	t.Parallel()
	out, err := normalizeGrepPatterns([]string{"  foo ", "bar"})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if len(out) != 2 || out[0] != "foo" || out[1] != "bar" {
		t.Fatalf("normalize = %#v", out)
	}
	if _, err := normalizeGrepPatterns(nil); err == nil {
		t.Fatal("expected error for empty patterns")
	}
	if _, err := normalizeGrepPatterns([]string{"ok", "  "}); err == nil {
		t.Fatal("expected error for blank pattern")
	}
	if _, err := validateGrepPatterns(false, []string{""}); err != nil {
		t.Fatalf("!set should ignore patterns: %v", err)
	}
	if _, err := validateGrepPatterns(true, nil); err == nil {
		t.Fatal("GrepSet with empty slice should error")
	}
}

func TestHighlightAllLiteralCI(t *testing.T) {
	t.Parallel()
	got := highlightAllLiteralCI("fix the Timeout now", []string{"fix", "timeout"})
	if !strings.Contains(got, "\x1b[1m\x1b[31mfix\x1b[0m") {
		t.Fatalf("missing fix highlight in %q", got)
	}
	if !strings.Contains(got, "\x1b[1m\x1b[31mTimeout\x1b[0m") {
		t.Fatalf("missing Timeout highlight in %q", got)
	}
}
