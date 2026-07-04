package main

import (
	"os"
	"strings"
	"testing"
)

func TestRenderSnapshotScrollbackCodexTrustPrompt(t *testing.T) {
	raw, err := os.ReadFile("/tmp/tty-watch-raw-scrollback.bin")
	if err != nil {
		// Fallback fixture mirrors live codex trust prompt (?2026h + CUP, no 2J).
		raw = []byte("\x1b[?1049l\x1b[0m\x1b[?2026h\x1b[1;1H\x1b[J\x1b[2;2H> You are in /tmp/work\x1b[5;58HDo you trust the contents of this directory?\x1b[10;26H\x203a 1. Yes, continue\x1b[11;2H  2. No, quit\x1b[12;2H  Press enter to continue")
	}
	got := renderSnapshotScrollback(string(raw), 80, 24)
	if got == "" {
		t.Fatal("empty snapshot text")
	}
	if strings.Contains(got, "doyoutrustthecontentsofthisdirectory") {
		t.Fatalf("smeared sanitize fallback, got %q", got)
	}
	if !strings.Contains(got, "Do you trust the contents of this directory?") {
		t.Fatalf("missing trust prompt, got %q", got)
	}
	for _, line := range strings.Split(got, "\n") {
		if len(line) > 120 {
			t.Fatalf("smeared long line (%d chars): %q", len(line), line)
		}
	}
}

func TestRenderSnapshotScrollbackCodexCursorDrawnUI(t *testing.T) {
	scrollback := []byte("\x1b[?2026h\x1b[3;1H│ >_ \x1b[1mOpenAI Codex\x1b[22m (v0.142.5) │\x1b[5;1H│ model: gpt-5.5 medium │\x1b[6;1H│ directory: /tmp/work │\x1b[10;1H›\x1b[10;3HWrite tests for @filename\x1b]0;title\a")
	got := renderSnapshotScrollback(string(scrollback), 80, 24)
	if strings.Count(got, "OpenAI Codex") != 1 {
		t.Fatalf("want single UI header, got %q", got)
	}
	if !strings.Contains(got, "Write tests for @filename") {
		t.Fatalf("missing prompt line, got %q", got)
	}
}