package agentrunapi

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestShouldSpillPromptThreshold(t *testing.T) {
	t.Parallel()
	if ShouldSpillPrompt(strings.Repeat("a", PromptFileSpillMinRunes)) {
		t.Fatalf("exactly %d runes must not spill", PromptFileSpillMinRunes)
	}
	if !ShouldSpillPrompt(strings.Repeat("a", PromptFileSpillMinRunes+1)) {
		t.Fatalf("%d runes must spill", PromptFileSpillMinRunes+1)
	}
	if ShouldSpillPrompt("  " + strings.Repeat("a", PromptFileSpillMinRunes) + "  ") {
		t.Fatal("threshold is after TrimSpace")
	}
}

func TestMaybeSpillPromptUnderThresholdNoFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path, spilled, err := MaybeSpillPrompt("hello", PromptSpillOpts{SpillDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if spilled || path != "" {
		t.Fatalf("short prompt: spilled=%v path=%q", spilled, path)
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) != 0 {
		t.Fatalf("spill dir should be empty; got %v", ents)
	}
}

func TestMaybeSpillPromptOverThresholdWritesBody(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	body := strings.Repeat("字", PromptFileSpillMinRunes+1)
	path, spilled, err := MaybeSpillPrompt(body, PromptSpillOpts{
		SpillDir:  dir,
		SessionID: "sess/spill",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !spilled {
		t.Fatal("expected spill")
	}
	if !filepath.IsAbs(path) {
		t.Fatalf("path must be abs; got %q", path)
	}
	rel, err := filepath.Rel(dir, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		t.Fatalf("path %q not under %q", path, dir)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Fatalf("spill body mismatch: got %d bytes want %d runes", len(got), utf8.RuneCountInString(body))
	}
	base := filepath.Base(path)
	if !strings.HasPrefix(base, "prompt-sess_spill-") {
		t.Fatalf("session id should be sanitized into name; got %q", base)
	}
}

func TestMaybeSpillPromptForceUnderThreshold(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path, spilled, err := MaybeSpillPrompt("short", PromptSpillOpts{SpillDir: dir, Force: true})
	if err != nil {
		t.Fatal(err)
	}
	if !spilled {
		t.Fatal("Force must write")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "short" {
		t.Fatalf("got %q", got)
	}
}
