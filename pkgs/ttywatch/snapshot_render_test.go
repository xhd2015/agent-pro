package ttywatch

import (
	"strings"
	"testing"
)

// grokChangelogGoodScrollback is the final grok changelog screen drawn via absolute CUP.
func grokChangelogGoodScrollback() string {
	var b strings.Builder
	b.WriteString("\x1b[?1049h\x1b[?25l")
	b.WriteString("\x1b[1;1H\x1b[K  Quit ctrl+q    Changelog    Settings")
	b.WriteString("\x1b[3;1H╭──────────────────────────────────────────────────────────╮")
	b.WriteString("\x1b[4;1H│ Grok Build Changelog                                     │")
	b.WriteString("\x1b[5;1H│                                                          │")
	b.WriteString("\x1b[6;1H│ • Snapshot screen-frame parity fix                       │")
	b.WriteString("\x1b[7;1H╰──────────────────────────────────────────────────────────╯")
	b.WriteString("\x1b[20;1H❯ Ask anything")
	b.WriteString("\x1b[24;1HLogged in with API key · Grok Build")
	return b.String()
}

// grokChangelogGarbledScrollback mimics scrollback polluted by changelog-only boot redraws
// that double vt10x replay collapses to Quit q without ctrl+q and no bordered box.
func grokChangelogGarbledScrollback() string {
	var b strings.Builder
	b.WriteString("\x1b[?1049h\x1b[?25l")
	b.WriteString("\x1b[1;1HChangelog")
	b.WriteString("\x1b[2;1HQuit q")
	b.WriteString("\x1b[3;1HGrok Build")
	b.WriteString("\x1b[4;1H• Snapshot screen-frame parity fix")
	b.WriteString("\x1b[20;1HAsk anything")
	return b.String()
}

func grokChangelogWants() []string {
	return []string{"ctrl+q", "Grok Build Changelog", "❯", "Logged in with API key"}
}

func TestRenderSnapshotOutput_prefersFrameOverGarbledScrollback(t *testing.T) {
	good := grokChangelogGoodScrollback()
	frame, ok := renderScreenSnapshotFrame([]byte(good), 80, 24)
	if !ok {
		t.Fatal("failed to build grok changelog screen snapshot frame")
	}
	badScroll := grokChangelogGarbledScrollback()

	scrollOnly := renderSnapshotOutput("", badScroll, 80, 24)
	if strings.Contains(scrollOnly, "ctrl+q") {
		t.Fatalf("garbled scrollback fixture should lack ctrl+q, got %q", scrollOnly)
	}

	out := renderSnapshotOutput(string(frame), badScroll, 80, 24)
	for _, want := range grokChangelogWants() {
		if !strings.Contains(out, want) {
			t.Fatalf("RenderSnapshotOutput must prefer screen frame over scrollback; missing %q, got %q", want, out)
		}
	}
	if strings.Contains(out, "Quit q") && !strings.Contains(out, "ctrl+q") {
		t.Fatalf("RenderSnapshotOutput used garbled scrollback (Quit q without ctrl+q), got %q", out)
	}
}

// grokPostTurnGoodScrollback mimics a completed grok turn: prompt, assistant reply,
// stop indicator, turn-complete status, and input box — content between UI chrome
// must not be dropped by snapshot rendering.
func grokPostTurnGoodScrollback() string {
	var b strings.Builder
	b.WriteString("\x1b[?1049h\x1b[?25l")
	b.WriteString("\x1b[2;1H  worktree header")
	b.WriteString("\x1b[5;1H     ❯ one word of France captial                                   11:39 AM")
	b.WriteString("\x1b[8;1H     making changes. Looking at the task: one word of France captial.")
	b.WriteString("\x1b[9;1H     Your request has been processed. Everything looks good.")
	b.WriteString("\x1b[14;1H     ◆ stop  [hooks: 1]")
	b.WriteString("\x1b[16;1H     Turn completed in 1.4s.")
	b.WriteString("\x1b[19;1H  ╭──────────────────────────────────────────────────────────╮")
	b.WriteString("\x1b[20;1H  │ ❯                                                        │")
	b.WriteString("\x1b[21;1H  ╰──────────────────────────────────────────────────────────╯")
	b.WriteString("\x1b[23;1H  Shift+Tab:mode  │  Ctrl+.:shortcuts")
	return b.String()
}

func grokPostTurnWants() []string {
	return []string{
		"one word of France captial",
		"processed. Everything looks good",
		"stop  [hooks: 1]",
		"Turn completed in 1.4s.",
		"╭",
		"Ctrl+.:shortcuts",
	}
}

func TestRenderSnapshotOutput_grokPostTurnFrameKeepsConversation(t *testing.T) {
	good := grokPostTurnGoodScrollback()
	frame, ok := renderScreenSnapshotFrame([]byte(good), 80, 24)
	if !ok {
		t.Fatal("failed to build grok post-turn screen snapshot frame")
	}

	out := renderSnapshotOutput(string(frame), "", 80, 24)
	for _, want := range grokPostTurnWants() {
		if !strings.Contains(out, want) {
			t.Fatalf("snapshot must keep grok post-turn conversation (watch parity); missing %q, got %q", want, out)
		}
	}
}

func TestScrollbackToScreenText_singlePassMatchesGoodFixture(t *testing.T) {
	good := grokChangelogGoodScrollback()
	direct, ok := screenSnapshotToText([]byte(good), 80, 24)
	if !ok {
		t.Fatal("single-pass screenSnapshotToText failed on good grok fixture")
	}
	for _, want := range grokChangelogWants() {
		if !strings.Contains(string(direct), want) {
			t.Fatalf("single-pass replay missing %q, got %q", want, direct)
		}
	}

	// Double vt10x scrollback replay must not drop bordered box from good fixture.
	out := renderSnapshotOutput("", good, 80, 24)
	for _, want := range grokChangelogWants() {
		if !strings.Contains(out, want) {
			t.Fatalf("scrollback render missing %q after replay, got %q", want, out)
		}
	}
}