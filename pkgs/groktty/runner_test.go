package groktty

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtractAssistantTextSubmittedLine(t *testing.T) {
	scrollback := []byte("GROK_TTY_BANNER\nGrok › SUBMITTED:run ls\n")
	got := extractAssistantText(scrollback, "run ls")
	if got != "run ls" {
		t.Fatalf("got %q", got)
	}
}

func TestCodexBannerDetectedFromCursorDrawnStartupTip(t *testing.T) {
	scrollback := []byte("\x1b[?2026h\x1b[3;1H│ >_ \x1b[1mOpenAI Codex\x1b[22m (v0.142.5) │\x1b[5;1H│ model: gpt-5.5 medium │\x1b[6;1H│ directory: /tmp/work │\x1b[10;1H›\x1b[10;3HWrite tests for @filename\x1b]0;title\a")
	if !bannerDetectedConfig(scrollback, "codex", nil) {
		t.Fatalf("Codex input screen was not detected")
	}
}

func TestCodexModelLoadingScreenIsNotReadiness(t *testing.T) {
	scrollback := []byte("\x1b[?2026h\x1b[3;1H│ >_ \x1b[1mOpenAI Codex\x1b[22m (v0.142.5) │\x1b[5;1H│ model: loading /model to change │\x1b[6;1H│ directory: /tmp/work │\x1b[10;1H›\x1b[10;3HWrite tests for @filename\x1b]0;title\a")
	if bannerDetectedConfig(scrollback, "codex", nil) {
		t.Fatalf("Codex model-loading screen should not be treated as input readiness")
	}
}

func TestCodexStartupWarningIsNotReadiness(t *testing.T) {
	scrollback := []byte("`\n[\nf\ne\na\nt\nu\nr\ne\ns\n]\n.\nc\no\nd\ne\nx\n_\nh\no\no\nk\ns\n`\ni\ns\nd\ne\np\nr\ne\nc\na\nt\ne\nd\n.\n")
	if bannerDetectedConfig(scrollback, "codex", nil) {
		t.Fatalf("Codex hooks warning should not be treated as input readiness")
	}
}

func TestCodexTrustPromptDetected(t *testing.T) {
	scrollback := []byte(">4;0m>7u>You are in /tmp/TestGeneratedCaseDo you trust the contents of this directory? Working with untrusted contents comes with higher risk. › 1. Yes, continue 2. No, quit Press enter to continue")
	if !codexTrustPromptDetected(scrollback, "codex") {
		t.Fatalf("Codex trust prompt was not detected")
	}
	if codexTrustPromptDetected(scrollback, "grok") {
		t.Fatalf("Codex trust prompt should not apply to grok")
	}
}

func TestCodexTurnCompleteForExit(t *testing.T) {
	idle := []byte("OpenAI Codex\n› say hi\n• Hi.\n› Write tests for @filename\ngpt-5.5 medium")
	if !codexTurnCompleteForExit(idle, compactBannerText("say hi")) {
		t.Fatalf("completed Codex turn was not detected")
	}
	busy := []byte("OpenAI Codex\n› say hi\n• Working (6s • esc to interrupt)\n› Write tests for @filename")
	if codexTurnCompleteForExit(busy, compactBannerText("say hi")) {
		t.Fatalf("busy Codex turn should not be treated as complete")
	}
}

func TestRunSubmitsPromptWithCarriageReturn(t *testing.T) {
	home := t.TempDir()
	fake := `sh -c 'printf "GROK_TTY_BANNER\nGrok › "; read -r line; echo "SUBMITTED:$line"'`
	t.Setenv("AGENT_RUN_GROK_TTY_COMMAND", fake)

	captured, _, err := Run(context.Background(), RunOptions{
		Home:      filepath.Join(home, ".agent-run"),
		Prompt:    "run ls",
		Workspace: home,
		Stderr:    os.Stderr,
	})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if captured != "run ls" {
		t.Fatalf("captured = %q, want run ls (scrollback should contain SUBMITTED:run ls)", captured)
	}
	if strings.Contains(captured, "UNSUBMITTED") {
		t.Fatalf("prompt submitted with bare LF: %q", captured)
	}
}
